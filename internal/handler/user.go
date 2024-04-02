package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/marco-almeida/gobank/internal"
	"github.com/sirupsen/logrus"
)

// UserService defines the methods that the user handler will use
type UserService interface {
	GetAll(limit, offset int64) ([]internal.User, error)
	Get(id int64) (internal.User, error)
	Create(user internal.User) error
	Delete(id int64) error
	Update(id int64, user internal.User) (internal.User, error)
	PartialUpdate(id int64, user internal.User) (internal.User, error)
	// returns user id and jwt token
	Login(email, password string) (int64, string, error)
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

// UserHandler is the handler for the user service
type UserHandler struct {
	svc UserService
	log *logrus.Logger
}

// NewUser creates a new user handler
func NewUser(svc UserService, logger *logrus.Logger) *UserHandler {
	return &UserHandler{
		svc: svc,
		log: logger,
	}
}

// RegisterRoutes connects the handlers to the router
func (h *UserHandler) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("GET /v1/users", h.handleGetAllUsers)
	r.HandleFunc("GET /v1/users/{user_id}", h.handleGetUser)
	r.HandleFunc("POST /v1/users/register", h.handleUserRegister)
	r.HandleFunc("POST /v1/users/login", h.handleUserLogin)
	// r.HandleFunc("DELETE /v1/users/{user_id}", h.handleUserDelete)
	// r.HandleFunc("PUT /v1/users/{user_id}", h.handleUpdateUser)
	// r.HandleFunc("PATCH /v1/users/{user_id}", h.handlePartialUpdateUser)
}

// RegisterUserRequest defines the request payload for registering a new user
type RegisterUserRequest struct {
	FirstName string `json:"firstName" validate:"required"`
	LastName  string `json:"lastName" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=64"`
}

func (h *UserHandler) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteErrorResponse(w, r, "error decoding payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "error decoding payload"))
		return
	}

	if err := validate.Struct(payload); err != nil {
		h.log.Errorf("error validating payload: %v", err)
		WriteErrorResponse(w, r, "invalid payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid payload"))
		return
	}

	err := h.svc.Create(internal.User{
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

func (h *UserHandler) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	var payload LoginUserRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteErrorResponse(w, r, "error decoding payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "error decoding payload"))
		return
	}

	if err := validate.Struct(payload); err != nil {
		h.log.Errorf("error validating payload: %v", err)
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

type DataResponse struct {
	Data *[]internal.User `json:"data,omitempty"`
}

func (h *UserHandler) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := getLimitAndOffset(r)
	users, err := h.svc.GetAll(limit, offset)
	if err != nil {
		h.log.Errorf("error getting all users: %v", err)
		WriteErrorResponse(w, r, "error getting all users", err)
		return
	}

	if users == nil {
		users = []internal.User{}
	}

	WriteJSON(w, http.StatusOK, users)
}

func (h *UserHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid user id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid user id"))
		return
	}

	user, err := h.svc.Get(userID)
	if err != nil {
		h.log.Errorf("error getting user: %v", err)
		WriteErrorResponse(w, r, "error getting user", err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

// getLimitAndOffset returns the limit and offset from the request query parameters
func getLimitAndOffset(r *http.Request) (int64, int64) {
	limitStr := r.URL.Query().Get("limit")
	limit, err := strconv.ParseInt(limitStr, 10, 64)
	if err != nil {
		limit = 10
	}

	offsetStr := r.URL.Query().Get("offset")
	offset, err := strconv.ParseInt(offsetStr, 10, 64)
	if err != nil {
		offset = 0
	}

	return limit, offset
}
