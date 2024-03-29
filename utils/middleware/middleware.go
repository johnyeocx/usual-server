package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/constants"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/secure"
)

type contextKey struct {
	key string
}
var UserCtxKey = contextKey{
	key: "user_id"}

var UserTypeCtxKey = contextKey{
	key: "user_type"}
	
func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		
		// accessToken, _ := c.Cookie("access_token")
		
		var accessToken string
		accessToken, err := c.Cookie("access_token")
		if err != nil || len(accessToken) == 0 {
			const BEARER_SCHEMA = "Bearer "
			authHeader := c.GetHeader("Authorization")
		
			if authHeader == "" || len(authHeader) < len("Bearer  "){
				c.Next()
				return
			}

			accessToken = authHeader[len(BEARER_SCHEMA):]
		}

	
		userId, userType, err := secure.ParseAccessToken(accessToken)
		
	
		if err != nil {
			log.Printf("Could not parse access token: %s", err.Error())
			c.Next();
			return
		}
		
		if err != nil {
			log.Printf("User ID from token is invalid: %s", err.Error())
			c.Next()
			return
		}

		

		c.Set(UserCtxKey.key, userId)
		c.Set(UserTypeCtxKey.key, userType)
		c.Next()
	}
}

func UserCtx(c *gin.Context) (interface{}, interface{}, error) {
	userID, exists := c.Get(UserCtxKey.key)

	if !exists {
		err := models.RequestError{
			StatusCode: http.StatusUnauthorized, 
			Err: errors.New("user id not found in context"),
		}

		return nil, nil, err.Err
	}

	userType, exists := c.Get(UserTypeCtxKey.key)

	if !exists {
		err := models.RequestError{
			StatusCode: http.StatusUnauthorized, 
			Err: errors.New("user type not found in context"),
		}
		return nil, nil, err.Err
	}
	return userID, userType,  nil
}

func AuthenticateBId(c *gin.Context, sqlDB *sql.DB) (*int, error) {

	businessId, userType, err := UserCtx(c)
	
	if err != nil {
		return nil, err
	}

	if userType != constants.UserTypes.Business {
		return nil, errors.New("wrong type token")
	}

	businessIdInt, err := strconv.Atoi(businessId.(string))
	if err != nil {
		return nil, err
	}
	
	if ok := db.ValidateBusinessId(sqlDB, businessIdInt); !ok {
		return nil, fmt.Errorf("invalid business id")
	}

	return &businessIdInt, nil
}

func AuthenticateCId(c *gin.Context, sqlDB *sql.DB) (*int, error) {

	customerId, cusType, err := UserCtx(c)



	
	if err != nil {
		return nil, err
	}
	
	if cusType != constants.UserTypes.Customer {
		return nil, errors.New("wrong type token")
	}

	customerIdInt, err := strconv.Atoi(customerId.(string))
	if err != nil {
		return nil, err
	}
	

	if ok := db.ValidateCustomerId(sqlDB, customerIdInt); !ok {
		return nil, fmt.Errorf("invalid customer id")
	}

	return &customerIdInt, nil
}


