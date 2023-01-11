package my_stripe

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/johnyeocx/usual/server/db/models"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
)

var (
	stripeSecretKey = func () string {
		return os.Getenv("STRIPE_SECRET_KEY")
	}
)

func CreateConnectedAccount(
	country string, 
	email string, 
	ipAddress string,
	businessCategory string,
	businessUrl string,
	user *models.Person,
) (*string, error) {
	// hardcode business category
	businessCategory = "5734"

	stripe.Key = stripeSecretKey()
	unixNow := time.Now().Unix()

	params := &stripe.AccountParams{
		Country: stripe.String("GB"),
		Email: stripe.String(email),

		Type: stripe.String(string(stripe.AccountTypeCustom)),
		Capabilities: &stripe.AccountCapabilitiesParams{
			CardPayments: &stripe.AccountCapabilitiesCardPaymentsParams{
				Requested: stripe.Bool(true),
			},
			Transfers: &stripe.AccountCapabilitiesTransfersParams{Requested: stripe.Bool(true)},
		},
		TOSAcceptance: &stripe.AccountTOSAcceptanceParams{
			Date: &unixNow,
			IP: &ipAddress,
		},

		BusinessProfile: &stripe.AccountBusinessProfileParams{
			MCC: stripe.String(businessCategory),
			URL: stripe.String(businessUrl),
		},
		BusinessType: stripe.String("individual"),

		Individual: &stripe.PersonParams{
			FirstName: stripe.String(user.FirstName),
			LastName: stripe.String(user.LastName),
			Email: stripe.String(email),

			Address: &stripe.AddressParams{
				Line1: &user.Address.Line1,
				Line2: &user.Address.Line2,
				PostalCode: &user.Address.PostalCode,
				City: &user.Address.City,
			},

			DOB: &stripe.PersonDOBParams{
				Day: stripe.Int64(int64(user.DOB.Day)),
				Month: stripe.Int64(int64(user.DOB.Month)),
				Year: stripe.Int64(int64(user.DOB.Year)),
			},

			Phone: stripe.String("+" + user.Mobile.DialingCode + user.Mobile.Number),
		},
	};

	result, err := account.New(params);
	if err != nil {
		return nil, err
	}

	return &result.ID, nil
}

func UpdateAccountMCC(
	stripeId string,
	mcc string,
) (error) {
	stripe.Key = stripeSecretKey()

	params := &stripe.AccountParams{
		BusinessProfile: &stripe.AccountBusinessProfileParams{
			MCC: stripe.String(mcc),
		},
	};

	_, err := account.Update(stripeId, params);
	return err
}

func UpdateAccountEmail(
	stripeId string,
	email string,
) (error) {
	stripe.Key = stripeSecretKey()

	params := &stripe.AccountParams{
		Email: stripe.String(email),
		Individual: &stripe.PersonParams{
			Email: stripe.String(email),
		},
	};

	_, err := account.Update(stripeId, params);
	return err
}

func UpdateAccountUrl(
	stripeId string,
	url string,
) (error) {
	stripe.Key = stripeSecretKey()

	params := &stripe.AccountParams{
		BusinessProfile: &stripe.AccountBusinessProfileParams{
			URL: stripe.String(url),
		},
	};

	_, err := account.Update(stripeId, params);
	return err
}

func UpdateBusinessProfile(
	accountId string,
	businessCategory string,
	businessUrl *string,
	accountCard *models.CreditCard,
) (error) {
	stripe.Key = stripeSecretKey()

	// token, err :=  CreateBankToken()
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }

	businessType := "individual"
	businessCategory = "5734"
	
	// 1. Update account details with business profile
	params := &stripe.AccountParams{
		BusinessProfile: &stripe.AccountBusinessProfileParams{
			MCC: &businessCategory,
			URL: businessUrl,
		},
		BusinessType: &businessType,
		// ExternalAccount: &stripe.AccountExternalAccountParams{
		// 	Token: &token.ID,
		// },
	}

	_, err := account.Update(
		accountId,
		params,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}


func UpdateIndividualAccount(
	accountId string, 
	user *models.Person,
) (error) {
	stripe.Key = stripeSecretKey()

	userMobile := user.Mobile.DialingCode + user.Mobile.Number
	fmt.Println(userMobile)
	params := &stripe.AccountParams{
		Individual: &stripe.PersonParams{
			FirstName: stripe.String(user.FirstName),
			LastName: stripe.String(user.LastName),
			Email: stripe.String(user.Email),

			Address: &stripe.AddressParams{
				Line1: &user.Address.Line1,
				Line2: &user.Address.Line2,
				PostalCode: &user.Address.PostalCode,
				City: &user.Address.City,
			},

			DOB: &stripe.PersonDOBParams{
				Day: stripe.Int64(int64(user.DOB.Day)),
				Month: stripe.Int64(int64(user.DOB.Month)),
				Year: stripe.Int64(int64(user.DOB.Year)),
			},

			Phone: stripe.String("8888675309"),
		},
	}

	_, err := account.Update(
		accountId,
		params,
	)

	if err != nil {
		return err
	}
	fmt.Println("Successfully updated company account")

	return nil
}