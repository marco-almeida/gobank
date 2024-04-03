package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/marco-almeida/gobank/internal"
	"github.com/sirupsen/logrus"
)

// AccountService defines the methods that the account handler will use
type AccountService interface {
	Create(userID int64) error
	GetAllByUserID(userID, offset, limit int64) ([]internal.Account, error)
	GetByID(userID int64, accountID int64) (internal.Account, error)
	DeleteByID(userID, accountID int64) error
	UpdateBalanceByID(userID int64, accountID int64, balance internal.USD) (internal.Account, error)
}

// AccountHandler is the handler for the account service
type AccountHandler struct {
	svc AccountService
	log *logrus.Logger
}

// NewAccount creates a new account handler
func NewAccount(svc AccountService, logger *logrus.Logger) *AccountHandler {
	return &AccountHandler{
		svc: svc,
		log: logger,
	}
}

// RegisterRoutes connects the handlers to the router
func (h *AccountHandler) RegisterRoutes(r *http.ServeMux) {
	r.HandleFunc("POST /v1/users/{user_id}/accounts", h.handleCreateAccount)
	r.HandleFunc("GET /v1/users/{user_id}/accounts/{account_id}", h.handleGetAccountByID)
	r.HandleFunc("GET /v1/users/{user_id}/accounts", h.handleGetAllAccountsByID)
	r.HandleFunc("POST /v1/users/{user_id}/accounts/{account_id}/updateBalance", h.handleUpdateBalance)
	r.HandleFunc("DELETE /v1/users/{user_id}/accounts/{account_id}", h.handleDeleteAccount)
}

func (h *AccountHandler) handleCreateAccount(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid user id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid user id"))
		return
	}

	if err := h.svc.Create(userID); err != nil {
		h.log.Info(err)
		WriteErrorResponse(w, r, "error creating account", err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *AccountHandler) handleGetAccountByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid user id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid user id"))
		return
	}

	accountIDStr := r.PathValue("account_id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid account id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid account id"))
		return
	}

	account, err := h.svc.GetByID(userID, accountID)
	if err != nil {
		h.log.Info(err)
		WriteErrorResponse(w, r, "error getting account", err)
		return
	}

	WriteJSON(w, http.StatusOK, account)
}

func (h *AccountHandler) handleGetAllAccountsByID(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid user id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid user id"))
		return
	}

	limit, offset := getLimitAndOffset(r)
	accounts, err := h.svc.GetAllByUserID(userID, offset, limit)
	if err != nil {
		h.log.Info(err)
		WriteErrorResponse(w, r, "error getting all accounts", err)
		return
	}

	if accounts == nil {
		accounts = []internal.Account{}
	}

	WriteJSON(w, http.StatusOK, accounts)
}

type BalanceUpdateRequest struct {
	Amount internal.USD `json:"amount" validate:"required"`
}

func (h *AccountHandler) handleUpdateBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid user id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid user id"))
		return
	}

	accountIDStr := r.PathValue("account_id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid account id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid account id"))
		return
	}

	var req BalanceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteErrorResponse(w, r, "error decoding payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "error decoding payload"))
		return
	}

	if err := validate.Struct(req); err != nil {
		WriteErrorResponse(w, r, "invalid payload", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid payload"))
		return
	}

	newAccount, err := h.svc.UpdateBalanceByID(userID, accountID, req.Amount)
	if err != nil {
		h.log.Info(err)
		WriteErrorResponse(w, r, "error updating balance", err)
		return
	}

	WriteJSON(w, http.StatusOK, newAccount)
}

func (h *AccountHandler) handleDeleteAccount(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid user id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid user id"))
		return
	}

	accountIDStr := r.PathValue("account_id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		WriteErrorResponse(w, r, "invalid account id", internal.WrapErrorf(err, internal.ErrorCodeInvalidArgument, "invalid account id"))
		return
	}

	if err := h.svc.DeleteByID(userID, accountID); err != nil {
		h.log.Info(err)
		WriteErrorResponse(w, r, "error deleting account", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
