package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/proxy/proxy"
	"github.com/Ishan27g/ryo-Faas/store"
)

// var host = "localhost"

// Optional flags to change ports
var httpPort = flag.String("http", proxy.DefaultHttp, "http port")
var grpcPort = flag.String("rpc", proxy.DefaultRpc, "rpc port")

var jaegerUrl = os.Getenv("JAEGER")
var zipKinUrl = os.Getenv("ZIPKIN")

var appName = proxy.ServiceName
var serviceName = proxy.ServiceName

func main() {
	flag.Parse()
	var provider tracing.TraceProvider
	if jaegerUrl == "" && zipKinUrl != "" {
		provider = tracing.Init("zipkin", appName, serviceName)
	}
	if zipKinUrl == "" && jaegerUrl != "" {
		provider = tracing.Init("jaeger", appName, serviceName)
	}
	if provider == nil {
		provider = tracing.Init("jaeger", appName, serviceName)
	}
	defer provider.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	proxy.Start(ctx, *grpcPort, *httpPort)

	<-time.After(5 * time.Second)
	// todo only to check connectivity
	transport.NatsPublish("hello", "ok", nil)
	store.Get("any")

	<-make(chan bool)
}
