package models

type Address struct {
	Country string `json:"country"`
	Line1 string `json:"line1"`
	Line2 string `json:"line2"`
	PostalCode string `json:"postal_code"`
	City string `json:"city"`
}