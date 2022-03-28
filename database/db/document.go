package database

import (
	"fmt"

	"github.com/google/uuid"
)

type NatsDoc interface {
	Table() string
	Id() string
	Document() map[string]interface{} // natsDoc as map
	DocumentString() string

	Print()
}

func FromJson(table string, doc map[string]interface{}, id ...string) NatsDoc {
	var document NatsDoc
	if doc["Id"] == nil && len(id) == 0 {
		var m = make(map[string]interface{})
		m["Id"] = doc
		for _, data := range m {
			document = NewDocument(table, "", data.(map[string]interface{}))
		}
	} else {
		document = NewDocument(table, id[0], doc)
	}

	return document
}
func NewDocument(table, id string, data map[string]interface{}) NatsDoc {
	if id == "" {
		return &natsDoc{table: table, id: uuid.New().String(), data: data}
	}
	return &natsDoc{table: table, id: id, data: data}
}

type natsDoc struct {
	table string
	id    string
	data  map[string]interface{}
}

func (d *natsDoc) Table() string {
	return d.table
}

func (d *natsDoc) Print() {
	fmt.Println(d.Id(), fmt.Sprintf("%v", d.data))
}
func (d *natsDoc) Document() map[string]interface{} {
	return d.data
}

func (d *natsDoc) DocumentString() string {
	return fmt.Sprintf("%v", d.Document())
}
func (d *natsDoc) Id() string {
	return d.id
}
