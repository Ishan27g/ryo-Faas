package actions

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func CheckHeader(req *http.Request) bool {
	hasActions := req.Header.Get(actionKey)
	return hasActions != ""
}

// Inject adds actions to request's header present in request's context
func Inject(req *http.Request) *http.Request {
	var a *Actions
	if a = FromCtx(req.Context()); a == nil {
		return req
	}
	for _, event := range a.GetEvents() {
		e, _ := json.Marshal(event)
		req = addHeader(req, string(e))
	}
	return req
}

// Extract actions to request's context if present in request's header
func Extract(req *http.Request) *http.Request {
	a := fromHeader(req.Header)
	if a == nil {
		return req
	}
	return req.Clone(NewCtx(context.Background(), a))
}

func addHeader(req *http.Request, val string) *http.Request {
	req.Header.Add(actionKey, val)
	return req
}

func fromHeader(header http.Header) *Actions {
	var a = New()
	for _, ev := range header.Values(actionKey) {
		var e = Event{}
		err := json.Unmarshal([]byte(ev), &e)
		if err != nil {
			fmt.Println(err.Error())
		}
		a.AddEvent(e)
	}
	return a
}
