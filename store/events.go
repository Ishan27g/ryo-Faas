package store

import "github.com/Ishan27g/ryo-Faas/types"

type EventCb func(document types.NatsDoc)

type StoreEvents struct {
	OnCreate []EventCb
	OnGet    []EventCb
	OnUpdate []EventCb
	OnDelete []EventCb
}

func (s StoreEvents) Apply() (ok bool) {
	ok = false
	for _, cb := range s.OnCreate {
		Documents.OnCreate(cb)
		ok = true
	}
	for _, cb := range s.OnGet {
		Documents.OnGet(cb)
		ok = true
	}
	for _, cb := range s.OnUpdate {
		Documents.OnUpdate(cb)
		ok = true
	}
	for _, cb := range s.OnDelete {
		Documents.OnDelete(cb)
		ok = true
	}
	return
}
