package c_auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/johnyeocx/usual/server/utils/secure"
)

// AUTH ROUTES
func Routes(authRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session) {
	

	authRouter.GET("/pkpass", getPkPassPresignedUrlHandler(sqlDB, s3Sess))


	authRouter.POST("/validate", validateTokenHandler(sqlDB))
	authRouter.POST("/refresh_token", refreshTokenHandler(sqlDB))
	authRouter.POST("/login", loginHandler(sqlDB))
	// authRouter.POST("/verify_msg_otp", verifyRegisterOTPHandler(conn))
	// authRouter.POST("/register_user_details", registerUserDetailsHandler(conn))
}

func getPkPassPresignedUrlHandler(sqlDB *sql.DB, s3sess *session.Session) gin.HandlerFunc {
	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		key := fmt.Sprintf("customer/pkpass/%d.pkpass", *cusId)
		fmt.Println("KEY:", key)

		url, err := cloud.GetObjectPresignedURL(s3sess, key)
		if err != nil {
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, url)
	}
}

func validateTokenHandler(sqlDB *sql.DB) gin.HandlerFunc {
	return func (c *gin.Context) {
		_, err := middleware.AuthenticateCId(c, sqlDB)
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

		cId, err := refreshToken(sqlDB, reqBody.RefreshToken)
		if err != nil {
			log.Printf("Failed to authenticate refresh token: %v\n", err)
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		accessToken, refreshToken, err := secure.GenerateTokensFromId(*cId, "customer")
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

		res, reqErr := login(sqlDB, reqBody.Email, reqBody.Password)
		if reqErr != nil {
			log.Println(reqErr.Err)
			c.JSON(reqErr.StatusCode, reqErr.Err)
			return 
		}
		
		c.JSON(http.StatusOK, res)
	}
}

