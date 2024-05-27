package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// AccountService defines the methods that the account handler will use
type AccountService interface {
	Create(context context.Context, account db.CreateAccountParams) (db.Account, error)
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
func (h *AccountHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/api/v1/accounts", h.handleCreateAccount)
}

type createAccountRequest struct {
	Currency string `json:"currency" binding:"required,currency"`
}

func (h *AccountHandler) handleCreateAccount(ctx *gin.Context) {
	var req createAccountRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	arg := db.CreateAccountParams{
		Owner:    "authPayloadUsername",
		Currency: req.Currency,
		Balance:  0,
	}

	account, err := h.accountSvc.Create(ctx, arg)
	if err != nil {
		errCode := db.ErrorCode(err)
		if errCode == db.ForeignKeyViolation || errCode == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, account)
}

// type getAccountRequest struct {
// 	ID int64 `uri:"id" binding:"required,min=1"`
// }

// func (server *Server) getAccount(ctx *gin.Context) {
// 	var req getAccountRequest
// 	if err := ctx.ShouldBindUri(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	account, err := h.store.GetAccount(ctx, req.ID)
// 	if err != nil {
// 		if errors.Is(err, db.ErrRecordNotFound) {
// 			ctx.JSON(http.StatusNotFound, errorResponse(err))
// 			return
// 		}

// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
// 	if account.Owner != authPayload.Username {
// 		err := errors.New("account doesn't belong to the authenticated user")
// 		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, account)
// }

// type listAccountRequest struct {
// 	PageID   int32 `form:"page_id" binding:"required,min=1"`
// 	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
// }

// func (server *Server) listAccounts(ctx *gin.Context) {
// 	var req listAccountRequest
// 	if err := ctx.ShouldBindQuery(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
// 	arg := db.ListAccountsParams{
// 		Owner:  authPayload.Username,
// 		Limit:  req.PageSize,
// 		Offset: (req.PageID - 1) * req.PageSize,
// 	}

// 	accounts, err := h.store.ListAccounts(ctx, arg)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	ctx.JSON(http.StatusOK, accounts)
// }
