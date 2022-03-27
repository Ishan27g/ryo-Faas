package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/Ishan27g/ryo-Faas/examples/acl"
	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

func main() {
	// notMain.Init()

	FuncFw.Export.Http("AddRole", "/add", acl.AddRole)
	FuncFw.Export.Http("AddChildPermissions", "/update", acl.AddChildPermission)
	FuncFw.Export.Http("CheckPermission", "/get", acl.CheckPermission)

	FuncFw.Start("9999")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
	FuncFw.Stop()
}
