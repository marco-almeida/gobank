package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
)

var JWTSecret *string = new(string)

func InitAuth(secret string) {
	*JWTSecret = secret
}

func JWTMiddleware(log *logrus.Logger, userService UserService, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		// check if type is Bearer
		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			log.Info("invalid token format")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenString = tokenString[7:]

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
		tokenUserIDStr := claims["userID"].(string)
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

		userID, err := strconv.ParseInt(tokenUserIDStr, 10, 64)
		if err != nil {
			log.Infof("failed to parse userID: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// if path contains user_id path param, compare it to the user_id in the token
		pathIDStr := r.PathValue("user_id")
		if pathIDStr != "" {
			if pathIDStr != tokenUserIDStr {
				log.Info("userID in token does not match path param")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		_, err = userService.Get(userID)
		if err != nil {
			log.Infof("failed to get user by id: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Call the function if the token is valid
		handlerFunc(w, r)
	}
}

func validateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(*JWTSecret), nil
	})
}
