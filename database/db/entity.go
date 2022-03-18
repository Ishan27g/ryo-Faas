package database

import "time"

type Data struct {
	Value map[string]interface{} `json:"Value"`
}
type Entity struct {
	Id        string    `json:"Id"`
	CreatedAt time.Time `json:"CreatedAt"`
	EditedAt  time.Time `json:"EditedAt"`
	Data      Data      `json:"Data"`
}

//ID any struct that needs to persist should implement this function defined
//in Entity interface.
func (e Entity) ID() (jsonField string, value interface{}) {
	value = e.Id
	jsonField = "Id"
	return
}
