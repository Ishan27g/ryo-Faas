package main

import (
	"fmt"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/model"
	"github.com/Ishan27g/ryo-Faas/store"
)

const TableName = "payment"

var toPayment = func(doc store.NatsDoc) model.Payment {
	m := doc.Document()
	return m[TableName].(model.Payment)
}

func PaymentMade(document store.NatsDoc) {
	fmt.Println("Document.OnCreate")
	document.Print()
	payment := toPayment(document)
	fmt.Println("New payment:", payment)
}
func PaymentsUpdated(document store.NatsDoc) {
	fmt.Println("Document.OnUpdate")
	document.Print()
	payment := toPayment(document)
	fmt.Println("Updated payment:", payment)
}
