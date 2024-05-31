package service

import (
	"context"

	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// UserRepository defines the methods that any User repository should implement.
type UserRepository interface {
	Create(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	Get(ctx context.Context, username string) (db.User, error)
	CreateWithTx(ctx context.Context, arg db.CreateUserTxParams) (db.CreateUserTxResult, error)
}

// AuthService defines the application service in charge of interacting with Users.
type AuthService interface {
	Create(ctx context.Context, arg db.CreateUserTxParams) (db.CreateUserTxResult, error)
	Login(ctx context.Context, req LoginUserParams) (LoginUserResponse, error)
	RenewAccessToken(ctx context.Context, req RenewAccessTokenParams) (RenewAccessTokenResponse, error)
	// CreateWithTx(ctx context.Context, arg db.CreateUserTxParams) (db.CreateUserTxResult, error)
}

type EmailService interface {
	SendWelcomeEmail(username string) error
}

// UserService defines the application service in charge of interacting with Users.
type UserService struct {
	repo     UserRepository
	authSvc  AuthService
	emailSvc EmailService
}

// NewUserService creates a new User service.
func NewUserService(repo UserRepository, authSvc AuthService, emailSvc EmailService) *UserService {
	return &UserService{
		repo:     repo,
		authSvc:  authSvc,
		emailSvc: emailSvc,
	}
}

func (s *UserService) Create(ctx context.Context, req CreateUserParams) (db.User, error) {

	txParams := db.CreateUserTxParams{
		CreateUserParams: db.CreateUserParams{
			Username:       req.Username,
			HashedPassword: req.PlaintextPassword,
			FullName:       req.FullName,
			Email:          req.Email,
		},
		AfterCreate: func(user db.User) error {
			return s.emailSvc.SendWelcomeEmail(user.Username) // publishes task to queue
		},
	}
	txResult, err := s.authSvc.Create(ctx, txParams)

	if err != nil {
		return db.User{}, err
	}

	return txResult.User, nil
}

// func (s *UserService) Create(ctx context.Context, user CreateUserParams) (db.User, error) {
// 	return s.authSvc.Create(ctx, user)
// }

func (s *UserService) Get(ctx context.Context, username string) (db.User, error) {
	return s.repo.Get(ctx, username)
}

func (s *UserService) Login(ctx context.Context, req LoginUserParams) (LoginUserResponse, error) {
	return s.authSvc.Login(ctx, req)
}

func (s *UserService) RenewAccessToken(ctx context.Context, req RenewAccessTokenParams) (RenewAccessTokenResponse, error) {
	return s.authSvc.RenewAccessToken(ctx, req)
}
