package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco-almeida/mybank/internal"
	"github.com/marco-almeida/mybank/internal/middleware"
	"github.com/marco-almeida/mybank/internal/pkg"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
	"github.com/marco-almeida/mybank/internal/token"
)

// AccountService defines the methods that the account handler will use
type AccountService interface {
	Create(context context.Context, account db.CreateAccountParams) (db.Account, error)
	Get(context context.Context, id int64) (db.Account, error)
	List(ctx context.Context, arg db.ListAccountsParams) ([]db.Account, error)
	Delete(ctx context.Context, id int64) error
}

// AccountHandler is the handler for the account service
type AccountHandler struct {
	accountSvc AccountService
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(accountSvc AccountService) *AccountHandler {
	return &AccountHandler{
		accountSvc: accountSvc,
	}
}

// RegisterRoutes connects the handlers to the router
func (h *AccountHandler) RegisterRoutes(r *gin.Engine, tokenMaker token.Maker) {
	authRoutes := r.Group("/api").Use(middleware.Authentication(tokenMaker, []string{pkg.DepositorRole, pkg.BankerRole}))
	authRoutes.POST("/v1/accounts", h.handleCreateAccount)
	authRoutes.GET("/v1/accounts/:id", h.handleGetAccount)
	authRoutes.GET("/v1/accounts", h.handleListAccounts)

	adminRoutes := r.Group("/api").Use(middleware.Authentication(tokenMaker, []string{pkg.BankerRole}))
	adminRoutes.DELETE("/v1/accounts/:id", h.handleDeleteAccount) // only accessible by bank workers (or admins)
}

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

func (h *AccountHandler) handleCreateAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := h.accountSvc.Create(ctx, arg)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, account)
}

type getAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (h *AccountHandler) handleGetAccount(ctx *gin.Context) {
	var req getAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	account, err := h.accountSvc.Get(ctx, req.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	overridePermission := ctx.MustGet(middleware.OverridePermissionKey).(bool)
	// needs refactoring
	if !overridePermission {
		if account.Owner != authPayload.Username {
			err := errors.New("account doesn't belong to the authenticated user")
			ctx.Error(fmt.Errorf("%w: %s", internal.ErrNoRows, err.Error())) // user shouldnt know about other accounts
			return
		}
	}

	ctx.JSON(http.StatusOK, account)
}

type listAccountRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

func (h *AccountHandler) handleListAccounts(ctx *gin.Context) {
	var req listAccountRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	arg := db.ListAccountsParams{
		Owner:  authPayload.Username,
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	accounts, err := h.accountSvc.List(ctx, arg)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, accounts)
}

type deleteAccountRequest struct {
	ID int64 `uri:"id" binding:"required,min=1"`
}

func (h *AccountHandler) handleDeleteAccount(ctx *gin.Context) {
	var req deleteAccountRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	err := h.accountSvc.Delete(ctx, req.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}
