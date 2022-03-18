package main

import (
	"os"
	"os/signal"
	"syscall"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

func main() {

	FuncFw.Export.Http("MakePayment", "/pay", MakePayment)
	FuncFw.Export.Http("GetPayment", "/get", GetPayment)

	//FuncFw.Export.Events(FuncFw.StoreEvents{
	//	OnCreate: FuncFw.Events{paymentMade},
	//	OnGet:    FuncFw.Events{paymentsRetrieved},
	//	OnUpdate: FuncFw.Events{paymentsUpdated},
	//	OnDelete: FuncFw.Events{paymentsDeleted},
	//})

	FuncFw.Start("9999")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
	FuncFw.Stop()
}
