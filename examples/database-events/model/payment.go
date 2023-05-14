package payment

import (
	"encoding/json"
	"fmt"

	"github.com/Ishan27g/ryo-Faas/store"
)

type Payment struct {
	Id     string  `json:"Id"`
	From   string  `json:"From"`
	To     string  `json:"To"`
	Amount float64 `json:"Amount"`

	InvoiceData map[string]interface{} `json:"InvoiceData,omitempty"`
	EmailSent   bool                   `json:"EmailSent"`
}

func (p *Payment) Marshal() map[string]interface{} {
	m := make(map[string]interface{})
	rec, _ := json.Marshal(p)
	err := json.Unmarshal(rec, &m)
	if err != nil {
		fmt.Println(err.Error())
		return m
	}
	return m
}

func (p *Payment) BuildInvoice() {
	p.InvoiceData = make(map[string]interface{})
	p.InvoiceData = map[string]interface{}{
		"anything": p.Amount,
	}
}
func FromDocument(doc store.Doc) Payment {
	var p = Payment{}
	err := doc.Unmarshal(&p)
	if err != nil {
		fmt.Println(err.Error())
		return p
	}
	return p
}
