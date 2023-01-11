package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"regexp"
	"strconv"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/secure"
)

// REGISTER
func createBusiness(
	sqlDB *sql.DB, 
	business *models.BusinessDetails,
) (*int64, *models.RequestError) {
	
	// 1. check that email is not already taken
	businessDB := db.BusinessDB{DB: sqlDB}
	_, err := businessDB.GetBusinessByEmail(business.Email)

	if err != sql.ErrNoRows {
		return nil, &models.RequestError{
			Err: fmt.Errorf("account already created"),
			StatusCode: http.StatusConflict,
		}
	}
	
	// businessDB.
	// 2. check that email is valid
	if !isEmailValid(business.Email) {
		return nil, &models.RequestError{
			Err: fmt.Errorf("invalid email"),
			StatusCode: http.StatusBadRequest,
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

func isEmailValid(e string) bool {
    emailRegex := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
    return emailRegex.MatchString(e)
}

func refreshToken(sqlDB *sql.DB, refreshToken string) (*int, error) {
	businessId, err := secure.ParseRefreshToken(refreshToken)
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

// VALIDATION