package usage

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/utils/middleware"
)


func Routes(usageRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	usageRouter.POST("/scan", scanCusQRHandler(sqlDB))
	usageRouter.POST("/insert_usage", insertCusUsageHandler(sqlDB))
}

func scanCusQRHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}
		reqBody := struct {
			CusUUID		string `json:"customer_uuid"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		res, reqErr := ScanCusQR(sqlDB, reqBody.CusUUID, *businessId)
		if reqErr != nil {
			log.Println("Failed to scan cus QR: ", reqErr)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}


		c.JSON(200, res)
	}
}

func insertCusUsageHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}
		reqBody := struct {
			CusUUID			string `json:"customer_uuid"`
			SubUsageID 		int `json:"sub_usage_id"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		res, reqErr := InsertCusUsage(sqlDB,  reqBody.CusUUID, *businessId, reqBody.SubUsageID)
		if reqErr != nil {
			log.Println("Failed to insert cus usage: ", reqErr)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, res)
	}
}