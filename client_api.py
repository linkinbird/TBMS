#!/bin/python
# -*- coding: utf8 -*-

from tbms import client

ENDPOINT = 'http://serving.tbms.io'
ACCESSID = 'asssv'
ACCESSKEY = '69Pk88nj'
INSTANCENAME = 'nianjianQA'

tbms_client = tbmsClient(ENDPOINT, ACCESSID, ACCESSKEY, INSTANCENAME)
tbms_models = tbmsList({"embedding":{"est":35},
                        "svm":{"est":25},
                        "bayes":{"est":15},
                        "keysearch"{"est":5}
                       })

tbms_mixModels = tbmsMixList({"mix1":["embedding":{"est":35},"svm":{"est":25},"mix":"max"],
                              "mix2":["bayes":{"est":15},"keysearch"{"est":5},"mix":"avg"]
                             })

def main():
    questionString = sys.argv
    if len(sys.argv)>=1:
        questionString = sys.argv[0]
        answer = tbmsTry(tbmsClient,tbms_models,questionString,tloc=50,crossRequest=1,crossLag=10,priority=0)
        # answer = tbmsMixTry(tbmsClient,tbms_mixModels,questionString,tloc=50,crossRequest=1,crossLag=10)
        if len(answer) >0:
            print answer
        else:
            print u'human assist'

if __name__ == '__main__':
    main()
