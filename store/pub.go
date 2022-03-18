package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/transport"
)

var publish = func(subjId, data string) bool {
	return transport.NatsPublish(subjId, data, nil)
}

func marshal(doc NatsDoc) ([]byte, bool) {
	docData, err := json.Marshal(doc.Document())
	if err != nil {
		fmt.Println("json.Marshal", err.Error())
		return nil, false
	}
	return docData, true
}

func (d *store) natsSub(id string) string {
	return "." + d.table + "." + id
}

func (d *store) Create(id string, data map[string]interface{}) string {
	ctx, can := context.WithTimeout(context.Background(), time.Second*2)
	defer can()
	doc := d.new(d.table, id, data)
	docData, done := marshal(doc)
	if !done {
		fmt.Println("Cannot marshal")
		return ""
	}
	ids, err := d.dbClient.New(ctx, &deploy.Documents{Document: []*deploy.Document{
		{
			Table: d.table,
			Id:    doc.Id(),
			Data:  docData,
		},
	}})
	if err != nil {
		fmt.Println("store.New()", err.Error())
		return ""
	}
	if ids.Id[0] != doc.Id() {
		fmt.Println("who dun it")
	}
	if publish(transport.DocumentCREATE+d.natsSub(doc.Id()), doc.DocumentString()) { // map[id]:data
		return ids.Id[0]
	}
	return ""
}

func (d *store) Update(id string, data map[string]interface{}) bool {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	document := d.new(d.table, id, data)
	docData, done := marshal(document)
	if !done {
		return false
	}
	_, err := d.dbClient.Update(ctx, &deploy.Documents{Document: []*deploy.Document{
		{
			Table: d.table,
			Id:    document.Id(),
			Data:  docData,
		},
	}})
	if err != nil {
		fmt.Println("store.Update()", err.Error())
		return false
	}
	return publish(transport.DocumentUPDATE+d.natsSub(id), document.DocumentString())

}

func (d *store) Get(ids ...string) []*NatsDoc {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	var docs []*NatsDoc
	documents, err := d.dbClient.Get(ctx, &deploy.Ids{Id: ids})
	if err != nil {
		fmt.Println("store.Get()", err.Error())
		return nil
	}
	fmt.Println(documents)
	for _, document := range documents.Document {
		var data map[string]interface{}
		err := json.Unmarshal(document.GetData(), &data)
		if err != nil {
			fmt.Println("json.Unmarshal", err.Error())
			return nil
		}
		doc := NewDocument(d.table, document.Id, data)
		docs = append(docs, &doc)
		go publish(transport.DocumentGET+d.natsSub(doc.Id()), doc.DocumentString())
	}
	return docs
}

func (d *store) Delete(ids ...string) bool {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	_, err := d.dbClient.Delete(ctx, &deploy.Ids{Id: ids})
	if err != nil {
		fmt.Println("store.Delete()", err.Error())
		return false
	}
	return true
}
