package subscription

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/stripe/stripe-go/v74"
)


func Routes(subRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	subRouter.GET("/:id", getSubscriptionDataHandler(sqlDB))

	subRouter.POST("create", CreateSubscriptionHandler(sqlDB))
	subRouter.POST("resolve_payment_intent", ResolvePaymentIntentHandler(sqlDB))
	subRouter.PATCH("resume", ResumeSubscriptionHandler(sqlDB))
	
	subRouter.PATCH("default_card", ChangeSubDefaultCardHandler(sqlDB))
	
	subRouter.DELETE("cancel/:subId", CancelSubscriptionHandler(sqlDB))
}


func getSubscriptionDataHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		productId := c.Param("id")
		productIdInt, err := strconv.Atoi(productId)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		cusId, err := middleware.AuthenticateCId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		res, reqErr := GetSubscriptionData(sqlDB, *cusId, productIdInt)
		if reqErr != nil {
			log.Println("Failed to get customer: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}
		
		// time.Sleep(time.Second * 2)
		c.JSON(http.StatusOK, res)
	}
}


func CreateSubscriptionHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {

		// GET CUSTOMER ID
		customerId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		reqBody := struct {
			ProductID		int		`json:"product_id"`
			CardID			int		`json:"card_id"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		res, reqErr := CreateSubscription(sqlDB, *customerId, reqBody.CardID, reqBody.ProductID)
		if reqErr != nil {
			log.Println("Failed to create subscription:", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, res)
	}
}

func ResumeSubscriptionHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		
		// GET CUSTOMER ID
		customerId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		reqBody := struct {
			CardID			int		`json:"card_id"`
			SubID 			int 	`json:"sub_id"`
		}{}


		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body : %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := ResumeSubscription(sqlDB, *customerId, reqBody.CardID, reqBody.SubID)
		if reqErr != nil {
			log.Println("Failed to resume subscription: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func CancelSubscriptionHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		
		// GET CUSTOMER ID
		customerId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		productId := c.Param("subId")
		productIdInt, err := strconv.Atoi(productId)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		res, reqErr := CancelSubscription(sqlDB, *customerId, productIdInt)
		if reqErr != nil {
			log.Println("Failed to cancel subscription: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, res)
	}
}

func ChangeSubDefaultCardHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		
		// GET CUSTOMER ID
		customerId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		reqBody := struct {
			SubID	int `json:"sub_id"`
			CardID int `json:"card_id"`
		}{}
		err = c.BindJSON(&reqBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		reqErr := ChangeSubDefaultCard(sqlDB, *customerId, reqBody.SubID, reqBody.CardID)
		if reqErr != nil {
			log.Println("Failed to cancel subscription: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func ResolvePaymentIntentHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c * gin.Context) {
		customerId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		reqBody := struct {
			SubID 	int `json:"sub_id"`
			CardID 	int `json:"card_id"`
		}{}
		err = c.BindJSON(&reqBody)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}


		res, reqErr := ResolvePaymentIntent(sqlDB, *customerId, reqBody.CardID, reqBody.SubID)
		if reqErr != nil {
			stripeErr := stripe.Error{}
			err := json.Unmarshal([]byte(reqErr.Err.Error()), &stripeErr)
			if err == nil {
				log.Println("Stripe error: ", stripeErr.Code)
				c.JSON(reqErr.StatusCode, stripeErr.Code)
				return
			}

			log.Println("Failed to resolve payment intent:", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}