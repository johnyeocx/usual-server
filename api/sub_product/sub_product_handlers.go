package sub_product

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/johnyeocx/usual/server/utils/middleware"
)


func Routes(subProductRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	subProductRouter.POST("/create", createSubProductHandler(sqlDB, s3Sess))
	subProductRouter.POST("/product_stats", getSubProductStatsHandler(sqlDB))
	subProductRouter.POST("/usage", addProductUsageHandler(sqlDB))

	subProductRouter.PATCH("/name", updateProductNameHandler(sqlDB))
	subProductRouter.PATCH("/category", updateProductCategoryHandler(sqlDB))
	subProductRouter.PATCH("/usage", updateProductUsageHandler(sqlDB))


	subProductRouter.DELETE("/:productId", deleteSubProductHandler(sqlDB, s3Sess))
}


func createSubProductHandler(sqlDB *sql.DB, s3Sess *session.Session)  gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			ProductCategory	models.ProductCategory `json:"category"`
			Product			models.Product	`json:"product"`
			SubPlan			models.SubscriptionPlan `json:"subscription_plan"`
			Usages			[]models.SubUsage `json:"usages"`
		}{}
			
		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
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
			reqBody.Usages,
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

		c.JSON(200, resBody)
	}
}

func getSubProductStatsHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {

		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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

func updateProductNameHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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
		
		reqErr := UpdateProductName(sqlDB, *businessId, reqBody.ProductID, reqBody.Name)
		if reqErr != nil {
			log.Println("Failed to update product name:", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateProductCategoryHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func  (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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

		catId, reqErr := UpdateProductCategory(
			sqlDB, *businessId, reqBody.ProductID, reqBody.Title, reqBody.CategoryID)
		
		if reqErr != nil {
			c.JSON(reqErr.StatusCode, reqErr.Err)
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

func updateProductUsageHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func  (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			SubUsageID		int 	`json:"sub_usage_id"`
			SubUsage 		models.SubUsage 	`json:"sub_usage"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := UpdateProductUsage(sqlDB, *businessId, reqBody.SubUsageID, reqBody.SubUsage)
		
		if reqErr != nil {
			log.Printf("Failed to update product usage: %v\n", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func addProductUsageHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func  (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			PlanID			int 				`json:"plan_id"`
			SubUsage 		models.SubUsage 	`json:"sub_usage"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		newId, reqErr := AddProductUsage(sqlDB, *businessId, reqBody.PlanID, reqBody.SubUsage)
		
		if reqErr != nil {
			log.Printf("Failed to update product usage: %v\n", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, map[string]int {
			"sub_usage_id": *newId,
		})
	}
}

func deleteSubProductHandler(sqlDB *sql.DB, s3Sess *session.Session) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		productId := c.Param("productId")
		productIdInt, err := strconv.Atoi(productId)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}

		reqErr := DeleteSubProduct(sqlDB, s3Sess, *businessId, productIdInt)
		if reqErr != nil {
			log.Println(reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}
	}
}