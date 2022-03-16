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
	documents: types.NewMap(),
	driver:    nil,
	Mutex:     sync.Mutex{},
	Logger:    mLogger.Get("DATABASE"),
}

// Simple Json Database over grpc
type Database interface {
	New(doc types.NatsDoc)
	Update(doc types.NatsDoc)
	Delete(id string)
	Get(id string) *types.NatsDoc
	All() []*types.NatsDoc
	After(fromTime string) []*types.NatsDoc
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
		EditAt:    time.Time{},
		Data:      Data{Value: doc.Document()},
	}
}

func (d *dbStore) New(doc types.NatsDoc) {
	d.Lock()
	defer d.Unlock()

	document := toEntity(doc)
	document.CreatedAt = time.Now()
	document.EditAt = time.Now()

	err := d.driver.Insert(document)
	if err != nil {
		d.Logger.Error("driver.Insert", "id", document.Id)
	}
	d.documents.Add(doc.Id(), format(document.CreatedAt)) // value=createAttime
}

func (d *dbStore) Update(document types.NatsDoc) {
	d.Lock()
	defer d.Unlock()

	docu := toEntity(document)
	docu.EditAt = time.Now()

	err := databaseStore.driver.Update(docu)
	if err != nil {
		d.Logger.Error("driver.Update", "id", docu.Id)
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

func (d *dbStore) get(id string) types.NatsDoc {
	var entity Entity
	var document types.NatsDoc

	err := d.driver.Open(Entity{}).Where("Id", "=", id).First().AsEntity(&entity)
	if err != nil {
		panic(err)
	}
	document = types.NewNatsDoc(entity.Id, entity.Data.Value)
	return document
}

func (d *dbStore) Get(id string) *types.NatsDoc {
	d.Lock()
	defer d.Unlock()
	document := d.get(id)
	return &document
}

func (d *dbStore) All() []*types.NatsDoc {
	d.Lock()
	defer d.Unlock()
	var documents []*types.NatsDoc
	for id := range d.documents.All() {
		doc := d.get(id)
		documents = append(documents, &doc)
	}
	return documents
}

func (d *dbStore) After(fromTime string) []*types.NatsDoc {
	d.Lock()
	defer d.Unlock()
	from := parse(fromTime)

	var documents []*types.NatsDoc
	for id, at := range d.documents.All() {
		createdAt := format(at.(time.Time))
		if createdAt.Before(from) {
			doc := d.get(id)
			documents = append(documents, &doc)
		}
	}
	return documents
}
