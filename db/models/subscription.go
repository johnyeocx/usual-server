package models

type ProductCategory struct {
	Title		string `json:"title"`
	CategoryID	*int `json:"category_id"`
	BusinessID	*string `json:"business_id"`
}

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
}

type SubscriptionPlan struct {
	PlanID				int 			`json:"plan_id"`
	ProductID 			int				`json:"product_id"`
	RecurringDuration	TimeFrame		`json:"recurring_duration"`
	UnitAmount			int			`json:"unit_amount"`
	Currency			string			`json:"currency"`
	UsageUnlimited		bool			`json:"usage_unlimited"`
	UsageDuration		*TimeFrame		`json:"usage_duration"`
	UsageAmount			*JsonNullInt16	`json:"usage_amount"`
	StripePriceID 		*string 		`json:"stripe_price_id"`
}

type TimeFrame struct {
	Interval		JsonNullString		`json:"interval"`
	IntervalCount	JsonNullInt16		`json:"interval_count"`
}

