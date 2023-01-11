package middleware

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/johnyeocx/usual/server/db"
	"github.com/johnyeocx/usual/server/db/models"
	"github.com/johnyeocx/usual/server/utils/secure"
)

type contextKey struct {
	key string
}
var UserCtxKey = contextKey{key: "user_id"}

func AuthMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		

		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			log.Printf("No authorization header found\n")
			c.Next()
			return
		}

		userId, err := secure.ParseAccessToken(authHeader[len(BEARER_SCHEMA):])
		
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
		c.Next()
	}
}

func UserCtx(c *gin.Context) (interface{}, error) {
	userID, exists := c.Get(UserCtxKey.key)

	if !exists {
		err := models.RequestError{
			StatusCode: http.StatusUnauthorized, 
			Err: errors.New("user id not found in context"),
		}
		return nil, err.Err
	}
	return userID, nil
}


func AuthenticateId(c *gin.Context, sqlDB *sql.DB) (*int, error) {

	businessId, err := UserCtx(c)
	
	if err != nil {
		return nil, err
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

	customerId, err := UserCtx(c)
	
	if err != nil {
		return nil, err
	}

	customerIdInt, err := strconv.Atoi(customerId.(string))
	if err != nil {
		return nil, err
	}
	
	if ok := db.ValidateCustomerId(sqlDB, customerIdInt); !ok {
		return nil, fmt.Errorf("invalid business id")
	}

	return &customerIdInt, nil
}


