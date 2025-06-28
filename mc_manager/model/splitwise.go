package model

type SplitwiseUser struct {
	Id    string  `json:"id"`
	Name  *string `json:"name"`
	Email *string `json:"email"`
}

type SplitwiseConfig struct {
	InvoiceCurrency string           `json:"invoice_currency"`
	Users           []*SplitwiseUser `json:"users"`
	Currencies      []string         `json:"currencies"`
}
