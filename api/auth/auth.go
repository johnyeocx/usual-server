package auth

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/johnyeocx/usual/server/constants"
	"github.com/johnyeocx/usual/server/db"
	busdb "github.com/johnyeocx/usual/server/db/bus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/media"
	"github.com/johnyeocx/usual/server/utils/secure"
)

// REGISTER
func createBusiness(
	sqlDB *sql.DB, 
	business *models.BusinessDetails,
) (*int64, *models.RequestError) {
	
	// 1. check that email is not already taken
	b := busdb.BusinessDB{DB: sqlDB}

	// 2. check that email is valid
	if !constants.EmailValid(business.Email) || !constants.PasswordValid(business.Password) {
		return nil, &models.RequestError{
			Err: fmt.Errorf("invalid email"),
			StatusCode: http.StatusBadRequest,
		}
	}

	// step 1: Check if email already taken
	verified, err := b.GetBusinessEmailVerified(business.Email)
	if err != nil && err != sql.ErrNoRows{
		// ERROR
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	} else if verified {
		// TAKEN
		return nil, &models.RequestError{
			Err: errors.New("email already exists"),
			StatusCode: http.StatusConflict,
		}
	}

	
	// 3. insert into DB
	authDB := db.AuthDB{DB: sqlDB}
	id, err := authDB.InsertBusinessDetails(business)
	if err != nil {
		return nil, &models.RequestError{
			Err: fmt.Errorf("failed to insert into db: \n%v", err),
			StatusCode: http.StatusBadGateway,
		}
	}

	return id, nil
}

func refreshToken(sqlDB *sql.DB, refreshToken string) (*int, error) {
	businessId, userType, err := secure.ParseRefreshToken(refreshToken)

	if userType != constants.UserTypes.Business {
		return nil, err
	}
	
	if err != nil {
		return nil, err
	}
	
	businessIdInt, err := strconv.Atoi(businessId)
	if err != nil {
		return nil, err
	}
	
	if ok := db.ValidateBusinessId(sqlDB, businessIdInt); !ok {
		return nil, fmt.Errorf("invalid business id")
	}

	return &businessIdInt, nil
}

func sendBusRegEmailOTP(
	sqlDB *sql.DB,
	newEmail string,
) (*models.RequestError){

	b := busdb.BusinessDB{DB: sqlDB}

	// check email valid
	if !constants.EmailValid(newEmail) {
		return &models.RequestError{
			Err: errors.New("invalid email"),
			StatusCode: http.StatusBadRequest,
		}
	}

	// step 1: Check if email already taken
	verified, err := b.GetBusinessEmailVerified(newEmail)
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
	bus, err := b.GetBusinessByEmail(newEmail)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadRequest,
		}
	}

	// step 2: send verification email
	otp, reqErr := GenerateEmailOTP(
		sqlDB, 
		newEmail, 
		constants.OtpTypes.RegisterBusEmail,
	)
	if reqErr != nil {
		return reqErr
	}

	err = media.SendEmailVerification(newEmail, bus.Name, *otp)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return nil
}