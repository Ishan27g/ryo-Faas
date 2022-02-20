package main

import (
	"context"
	"flag"

	"github.com/Ishan27g/ryo-Faas/agent/registry"
	"github.com/Ishan27g/ryo-Faas/metrics"
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
	agent := registry.Init(*port)
	jp := metrics.InitJaeger(ctx, "ryo-Faas-agent", "", "http://localhost:14268/api/traces")
	defer jp.Close()
	transport.Init(ctx, agent, rpcAddr, nil, "").Start()

	agent.Println(*agent)
	<-make(chan bool)
}
