package sub_product

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/utils/middleware"
)


func Routes(subProductRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	subProductRouter.POST("/product_stats", getSubProductStats(sqlDB))
	subProductRouter.POST("/products_stats", getSubProductsStats(sqlDB))
}

func getSubProductStats(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {

		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, errors.New("failed to get business id"))
			return
		}

		var reqBody struct {
			ProductID 	int	`json:"product_id"`
		}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Println("Failed to provide req body")
			c.JSON(http.StatusBadRequest, err)
			return
		}


		// 1. get list of subscriptions for product
		data, reqErr := GetSubProductStats(sqlDB, *businessId, reqBody.ProductID)
		if err != nil {
			log.Println("Failed to get sub product stats:", err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		
		c.JSON(200, data);
	}
}


// get stats for multiple sub products
func getSubProductsStats(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
	}
}