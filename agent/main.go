package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ishan27g/ryo-Faas/agent/registry"
	"github.com/Ishan27g/ryo-Faas/examples/plugins"
	"github.com/Ishan27g/ryo-Faas/transport"
)

var DefaultPort = ":9000"

// var host = "localhost"

var rpcAddr = DefaultPort

// Optional flags to change config
var port = flag.String("port", rpcAddr, "--port :9000")

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	flag.Parse()

	jp := plugins.InitJaeger(ctx, "ryo-Faas-agent", "agent", "http://jaeger:14268/api/traces") //match with docker hostname
	defer jp.Close()

	agent := registry.Init(*port)
	transport.Init(ctx, struct {
		IsDeploy bool
		Server   interface{}
	}{IsDeploy: true, Server: agent}, rpcAddr, nil, "").Start()

	closeLogs := make(chan os.Signal, 1)
	signal.Notify(closeLogs, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-closeLogs

	fmt.Println("EXITING?")
}
