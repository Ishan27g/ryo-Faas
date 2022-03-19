package FuncFw

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/Ishan27g/ryo-Faas/transport"
)

func (hn HttpAsyncNats) SubscribeAsync(fn HttpFn) {
	go transport.NatsSubscribeJson(hn.getSubj(), func(msg *transport.AsyncNats) {
		ww := httptest.NewRecorder()
		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(msg.Req)))
		fn(ww, req)
		_, err = http.Post(msg.Callback, "application/json", ww.Result().Body)
		if err != nil {
			log.Println(err.Error())
		}
	})
}
func (hn HttpAsyncNats) getSubj() string {
	return transport.HttpAsync + "." + hn.entrypoint
}
func (hn HttpAsyncNats) HandleAsyncNats(w http.ResponseWriter, r *http.Request) {
	callback := r.Header.Get("X-Callback-Url")
	if callback == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing header, X-Callback-Url"))
		return
	}
	ha := NewAsyncNats(hn.entrypoint, callback)
	*ha.req = *r
	var b = &bytes.Buffer{}
	if err := ha.req.Write(b); err != nil {
		fmt.Println(" ha.req.Write", err.Error())
		return
	}

	if transport.NatsPublishJson(ha.getSubj(), transport.AsyncNats{
		Callback:   ha.callback,
		Entrypoint: ha.entrypoint,
		Req:        b.Bytes(),
	}, nil) {
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Ok"))
		return
	}
	w.WriteHeader(http.StatusExpectationFailed)
	w.Write([]byte("dunno"))
	return
}

func NewAsyncNats(entrypoint, callback string) HttpAsyncNats {
	return HttpAsyncNats{
		req:        new(http.Request),
		callback:   callback,
		entrypoint: entrypoint,
		//HttpFunction: HttpFunction{
		//	Entrypoint: entrypoint,
		//	UrlPath:    "",
		//	HttpFn:     httpFn,
		//},
	}
}
