package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ishan27g/ryo-Faas/database/handler"
	"github.com/Ishan27g/ryo-Faas/transport"
)

var DefaultHttp = ":5000"
var DefaultGrpc = ":5001"

// Optional flags to change config
var grpcPort = flag.String("grpc", DefaultHttp, "--grpc "+DefaultHttp)
var httpPort = flag.String("htpp", DefaultGrpc, "--http "+DefaultGrpc)

func main() {

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var handler = handler.GetHandler()
	transport.Init(ctx, struct {
		IsDeploy bool
		Server   interface{}
	}{IsDeploy: false, Server: &handler.Rpc}, *grpcPort, handler.Gin, *httpPort).Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
}
