package store

import (
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/Ishan27g/ryo-Faas/types"
)

var publish = func(subjId, data string) {
	transport.NatsPublish(subjId, data, nil)
}

func (d *store) Create(data string) {
	doc := d.new("", data)
	d.documents.Add(doc.Id(), doc)
	// DatabaseStore.Create(data)
	defer publish(DocumentCREATE+doc.Id(), doc.Data())
}

func (d *store) Update(id string, data string) {
	doc := d.new(id, data)
	d.documents.Add(doc.Id(), doc)
	//	DatabaseStore.Update(id, data)
	defer publish(DocumentUPDATE+doc.Id(), doc.Data())
}

func (d *store) Get(id string) *types.DocData {
	var doc types.DocData
	if dc := d.documents.Get(id); dc != nil {
		doc = dc.(types.DocData)
		defer publish(DocumentGET+doc.Id(), doc.Data())
	}
	return &doc
}

func (d *store) Delete(id string) {
	if dc := d.documents.Get(id); dc != nil {
		d.documents.Delete(id)
		doc := dc.(types.DocData)
		defer publish(DocumentDELETE+doc.Id(), doc.Data())
	}
}
