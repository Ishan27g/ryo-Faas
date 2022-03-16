package types

import (
	"fmt"

	"github.com/google/uuid"
)

type NatsDoc interface {
	Id() string
	Data() string
	Document() map[string]interface{} // natsDoc as map
	Print()
}

func FromJson(doc map[string]interface{}, id ...string) NatsDoc {
	var document NatsDoc
	if doc["Id"] == nil && len(id) == 0 {
		var m map[string]interface{} = make(map[string]interface{})
		m["Id"] = doc
		for _, data := range m {
			document = NewNatsDoc("", data.(map[string]interface{}))
		}
	} else {
		//for id, data := range doc {
		document = NewNatsDoc(id[0], doc)
		//}
	}

	return document
}
func NewNatsDoc(id string, data map[string]interface{}) NatsDoc {
	if id == "" {
		return &natsDoc{id: uuid.New().String(), data: data}
	}
	return &natsDoc{id: id, data: data}
}

type natsDoc struct {
	id   string
	data map[string]interface{}
}

func (d *natsDoc) Print() {
	fmt.Println(d.Id(), fmt.Sprintf("%v", d.data))
}
func (d *natsDoc) Document() map[string]interface{} {
	m := make(map[string]interface{})
	m[d.id] = d.data
	return m
}

func (d *natsDoc) Data() string {
	return fmt.Sprintf("%v", d.data)
}
func (d *natsDoc) Id() string {
	return d.id
}
