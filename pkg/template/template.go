package main

import (
	"flag"
	"fmt"
	"time"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/store"
)

// init definition gets generated
func init() {
}

func main() {
	var port = flag.String("port", "", "--port :9000")

	flag.Parse()

	if *port == "" {
		return
	}
	FuncFw.Start(*port, "")

	go func() {
		<-time.After(3 * time.Second)
		// todo only to check connectivity
		transport.NatsPublish("hello", "ok", nil)
		store.Get("any")
	}()

	<-make(chan byte)
	fmt.Println("exited.....")
}
