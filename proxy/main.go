package main

import (
	"context"
	"flag"
	"time"

	"github.com/Ishan27g/ryo-Faas/examples/plugins"
	"github.com/Ishan27g/ryo-Faas/proxy/proxy"
	"github.com/Ishan27g/ryo-Faas/store"
	"github.com/Ishan27g/ryo-Faas/transport"
)

// var host = "localhost"

// Optional flags to change ports
var httpPort = flag.String("http", proxy.DefaultHttp, "http port")
var grpcPort = flag.String("rpc", proxy.DefaultRpc, "rpc port")

func main() {
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	url := "http://localhost:14268/api/traces"
	provider := plugins.InitJaeger(ctx, "ryo-Faas-proxy", "proxy-server", url)
	defer provider.Close()

	proxy.Start(ctx, *grpcPort, *httpPort)

	<-time.After(5 * time.Second)
	// todo only to check connectivity
	transport.NatsPublish("hello", "ok", nil)
	store.Get("any")

	// handler.AgentConnectionType = transport.RPC
	<-make(chan bool)
}
