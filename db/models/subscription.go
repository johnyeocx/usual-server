package models

import "time"

type Subscription struct {
	ID 			 int 					`json:"sub_id"`
	StripeSubID	 string 				`json:"stripe_sub_id"`
	CustomerID	 int 					`json:"customer_id"`
	PlanID		 int					`json:"plan_id"`
	StartDate	 time.Time				`json:"start_date"`

	Cancelled 		bool					`json:"cancelled"`
	CancelledDate 	JsonNullTime			`json:"cancelled_date"`
	Expires			JsonNullTime			`json:"expires"`

	// additional for customer
	BusinessName *string 				`json:"business_name"`
	BusinessID 	 *int 					`json:"business_id"`
	SubProduct	 *SubscriptionProduct 	`json:"sub_product"`
	CardID		int						`json:"card_id"`
}