package customer

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/johnyeocx/usual/server/api/auth"
	"github.com/johnyeocx/usual/server/constants"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/media"
	"github.com/johnyeocx/usual/server/external/my_stripe"
	"github.com/johnyeocx/usual/server/utils/secure"
)

func updateCusName(
	sqlDB *sql.DB,
	cusId int,
	name string,
) (*models.RequestError) {

	// 1. get stripe id from db
	c := db.CustomerDB{DB: sqlDB}
	stripeId, err := c.GetCustomerStripeId(cusId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// // 2. update stripe profile
	err = my_stripe.UpdateCusName(*stripeId, name)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 3. update sql
	err = c.UpdateCusName(cusId, name)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	return nil
}

func sendUpdateEmailVerification(
	sqlDB *sql.DB,
	cusId int,
	newEmail string,
) (*models.RequestError){
	c := db.CustomerDB{DB: sqlDB}

	// check email valid
	if !constants.EmailValid(newEmail) {
		return &models.RequestError{
			Err: errors.New("invalid customer email"),
			StatusCode: http.StatusBadRequest,
		}
	}

	// step 1: Check if user already exists
	verified, err := c.GetCustomerEmailVerified(newEmail)
	if err != nil && err != sql.ErrNoRows{
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	} else if verified {
		return &models.RequestError{
			Err: errors.New("email already exists"),
			StatusCode: http.StatusConflict,
		}
	}

	// step 1: get cus name
	cus, err := c.GetCustomerByID(cusId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadRequest,
		}
	}

	// step 2: send verification email
	otp, reqErr := auth.GenerateEmailOTP(
		sqlDB, 
		newEmail, 
		constants.OtpTypes.UpdateCusEmail,
	)
	if reqErr != nil {
		return reqErr
	}
	
	err = media.SendEmailVerification(newEmail, cus.Name, *otp)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return nil
}

func verifyUpdateCusEmail(
	sqlDB *sql.DB,
	cusId int,
	otp string,
	email string,
) (*models.RequestError) {

	// 1. verify email otp
	emailOtp, reqErr := auth.VerifyEmailOTP(sqlDB, email, otp, constants.OtpTypes.UpdateCusEmail)
	if reqErr != nil {
		return reqErr
	}


	// 1. get stripe id from db
	c := db.CustomerDB{DB: sqlDB}
	stripeId, err := c.GetCustomerStripeId(cusId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// // 2. update stripe profile
	err = my_stripe.UpdateCusEmail(*stripeId, emailOtp.Email)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 3. update sql
	err = c.UpdateCusEmail(cusId, emailOtp.Email)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	return nil
}

func updateCusAddress(
	sqlDB *sql.DB,
	cusId int,
	address models.Address,
) (*models.RequestError) {

	// 1. get stripe id from db
	c := db.CustomerDB{DB: sqlDB}
	stripeId, err := c.GetCustomerStripeId(cusId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// // 2. update stripe profile
	err = my_stripe.UpdateCusAddress(*stripeId, address)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	err = c.UpdateCusAddress(cusId, address)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	return nil
}

func updateCusPassword(
	sqlDB *sql.DB,
	cusId int,
	oldPassword string,
	newPassword string,
) (*models.RequestError) {

	// 1. get stripe id from db
	c := db.CustomerDB{DB: sqlDB}
	oldPasswordHash, err := c.GetCusPasswordFromID(cusId)
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

	err = c.UpdateCusPassword(cusId, newPassHash)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	return nil
}