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
	store.Get("payments").Create(payment.Id, payment.Marshal())
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "Made payment:"+fmt.Sprintf("%v", payment)+"\n")

}
func GetPayment(w http.ResponseWriter, r *http.Request) {
	// create random payment
	// payment := controller.RandomPayment()

	// add to database
	docs := store.Get("payments").Get()
	for _, doc := range docs {
		(*doc).Print()
	}
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "All payment:"+fmt.Sprintf("%v", (*docs[0]).Document())+"\n")
}
