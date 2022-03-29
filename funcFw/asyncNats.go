package FuncFw

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func NewAsyncNats(entrypoint, callback string) HttpAsyncNats {
	return HttpAsyncNats{
		req:        new(http.Request),
		callback:   callback,
		entrypoint: entrypoint,
	}
}

func (hn HttpAsyncNats) getSubj() string {
	return transport.HttpAsync + "." + strings.ToLower(hn.entrypoint)
}

func (hn HttpAsyncNats) HandleAsyncNats(w http.ResponseWriter, r *http.Request) {
	callback := r.Header.Get(tracing.XCallback)
	span := trace.SpanFromContext(r.Context())
	span.SetAttributes(attribute.Key(tracing.XCallback).String(callback))
	span.SetAttributes(attribute.Key(tracing.Pub).String(hn.entrypoint))

	ha := NewAsyncNats(hn.entrypoint, callback)
	*ha.req = *r
	var b = &bytes.Buffer{}
	if err := ha.req.Write(b); err != nil {
		span.SetAttributes(attribute.Key(tracing.Error).String(tracing.ErrorWrite))
		fmt.Println("ha.req.Write", err.Error())
		return
	}
	span.SetAttributes(attribute.Key(tracing.PubAt).String(time.Now().String()))

	published := transport.NatsPublishJson(ha.getSubj(), transport.AsyncNats{
		Callback:   ha.callback,
		Entrypoint: ha.entrypoint,
		Req:        b.Bytes(),
		Header:     tracing.ExtractHeader(span),
	}, nil)

	if published {
		span.SetAttributes(attribute.Key(tracing.Success).String(tracing.SuccessPub))
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Ok " + callback + "\n"))
		return
	}

	span.SetAttributes(attribute.Key(tracing.Error).String(tracing.ErrorPub))
	w.WriteHeader(http.StatusExpectationFailed)
	w.Write([]byte("dunno"))
}

func (hn HttpAsyncNats) SubscribeAsync(fn HttpFn) {
	go transport.NatsSubscribeJson(hn.getSubj(), func(msg *transport.AsyncNats) {
		// register next subscription
		go hn.SubscribeAsync(fn)
		
		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(msg.Req)))

		tracer := provider.Get().Tracer("function-async-receiver")
		_, span := tracer.Start(tracing.ExtractSpan(msg.Header), "proxy-function-call"+"-"+hn.entrypoint)
		defer span.End()
		span.SetAttributes(attribute.Key(tracing.Sub).String(hn.entrypoint))
		span.SetAttributes(attribute.Key(tracing.SubAt).String(time.Now().String()))

		ww := httptest.NewRecorder()
		fn(ww, req)

		span.SetAttributes(attribute.Key(tracing.SubEnd).String(time.Now().String()))
		_, err = http.Post(msg.Callback, "application/json", ww.Result().Body)
		if err != nil {
			log.Println(err.Error())
		}
	})
}
