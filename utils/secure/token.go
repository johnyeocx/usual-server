package secure

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var (
	accessTokenExpiry = time.Minute * 20
	refreshTokenExpiry = time.Hour * 500
)

func GenerateAccessToken(userID string, userType string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["user_type"] = userType

	claims["exp"] = time.Now().Add(accessTokenExpiry).Unix()
	tokenString, err := token.SignedString([]byte(secretKey))


	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateRefreshToken(userID string, userType string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	secretKey := os.Getenv("JWT_REFRESH_SECRET")

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID

	claims["user_type"] = userType


	claims["exp"] = time.Now().Add(refreshTokenExpiry).Unix()
	tokenString, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseAccessToken(tokenStr string) (string, string, error) {
	
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("invalid token: %v", token.Header["alg"])
		}
		secretKey := os.Getenv("JWT_ACCESS_SECRET")
		return []byte(secretKey), nil
	})

	if err != nil {
		fmt.Println("Invalid jwt token:", err)
		return "", "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := claims["user_id"].(string)
		userType := claims["user_type"].(string)
		return userId, userType, nil
		
	} else {
		log.Print("Err", err)
		return "", "", err
	}
}

func ParseRefreshToken(tokenStr string) (string, string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("invalid token: %v", token.Header["alg"])
		}
		secretKey := os.Getenv("JWT_REFRESH_SECRET")
		return []byte(secretKey), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)
		userType := claims["user_type"].(string)
		return userID, userType, nil
	} else {
		log.Print("Err", err)
		return "", "", err
	}
}

func GenerateTokensFromId(id int, userType string) (*string, *string, error) {
	accessToken, err := GenerateAccessToken(strconv.Itoa(id), userType)
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := GenerateRefreshToken(strconv.Itoa(id), userType)
	if err != nil {
		return nil, nil, err
	}

	return &accessToken, &refreshToken, nil
}