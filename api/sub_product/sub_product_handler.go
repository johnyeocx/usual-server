package sub_product

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/utils/middleware"
)


func Routes(subProductRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	subProductRouter.GET("/product_stats/:productId", getSubProductStats(sqlDB))
	subProductRouter.POST("/products_stats", getSubProductsStats(sqlDB))
}

func getSubProductStats(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		productId, ok := c.Params.Get("productId")
		if !ok {
			c.JSON(http.StatusBadRequest, errors.New("missing product id param"))
		}

		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, errors.New("failed to get business id"))
			return
		}


		// 1. get list of subscriptions for product


		fmt.Println(productId, businessId)
		c.JSON(200, nil);
	}
}


// get stats for multiple sub products
func getSubProductsStats(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
	}
}