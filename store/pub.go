package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	database "github.com/Ishan27g/ryo-Faas/database/db"
	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/transport"
)

var publish = func(subjId, data string) bool {
	return transport.NatsPublish(subjId, data, nil)
}

func marshal(doc database.NatsDoc) ([]byte, bool) {
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

	docs := d.Get(ids.Id[0])

	out, err := json.Marshal(docs[0])
	if err != nil {
		panic(err)
	}
	if publish(DocumentCREATE+d.natsSub(doc.Id()), string(out)) { // map[id]:data
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
	return publish(DocumentUPDATE+d.natsSub(id), document.DocumentString())

}

func (d *store) Get(ids ...string) []Doc {
	ctx, can := context.WithTimeout(context.Background(), time.Second*6)
	defer can()
	documents, err := d.dbClient.Get(ctx, &deploy.Ids{Id: ids})
	if err != nil {
		fmt.Println("store.Get()", err.Error())
		return nil
	}
	fmt.Println(documents)
	docs := ToDocs(documents, d.table)
	for _, doc := range docs {
		b, _ := json.Marshal(doc)
		go publish(DocumentGET+d.natsSub(doc.Id), string(b))
	}
	return docs
}

func ToDocs(documents *deploy.Documents, table string) []Doc {
	var docs []*database.NatsDoc

	for _, document := range documents.Document {
		var data map[string]interface{}
		err := json.Unmarshal(document.GetData(), &data)
		if err != nil {
			fmt.Println("json.Unmarshal", err.Error())
			return nil
		}
		doc := database.NewDocument(table, document.Id, data)
		docs = append(docs, &doc)
	}
	var entities []Doc
	for i, _ := range docs {
		value := (*docs[i]).Document()["Data"].(map[string]interface{})
		entities = append(entities, Doc{
			Id:        (*docs[i]).Document()["Id"].(string),
			CreatedAt: (*docs[i]).Document()["CreatedAt"].(string),
			EditedAt:  (*docs[i]).Document()["EditedAt"].(string),
			Data:      database.Data(struct{ Value map[string]interface{} }{Value: value["Value"].(map[string]interface{})}),
		})
	}
	return entities
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
