package main

import (
	"os"
	"os/signal"
	"syscall"

	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
)

// var events = func() store.StoreEvents {
// 	return store.StoreEvents{
// 		OnCreate: []store.EventCb{paymentMade},
// 		OnGet:    []store.EventCb{paymentsRetrieved},
// 		OnUpdate: []store.EventCb{paymentsUpdated},
// 		OnDelete: []store.EventCb{paymentsDeleted},
// 	}

// }

//var _ = events().Apply()

func main() {

	FuncFw.Export.Http("MakePayment", "/pay", MakePayment)

	FuncFw.Export.Events(FuncFw.StoreEvents{
		OnCreate: FuncFw.EventCbs{paymentMade},
		OnGet:    FuncFw.EventCbs{paymentsRetrieved},
		OnUpdate: FuncFw.EventCbs{paymentsUpdated},
		OnDelete: FuncFw.EventCbs{paymentsDeleted},
	})

	FuncFw.Start("9999")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
	FuncFw.Stop()
}
