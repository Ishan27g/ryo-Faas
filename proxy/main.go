package main

import (
	"context"
	"flag"
	"strings"

	"github.com/Ishan27g/ryo-Faas/examples/plugins"
	"github.com/Ishan27g/ryo-Faas/proxy/proxy"
)

var DefaultRpc = ":9001"
var DefaultHttp = ":9002"
var host = "localhost"

// Optional flags to change ports
var httpPort = flag.String("http", DefaultHttp, "http port")
var grpcPort = flag.String("rpc", DefaultRpc, "rpc port")

var agents = flag.String("agents", "", "agent address's")

func main() {
	flag.Parse()

	var ag []string
	if *agents != "" {
		ag = append(strings.Split(*agents, " "), flag.Args()...)
	}
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()
	url := "http://localhost:14268/api/traces"
	provider := plugins.InitJaeger(ctx, "ryo-Faas-proxy", "proxy-server", url)
	defer provider.Close()

	proxy.Start(ctx, host+*grpcPort, host+*httpPort, ag...)
	// handler.AgentConnectionType = transport.RPC
	<-make(chan bool)
}
