package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/marco-almeida/mybank/internal"
	"github.com/marco-almeida/mybank/internal/postgresql/db"
)

// AccountRepository defines the methods that any Account repository should implement.
type AccountRepository interface {
	Create(context context.Context, account db.CreateAccountParams) (db.Account, error)
	Get(context context.Context, id int64) (db.Account, error)
	List(ctx context.Context, arg db.ListAccountsParams) ([]db.Account, error)
	Delete(ctx context.Context, id int64) error
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
	acc, err := s.repo.Create(ctx, account)
	if err != nil {
		if errors.Is(err, internal.ErrUniqueConstraintViolation) {
			return db.Account{}, fmt.Errorf("%w: %s", internal.ErrAccountAlreadyExists, err.Error())
		}
		return db.Account{}, err
	}

	return acc, nil
}

func (s *AccountService) Get(ctx context.Context, id int64) (db.Account, error) {
	return s.repo.Get(ctx, id)
}

func (s *AccountService) List(ctx context.Context, arg db.ListAccountsParams) ([]db.Account, error) {
	return s.repo.List(ctx, arg)
}

func (s *AccountService) Delete(ctx context.Context, id int64) error {
	acc, err := s.repo.Get(ctx, id)
	if err != nil {
		return err
	}
	if acc.Balance != 0 {
		return internal.ErrBalanceNotZero
	}
	return s.repo.Delete(ctx, id)
}
