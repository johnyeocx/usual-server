package business

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/stripe/stripe-go/v74"
)

// UPDATE
func updateIndividualNameHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			FirstName	string `json:"first_name"`
			LastName	string `json:"last_name"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := updateIndividualName(sqlDB, *businessId, reqBody.FirstName, reqBody.LastName)
		if reqErr != nil {
			log.Printf("Failed to update individual name: %v\n", reqErr.Err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateIndividualDOBHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Day		int `json:"day"`
			Month	int `json:"month"`
			Year	int `json:"year"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		err = updateIndividualDOB(sqlDB, *businessId, reqBody.Day, reqBody.Month, reqBody.Year)
		if err != nil {
			var stripeErr  stripe.Error
			err := json.Unmarshal([]byte(err.Error()), &stripeErr)
			if err == nil {
				c.JSON(http.StatusNotAcceptable, map[string]string {
					"param": stripeErr.Param,
					"message": stripeErr.Msg,
				})
				return;
			}
			
			log.Printf("Failed to update individual name: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateIndividualAddressHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Line1		string `json:"line1"`
			Line2		string `json:"line2"`
			City		string `json:"city"`
			PostalCode	string `json:"postal_code"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		err = updateIndividualAddress(
			sqlDB, 
			*businessId, 
			reqBody.Line1, 
			reqBody.Line2, 
			reqBody.PostalCode, 
			reqBody.City,
		)

		if err != nil {
			log.Printf("Failed to update individual name: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateIndividualMobileHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			DialingCode		string `json:"dialing_code"`
			Number		string `json:"number"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		err = updateIndividualMobile(
			sqlDB, 
			*businessId, 
			reqBody.DialingCode, 
			reqBody.Number, 
		)

		if err != nil {
			var stripeErr  stripe.Error
			err := json.Unmarshal([]byte(err.Error()), &stripeErr)
			if err == nil {
				c.JSON(http.StatusNotAcceptable, map[string]string {
					"param": stripeErr.Param,
					"message": stripeErr.Msg,
				})
				return;
			}
			
			log.Printf("Failed to update individual mobile: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}