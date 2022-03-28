package FuncFw

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
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
	callback := r.Header.Get("X-Callback-Url")
	if callback == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("missing header, X-Callback-Url"))
		return
	}

	span := trace.SpanFromContext(r.Context())
	span.SetAttributes(attribute.Key("X-Callback-Url").String(callback))

	ha := NewAsyncNats(hn.entrypoint, callback)
	*ha.req = *r
	var b = &bytes.Buffer{}
	if err := ha.req.Write(b); err != nil {
		span.SetAttributes(attribute.Key("Request write error").String(err.Error()))
		fmt.Println("ha.req.Write", err.Error())
		return
	}
	span.SetAttributes(attribute.Key("nats-publish at").String(time.Now().String()))

	var makeSpanHeader = func(span trace.Span) (header nats.Header) {
		header = make(nats.Header)
		getTraceID := span.SpanContext().TraceID().String()
		header.Set("traceid", getTraceID)
		getSpanID := span.SpanContext().SpanID().String()
		header.Set("spanid", getSpanID)
		return
	}

	published := transport.NatsPublishJson(ha.getSubj(), transport.AsyncNats{
		Callback:   ha.callback,
		Entrypoint: ha.entrypoint,
		Req:        b.Bytes(),
		Header:     makeSpanHeader(span),
	}, nil)

	if published {
		span.SetAttributes(attribute.Key("nats-publish").String(ha.entrypoint))
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Ok"))
		return
	}

	span.SetAttributes(attribute.Key("Error").String("publishing to nats"))
	w.WriteHeader(http.StatusExpectationFailed)
	w.Write([]byte("dunno"))
}

func (hn HttpAsyncNats) SubscribeAsync(fn HttpFn) {
	go transport.NatsSubscribeJson(hn.getSubj(), func(msg *transport.AsyncNats) {
		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(msg.Req)))

		_, span := otel.Tracer("function-async-receiver").Start(extractSpan(msg, err),
			"proxy-function-call"+"-"+hn.entrypoint)
		defer span.End()
		span.SetAttributes(attribute.Key("nats-subscribe").String(hn.entrypoint))
		span.SetAttributes(attribute.Key("nats-subscribe begin-at").String(time.Now().String()))

		ww := httptest.NewRecorder()
		fn(ww, req)

		span.SetAttributes(attribute.Key("nats-subscribe end-at").String(time.Now().String()))
		_, err = http.Post(msg.Callback, "application/json", ww.Result().Body)
		if err != nil {
			log.Println(err.Error())
		}
	})
}

// https://stackoverflow.com/questions/70378025/how-to-create-opentelemetry-span-from-a-string-traceid
func createSpanContext(traceID trace.TraceID, spanID trace.SpanID) context.Context {
	var spanContextConfig trace.SpanContextConfig
	spanContextConfig.TraceID = traceID
	spanContextConfig.SpanID = spanID
	spanContextConfig.TraceFlags = 01
	spanContextConfig.Remote = false
	return trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(spanContextConfig))
}
func extractSpan(msg *transport.AsyncNats, err error) context.Context {
	traceID, err := trace.TraceIDFromHex(msg.Header.Get("traceid"))
	if err != nil {
		fmt.Println("no trace Id in header: ", err)
		fmt.Println("expected otel middleware to inject span into original request")
	}
	spanID, err := trace.SpanIDFromHex(msg.Header.Get("spanid"))
	if err != nil {
		fmt.Println("no span Id in header: ", err)
		fmt.Println("expected otel middleware to inject span into original request")
	}
	return createSpanContext(traceID, spanID)
}
