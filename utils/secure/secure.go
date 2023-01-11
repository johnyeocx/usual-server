package secure

import (
	"golang.org/x/crypto/bcrypt"
)

var (
	DefaultDifficulty int = 6
	// accessTokenExpiry = time.Minute * 15
	// refreshTokenExpiry = time.Hour * 500
)

func GenerateHashFromStr(str string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(str), DefaultDifficulty)
	return string(bytes), err
}

func StringMatchesHash(str, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(str))
    return err == nil
}
