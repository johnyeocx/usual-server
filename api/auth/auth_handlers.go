package auth

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/johnyeocx/usual/server/utils/secure"
)

// AUTH ROUTES
func Routes(authRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	
	authRouter.POST("/validate", validateTokenHandler(sqlDB))
	authRouter.POST("/refresh_token", refreshTokenHandler(sqlDB))

	authRouter.POST("/create_business", createBusinessHandler(sqlDB, s3Sess))
	authRouter.POST("/verify_email", verifyEmailHandler(sqlDB))
	authRouter.POST("/login", loginHandler(sqlDB))
	// authRouter.POST("/verify_msg_otp", verifyRegisterOTPHandler(conn))
	// authRouter.POST("/register_user_details", registerUserDetailsHandler(conn))
}

func validateTokenHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		_, err := middleware.AuthenticateId(c, sqlDB)
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
		reqErr = generateRegisterOTP(sqlDB, reqBody.Email, reqBody.Name)
		if reqErr != nil {
			log.Printf("Failed to generate email verification otp: %v\n", reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}
		
		c.JSON(200, url)
	}
}

func verifyEmailHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

		// 1. Get user email and search if exists in db
		reqBody := struct {
			Email  		 string `json:"email"`
			OTP          string `json:"otp"`
		}{}
		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body for verify otp: %v\n", err)
			c.JSON(400, err)
			return
		}
		
		// 2. Verify email
		business, reqErr := verifyEmailOTP(sqlDB, reqBody.Email, reqBody.OTP)
		if reqErr != nil {
			log.Println(reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return
		}

		// 3. Return jwt token
		accessToken, err := secure.GenerateAccessToken(strconv.Itoa(business.BusinessID))
		if err != nil {
			c.JSON(500, err)
			return
		}

		refreshToken, err := secure.GenerateRefreshToken(strconv.Itoa(business.BusinessID))
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
		
		accessToken, refreshToken, err := generateTokens(businessObj.BusinessID)
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
	accessToken, err := secure.GenerateAccessToken(strconv.Itoa(businessId))
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := secure.GenerateRefreshToken(strconv.Itoa(businessId))
	if err != nil {
		return nil, nil, err
	}

	return &accessToken, &refreshToken, nil
}