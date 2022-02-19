package metrics

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
)

type JaegerProvider struct {
	ctx context.Context
	url string
	*tracesdk.TracerProvider
}

func Jaeger(ctx context.Context, url string) *JaegerProvider {
	exp, err := jaeger.New(jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(url)))
	if err != nil {
		fmt.Println("CANNOT CONNECT TO JAEGER ", err.Error())
		return nil
	}
	fmt.Println("CONNECTED TO JAEGER ", url)
	tp := JaegerExporter(exp)
	otel.SetTracerProvider(tp)
	// https://github.com/open-telemetry/opentelemetry-go-contrib/blob/c59b4421d98c35a7fa9735647507304da112ff51/instrumentation/net/http/otelhttp/example/server/server.go#L52
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	return &JaegerProvider{
		ctx,
		url,
		tp,
	}
}

func JaegerExporter(exp *jaeger.Exporter) *tracesdk.TracerProvider {
	return tracesdk.NewTracerProvider(
		tracesdk.WithBatcher(exp),
		tracesdk.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(service),
			attribute.String("environment", environment),
			attribute.Int64("ID", id),
		)),
	)
}

func (j *JaegerProvider) Close() {
	ctx, cancel := context.WithCancel(j.ctx)
	defer cancel()
	defer func(ctx context.Context) {
		ctx, cancel = context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		if err := j.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}(ctx)
}
