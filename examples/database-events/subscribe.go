package main

import (
	"fmt"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/model"
	"github.com/Ishan27g/ryo-Faas/store"
)

var toPayment = func(doc store.NatsDoc) model.Payment {
	m := doc.Document()
	return m["payment"].(model.Payment)
}

func paymentMade(document store.NatsDoc) {
	fmt.Println("Document.OnCreate")
	document.Print()
	payment := toPayment(document)
	fmt.Println("New payment:", payment)
}
func paymentsRetrieved(document store.NatsDoc) {
	fmt.Println("Document.OnGet")
	document.Print()
	payment := toPayment(document)
	fmt.Println("Retrived payment:", payment)
}
func paymentsUpdated(document store.NatsDoc) {
	fmt.Println("Document.OnUpdate")
	document.Print()
	payment := toPayment(document)
	fmt.Println("Updated payment:", payment)
}
func paymentsDeleted(document store.NatsDoc) {
	fmt.Println("Document.OnDelete")
	document.Print()
	payment := toPayment(document)
	fmt.Println("Deleted payment:", payment)
}
