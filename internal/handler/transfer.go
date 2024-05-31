package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/marco-almeida/mybank/internal"
	"github.com/marco-almeida/mybank/internal/middleware"
	"github.com/marco-almeida/mybank/internal/pkg"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
	"github.com/marco-almeida/mybank/internal/token"
)

func init() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("currency", validCurrency)
	}
}

var validCurrency validator.Func = func(fieldLevel validator.FieldLevel) bool {
	if currency, ok := fieldLevel.Field().Interface().(string); ok {
		return pkg.IsSupportedCurrency(currency)
	}
	return false
}

// TransferService defines the methods that the transfer handler will use
type TransferService interface {
	CreateTx(context context.Context, arg db.TransferTxParams) (db.TransferTxResult, error)
}

// TransferHandler is the handler for the account service
type TransferHandler struct {
	transferSvc TransferService
	accountSvc  AccountService
}

// NewTransferHandler creates a new transfer handler
func NewTransferHandler(transferSvc TransferService, accountSvc AccountService) *TransferHandler {
	return &TransferHandler{
		transferSvc: transferSvc,
		accountSvc:  accountSvc,
	}
}

// RegisterRoutes connects the handlers to the router
func (h *TransferHandler) RegisterRoutes(r *gin.Engine, tokenMaker token.Maker) {
	authRoutes := r.Group("/api").Use(middleware.Authentication(tokenMaker, []string{pkg.DepositorRole}))
	authRoutes.POST("/v1/transfers", h.handleCreateTransfer)
}

type transferRequest struct {
	FromAccountID int64  `json:"from_account_id" binding:"required,min=1"`
	ToAccountID   int64  `json:"to_account_id" binding:"required,min=1"`
	Amount        int64  `json:"amount" binding:"required,gt=0"`
	Currency      string `json:"currency" binding:"required,currency"`
}

func (h *TransferHandler) handleCreateTransfer(ctx *gin.Context) {
	var req transferRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	fromAccount, err := h.accountSvc.Get(ctx, req.FromAccountID)
	if err != nil {
		if errors.Is(err, internal.ErrNoRows) {
			ctx.Error(fmt.Errorf("%w: %w", internal.ErrInvalidFromAccount, err))
			return
		}
		ctx.Error(err)
		return
	}

	if fromAccount.Currency != req.Currency {
		ctx.Error(internal.ErrCurrencyMismatch)
		return
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		ctx.Error(fmt.Errorf("%w; from account doesn't belong to the authenticated user: %w", internal.ErrForbidden, err))
		return
	}

	toAccount, err := h.accountSvc.Get(ctx, req.ToAccountID)
	if err != nil {
		if errors.Is(err, internal.ErrNoRows) {
			ctx.Error(fmt.Errorf("%w: %w", internal.ErrInvalidToAccount, err))
			return
		}
		ctx.Error(err)
		return
	}

	if toAccount.Currency != req.Currency {
		err := fmt.Errorf("%w: account [%d] currency mismatch: %s vs %s", internal.ErrCurrencyMismatch, toAccount.ID, toAccount.Currency, req.Currency)
		ctx.Error(err)
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := h.transferSvc.CreateTx(ctx, arg)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, result)
}
