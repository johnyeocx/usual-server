package c_auth

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/johnyeocx/usual/server/constants"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
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
	c := db.CustomerDB{DB: sqlDB}
	cusId, hashedPassword, err := c.GetCustomerHashedPassword(email)
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