package middleware

import (
	"net/http"

	"github.com/Ishan27g/noware/pkg/actions"
	"github.com/Ishan27g/noware/pkg/noop"
	"github.com/gin-gonic/gin"
)

// Http middleware extracts `noop` and `actions` from the request's header and adds to the request's context
func Http(n http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if noop.CheckHeader(req) {
			if actions.CheckHeader(req) {
				req = actions.Extract(req)
			}
			req = httpReqInjectNoop(req)
		}
		n.ServeHTTP(w, req)
	})
}

// Gin middleware extracts `noop` and `actions` from the request's header and adds to the request's context
func Gin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if noop.CheckHeader(c.Request) {
			if actions.CheckHeader(c.Request) {
				c.Request = actions.Extract(c.Request)
			}
			c.Request = httpReqInjectNoop(c.Request)
		}
	}
}

// adds noop ctx to request
func httpReqInjectNoop(req *http.Request) *http.Request {
	return req.Clone(noop.NewCtxWithNoop(req.Context()))
}
