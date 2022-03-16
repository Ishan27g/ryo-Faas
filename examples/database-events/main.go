package main

import (
	"os"
	"os/signal"
	"syscall"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/store"
	"github.com/Ishan27g/ryo-Faas/types"
)

var events = func() store.StoreEvents {
	return store.StoreEvents{
		OnCreate: []store.EventCb{paymentMade},
		OnGet:    []store.EventCb{paymentsRetrieved},
		OnUpdate: []store.EventCb{paymentsUpdated},
		OnDelete: []store.EventCb{paymentsDeleted},
	}
}

var _ = events().Apply()

func main() {

	store.Documents.OnDelete(func(document types.NatsDoc) {
		paymentsDeleted(document)
	})

	FuncFw.Export.Http("MakePayment", "/pay", MakePayment)

	FuncFw.Start("9999")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
	FuncFw.Stop()
}
