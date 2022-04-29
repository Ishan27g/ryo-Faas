package noop

import (
	"context"
	"encoding/json"
	"errors"
)

const actionKey keyType = "action_key"

type Json struct {
	Events []interface{} `json:"Events"`
}
type Actions struct {
	Events chan []interface{}
}
type Event struct {
	Name        string      `json:"name"` // ",omitempty"`
	NextSubject string      `json:"nextSubject"`
	Meta        interface{} `json:"meta,omitempty"`
}

func ActionsFromCtx(ctx context.Context) *Actions {
	if ctx == nil {
		return nil
	}
	// return nil or actions if present
	if a := ctx.Value(actionKey); a != nil {
		return a.(*Actions)
	}
	return nil
}
func NewCtxWithActions(ctx context.Context, actions *Actions) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if actions == nil || actions.Events == nil {
		actions = NewActions()
	}
	return context.WithValue(ctx, actionKey, actions)
}

func NewActions() *Actions {
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
		// fmt.Println("actions or Events is nil")
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
	return json.Marshal(Json{Events: e})
}
