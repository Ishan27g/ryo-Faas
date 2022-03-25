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
	EmailSent   bool                   `json:"EmailSent"`
}

func (p *Payment) Marshal() map[string]interface{} {
	m := make(map[string]interface{})
	pm := p
	m["payment"] = pm
	return m
}

func (p *Payment) BuildInvoice() {
	p.InvoiceData = make(map[string]interface{})
	p.InvoiceData = map[string]interface{}{
		"anything": p.Amount,
	}
}
func FromDocument(doc store.Doc) Payment {
	m := doc.Data.Value["payment"].(map[string]interface{})
	p := Payment{
		Id:          m["Id"].(string),
		From:        m["From"].(string),
		To:          m["To"].(string),
		Amount:      m["Amount"].(float64),
		InvoiceData: make(map[string]interface{}),
		EmailSent:   m["EmailSent"].(bool),
	}
	if m["InvoiceData"] != nil {
		p.InvoiceData = m["InvoiceData"].(map[string]interface{})
	}
	return p
}
