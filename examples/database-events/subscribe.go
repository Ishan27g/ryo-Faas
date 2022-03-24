package notMain

import (
	"fmt"
	"time"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/model"
	"github.com/Ishan27g/ryo-Faas/store"
)

const TableName = "payments"

// background job to generate pdf for a new payment
func GeneratePaymentPdf(document store.Doc) {
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
func EmailUsers(document store.Doc) {
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

func PaymentsUpdated(document store.Doc) {
	fmt.Println("Document.OnUpdate")
	payment := payment.FromDocument(document)
	fmt.Println("Updated payment:", payment)
}
