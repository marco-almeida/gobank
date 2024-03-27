package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/marco-almeida/gobank/internal/service"
	"github.com/marco-almeida/gobank/internal/types"
)

// UserService ...
type UserService interface {
	GetAll(limit, offset int64) ([]types.User, error)
	Get(id int64) (types.User, error)
	Create(user types.User) (types.User, error)
	Delete(id int64) error
	Update(id int64, user types.User) (types.User, error)
	PartialUpdate(id int64, user types.User) (types.User, error)
}

// UserHandler ...
type UserHandler struct {
	svc service.UsersService
}

// NewUser ...
func NewUser(svc service.UsersService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

// RegisterRoutes connects the handlers to the router
func (h *UserHandler) RegisterRoutes(r *http.ServeMux) {
	// r.HandleFunc("GET /api/v1/users", h.handleGetAllUsers)
	// r.HandleFunc("GET /api/v1/users/{user_id}", h.handleGetUser)
	r.HandleFunc("POST /api/v1/users/register", h.handleUserRegister)
	// r.HandleFunc("POST /api/v1/users/login", h.handleUserLogin)
	// r.HandleFunc("DELETE /api/v1/users/{user_id}", h.handleUserDelete)
	// r.HandleFunc("PUT /api/v1/users/{user_id}", h.handleUpdateUser)
	// r.HandleFunc("PATCH /api/v1/users/{user_id}", h.handlePartialUpdateUser)
}

// RegisterUserRequest defines the request payload for registering a new user
type RegisterUserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (r *RegisterUserRequest) Validate() error {
	// iterate over struct fields
	val := reflect.ValueOf(r).Elem()
	for i := 0; i < val.NumField(); i++ {
		// if attribute value is empty, return error
		if val.Field(i).Interface() == "" {
			return errors.New(val.Type().Field(i).Tag.Get("json") + " is required")
		}
	}
	return nil
}

type User struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"-"`
}

func (h *UserHandler) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	var payload RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "Invalid request payload"})
		return
	}

	if err := payload.Validate(); err != nil {
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	err := h.svc.Create(types.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  payload.Password,
	})

	if err != nil {
		// return error response according to the error

	}

	w.WriteHeader(http.StatusCreated)
}
