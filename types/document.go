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

func FromNats(doc map[string]interface{}) NatsDoc {
	var document NatsDoc
	for id, data := range doc {
		document = NewNatsDoc(id, data.(map[string]interface{}))
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
