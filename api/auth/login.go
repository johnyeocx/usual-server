package auth

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/secure"
)

func login(sqlDB *sql.DB, email string, password string) (
	*models.Business, *models.Person, *models.RequestError,
) {

	// 1. Authenticate
	authDB := db.AuthDB{DB: sqlDB}
	hashedPassword, err := authDB.GetHashedPassword(email)
	if err != nil {
		return nil,nil, &models.RequestError{
			Err: fmt.Errorf("failed to get hashed password from email\n%v", err),
			StatusCode: http.StatusBadRequest,
		}
	}
	
	// 2. Check if password matches
	matches := secure.StringMatchesHash(password, *hashedPassword)
	
	if !matches {
		return nil,nil,  &models.RequestError{
			Err: fmt.Errorf("password invalid\n%v", err),
			StatusCode: http.StatusUnauthorized,
		}
	}

	// 3. Get business
	businessDB := db.BusinessDB{DB: sqlDB}
	business, err := businessDB.GetBusinessByEmail(email)
	if err != nil {
		return nil, nil, &models.RequestError{
			Err: fmt.Errorf("failed to get business from email\n%v", err),
			StatusCode: http.StatusBadRequest,
		}
	}

	// 4. Get individual
	if business.IndividualID == nil {
		return business, nil, nil
	}
	individual, err := businessDB.GetIndividualByID(*business.IndividualID)
	if err != nil {
		return nil, nil, &models.RequestError{
			Err: fmt.Errorf("failed to get individual from individual id\n%v", err),
			StatusCode: http.StatusBadRequest,
		}
	}

	return business, individual, nil
}