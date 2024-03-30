package handler

import (
	"encoding/json"
	"errors"
	"net/http"

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
	// r.HandleFunc("GET /v1/users", h.handleGetAllUsers)
	// r.HandleFunc("GET /v1/users/{user_id}", h.handleGetUser)
	r.HandleFunc("POST /v1/users/register", h.handleUserRegister)
	// r.HandleFunc("POST /v1/users/login", h.handleUserLogin)
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
		if !errors.As(err, &ierr) {
			WriteErrorResponse(w, r, "error creating user", err)
			return
		}
		// let user know if the email is already in use
		if ierr.Code() == internal.ErrorCodeDuplicate {
			WriteErrorResponse(w, r, ierr.Message(), ierr)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
}
