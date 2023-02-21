package c_auth

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	firebase "firebase.google.com/go"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gin-gonic/gin"
	my_enums "github.com/johnyeocx/usual/server/constants/enums"
	"github.com/johnyeocx/usual/server/external/cloud"
	"github.com/johnyeocx/usual/server/utils/middleware"
	"github.com/johnyeocx/usual/server/utils/secure"
)

// AUTH ROUTES
func Routes(authRouter *gin.RouterGroup, sqlDB *sql.DB, s3Sess *session.Session, fbApp *firebase.App) {
	

	authRouter.GET("/pkpass", getPkPassPresignedUrlHandler(sqlDB, s3Sess))
	authRouter.GET("/qr", getQRPresignedUrlHandler(sqlDB, s3Sess))

	
	
	authRouter.POST("/google_sign_in", googleSignInHandler(sqlDB, fbApp))
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

		url, err := cloud.GetObjectPresignedURL(s3sess, key, time.Minute)
		if err != nil {
			c.JSON(http.StatusBadGateway, err)
			return
		}

		c.JSON(200, url)
	}
}

func getQRPresignedUrlHandler(sqlDB *sql.DB, s3sess *session.Session) gin.HandlerFunc {
	return func (c *gin.Context) {
		cusId, err := middleware.AuthenticateCId(c, sqlDB)

		if err != nil {
			c.JSON(http.StatusUnauthorized, err)
			return
		}

		key := fmt.Sprintf("customer/profile_qr/%d", *cusId)

		url, err := cloud.GetObjectPresignedURL(s3sess, key, time.Hour)
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
			log.Println("Failed to validate customer:", err)
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
			log.Println("Failed to login:", reqErr.Err)
			c.JSON(reqErr.StatusCode, res)
			return 
		}
		
		accessToken := res["access_token"].(string)
		refreshToken := res["refresh_token"].(string)

		c.SetCookie("access_token", accessToken, 60 * 60 * 24, "/", "localhost", false, true);
		c.SetCookie("refresh_token", refreshToken, 60 * 60 * 24, "/", "localhost", false, true);

		c.JSON(http.StatusOK, res)
	}
}

func googleSignInHandler(sqlDB *sql.DB, fbApp *firebase.App) gin.HandlerFunc {
	return func (c *gin.Context) {
		
		// 1. Get user email and search if exists in db
		reqBody := struct {
			Token  		 string `json:"token"`
		}{}

		if err := c.BindJSON(&reqBody); err != nil {
			log.Printf("Failed to decode req body: %v\n", err)
			c.JSON(400, err)
			return
		}

		client, err := fbApp.Auth(c)
		if err != nil {
				log.Fatalf("error getting Auth client: %v\n", err)
		}

		// Verify the ID token first.
		token, err := client.VerifyIDToken(c, reqBody.Token)
		if err != nil {
				log.Fatal(err)
		}

		email := token.Claims["email"]
		siginProvider := token.Firebase.SignInProvider
		res, reqErr := ExternalSignIn(sqlDB, email.(string), my_enums.CusSignInProvider(siginProvider))

		if reqErr != nil {
			fmt.Println("Failed to sign in with external provider: ", reqErr.Err)
			c.JSON(reqErr.StatusCode, res)
			return
		}

		c.JSON(http.StatusOK, res)
	}
}