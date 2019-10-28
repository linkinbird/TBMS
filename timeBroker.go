package TBMS

import {
	"io"
	"log"
	"net"
	"time"
	"os"
	"sync"
}

// main class
type TimeBroker struct {
	treeOfQueue timeTree
	messQueue map[int]*timeTree
	nextCache map[*timeTree]bool	//cache as locks
	nextChan chan *timeTree			//concurrent channel
	mu sync.Mutex
}

func (t *TimeBroker) Init() {
	t.treeOfQueue = make(timeTree)
	t.messQueue = make(map[int]*timeTree)
	t.nextCache = make(map[*timeTree]bool)
	t.nextChan = make(chan *timeTree, 10)
}

func (t *TimeBroker) Listen(proto string, address string) {
	listener, err := net.Listen(proto, address)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Listening to %s\n", address)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err) // e.g., connection aborted
			continue
		}
		go t.append(conn)
	}
}

func (t *TimeBroker) append(conn net.Listener) {
	data, err := json.Unmarshal(conn)
	if err != nil {
		log.Fatalf("JSON of model request unmarshaling failed: %s", err)
	}
	new = timeRequest{data, nil, 0}	//fill the missing slot
	go t.requestOne(nil, new.Pop())

	if new.Init() {
		var checkLeft bool
		t.treeOfQueue, checkLeft = t.treeOfQueue.Insert(new, true, true)	// sometimes the root of the tree will change
		t.nextUpdate(checkLeft)
	}
}

func (t *TimeBroker) nextUpdate(checkLeft bool) {
	if checkLeft {
		tmp := t.treeOfQueue.Leftest()
		t.mu.Lock()	//mutex loc
		if !t.nextCache[tmp] {
			t.nextCache[tmp] = true
			t.nextChan <- tmp
		}
		t.mu.Unlock()
	}
}

func (t *TimeBroker) Request() {
	maxDelay := time.second
	var next *timeTree
	for {
		// Looping in Parallel with channel
		if len(t.nextChan)>0 {
			next <- t.nextChan
			go t.requestNext(next, maxDelay)
		}
	}
}

func (t *TimeBroker) requestNext(next *timeTree, maxDelay time.second) {
	nodeTime := next.NodeTime()
	for time.Now() < nodeTime{
		delay := min(maxDelay, (nodeTime - time.Now()) / 10)
		time.Sleep(delay)
	}

	model, receiver, checkLeft := t.treeOfQueue.PopUpdate(next)
	go t.requestOne(next, model, receiver)
	delete(t.nextCache, next)	//concurrency-safe to delete
	t.nextUpdate(true)	//also check after poped
}

func (t *TimeBroker) requestOne(node *timeTree, model *model, receiver net.Endpoint) {
	messID = getUDID()
	t.messQueue[messID] = node

	conn = net.Connect(model.remote)
	conn.Send(model, receiver, messID)	// to service mesh
	defer conn.Close()
}

func (t *TimeBroker) ReceiveFoward(proto string, address string) {
	listener, err := net.Listen(proto, address)
	if err != nil {
		log.Fatal(err)
	} else {
		fmt.Printf("Receving service from %s\n", address)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err) // e.g., connection aborted
			continue
		}
		data, err := json.Unmarshal(conn)
		if err != nil {
			log.Fatalf("JSON of model request unmarshaling failed: %s", err)
		}
		connf, err := net.Dial(data.receiver)
		if err != nil {
			log.Fatal(err)
		}
		connf.Send(data.value)
		connf.Close()
		t.treeOfQueue.Remove(t.messQueue[data.messID])	//delete from tree by ID
		delete(t.messQueue, data.messID)
	}
}

func getUDID () int{
	return random(unique)
}

// sub class is a balanced fork tree
// Almost concurrent-safe
// example in Java https://github.com/npgall/concurrent-trees
type timeTree struct {
	value timeRequest
	left, right *timeTree
	muL, muR sync.Mutex	//locks of leafs
}

func (tree *timeTree) Insert(new timeRequest, balance bool, checkLeft bool) *timeTree, bool{
	// insert root
	if tree.value = nil & tree.left = nil & tree.right = nil {
		tree.value = new
		return tree, true
	}

	// Mutex lock early will lock whole tree and all inserts
	// Separate lock logic in every branch and check flag after lock
	rightFlag := tree.right

	// fork tree insert
	if tree.NodeTime() > new.NextTime() {
		tree.muL.Lock()
		if tree.left != nil {
			tree.muL.Unlock()	//Unlock early in recursion
			_, checkLeft = tree.left.Insert(new, false, checkLeft)
		} else {
			newTree = make(timeTree)
			newTree.value = new
			newTree.left = nil
			newTree.right = nil
			tree.left = *newTree
			tree.muL.Unlock()
		}
	} else if tree.right != nil & tree.right.NodeTime() > new.NextTime() {
		tree.muR.Lock()	//right leaf may change during lock time
		checkLeft = false	// if search turn right, means leftest node not altered
		if tree.right == rightFlag {
			newTree = make(timeTree)
			newTree.value = new
			newTree.left = nil
			newTree.right = tree.right
			tree.right = *newTree
			tree.muR.Unlock()
		} else {
			tree.muR.Unlock()
			tree.right.Insert(new,false,checkLeft)
		}
	} else {
		tree.muR.Lock()
		checkLeft = false
		if tree.right != nil {
			tree.muR.Unlock()	//Unlock early in recursion
			tree.right.Insert(new, false, checkLeft)
		} else {
			newTree = make(timeTree)
			newTree.value = new
			newTree.left = nil
			newTree.right = nil
			tree.right = *newTree
			tree.muR.Unlock()
		}
		// Only balance on the root or will broken the branch
		// change root to the right one
		if balance {
			tree.muR.Lock()
			oldRoot := tree
			leaf := Leftest(oldRoot.right)
			leaf.muL.Lock()	//may be late
			leaf.left = oldRoot
			leaf.muL.Unlock()
			tree = oldRoot.right	//OK to accept new traffic
			oldRoot.right = nil
			tree.muR.Unlock()
		}
	}
	return tree, checkLeft
}

func (tree *timeTree) Remove(node *timeTree) *timeTree {
	if node != nil {
		// fork tree remove
	}
}

func (tree *timeTree) PopUpdate(node *timeTree) *model, net.Endpoint {
	// pop and fork tree update
	nodeValue, anyLeft := node.value.Pop()
	if anyLeft {
		tree, _ = tree.Insert(node.value,true,true)
	}
	tree.Remove(node)
	return nodeValue
}

func (tree *timeTree) Leftest() *timeTree{
	if tree.left != nil {
		return leftest(tree.left)
	} else {
		return tree
	}
}

func (tree *timeTree) NodeTime() time {
	return tree.value.NextTime()
}

// parse from json
type timeRequest struct {
	requests []*model
	receiver net.Endpoint
	tloc time
}
//{"embedding":{"delay":0},
//"svm":{"delay":35},
//"bayes":{"delay":25},
//"keysearch":{"delay":15},
//"tloc":60
//}

func (tr *timeRequest) Init() bool{
	if len(tr.requests) >0 {
		// find next model and time
		trIncrease := time.Now()
		tr.tloc := tr.tloc + trIncrease
		for model := range tr.requests {
			trIncrease += model.time
			if trIncrease > tr.tloc {
				trIncrease = tr.tloc
			}
			model.time = trIncrease
		}
		return true
	} else {
		return false
	}
}

func (tr *timeRequest) Pop() *model, net.Endpoint, bool{
	var poped *model
	anyLeft := false
	if len(tr.requests) >0 {
		poped = tr.requests[0]
		tr.requests = tr.requests[1:len(tr.requests)-1]
	}
	if len(tr.requests) >0 {
		anyLeft = true
	}
	return poped, tr.receiver, anyLeft
}

func (tr *timeRequest) NextTime() time {
	if len(tr.requests)>0 {
		return tr.requests[0].Time
	}
}

type model struct {
	Remote RemoteModel
	Value string	//parse in the model handler
	Time time
}

// export to model house
type RemoteModel struct {
	Name string
	Endpoint net
	Attribute map[string]int
}

func main() {
	timeBroker := make(TimeBroker)	//make init for sub class too
	timeBroker.Init()
	addresses := map[string]string {
		"tcp","localhost:8000"
		"tcp","bcrb.com"
	}
	servicess := map[string]string {
		"tcp","localhost:8000"
		"tcp","bcrb.com"
	}

	for proto, address := range addresses {
		go timeBroker.Listen(proto,address)
	}

	go timeBroker.Request()
	for proto, address := range servicess {
		go timeBroker.ReceiveFoward(proto,address)
	}
}