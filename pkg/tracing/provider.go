package tracing

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/zipkin"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

type TraceProvider interface {
	Get() trace.Tracer
	Close()
}

type provider struct {
	app, service string
	*tracesdk.TracerProvider
}

func (j *provider) Get() trace.Tracer {
	return j.TracerProvider.Tracer(j.service)
}

func (j *provider) Close() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer func(ctx context.Context) {
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := j.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)
}

func Init(exporter, app, service string) TraceProvider {

	var exp tracesdk.SpanExporter = nil
	var err error

	switch exporter {
	case "jaeger":
		jaegerUrl := "http://" + os.Getenv("JAEGER") + ":14268/api/traces"
		exp, err = jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(jaegerUrl)))
		if err != nil {
			fmt.Println("Cannot connect to Jaeger ", err.Error())
			return nil
		}
		return initProvider(app, service, exp)
	case "zipkin":
		zipkinUrl := "http://" + os.Getenv("ZIPKIN") + ":9411/api/v2/spans"
		exp, err = zipkin.New(zipkinUrl)
		if err != nil {
			fmt.Println("Cannot connect to Zipkin ", err.Error())
		}
		return initProvider(app, service, exp)
	}
	return &provider{app, service, nil}
}

func initProvider(app, service string, exporter tracesdk.SpanExporter) TraceProvider {
	tp := tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exporter),
		//		tracesdk.WithSpanProcessor(tracesdk.NewBatchSpanProcessor(exporter)),
		tracesdk.WithSampler(tracesdk.TraceIDRatioBased(1)),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(app),
			semconv.ServiceNamespaceKey.String(service),
		)),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{}))

	return &provider{app, service, tp}
}
