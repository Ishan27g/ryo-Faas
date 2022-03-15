package store

import (
	"github.com/Ishan27g/ryo-Faas/types"
)

var (
	st Store = &store{documents: types.NewMap(), new: types.NewDocData}
)

const (
	DocumentCREATE = "new"
	DocumentUPDATE = "update"
	DocumentGET    = "get"
	DocumentDELETE = "delete"
)

type Event func(document types.DocData)

// Store implements a document store that publishes nats-messages for the respective document action
type Store interface {
	// publish

	Create(data string)
	Update(id string, data string)
	Get(id string) *types.DocData
	Delete(id string)

	// subscribe

	OnCreate(do Event)
	OnUpdate(do Event, ids ...string) // subscribe to all ids if nil
	OnDelete(do Event, ids ...string) // subscribe to all ids if nil
	OnGet(do Event, ids ...string)

	On(subjId string, do Event)
}

func GetStore() Store {
	return st
}

type store struct {
	documents types.SyncMap
	new       func(id, data string) types.DocData
}
