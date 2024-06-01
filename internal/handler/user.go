package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/marco-almeida/mybank/internal"
	"github.com/marco-almeida/mybank/internal/middleware"
	"github.com/marco-almeida/mybank/internal/pkg"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
	"github.com/marco-almeida/mybank/internal/service"
	"github.com/marco-almeida/mybank/internal/token"
)

// UserService defines the methods that the user handler will use
type UserService interface {
	// Create(context context.Context, user service.CreateUserParams) (db.User, error)
	Get(context context.Context, username string) (db.User, error)
	Login(context context.Context, req service.LoginUserParams) (service.LoginUserResponse, error)
	Create(ctx context.Context, arg service.CreateUserParams) (db.User, error)
	RenewAccessToken(context context.Context, req service.RenewAccessTokenParams) (service.RenewAccessTokenResponse, error)
	VerifyEmail(ctx context.Context, req db.VerifyEmailTxParams) (db.VerifyEmailTxResult, error)
	Update(ctx context.Context, arg service.UpdateUserParams) (db.User, error)
}

// UserHandler is the handler for the user service
type UserHandler struct {
	userSvc UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(userSvc UserService) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
	}
}

// RegisterRoutes connects the handlers to the router
func (h *UserHandler) RegisterRoutes(r *gin.Engine, tokenMaker token.Maker) {
	groupRoutes := r.Group("/api")
	groupRoutes.POST("/v1/users", h.handleCreateUser)
	groupRoutes.POST("/v1/users/login", h.handleLoginUser)
	groupRoutes.POST("/v1/users/renew_access", h.handleRenewAccessToken)
	groupRoutes.GET("/v1/users/verify_email", h.handleVerifyEmail)

	groupRoutes.PATCH("/v1/users/:username", middleware.Authentication(tokenMaker, []string{pkg.DepositorRole, pkg.BankerRole}), h.handleUpdateUser)
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
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	arg := service.CreateUserParams{
		Username:          req.Username,
		PlaintextPassword: req.Password,
		FullName:          req.FullName,
		Email:             req.Email,
	}

	user, err := h.userSvc.Create(ctx, arg)
	if err != nil {
		ctx.Error(err)
		return
	}

	rsp := newUserResponse(user)
	ctx.JSON(http.StatusAccepted, rsp)
}

type loginUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) handleLoginUser(ctx *gin.Context) {
	var req loginUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	rsp, err := h.userSvc.Login(ctx, service.LoginUserParams{
		Username:  req.Username,
		Password:  req.Password,
		UserAgent: ctx.Request.UserAgent(),
		ClientIP:  ctx.ClientIP(),
	})

	if err != nil {
		ctx.Error(err)
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
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	rsp, err := h.userSvc.RenewAccessToken(ctx, service.RenewAccessTokenParams{
		RefreshToken: req.RefreshToken,
	})

	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, rsp)
}

type verifyEmailRequest struct {
	EmailID    int64  `form:"email_id" binding:"required"`
	SecretCode string `form:"secret_code" binding:"required"`
}

type verifyEmailResponse struct {
	IsEmailVerified bool `json:"success"`
}

func (h *UserHandler) handleVerifyEmail(ctx *gin.Context) {
	var req verifyEmailRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	result, err := h.userSvc.VerifyEmail(ctx, db.VerifyEmailTxParams{
		EmailId:    req.EmailID,
		SecretCode: req.SecretCode,
	})

	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, verifyEmailResponse{IsEmailVerified: result.User.IsEmailVerified})
}

type updateUserBodyRequest struct {
	FullName string `json:"full_name" `
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateUserUriRequest struct {
	Username string `uri:"username" binding:"required,alphanum"`
}

func (h *UserHandler) handleUpdateUser(ctx *gin.Context) {
	var req updateUserUriRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	var args updateUserBodyRequest
	if err := ctx.ShouldBindJSON(&args); err != nil {
		ctx.Error(fmt.Errorf("%w; %w", internal.ErrInvalidParams, err))
		return
	}

	arg := service.UpdateUserParams{
		Username:          req.Username,
		PlaintextPassword: args.Password,
		FullName:          args.FullName,
		Email:             args.Email,
	}

	authPayload := ctx.MustGet(middleware.AuthorizationPayloadKey).(*token.Payload)
	overridePermission := ctx.MustGet(middleware.OverridePermissionKey).(bool)
	if !overridePermission && req.Username != authPayload.Username {
		err := errors.New("user doesn't belong to the authenticated user")
		ctx.Error(fmt.Errorf("%w; user doesn't belong to the authenticated user: %w", internal.ErrForbidden, err))
		return
	}
	user, err := h.userSvc.Update(ctx, arg)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, newUserResponse(user))
}
