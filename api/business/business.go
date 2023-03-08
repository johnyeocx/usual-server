package business

import (
	"database/sql"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"

	"github.com/johnyeocx/usual/server/constants"
	busdb "github.com/johnyeocx/usual/server/db/bus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/errors/bus_errors"
	"github.com/johnyeocx/usual/server/external/my_stripe"
	"github.com/johnyeocx/usual/server/utils/secure"
)


func getBusinessData(
	sqlDB *sql.DB,
	bId int,
) (map[string]interface{}, *models.RequestError) {
	businessDB := busdb.BusinessDB{DB: sqlDB}
	business, payoutTotal, receivedTotal, err := businessDB.GetBusinessAndTotal(bId)

	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	var individual models.Person
	if business.IndividualID != nil {
		indiv, err := businessDB.GetIndividualByID(*business.IndividualID)
		if err != nil && err != sql.ErrNoRows {
			return nil, &models.RequestError{
				Err: err,
				StatusCode: http.StatusBadGateway,
			}
		}

		individual = *indiv
	}

	categories, subProducts, err := GetBusinessProducts(sqlDB, bId)
	if err != nil && err != sql.ErrNoRows{
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	stats, err := getBusinessStats(sqlDB, bId)
	if err != nil && err != sql.ErrNoRows {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	res := map[string]interface{} {
		"business": business,
		"payout_total": payoutTotal,
		"received_total": receivedTotal,
		"individual": individual,
		"product_categories": categories,
		"subscription_products": subProducts,
		"sub_infos": stats["sub_infos"],
		"invoices": stats["invoices"],
		"usage_infos": stats["usage_infos"],
		"bank_accounts": stats["bank_accounts"],
	}

	return res, nil
}

func getBusinessTransactions(
	sqlDB *sql.DB,
	bId int,
) (map[string]interface{}, *models.RequestError) {
	b := busdb.BusinessDB{DB: sqlDB}

	invoices, err := b.GetBusinessInvoices(bId, 50)
	if err != nil {
		return nil, bus_errors.GetInvoicesFailedErr(err)
	}

	payouts, err := b.GetBusinessPayouts(bId, 50)
	if err != nil {
		return nil, bus_errors.GetInvoicesFailedErr(err)
	}

	res := map[string]interface{} {
		"invoices": invoices,
		"payouts": payouts,
	}

	return res, nil
}


func setBusinessProfile(
	sqlDB *sql.DB, 
	businessId int,
	businessCategory string, 
	businessUrl string, 
	individual *models.Person,
	ipAddress string,
) (*models.RequestError) {

	// 1. get business by id
	businessDB := busdb.BusinessDB{DB: sqlDB}
	business, err := businessDB.GetBusinessByID(businessId)
	if err != nil {
		return &models.RequestError{
			Err: errors.New("invalid business id"),
			StatusCode: http.StatusBadRequest,
		}
	}

	// 1. Get MCC associated to category
	foundMcc := false
	var mcc string
	for _, cat := range constants.BusinessCategories {
		if cat["label"] == businessCategory {
			mcc = cat["mcc"].(string)
			foundMcc = true
		}
	}
	

	if (!foundMcc) {
		return &models.RequestError{
			Err: errors.New("invalid business category"),
			StatusCode: http.StatusBadRequest,
		}
	}

	err = my_stripe.UpdateAccountBusinessProfile(
		*business.StripeID,
		business.Email,
		mcc,
		businessUrl,
		individual,
	)

	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}


	// 2. create individual	
	individualId, err := businessDB.CreateIndividual(individual)
	if err != nil {
		return &models.RequestError{
			Err: fmt.Errorf("failed to create individual\n%v", err),
			StatusCode: http.StatusBadGateway,
		}
	}

	err = businessDB.CreateBusinessProfile(
		businessId, 
		businessCategory, 
		businessUrl, 
		*individualId,
	)

	if err != nil {
		return &models.RequestError{
			Err: fmt.Errorf("failed to create business profile\n%v", err),
			StatusCode: http.StatusBadGateway,
		}
	}

	return  nil
}

// UPDATE
func updateBusinessCategory(
	sqlDB *sql.DB,
	businessId int,
	category string,
) (error) {

	// 1. get stripe id from db
	b := busdb.BusinessDB{DB: sqlDB}
	stripeId, err := b.GetBusinessStripeID(businessId)
	if err != nil {
		return err
	}

	foundMcc := false
	var mcc string
	for _, cat := range constants.BusinessCategories {
		if cat["label"] == category {
			mcc = cat["mcc"].(string)
			foundMcc = true
		}
	}

	if (!foundMcc) {
		return errors.New("unknown business category")
	}

	// 2. update stripe profile
	err = my_stripe.UpdateAccountMCC(*stripeId, mcc)
	if err != nil {
		return err
	}

	// 3. update sql
	err = b.SetBusinessCategory(businessId, category)
	return err
}

func updateBusinessName(
	sqlDB *sql.DB,
	businessId int,
	name string,
) (*models.RequestError) {

	// 1. get stripe id from db
	b := busdb.BusinessDB{DB: sqlDB}

	bus, err := b.GetBusinessByName(name)
	if err != nil && err != sql.ErrNoRows {
		return &models.RequestError{
			StatusCode: http.StatusBadGateway,
			Err: err,
		}
	} else if bus != nil && *bus.EmailVerified && bus.Name == name {
		return &models.RequestError{
			StatusCode: http.StatusConflict,
			Err: err,
		}
	}
	
	err = b.SetBusinessName(businessId, name)
	if err != nil && err != sql.ErrNoRows {
		return &models.RequestError{
			StatusCode: http.StatusBadGateway,
			Err: err,
		}
	}
	return nil
}


func updateBusinessPassword(
	sqlDB *sql.DB,
	busId int,
	oldPassword string,
	newPassword string,
) (*models.RequestError) {
	if !constants.PasswordValid(newPassword) {
		return &models.RequestError{
			Err: errors.New("invalid password"),
			StatusCode: http.StatusBadRequest,
		}
	}

	// 1. get stripe id from db
	b := busdb.BusinessDB{DB: sqlDB}
	oldPasswordHash, err := b.GetBusinessPasswordByID(busId)

	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	if !secure.StringMatchesHash(oldPassword, *oldPasswordHash) {
		return &models.RequestError{
			Err: errors.New("passwords don't match"),
			StatusCode: http.StatusNotAcceptable,
		}
	}

	// update password
	newPassHash, err := secure.GenerateHashFromStr(newPassword)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	err = b.UpdateBusinessPassword(busId, newPassHash)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	return nil
}

func updateBusinessEmail(
	sqlDB *sql.DB,
	businessId int,
	email string,
) (error) {

	// 1. get stripe id from db
	b := busdb.BusinessDB{DB: sqlDB}
	stripeId, err := b.GetBusinessStripeID(businessId)
	if err != nil {
		return err
	}

	// 2. update stripe profile
	err = my_stripe.UpdateAccountEmail(*stripeId, email)
	if err != nil {
		return err
	}

	// 3. update sql
	err = b.SetBusinessEmail(businessId, email)
	return err
}


func updateIndividualDetailsStripe(
	sqlDB *sql.DB,
	bId int,
	firstName string,
	lastName string,
	dialingCode string,
	mobileNumber string,
	dob models.Date,
) (error) {

	// 1. get stripe id from db
	b := busdb.BusinessDB{DB: sqlDB}
	stripeId, err := b.GetBusinessStripeID(bId)
	if err != nil {
		return err
	}

	// 2. update stripe profile
	err = my_stripe.UpdateIndividualInfo(*stripeId, firstName, lastName, dialingCode, mobileNumber, dob)

	return err
}

func updateBusinessUrl(
	sqlDB *sql.DB,
	businessId int,
	url string,
) (error) {

	// 1. get stripe id from db

	b := busdb.BusinessDB{DB: sqlDB}
	stripeId, err := b.GetBusinessStripeID(businessId)
	if err != nil {
		return err
	}

	// 2. update stripe profile
	err = my_stripe.UpdateAccountUrl(*stripeId, url)
	if err != nil {
		return err
	}

	// 3. update sql
	err = b.SetBusinessUrl(businessId, url)
	return err
}

func updateBusinessBankInfo(sqlDB *sql.DB, bId int, bankInfo models.BankInfo) (*models.BankAccount, *models.RequestError) {

	b := busdb.BusinessDB{DB: sqlDB}

	bus, err := b.GetBusinessByID(bId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	stripeBankAccount, err := my_stripe.UpdateAccountBankInfo(*bus.StripeID, bankInfo, bus.Country)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}


	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	bankAccount, err := b.UpdateBusinessBankAccount(bId, stripeBankAccount)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	return bankAccount, nil
}

func uploadIndVerificationDoc(sqlDB *sql.DB, bId int, frontFile multipart.File, backFile *multipart.File) (*models.RequestError) {

	b := busdb.BusinessDB{DB: sqlDB}

	bus, err := b.GetBusinessByID(bId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	reqErr := my_stripe.UploadIdentityDocument(*bus.StripeID, frontFile, backFile)
	if reqErr != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	err = b.UpdateIndividualVerificationDocumentRequired(*bus.IndividualID, false)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	return nil
}