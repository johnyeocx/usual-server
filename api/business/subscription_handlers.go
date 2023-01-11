package business

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/johnyeocx/usual/server/utils/middleware"
)

func createSubProductHandler(sqlDB *sql.DB, s3Sess *session.Session)  gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			ProductCategory	models.ProductCategory `json:"category"`
			Product			models.Product	`json:"product"`
			SubPlan			models.SubscriptionPlan `json:"subscription_plan"`
		}{}
			
		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}


		// 1. get business by id
		reqBody.SubPlan.Currency = "GBP" // default for now
		newCatId, subProduct, err := createSubProduct(
			sqlDB, 
			*businessId, 
			&reqBody.ProductCategory, 
			&reqBody.Product, 
			&reqBody.SubPlan,
		)


		if err != nil {
			log.Printf("Failed to create sub product: %v", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		key := "./business/product_image/" + strconv.Itoa(subProduct.Product.ProductID)
		url, err := cloud.GetImageUploadUrl(s3Sess, key)
		if err != nil {
			log.Printf("Failed to decode req body for register business details: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}
		

		resBody := map[string]interface{} {
			"subscription_product": subProduct,
			"upload_url": url,
		}

		if (newCatId != nil) {
			resBody["new_category"] = models.ProductCategory {
				CategoryID: newCatId,
				Title: reqBody.ProductCategory.Title,
			}
		}

		fmt.Println(resBody)

		c.JSON(200, resBody)
	}
}

func setProductNameHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Name	string `json:"name"`
			ProductID int `json:"product_id"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}

		b := db.BusinessDB{DB: sqlDB}
		err = b.SetProductName(*businessId, reqBody.ProductID, reqBody.Name)
		if err != nil {
			log.Printf("Failed to update product name: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func setProductDescriptionHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Description	string `json:"description"`
			ProductID int `json:"product_id"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		b := db.BusinessDB{DB: sqlDB}
		err = b.SetProductDescription(*businessId, reqBody.ProductID, reqBody.Description)
		if err != nil {
			log.Printf("Failed to update product description: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func setProductCategoryHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			ProductID int `json:"product_id"`
			CategoryID *int `json:"category_id"`
			Title string `json:"title"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}


		b := db.BusinessDB{DB: sqlDB}
		catId, err := b.SetProductCategory(*businessId, reqBody.ProductID, reqBody.CategoryID, reqBody.Title)
		if err != nil {
			log.Printf("Failed to update product category: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		if catId != nil {
			c.JSON(200, map[string]int {
				"new_category_id": *catId,
			})
		} else {
			c.JSON(200, nil)
		}
	}
}

func setSubProductPricingHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			ProductID	int `json:"product_id"`
			PlanID int `json:"plan_id"`
			RecurringDuration models.TimeFrame `json:"recurring_duration"`
			UnitAmount int `json:"unit_amount"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		err = UpdateSubProductPricing(
			sqlDB,
			*businessId, reqBody.ProductID, reqBody.PlanID, reqBody.RecurringDuration, reqBody.UnitAmount,
		)
		if err != nil {
			log.Printf("Failed to update product pricing: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func setSubProductUsageHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			ProductID	int `json:"product_id"`
			PlanID int `json:"plan_id"`
			UsageUnlimited bool `json:"usage_unlimited"`
			UsageDuration *models.TimeFrame `json:"usage_duration"`
			UsageAmount *int `json:"usage_amount"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}


		b := db.BusinessDB{DB: sqlDB}
		err = b.SetSubProductUsage(
			*businessId, 
			reqBody.ProductID, 
			reqBody.PlanID, 
			reqBody.UsageUnlimited, 
			reqBody.UsageDuration, 
			reqBody.UsageAmount,
		)
		
		if err != nil {
			log.Printf("Failed to update product description: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}