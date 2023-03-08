package auth

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/constants"
	"github.com/johnyeocx/usual/server/db"
	busdb "github.com/johnyeocx/usual/server/db/bus_db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/johnyeocx/usual/server/external/media"
	"github.com/johnyeocx/usual/server/external/my_stripe"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/johnyeocx/usual/server/utils/secure"
)

// AUTH ROUTES
func Routes(authRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	
	authRouter.POST("/validate", validateTokenHandler(sqlDB))
	authRouter.POST("/refresh_token", refreshTokenHandler(sqlDB))
	authRouter.POST("/resend_email_otp", resendEmailOTPHandler(sqlDB))

	authRouter.POST("/create_business", createBusinessHandler(sqlDB, s3Sess))
	authRouter.POST("/verify_email", verifyEmailHandler(sqlDB))
	authRouter.POST("/login", loginHandler(sqlDB))
	// authRouter.POST("/verify_msg_otp", verifyRegisterOTPHandler(conn))
	// authRouter.POST("/register_user_details", registerUserDetailsHandler(conn))
}

func createBusinessHandler(sqlDB *sql.DB, s3Sess *session.Session) gin.HandlerFunc {
	return func(c *gin.Context) {

		reqBody := models.BusinessDetails{}
		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for register business details: %v\n", err)
			c.JSON(400, err)
			return
		}

		// 1. register business to sql
		id, reqErr := createBusiness(sqlDB, &reqBody)
		if reqErr != nil {
			log.Println(reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		// 2. generate presigned url
		key := "./business/profile_image/" + strconv.Itoa(int(*id))
		url, err := cloud.GetImageUploadUrl(s3Sess, key)
		if err != nil {
			log.Printf("Failed to decode req body for register business details: %v\n", err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		// 3. email verification code
		emailOtp, reqErr := GenerateEmailOTP(sqlDB, reqBody.Email, "register")
		if reqErr != nil {
			log.Printf("Failed to generate email verification otp: %v\n", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		// 4. send verification via email
		err = media.SendEmailVerification(reqBody.Email, reqBody.Name, *emailOtp)
		if err != nil {
			log.Println("Failed to email verification code: ", err)
			c.JSON(http.StatusBadGateway, err)
		}
		
		c.JSON(200, url)
	}
}

func validateTokenHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		_, err := middleware.AuthenticateBId(c, sqlDB)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		c.JSON(200, nil)
	}
}

func refreshTokenHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {

		reqBody := struct {
			RefreshToken	string `json:"refresh_token"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for refresh token: %v\n", err)
			c.JSON(400, err)
			return
		}

		businessId, err := refreshToken(sqlDB, reqBody.RefreshToken)
		if err != nil {
			log.Printf("Failed to authenticate refresh token: %v\n", err)
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		accessToken, refreshToken, err := generateTokens(*businessId)
		if err != nil {
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, map[string]string{
			"access_token": *accessToken,
			"refresh_token": *refreshToken,
		})
	}
}

func verifyEmailHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. Get user email and search if exists in db
		reqBody := struct {
			Email  		 string `json:"email"`
			OTP          string `json:"otp"`
			IPAddress	string `json:"ip_address"`
		}{}
		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		// 2. Verify email
		_, reqErr := VerifyEmailOTP(sqlDB, reqBody.Email, reqBody.OTP, "register")
		if reqErr != nil {
			log.Println(reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		// 4. Get business by email
		businessDB := busdb.BusinessDB{DB: sqlDB}
		business, err := businessDB.GetBusinessByEmail(reqBody.Email); 
		if err != nil {
			c.JSON(http.StatusBadGateway, err)
			return
		}

		// 5. Create stripe account
		bStripeId, err := my_stripe.CreateBasicConnectedAccount(business.Country, business.Email, reqBody.IPAddress)
		if err != nil {
			c.JSON(http.StatusBadGateway, errors.New("stripe_creation_failed"))
			return
		}
		
		err = businessDB.SetBusinessStripeID(business.ID, *bStripeId)
		if err != nil {
			c.JSON(http.StatusBadGateway, errors.New("stripe_creation_failed"))
			return
		}

		// 3. Success, set user email verified
		authDB := db.AuthDB{DB: sqlDB}
		if err := authDB.SetBusinessVerified(reqBody.Email, true); err != nil {
			c.JSON(http.StatusBadGateway, err)
			return
		}

		// 3. Return jwt token
		accessToken, err := secure.GenerateAccessToken(strconv.Itoa(business.ID), "business")
		if err != nil {
			c.JSON(500, err)
			return
		}

		refreshToken, err := secure.GenerateRefreshToken(strconv.Itoa(business.ID), "business")
		if err != nil {
			c.JSON(500, err)
			return
		}
		
		resBody := map[string]interface{} {
			"access_token": accessToken,
			"refresh_token":refreshToken,
			"business": business,
		}

		c.JSON(200, resBody)
	}
}

func loginHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		// 1. Get user email and search if exists in db
		reqBody := struct {
			Email  		 string `json:"email"`
			Password     string `json:"password"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for login: %v\n", err)
			c.JSON(400, err)
			return
		}

		businessObj, reqErr := login(sqlDB, reqBody.Email, reqBody.Password)
		if reqErr != nil {
			log.Println(reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return 
		}
		
		accessToken, refreshToken, err := generateTokens(businessObj.ID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusBadGateway, err)
			return
		}

		resBody := map[string]interface{} {
			"access_token": *accessToken,
			"refresh_token": *refreshToken,
		}

		c.JSON(200, resBody)
	}
}

func generateTokens(businessId int) (*string, *string, error) {
	accessToken, err := secure.GenerateAccessToken(strconv.Itoa(businessId), "business")
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := secure.GenerateRefreshToken(strconv.Itoa(businessId), "business")
	if err != nil {
		return nil, nil, err
	}

	return &accessToken, &refreshToken, nil
}

func resendEmailOTPHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c* gin.Context) {
		reqBody := struct {
			Email 	string	`json:"email"`
			OtpType  string  `json:"otp_type"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		var reqErr *models.RequestError
		if reqBody.OtpType == constants.OtpTypes.RegisterBusEmail  {
			reqErr = sendBusRegEmailOTP(sqlDB, reqBody.Email)
		} else if reqBody.OtpType == constants.OtpTypes.UpdateCusEmail {
			// bId, err := middleware.AuthenticateBId(c, sqlDB)
			// if err != nil {
			// 	c.JSON(http.StatusUnauthorized, err)
			// 	return
			// }
			// reqErr = sendUpdateEmailOTP(sqlDB, *cusId, reqBody.Email)
		} else {
			c.JSON(http.StatusBadRequest, errors.New("invalid otp type"))
		}

		if reqErr != nil {
			log.Printf("Failed to resend bus email verification: %v\n", reqErr.Err)
			c.JSON(http.StatusBadGateway, reqErr.Err)
			return
		}

		c.JSON(200, nil)
	}
}