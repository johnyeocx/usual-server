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
	businessRouter.GET("/data/:id", getBusinessDataHandler(sqlDB))
	
	businessRouter.GET("/accounts", accountSearch(sqlDB))
	businessRouter.GET("/sub_products", searchSubProductsHandler(sqlDB))

	businessRouter.GET("/sub_product/:id", getSubProductHandler(sqlDB))
}

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

func searchSubProductsHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		query := c.Query("q")
		if query == "" {
			c.JSON(http.StatusBadRequest, errors.New("query field empty"))
			return
		}
		
		res, err := SearchSubProducts(sqlDB, strings.ToLower(query))
		if err != nil {
			log.Println("Failed to search for sub products: ", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}
		

		c.JSON(http.StatusOK, res)
		
	}
}

func getBusinessDataHandler(sqlDB *sql.DB) gin.HandlerFunc {
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

func getBusinessHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId := c.Param("id")
		businessIdInt, err := strconv.Atoi(businessId)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		business, reqErr := GetBusiness(sqlDB, businessIdInt)
		if reqErr != nil {
			log.Println("Failed to get business for customer: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, err)
			return
		}

		c.JSON(http.StatusOK, business)
	}
}

func getSubProductHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		subProductId := c.Param("id")
		subProductIdInt, err := strconv.Atoi(subProductId)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		res, err := GetCSubProduct(sqlDB, subProductIdInt)
		if err != nil {
			log.Println("Failed to get sub product for customer: ", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}


		c.JSON(http.StatusOK, res)
	}
}