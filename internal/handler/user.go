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
	GetByEmail(email string) (internal.User, error)
	Create(user internal.User) error
	Delete(id int64) error
	Update(id int64, user internal.User) (internal.User, error)
	PartialUpdate(id int64, user internal.User) (internal.User, error)
}

// use a single instance of Validate, it caches struct info
var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

// UserHandler is the handler for the user service
type UserHandler struct {
	svc     UserService
	log     *logrus.Logger
	authSvc AuthService
}

// NewUser creates a new user handler
func NewUser(svc UserService, logger *logrus.Logger, authSvc AuthService) *UserHandler {
	return &UserHandler{
		svc:     svc,
		log:     logger,
		authSvc: authSvc,
	}
}

// RegisterRoutes connects the handlers to the router
func (h *UserHandler) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("GET /v1/users", h.authSvc.WithJWTMiddleware(h.handleGetAllUsers))
	r.HandleFunc("GET /v1/users/{user_id}", h.authSvc.WithJWTMiddleware(h.handleGetUser))
	r.HandleFunc("DELETE /v1/users/{user_id}", h.authSvc.WithJWTMiddleware(h.handleUserDelete))
	r.HandleFunc("PUT /v1/users/{user_id}", h.authSvc.WithJWTMiddleware(h.handleUpdateUser))
	r.HandleFunc("PATCH /v1/users/{user_id}", h.authSvc.WithJWTMiddleware(h.handlePartialUpdateUser))
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

func (h *UserHandler) handleUserDelete(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid user id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid user id"))
		return
	}

	err = h.svc.Delete(userID)
	if err != nil {
		h.log.Errorf("error deleting user: %v", err)
		WriteErrorResponse(w, r, "error deleting user", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid user id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid user id"))
		return
	}

	var userPayload RegisterUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userPayload); err != nil {
		WriteErrorResponse(w, r, "error decoding payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "error decoding payload"))
		return
	}

	if err := validate.Struct(userPayload); err != nil {
		WriteErrorResponse(w, r, "invalid payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid payload"))
		return
	}

	user, err := h.svc.Update(userID, internal.User{
		FirstName: userPayload.FirstName,
		LastName:  userPayload.LastName,
		Email:     userPayload.Email,
		Password:  userPayload.Password,
	})
	if err != nil {
		h.log.Errorf("error updating user: %v", err)
		var ierr *internal.Error
		// let user know if the email is already in use
		if errors.As(err, &ierr) {
			if ierr.Code() == internal.ErrorCodeDuplicate {
				WriteErrorResponse(w, r, ierr.Message(), ierr)
				return
			}
		}
		WriteErrorResponse(w, r, "error updating user", err)
		return
	}

	WriteJSON(w, http.StatusOK, user)
}

type PartialUpdateUserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func (h *UserHandler) handlePartialUpdateUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid user id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid user id"))
		return
	}

	var userPayload PartialUpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&userPayload); err != nil {
		WriteErrorResponse(w, r, "error decoding payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "error decoding payload"))
		return
	}

	if userPayload.Password != "" && (len(userPayload.Password) < 8 || len(userPayload.Password) > 64) {
		WriteErrorResponse(w, r, "invalid password", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid password"))
		return
	}

	user, err := h.svc.PartialUpdate(userID, internal.User{
		FirstName: userPayload.FirstName,
		LastName:  userPayload.LastName,
		Email:     userPayload.Email,
		Password:  userPayload.Password,
	})
	if err != nil {
		h.log.Errorf("error updating user: %v", err)
		var ierr *internal.Error
		// let user know if the email is already in use
		if errors.As(err, &ierr) {
			if ierr.Code() == internal.ErrorCodeDuplicate {
				WriteErrorResponse(w, r, ierr.Message(), ierr)
				return
			}
		}
		WriteErrorResponse(w, r, "error updating user", err)
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
