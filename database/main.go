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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var handler = handler.GetHandler()
	transport.Init(ctx, struct {
		IsDeploy bool
		Server   interface{}
	}{IsDeploy: false, Server: &handler.Rpc}, *grpcPort, handler.Gin, *httpPort).Start()

	<-time.After(5 * time.Second)
	// todo only to check connectivity
	transport.NatsPublish("hello", "ok", nil)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
}
