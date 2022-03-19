package FuncFw

import (
	"net/http"

	"github.com/Ishan27g/ryo-Faas/store"
)

var (
	Export = funcFw{
		httpFns:     make(map[string]*HttpFunction),
		httpAsync:   make(map[string]*HttpAsync),
		storeEvents: nil,
	} // entrypoint:fn
)

type HttpFn func(w http.ResponseWriter, r *http.Request)

type HttpAsync HttpFunction
type DatabaseEvent store.EventCb
type HttpFunction struct {
	Entrypoint string
	UrlPath    string
	HttpFn
}
type funcFw struct {
	httpFns     map[string]*HttpFunction
	httpAsync   map[string]*HttpAsync
	storeEvents map[string]StoreEventsI
}

func (f funcFw) Http(entrypoint, url string, fn HttpFn) {
	f.httpFns[entrypoint] = &HttpFunction{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
}
func (f funcFw) GetHttp() map[string]*HttpFunction {
	return f.httpFns
}
func (f funcFw) GetHttpAsync() map[string]*HttpAsync {
	return f.httpAsync
}
func (f funcFw) HttpAsync(entrypoint, url string, fn HttpFn) {
	f.httpAsync[entrypoint] = &HttpAsync{
		Entrypoint: entrypoint,
		UrlPath:    url,
		HttpFn:     fn,
	}
	//go func() {
	//	transport.NatsSubscribeJson(transport.HttpAsync+"."+entrypoint, func(msg interface{}) {
	//		asyncFn := msg.(HttpAsync)
	//
	//		var w http.ResponseWriter
	//		var r io.ReadCloser
	//
	//		asyncFn.HttpFn(w, asyncFn.r)
	//		_, err := io.Copy(w, r)
	//		if err != nil {
	//			log.Println(err.Error())
	//			return
	//		}
	//
	//		_, err = http.Post(asyncFn.callback, "application/json", r)
	//		if err != nil {
	//			log.Println(err.Error())
	//		}
	//
	//	})
	//}()

}

//
//func HandleAsync(entrypoint string, fn HttpFn, w http.ResponseWriter, r *http.Request) {
//	callback := r.Header.Get("X-Callback-Url")
//	if callback == "" {
//		w.WriteHeader(http.StatusBadRequest)
//		w.Write([]byte("missing header, X-Callback-Url"))
//		return
//	}
//	req := *r
//	ha := HttpAsync{
//		callback: callback,
//		r:        &req,
//		HttpFn:   fn,
//	}
//	if transport.NatsPublishJson(transport.HttpAsync+"."+entrypoint, &ha, nil) {
//		w.WriteHeader(http.StatusAccepted)
//		w.Write([]byte("Ok"))
//		return
//	}
//	w.WriteHeader(http.StatusExpectationFailed)
//	w.Write([]byte("dunno"))
//	return
//}
