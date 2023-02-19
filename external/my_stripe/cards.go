package my_stripe

import (
	"log"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/token"
)


func CreateBankToken() (*stripe.Token, error) {
	stripe.Key = stripeSecretKey()

	t, err := token.New(&stripe.TokenParams{
		BankAccount: &stripe.BankAccountParams{
		  	Country: stripe.String("GB"),
		  	Currency: stripe.String(string(stripe.CurrencyGBP)),
		  	AccountHolderName: stripe.String("John Yeo"),
		//   AccountHolderType: stripe.String(string(stripe.BankAccountAccountHolderTypeIndividual)),
			AccountNumber: stripe.String("GB82WEST12345698765432"),
		  	RoutingNumber: stripe.String("108800"),
		},
	  })

	if err != nil {
		log.Printf("Failed to create card token")
		return nil, err
	}

	return t, nil
}