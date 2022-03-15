package types

import (
	"github.com/google/uuid"
)

type DocData interface {
	Id() string
	Data() string
}

func NewDocData(id, data string) DocData {
	if id == "" {
		return &docData{id: uuid.New().String(), data: data}
	}
	return &docData{id: id, data: data}
}

type docData struct {
	id   string
	data string
}

func (d *docData) Data() string {
	return d.data
}
func (d *docData) Id() string {
	return d.id
}
