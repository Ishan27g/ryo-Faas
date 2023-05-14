package FuncFw

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HttpFn http.HandlerFunc
type httpFunction struct {
	Entrypoint string
	UrlPath    string
	HttpFn
}
type HttpAsync httpFunction
type HttpAsyncNats struct {
	callback   string
	entrypoint string
	req        *http.Request
}

type funcFw struct {
	httpFns       map[string]*httpFunction
	httpFnsGin    map[string]*httpFnGin
	httpAsync     map[string]*HttpAsync
	httpAsyncNats map[string]*HttpAsync
	storeEvents   map[string]StoreEventsI
	funcCtx
}

func wrap(f HttpFn) gin.HandlerFunc {
	return gin.WrapH(http.HandlerFunc(f))
}
func (f httpFunction) AsGin() gin.HandlerFunc {
	return wrap(f.HttpFn)
}

func (f HttpAsync) AsGin() gin.HandlerFunc {
	return wrap(wrapAsync(f))
}

func (f *funcFw) EventsFor(tableName string) StoreEventsI {
	if f.storeEvents == nil {
		f.storeEvents = make(map[string]StoreEventsI)
	}
	if f.storeEvents[tableName] == nil {
		f.storeEvents[tableName] = &storeEvents{
			Table: tableName,
			on:    nil,
		}
	}
	return f.storeEvents[tableName]
}

func (f *funcFw) Http(entrypoint, url string, fn HttpFn) {
	f.httpFns[entrypoint] = &httpFunction{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
}
func (f *funcFw) getHttp() map[string]*httpFunction {
	return f.httpFns
}
func (f *funcFw) HttpAsync(entrypoint, url string, fn HttpFn) {
	f.httpAsync[entrypoint] = &HttpAsync{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
}
func (f *funcFw) getHttpAsync() map[string]*HttpAsync {
	return f.httpAsync
}
func (f *funcFw) NatsAsync(entrypoint string, url string, fn HttpFn) {
	f.httpAsyncNats[entrypoint] = &HttpAsync{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
}
func (f *funcFw) getHttpAsyncNats() map[string]*HttpAsync {
	return f.httpAsyncNats
}
func (f *funcFw) HttpGin(entrypoint, url string, fn HttpFnGin) {
	f.httpFnsGin[entrypoint] = &httpFnGin{
		httpFunction: httpFunction{
			Entrypoint: entrypoint,
			UrlPath:    url,
		},
		gf: gin.HandlerFunc(fn),
	}
}
func (f *funcFw) getHttpGin() map[string]*httpFnGin {
	return f.httpFnsGin
}
