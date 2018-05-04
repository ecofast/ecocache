package lrucache

import (
	"cacheserver/msgnode"
	"lru"
	. "protocols"
	"sync"
	"unsafe"
)

func newCache(cacheBytes int, msgChan <-chan *msgnode.MsgNode) *cache {
	c := &cache{
		cacheBytes: cacheBytes,
		msgChan:    msgChan,
	}
	c.lru = &lru.Cache{
		OnEvicted: func(key lru.Key, value lru.Value) {
			c.numBytes -= len(key) + len(value)
		},
	}
	return c
}

// cache is a wrapper around an *lru.Cache that adds synchronization
type cache struct {
	lru        *lru.Cache
	numBytes   int // of all keys and values
	cacheBytes int
	msgChan    <-chan *msgnode.MsgNode
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func (self *cache) run(exitChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	self.warmup()
	for {
		select {
		case <-exitChan:
			self.lru.Clear()
			return
		case msg := <-self.msgChan:
			if msg.Consumer != nil && msg.Cmd == CM_GET {
				if v, ok := self.get(bytes2str(msg.Body)); ok {
					msg.Consumer <- msgnode.New(nil, SM_GET, msg.Param, v)
				} else {
					msg.Consumer <- msgnode.New(nil, SM_GET, msg.Param, nil)
				}
			}
		}
	}
}

func (self *cache) warmup() {
	// to do
	//
}

func (self *cache) checkBytes(extra int) {
	for {
		if self.numBytes <= self.cacheBytes-extra {
			return
		}
		self.lru.RemoveOldest()
	}
}

func (self *cache) add(key string, value []byte) {
	self.lru.Add(key, value)
	self.numBytes += len(key) + len(value)
}

func (self *cache) load(key string) (value []byte, ok bool) {
	// to do
	//
	return nil, false
}

func (self *cache) get(key string) (value []byte, ok bool) {
	if v, ok := self.lru.Get(key); ok {
		return v, ok
	}
	if v, ok := self.load(key); ok {
		self.checkBytes(len(key) + len(v))
		self.add(key, v)
		return v, ok
	}
	return nil, false
}
