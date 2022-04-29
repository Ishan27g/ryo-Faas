package tracing

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/trace"
)

const (
	TID       = "traceid"
	SID       = "spanid"
	XCallback = "X-Callback-Url"

	prefix     = "ryo-faas-"
	Entrypoint = prefix + "entrypoint"
	Url        = prefix + "url"
	Status     = prefix + "status"
	IsMain     = prefix + "isMain"
	IsAsync    = prefix + "isAsync"
	Success    = prefix + "success"
	Error      = prefix + "success"

	Pub    = prefix + "pub"
	Sub    = prefix + "sub"
	PubAt  = prefix + "published at"
	SubAt  = prefix + "begin function at"
	SubEnd = prefix + "ended function at"

	ErrorWrite = "could not parse incoming http to nats type"
	SuccessPub = "published http request to nats"
	ErrorPub   = "could not publish to nats"
)

var status = make(chan bool, 1)
var tp = new(provider)

func Disable() {
	enabled := <-status
	defer func() { status <- false }()
	if !enabled {
		return
	}
	tp.Close()
	tp = &provider{
		exporter:       tp.exporter,
		app:            tp.app,
		service:        tp.service,
		TracerProvider: nil,
	}
}
func Enable() {
	enabled := <-status
	defer func() { status <- true }()
	if !enabled {
		Init(tp.exporter, tp.app, tp.service)
	}
}

// https://stackoverflow.com/questions/70378025/how-to-create-opentelemetry-span-from-a-string-traceid
func CreateSpanContext(traceID trace.TraceID, spanID trace.SpanID) context.Context {
	var spanContextConfig trace.SpanContextConfig
	spanContextConfig.TraceID = traceID
	spanContextConfig.SpanID = spanID
	spanContextConfig.TraceFlags = 01
	spanContextConfig.Remote = false
	return trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(spanContextConfig))
}

func ExtractSpan(header nats.Header) context.Context {

	traceID, err := trace.TraceIDFromHex(header.Get(TID))
	if err != nil {
		fmt.Println("no trace Id in header: ", err)
		// fmt.Println("expected otel middleware to inject span into original request")
	}
	spanID, err := trace.SpanIDFromHex(header.Get(SID))
	if err != nil {
		fmt.Println("no span Id in header: ", err)
		// fmt.Println("expected otel middleware to inject span into original request")
	}
	return CreateSpanContext(traceID, spanID)
}

func ExtractHeader(span trace.Span) (header nats.Header) {
	header = make(nats.Header)
	getTraceID := span.SpanContext().TraceID().String()
	header.Set(TID, getTraceID)
	getSpanID := span.SpanContext().SpanID().String()
	header.Set(SID, getSpanID)
	return
}
