package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
	"github.com/marco-almeida/mybank/internal/service"
)

// UserService defines the methods that the user handler will use
type UserService interface {
	// GetAll(context context.Context, limit, offset int64) ([]db.User, error)
	// Get(context context.Context, id int64) (db.User, error)
	// GetByEmail(context context.Context, email string) (db.User, error)
	Create(context context.Context, user db.CreateUserParams) (db.User, error)
	Get(context context.Context, username string) (db.User, error)
	// Delete(context context.Context, id int64) error
	// Update(context context.Context, user db.UpdateUserParams) (db.User, error)
	// PartialUpdate(context context.Context, id int64, user db.User) (db.User, error)
}

// UserHandler is the handler for the user service
type UserHandler struct {
	userSvc UserService
	authSvc AuthService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userSvc UserService, authSvc AuthService) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		authSvc: authSvc,
	}
}

// RegisterRoutes connects the handlers to the router
func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/api/v1/users", h.handleCreateUser)
	r.POST("/api/v1/users/login", h.handleLoginUser)
	r.POST("/api/v1/users/renew_access", h.handleRenewAccessToken)
	// r.HandleFunc("GET /v1/users", h.authSvc.WithJWTMiddleware(h.handleGetAllUsers))
	// r.HandleFunc("GET /v1/users/{user_id}", h.authSvc.WithJWTMiddleware(h.handleGetUser))
	// r.HandleFunc("DELETE /v1/users/{user_id}", h.authSvc.WithJWTMiddleware(h.handleUserDelete))
	// r.HandleFunc("PUT /v1/users/{user_id}", h.authSvc.WithJWTMiddleware(h.handleUpdateUser))
	// r.HandleFunc("PATCH /v1/users/{user_id}", h.authSvc.WithJWTMiddleware(h.handlePartialUpdateUser))
}

type createUserRequest struct {
	Username string `json:"username" binding:"required,alphanum"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func newUserResponse(user db.User) userResponse {
	return userResponse{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: user.PasswordChangedAt,
		CreatedAt:         user.CreatedAt,
	}
}

func (h *UserHandler) handleCreateUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := service.CreateUserParams{
		Username:          req.Username,
		PlaintextPassword: req.Password,
		FullName:          req.FullName,
		Email:             req.Email,
	}

	user, err := h.authSvc.Create(ctx, arg)
	if err != nil {
		if db.ErrorCode(err) == db.UniqueViolation {
			ctx.JSON(http.StatusForbidden, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	rsp := newUserResponse(user)
	ctx.JSON(http.StatusOK, rsp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) handleLoginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	rsp, err := h.authSvc.Login(ctx, service.LoginUserParams{
		Username:  req.Username,
		Password:  req.Password,
		UserAgent: ctx.Request.UserAgent(),
		ClientIP:  ctx.ClientIP(),
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, rsp)
}

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

func (h *UserHandler) handleRenewAccessToken(ctx *gin.Context) {
	var req renewAccessTokenRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	rsp, err := h.authSvc.RenewAccessToken(ctx, service.RenewAccessTokenRequest{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, rsp)
}
