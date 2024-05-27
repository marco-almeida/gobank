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

// UserService defines the application service in charge of interacting with Users.
type UserService struct {
	repo UserRepository
}

// NewUserService creates a new User service.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) Create(ctx context.Context, user db.CreateUserParams) (db.User, error) {
	return s.repo.Create(ctx, user)
}

func (s *UserService) Get(ctx context.Context, username string) (db.User, error) {
	return s.repo.Get(ctx, username)
}