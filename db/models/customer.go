package models

import "time"

type Customer struct {
	ID 					int 			`json:"customer_id"`
	Name				string 			`json:"name"`
	Email 				string 			`json:"email"`
	Address 			*Address 		`json:"address"`
	StripeID 			string 			`json:"stripe_id"`
	DefaultCardID	 	JsonNullInt16	`json:"default_card_id"`
	Uuid 				string 			`json:"uuid"`
}

type CardInfo struct {
	ID 			int 	`json:"card_id"`
	Last4		string	`json:"last4"`
	StripeID 	string	`json:"stripe_id"`
	CusID		int		`json:"customer_id"`
	Brand		string	`json:"brand"`
}


type Invoice struct {
	ID					int 			`json:"invoice_id"`
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
	DefaultPaymentMethod	JsonNullString		`json:"default_payment_method"`
	ApplicationFeeAmt	JsonNullInt64 	`json:"app_fee_amt"`

	// NULLLABLES
	Subscription		*Subscription   `json:"sub"`
	CardInfo			*CardInfo		`json:"card_info"`
}

type ExploreResult struct {
	Business 	Business			`json:"business"`
	SubProduct 	SubscriptionProduct	`json:"sub_product"`
}

