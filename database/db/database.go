package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/go-hclog"
	"github.com/patrickmn/go-cache"
)

const TFormat = time.RFC850

var databaseStore = dbStore{
	documents:    types.NewMap[string, string](), // entityId:createdAt
	badgerDriver: nil,                            // database driver
	Mutex:        sync.Mutex{},
	Logger:       mLogger.Get("DATABASE"),
	cache:        cache.New(1*time.Minute, 5*time.Minute),
}

type Database interface {
	New(doc NatsDoc)
	Update(doc NatsDoc)
	Delete(id string)
	Get(id string) *Entity
	All() []*Entity
}
type badgerDriver struct {
	badgerDb *badger.DB
	close    func() error
}
type dbStore struct {
	documents    types.SyncMap[string, string]
	badgerDriver *badgerDriver
	cache        *cache.Cache
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
	if err != nil {
		panic(err)
	}

}

func GetDatabase() Database { return &databaseStore }

func toEntity(doc NatsDoc) Entity {
	return Entity{
		Id:        doc.Id(),
		CreatedAt: "",
		EditedAt:  "",
		Data:      doc.Document(),
	}
}
func (d *dbStore) getBadgerDriver() *badgerDriver {
	if d.badgerDriver == nil {
		driver, close := Init()
		d.badgerDriver = &badgerDriver{
			badgerDb: driver,
			close:    close,
		}
	}
	return d.badgerDriver
}
func (d *dbStore) New(doc NatsDoc) {
	d.Lock()
	defer d.Unlock()

	entity := toEntity(doc)
	entity.CreatedAt = time.Now().String()
	entity.EditedAt = time.Now().String()

	err := set(d.getBadgerDriver().badgerDb, doc.Table(), &entity)
	if err != nil {
		d.Logger.Error("driver.Insert", "id", entity.Id)
	}
	d.documents.Add(doc.Id(), doc.Table()) // value=createAttime
	if _, found := d.cache.Get(doc.Id()); found {
		d.cache.Delete(doc.Id())
	}
	d.cache.Set(doc.Id(), entity, cache.DefaultExpiration)
	fmt.Println("Added", doc.Id(), " to", doc.Table())
}

func (d *dbStore) Update(doc NatsDoc) {

	var existing *Entity
	if existing = d.Get(doc.Id()); existing == nil {
		d.Logger.Error("driver.Update - not found", "id", doc.Id())
		return
	}
	d.Lock()
	defer d.Unlock()

	entity := toEntity(doc)
	entity.EditedAt = time.Now().String()
	entity.CreatedAt = existing.CreatedAt

	err := set(d.getBadgerDriver().badgerDb, doc.Table(), &entity)
	if err != nil {
		d.Logger.Error("driver.Update", "id", entity.Id, "err", err.Error())
	}
	if _, found := d.cache.Get(doc.Id()); found {
		d.cache.Delete(doc.Id())
		d.cache.Set(doc.Id(), entity, cache.DefaultExpiration)
	}
	// no need to update d.doc
}

func (d *dbStore) Delete(id string) {
	d.Lock()
	defer d.Unlock()
	tableName := d.documents.Get(id)
	fmt.Println("Deleting ", id, " from ", tableName)
	err := delete(d.getBadgerDriver().badgerDb, tableName, &Entity{Id: id})
	if err != nil {
		d.Logger.Error("driver.Delete", "id", id)
	}
	d.documents.Delete(id)
	d.cache.Delete(id)
}

func (d *dbStore) get(id string) Entity {
	if e, found := d.cache.Get(id); found {
		return e.(Entity)
	}
	var entity Entity
	tableName := d.documents.Get(id)
	if tableName != "" {
		fmt.Println("Table found for ", id)
		e, err := get(d.getBadgerDriver().badgerDb, tableName, id)
		if err != nil {
			panic(err)
		}
		entity = e
	} else {
		fmt.Println("Table not found for ", id)
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
