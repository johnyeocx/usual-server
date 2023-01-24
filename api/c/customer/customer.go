package customer

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/johnyeocx/usual/server/api/auth"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/media"
	"github.com/johnyeocx/usual/server/external/my_stripe"
	"github.com/johnyeocx/usual/server/utils/secure"
)

var (
	otpType = "customer_register"
)

func CreateCustomer(
	sqlDB *sql.DB,
	name string,
	email string,
	password string,
) (*models.RequestError) {
	c := db.CustomerDB{DB: sqlDB}

	// step 1: Check if user already exists
	verified, err := c.GetCustomerEmailVerified(email)
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

	// if no rows
	if err != nil && err == sql.ErrNoRows {
		_, err = c.CreateCustomer(name, email, password)
		if err != nil {
			return &models.RequestError{
				Err: err,
				StatusCode: http.StatusBadGateway,
			}
		}
	}

	// step 3: send verification email
	otp, reqErr := auth.GenerateEmailOTP(sqlDB, email, "customer_register")
	if reqErr != nil {
		return reqErr
	}
	
	err = media.SendEmailVerification(email, name, *otp)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return nil
}

func VerifyCustomerEmail(
	sqlDB *sql.DB,
	email string,
	otp string,
) (map[string]string, *models.RequestError) {

	// 1. verify email otp
	reqErr := auth.VerifyEmailOTP(sqlDB, email, otp, otpType)
	if reqErr != nil {
		return nil, reqErr
	}

	// 2. Success, set user email verified
	authDB := db.AuthDB{DB: sqlDB}
	if err := authDB.SetCustomerVerified(email, true); err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadRequest,
		}
	}

	// 3. Get business by email
	c := db.CustomerDB{DB: sqlDB}
	cus, err := c.GetCustomerByEmail(email)
	if err != nil {		
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 4. Create stripe customer
	cusStripeId, err := my_stripe.CreateCustomerNoPayment(cus)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 5. Set customer stripe id in sql
	err = c.InsertCustomerStripeID(cus.ID, *cusStripeId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 6. Return jwt token
	accessToken, err := secure.GenerateAccessToken(strconv.Itoa(cus.ID))
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	refreshToken, err := secure.GenerateRefreshToken(strconv.Itoa(cus.ID))
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return map[string]string{
		"access_token": accessToken,
		"refresh_token": refreshToken,
	}, nil
}

func CreateCFromSubscribe(
	sqlDB *sql.DB,
	name string,
	email string,
	card *models.CreditCard,
) (*int, *models.RequestError) {

	// 1. check if customer already created
	c := db.CustomerDB{DB: sqlDB}
	_, err := c.GetCustomerByEmail(email) 

	if err == nil || err != sql.ErrNoRows {
		return  nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusConflict,
		}
	}
	
	newC := models.Customer {
		Name: name,
		Email: email,
	}
	// 2. Create stripe customer
	stripeId, err := my_stripe.CreateCustomer(&newC, card)
	if err != nil {
		errMap := map[string]interface{}{}
		json.Unmarshal([]byte(err.Error()), &errMap)
		fmt.Println(errMap["code"])

		// handle card declined
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 3. Insert into db
	cId, err := c.CreateCFromSubscribe(name, email, *stripeId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	} else {
		return cId, nil
	}
}

func GetCustomer(
	sqlDB *sql.DB,
	cusId int,
) (map[string]interface{}, *models.RequestError) {

	c := db.CustomerDB{DB: sqlDB}
	
	cus, err := c.GetCustomerByID(cusId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return map[string]interface{}{
		"customer": cus,
	}, nil
}