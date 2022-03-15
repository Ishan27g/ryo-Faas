package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/Ishan27g/ryo-Faas/types"
	"github.com/nats-io/nats.go"
)

func (d *store) onNatsMsg(msg *nats.Msg, do Event) {
	docId := msg.Subject // todo strings.Trim(subj.DataId)
	var docData map[string]interface{}
	err := json.Unmarshal(msg.Data, &docData)
	if err != nil {
		fmt.Println("json.Unmarshal", err.Error())
		if doc := d.Get(docId); doc == nil {
			fmt.Println("database.Get", docId+" not found")
			return
		}
	}
	do(types.NewDocData(docId, docData))
}

func (d *store) on(subject string, do Event, ids ...string) {
	switch len(ids) {
	case 0:
		ctx, can := context.WithTimeout(context.Background(), time.Second*6)
		defer can()
		all, _ := d.database.All(ctx, &deploy.Ids{Id: ids})
		for _, doc := range all.Document { // todo
			transport.NatsSubscribe(subject+"."+doc.Id, func(msg *nats.Msg) {
				d.onNatsMsg(msg, do)
			})
		}
	default:
		for _, id := range ids {
			transport.NatsSubscribe(subject+"."+id, func(msg *nats.Msg) {
				d.onNatsMsg(msg, do)
			})
		}
	}
}
func (d *store) OnCreate(do Event) {
	transport.NatsSubscribe(DocumentCREATE, func(msg *nats.Msg) {
		d.onNatsMsg(msg, do)
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
