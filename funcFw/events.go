package FuncFw

import "github.com/Ishan27g/ryo-Faas/store"

type EventCbs []store.EventCb

type StoreEvents struct {
	OnCreate EventCbs
	OnGet    EventCbs
	OnUpdate EventCbs
	OnDelete EventCbs
}

func (s StoreEvents) Apply() (ok bool) {
	ok = false
	for _, cb := range s.OnCreate {
		store.Documents.OnCreate(cb)
		ok = true
	}
	for _, cb := range s.OnGet {
		store.Documents.OnGet(cb)
		ok = true
	}
	for _, cb := range s.OnUpdate {
		store.Documents.OnUpdate(cb)
		ok = true
	}
	for _, cb := range s.OnDelete {
		store.Documents.OnDelete(cb)
		ok = true
	}
	return
}
