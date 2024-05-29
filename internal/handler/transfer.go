package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco-almeida/mybank/internal/pkg"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
	"github.com/marco-almeida/mybank/internal/token"
)

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
	authRoutes := r.Group("/").Use(authMiddleware(tokenMaker, []string{pkg.DepositorRole}))
	authRoutes.POST("/api/v1/transfers", h.handleCreateTransfer)
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
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	fromAccount, err := h.accountSvc.Get(ctx, req.FromAccountID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if fromAccount.Currency != req.Currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", fromAccount.ID, fromAccount.Currency, req.Currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	if fromAccount.Owner != authPayload.Username {
		err := errors.New("from account doesn't belong to the authenticated user")
		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
		return
	}

	toAccount, err := h.accountSvc.Get(ctx, req.ToAccountID)
	if err != nil {
		if errors.Is(err, db.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}

		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	if toAccount.Currency != req.Currency {
		err := fmt.Errorf("account [%d] currency mismatch: %s vs %s", toAccount.ID, toAccount.Currency, req.Currency)
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.TransferTxParams{
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
	}

	result, err := h.transferSvc.CreateTx(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, result)
}
