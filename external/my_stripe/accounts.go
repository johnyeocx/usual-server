package my_stripe

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"time"

	"github.com/johnyeocx/usual/server/db/models"
	"github.com/stripe/stripe-go/v74"
	"github.com/stripe/stripe-go/v74/account"
	"github.com/stripe/stripe-go/v74/bankaccount"
	"github.com/stripe/stripe-go/v74/file"
	"github.com/stripe/stripe-go/v74/token"
)

var (
	stripeSecretKey = func () string {
		return os.Getenv("STRIPE_SECRET_KEY")
	}
)

func CreateBasicConnectedAccount(
	country string, 
	email string, 
	ipAddress string,
) (*string, error) {

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
		BusinessType: stripe.String("individual"),
	};

	result, err := account.New(params);
	if err != nil {
		return nil, err
	}

	return &result.ID, nil
}

func CreateConnectedAccount(
	country string, 
	email string, 
	ipAddress string,
	mcc string,
	businessUrl string,
	user *models.Person,
) (*string, error) {

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
			MCC: stripe.String(mcc),
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


func UpdateAccountBusinessProfile(
	stripeId string,
	email string, 
	mcc string,
	businessUrl string,
	user *models.Person,
) (error) {

	stripe.Key = stripeSecretKey()

	params := &stripe.AccountParams{

		BusinessProfile: &stripe.AccountBusinessProfileParams{
			MCC: stripe.String(mcc),
			URL: stripe.String(businessUrl),
		},

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

	_, err := account.Update(stripeId, params);
	return err
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

func UpdateIndividualInfo(
	accountId string,
	firstName string,
	lastName string,
	dialingCode string,
	mobileNumber string,
	dob models.Date,
) (error){
	stripe.Key = stripeSecretKey()

	params := &stripe.AccountParams{
		BusinessType: stripe.String("individual"),

		Individual: &stripe.PersonParams{
			FirstName: stripe.String(firstName),
			LastName: stripe.String(lastName),
			DOB: &stripe.PersonDOBParams{
				Day: stripe.Int64(int64(dob.Day)),
				Month: stripe.Int64(int64(dob.Month)),
				Year: stripe.Int64(int64(dob.Year)),
			},

			Phone: stripe.String("+" + dialingCode + mobileNumber),
		},
	};

	_, err := account.Update(accountId, params);
	return err
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

func UpdateAccountBankInfo(
	accountId string, 
	bankInfo models.BankInfo,
	country string,
) (*stripe.BankAccount, error) {
	stripe.Key = stripeSecretKey()


	t, err := token.New(&stripe.TokenParams{
		BankAccount: &stripe.BankAccountParams{
			Country: &country,
			Currency: stripe.String(string(stripe.CurrencyGBP)),
			AccountHolderName: &bankInfo.AccountHolder,
			// AccountHolderType: stripe.String(string(stripe.BankAccountAccountHolderTypeIndividual)),
			RoutingNumber: &bankInfo.RoutingNumber,
			AccountNumber: &bankInfo.AccountNumber,
		},
	})
	if err != nil {
		return nil, err
	}

	bankParams := &stripe.BankAccountParams{
		Account: stripe.String(accountId),
		Token: stripe.String(t.ID),
		DefaultForCurrency: stripe.Bool(true),
	}
	
	ba, err := bankaccount.New(bankParams)
	if err != nil {
		return nil, err
	}

	return ba, nil
}

func UploadIdentityDocument(
	accountId string, 
	frontFile multipart.File,
	backFile *multipart.File,
) (error) {
	stripe.Key = stripeSecretKey()

	params := &stripe.FileParams{
		FileReader: frontFile,
		Filename: stripe.String("front.jpg"),
		Purpose: stripe.String(string(stripe.FilePurposeIdentityDocument)),
		
	}
	front, err := file.New(params)
	if err != nil {
		return err
	}

	var back *stripe.File
	if (backFile != nil) {
		backParams := &stripe.FileParams{
			FileReader: frontFile,
			Filename: stripe.String("back.jpg"),
			Purpose: stripe.String(string(stripe.FilePurposeIdentityDocument)),
			
		}
		back, err = file.New(backParams)
		if err != nil {
			return err
		}
	}
	
	var veriDocuParams stripe.PersonVerificationDocumentParams

	if (backFile != nil ) {
		veriDocuParams = stripe.PersonVerificationDocumentParams {
			Front: stripe.String(front.ID),
			Back: stripe.String(back.ID),
		}
	} else {
		veriDocuParams = stripe.PersonVerificationDocumentParams {
			Front: stripe.String(front.ID),
		}
	}
	

	accUpdateParams := stripe.AccountParams{
		Individual: &stripe.PersonParams{
			Verification: &stripe.PersonVerificationParams{
				Document: &veriDocuParams,
			},
		},
	}
	
	_, err = account.Update(accountId, &accUpdateParams)
	return err
}

func DeleteConnectedAccounts(
) {

	stripe.Key = stripeSecretKey()
	params := &stripe.AccountListParams{}
	i := account.List(params)
	for i.Next() {
		a := i.Account()
		account.Del(a.ID, nil)
	}
}
