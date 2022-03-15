package database

import (
	"sync"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27g/ryo-Faas/store"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/Ishan27g/ryo-Faas/types"
	"github.com/hashicorp/go-hclog"
	db "github.com/sonyarouje/simdb"
)

var databaseStore = dbStore{
	documents: types.NewMap(),
	driver:    nil,
	Mutex:     sync.Mutex{},
	Logger:    mLogger.Get("DATABASE"),
}

// Database publishes corresponding event to Nats after doing the DB operation
type Database interface {
	New(doc types.DocData)
	Update(doc types.DocData)
	Delete(id string)
	Get(id string) *types.DocData
	All() []*types.DocData
}

func GetDatabase() Database {
	return &databaseStore
}

type dbStore struct {
	documents types.SyncMap
	driver    *db.Driver
	sync.Mutex
	hclog.Logger
}

func init() {
	var err error
	databaseStore.driver, err = db.New("data")
	if err != nil {
		panic(err)
	}
}

var publish = func(subjId, data string) {
	transport.NatsPublish(subjId, data, nil)
}

func toEntity(doc types.DocData) Entity {
	return Entity{
		Id:        doc.Id(),
		CreatedAt: time.Time{},
		EditAt:    time.Time{},
		Data:      Data{Value: doc.DataJson()},
	}
}
func (d *dbStore) New(doc types.DocData) {
	d.Lock()
	defer d.Unlock()
	document := toEntity(doc)
	document.CreatedAt = time.Now()
	document.EditAt = time.Now()
	err := databaseStore.driver.Insert(document)
	if err != nil {
		d.Logger.Error("driver.Insert", "id", document.Id)
	}
	d.documents.Add(doc.Id(), doc.Id()) // value unused
	defer publish(store.DocumentCREATE+doc.Id(), doc.Data())
}

func (d *dbStore) Update(doc types.DocData) {
	d.Lock()
	defer d.Unlock()
	document := toEntity(doc)
	document.EditAt = time.Now()
	err := databaseStore.driver.Update(document)
	if err != nil {
		d.Logger.Error("driver.Update", "id", document.Id)
	}
	d.documents.Add(doc.Id(), doc.Id()) // value unused
	defer publish(store.DocumentUPDATE+doc.Id(), doc.Data())

}

func (d *dbStore) Delete(id string) {
	d.Lock()
	defer d.Unlock()
	err := databaseStore.driver.Delete(Entity{Id: id})
	if err != nil {
		d.Logger.Error("driver.Delete", "id", id)
	}
	defer publish(store.DocumentDELETE+id, "deleted")
	d.documents.Delete(id)
}

func (d *dbStore) get(id string) types.DocData {
	var entity Entity
	var document types.DocData

	err := d.driver.Open(Entity{}).Where("Id", "=", id).First().AsEntity(&entity)
	if err != nil {
		panic(err)
	}
	defer publish(store.DocumentGET+id, "deleted")

	document = types.NewDocData(entity.Id, entity.Data.Value)
	return document
}

func (d *dbStore) Get(id string) *types.DocData {
	d.Lock()
	defer d.Unlock()
	document := d.get(id)
	return &document
}

func (d *dbStore) All() []*types.DocData {
	d.Lock()
	defer d.Unlock()
	var documents []*types.DocData
	for id, _ := range d.documents.All() {
		doc := d.get(id)
		documents = append(documents, &doc)
	}
	return documents
}

type Data struct {
	Value map[string]interface{}
}
type Entity struct {
	Id        string    `json:"Id"`
	CreatedAt time.Time `json:"CreatedAt"`
	EditAt    time.Time `json:"EditAt"`
	Data      Data      `json:"Data"`
}

//ID any struct that needs to persist should implement this function defined
//in Entity interface.
func (e Entity) ID() (jsonField string, value interface{}) {
	value = e.Id
	jsonField = "Id"
	return
}
