package rest

import (
	"net/http"
	"strconv"

	"github.com/marco-almeida/gobank/internal/storage"
	u "github.com/marco-almeida/gobank/pkg/utils"
	"github.com/sirupsen/logrus"
)

type AccountsService struct {
	log   *logrus.Logger
	store storage.Storer
}

func NewAccountsService(logger *logrus.Logger, s storage.Storer) *AccountsService {
	return &AccountsService{
		log:   logger,
		store: s,
	}
}

func (s *AccountsService) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("GET /api/v1/users/{user_id}/accounts", JWTMiddleware(s.log, s.store, s.handleGetAllAccountsByID))
	r.HandleFunc("GET /api/v1/users/{user_id}/accounts/{account_id}", JWTMiddleware(s.log, s.store, s.handleGetAccountByID))
	r.HandleFunc("POST /api/v1/users/{user_id}/accounts", JWTMiddleware(s.log, s.store, s.handleCreateAccount))
}

func (s *AccountsService) handleGetAllAccountsByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	id, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		s.log.Infof("Invalid user id: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid user id"})
		return
	}

	accounts, err := s.store.GetAllAccountsByUserID(id)
	if err != nil {
		s.log.Infof("Failed to get accounts: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Failed to get accounts"})
		return
	}

	u.WriteJSON(w, http.StatusOK, accounts)
}

func (s *AccountsService) handleGetAccountByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		s.log.Infof("Invalid user id: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid user id"})
		return
	}

	accountIDStr := r.PathValue("account_id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		s.log.Infof("Invalid account id: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid account id"})
		return
	}

	account, err := s.store.GetAccountByID(userID, accountID)
	if err != nil {
		s.log.Infof("Failed to get account: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Failed to get account"})
		return
	}

	u.WriteJSON(w, http.StatusOK, account)
}

func (s *AccountsService) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		s.log.Infof("Invalid user id: %v", err)
		u.WriteJSON(w, http.StatusBadRequest, u.ErrorResponse{Error: "Invalid user id"})
		return
	}

	err = s.store.CreateAccount(userID)
	if err != nil {
		s.log.Infof("Failed to create account: %v", err)
		u.WriteJSON(w, http.StatusInternalServerError, u.ErrorResponse{Error: "Failed to create account"})
		return
	}

	w.WriteHeader(http.StatusCreated)
}
