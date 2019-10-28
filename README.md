# TBMS
time based model serving, still a proof of concept

## Model serving problem in SysML

We have a bunch of model training tools for AI and machine learning. But lack of the tool to put your model in distributed service. [Clipper](https://github.com/ucbrise/clipper) from UC Berkeley is one tool to use, but only links the existing services and add some work flow. TBMS want to design it from the very beginning, and use the most advanced tech to build a solution framework. 

![tech_eco](/ref/tech_eco.png)

In the ecosystem, from bottom up are processing chips, execution lib, design tools and model serving. The hardware team will focus on the execution layer, the software team will focus on the model design tools, but the business team who want to put those models into use cares a lot about the serving management. The bottom layers are changing fast, we need good tools to handle those complexity and make the service stable while evolving.

### Opensoure or Cloud

The easy choice is to use the cloud services, which are easy to train and easy to publish. Traditional company without an AI team will go this way. They get the benefit of the advanced AI with very little cost, but lost the freedom. Most tech company will choose to use open source AI tools. But those tools are changing fast and the advanced AI team are expensive. So the solution we pursue is a model as a service platform. The IT team can handle this platform as easy as any other service, while the model team can use the opensource module like [bert](https://github.com/google-research/bert) in NLP to deliver a fast solution with much lower research cost. See [bert-as-service](https://github.com/hanxiao/bert-as-service) as an example.

### Visual machine or K8s

Google has down a lot of work in this field. By combining the tensorflow and kubernetes they get [kuberflow](https://github.com/kubeflow/kubeflow) which is an elastic AI model serving tool with the benefit of both sides. Yes we know electric cars are good but there're still a lot of petrol engines running on the road. So as ordinary business companies we want the freedom to handle all the possibilities and welcome the competition of the service provider, so both VM and K8s are considered.

## Time broker and service mesh solution

Here is our solution: First of all, AI models are time consuming and the users are not time sensitive, so we need to handle the time variable. Second, AI models are fast evolving so we need A/B serving and always have a backup. Then most models run on GPUs so we need batching optimization and do trade of when there are not enough resources. Finally to decouple the model management and the execution management we need a model repository. 

Put the whole framework in a picture is like this, let's discuss each component in detail.

![worlfow](/ref/worlfow.png)

### Time broker

Why time matters? See modern chatbot from DeepPavlov here. The whole service is provided by a workflow with a lot of components. Some components are slow some are fast. Every time you upgrade to a more sophisticated solution you need a backup.

![chatbot](/ref/chatbot.png)

So we redesign the client request to include the backups and time limits. See example in **client_api.py** (*pseudocode only for poc*), each model requests list the backups and the estimated time. If timeout, backup models will be executed asynchronously. If all the execution time together pass over the 'tloc' limit, at least some thing will return.

```python
tbms_models = tbmsList({"embedding":{"est":35},
                        "svm":{"est":25},
                        "bayes":{"est":15},
                        "keysearch"{"est":5}
                       })
...
answer = tbmsTry(tbmsClient,tbms_models,questionString,tloc=50,crossRequest=1,crossLag=10,priority=0)
```

And for the server side, we send the first model request to the service mesh, and then add the backups to a timer. The timer is designed as an concurrent-safe fork tree in our example **timeBroker.go** (*pseudocode only for poc*). The timer can also use an ordinary queue but need to handle the concurrent insert and ranking nicely. This design allows us to use goroutines to concurrently send the timed request, with channel and locker to handle parallelization.

### Service mesh

Every request need to pass to a server to execute, but those micro services maybe changing, so we use existing service mesh to handle the request from the timeBroker. As we discuss above, VMs and kubernetes have to be considered, so [kuma](https://github.com/Kong/kuma) from Kong is recommended. Load balance, monitoring, service discover, identity and access management are all done here. But GPUs load balance are different, we have to balance between batch, lag and require more GPUs.

#### Batch optimization

GPUs can batch the same request together, use multi core parallelization like SIMD to speed up the process. But it has dilemma pictured below. The full batch has the best QPS, but waiting for the batch to be filled in less traffic will increase the lag. The user has to make the choice in configure file.

![batch_opi](/ref/batch_opi.png)

### GPU virtualization

Finally the load will be executed on a CPU, GPU or FPGA. GPU and TPU are the best choice for today's AI cloud serving. You can use a single GPU to handle multi models and multi requests. But the context switch will cost thousands times more then CPU. And the GPU level optimization has to rely on the GPU producer like Nvidia. Luckly Nvidia opensourced its [TensorRT](https://github.com/NVIDIA/tensorrt-inference-server) which has streaming. It can load the next context while processing the previous one. But in a multi GPU cloud, the TensorRT has to collaborate with service mesh to get the best batch performance.

### Model management

Ensemble model means we can recursively add an ensembled client model to the model repository. See model repository in tensorRT inference server.

![TensorRT](/ref/TensorRT.png)

## Vision and mission

We want AI power free to everyone.

We help every business no matter small or big to get the benefit of AI.
