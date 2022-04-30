package noop

import (
	"context"
)

type keyType string

const noopKey keyType = "noop_key"

func ContainsNoop(ctx context.Context) bool {
	if ctx == nil {
		return false
	}
	if v := ctx.Value(noopKey); v != nil {
		return v.(bool)
	}
	return false
}

func NewCtxWithNoop(ctx context.Context, isNoop bool) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if ContainsNoop(ctx) { // todo needed?
		return ctx
	}
	return context.WithValue(ctx, noopKey, isNoop)
}
