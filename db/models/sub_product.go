package models

type InvoiceData struct {
	CustomerID			int 			`json:"customer_id"`
	CustomerStripeID	string 			`json:"stripe_cus_id"`
	CustomerName		string 			`json:"customer_name"`
	InvoiceID			string			`json:"invoice_id"`
	Total				int				`json:"total"`
	InvoiceURL			string			`json:"invoice_url"`
	Status				string			`json:"status"`
	Attempted			bool			`json:"attempted"`
	ApplicationFeeAmt	JsonNullInt64	`json:"app_fee_amt"`
}