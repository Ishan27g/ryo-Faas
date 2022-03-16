package main

import (
	"fmt"
	"net/http"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/controller"

	"github.com/Ishan27g/ryo-Faas/store"
)

// MakePayment
func MakePayment(w http.ResponseWriter, r *http.Request) {

	// create random payment
	payment := controller.RandomPayment()

	// add to database
	store.Documents.Create(payment.Id, payment.Marshal())

	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Made payment:"+fmt.Sprintf("%v", payment)+"\n")
}
