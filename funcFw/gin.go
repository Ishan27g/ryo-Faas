package FuncFw

import "github.com/gin-gonic/gin"

type HttpFnGin func(ctx *gin.Context)
type httpFnGin struct {
	httpFunction
	gf gin.HandlerFunc
}
