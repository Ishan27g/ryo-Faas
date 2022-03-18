package types

import "sync"

type SyncMap struct {
	sync.RWMutex
	data map[string]interface{}
}

func NewMap() SyncMap {
	return SyncMap{
		RWMutex: sync.RWMutex{},
		data:    make(map[string]interface{}),
	}
}
func (sm *SyncMap) Add(id string, val interface{}) {
	sm.Lock()
	defer sm.Unlock()
	sm.data[id] = val
}

func (sm *SyncMap) Delete(id string) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.data, id)
}
func (sm *SyncMap) Get(id string) interface{} {
	sm.RLock()
	defer sm.RUnlock()
	return sm.data[id]
}
func (sm *SyncMap) All() map[string]interface{} {
	sm.RLock()
	defer sm.RUnlock()
	var m = make(map[string]interface{})
	m = sm.data
	return m
}
