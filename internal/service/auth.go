package service

import (
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var JWTSecret *string = new(string)

func InitAuth(secret string) {
	*JWTSecret = secret
}

func CreateJWT(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(int(userID)),
		"expiresAt": time.Now().Add(3 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(*JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, err
}
