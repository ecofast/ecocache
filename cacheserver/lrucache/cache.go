package lrucache

import (
	"sync"

	"github.com/ecofast/ecocache/cacheserver/errcode"
	"github.com/ecofast/ecocache/cacheserver/msgnode"
	"github.com/ecofast/ecocache/cacheserver/utils"
	"github.com/ecofast/ecocache/lru"
	. "github.com/ecofast/ecocache/protocols"
)

func newCache(index int, maxEntries, cacheBytes int, msgChan <-chan *msgnode.MsgNode) *cache {
	c := &cache{
		index:      index,
		cacheBytes: cacheBytes,
		msgChan:    msgChan,
	}
	c.lru = &lru.Cache{
		MaxEntries: maxEntries,
		OnEvicted: func(key lru.Key, value lru.Value) {
			c.numBytes -= len(key) + len(value)
		},
	}
	return c
}

type cache struct {
	index      int
	lru        *lru.Cache
	numBytes   int // of all keys and values
	cacheBytes int
	msgChan    <-chan *msgnode.MsgNode
}

func (self *cache) run(exitChan chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-exitChan:
			self.lru.Clear()
			return
		case msg := <-self.msgChan:
			if msg.Consumer != nil {
				switch msg.Cmd {
				case CM_GET:
					sz := len(msg.Body)
					v, ok := self.get(utils.BytesToStr(msg.Body[uint8(msg.Param>>8) : uint8(msg.Param>>8)+uint8(msg.Param)]))
					if ok {
						sz += len(v)
					}
					buf := make([]byte, sz)
					copy(buf, msg.Body)
					if ok {
						copy(buf[len(msg.Body):], v)
						msg.Consumer <- msgnode.New(nil, SM_GET, 0, msg.Param, buf)
						break
					}
					msg.Consumer <- msgnode.New(nil, SM_GET, errcode.ErrKeyDoNotExists, msg.Param, buf)
				case CM_SET:
					self.set(utils.BytesToStr(msg.Body[uint8(msg.Param>>8):uint8(msg.Param>>8)+uint8(msg.Param)]), msg.Body[uint8(msg.Param>>8)+uint8(msg.Param):])
					msg.Consumer <- msgnode.New(nil, SM_SET, 0, msg.Param, msg.Body)
				}
			}
		}
	}
}

func (self *cache) checkBytes(extra int) {
	for {
		if self.numBytes <= self.cacheBytes-extra || self.lru.Len() == 0 {
			return
		}
		self.lru.RemoveOldest()
	}
}

func (self *cache) set(key string, value []byte) {
	self.lru.Remove(key)
	self.checkBytes(len(key) + len(value))
	self.lru.Add(key, value)
	self.numBytes += len(key) + len(value)
}

func (self *cache) get(key string) (value []byte, ok bool) {
	return self.lru.Get(key)
}
