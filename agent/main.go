package main

import (
	"context"
	"flag"

	"github.com/Ishan27g/ryo-Faas/agent/registry"
	"github.com/Ishan27g/ryo-Faas/transport"
	"github.com/Ishan27g/ryo-Faas/plugins"

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
	transport.Init(ctx, agent, rpcAddr, nil, "").Start()

	agent.Println(*agent)
	<-make(chan bool)
}
