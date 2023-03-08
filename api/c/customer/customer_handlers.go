package customer

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/constants"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/middleware"
)

func Routes(customerRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	customerRouter.GET("data", getCustomerDataHandler(sqlDB))
	customerRouter.GET("subs", getCusSubsAndInvoicesHandler(sqlDB))

	customerRouter.POST("fcm_token", saveCusFCMTokenHandler(sqlDB))


	customerRouter.POST("create", createCustomerHandler(sqlDB))
	customerRouter.POST("create_pass", createCusPassHandler(sqlDB, s3Sess))
	customerRouter.POST("verify_email", verifyCustomerEmailHandler(sqlDB, s3Sess))
	customerRouter.POST("add_card", addCustomerCardHandler(sqlDB))
	customerRouter.POST("resend_email_otp", resendEmailOTPHandler(sqlDB))

	customerRouter.PATCH("name", updateCusNameHandler(sqlDB))
	customerRouter.PATCH("email", sendCusUpdateEmailVerificationHandler(sqlDB))
	customerRouter.PATCH("verify_email", verifyCusUpdateEmailHandler(sqlDB))
	customerRouter.PATCH("address", updateCusAddressHandler(sqlDB))
	customerRouter.PATCH("password", updateCusPasswordHandler(sqlDB))
	customerRouter.PATCH("default_payment", updateCusDefaultPaymentHandler(sqlDB))

	customerRouter.DELETE("card/:cardId", deleteCusCardHandler(sqlDB))
}

func saveCusFCMTokenHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {

		cusId , err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, errors.New("unauthenticated user"))
			return
		}
		
		reqBody := struct {FCMToken string `json:"fcm_token"`}{}

		err = c.BindJSON(&reqBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		reqErr  := updateCusFCMToken(sqlDB, *cusId, reqBody.FCMToken)
		if reqErr != nil {
			log.Println("Failed to update cus fcm token: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}
		
		c.JSON(http.StatusOK, nil)
	}
}

func getCustomerDataHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		
		res, reqErr := GetCustomerData(sqlDB, *cusId)
		if reqErr != nil {
			log.Println("Failed to get customer: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

func getCusSubsAndInvoicesHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}


		res, reqErr := GetCusSubsAndInvoices(sqlDB, *cusId)
		if reqErr != nil {
			log.Println("Failed to get customer: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}


func createCustomerHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		reqBody := struct {
			FirstName			string 	`json:"first_name"`
			LastName			string 	`json:"last_name"`
			Email 			string 	`json:"email"`
			Password	 	string 	`json:"password"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := CreateCustomer(sqlDB, reqBody.FirstName, reqBody.LastName, reqBody.Email, reqBody.Password)
		if reqErr != nil {
			log.Println("Failed to create customer:", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func createCusPassHandler(sqlDB *sql.DB, s3sess *session.Session) gin.HandlerFunc {
	return func (c *gin.Context) {
		cId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		reqErr := CreateCusPass(sqlDB, s3sess, *cId)
		if reqErr != nil {
			c.JSON(200, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func verifyCustomerEmailHandler(sqlDB *sql.DB, s3Sess *session.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. Get user email and search if exists in db
		reqBody := struct {
			Email  		 string `json:"email"`
			OTP          string `json:"otp"`
		}{}

		
		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}
		fmt.Println(reqBody)
		
		// 2. Verify email
		res, reqErr := VerifyCustomerRegEmail(s3Sess, sqlDB, reqBody.Email, reqBody.OTP)
		if reqErr != nil {
			log.Println(reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, res)
	}
}

func resendEmailOTPHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c* gin.Context) {
		reqBody := struct {
			Email 	string	`json:"email"`
			OtpType  string  `json:"otp_type"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		var reqErr *models.RequestError
		if reqBody.OtpType == constants.OtpTypes.RegisterCusEmail  {

			reqErr = sendRegEmailOTP(sqlDB, reqBody.Email)
		} else if reqBody.OtpType == constants.OtpTypes.UpdateCusEmail {
			cusId, err := middleware.AuthenticateCId(c, sqlDB)
			if err != nil {
				c.JSON(http.StatusUnauthorized, err)
				return
			}
			reqErr = sendUpdateEmailOTP(sqlDB, *cusId, reqBody.Email)
		} else {
			c.JSON(http.StatusBadRequest, errors.New("invalid otp type"))
		}

		if reqErr != nil {
			log.Printf("Failed to resend cus update email verification: %v\n", reqErr.Err)
			c.JSON(http.StatusBadGateway, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func addCustomerCardHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Number		string `json:"number"`
			ExpMonth 	int64 `json:"expiry_month"`
			ExpYear 	int64 `json:"expiry_year"`
			CVC 		string `json:"cvc"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		res, reqErr := AddCusCreditCard(sqlDB, *cusId, models.CreditCard{
			Number: reqBody.Number,
			ExpMonth: reqBody.ExpMonth,
			ExpYear: reqBody.ExpYear,
			CVC: reqBody.CVC,
		})

		if reqErr != nil {
			log.Println("Failed to add custoemr credit card:", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, res)
	}
}

func updateCusNameHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			FirstName	string `json:"first_name"`
			LastName	string `json:"last_name"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := updateCusName(sqlDB, *cusId, reqBody.FirstName, reqBody.LastName)
		if reqErr != nil {
			log.Printf("Failed to update cus name: %v\n", reqErr.Err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func sendCusUpdateEmailVerificationHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Email	string `json:"email"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := sendUpdateEmailOTP(sqlDB, *cusId, reqBody.Email)
		if reqErr != nil {
			log.Printf("Failed to send cus update email verification: %v\n", reqErr.Err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func verifyCusUpdateEmailHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Email	string `json:"email"`
			OTP 	string `json:"otp"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := verifyUpdateCusEmail(sqlDB, *cusId, reqBody.OTP, reqBody.Email)
		if reqErr != nil {
			log.Printf("Failed to send cus update email verification: %v\n", reqErr.Err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateCusAddressHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Address models.Address `json:"address"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		fmt.Println(reqBody)

		reqErr := updateCusAddress(sqlDB, *cusId, reqBody.Address)
		if reqErr != nil {
			log.Printf("Failed to update cus name: %v\n", reqErr.Err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateCusPasswordHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := updateCusPassword(sqlDB, *cusId, reqBody.OldPassword, reqBody.NewPassword)
		if reqErr != nil {
			log.Printf("Failed to update cus password: %v\n", reqErr.Err)
			c.JSON(reqErr.StatusCode, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateCusDefaultPaymentHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			CardID	int `json:"card_id"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := updateCusDefaultPayment(sqlDB, *cusId, reqBody.CardID)
		if reqErr != nil {
			log.Printf("Failed to update cus default card: %v\n", reqErr.Err)
			c.JSON(reqErr.StatusCode, err)
			return
		}

		c.JSON(200, nil)
	}
}

func deleteCusCardHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		cardId, _ := c.Params.Get("cardId")
		cardIdInt, err := strconv.Atoi(cardId)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.New("invalid card id"))
			return
		}

		reqErr := deleteCard(sqlDB, *cusId, cardIdInt)
		if reqErr != nil {
			log.Printf("Failed to delete cus card: %v\n", reqErr.Err)
			c.JSON(reqErr.StatusCode, map[string]string{
				"code": reqErr.Err.Error(),
			})
			return
		}

		c.JSON(200, nil)
	}
}
