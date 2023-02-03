package customer

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/johnyeocx/usual/server/utils/secure"
)

func Routes(customerRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	customerRouter.GET("data", getCustomerDataHandler(sqlDB))

	customerRouter.POST("create", createCustomerHandler(sqlDB))
	customerRouter.POST("verify_email", verifyCustomerEmailHandler(sqlDB))
	customerRouter.POST("create_from_subscribe", createCFromSubscribeHandler(sqlDB))
	
	customerRouter.POST("add_card", addCustomerCardHandler(sqlDB))
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

func createCustomerHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		reqBody := struct {
			Name			string 	`json:"name"`
			Email 			string 	`json:"email"`
			Password	 	string 	`json:"password"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := CreateCustomer(sqlDB, reqBody.Name, reqBody.Email, reqBody.Password)
		if reqErr != nil {
			log.Println("Failed to create customer:", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func verifyCustomerEmailHandler(sqlDB *sql.DB) gin.HandlerFunc {
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
		
		// 2. Verify email
		res, reqErr := VerifyCustomerEmail(sqlDB, reqBody.Email, reqBody.OTP)
		if reqErr != nil {
			log.Println(reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, res)
	}
}

func createCFromSubscribeHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		reqBody := struct {
			Name			string 				`json:"name"`
			Email 			string 				`json:"email"`
			Card	 	*models.CreditCard 	`json:"card"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		cId, reqErr := CreateCFromSubscribe(sqlDB, reqBody.Name, reqBody.Email, reqBody.Card)
		if reqErr != nil {
			// log.Println("Failed to create customer from subscribe:", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		accessToken, err := secure.GenerateAccessToken(strconv.Itoa(*cId), "customer")
		if err != nil {
			c.JSON(http.StatusBadGateway, err)
		}

		c.JSON(200, map[string]string{
			"access_token": accessToken,
		})
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