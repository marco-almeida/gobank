package service

import (
	"context"

	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// UserRepository defines the methods that any User repository should implement.
type UserRepository interface {
	Create(ctx context.Context, arg db.CreateUserParams) (db.User, error)
}

// User defines the application service in charge of interacting with Users.
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
