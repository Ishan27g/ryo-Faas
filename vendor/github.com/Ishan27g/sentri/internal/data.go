package internal

import "time"

const (
	Channel = "sentri.*"
)

type Output struct {
	From      string
	Timestamp time.Time
	Log       string // []byte
}
