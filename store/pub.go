package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/Ishan27g/ryo-Faas/types"
)

var publish = func(subjId, data string) {
	transport.NatsPublish(subjId, data, nil)
}

func marshal(doc types.NatsDoc) ([]byte, bool) {
	docData, err := json.Marshal(doc.Document())
	if err != nil {
		fmt.Println("json.Marshal", err.Error())
		return nil, false
	}
	return docData, true
}

func (d *store) Create(id string, data map[string]interface{}) {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	doc := d.new(id, data)
	docData, done := marshal(doc)
	if !done {
		return
	}
	d.database.New(ctx, &deploy.Documents{Document: []*deploy.Document{
		{
			Id:   doc.Id(),
			Data: docData,
		},
	}})
	defer publish(transport.DocumentCREATE+doc.Id(), doc.Data()) // map[id]:data

}

func (d *store) Update(id string, data map[string]interface{}) {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	document := d.new(id, data)
	docData, done := marshal(document)
	if !done {
		return
	}
	d.database.Update(ctx, &deploy.Documents{Document: []*deploy.Document{
		{
			Id:   document.Id(),
			Data: docData,
		},
	}})
	defer publish(transport.DocumentUPDATE+"."+document.Id(), document.Data())

}

func (d *store) Get(ids ...string) []*types.NatsDoc {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	var docs []*types.NatsDoc
	documents, err := d.database.Get(ctx, &deploy.Ids{Id: ids})
	if err != nil {
		return nil
	}
	for _, document := range documents.Document {
		var data map[string]interface{}
		err := json.Unmarshal(document.GetData(), &data)
		if err != nil {
			fmt.Println("json.Unmarshal", err.Error())
			return nil
		}
		d := types.NewNatsDoc(document.Id, data)
		docs = append(docs, &d)
		go publish(transport.DocumentGET+"."+d.Id(), d.Data())
	}
	return docs
}

func (d *store) Delete(ids ...string) {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	d.database.Delete(ctx, &deploy.Ids{Id: ids})
}
