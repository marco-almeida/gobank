package rest

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	config "github.com/marco-almeida/gobank/configs"
	s "github.com/marco-almeida/gobank/internal/storage"
	"github.com/sirupsen/logrus"
)

func JWTMiddleware(log *logrus.Logger, store s.Storer, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")

		token, err := validateJWT(tokenString)
		if err != nil {
			log.Infof("failed to validate token: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			log.Info("invalid token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		userIDString := claims["userID"].(string)
		expiresAtFloat, ok := claims["expiresAt"].(float64)
		if !ok {
			log.Info("failed to get expiresAt from token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expiresAt := time.Unix(int64(expiresAtFloat), 0)
		if time.Now().After(expiresAt) {
			log.Info("token expired")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseInt(userIDString, 10, 64)
		if err != nil {
			log.Infof("failed to parse userID: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// if path contains user_id path param, compare it to the user_id in the token
		pathID := r.PathValue("user_id")
		if pathID != "" {
			if pathID != userIDString {
				log.Info("user_id in token does not match path param")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		_, err = store.GetUserByID(userID)
		if err != nil {
			log.Printf("failed to get user by id: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Call the function if the token is valid
		handlerFunc(w, r)
	}
}

func CreateJWT(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(int(userID)),
		"expiresAt": time.Now().Add(3 * time.Hour).Unix(),
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
