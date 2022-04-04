package main

import (
	FuncFw "github.com/Ishan27g/ryo-Faas/funcFw"
	"github.com/Ishan27g/ryo-Faas/store"
)

func Init() {

	FuncFw.Export.Http("MakePayment", "/pay", MakePayment)
	FuncFw.Export.Http("GetPayment", "/get", GetPayments)

	// register functions that subscribe to respective queries to the `payments` table
	// when a new payment document is created, generate its invoice
	FuncFw.Export.EventsFor(TableName).On(store.DocumentCREATE, GeneratePaymentPdf)
	FuncFw.Export.EventsFor(TableName).On(store.DocumentGET, Retrieved)
	// when a payment is updated, send email to users
	FuncFw.Export.EventsFor(TableName).On(store.DocumentUPDATE, EmailUsers)

	// or subscribe to respective queries for a specific documents in the table
	//FuncFw.Export.EventsFor(TableName).OnIds(store.DocumentUPDATE, PaymentsUpdated,
	//	"some-known-id", "another-known-id")
	//
}
