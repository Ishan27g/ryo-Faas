package noop

import (
	"net/http"

	"github.com/Ishan27g/noware/pkg/actions"
)

// HttpRequest returns a clone of the passed request with `noop`
// into the request's header if the context is `noop`. If `noop`, then `actions` are
// injected is context has `actions`
func HttpRequest(req *http.Request) *http.Request {
	if ContainsNoop(req.Context()) {
		return actions.Inject(AddHeader(req))
	}
	return req
}
