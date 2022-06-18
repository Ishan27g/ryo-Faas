package actions

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

const actionKey string = "action-key"

type Actions struct {
	Events chan []interface{}
}
type Event struct {
	Name string      `json:"name"`
	Meta interface{} `json:"meta,omitempty"`
}

// FromCtx returns actions if present in the context or nil
func FromCtx(ctx context.Context) *Actions {
	if ctx == nil {
		return nil
	}
	if a := ctx.Value(actionKey); a != nil {
		return a.(*Actions)
	}
	return nil
}

// NewCtx returns a context with actions
func NewCtx(ctx context.Context, actions *Actions) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if actions == nil || actions.Events == nil {
		actions = New()
	}
	return context.WithValue(ctx, actionKey, actions)
}

func New() *Actions {
	a := &Actions{Events: make(chan []interface{}, 1)}
	a.Events <- make([]interface{}, 0)
	return a
}
func (a *Actions) GetEvents() []Event {
	if a == nil || a.Events == nil {
		return nil
	}
	events := <-a.Events
	defer func() { a.Events <- events }()
	var newEvents []Event
	for _, event := range events {
		newEvents = append(newEvents, event.(Event))
	}
	return newEvents
}
func (a *Actions) AddEvent(events ...Event) {
	// don't add if not present
	if a == nil || a.Events == nil {
		return
	}
	e := <-a.Events
	var newEvents []interface{}
	for _, event := range events {
		newEvents = append(newEvents, event)
	}
	a.Events <- append(e, newEvents...)
}

func (a *Actions) Marshal() ([]byte, error) {
	if a.Events == nil {
		return nil, errors.New("nil events")
	}
	e := <-a.Events
	defer func() { a.Events <- e }()
	return json.Marshal(e)
}
func UnMarshal(by []byte) (Actions, error) {
	a := New()
	var events []Event
	var b interface{}
	err := json.Unmarshal(by, &b)
	if err != nil {
		fmt.Println(err.Error())
		return *a, nil
	}
	j, _ := json.Marshal(b)
	_ = json.Unmarshal(j, &events)
	for _, event := range events {
		a.AddEvent(event)
	}
	return *a, nil
}
