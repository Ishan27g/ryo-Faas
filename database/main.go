package database

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ishan27g/ryo-Faas/transport"
)

var DefaultPort = ":9000"

// Optional flags to change config
var port = flag.String("port", DefaultPort, "--port :9000")

func main() {

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	transport.Init(ctx, handler{GetDatabase()}, *port, nil, "").Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
}
