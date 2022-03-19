package model

type Payment struct {
	Id     string  `json:"Id"`
	From   string  `json:"From"`
	To     string  `json:"To"`
	Amount float64 `json:"Amount"`
}

func (p Payment) Marshal() map[string]interface{} {
	m := make(map[string]interface{})
	pm := p
	m["payment"] = pm
	return m
}
