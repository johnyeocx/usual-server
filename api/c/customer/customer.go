package customer

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/uuid"
	"github.com/johnyeocx/usual/server/api/auth"
	"github.com/johnyeocx/usual/server/constants"
	my_enums "github.com/johnyeocx/usual/server/constants/enums"
	"github.com/johnyeocx/usual/server/db"
	cusdb "github.com/johnyeocx/usual/server/db/cus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/media"
	"github.com/johnyeocx/usual/server/external/my_stripe"
	"github.com/johnyeocx/usual/server/passes"
	"github.com/johnyeocx/usual/server/utils/secure"
)

// var (
// 	otpType = "customer_register"
// )

func CreateCustomer(
	sqlDB *sql.DB,
	firstName string,
	lastName string,
	email string,
	password string,
) (*models.RequestError) {
	c := cusdb.CustomerDB{DB: sqlDB}

	// check email valid
	if !constants.EmailValid(email) || !constants.PasswordValid(password) {
		return &models.RequestError{
			Err: errors.New("invalid customer email"),
			StatusCode: http.StatusBadRequest,
		}
	}

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

	// 1. generate random 16 digit number
	uuid := uuid.NewString()	

	// if no rows
	_, err = c.CreateCustomer(firstName, lastName, email, password, uuid, my_enums.Custom)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// step 3: send verification email
	otp, reqErr := auth.GenerateEmailOTP(sqlDB, email, "customer_register")
	if reqErr != nil {
		return reqErr
	}
	
	err = media.SendEmailVerification(email, firstName, *otp)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return nil
}

func VerifyCustomerRegEmail(
	s3sess *session.Session,
	sqlDB *sql.DB,
	email string,
	otp string,
) (map[string]string, *models.RequestError) {

	// 1. verify email otp
	_, reqErr := auth.VerifyEmailOTP(sqlDB, email, otp, constants.OtpTypes.RegisterCusEmail)
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
	c := cusdb.CustomerDB{DB: sqlDB}
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

	// 6. Add pkpass and image to cloud
	fullName := cus.FirstName + " " + cus.LastName
	err = passes.GenerateCustomerPass(s3sess, fullName, cus.Uuid, cus.ID)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	err = media.GenerateCusQR(s3sess, cus.Uuid, cus.ID)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 7. Return jwt token
	accessToken, err := secure.GenerateAccessToken(strconv.Itoa(cus.ID), "customer")
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	refreshToken, err := secure.GenerateRefreshToken(strconv.Itoa(cus.ID), "customer")
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

func sendRegEmailOTP(
	sqlDB *sql.DB,
	newEmail string,
) (*models.RequestError){
	c := cusdb.CustomerDB{DB: sqlDB}

	// check email valid
	if !constants.EmailValid(newEmail) {
		return &models.RequestError{
			Err: errors.New("invalid email"),
			StatusCode: http.StatusBadRequest,
		}
	}

	// step 1: Check if email already taken
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
	cus, err := c.GetCustomerByEmail(newEmail)
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
		constants.OtpTypes.RegisterCusEmail,
	)
	if reqErr != nil {
		return reqErr
	}

	err = media.SendEmailVerification(newEmail, cus.FirstName, *otp)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return nil
}

func GetCustomerData(
	sqlDB *sql.DB,
	cusId int,
) (map[string]interface{}, *models.RequestError) {

	c := cusdb.CustomerDB{DB: sqlDB}
	
	cus, err := c.GetCustomerByID(cusId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}



	subs, err := c.GetCustomerSubscriptions(cusId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}



	// get customer cards
	cards, err := c.GetCustomerCards(cusId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// get invoices
	invoices, err := c.GetCustomerInvoices(cusId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}


	return map[string]interface{}{
		"customer": cus,
		"subscriptions": subs,
		"cards": cards,
		"invoices": invoices,
	}, nil
}

func GetCusSubsAndInvoices(
	sqlDB *sql.DB,
	cusId int,
) (map[string]interface{}, *models.RequestError) {

	c := cusdb.CustomerDB{DB: sqlDB}

	subs, err := c.GetCustomerSubscriptions(cusId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// get invoices
	invoices, err := c.GetCustomerInvoices(cusId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}


	return map[string]interface{}{
		"subscriptions": subs,
		"invoices": invoices,
	}, nil
}

func AddCusCreditCard(
	sqlDB *sql.DB,
	cusId int,
	card models.CreditCard,
) (map[string]interface{}, *models.RequestError) {

	c := cusdb.CustomerDB{DB: sqlDB}

	// 1. Get customer stripe id
	cusStripeId, err := c.GetCustomerStripeId(cusId) 
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}
	
	// 2. Add credit card to stripe
	pm, err := my_stripe.AddNewCustomerCard(*cusStripeId, &card)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 3. Update customer default payment id
	cardId, err := c.AddNewCustomerCard(cusId, models.CardInfo{
		Last4: pm.Card.Last4,
		StripeID: pm.ID,
		CusID: cusId,
		Brand: string(pm.Card.Brand),
	})

	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	err = c.UpdateCusDefaultCard(cusId, *cardId)
	if err != nil {
		return nil, &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	// 4. Set default
	return map[string]interface{} {
		"card_id": cardId,
		"brand": pm.Card.Brand,
		"last4": pm.Card.Last4,
	}, nil
}



func CreateCusPass(sqlDB *sql.DB, s3sess *session.Session, cusId int) (*models.RequestError) {
	// 6. Add pkpass and image to cloud

	c := cusdb.CustomerDB{DB: sqlDB}
	cus, err := c.GetCustomerByID(cusId)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	fullName := cus.FullName()

	err = passes.GenerateCustomerPass(s3sess, fullName, cus.Uuid, cus.ID)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	err = media.GenerateCusQR(s3sess, cus.Uuid, cus.ID)
	if err != nil {
		return &models.RequestError{
			Err: err,
			StatusCode: http.StatusBadGateway,
		}
	}

	return nil
}