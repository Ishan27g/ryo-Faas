package noop

import (
	"context"
	"net/http"
)

type keyType string

const noopKey keyType = "noop-key"

func ContainsNoop(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if v := ctx.Value(noopKey); v != nil {
		return v.(bool)
	}
	return false
}

func NewCtxWithNoop(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if ContainsNoop(ctx) { // todo needed?
		return ctx
	}
	return context.WithValue(ctx, noopKey, true)
}
func AddHeader(req *http.Request) *http.Request {
	req.Header.Add(string(noopKey), "true")
	return req
}
func CheckHeader(req *http.Request) bool {
	isNoop := req.Header.Get(string(noopKey))
	return isNoop == "true"
}
