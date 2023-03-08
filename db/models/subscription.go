package models

import (
	"time"

	"github.com/stripe/stripe-go/v74"
)

type  Subscription struct {
	ID 			 int 					`json:"sub_id"`
	StripeSubID	 string 				`json:"stripe_sub_id"`
	CustomerID	 int 					`json:"customer_id"`
	PlanID		 int					`json:"plan_id"`
	StartDate	 time.Time				`json:"start_date"`
	
	Cancelled 		bool					`json:"cancelled"`
	CancelledDate 	JsonNullTime			`json:"cancelled_date"`
	Expires			JsonNullTime			`json:"expires"`
	
	// additional for customer
	CardID			int						`json:"card_id"`
	BusinessName 	*string 				`json:"business_name"`
	BusinessID 	 	*int 					`json:"business_id"`
	SubProduct	 	*SubscriptionProduct 	`json:"sub_product"`

	LastInvoice		*Invoice			`json:"last_invoice"`
}

type CreateSubReturn struct {
	Sub 			Subscription `json:"sub"`
	Status			stripe.PaymentIntentStatus 		`json:"status"`
	PaymentIntent 	*stripe.PaymentIntent `json:"payment_intent"`
}

type ResumeSubReturn struct {
	Status			stripe.PaymentIntentStatus 		`json:"status"`
	PaymentIntent 	*stripe.PaymentIntent 			`json:"payment_intent"`
	LastInvoice 	*Invoice 						`json:"last_invoice"`
}