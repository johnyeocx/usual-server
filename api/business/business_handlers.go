package business

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/my_stripe"
	"github.com/johnyeocx/usual/server/utils/middleware"
)


func Routes(businessRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	businessRouter.GET("", getBusinessHandler(sqlDB))

	businessRouter.POST("set_profile", setBusinessProfileHandler(sqlDB))
	businessRouter.POST("set_description", updateBusinessDescriptionHandler(sqlDB))
	businessRouter.POST("add_subscription_product", createSubProductHandler(sqlDB, s3Sess))
	
	businessRouter.PATCH("account/category", updateBusinessCategoryHandler(sqlDB))
	businessRouter.PATCH("account/name", updateBusinessNameHandler(sqlDB))
	businessRouter.PATCH("account/email", updateBusinessEmailHandler(sqlDB))
	businessRouter.PATCH("account/url", updateBusinessUrlHandler(sqlDB))
	businessRouter.PATCH("account/description", updateBusinessDescriptionHandler(sqlDB))
	
	businessRouter.PATCH("individual/name", updateIndividualNameHandler(sqlDB))
	businessRouter.PATCH("individual/dob", updateIndividualDOBHandler(sqlDB))
	businessRouter.PATCH("individual/address", updateIndividualAddressHandler(sqlDB))
	businessRouter.PATCH("individual/mobile", updateIndividualMobileHandler(sqlDB))

	businessRouter.PATCH("subscription_product/name", setProductNameHandler(sqlDB))
	businessRouter.PATCH("subscription_product/category", setProductCategoryHandler(sqlDB))
	businessRouter.PATCH("subscription_product/description", setProductDescriptionHandler(sqlDB))
	businessRouter.PATCH("subscription_product/subscription_pricing", setSubProductPricingHandler(sqlDB))
	businessRouter.PATCH("subscription_product/subscription_usage", setSubProductUsageHandler(sqlDB))
	
	// businessRouter.PATCH("subscription_product/category", createSubProductHandler(sqlDB, s3Sess))
}

func getBusinessHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		// 3. Get business
		businessDB := db.BusinessDB{DB: sqlDB}
		business, err := businessDB.GetBusinessByID(*businessId)

		if err != nil {
			log.Println(err.Error())
			c.JSON(http.StatusBadGateway, nil)
			return
		}
		var individual models.Person
		if business.IndividualID != nil {
			indiv, err := businessDB.GetIndividualByID(*business.IndividualID)
			if err != nil && err != sql.ErrNoRows {
				log.Println(err)
				c.JSON(http.StatusBadGateway, err)
				return
			}

			individual = *indiv
		}

		categories, subProducts, err := GetBusinessProducts(sqlDB, *businessId)
		if err != nil && err != sql.ErrNoRows{
			log.Println(err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		subscribers, err := getBusinessStats(sqlDB, *businessId)
		if err != nil && err != sql.ErrNoRows {
			log.Println("Failed to get biz subscribers", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		resBody := map[string]interface{} {
			"business": business,
			"individual": individual,
			"product_categories": categories,
			"subscription_products": subProducts,
			"subscribers": subscribers,
		}

		c.JSON(200, resBody)
	}
}

func setBusinessProfileHandler(sqlDB *sql.DB)  gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		reqBody := struct {
			IPAddress			string `json:"ip_address"`
			BusinessCategory string `json:"business_category"`
			BusinessUrl 	string `json:"business_url"`
			Individual		models.Person `json:"individual"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}

		// 1. get business by id
		businessDB := db.BusinessDB{DB: sqlDB}
		business, err := businessDB.GetBusinessByID(*businessId)
		if err != nil {
			log.Printf("Failed to get user by id: %v", err)
			c.JSON(400, err)
			return
		}

		// 2. set up stripe account
		stripeId, err := my_stripe.CreateConnectedAccount(
			business.Country,
			business.Email,
			reqBody.IPAddress,
			reqBody.BusinessCategory,
			reqBody.BusinessUrl,
			&reqBody.Individual,
		)
		if err != nil {
			log.Printf("Failed to create stripe connected account: %v", err)
			c.JSON(400, err)
			return
		}

		// 3. set up business profile
		individualId, reqErr := setBusinessProfile(
			sqlDB,
			*businessId,
			reqBody.BusinessCategory, 
			reqBody.BusinessUrl, 
			&reqBody.Individual,
			*stripeId,
		) 

		if reqErr != nil {
			log.Printf("Failed to set business profile: %v", reqErr.Err)
			c.JSON(400, err)
			return
		}
		
		c.JSON(200, struct {
			IndividualID int `json:"individual_id"`
			StripeID 	string 	`json:"stripe_id"`
		}{
			IndividualID: *individualId,
			StripeID: *stripeId,
		})
	}
}

// UPDATE 

func updateBusinessNameHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Name	string `json:"name"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		err = updateBusinessName(sqlDB, *businessId, reqBody.Name)
		if err != nil {
			log.Printf("Failed to update business name: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateBusinessEmailHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Email	string `json:"email"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		err = updateBusinessEmail(sqlDB, *businessId, reqBody.Email)
		if err != nil {
			log.Printf("Failed to update business email: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateBusinessDescriptionHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Description			string `json:"description"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}

		businessDB := db.BusinessDB{DB: sqlDB}
		err = businessDB.SetBusinessDescription(*businessId, reqBody.Description)
		if err != nil {
			log.Printf("Failed to set business description: %v", err)
			c.JSON(http.StatusBadGateway, err)
			return;
		}

		c.JSON(200, nil)
	}
}

func updateBusinessCategoryHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Category	string `json:"category"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		err = updateBusinessCategory(sqlDB, *businessId, reqBody.Category)
		if err != nil {
			log.Printf("Failed to update product category: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateBusinessUrlHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			Url	string `json:"url"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		err = updateBusinessUrl(sqlDB, *businessId, reqBody.Url)
		if err != nil {
			log.Printf("Failed to update product category: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}


