package FuncFw

import (
	"net/http"

	"github.com/Ishan27g/ryo-Faas/store"
)

var (
	applied = false
	Export  = funcFw{fns: make(map[string]*HttpFunction), s: nil} // entrypoint:fn
)

type HttpFn func(w http.ResponseWriter, r *http.Request)
type DatabaseEvent store.EventCb
type HttpFunction struct {
	Entrypoint string
	UrlPath    string
	HttpFn
}
type funcFw struct {
	fns map[string]*HttpFunction
	s   *StoreEvents
}

func (f funcFw) Events(s StoreEvents) {
	f.s = new(StoreEvents)
	f.s = &s
	s.Apply()
	applied = true
}
func (f funcFw) Http(entrypoint, url string, fn HttpFn) {
	f.fns[entrypoint] = &HttpFunction{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
}
func (f funcFw) Get() map[string]*HttpFunction {
	return f.fns
}
