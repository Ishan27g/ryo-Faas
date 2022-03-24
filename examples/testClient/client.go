package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Ishan27g/ryo-Faas/examples/plugins"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

// https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/net/http/otelhttp/example
var url = "http://localhost:9999/functions/methodwithotel"

// example of a jaeger metric
// assumes `methodwithotel` is deployed and available via proxy at `url`
// starts a span that gets propagated from this client to the proxy and then to the deployed function.
func main() {
	requestWithOtel()
}

func requestWithOtel() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// connect to jaeger
	jp := plugins.InitJaeger(ctx, "otel-client", "test-Client", "http://localhost:14268/api/traces")
	defer jp.Close()

	tr := jp.Tracer("otel-client")

	// new span
	ctx2, span := tr.Start(ctx, "client-with-otel-header", trace.WithAttributes(semconv.MessagingDestinationKey.String(url)))

	// add baggage to span
	bag, err := baggage.Parse("username=goku")
	if err != nil {
		panic(err.Error())
	}
	ctx3 := baggage.ContextWithBaggage(ctx2, bag)

	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	now := time.Now()
	defer func() {
		span.SetAttributes(attribute.String("took time", time.Since(now).String()))
		// end the span
		span.End()
	}()

	// add ctx to http request
	req, _ := http.NewRequestWithContext(ctx3, "GET", url, nil)

	// receiving server's will extract the span from the context
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	fmt.Println("response is ", string(body))

	// set response status as an attribute
	span.SetAttributes(attribute.String("resp-status", res.Status))

}