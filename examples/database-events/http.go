package notMain

import (
	"fmt"
	"net/http"

	"github.com/Ishan27g/ryo-Faas/examples/database-events/controller"
	"github.com/Ishan27g/ryo-Faas/examples/db-events/model"
	"github.com/Ishan27g/ryo-Faas/store"
)

// MakePayment creates random payment and adds to `payments` table in the db
func MakePayment(w http.ResponseWriter, r *http.Request) {

	// create random payment
	payment := controller.RandomPayment()

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
		v := doc.Data.Value
		fmt.Println(v["Value"].(model.Payment))
	}
	w.WriteHeader(http.StatusAccepted)
	fmt.Fprint(w, "All payments:"+fmt.Sprintf("%v", docs))
}
