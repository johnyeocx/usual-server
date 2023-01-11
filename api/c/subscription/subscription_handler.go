package subscription

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/utils/middleware"
)


func Routes(subRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	subRouter.POST("create", CreateSubscriptionHandler(sqlDB))
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
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		err = CreateSubscription(sqlDB, *customerId, reqBody.ProductIDs)
		if err != nil {
			c.JSON(http.StatusBadGateway, err)
			return
		}
		c.JSON(200, nil)
	}
}