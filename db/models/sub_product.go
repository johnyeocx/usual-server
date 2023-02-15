package models

type SubscriptionProduct struct {
	Product 	Product 			`json:"product"`
	SubPlan		SubscriptionPlan 	`json:"subscription_plan"`	
}

type Product struct {
	ProductID		int 	`json:"product_id"`
	BusinessID		int		`json:"business_id"`
	Name	 		string 	`json:"name"`
	Description		string	`json:"description"`
	CategoryID		*int	`json:"category_id"`
	StripeProductID	*string	`json:"stripe_product_id"`
	SubCount		*int	`json:"sub_count"`
	CatTitle		*string `json:"category_title"`
}

type SubscriptionPlan struct {
	PlanID				int 			`json:"plan_id"`
	ProductID 			int				`json:"product_id"`
	RecurringDuration	TimeFrame		`json:"recurring_duration"`
	UnitAmount			int			`json:"unit_amount"`
	Currency			string			`json:"currency"`
	StripePriceID 		*string 		`json:"stripe_price_id"`
	Usages				*[]SubUsage		`json:"usages"`
}

type SubUsage struct {
	ID 				int 			`json:"sub_usage_id"`
	PlanID 			int 			`json:"plan_id"`
	Title 			string 			`json:"title"`
	Unlimited 		bool 			`json:"unlimited"`
	Interval		JsonNullString 	`json:"interval"`
	Amount			JsonNullInt16 	`json:"amount"`
}

type InvoiceData struct {
	CustomerID			int 			`json:"customer_id"`
	CustomerStripeID	string 			`json:"stripe_cus_id"`
	CusLastName			string 			`json:"cus_first_name"`
	CusFirstName		string 			`json:"cus_last_name"`
	InvoiceID			string			`json:"invoice_id"`
	Total				int				`json:"total"`
	InvoiceURL			string			`json:"invoice_url"`
	Status				string			`json:"status"`
	Attempted			bool			`json:"attempted"`
	ApplicationFeeAmt	JsonNullInt64	`json:"app_fee_amt"`

	ProductName 		*string 			`json:"product_name"`
	ProductID 			*int 			`json:"product_id"`
	// SubPlan 			*string 			`json:"product_id"`
}

type ProductCategory struct {
	Title		string `json:"title"`
	CategoryID	*int `json:"category_id"`
	BusinessID	*string `json:"business_id"`
}

type TimeFrame struct {
	Interval		JsonNullString		`json:"interval"`
	IntervalCount	JsonNullInt16		`json:"interval_count"`
}

