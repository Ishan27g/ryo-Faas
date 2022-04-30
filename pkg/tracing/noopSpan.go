package tracing

import (
	"context"
	"strings"

	"github.com/Ishan27g/go-utils/noop/noop"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

/*
	Noop based on url query param
*/
// NoopSpanFromGin returns a span with noop-ctx if url has a query => `?noop=true`
// otherwise it returns the request's span (if any)
// if the request's ctx has noop, it overrides presence of the query param
func NoopSpanFromGin(c *gin.Context) (trace.Span, context.Context) {
	isNoop := strings.EqualFold(c.Query("noop"), "true")
	ctx := noop.NewCtxWithNoop(c.Request.Context(), isNoop)
	// span := trace.SpanFromContext(ctx)
	return trace.SpanFromContext(ctx), ctx
}

/*
	Wraps Span with noop
*/
//
//type NoopSpan interface {
//	End(options ...trace.SpanEndOption)
//	AddEvent(name string, options ...trace.EventOption)
//	IsRecording() bool
//	RecordError(err error, options ...trace.EventOption)
//	SpanContext() trace.SpanContext
//	SetStatus(code codes.Code, description string)
//	SetName(name string)
//	SetAttributes(kv ...attribute.KeyValue)
//	TracerProvider() trace.TracerProvider
//}
//type noopSpan struct {
//	Span *trace.Span
//}
//
//// NoopSpanFromSpan returns a NoopSpan interface for the provided span
//func NoopSpanFromSpan(existingSpan *trace.Span) NoopSpan {
//	return &noopSpan{existingSpan}
//}

var status = make(chan bool, 1)

func init() {
	status <- true
}

// Disable tracing
func Disable() {
	enabled := <-status
	defer func() { status <- false }()
	if enabled {
		tp.Close()
	}

}
func Enable() {
	enabled := <-status
	defer func() { status <- true }()
	if !enabled {
		Init(exporter, app, tp.service)
	}
}

//func (s *noopSpan) End(options ...trace.SpanEndOption) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	(*s.Span).End(options...)
//}
//
//func (s *noopSpan) AddEvent(name string, options ...trace.EventOption) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	(*s.Span).AddEvent(name, options...)
//}
//
//func (s *noopSpan) IsRecording() bool {
//	//enabled := <-status
//	//defer func() { status <- enabled }()
//	//if !enabled {
//	//	return true
//	//}
//	return (*s.Span).IsRecording()
//}
//
//func (s *noopSpan) RecordError(err error, options ...trace.EventOption) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	(*s.Span).RecordError(err, options...)
//}
//
//func (s *noopSpan) SpanContext() trace.SpanContext {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return trace.NewSpanContext(trace.SpanContextConfig{})
//	}
//	return (*s.Span).SpanContext()
//
//}
//
//func (s *noopSpan) SetStatus(code codes.Code, description string) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	(*s.Span).SetStatus(code, description)
//}
//
//func (s *noopSpan) SetName(name string) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	(*s.Span).SetName(name)
//}
//
//func (s *noopSpan) SetAttributes(kv ...attribute.KeyValue) {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return
//	}
//	(*s.Span).SetAttributes(kv...)
//}
//
//func (s *noopSpan) TracerProvider() trace.TracerProvider {
//	enabled := <-status
//	defer func() { status <- enabled }()
//	if !enabled {
//		return trace.NewNoopTracerProvider()
//	}
//	return (*s.Span).TracerProvider()
//}
