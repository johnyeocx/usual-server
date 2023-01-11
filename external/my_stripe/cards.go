package my_stripe

import (
	"log"

	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/token"
)

// func CreateCardToken(card *models.CreditCard) (*stripe.Token, error) {
// 	stripe.Key = stripeSecretKey()

// 	params := &stripe.TokenParams{
// 		Card: &stripe.CardParams{
// 			Number: stripe.String(card.Number),
// 			ExpMonth: stripe.String(card.ExpMonth),
// 			ExpYear: stripe.String(card.ExpYear),
// 			CVC: stripe.String(card.CVC),
// 			Currency: stripe.String(card.Currency),
// 		},
// 	}

// 	t, err := token.New(params)

// 	if err != nil {
// 		log.Printf("Failed to create card token")
// 		return nil, err
// 	}

// 	return t, nil
// }

// func CreateCard(accountId string, userCard *models.CreditCard) (*stripe.Card, error){
// 	stripe.Key = stripeSecretKey()

// 	// token, err := CreateCardToken(userCard)
// 	// if err != nil {
// 	// 	return nil, err
// 	// }

// 	// fmt.Println("New token | card: ", token.Card.ID)
// 	// fmt.Println("New token | id: ", token.ID)

// 	params := &stripe.CardParams{
// 		Account: stripe.String(accountId),
// 		Token: stripe.String("tok_mastercard_debit"),
// 	}

// 	c, err := card.New(params)
// 	if err != nil {
// 		log.Println(err)
// 		log.Println()
// 		return nil, err
// 	}

// 	return c, nil
// }

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