package main

import (
	"flag"
	"fmt"
	databaseevents "github.com/Ishan27g/ryo-Faas/deployments/tmp/database-events"
	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

// init definition gets generated
func init() {
	databaseevents.Init()
}

func main() {
	var port = flag.String("port", "", "--port :9000")

	flag.Parse()

	if *port == "" {
		return
	}
	FuncFw.Start(*port)

	<-make(chan byte)
	fmt.Println("exited.....")
}
