package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Ishan27g/ryo-Faas/database/handler"
	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
)

var DefaultGrpc = ":5000"
var DefaultHttp = ":5001"

// Optional flags to change config
var grpcPort = flag.String("grpc", DefaultGrpc, "--grpc "+DefaultGrpc)
var httpPort = flag.String("http", DefaultHttp, "--http "+DefaultHttp)
var jaegerUrl = os.Getenv("JAEGER")
var zipKinUrl = os.Getenv("ZIPKIN")

var appName = "rfa-database"
var serviceName = ""

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
	var db = handler.GetHandler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config := []transport.Config{transport.WithRpcPort(*grpcPort), transport.WithDatabaseServer(&db.Rpc)}
	transport.Init(ctx, config...).Start()

	FuncFw.Start(strings.TrimPrefix(DefaultHttp, ":"))

	<-time.After(5 * time.Second)
	// todo only to check connectivity
	transport.NatsPublish("hello", "ok", nil)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
}
