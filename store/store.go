package store

import (
	"fmt"
	"os"

	database "github.com/Ishan27g/ryo-Faas/database/client"
	db "github.com/Ishan27g/ryo-Faas/database/db"
)

var (
	defaultTable       = "data"
	databaseAddressEnv = "DATABASE"
	databaseAddress    = "localhost:5000"          // os.Getenv("Database")
	documents          = make(map[string]DocStore) // per table
)

const (
	DocumentCREATE string = "new"
	DocumentUPDATE string = "update"
	DocumentGET    string = "get"
	DocumentDELETE string = "delete"
)

type EventCb func(document Doc)
type Doc db.Entity
type DocStore interface {
	// publish

	Create(id string, data map[string]interface{}) string
	Update(id string, data map[string]interface{}) bool
	Get(id ...string) []Doc
	Delete(id ...string) bool

	// subscribe

	// OnCreate(do EventCb)
	// OnUpdate(do EventCb, ids ...string) // subscribe to all ids if nil
	// OnDelete(do EventCb, ids ...string) // subscribe to all ids if nil
	// OnGet(do EventCb, ids ...string)    // subscribe to all ids if nil

	On(eventType string, do EventCb, ids ...string) (ok bool) // subscribe to all ids if nil

	//	On(subjId string, do EventCb)
}

type store struct {
	table    string
	new      func(table, id string, data map[string]interface{}) db.NatsDoc
	dbClient database.Client
}

func new(table string) DocStore {

	d := os.Getenv(databaseAddressEnv)
	if d != "" {
		databaseAddress = d
	}

	var dbClient database.Client
	if dbClient = database.Connect(databaseAddress); dbClient == nil {
		fmt.Println("cannot connect to database")
		return nil
	}
	if table == "" {
		table = defaultTable
	}
	documents[table] = &store{table: table, new: db.NewDocument, dbClient: dbClient}

	return documents[table]
}

func Get(table string) DocStore {
	if documents[table] == nil {
		return new(table)
	}
	return documents[table]
}

func ok() {
	// get handler for `payments` document
	docStore := Get("payments")

	// data to add
	data := map[string]interface{}{
		"from":   "bob",
		"to":     "alice",
		"amount": 42,
	}

	// subscribe event functions for this document
	// go func() {
	// 	go func() {
	// 		docStore.OnCreate(func(document NatsDoc) {
	// 			fmt.Println("New payment ")
	// 			document.Print()
	// 		})
	// 	}()
	// 	go func() {
	// 		docStore.OnGet(func(document NatsDoc) {
	// 			fmt.Println("Retrived payment ")
	// 			document.Print()
	// 		})
	// 	}()
	// 	go func() {
	// 		docStore.OnUpdate(func(document NatsDoc) {
	// 			fmt.Println("Updated payment ")
	// 			document.Print()
	// 		})
	// 	}()
	// 	go func() {
	// 		docStore.OnDelete(func(document NatsDoc) {
	// 			fmt.Println("Deleted payment ")
	// 			document.Print()
	// 		})
	// 	}()

	// }()

	// add a new `payment` to the db
	id := docStore.Create("", data)

	// get it from the db
	dataReturned := docStore.Get(id)

	// dataReturned == data
	fmt.Println(dataReturned)

	// update some field
	data["amount"] = 43
	docStore.Update(id, data)

	// delete it
	docStore.Delete(id)

}
