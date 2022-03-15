package store

import (
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/nats-io/nats.go"
)

var onNatsMsg = func(msg *nats.Msg, do Event) {
	docId := string(msg.Data)
	if doc := GetDatabase().Get(docId); doc != nil {
		do(*doc)
	}
}

func (d *store) on(subject string, do Event, ids ...string) {
	switch len(ids) {
	case 0:
		docs := GetDatabase().all()
		for _, doc := range docs {
			transport.NatsSubscribe(subject+"."+doc.Id(), func(msg *nats.Msg) {
				onNatsMsg(msg, do)
			})
		}
	default:
		for _, id := range ids {
			transport.NatsSubscribe(subject+"."+id, func(msg *nats.Msg) {
				onNatsMsg(msg, do)
			})
		}
	}
}
func (d *store) OnCreate(do Event) {
	transport.NatsSubscribe(DocumentCREATE, func(msg *nats.Msg) {
		onNatsMsg(msg, do)
	})
}
func (d *store) OnGet(do Event, ids ...string) {
	d.on(DocumentGET, do, ids...)
}
func (d *store) OnUpdate(do Event, ids ...string) {
	d.on(DocumentUPDATE, do, ids...)
}
func (d *store) OnDelete(do Event, ids ...string) {
	d.on(DocumentDELETE, do, ids...)
}

// On For all ids in database, subscribe to subject and call do() on subscription
func (d *store) On(subject string, do Event) {
	d.on(subject, do)
}
