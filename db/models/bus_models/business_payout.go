package bus_models

import (
	"time"

	"github.com/stripe/stripe-go/v74"
)

type BusinessPayout struct {
	ID				int 					`json:"payout_id"`
	Amount			int 					`json:"amount"`
	BusinessID		int 					`json:"business_id"`
	Currency		string 					`json:"currency"`
	Status			stripe.PayoutStatus 	`json:"status"`
	ArrivalDate		time.Time 				`json:"arrival_date"`
	StripePayoutID	string 					`json:"stripe_payout_id"`
	StripeDestID		string 					`json:"stripe_dest_id"`
	Type			string 					`json:"type"`
	ExternalAccountID int 					`json:"external_account_id"`
}