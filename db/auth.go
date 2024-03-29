package db

import (
	"database/sql"
	"time"

	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/secure"
)

type AuthDB struct {
	DB *sql.DB
}

func (a *AuthDB) InsertBusinessDetails(business *models.BusinessDetails) (*int64, error) {

	hashedPassword, err := secure.GenerateHashFromStr(business.Password)
	if err != nil {
		return nil, err
	}

	insertStatement := `
		INSERT INTO business (name, country, email, password) VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE 
		SET name=$1, country=$2, password=$4
		RETURNING business_id
	`
	
	var insertedId int64
	err = a.DB.QueryRow(
		insertStatement, 
		business.Name, 
		business.Country,
		business.Email,
		hashedPassword,
	).Scan(&insertedId)

	if err != nil {
		return nil, err
	}

	return &insertedId, nil
}

func (a *AuthDB) GetEmailVerification(email string, verificationType string) (*models.EmailOTP, error) {

	selectStatement := `
		SELECT hashed_otp, email from email_otp WHERE
		email=$1 AND type=$2 AND $3 <= expiry
	`

	row := a.DB.QueryRow(selectStatement, email, verificationType, time.Now().UTC())

	var emailOtp models.EmailOTP
	err := row.Scan(&emailOtp.HashedOTP, &emailOtp.Email);

	if err == sql.ErrNoRows {
		return nil, err
	}

	return &emailOtp, nil
}

func (a *AuthDB) DeleteEmailVerification(email string, verificationType string) (error) {
	deleteStatement := `
        DELETE from email_otp WHERE email=$1 AND type=$2
    `

    if _, err := a.DB.Exec(deleteStatement, email, verificationType); err != nil {
        return err
    }

	return nil
}

func (a *AuthDB) SetBusinessVerified (email string, verified bool) (error) {
	insertStatement := `
		UPDATE business SET email_verified=$1 WHERE email=$2
	`
	
	_, err := a.DB.Exec(insertStatement, verified, email)

	if err != nil {
		return err
	}

	return nil
}

func (a *AuthDB) SetCustomerVerified (email string, verified bool) (error) {
	insertStatement := `
		UPDATE customer SET email_verified=$1 WHERE email=$2
	`
	
	_, err := a.DB.Exec(insertStatement, verified, email)

	if err != nil {
		return err
	}

	return nil
}

func (a *AuthDB) GetHashedPassword( email string) (*string, error) {

	var hashedPassword  string
	err := a.DB.QueryRow(`SELECT password FROM business WHERE email=$1`, email).Scan(&hashedPassword)

	if err != nil {
		return nil, err
	}
	
	return  &hashedPassword, nil
}

func ValidateBusinessId (sqlDB *sql.DB, businessId int) (bool) {
	verified := false

	err := sqlDB.QueryRow("SELECT email_verified=true FROM business WHERE business_id=$1", 
		businessId,
	).Scan(&verified) 
	
	if err != nil {
		return false
	}
	
	return verified
}

func ValidateCustomerId (sqlDB *sql.DB, id int) (bool) {
	var email string

	err := sqlDB.QueryRow("SELECT email FROM customer WHERE customer_id=$1", 
		id,
	).Scan(&email) 
	
	if err != nil {
		return false
	}
	

	return email != ""
}


