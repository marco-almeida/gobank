package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/marco-almeida/mybank/internal"
	"github.com/marco-almeida/mybank/internal/pkg"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
	"github.com/marco-almeida/mybank/internal/token"
)

// use a single instance of Validate, it caches struct info
var validate = validator.New(validator.WithRequiredStructEnabled())

// SessionRepository defines the methods that any Session repository should implement.
type SessionRepository interface {
	Create(ctx context.Context, arg db.CreateSessionParams) (db.Session, error)
	Get(ctx context.Context, id uuid.UUID) (db.Session, error)
}

// AuthServiceImpl defines the application service in charge of interacting with Auth.
type AuthServiceImpl struct {
	sessionRepo          SessionRepository
	userRepo             UserRepository
	verifyEmailRepo      VerifyEmailRepository
	tokenMaker           token.Maker
	accessTokenDuration  time.Duration
	refreshTokenDuration time.Duration
}

// NewAuthService creates a new Auth service.
func NewAuthService(userRepo UserRepository, sessionRepo SessionRepository, tokenMaker token.Maker, accessTokenDuration time.Duration, refreshTokenDuration time.Duration, verifyEmailRepo VerifyEmailRepository) *AuthServiceImpl {
	return &AuthServiceImpl{
		userRepo:             userRepo,
		sessionRepo:          sessionRepo,
		tokenMaker:           tokenMaker,
		accessTokenDuration:  accessTokenDuration,
		refreshTokenDuration: refreshTokenDuration,
		verifyEmailRepo:      verifyEmailRepo,
	}
}

type CreateUserParams struct {
	Username          string `json:"username" validate:"required,alphanum"`
	PlaintextPassword string `json:"plaintext_password" validate:"required,min=6"`
	FullName          string `json:"full_name" validate:"required"`
	Email             string `json:"email" validate:"required,email"`
}

type CreateUserTxParams struct {
	Username          string
	PlaintextPassword string
	FullName          string
	Email             string
	AfterCreate       func(user db.User) error
}

func (s *AuthServiceImpl) Create(ctx context.Context, req CreateUserTxParams) (db.CreateUserTxResult, error) {
	// validate CreateUserParams
	err := validate.Struct(req)
	if err != nil {
		return db.CreateUserTxResult{}, err
	}

	// hash password, call auth for this
	hashedPassword, err := pkg.HashPassword(req.PlaintextPassword)
	if err != nil {
		return db.CreateUserTxResult{}, err
	}

	return s.userRepo.CreateWithTx(ctx, db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.Username,
			HashedPassword: hashedPassword,
			FullName:       req.FullName,
			Email:          req.Email,
		},
		AfterCreate: req.AfterCreate,
	})

}

type LoginUserParams struct {
	Username  string `json:"username" validate:"required,alphanum"`
	Password  string `json:"password" validate:"required,min=6"`
	UserAgent string `json:"user_agent" validate:"required"`
	ClientIP  string `json:"client_ip" validate:"required"`
}

type LoginUserResponse struct {
	SessionID             uuid.UUID    `json:"session_id"`
	AccessToken           string       `json:"access_token"`
	AccessTokenExpiresAt  time.Time    `json:"access_token_expires_at"`
	RefreshToken          string       `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time    `json:"refresh_token_expires_at"`
	User                  userResponse `json:"user"`
}

type userResponse struct {
	Username          string    `json:"username"`
	FullName          string    `json:"full_name"`
	Email             string    `json:"email"`
	PasswordChangedAt time.Time `json:"password_changed_at"`
	CreatedAt         time.Time `json:"created_at"`
}

func (s *AuthServiceImpl) Login(ctx context.Context, req LoginUserParams) (LoginUserResponse, error) {
	user, err := s.userRepo.Get(ctx, req.Username)
	if err != nil {
		if errors.Is(err, internal.ErrNoRows) {
			return LoginUserResponse{}, fmt.Errorf("%w; user not found: %w", internal.ErrInvalidCredentials, err)
		}
		return LoginUserResponse{}, err
	}

	if !user.IsEmailVerified {
		return LoginUserResponse{}, internal.ErrUnverifiedAccount
	}

	err = pkg.CheckPassword(req.Password, user.HashedPassword)
	if err != nil {
		return LoginUserResponse{}, fmt.Errorf("%w; %w", internal.ErrInvalidCredentials, err)
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(
		user.Username,
		user.Role,
		s.accessTokenDuration,
	)
	if err != nil {
		return LoginUserResponse{}, fmt.Errorf("cannot create access token: %w", err)
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(
		user.Username,
		user.Role,
		s.refreshTokenDuration,
	)
	if err != nil {
		return LoginUserResponse{}, fmt.Errorf("cannot create refresh token: %w", err)
	}

	session, err := s.sessionRepo.Create(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    req.UserAgent,
		ClientIp:     req.ClientIP,
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return LoginUserResponse{}, fmt.Errorf("cannot create session: %w", err)
	}

	return LoginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User: userResponse{
			Username:          user.Username,
			FullName:          user.FullName,
			Email:             user.Email,
			PasswordChangedAt: user.PasswordChangedAt,
			CreatedAt:         user.CreatedAt,
		},
	}, nil
}

type RenewAccessTokenParams struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RenewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (s *AuthServiceImpl) RenewAccessToken(ctx context.Context, req RenewAccessTokenParams) (RenewAccessTokenResponse, error) {
	refreshPayload, err := s.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		return RenewAccessTokenResponse{}, fmt.Errorf("%w; %w", internal.ErrInvalidToken, err)
	}

	session, err := s.sessionRepo.Get(ctx, refreshPayload.ID)
	if err != nil {
		if errors.Is(err, internal.ErrNoRows) {
			return RenewAccessTokenResponse{}, fmt.Errorf("session not found: %w", err)
		}
		return RenewAccessTokenResponse{}, fmt.Errorf("internal server error: %w", err)
	}

	if session.IsBlocked {
		err := fmt.Errorf("blocked session")
		return RenewAccessTokenResponse{}, fmt.Errorf("%w; session is blocked: %w", internal.ErrInvalidToken, err)
	}

	if session.Username != refreshPayload.Username {
		err := fmt.Errorf("incorrect session user")
		return RenewAccessTokenResponse{}, fmt.Errorf("%w; session user mismatch: %w", internal.ErrInvalidToken, err)
	}

	if session.RefreshToken != req.RefreshToken {
		err := fmt.Errorf("mismatched session token")
		return RenewAccessTokenResponse{}, fmt.Errorf("%w; session token mismatch: %w", internal.ErrInvalidToken, err)
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("expired session")
		return RenewAccessTokenResponse{}, fmt.Errorf("%w; session expired: %w", internal.ErrInvalidToken, err)
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(
		refreshPayload.Username,
		refreshPayload.Role,
		s.accessTokenDuration,
	)
	if err != nil {
		return RenewAccessTokenResponse{}, fmt.Errorf("cannot create access token: %w", err)
	}

	return RenewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}, nil
}

func (s *AuthServiceImpl) VerifyEmail(ctx context.Context, req db.VerifyEmailTxParams) (db.VerifyEmailTxResult, error) {
	return s.verifyEmailRepo.Verify(ctx, req)
}
