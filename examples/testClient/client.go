package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"go.opentelemetry.io/otel/trace"
)

var url = "http://localhost:9999/functions/methodwithotel"
var urlNoop = "http://localhost:9999/functions/methodwithotel/some/thing?noop"

// example of using opentelemetry metrics
// assumes `MethodWithOtel` is deployed and available via proxy at `url`
func main() {
	//requestTest("http://localhost:9997/test")
	//requestTest("http://localhost:9997/test/1212?noop=123")
	os.Setenv("JAEGER", "localhost")
	// connect to jaeger
	jp := tracing.Init("jaeger", "otel-client", "test-Client")
	defer jp.Close()

	//requestWithOtel(url, jp.Get())
	//<-time.After(2 * time.Second)
	//requestWithOtel(urlNoop+"=true", jp.Get())
	//<-time.After(2 * time.Second)
	rand.Seed(time.Now().Unix())
	for i := 0; i < 5; i++ {
		r := time.Duration(rand.Intn(1500))
		<-time.After(r * time.Millisecond)
		for i := 0; i < 5; i++ {
			r := time.Duration(rand.Intn(500))
			<-time.After(r * time.Millisecond)
			requestWithOtel(urlNoop+"=false", jp.Get())
		}
	}
}

// starts a span that gets propagated from this client to the proxy and then to the deployed function.
func requestWithOtel(atUrl string, tr trace.Tracer) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx2, span := tr.Start(ctx, "client-with-otel-header", trace.WithAttributes(semconv.MessagingDestinationKey.String(atUrl)))

	// add baggage to span
	bag, err := baggage.Parse("username=Sensu,id=Bean")
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
	req, _ := http.NewRequestWithContext(ctx3, "GET", atUrl, nil)

	// receiving server will extract the span from the context, add events/attributes
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	//body, err := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	// set response status as an attribute
	fmt.Println("response status is ", res.StatusCode)
	span.SetAttributes(attribute.String("resp-status", res.Status))
	//fmt.Println("response is ", string(body))
}

func requestTest(to string) {
	client := http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}

	// add ctx to http request
	req, _ := http.NewRequest("GET", to, nil)

	// receiving server will extract the span from the context, add events/attributes
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	_ = res.Body.Close()
	fmt.Println(string(body))
}
