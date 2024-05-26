package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/marco-almeida/mybank/internal/pkg"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// UserService defines the methods that the user handler will use
type UserService interface {
	// GetAll(context context.Context, limit, offset int64) ([]db.User, error)
	// Get(context context.Context, id int64) (db.User, error)
	// GetByEmail(context context.Context, email string) (db.User, error)
	Create(context context.Context, user db.CreateUserParams) (db.User, error)
	// Delete(context context.Context, id int64) error
	// Update(context context.Context, user db.UpdateUserParams) (db.User, error)
	// PartialUpdate(context context.Context, id int64, user db.User) (db.User, error)
}

// UserHandler is the handler for the user service
type UserHandler struct {
	svc UserService
}

// NewUserHandler creates a new user handler
func NewUserHandler(svc UserService) *UserHandler {
	return &UserHandler{
		svc: svc,
	}
}

// RegisterRoutes connects the handlers to the router
func (h *UserHandler) RegisterRoutes(r *gin.Engine) {
	r.POST("/api/v1/users", h.handleCreateUser)
	// r.POST("/users/login", h.handleCreateUser)
	// r.POST("/users/renew_token", h.handleCreateUser)
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

func (server *UserHandler) handleCreateUser(ctx *gin.Context) {
	var req createUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	// TODO: call auth service

	hashedPassword, err := pkg.HashPassword(req.Password)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
	}

	user, err := server.svc.Create(ctx, arg)
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

// type loginUserRequest struct {
// 	Username string `json:"username" binding:"required,alphanum"`
// 	Password string `json:"password" binding:"required,min=6"`
// }

// type loginUserResponse struct {
// 	SessionID             uuid.UUID    `json:"session_id"`
// 	AccessToken           string       `json:"access_token"`
// 	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
// 	RefreshToken          string       `json:"refresh_token"`
// 	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
// 	User                  userResponse `json:"user"`
// }

// func (server *UserHandler) loginUser(ctx *gin.Context) {
// 	var req loginUserRequest
// 	if err := ctx.ShouldBindJSON(&req); err != nil {
// 		ctx.JSON(http.StatusBadRequest, errorResponse(err))
// 		return
// 	}

// 	user, err := server.store.GetUser(ctx, req.Username)
// 	if err != nil {
// 		if errors.Is(err, db.ErrRecordNotFound) {
// 			ctx.JSON(http.StatusNotFound, errorResponse(err))
// 			return
// 		}
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	err = util.CheckPassword(req.Password, user.HashedPassword)
// 	if err != nil {
// 		ctx.JSON(http.StatusUnauthorized, errorResponse(err))
// 		return
// 	}

// 	accessToken, accessPayload, err := server.tokenMaker.CreateToken(
// 		user.Username,
// 		user.Role,
// 		server.config.AccessTokenDuration,
// 	)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	refreshToken, refreshPayload, err := server.tokenMaker.CreateToken(
// 		user.Username,
// 		user.Role,
// 		server.config.RefreshTokenDuration,
// 	)
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	session, err := server.store.CreateSession(ctx, db.CreateSessionParams{
// 		ID:           refreshPayload.ID,
// 		Username:     user.Username,
// 		RefreshToken: refreshToken,
// 		UserAgent:    ctx.Request.UserAgent(),
// 		ClientIp:     ctx.ClientIP(),
// 		IsBlocked:    false,
// 		ExpiresAt:    refreshPayload.ExpiredAt,
// 	})
// 	if err != nil {
// 		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
// 		return
// 	}

// 	rsp := loginUserResponse{
// 		SessionID:             session.ID,
// 		AccessToken:           accessToken,
// 		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
// 		RefreshToken:          refreshToken,
// 		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
// 		User:                  newUserResponse(user),
// 	}
// 	ctx.JSON(http.StatusOK, rsp)
// }
