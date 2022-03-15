package store

import "github.com/Ishan27g/ryo-Faas/types"

var DatabaseStore = dbStore{store{documents: types.NewMap(), new: types.NewDocData}} // todo as db (redis?)
type dbStore struct {
	store // todo
}

func GetDatabase() *dbStore {
	return &DatabaseStore
}

// todo database
func (d *dbStore) all() []types.DocData {
	var docs []types.DocData
	for _, d := range d.documents.All() {
		docs = append(docs, d.(types.DocData))
	}
	return docs
}
