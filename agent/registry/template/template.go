package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

// init definition gets generated
func init() {
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		return
	}
	FuncFw.Start(port)

	closeLogs := make(chan os.Signal, 1)
	signal.Notify(closeLogs, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-closeLogs

	fmt.Println("EXITING?")
	FuncFw.Stop()
}
