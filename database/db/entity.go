package database

type Data struct {
	Value map[string]interface{} `json:"Value"`
}
type Entity struct {
	Id        string `json:"Id"`
	CreatedAt string `json:"CreatedAt"`
	EditedAt  string `json:"EditedAt"`
	Data      Data   `json:"Data"`
}

func (e Entity) ID() (jsonField string, value interface{}) {
	value = e.Id
	jsonField = "Id"
	return
}
