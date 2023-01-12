package customer

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/secure"
)

func Routes(customerRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	customerRouter.POST("create_from_subscribe", createCFromSubscribeHandler(sqlDB))
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

		accessToken, err := secure.GenerateAccessToken(strconv.Itoa(*cId))
		if err != nil {
			c.JSON(http.StatusBadGateway, err)
		}

		c.JSON(200, map[string]string{
			"access_token": accessToken,
		})
	}
}