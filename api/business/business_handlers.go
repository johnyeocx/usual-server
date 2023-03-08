package business

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	busdb "github.com/johnyeocx/usual/server/db/bus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/media"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/stripe/stripe-go/v74"
)


func Routes(businessRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	businessRouter.GET("", getBusinessHandler(sqlDB))
	businessRouter.GET("/total_and_payouts", getTotalAndPayoutsHandler(sqlDB))
	businessRouter.GET("/transactions", getBusinessTransactionsHandler(sqlDB))
	businessRouter.GET("/email_taken/:email", checkBusinessEmailTaken(sqlDB))

	businessRouter.POST("set_profile", setBusinessProfileHandler(sqlDB, s3Sess))
	businessRouter.POST("set_description", updateBusinessDescriptionHandler(sqlDB))

	businessRouter.POST("identity_document", uploadIdentityDocumentHandler(sqlDB))
	
	businessRouter.PATCH("account/category", updateBusinessCategoryHandler(sqlDB))
	businessRouter.PATCH("account/name", updateBusinessNameHandler(sqlDB))
	// businessRouter.PATCH("account/email", updateBusinessEmailHandler(sqlDB))
	businessRouter.PATCH("account/url", updateBusinessUrlHandler(sqlDB))
	businessRouter.PATCH("account/password", updateBusinessPasswordHandler(sqlDB))
	businessRouter.PATCH("account/personal_info", setPersonalInfoHandler(sqlDB))
	businessRouter.PATCH("account/bank_account", updateBusinessBankAccountHandler(sqlDB))
	

	businessRouter.PATCH("account/description", updateBusinessDescriptionHandler(sqlDB))
	
	businessRouter.PATCH("individual/name", updateIndividualNameHandler(sqlDB))
	businessRouter.PATCH("individual/dob", updateIndividualDOBHandler(sqlDB))
	businessRouter.PATCH("individual/address", updateIndividualAddressHandler(sqlDB))
	businessRouter.PATCH("individual/mobile", updateIndividualMobileHandler(sqlDB))

	businessRouter.PATCH("subscription_product/description", setProductDescriptionHandler(sqlDB))
	businessRouter.PATCH("subscription_product/subscription_pricing", setSubProductPricingHandler(sqlDB))
}

func checkBusinessEmailTaken(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		email, ok := c.Params.Get("email")
		if !ok {
			c.JSON(http.StatusBadRequest, errors.New("invalid email"))
			return
		}

		businessDB := busdb.BusinessDB{DB: sqlDB}
		business, err := businessDB.GetBusinessByEmail(email)

		if err == sql.ErrNoRows {
			c.JSON(200, nil)
			return
		} else if err != nil {
			log.Println("Failed to check if email taken:", err)
			c.JSON(http.StatusBadGateway, nil)
			return
		} else if *business.EmailVerified {
			c.JSON(http.StatusConflict, nil)
			return
		} 

		c.JSON(200, nil)
	}
}

func getBusinessHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		res, reqErr := getBusinessData(sqlDB, *businessId)
		if reqErr != nil {
			log.Println("Failed to get business data:", reqErr)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, res)
	}
}

func getBusinessTransactionsHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		res, reqErr := getBusinessTransactions(sqlDB, *businessId)
		if reqErr != nil {
			reqErr.Log()
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, res)
	}
}

func setBusinessProfileHandler(sqlDB *sql.DB, s3Sess *session.Session)  gin.HandlerFunc {
	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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

		// 3. set up business profile
		reqErr := setBusinessProfile(
			sqlDB,
			*businessId,
			reqBody.BusinessCategory, 
			reqBody.BusinessUrl, 
			&reqBody.Individual,
			reqBody.IPAddress,
		) 

		if reqErr != nil {

			var stripeErr stripe.Error
			err := json.Unmarshal([]byte(reqErr.Err.Error()), &stripeErr)
			if err == nil {
				c.JSON(http.StatusNotAcceptable, map[string]string {
					"param": stripeErr.Param,
					"message": stripeErr.Msg,
					"code": string(stripeErr.Code),
				})
				return;
			}

			log.Printf("Failed to set business profile: %v", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}
		
		// 4. Create Business QR
		media.GenerateSubscribeQRCode(s3Sess, *businessId)
		
		c.JSON(200, nil)
	}
}

func updateBusinessNameHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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

		reqErr := updateBusinessName(sqlDB, *businessId, reqBody.Name)
		if reqErr != nil {
			log.Printf("Failed to update business name: %v\n", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateBusinessPasswordHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			OldPassword	string `json:"old_password"`
			NewPassword	string `json:"new_password"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		reqErr := updateBusinessPassword(sqlDB, *businessId, reqBody.OldPassword, reqBody.NewPassword)
		if reqErr != nil {
			log.Printf("Failed to update business password: %v\n", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateBusinessEmailHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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

		businessDB := busdb.BusinessDB{DB: sqlDB}
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
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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
		
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
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
			var stripeErr  stripe.Error
			err := json.Unmarshal([]byte(err.Error()), &stripeErr)
			if err == nil {
				c.JSON(http.StatusNotAcceptable, stripeErr.Code)
				return;
			}

			log.Printf("Failed to update business url: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func setPersonalInfoHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		reqBody := struct {
			FirstName string `json:"first_name"`
			LastName string `json:"last_name"`
			DialingCode string `json:"dialing_code"`
			MobileNumber string `json:"mobile_number"`
			DOB models.Date 	`json:"dob"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		err = updateIndividualDetailsStripe(sqlDB, *businessId, reqBody.FirstName, reqBody.LastName, reqBody.DialingCode, reqBody.MobileNumber, reqBody.DOB)
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

			log.Printf("Failed to update business mobile / dob: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, nil)
	}
}

func updateBusinessBankAccountHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {
		
		businessId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
		}

		var reqBody models.BankInfo

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}
		bankAccount, reqErr := updateBusinessBankInfo(sqlDB, *businessId, reqBody)

		if reqErr != nil {
			var stripeErr  stripe.Error
			err := json.Unmarshal([]byte(reqErr.Err.Error()), &stripeErr)
			if err == nil {
				c.JSON(http.StatusNotAcceptable, stripeErr.Param)
				return;
			}

			log.Printf("Failed to update business bank account: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, bankAccount)
	}
}

func uploadIdentityDocumentHandler(sqlDB *sql.DB) gin.HandlerFunc {

	return func (c *gin.Context) {

		bId, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		frontFile, err := c.FormFile("front")
		if err != nil {
			fmt.Println("Failed to get form file front: ", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		openedFront, err := frontFile.Open()
		if err != nil {
			log.Println("Failed to open front image: ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		defer openedFront.Close()

		backFile, err := c.FormFile("back")
		var openedBack *multipart.File
	
		if err == nil {
			openedBack, err := backFile.Open()
			if err != nil {
				log.Println("Failed to open back image: ", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			} 
			defer openedBack.Close()
		} else {
			openedBack = nil
		}
		
		reqErr := uploadIndVerificationDoc(sqlDB, *bId, openedFront, openedBack)
		if reqErr != nil {
			log.Println("Failed to upload individual verification doc:", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}
		
		c.JSON(http.StatusOK, gin.H{
			"message": "File uploaded successfully",
		})

	}
}


