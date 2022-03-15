package types

import "sync"

type SyncMap struct {
	s sync.Map
}

func NewMap() SyncMap {
	return SyncMap{
		s: sync.Map{},
	}
}
func (sm *SyncMap) Add(id string, val interface{}) {
	sm.s.Store(id, val)
}
func (sm *SyncMap) Get(id string) interface{} {
	if s, f := sm.s.Load(id); f {
		return s
	}
	return nil
}
func (sm *SyncMap) Delete(id string) {
	sm.s.Delete(id)
}
func (sm *SyncMap) All() map[string]interface{} {
	var m = make(map[string]interface{})
	sm.s.Range(func(key, value interface{}) bool {
		m[key.(string)] = value
		return true
	})
	return m
}
