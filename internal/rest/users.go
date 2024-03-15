package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/lib/pq"
	"github.com/marco-almeida/gobank/internal/storage"
	t "github.com/marco-almeida/gobank/internal/types"
	u "github.com/marco-almeida/gobank/pkg/utils"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	log   *logrus.Logger
	store storage.Storer
}

func NewUserService(logger *logrus.Logger, s storage.Storer) *UserService {
	return &UserService{
		log:   logger,
		store: s,
	}
}

func (s *UserService) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("GET /api/v1/users", JWTMiddleware(s.log, s.store, s.handleGetAllUsers))
	r.HandleFunc("GET /api/v1/users/{user_id}", JWTMiddleware(s.log, s.store, s.handleGetUser))
	r.HandleFunc("POST /api/v1/users/register", s.handleUserRegister)
	r.HandleFunc("POST /api/v1/users/login", s.handleUserLogin)
	r.HandleFunc("DELETE /api/v1/users/{user_id}", JWTMiddleware(s.log, s.store, s.handleUserDelete))
	r.HandleFunc("PUT /api/v1/users/{user_id}", JWTMiddleware(s.log, s.store, s.handleUpdateUser))
	r.HandleFunc("PATCH /api/v1/users/{user_id}", JWTMiddleware(s.log, s.store, s.handlePartialUpdateUser))
}

func (s *UserService) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Errorf("Error reading request body: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error reading request body"})
		return
	}

	defer r.Body.Close()

	var payload t.RegisterUserRequest
	err = json.Unmarshal(body, &payload)
	if err != nil {
		s.log.Infof("Invalid request payload: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	err = t.ValidateRegisterUserRequest(&payload)
	if err != nil {
		s.log.Infof("Invalid request payload: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: err.Error()})
		return
	}

	err = s.store.CreateUser(&payload)
	if err != nil {
		s.log.Infof("Error creating user: %v", err)
		// check if error of type duplicate key
		pgErr, ok := err.(*pq.Error)
		if ok {
			if pgErr.Code == "23505" {
				u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "email address is already in use"})
				return
			}
		}
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error creating user"})
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *UserService) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	// 3. Create JWT and set it in a cookie
	// 4. Return JWT in response
	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Errorf("Error reading request body: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error reading request body"})
		return
	}

	defer r.Body.Close()

	var payload t.LoginUserRequest
	err = json.Unmarshal(body, &payload)
	if err != nil {
		s.log.Infof("Invalid request payload: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	user, err := s.store.GetUserByEmail(payload.Email)
	if err != nil {
		s.log.Infof("Error getting user: %v", err)
		u.WriteJSON(w, http.StatusUnauthorized, u.ErrorResponse{Error: "Access denied"})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		s.log.Infof("Invalid password: %v", err)
		u.WriteJSON(w, http.StatusUnauthorized, u.ErrorResponse{Error: "Access denied"})
		return
	}

	// create JWT
	token, err := CreateJWT(user.ID)
	if err != nil {
		s.log.Errorf("Error creating JWT: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error creating token"})
		return
	}

	s.log.Infof("User %d logged in, token %+v", user.ID, token)

	// set JWT in cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		HttpOnly: true,
	})

	// return JWT in response
	w.WriteHeader(http.StatusOK)
}

func (s *UserService) handleUserDelete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.log.Infof("Invalid user id: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid user id"})
		return
	}

	err = s.store.DeleteUserByID(id)
	if err != nil {
		s.log.Errorf("Error deleting user: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error deleting user"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *UserService) handleGetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := s.store.GetAllUsers()
	if err != nil {
		s.log.Errorf("Error getting users: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error getting users"})
		return
	}

	u.WriteJSON(w, http.StatusOK, users)
}

func (s *UserService) handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.log.Infof("Invalid user id: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid user id"})
		return
	}

	user, err := s.store.GetUserByID(id)
	if err != nil {
		s.log.Errorf("Error getting user: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error getting user"})
		return
	}

	u.WriteJSON(w, http.StatusOK, user)
}

func (s *UserService) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.log.Infof("Invalid user id: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid user id"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Errorf("Error reading request body: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error reading request body"})
		return
	}

	defer r.Body.Close()

	var payload t.RegisterUserRequest
	err = json.Unmarshal(body, &payload)
	if err != nil {
		s.log.Infof("Invalid request payload: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	err = t.ValidateRegisterUserRequest(&payload)
	if err != nil {
		s.log.Infof("Invalid request payload: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: err.Error()})
		return
	}

	// TODO: get id by jwt
	err = s.store.UpdateUserByID(id, &payload)
	if err != nil {
		s.log.Errorf("Error updating user: %v", err)

		pgErr, ok := err.(*pq.Error)
		if ok {
			if pgErr.Code == "23505" {
				u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "email address is already in use"})
				return
			}
		}

		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error updating user"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *UserService) handlePartialUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("user_id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		s.log.Infof("Invalid user id: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid user id"})
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		s.log.Errorf("Error reading request body: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error reading request body"})
		return
	}

	defer r.Body.Close()

	var payload t.RegisterUserRequest
	err = json.Unmarshal(body, &payload)
	if err != nil {
		s.log.Infof("Invalid request payload: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid request payload"})
		return
	}

	s.log.Infof("Partial update user %d: %+v", id, payload)
	err = s.store.PartialUpdateUserByID(id, &payload)
	if err != nil {
		s.log.Errorf("Error updating user: %v", err)

		pgErr, ok := err.(*pq.Error)
		if ok {
			if pgErr.Code == "23505" {
				u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "email address is already in use"})
				return
			}
		}

		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Error updating user"})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
