package dbMain

import (
	"fmt"
	"time"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/model"
	"github.com/Ishan27g/ryo-Faas/store"
)

const TableName = "payments"

// time consuming background job to generate pdf for a new payment
func GeneratePaymentPdf(document store.Doc) {
	fmt.Println("[StoreEvent][Document.Create] - {generatePaymentPdf}")
	payment := payment.FromDocument(document)
	//
	// generateInvoice
	fmt.Println("[StoreEvent][Document.Create] - Building invoice")
	<-time.After(5 * time.Second)
	payment.BuildInvoice()

	// update database
	fmt.Println("[StoreEvent][Document.Create] Generated invoice for payment:", payment.Id)
	store.Get(TableName).Update(payment.Id, payment.Marshal())

	// add event for a particular ID
	go func() {
		store.Get(TableName).On(store.DocumentUPDATE, EmailSent, payment.Id)
	}()
}

// background job to email users once invoice is created
func EmailUsers(document store.Doc) {
	fmt.Println("[StoreEvent][Document.Update] - {emailUsers}")
	payment := payment.FromDocument(document)
	if payment.EmailSent {
		return
	}

	emailBody := payment.InvoiceData
	fmt.Println("[StoreEvent][Document.Update] - Emailing invoice to users ")
	// email users
	{
		<-time.After(5 * time.Second)
		fmt.Println("Email - ", emailBody)
	}
	payment.EmailSent = true
	store.Get(TableName).Update(payment.Id, payment.Marshal())
	fmt.Println("[StoreEvent][Document.Update] - Emailed users: ", payment.Id)
}

func EmailSent(document store.Doc) {
	fmt.Println("[StoreEvent][Document.OnUpdate][ID] ", document.Id)
	//payment := payment.FromDocument(document)
	//fmt.Println(payment)
}

func Retrieved(document store.Doc) {
	fmt.Println("[StoreEvent][Document.OnGet] ", document.Id)
	//payment := payment.FromDocument(document)
	//fmt.Println(payment)
}
