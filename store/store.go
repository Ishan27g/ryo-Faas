package store

import (
	"log"

	"github.com/Ishan27g/ryo-Faas/database/client"
)

var (
	defaultTable    = "data"
	databaseAddress = "localhost:5000" // os.Getenv("Database")
	documents       = make(map[string]DocStore)
)

func init() {

}

type EventCb func(document NatsDoc)

// DocStore exposes methods that
// - publish a nats-message after completing the respective dbClient action
// - register EventCb's against respective dbClient operations
type DocStore interface {
	// publish

	Create(id string, data map[string]interface{}) bool
	Update(id string, data map[string]interface{}) bool
	Get(id ...string) []*NatsDoc
	Delete(id ...string) bool

	// subscribe

	OnCreate(do EventCb)
	OnUpdate(do EventCb, ids ...string) // subscribe to all ids if nil
	OnDelete(do EventCb, ids ...string) // subscribe to all ids if nil
	OnGet(do EventCb, ids ...string)    // subscribe to all ids if nil

	On(subjId string, do EventCb)
}

type store struct {
	table    string
	new      func(table, id string, data map[string]interface{}) NatsDoc
	dbClient database.Client
}

func new(table string) DocStore {
	var dbClient database.Client
	if dbClient = database.Connect(databaseAddress); dbClient == nil {
		log.Fatal("cannot connect to database")
	}
	if table == "" {
		table = defaultTable
	}
	documents[table] = &store{table: table, new: NewDocument, dbClient: dbClient}
	return documents[table]
}

func Get(table string) DocStore {
	if documents[table] == nil {
		return new(table)
	}
	return documents[table]
}
