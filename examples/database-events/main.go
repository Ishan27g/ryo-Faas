package main

import (
	"os"
	"os/signal"
	"syscall"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/store"
)

func main() {

	FuncFw.Export.Http("MakePayment", "/pay", MakePayment)
	FuncFw.Export.Http("GetPayment", "/get", GetPayments)

	// register functions that subscribe to respective queries to the `payments` table
	// when a new payment document is created, generate its invoice
	FuncFw.Export.EventsFor(TableName).On(store.DocumentCREATE, GeneratePaymentPdf)
	FuncFw.Export.EventsFor(TableName).On(store.DocumentGET, Retrieved)
	// when a payment is updated, send email to users
	FuncFw.Export.EventsFor(TableName).On(store.DocumentUPDATE, EmailUsers)

	// or subscribe to respective queries for a specific documents in the table
	//FuncFw.Export.EventsFor(TableName).OnIds(store.DocumentUPDATE, PaymentsUpdated,
	//	"some-known-id", "another-known-id")
	//
	FuncFw.Start("9997")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
	FuncFw.Stop()
}
