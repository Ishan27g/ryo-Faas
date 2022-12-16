package database

type Entity struct {
	Id        string                 `json:"Id"`
	CreatedAt string                 `json:"CreatedAt"`
	EditedAt  string                 `json:"EditedAt"`
	Data      map[string]interface{} `json:"Data"`
}
