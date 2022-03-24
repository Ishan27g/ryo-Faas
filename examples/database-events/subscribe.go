package main

import (
	"fmt"
	"time"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/model"
	"github.com/Ishan27g/ryo-Faas/store"
)

// background job to generate pdf for a new payment
func generatePaymentPdf(document store.Doc) {
	fmt.Println("Document.Create - generatePaymentPdf")
	payment := payment.FromDocument(document)

	// generateInvoice
	<-time.After(5 * time.Second)
	payment.BuildInvoice()

	// update database
	store.Get(TableName).Update(payment.Id, payment.Marshal())
	fmt.Println("Generated Pdf for payment:", payment.Id)
}

// background job to email users once invoice is created
func emailUsers(document store.Doc) {
	fmt.Println("Document.Update - emailUsers")
	payment := payment.FromDocument(document)

	emailBody := payment.InvoiceData

	// email users
	{
		<-time.After(5 * time.Second)
		fmt.Println(emailBody)
	}

	fmt.Println("Emailed users pdf payment:", payment.Id)
}

func paymentsUpdated(document store.Doc) {
	fmt.Println("Document.OnUpdate")
	payment := payment.FromDocument(document)
	fmt.Println("Updated payment:", payment)
}
