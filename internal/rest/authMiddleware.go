package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	config "github.com/marco-almeida/gobank/configs"
	s "github.com/marco-almeida/gobank/internal/storage"
	u "github.com/marco-almeida/gobank/pkg/utils"
	"github.com/sirupsen/logrus"
)

func JWTMiddleware(log *logrus.Logger, store s.Storer, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		log.Infof("token obtained: %v", tokenString)

		token, err := validateJWT(tokenString)
		if err != nil {
			log.Infof("failed to validate token: %v", err)
			permissionDenied(w)
			return
		}

		if !token.Valid {
			log.Infof("invalid token")
			permissionDenied(w)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		log.Infof("claims: %v\nuserId: %v", claims, claims["userID"])
		userIDString := claims["userID"].(string)
		userID, err := strconv.ParseInt(userIDString, 10, 64)
		if err != nil {
			log.Printf("failed to parse userID: %v", err)
			permissionDenied(w)
			return
		}

		_, err = store.GetUserByID(userID)
		if err != nil {
			log.Printf("failed to get user by id: %v", err)
			permissionDenied(w)
			return
		}

		// Call the function if the token is valid
		handlerFunc(w, r)
	}
}

func CreateJWT(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(int(userID)),
		"expiresAt": time.Now().Add(time.Hour * 24 * 120).Unix(),
	})

	tokenString, err := token.SignedString([]byte(config.Envs.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, err
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	secret := config.Envs.JWTSecret

	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
}

func permissionDenied(w http.ResponseWriter) {
	u.WriteJSON(w, http.StatusUnauthorized, u.ErrorResponse{
		Error: fmt.Errorf("permission denied").Error(),
	})
}
