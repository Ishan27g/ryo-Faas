package store

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	database "github.com/Ishan27g/ryo-Faas/database/client"
	db "github.com/Ishan27g/ryo-Faas/database/db"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
)

var (
	defaultTable    = "data"
	databaseAddress = os.Getenv("DATABASE")            // set "" if local
	documents       = types.NewMap[string, DocStore]() // per table
)

const (
	DocumentCREATE string = "new"
	DocumentUPDATE string = "update"
	DocumentGET    string = "get"
	DocumentDELETE string = "delete"
)

func init() {
	//if dbClient := database.Connect(databaseAddress); dbClient == nil {
	//	fmt.Println("cannot connect to database -", databaseAddress)
	//	return
	//}
	//fmt.Println("Connected to database -", databaseAddress)
}

type EventCb func(document Doc)
type Doc db.Entity
type DocStore interface {
	// publish

	Create(id string, data map[string]interface{}) string
	Update(id string, data map[string]interface{}) bool
	Get(id ...string) []Doc
	Delete(id ...string) bool

	// subscribe

	On(eventType string, do EventCb, ids ...string) (ok bool) // subscribe to all ids if nil
}

type store struct {
	table    string
	new      func(table, id string, data map[string]interface{}) db.NatsDoc
	dbClient database.Client
	*log.Logger
}

func (d Doc) Unmarshal(to interface{}) error {
	rec, _ := json.Marshal(d.Data)
	err := json.Unmarshal(rec, &to)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}
func newTable(table string) DocStore {

	if databaseAddress == "" {
		databaseAddress = "localhost:5000"
	}

	var dbClient database.Client
	if dbClient = database.Connect(databaseAddress); dbClient == nil {
		fmt.Println("cannot connect to database [" + databaseAddress + "]")
		return nil
	}
	if table == "" {
		table = defaultTable
	}
	documents.Add(table, newStore(table, dbClient))
	return documents.Get(table)
}

func newStore(table string, dbClient database.Client) *store {
	return &store{table: table, new: db.NewDocument, dbClient: dbClient,
		Logger: log.New(os.Stdout, "[store]["+table+"]", log.LstdFlags)}
}

func Get(table string) DocStore {
	if documents.Get(table) == nil {
		return newTable(table)
	}
	return documents.Get(table)
}
