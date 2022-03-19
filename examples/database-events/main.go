package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/controller"
	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/store"
)

const TableName = "payments"

// MakePayment creates random payment and adds to `payments` table in the db
func MakePayment(w http.ResponseWriter, r *http.Request) {
	// create random payment
	payment := controller.RandomPayment()
	// add to database
	_ = store.Get(TableName).Create(payment.Id, payment.Marshal())
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Made payment:"+fmt.Sprintf("%v", payment)+"\n")
}

// GetPayments return all entries from the `payments` table in the db
func GetPayments(w http.ResponseWriter, r *http.Request) {
	// retrieve from db
	docs := store.Get(TableName).Get()
	for _, doc := range docs {
		(*doc).Print()
	}
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "All payments:"+fmt.Sprintf("%v", (*docs[0]).Document())+"\n")
}
func main() {

	FuncFw.Export.Http("MakePayment", "/pay", MakePayment)
	FuncFw.Export.Http("GetPayment", "/get", GetPayments)

	// register functions that subscribe to respective queries to the `payments` table
	FuncFw.Export.EventsFor("payments").On(store.DocumentCREATE, paymentMade)
	FuncFw.Export.EventsFor("payments").On(store.DocumentGET, paymentsRetrieved)
	FuncFw.Export.EventsFor("payments").On(store.DocumentDELETE, paymentsDeleted)

	// or subscribe to respective queries for a specific documents in the table
	FuncFw.Export.EventsFor("payments").OnIds(store.DocumentUPDATE, paymentsUpdated,
		"some-known-id", "another-known-id")

	FuncFw.Start("9999")

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-stop
	FuncFw.Stop()
}
