package FuncFw

import "github.com/Ishan27g/ryo-Faas/store"

var events = make(map[string]StoreEventsI) // tableName:

type StoreEventsI interface {
	OnCreate(...store.EventCb)
	OnGet(...store.EventCb)
	OnUpdate(...store.EventCb)
	OnDelete(...store.EventCb)
	OnGetIds(cb store.EventCb, id string)
	OnUpdateIds(cb store.EventCb, id string)
	OnDeleteIds(cb store.EventCb, id string)
	get() storeEvents
}
type eventCb struct {
	cb  store.EventCb
	ids []string
}
type storeEvents struct {
	Table    string // table name
	onCreate []eventCb
	onGet    []eventCb
	onUpdate []eventCb
	onDelete []eventCb
}

func EventsForTable(tableName string) StoreEventsI {
	if events[tableName] == nil {
		events[tableName] = &storeEvents{
			Table:    tableName,
			onCreate: nil,
			onGet:    nil,
			onUpdate: nil,
			onDelete: nil,
		}
	}
	return events[tableName]
}

func ApplyEvents() (ok bool) {
	ok = false
	for tableName, se := range events {
		s := se.get()
		docs := store.Get(tableName)
		for _, e := range s.onCreate {
			docs.OnCreate(e.cb)
			ok = true
		}
		for _, e := range s.onGet {
			docs.OnGet(e.cb, e.ids...)
			ok = true
		}
		for _, e := range s.onUpdate {
			docs.OnUpdate(e.cb, e.ids...)
			ok = true
		}
		for _, e := range s.onDelete {
			docs.OnDelete(e.cb, e.ids...)
			ok = true
		}
	}
	return
}
func (se *storeEvents) OnCreate(cb ...store.EventCb) {
	for _, s := range cb {
		se.onCreate = append(se.onCreate, eventCb{
			cb:  s,
			ids: nil,
		})
	}
}
func (se *storeEvents) OnGet(cb ...store.EventCb) {
	for _, s := range cb {
		se.onGet = append(se.onGet, eventCb{
			cb:  s,
			ids: nil,
		})
	}
}
func (se *storeEvents) OnUpdate(cb ...store.EventCb) {
	for _, s := range cb {
		se.onUpdate = append(se.onUpdate, eventCb{
			cb:  s,
			ids: nil,
		})
	}
}
func (se *storeEvents) OnDelete(cb ...store.EventCb) {
	for _, s := range cb {
		se.onDelete = append(se.onDelete, eventCb{
			cb:  s,
			ids: nil,
		})
	}
}
func (se *storeEvents) get() storeEvents {
	return *se
}

func (se *storeEvents) OnGetIds(cb store.EventCb, id string) {
	se.onGet = append(se.onGet, eventCb{
		cb:  cb,
		ids: []string{id},
	})
}

func (se *storeEvents) OnUpdateIds(cb store.EventCb, id string) {
	se.onUpdate = append(se.onUpdate, eventCb{
		cb:  cb,
		ids: []string{id},
	})
}

func (se *storeEvents) OnDeleteIds(cb store.EventCb, id string) {
	se.onDelete = append(se.onDelete, eventCb{
		cb:  cb,
		ids: []string{id},
	})
}
