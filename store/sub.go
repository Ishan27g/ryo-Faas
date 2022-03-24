package store

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/nats-io/nats.go"
)

func (d *store) onNatsMsg(msg *nats.Msg, do EventCb) {
	// subject.table.id
	subJ := strings.Split(".", msg.Subject) // todo strings.Trim(subj.DataId)
	var docData map[string]interface{}      // map[id]:data
	var document Doc
	err := json.Unmarshal(msg.Data, &docData)
	if err == nil {
		//		document = database.FromJson(subJ[1], docData)
	} else {
		// if not be able to convert nats msg , go to db
		fmt.Println("json.Unmarshal", err.Error())
		if doc := d.Get(subJ[2]); doc == nil {
			fmt.Println("dbClient.Get: ", subJ[2]+" not found")
			return
		} else {
			document = doc[0]
		}
	}
	do(document)
}

func (d *store) on(subject string, do EventCb, ids ...string) {
	switch len(ids) {
	case 0:
		ctx, can := context.WithTimeout(context.Background(), time.Second*6)
		defer can()
		all, _ := d.dbClient.All(ctx, &deploy.Ids{Id: nil}) // ids unused
		for _, doc := range all.Document {                  // todo
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

// On For all ids in dbClient, subscribe to subject and call do() on subscription
func (d *store) On(eventType string, do EventCb, ids ...string) (ok bool) {
	ok = false
	switch eventType {
	case DocumentGET, DocumentUPDATE, DocumentDELETE:
		d.on(eventType+"."+d.table, do, ids...)
		ok = true
	case DocumentCREATE:
		d.on(eventType+"."+d.table, do)
		ok = true
	}
	if !ok {
		fmt.Println("Invalid eventType", eventType)
	}
	return
}
