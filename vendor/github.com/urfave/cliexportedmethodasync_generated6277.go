package main

import (
	"flag"
	"fmt"
	async "github.com/Ishan27g/ryo-Faas/deployments/async"
	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

var port = flag.String("port", "", "--port :9000")

// init definition gets generated
func init() {
	FuncFw.Export.HttpAsync("MethodAsync", "/methodasync", async.MethodAsync)
}

func main() {
	flag.Parse()

	if *port == "" {
		return
	}
	FuncFw.Start(*port)

	<-make(chan byte)
	fmt.Println("exited.....")
}
