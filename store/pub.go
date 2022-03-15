package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/types"
)

func marshal(doc types.DocData) ([]byte, bool) {
	docData, err := json.Marshal(doc.DataJson())
	if err != nil {
		fmt.Println("json.Marshal", err.Error())
		return nil, false
	}
	return docData, true
}

func (d *store) Create(data map[string]interface{}) {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	doc := d.new("", data)
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
}

func (d *store) Update(id string, data map[string]interface{}) {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	doc := d.new(id, data)
	docData, done := marshal(doc)
	if !done {
		return
	}
	d.database.Update(ctx, &deploy.Documents{Document: []*deploy.Document{
		{
			Id:   doc.Id(),
			Data: docData,
		},
	}})
}

func (d *store) Get(ids ...string) []*types.DocData {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	var docs []*types.DocData
	documents, err := d.database.Get(ctx, &deploy.Ids{Id: ids})
	if err != nil {
		return nil
	}
	for _, document := range documents.Document {
		var data map[string]interface{}
		err := json.Unmarshal(document.GetData(), data)
		if err != nil {
			fmt.Println("json.Unmarshal", err.Error())
			return nil
		}
		d := types.NewDocData(document.Id, data)
		docs = append(docs, &d)
	}
	return docs
}

func (d *store) Delete(ids ...string) {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	d.database.Delete(ctx, &deploy.Ids{Id: ids})
}
