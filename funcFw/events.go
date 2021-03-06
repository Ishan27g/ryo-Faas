package FuncFw

import (
	"github.com/Ishan27g/ryo-Faas/store"
)

type StoreEventsI interface {
	On(eventType string, cbs ...store.EventCb)
	OnIds(eventType string, cb store.EventCb, ids ...string)
	get() storeEvents
}
type eventCb struct {
	eventType string
	cb        store.EventCb
	ids       []string
}
type storeEvents struct {
	Table string // table name
	on    []eventCb
}

func applyEvents() (ok bool) {
	ok = false
	for tableName, se := range Export.storeEvents {
		s := se.get()
		docs := store.Get(tableName)
		for _, e := range s.on {
			if docs.On(e.eventType, e.cb, e.ids...) {
				ok = true
			} else {
				return false
			}
		}
	}
	return
}

func (se *storeEvents) get() storeEvents {
	return *se
}

func (se *storeEvents) On(eventType string, cbs ...store.EventCb) {
	for _, cb := range cbs {
		se.on = append(se.on, eventCb{
			eventType: eventType,
			cb:        cb,
			ids:       nil,
		})
	}
}
func (se *storeEvents) OnIds(eventType string, cb store.EventCb, ids ...string) {
	se.on = append(se.on, eventCb{
		eventType: eventType,
		cb:        cb,
		ids:       ids,
	})
}
