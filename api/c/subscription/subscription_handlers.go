package subscription

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/utils/middleware"
)


func Routes(subRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	subRouter.POST("create", CreateSubscriptionHandler(sqlDB))
	subRouter.DELETE("cancel/:subId", CancelSubscriptionHandler(sqlDB))
	// businessRouter.POST("", createSubscriptionHandler(sqlDB))
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
			ProductIDs		[]int	`json:"product_ids"`
			CardID			int		`json:"card_id"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		subs, reqErr := CreateSubscription(sqlDB, *customerId, reqBody.CardID, reqBody.ProductIDs)
		if reqErr != nil {
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, subs)
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

		reqErr := CancelSubscription(sqlDB, *customerId, productIdInt)
		if reqErr != nil {
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}
