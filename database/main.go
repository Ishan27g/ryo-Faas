package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ishan27g/ryo-Faas/database/handler"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
)

var DefaultGrpc = ":5000"
var DefaultHttp = ":5001"

// Optional flags to change config
var grpcPort = flag.String("grpc", DefaultGrpc, "--grpc "+DefaultGrpc)
var httpPort = flag.String("htpp", DefaultHttp, "--http "+DefaultHttp)

func main() {

	flag.Parse()

	var db = handler.GetHandler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config := []transport.Config{transport.WithRpcPort(*grpcPort), transport.WithDatabaseServer(&db.Rpc),
		transport.WithHttpPort(*httpPort), transport.WithHandler(db.Gin)}
	transport.Init(ctx, config...).Start()

	<-time.After(5 * time.Second)
	// todo only to check connectivity
	transport.NatsPublish("hello", "ok", nil)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
}
