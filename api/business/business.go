package business

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/johnyeocx/usual/server/constants"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
)

func getBusinessStats(
	sqlDB *sql.DB,
	businessId int,
) ( *[]models.Customer, error) {
	b := db.BusinessDB{DB: sqlDB}
	subscribers, err := b.GetBusinessSubscribers(businessId)
	if err != nil {
		return nil, err
	}

	return subscribers, nil
}

func setBusinessProfile(
	sqlDB *sql.DB, 
	businessId int,
	businessCategory string, 
	businessUrl string, 
	individual *models.Person,
	stripeId 	string,
) (*int, *models.RequestError) {
	// 1. create individual
	businessDB := db.BusinessDB{DB: sqlDB}
	
	individualId, err := businessDB.CreateIndividual(individual)
	if err != nil {
		return nil, &models.RequestError{
			Err: fmt.Errorf("failed to create individual\n%v", err),
			StatusCode: http.StatusBadGateway,
		}
	}

	err = businessDB.CreateBusinessProfile(
		businessId, 
		businessCategory, 
		businessUrl, 
		*individualId,
		stripeId,
	)

	if err != nil {
		return nil, &models.RequestError{
			Err: fmt.Errorf("failed to create business profile\n%v", err),
			StatusCode: http.StatusBadGateway,
		}
	}

	return individualId, nil
}

// UPDATE
func updateBusinessCategory(
	sqlDB *sql.DB,
	businessId int,
	category string,
) (error) {

	// 1. get stripe id from db
	b := db.BusinessDB{DB: sqlDB}
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
	category string,
) (error) {

	// 1. get stripe id from db
	b := db.BusinessDB{DB: sqlDB}
	
	// 2. update sql
	err := b.SetBusinessName(businessId, category)
	return err
}

func updateBusinessEmail(
	sqlDB *sql.DB,
	businessId int,
	email string,
) (error) {

	// 1. get stripe id from db
	b := db.BusinessDB{DB: sqlDB}
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

func updateBusinessUrl(
	sqlDB *sql.DB,
	businessId int,
	url string,
) (error) {

	// 1. get stripe id from db
	b := db.BusinessDB{DB: sqlDB}
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