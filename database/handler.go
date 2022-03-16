package database

import (
	"context"
	"encoding/json"
	"fmt"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
	"github.com/Ishan27g/ryo-Faas/types"
)

type handler struct {
	Database
}

func toDoc(document *deploy.Document) types.NatsDoc {
	var data map[string]interface{}
	err := json.Unmarshal(document.Data, &data)
	if err != nil {
		fmt.Println("toDoc", err.Error())
		return nil
	}
	return types.NewNatsDoc(document.Id, data)
}

func (d handler) forEachDoc(documents *deploy.Documents, cb func(data types.NatsDoc)) (*deploy.Ids, error) {
	var ids *deploy.Ids
	var err error
	for _, document := range documents.Document {
		if doc := toDoc(document); doc != nil {
			cb(doc)
			ids.Id = append(ids.Id, doc.Id())
		} else {
			return nil, err
		}
	}
	return ids, nil
}
func (d handler) New(ctx context.Context, documents *deploy.Documents) (*deploy.Ids, error) {
	return d.forEachDoc(documents, func(doc types.NatsDoc) {
		d.Database.New(doc)
	})
}

func (d handler) Update(ctx context.Context, documents *deploy.Documents) (*deploy.Ids, error) {
	return d.forEachDoc(documents, func(doc types.NatsDoc) {
		d.Database.Update(doc)
	})
}

func (d handler) Get(ctx context.Context, ids *deploy.Ids) (*deploy.Documents, error) {
	var documents []*deploy.Document
	for _, id := range ids.Id {
		if doc := d.Database.Get(id); doc != nil {
			data, _ := json.Marshal((*doc).Document())
			documents = append(documents, &deploy.Document{
				Id:   (*doc).Id(),
				Data: data,
			})
		}
	}
	return &deploy.Documents{Document: documents}, nil
}

func (d handler) Delete(ctx context.Context, ids *deploy.Ids) (*deploy.Ids, error) {
	for _, id := range ids.Id {
		d.Database.Delete(id)
	}
	return nil, nil
}
