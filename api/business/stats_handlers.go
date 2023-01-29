package business

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/utils/middleware"
)

func getTotalAndPayoutsHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		res, err := GetTotalAndPayouts(sqlDB, *businessId)
		if err != nil {
			log.Println("Failed to get business total and payout:", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, res)
	}
}