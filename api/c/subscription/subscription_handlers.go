package subscription

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/utils/middleware"
)


func Routes(subRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	subRouter.GET("/:id", getSubscriptionDataHandler(sqlDB))

	subRouter.POST("create", CreateSubscriptionHandler(sqlDB))
	subRouter.PATCH("resume", ResumeSubscriptionHandler(sqlDB))
	
	subRouter.PATCH("default_card", ChangeSubDefaultCardHandler(sqlDB))
	
	subRouter.DELETE("delete/:subId", DeleteSubscriptionHandler(sqlDB))
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

func DeleteSubscriptionHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {

		// GET CUSTOMER ID
		customerId, err := middleware.AuthenticateCId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		subId, _ := c.Params.Get("subId")
		subIdInt, err := strconv.Atoi(subId)
		if err != nil {
			c.JSON(http.StatusBadRequest, errors.New("invalid sub id provided"))
			return
		}
		
		reqErr := DeleteSubscription(sqlDB, *customerId, subIdInt)
		if reqErr != nil {
			log.Println("Failed to delete subscription:", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
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

		expires, reqErr := CancelSubscription(sqlDB, *customerId, productIdInt)
		if reqErr != nil {
			log.Println("Failed to cancel subscription: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, map[string]interface{} {
			"expires": expires,
		})
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
