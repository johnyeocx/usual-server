package my_stripe

import (
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
)

func UpdateIndividualName(
	accountId string,
	firstName string,
	lastName string,
) (error) {
	stripe.Key = stripeSecretKey()
	
	// 1. Update account details with business profile
	params := &stripe.AccountParams{
		Individual: &stripe.PersonParams{
			FirstName: stripe.String(firstName),
			LastName: stripe.String(lastName),
		},
	}

	_, err := account.Update(
		accountId,
		params,
	)

	return err
}


func UpdateIndividualDOB(
	accountId string,
	day int,
	month int,
	year int,
) (error) {
	stripe.Key = stripeSecretKey()
	
	// 1. Update account details with business profile
	params := &stripe.AccountParams{
		Individual: &stripe.PersonParams{
			DOB: &stripe.PersonDOBParams{
				Day: stripe.Int64(int64(day)),
				Month: stripe.Int64(int64(month)),
				Year: stripe.Int64(int64(year)),
			},
		},
	}

	_, err := account.Update(
		accountId,
		params,
	)

	return err
}

func UpdateIndividualAddress(
	accountId string,
	line1 string,
	line2 string,
	postalCode string,
	city string,
) (error) {
	stripe.Key = stripeSecretKey()
	
	params := &stripe.AccountParams{
		Individual: &stripe.PersonParams{
			Address: &stripe.AddressParams{
				Line1: stripe.String(line1),
				Line2: stripe.String(line2),
				PostalCode: stripe.String(postalCode),
				City: stripe.String(city),
			},
		},
	}

	_, err := account.Update(
		accountId,
		params,
	)

	return err
}


func UpdateIndividualMobile(
	accountId string,
	dialingCode string,
	number string,
) (error) {
	stripe.Key = stripeSecretKey()
	
	params := &stripe.AccountParams{
		Individual: &stripe.PersonParams{
			Phone: stripe.String("+" + dialingCode + number),
		},
	}

	_, err := account.Update(
		accountId,
		params,
	)

	return err
}




