package main

import (
	"context"
	"flag"

	"github.com/Ishan27g/ryo-Faas/plugins"
	"github.com/Ishan27g/ryo-Faas/proxy/proxy"
)

var DefaultRpc = ":9001"
var DefaultHttp = ":9002"
var host = "localhost"

// Optional flags to change ports
var httpPort = flag.String("http", DefaultHttp, "http port")
var grpcPort = flag.String("rpc", DefaultRpc, "rpc port")

func main() {
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	url := "http://localhost:14268/api/traces"
	provider := plugins.InitJaeger(ctx, "ryo-Faas-proxy", "proxy-server", url)
	defer provider.Close()

	proxy.Start(ctx, host+*grpcPort, host+*httpPort)
	// handler.AgentConnectionType = transport.RPC
	<-make(chan bool)
}
