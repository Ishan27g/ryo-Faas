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

	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (hn HttpAsyncNats) SubscribeAsync(fn HttpFn) {
	go transport.NatsSubscribeJson(hn.getSubj(), func(msg *transport.AsyncNats) {

		ww := httptest.NewRecorder()

		req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(msg.Req)))

		var traceID trace.TraceID
		traceID, err = trace.TraceIDFromHex(msg.Header.Get("traceid"))
		if err != nil {
			fmt.Println("error: ", err)
		}
		var spanID trace.SpanID
		spanID, err = trace.SpanIDFromHex(msg.Header.Get("spanid"))
		if err != nil {
			fmt.Println("error: ", err)
		}
		var spanContextConfig trace.SpanContextConfig
		spanContextConfig.TraceID = traceID
		spanContextConfig.SpanID = spanID
		spanContextConfig.TraceFlags = 01
		spanContextConfig.Remote = false
		spanContext := trace.NewSpanContext(spanContextConfig)

		fmt.Println("IS VALID? ", spanContext.IsValid())
		requestContext := context.Background()
		requestContext = trace.ContextWithSpanContext(requestContext, spanContext)

		var span trace.Span
		_, span = otel.Tracer("function-async-receiver").Start(requestContext, "proxy-function-call"+"-"+hn.entrypoint)
		defer span.End()
		span.AddEvent("processing....") //

		//span := trace.SpanFromContext()
		span.SetAttributes(attribute.Key("nats-subscribe").String(hn.entrypoint))
		span.SetAttributes(attribute.Key("nats-subscribe begin-at").String(time.Now().String()))
		fn(ww, req)
		span.SetAttributes(attribute.Key("nats-subscribe end-at").String(time.Now().String()))
		_, err = http.Post(msg.Callback, "application/json", ww.Result().Body)
		if err != nil {
			log.Println(err.Error())
		}
	})
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

	msg := new(nats.Msg)
	getTraceID := span.SpanContext().TraceID().String()
	header := make(nats.Header)
	msg.Header = header
	header.Set("traceid", getTraceID)
	getSpanID := span.SpanContext().SpanID().String()
	header.Set("spanid", getSpanID)
	msg.Header = header

	ha := NewAsyncNats(hn.entrypoint, callback)
	*ha.req = *r
	var b = &bytes.Buffer{}
	if err := ha.req.Write(b); err != nil {
		span.SetAttributes(attribute.Key("Request write error").String(err.Error()))
		fmt.Println("ha.req.Write", err.Error())
		return
	}
	fmt.Println("publishing json", ha.getSubj())
	if transport.NatsPublishJson(ha.getSubj(), transport.AsyncNats{
		Callback:   ha.callback,
		Entrypoint: ha.entrypoint,
		Req:        b.Bytes(),
		Header:     msg.Header,
	}, nil) {
		span.SetAttributes(attribute.Key("nats-publish").String(ha.entrypoint))
		span.SetAttributes(attribute.Key("nats-publish at").String(time.Now().String()))
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte("Ok"))
		return
	}
	span.SetAttributes(attribute.Key("Error").String("publishing to nats"))
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
