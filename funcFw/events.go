package FuncFw

import "github.com/Ishan27g/ryo-Faas/store"

type Events []struct {
	string // table name
	store.EventCb
}

type StoreEvents struct {
	OnCreate Events
	OnGet    Events
	OnUpdate Events
	OnDelete Events
}

func (s StoreEvents) Apply() (ok bool) {

	ok = false

	for _, cb := range s.OnCreate {
		docs := store.Get(cb.string)
		docs.OnCreate(cb.EventCb)
		ok = true
	}
	for _, cb := range s.OnGet {
		docs := store.Get(cb.string)
		docs.OnGet(cb.EventCb)
		ok = true
	}
	for _, cb := range s.OnUpdate {
		docs := store.Get(cb.string)
		docs.OnUpdate(cb.EventCb)
		ok = true
	}
	for _, cb := range s.OnDelete {
		docs := store.Get(cb.string)
		docs.OnDelete(cb.EventCb)
		ok = true
	}
	return
}
