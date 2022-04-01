package main

import (
	"fmt"
	"net/http"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/controller"
	payment "github.com/Ishan27g/ryo-Faas/examples/database-events/model"
	"github.com/Ishan27g/ryo-Faas/store"
)

// MakePayment creates random payment and adds to `payments` table in the db
func MakePayment(w http.ResponseWriter, r *http.Request) {

	// create random payment
	payment := controller.RandomPayment()
	fmt.Println("Made payment:" + fmt.Sprintf("%v", payment) + "\n")

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
		p := payment.FromDocument(doc)
		fmt.Println(p)
	}
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "All payments:"+fmt.Sprintf("%v", docs))
}
