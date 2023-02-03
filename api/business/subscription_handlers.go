package business

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/middleware"
)


func setProductDescriptionHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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


func setSubProductPricingHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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