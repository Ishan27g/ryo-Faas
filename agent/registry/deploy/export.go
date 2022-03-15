package deploy

import "net/http"

type HttpFn func(w http.ResponseWriter, r *http.Request)

type HttpFunction struct {
	Entrypoint string
	UrlPath    string
	HttpFn
}

var (
	exported = map[string]HttpFunction{} // entrypoint:fn
)

func init() {
	exported = make(map[string]HttpFunction)
}
func Export(entrypoint, url string, fn HttpFn) {
	exported[entrypoint] = HttpFunction{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
}
func Get() map[string]HttpFunction {
	return exported
}
