package service

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/marco-almeida/gobank/internal"
	"github.com/marco-almeida/gobank/internal/handler"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	log       *logrus.Logger
	userSvc   handler.UserService
	JWTSecret string
}

func NewAuth(log *logrus.Logger, userSvc handler.UserService, JWTSecret string) *Auth {
	return &Auth{
		log:       log,
		userSvc:   userSvc,
		JWTSecret: JWTSecret,
	}
}

func (h *Auth) WithJWTMiddleware(handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		// check if type is Bearer
		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			h.log.Info("invalid token format")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenString = tokenString[7:]

		token, err := h.validateJWT(tokenString)
		if err != nil {
			h.log.Infof("failed to validate token: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			h.log.Info("invalid token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		tokenUserIDStr := claims["userID"].(string)
		expiresAtFloat, ok := claims["expiresAt"].(float64)
		if !ok {
			h.log.Info("failed to get expiresAt from token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		expiresAt := time.Unix(int64(expiresAtFloat), 0)
		if time.Now().After(expiresAt) {
			h.log.Info("token expired")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseInt(tokenUserIDStr, 10, 64)
		if err != nil {
			h.log.Infof("failed to parse userID: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// if path contains user_id path param, compare it to the user_id in the token
		pathIDStr := r.PathValue("user_id")
		if pathIDStr != "" {
			if pathIDStr != tokenUserIDStr {
				h.log.Info("userID in token does not match path param")
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}

		_, err = h.userSvc.Get(userID)
		if err != nil {
			h.log.Infof("failed to get user by id: %v", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Call the function if the token is valid
		handlerFunc(w, r)
	}
}

func (h *Auth) Login(email, password string) (int64, string, error) {
	user, err := h.userSvc.GetByEmail(email)
	if err != nil {
		return 0, "", internal.WrapErrorf(err, internal.ErrorCodeUnauthorized, "failed to get user by email")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return 0, "", internal.WrapErrorf(err, internal.ErrorCodeUnauthorized, "invalid password")
	}

	token, err := h.CreateJWT(user.ID)

	return user.ID, token, err
}

func (h *Auth) Register(u internal.User) error {
	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "failed to hash password")
	}
	u.Password = string(hashedPassword)
	return h.userSvc.Create(u)
}

func (h *Auth) validateJWT(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(h.JWTSecret), nil
	})
}

func (h *Auth) CreateJWT(userID int64) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":    strconv.Itoa(int(userID)),
		"expiresAt": time.Now().Add(3 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.JWTSecret))
	if err != nil {
		return "", err
	}

	return tokenString, err
}
