package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/marco-almeida/gobank/internal"
	"github.com/sirupsen/logrus"
)

type AuthService interface {
	// Login returns user id and jwt token
	Login(email, password string) (int64, string, error)
	// Register hashes password and saves user
	Register(u internal.User) error
	WithJWTMiddleware(handlerFunc http.HandlerFunc) http.HandlerFunc
	CreateJWT(userID int64) (string, error)
}

type AuthHandler struct {
	svc AuthService
	log *logrus.Logger
}

func NewAuth(svc AuthService, logger *logrus.Logger) *AuthHandler {
	return &AuthHandler{
		svc: svc,
		log: logger,
	}
}

func (h *AuthHandler) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("POST /v1/auth/register", h.handleRegister)
	r.HandleFunc("POST /v1/auth/login", h.handleLogin)
}

// RegisterUserRequest defines the request payload for registering a new user
type RegisterUserRequest struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=64"`
}

func (h *AuthHandler) handleRegister(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteErrorResponse(w, r, "error decoding payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "error decoding payload"))
		return
	}

	if err := validate.Struct(payload); err != nil {
		WriteErrorResponse(w, r, "invalid payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid payload"))
		return
	}

	err := h.svc.Register(internal.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  payload.Password,
	})

	if err != nil {
		h.log.Errorf("error creating user: %v", err)
		var ierr *internal.Error
		// let user know if the email is already in use
		if errors.As(err, &ierr) {
			if ierr.Code() == internal.ErrorCodeDuplicate {
				WriteErrorResponse(w, r, ierr.Message(), ierr)
				return
			}
		}
		WriteErrorResponse(w, r, "error creating user", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// LoginUserRequest defines the request payload for logging in a user
type LoginUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	var payload LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteErrorResponse(w, r, "error decoding payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "error decoding payload"))
		return
	}

	if err := validate.Struct(payload); err != nil {
		WriteErrorResponse(w, r, "invalid payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid payload"))
		return
	}

	userID, token, err := h.svc.Login(payload.Email, payload.Password)
	if err != nil {
		h.log.Errorf("error logging in user: %v", err)
		WriteErrorResponse(w, r, "invalid credentials", err)
		return
	}

	// set JWT in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		Secure:   true,
		HttpOnly: true,
	})

	// return user id
	WriteJSON(w, http.StatusOK, map[string]int64{"userID": userID})
}
