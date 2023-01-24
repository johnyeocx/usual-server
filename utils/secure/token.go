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

func GenerateAccessToken(userID string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	secretKey := os.Getenv("JWT_ACCESS_SECRET")
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID

	claims["exp"] = time.Now().Add(accessTokenExpiry).Unix()
	tokenString, err := token.SignedString([]byte(secretKey))


	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func GenerateRefreshToken(userID string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	secretKey := os.Getenv("JWT_REFRESH_SECRET")

	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = userID
	claims["exp"] = time.Now().Add(refreshTokenExpiry).Unix()
	tokenString, err := token.SignedString([]byte(secretKey))

	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func ParseAccessToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("invalid token: %v", token.Header["alg"])
		}
		secretKey := os.Getenv("JWT_ACCESS_SECRET")
		return []byte(secretKey), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userId := claims["user_id"].(string)
		return userId, nil
		
	} else {
		log.Print("Err", err)
		return "", err
	}
}

func ParseRefreshToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, isvalid := token.Method.(*jwt.SigningMethodHMAC); !isvalid {
			return nil, fmt.Errorf("invalid token: %v", token.Header["alg"])
		}
		secretKey := os.Getenv("JWT_REFRESH_SECRET")
		return []byte(secretKey), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID := claims["user_id"].(string)
		return userID, nil
	} else {
		log.Print("Err", err)
		return "", err
	}
}

func GenerateTokensFromId(id int) (*string, *string, error) {
	accessToken, err := GenerateAccessToken(strconv.Itoa(id))
	if err != nil {
		return nil, nil, err
	}

	refreshToken, err := GenerateRefreshToken(strconv.Itoa(id))
	if err != nil {
		return nil, nil, err
	}

	return &accessToken, &refreshToken, nil
}