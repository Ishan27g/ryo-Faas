package FuncFw

import "github.com/gin-gonic/gin"

type HttpFnGin func(ctx *gin.Context)
type httpFnGin struct {
	httpFunction
	gf gin.HandlerFunc
}

type funcCtx struct{ any }

func newCtx[T any](val T) funcCtx { return funcCtx{val} }

// InjectCtx will set funcCtx to provided val
// call on init & use pointers
func InjectCtx[T any](val T) {
	Export.funcCtx = newCtx[T](val)
}

// ExtractCtx will return funcCtx
func ExtractCtx[T any]() T {
	return Export.funcCtx.any.(T)
}
