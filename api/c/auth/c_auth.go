package c_auth

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/johnyeocx/usual/server/constants"
	my_enums "github.com/johnyeocx/usual/server/constants/enums"
	"github.com/johnyeocx/usual/server/db"
	cusdb "github.com/johnyeocx/usual/server/db/cus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
	"github.com/johnyeocx/usual/server/utils/secure"
)

func refreshToken(sqlDB *sql.DB, refreshToken string) (*int, error) {
	cusId, cusType, err := secure.ParseRefreshToken(refreshToken)
	if cusType != constants.UserTypes.Customer {
		return nil, errors.New("unauthorized user")
	}

	if err != nil {
		return nil, err
	}
	
	customerIdInt, err := strconv.Atoi(cusId)
	if err != nil {
		return nil, err
	}
	
	if ok := db.ValidateCustomerId(sqlDB, customerIdInt); !ok {
		return nil, fmt.Errorf("invalid business id")
	}

	return &customerIdInt, nil
}

func login(
	sqlDB *sql.DB, 
	email string, 
	password string,
) (map[string]interface{}, *models.RequestError) {

	// 1. Get hashed password
	c := cusdb.CustomerDB{DB: sqlDB}

	cusId, hashedPassword, err := c.GetCusPasswordFromEmail(email)
	if err != nil {
		return nil, &models.RequestError{
			Err: fmt.Errorf("failed to get hashed password from email\n%v", err),
			StatusCode: http.StatusBadRequest,
		}
	}
	
	// 2. Check if password matches
	matches := secure.StringMatchesHash(password, *hashedPassword)
	
	if !matches {
		return nil, &models.RequestError{
			Err: fmt.Errorf("password invalid\n%v", err),
			StatusCode: http.StatusUnauthorized,
		}
	}


	accessToken, refreshToken, err := secure.GenerateTokensFromId(*cusId, "customer")
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}


	return map[string]interface{}{
		"access_token": accessToken,
		"refresh_token": refreshToken,
	}, nil
}

func ExternalSignIn(
	sqlDB *sql.DB,
	email string,
	signInProvider my_enums.CusSignInProvider,
) (map[string]interface{}, *models.RequestError) {
	c := cusdb.CustomerDB{DB: sqlDB}
	
	// 1. Check if email already exists. If 
	cus, err := c.GetCustomerByEmail(email)


	if err != nil && err != sql.ErrNoRows{
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 1. If email verified
	if cus != nil && *cus.EmailVerified && cus.SignInProvider != signInProvider {
		// wrong sign in method
		return map[string]interface{}{
			"signin_provider": cus.SignInProvider,
		}, &models.RequestError{
			Err: errors.New("wrong sign-in method"),
			StatusCode: http.StatusForbidden,
		}
	}

	var cId *int
	if err == sql.ErrNoRows {
		cusUuid := uuid.New()
		stripeId , err := my_stripe.CreateCustomerNoPayment(&models.Customer{
			FirstName: "",
			LastName: "",
			Email: email,
		})
		
		if err != nil {
			return nil, &models.RequestError{
				Err: err,
				StatusCode: http.StatusBadGateway,
			}
		}
		

		cId, err = c.CreateCustomerFromExtSignin(email, cusUuid.String(), *stripeId, signInProvider)
		if err != nil {
			return nil, &models.RequestError{
				Err: err,
				StatusCode: http.StatusBadGateway,
			}
		}
		// return auth tokens
	} else {
		cId = &cus.ID
	}


	accessToken, refreshToken, err := secure.GenerateTokensFromId(*cId, constants.UserTypes.Customer)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return map[string]interface{}{
		"access_token": accessToken,
		"refresh_token": refreshToken,
	}, nil
}
