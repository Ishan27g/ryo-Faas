package FuncFw

import "net/http"

var (
	Export = funcFw{fns: make(map[string]*HttpFunction)} // entrypoint:fn
)

type HttpFn func(w http.ResponseWriter, r *http.Request)

type HttpFunction struct {
	Entrypoint string
	UrlPath    string
	HttpFn
}
type funcFw struct {
	fns map[string]*HttpFunction
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
