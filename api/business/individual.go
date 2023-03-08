package business

import (
	"database/sql"
	"net/http"

	busdb "github.com/johnyeocx/usual/server/db/bus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
)

// UPDATES
func updateIndividualName(
	sqlDB *sql.DB,
	businessId int,
	firstName string,
	lastName string,
) (*models.RequestError) {

	// 1. get stripe id from db
	b := busdb.BusinessDB{DB: sqlDB}

	stripeId, err := b.GetBusinessStripeID(businessId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	indivId, err := b.GetIndividualID(businessId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// CHECK IF NAME IS TAKEN
	// b.GetBusinessByEmail()

	// 2. update stripe
	err = my_stripe.UpdateIndividualName(*stripeId, firstName, lastName)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	// 3. update sql
	err = b.SetIndividualName(*indivId, firstName, lastName)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	return nil
}

func updateIndividualDOB(
	sqlDB *sql.DB,
	businessId int,
	day int,
	month int,
	year int,
) (error) {

	// 1. get stripe id from db
	b := busdb.BusinessDB{DB: sqlDB}

	stripeId, err := b.GetBusinessStripeID(businessId)
	if err != nil {
		return err
	}

	indivId, err := b.GetIndividualID(businessId)
	if err != nil {
		return err
	}

	// 2. update stripe
	err = my_stripe.UpdateIndividualDOB(*stripeId, day, month, year)
	if err != nil {
		return err
	}
	
	// 3. update sql
	err = b.SetIndividualDOB(*indivId, day, month, year)
	return err
}

func updateIndividualAddress(
	sqlDB *sql.DB,
	businessId int,
	line1 string,
	line2 string,
	postalCode string,
	city string,
) (error) {

	// 1. get stripe id from db
	b := busdb.BusinessDB{DB: sqlDB}

	stripeId, err := b.GetBusinessStripeID(businessId)
	if err != nil {
		return err
	}

	indivId, err := b.GetIndividualID(businessId)
	if err != nil {
		return err
	}

	// 2. update stripe
	err = my_stripe.UpdateIndividualAddress(*stripeId, line1, line2, postalCode, city)
	if err != nil {
		return err
	}
	
	// 3. update sql
	err = b.SetIndividualAddress(*indivId,  line1, line2, postalCode, city)
	return err
}

func updateIndividualMobile(
	sqlDB *sql.DB,
	businessId int,
	dialingCode string,
	number string,
) (error) {

	// 1. get stripe id from db
	b := busdb.BusinessDB{DB: sqlDB}

	stripeId, err := b.GetBusinessStripeID(businessId)
	if err != nil {
		return err
	}

	indivId, err := b.GetIndividualID(businessId)
	if err != nil {
		return err
	}

	// 2. update stripe
	err = my_stripe.UpdateIndividualMobile(*stripeId, dialingCode, number)
	if err != nil {
		return err
	}
	
	// 3. update sql
	err = b.SetIndividualMobile(*indivId,  dialingCode, number)
	return err
}