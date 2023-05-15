package types

import "sync"

type SyncMap[k comparable, v any] struct {
	sync.RWMutex
	data map[k]v
}

func NewMap[k comparable, v any]() SyncMap[k, v] {
	return SyncMap[k, v]{
		RWMutex: sync.RWMutex{},
		data:    make(map[k]v),
	}
}
func (sm *SyncMap[k, v]) Add(id k, val v) {
	sm.Lock()
	defer sm.Unlock()
	sm.data[id] = val
}

func (sm *SyncMap[k, v]) Delete(id k) {
	sm.Lock()
	defer sm.Unlock()
	delete(sm.data, id)
}
func (sm *SyncMap[k, v]) Get(id k) v {
	sm.RLock()
	defer sm.RUnlock()
	return sm.data[id]
}
func (sm *SyncMap[k, v]) All() map[k]v {
	sm.RLock()
	defer sm.RUnlock()
	var m = make(map[k]v)
	m = sm.data
	return m
}
