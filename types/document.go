package types

import (
	"fmt"

	"github.com/google/uuid"
)

type DocData interface {
	Id() string
	Data() string
	DataJson() map[string]interface{}
}

func NewDocData(id string, data map[string]interface{}) DocData {
	if id == "" {
		return &docData{id: uuid.New().String(), data: data}
	}
	return &docData{id: id, data: data}
}

type docData struct {
	id   string
	data map[string]interface{}
}

func (d *docData) DataJson() map[string]interface{} {
	return d.data
}

func (d *docData) Data() string {
	return fmt.Sprintf("%v", d.data)
}
func (d *docData) Id() string {
	return d.id
}
