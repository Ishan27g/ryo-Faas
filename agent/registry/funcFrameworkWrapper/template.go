package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/Ishan27g/ryo-Faas/metrics"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

var handlerFunc func(w http.ResponseWriter, r *http.Request)
var entrypoint string

// init definition gets generated to call deploy()
func init() {

}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		return
	}
	url := "/" + os.Getenv("URL")
	fmt.Println("deploying at ", url)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jp := metrics.Jaeger(ctx, "http://jaeger:14268/api/traces") //match with docker hostname
	defer jp.Close()
	// _ = jp.Tracer("function-with-otel")

	// https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/net/http/otelhttp/example
	otelHandler := otelhttp.NewHandler(http.HandlerFunc(handlerFunc), "otel-function-"+entrypoint)
	http.Handle(url, otelHandler)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("httpListenAndServe: %v\n", err)
	}
}
