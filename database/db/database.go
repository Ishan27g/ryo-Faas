package database

import (
	"sync"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27g/ryo-Faas/types"
	"github.com/hashicorp/go-hclog"
	db "github.com/sonyarouje/simdb"
)

const TFormat = time.RFC850

var databaseStore = dbStore{
	documents: types.NewMap(), // entityId:createdAt
	driver:    nil,            // database driver
	Mutex:     sync.Mutex{},
	Logger:    mLogger.Get("DATABASE"),
}

// Simple Json Database over grpc
type Database interface {
	New(doc types.NatsDoc)
	Update(doc types.NatsDoc)
	Delete(id string)
	Get(id string) *Entity
	All() []*Entity
	After(fromTime string) []*Entity
}

type dbStore struct {
	documents types.SyncMap
	driver    *db.Driver
	sync.Mutex
	hclog.Logger
}

func parse(when string) time.Time {
	t, _ := time.Parse(TFormat, when)
	return t
}
func format(when time.Time) time.Time {
	t, _ := time.Parse(TFormat, when.String())
	return t
}
func init() {
	var err error
	databaseStore.driver, err = db.New("data")
	if err != nil {
		panic(err)
	}

}

func GetDatabase() Database {
	return &databaseStore
}

func toEntity(doc types.NatsDoc) Entity {
	return Entity{
		Id:        doc.Id(),
		CreatedAt: time.Time{},
		EditedAt:  time.Time{},
		Data:      Data{Value: doc.Document()},
	}
}

func (d *dbStore) New(doc types.NatsDoc) {
	d.Lock()
	defer d.Unlock()

	entity := toEntity(doc)
	entity.CreatedAt = time.Now()
	entity.EditedAt = time.Now()

	err := d.driver.Insert(entity)
	if err != nil {
		d.Logger.Error("driver.Insert", "id", entity.Id)
	}
	d.documents.Add(doc.Id(), format(entity.CreatedAt)) // value=createAttime
}

func (d *dbStore) Update(document types.NatsDoc) {

	var existing *Entity
	if existing = d.Get(document.Id()); existing == nil {
		d.Logger.Error("driver.Update - not found", "id", document.Id())
		return
	}

	d.Lock()
	defer d.Unlock()

	entity := toEntity(document)
	entity.EditedAt = time.Now()
	entity.CreatedAt = existing.CreatedAt

	err := databaseStore.driver.Update(entity)
	if err != nil {
		d.Logger.Error("driver.Update", "id", entity.Id, "err", err.Error())
	}
	// no need to update d.document
}

func (d *dbStore) Delete(id string) {
	d.Lock()
	defer d.Unlock()
	err := databaseStore.driver.Delete(Entity{Id: id})
	if err != nil {
		d.Logger.Error("driver.Delete", "id", id)
	}
	d.documents.Delete(id)
}

func (d *dbStore) get(id string) Entity {
	var entity Entity
	err := d.driver.Open(Entity{}).Where("Id", "=", id).First().AsEntity(&entity)
	if err != nil {
		panic(err)
	}
	return entity
}

func (d *dbStore) Get(id string) *Entity {
	d.Lock()
	defer d.Unlock()
	document := d.get(id)
	return &document
}

func (d *dbStore) All() []*Entity {
	d.Lock()
	defer d.Unlock()
	var documents []*Entity
	for id := range d.documents.All() {
		doc := d.get(id)
		documents = append(documents, &doc)
	}
	return documents
}

func (d *dbStore) After(fromTime string) []*Entity {
	d.Lock()
	defer d.Unlock()
	from := parse(fromTime)

	var documents []*Entity
	for id, at := range d.documents.All() {
		createdAt := format(at.(time.Time))
		if createdAt.Before(from) {
			doc := d.get(id)
			documents = append(documents, &doc)
		}
	}
	return documents
}
