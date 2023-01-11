package c_business

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
)


func Routes(businessRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	businessRouter.GET("", getBusinessHandler(sqlDB))
}

func getBusinessHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		// 1. get business id from params
		businessId, ok := c.GetQuery("business_id")
		if !ok {
			log.Printf("No business id param\n")
			c.JSON(http.StatusBadRequest, nil)
			return
		}

		businessIdInt, err := strconv.Atoi(businessId)
		if err != nil {
			log.Printf("invalid business id: %v", err)
			c.JSON(http.StatusBadRequest, err)
			return
		}

		res, err := GetBusinessByID(sqlDB, businessIdInt)
		if err != nil {
			log.Printf("failed to get business details: %v", err)
			c.JSON(http.StatusBadRequest, err) 
			return
		}

		c.JSON(http.StatusOK, res)
	}
}