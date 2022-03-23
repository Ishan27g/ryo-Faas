package main

import (
	"fmt"

	"github.com/Ishan27g/ryo-Faas/database/db"
	"github.com/Ishan27g/ryo-Faas/examples/database-events/model"
)

const TableName = "payment"

var toPayment = func(doc database.NatsDoc) payment.payment {
	m := doc.Document()
	return m[TableName].(payment.payment)
}

func PaymentMade(document database.NatsDoc) {
	fmt.Println("Document.OnCreate")
	document.Print()
	payment := toPayment(document)
	fmt.Println("New payment:", payment)
}
func PaymentsUpdated(document database.NatsDoc) {
	fmt.Println("Document.OnUpdate")
	document.Print()
	payment := toPayment(document)
	fmt.Println("Updated payment:", payment)
}
