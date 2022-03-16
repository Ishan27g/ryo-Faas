package database

import "time"

type Entity struct {
	Id        string    `json:"Id"`
	CreatedAt time.Time `json:"CreatedAt"`
	EditAt    time.Time `json:"EditAt"`
	Data      Data      `json:"Data"`
}

type Data struct {
	Value map[string]interface{}
}

//ID any struct that needs to persist should implement this function defined
//in Entity interface.
func (e Entity) ID() (jsonField string, value interface{}) {
	value = e.Id
	jsonField = "Id"
	return
}
