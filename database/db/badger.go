package database

import (
	"encoding/json"

	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
)

func Init() (*badger.DB, func() error) {
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}
	log.Warn("...USING BADGER DB...")
	return db, db.Close
}
func delete(db *badger.DB, table string, entity *Entity) error {
	err := db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(table + "." + entity.Id))
		return err
	})
	return err
}
func set(db *badger.DB, table string, entity *Entity) error {
	b, _ := json.Marshal(entity)
	err := db.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(table+"."+entity.Id), b)
		err := txn.SetEntry(e)
		return err
	})
	return err
}
func get(db *badger.DB, table, key string) (val Entity, err error) {
	err = db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(table + "." + key))
		if err != nil {
			return err
		}
		if valB, err := item.ValueCopy(nil); err == nil {
			return json.Unmarshal(valB, &val)
		} else {
			return err
		}
	})
	return
}
