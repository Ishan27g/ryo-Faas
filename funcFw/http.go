package FuncFw

import (
	"net/http"
)

type HttpFn func(w http.ResponseWriter, r *http.Request)

func (h HttpFunction) wrap() http.HandlerFunc {
	return http.HandlerFunc(h.HttpFn)
}

type HttpFunction struct {
	Entrypoint string
	UrlPath    string
	HttpFn
}
type HttpAsyncNats struct {
	callback   string
	entrypoint string
	req        *http.Request
	// HttpFunction
}

type HttpAsync HttpFunction

type funcFw struct {
	httpFns       map[string]*HttpFunction
	httpAsync     map[string]*HttpAsync
	httpAsyncNats map[string]*HttpAsync
	storeEvents   map[string]StoreEventsI
}

func (f *funcFw) Http(entrypoint, url string, fn HttpFn) {
	f.httpFns[entrypoint] = &HttpFunction{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
}
func (f *funcFw) GetHttp() map[string]*HttpFunction {
	return f.httpFns
}
func (f *funcFw) HttpAsync(entrypoint, url string, fn HttpFn) {
	f.httpAsync[entrypoint] = &HttpAsync{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
}
func (f *funcFw) GetHttpAsync() map[string]*HttpAsync {
	return f.httpAsync
}
func (f *funcFw) NatsAsync(entrypoint string, url string, fn HttpFn) {
	f.httpAsyncNats[entrypoint] = &HttpAsync{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
}
func (f *funcFw) GetHttpAsyncNats() map[string]*HttpAsync {
	return f.httpAsyncNats
}
