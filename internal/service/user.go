package service

import (
	"context"

	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// UserRepository defines the methods that any User repository should implement.
type UserRepository interface {
	Create(ctx context.Context, arg db.CreateUserParams) (db.User, error)
	Get(ctx context.Context, username string) (db.User, error)
}

type AuthService interface {
	Create(ctx context.Context, user CreateUserParams) (db.User, error)
	Login(ctx context.Context, req LoginUserParams) (LoginUserResponse, error)
	RenewAccessToken(ctx context.Context, req RenewAccessTokenParams) (RenewAccessTokenResponse, error)
}

// UserService defines the application service in charge of interacting with Users.
type UserService struct {
	repo    UserRepository
	authSvc AuthService
}

// NewUserService creates a new User service.
func NewUserService(repo UserRepository, authSvc AuthService) *UserService {
	return &UserService{
		repo:    repo,
		authSvc: authSvc,
	}
}

func (s *UserService) Create(ctx context.Context, user CreateUserParams) (db.User, error) {
	return s.authSvc.Create(ctx, user)
}

func (s *UserService) Get(ctx context.Context, username string) (db.User, error) {
	return s.repo.Get(ctx, username)
}

func (s *UserService) Login(ctx context.Context, req LoginUserParams) (LoginUserResponse, error) {
	return s.authSvc.Login(ctx, req)
}

func (s *UserService) RenewAccessToken(ctx context.Context, req RenewAccessTokenParams) (RenewAccessTokenResponse, error) {
	return s.authSvc.RenewAccessToken(ctx, req)
}
