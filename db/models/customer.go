package models

import "time"

type Customer struct {
	ID 			int 		`json:"id"`
	Name		string 		`json:"name"`
	Email 		string 		`json:"email"`
	Address 	*Address 	`json:"address"`
	StripeID 	string 		`json:"stripe_id"`
}

type Subscription struct {
	ID 			int 		`json:"id"`
	StripeSubID	string 		`json:"stripe_sub_id"`
	CustomerID	int 		`json:"customer_id"`
	PlanID		int			`json:"plan_id"`
	StartDate	time.Time	`json:"start_date"`
}