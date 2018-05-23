package lrucache

import (
	"fmt"
	"sync"

	"github.com/ecofast/ecocache/cacheserver/cfgmgr"
	"github.com/ecofast/ecocache/cacheserver/msgnode"
)

func Setup() {
	fmt.Printf("num of lru caches: %d\n", cfgmgr.NumLRUCache())
	fmt.Printf("max entries per bucket: %d\n", cfgmgr.MaxEntries())
	fmt.Printf("max cache bytes per bucket: %d\n", cfgmgr.MaxCacheBytesPerBucket())
}

func Run(exitChan chan struct{}, bucketChans []chan *msgnode.MsgNode, waitGroup *sync.WaitGroup) {
	defer waitGroup.Done()

	wg := &sync.WaitGroup{}
	for i := 0; i < cfgmgr.NumLRUCache(); i++ {
		wg.Add(1)
		go func(i int) {
			newCache(i, cfgmgr.MaxEntries(), cfgmgr.MaxCacheBytesPerBucket(), bucketChans[i]).run(exitChan, wg)
		}(i)
	}

	<-exitChan
	wg.Wait()
}
