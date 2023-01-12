package models

import "time"

type Customer struct {
	ID 			int 		`json:"customer_id"`
	Name		string 		`json:"name"`
	Email 		string 		`json:"email"`
	Address 	*Address 	`json:"address"`
	StripeID 	string 		`json:"stripe_id"`
}

type Subscription struct {
	ID 			int 		`json:"sub_id"`
	StripeSubID	string 		`json:"stripe_sub_id"`
	CustomerID	int 		`json:"customer_id"`
	PlanID		int			`json:"plan_id"`
	StartDate	time.Time	`json:"start_date"`
}

type Invoice struct {
	ID					int 			`json:"int"`
	InStripeID			string 			`json:"stripe_in_id"`
	CusStripeID			string 			`json:"stripe_cus_id"`
	SubStripeID			JsonNullString 	`json:"stripe_sub_id"`
	PriceStripeID		string			`json:"stripe_price_id"`
	ProdStripeID		string			`json:"stripe_prod_id"`
	Paid				bool 			`json:"paid"`
	Attempted			bool			`json:"attempted"`
	Status				string 			`json:"status"`
	Total				int				`json:"total"`
	Created				time.Time 		`json:"created"`
	InvoiceURL			string			`json:"invoice_url"`
	ApplicationFeeAmt	JsonNullInt64 	`json:"app_fee_amt"`
}