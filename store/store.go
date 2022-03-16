package store

import (
	databaseClient "github.com/Ishan27g/ryo-Faas/database/client"
	"github.com/Ishan27g/ryo-Faas/types"
)

var (
	databaseAddress          = ""
	st              DocStore = &store{new: types.NewNatsDoc, database: databaseClient.Connect(databaseAddress)}
	Documents                = st
)

type EventCb func(document types.NatsDoc)

// DocStore exposes methods that
// - publish a nats-message after completing the respective database action
// - register EventCb's against respective database operations
type DocStore interface {
	// publish

	Create(id string, data map[string]interface{})
	Update(id string, data map[string]interface{})
	Get(id ...string) []*types.NatsDoc
	Delete(id ...string)

	// subscribe

	OnCreate(do EventCb)
	OnUpdate(do EventCb, ids ...string) // subscribe to all ids if nil
	OnDelete(do EventCb, ids ...string) // subscribe to all ids if nil
	OnGet(do EventCb, ids ...string)    // subscribe to all ids if nil

	On(subjId string, do EventCb)
}

type store struct {
	new      func(id string, data map[string]interface{}) types.NatsDoc
	database databaseClient.Client
}
