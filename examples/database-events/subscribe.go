package main

import (
	"fmt"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/model"
	"github.com/Ishan27g/ryo-Faas/store"
)

var toPayment = func(doc store.Doc) model.Payment {
	m := doc.Data.Value["Value"]
	return m.(model.Payment)
}

func paymentMade(document store.Doc) {
	fmt.Println("Document.OnCreate")
	payment := toPayment(document)
	fmt.Println("New payment:", payment)
}
func paymentsRetrieved(document store.Doc) {
	fmt.Println("Document.OnGet")
	payment := toPayment(document)
	fmt.Println("Retrived payment:", payment)
}
func paymentsUpdated(document store.Doc) {
	fmt.Println("Document.OnUpdate")
	payment := toPayment(document)
	fmt.Println("Updated payment:", payment)
}
func paymentsDeleted(document store.Doc) {
	fmt.Println("Document.OnDelete")
	payment := toPayment(document)
	fmt.Println("Deleted payment:", payment)
}
