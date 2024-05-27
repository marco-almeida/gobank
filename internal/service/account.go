package service

import (
	"context"

	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// AccountRepository defines the methods that any Account repository should implement.
type AccountRepository interface {
	Create(context context.Context, account db.CreateAccountParams) (db.Account, error)
	Get(context context.Context, id int64) (db.Account, error)
	List(ctx context.Context, arg db.ListAccountsParams) ([]db.Account, error)
}

// AccountService defines the application service in charge of interacting with Accounts.
type AccountService struct {
	repo AccountRepository
}

// NewAccountService creates a new Account service.
func NewAccountService(repo AccountRepository) *AccountService {
	return &AccountService{
		repo: repo,
	}
}

func (s *AccountService) Create(ctx context.Context, account db.CreateAccountParams) (db.Account, error) {
	return s.repo.Create(ctx, account)
}

func (s *AccountService) Get(ctx context.Context, id int64) (db.Account, error) {
	return s.repo.Get(ctx, id)
}

func (s *AccountService) List(ctx context.Context, arg db.ListAccountsParams) ([]db.Account, error) {
	return s.repo.List(ctx, arg)
}
