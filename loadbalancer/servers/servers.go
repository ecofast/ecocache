package servers

import (
	"consistenthash"
	"loadbalancer/cfgmgr"
	"sync"
)

var (
	mutex   sync.RWMutex
	svrList map[string][]byte
	svrMap  *consistenthash.Map
)

func Setup() {
	svrList = make(map[string][]byte, 5)
	svrMap = consistenthash.New(cfgmgr.ServerReplicas(), nil)
}

func Store(key string, value []byte) {
	mutex.Lock()
	defer mutex.Unlock()
	svrList[key] = value
	svrMap.Add(value)
}

func Remove(key string) bool {
	mutex.Lock()
	defer mutex.Unlock()
	if v, ok := svrList[key]; ok {
		svrMap.Remove(v)
		delete(svrList, key)
		return true
	}
	return false
}

func Get(key string) []byte {
	mutex.RLock()
	defer mutex.RUnlock()
	return svrMap.Get(key)
}

func Count() int {
	mutex.RLock()
	defer mutex.RUnlock()
	return len(svrList)
}
