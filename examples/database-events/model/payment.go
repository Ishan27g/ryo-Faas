package payment

import (
	"github.com/Ishan27g/ryo-Faas/store"
)

type Payment struct {
	Id     string  `json:"Id"`
	From   string  `json:"From"`
	To     string  `json:"To"`
	Amount float64 `json:"Amount"`

	InvoiceData map[string]interface{} `json:"InvoiceData,omitempty"`
}

func (p Payment) Marshal() map[string]interface{} {
	m := make(map[string]interface{})
	pm := p
	m["payment"] = pm
	return m
}

func (p Payment) BuildInvoice() {
	p.InvoiceData = make(map[string]interface{})
	p.InvoiceData = map[string]interface{}{
		"anything": "ok",
	}
}
func FromDocument(doc store.Doc) Payment {
	m := doc.Data.Value["Value"]
	return m.(Payment)
}
