package c_business

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
)


func Routes(businessRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	// businessRouter.GET("", getBusinessHandler(sqlDB))
	businessRouter.GET("/explore", getExploreDataHandler(sqlDB))

	businessRouter.GET("/:id", getBusinessHandler(sqlDB))

	businessRouter.GET("/accounts", accountSearch(sqlDB))
}

// func getBusinessHandler(sqlDB *sql.DB) gin.HandlerFunc {
// 	return func (c *gin.Context) {
// 		// 1. get business id from params
// 		businessId, ok := c.GetQuery("business_id")
// 		if !ok {
// 			log.Printf("No business id param\n")
// 			c.JSON(http.StatusBadRequest, nil)
// 			return
// 		}

// 		businessIdInt, err := strconv.Atoi(businessId)
// 		if err != nil {
// 			log.Printf("invalid business id: %v", err)
// 			c.JSON(http.StatusBadRequest, err)
// 			return
// 		}

// 		res, err := GetBusinessByID(sqlDB, businessIdInt)
// 		if err != nil {
// 			log.Printf("failed to get business details: %v", err)
// 			c.JSON(http.StatusBadRequest, err) 
// 			return
// 		}
		
// 		if len(res["product_categories"].([]models.ProductCategory))== 0 {
// 			res["product_categories"] = make([]models.ProductCategory, 0)
// 		}
		
// 		if len(res["sub_products"].([]models.SubscriptionProduct))== 0 {
// 			res["sub_products"] = make([]models.SubscriptionProduct, 0)
// 		}

// 		c.JSON(http.StatusOK, res)
// 	}
// }

func getExploreDataHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		res, err := GetExploreData(sqlDB)
		if err != nil {
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}

func accountSearch(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, errors.New("query field empty"))
			return
		}
		
		accounts, err := SearchAccounts(sqlDB, strings.ToLower(query))
		if err != nil {
			log.Println("Failed to search for accounts: ", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(http.StatusOK, map[string]interface{}{
			"accounts": accounts,
		})
		
	}
}

func getBusinessHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId := c.Param("id")
		businessIdInt, err := strconv.Atoi(businessId)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		res, err := GetBusinessSubProducts(sqlDB, businessIdInt)
		if err != nil {
			log.Println("Failed to get business for customer: ", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}


		c.JSON(http.StatusOK, res)
	}
}