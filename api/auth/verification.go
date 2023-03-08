package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/johnyeocx/usual/server/db"
	busdb "github.com/johnyeocx/usual/server/db/bus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/otp"
	"github.com/johnyeocx/usual/server/utils/secure"
)

var (
    registerExpiry = time.Minute * 5
)

func GenerateEmailOTP(
    db *sql.DB, 
    email string,
    otpType string,
) (*string, *models.RequestError) {

    // 1. check if expiry otp already exists
    deleteStatement := `
        DELETE from email_otp WHERE
        email=$1 AND type=$2
    `
    db.Exec(deleteStatement, email, otpType)

    // 2. generate otp
    otp := otp.GenerateOTP(6)
    hashedOtp, err := secure.GenerateHashFromStr(otp)
    if err != nil {
        return nil, &models.RequestError{
            Err: fmt.Errorf("failed to hash otp: %v", err),
            StatusCode: http.StatusBadGateway,
        }
    }

    // 3. insert verification into sql table
    expiry := time.Now().Add(registerExpiry).UTC()
    insertStatement := `
        INSERT INTO email_otp 
        (email, type, hashed_otp, expiry) 
        VALUES($1, $2, $3, $4)
    `
	_, err = db.Exec(
        insertStatement, 
        email, 
        otpType,
        hashedOtp,
        expiry,
    )

	if err != nil {
		return nil, &models.RequestError{
            Err: fmt.Errorf("failed to insert otp into sql: %v", err),
            StatusCode: http.StatusBadGateway,
        }
	}

    return &otp, nil
}


func VerifyEmailOTP(
    sqlDB *sql.DB, 
    email string,
    otp string,
    otpType string,
) (*models.EmailOTP, *models.RequestError) {

    // 1. Find matching verification in sql
    authDB := db.AuthDB{DB: sqlDB}
    emailOtp, err := authDB.GetEmailVerification(email, otpType)

    if err != nil {
        return nil, &models.RequestError{
            Err: fmt.Errorf("failed to get verification from db\n%v", err),
            StatusCode: http.StatusBadRequest,
        }
    }

    // 2. Check otp match
    if !secure.StringMatchesHash(otp, emailOtp.HashedOTP) {
        return nil, &models.RequestError{
            Err: fmt.Errorf("invalid otp provided\n%v", err),
            StatusCode: http.StatusForbidden,
        };
    }

    // 3. Delete verification code from table
    if err := authDB.DeleteEmailVerification(email, otpType); err != nil {
        return nil, &models.RequestError{
            Err: fmt.Errorf("failed to delete email verification\n%v", err),
            StatusCode: http.StatusBadGateway,
        };
    }

    return emailOtp, nil  
}


func VerifyCustomerEmailOTP(
    sqlDB *sql.DB, 
    email string,
    otp string,
) (*models.Business, *models.RequestError) {

    // 1. Find matching verification in sql
    authDB := db.AuthDB{DB: sqlDB}
    emailOtp, err := authDB.GetEmailVerification(email, "register")

    if err != nil {
        return nil, &models.RequestError{
            Err: fmt.Errorf("failed to get verification from db\n%v", err),
            StatusCode: http.StatusBadRequest,
        }
    }

    // 2. Check otp match
    if !secure.StringMatchesHash(otp, emailOtp.HashedOTP) {
        return nil, &models.RequestError{
            Err: fmt.Errorf("invalid otp provided\n%v", err),
            StatusCode: http.StatusUnauthorized,
        };
    }

    // 3. Delete verification code from table
    if err := authDB.DeleteEmailVerification(email, "register"); err != nil {
        return nil, &models.RequestError{
            Err: fmt.Errorf("failed to delete email verification\n%v", err),
            StatusCode: http.StatusBadGateway,
        };
    }

    // 4. Success, set user email verified
    if err = authDB.SetBusinessVerified(email, true); err != nil {
        return nil, &models.RequestError{
            Err: fmt.Errorf("failed to set business email verified\n%v", err),
            StatusCode: http.StatusBadGateway,
        };
    }

    // 5. Get business by email
    businessDB := busdb.BusinessDB{DB: sqlDB}
    business, err := businessDB.GetBusinessByEmail(email); 
    if err != nil {
        return nil, &models.RequestError{
            Err: fmt.Errorf("failed to get business from db\n%v", err),
            StatusCode: http.StatusBadGateway,
        };
    }
    return business, nil  
}