package busdb

import (
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/stripe/stripe-go/v74"
)

func (businessDB *BusinessDB) SetBusinessStripeID(
	businessId int,
	stripeId string,
) (error) {
	_, err := businessDB.DB.Exec(
		`UPDATE business SET stripe_id=$1 WHERE business_id=$2`, 
		stripeId, 
		businessId,
	)
	if err != nil {
		return err
	}

	return nil
}


func (businessDB *BusinessDB) SetBusinessDescription(
	businessId int,
	description string,
) (error) {
	_, err := businessDB.DB.Exec(
		`UPDATE business SET description=$1 WHERE business_id=$2`, 
		description, 
		businessId,
	)
	if err != nil {
		return err
	}

	return nil
}

func (b *BusinessDB) SetBusinessCategory(
	businessId int,
	category string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET business_category=$1 WHERE business_id=$2`, 
		category, businessId,
	)

	return err
}

func (b *BusinessDB) SetBusinessName(
	businessId int,
	name string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET name=$1 WHERE business_id=$2`, 
		name, businessId,
	)

	return err
}

func (b *BusinessDB) SetBusinessEmail(
	businessId int,
	email string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET email=$1 WHERE business_id=$2`, 
		email, businessId,
	)

	return err
}

func (b *BusinessDB) SetBusinessCountry(
	businessId int,
	countryCode string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET country=$1 WHERE business_id=$2`, 
		countryCode, businessId,
	)

	return err
}

func (b *BusinessDB) SetBusinessUrl(
	businessId int,
	url string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET business_url=$1 WHERE business_id=$2`, 
		url, businessId,
	)

	return err
}

func (b *BusinessDB) UpdateBusinessPassword(
	businessId int,
	passHash string,
) (error) {
	_, err := b.DB.Exec(`UPDATE business SET password=$1 WHERE business_id=$2`, 
		passHash, businessId,
	)

	return err
}


func (b *BusinessDB) UpdateBusinessBankAccount(
	businessId int,
	bankAccount *stripe.BankAccount,
) (*models.BankAccount, error) {

	var bankAccountId int
	err := b.DB.QueryRow(`INSERT into business_bank_account (stripe_id, account_holder_name, bank_name, last4, routing_number, business_id) 
		VALUES($1, $2, $3, $4, $5, $6) RETURNING bank_account_id
	`, bankAccount.ID, bankAccount.AccountHolderName, bankAccount.BankName, bankAccount.Last4, bankAccount.RoutingNumber, businessId).Scan(
		&bankAccountId,
	)

	if err != nil {
		return nil, err
	}

	_, err = b.DB.Exec("UPDATE business SET external_account_id=$1, external_account_type='bank_account' WHERE business_id=$2", bankAccountId, businessId)
	if err != nil {
		return nil, err
	}

	return &models.BankAccount{
		ID: bankAccountId,
		StripeID: bankAccount.ID,
		AccountHolderName: bankAccount.AccountHolderName,
		Last4: bankAccount.Last4,
		RoutingNumber: bankAccount.RoutingNumber,
	}, err
}