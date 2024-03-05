package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	t "github.com/marco-almeida/golang-api-project-layout/internal/types"
	u "github.com/marco-almeida/golang-api-project-layout/pkg/utils"
	"github.com/sirupsen/logrus"
)

type UserService struct {
	log *logrus.Logger
}

func NewUserService(logger *logrus.Logger) *UserService {
	return &UserService{
		log: logger,
	}
}

func (s *UserService) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("POST /api/v1/users/register", s.handleUserRegister)
	r.HandleFunc("POST /api/v1/users/{id}/login", s.handleUserLogin)
	r.HandleFunc("DELETE /api/v1/users/{id}", s.handleUserDelete)
}

func (s *UserService) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	// user has first name, last name, email, password
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
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: err.Error()})
		return
	}

	u.WriteJSON(w, http.StatusOK, payload)
}

func (s *UserService) handleUserLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("User login with id %s", r.PathValue("id"))))
}

func (s *UserService) handleUserDelete(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("User logout"))
}
