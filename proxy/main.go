package main

import (
	"context"
	"flag"

	"github.com/Ishan27g/ryo-Faas/metrics"
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
	metrics.InitPrometheus()
	url := "http://localhost:14268/api/traces"
	provider := metrics.InitJaeger(ctx, "ryo-Faas-proxy", "proxy", url)
	defer provider.Close()

	proxy.Start(ctx, host+*grpcPort, host+*httpPort)
	// handler.AgentConnectionType = transport.RPC
	<-make(chan bool)
}
