package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/Ishan27g/go-utils/mLogger"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"github.com/hashicorp/go-hclog"
	"github.com/patrickmn/go-cache"
	db "github.com/sonyarouje/simdb"
)

const TFormat = time.RFC850

var databaseStore = dbStore{
	documents: types.NewMap(),              // entityId:createdAt
	driver:    make(map[string]*db.Driver), // database driver
	Mutex:     sync.Mutex{},
	Logger:    mLogger.Get("DATABASE"),
	cache:     cache.New(1*time.Minute, 5*time.Minute),
}

type Database interface {
	New(doc NatsDoc)
	Update(doc NatsDoc)
	Delete(id string)
	Get(id string) *Entity
	All() []*Entity
	After(fromTime string) []*Entity
}

type dbStore struct {
	documents types.SyncMap
	driver    map[string]*db.Driver
	cache     *cache.Cache
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
	// databaseStore.driver, err = db.New("data")
	if err != nil {
		panic(err)
	}

}

func GetDatabase() Database {
	return &databaseStore
}

func toEntity(doc NatsDoc) Entity {
	return Entity{
		Id:        doc.Id(),
		CreatedAt: "",
		EditedAt:  "",
		Data:      Data{Value: doc.Document()},
	}
}
func (d *dbStore) getDriver(table string) *db.Driver {
	if d.driver[table] == nil {
		if driver, err := db.New(table); err == nil {
			d.driver[table] = driver
		}
	}
	return d.driver[table]
}
func (d *dbStore) New(doc NatsDoc) {
	d.Lock()
	defer d.Unlock()

	entity := toEntity(doc)
	entity.CreatedAt = time.Now().String()
	entity.EditedAt = time.Now().String()

	err := d.getDriver(doc.Table()).Insert(entity)
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

	err := d.getDriver(doc.Table()).Update(entity)
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
	err := d.getDriver(tableName.(string)).Delete(Entity{Id: id})
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
	if tableName != nil {
		fmt.Println("Table found for ", id)
		err := d.getDriver(tableName.(string)).Open(Entity{}).Where("Id", "=", id).First().AsEntity(&entity)
		if err != nil {
			panic(err)
		}
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

func (d *dbStore) After(fromTime string) []*Entity {
	d.Lock()
	defer d.Unlock()
	var documents []*Entity

	from := parse(fromTime)
	for id, at := range d.documents.All() {
		createdAt := format(at.(time.Time))
		if createdAt.Before(from) {
			doc := d.get(id)
			documents = append(documents, &doc)
		}
	}
	return documents
}
