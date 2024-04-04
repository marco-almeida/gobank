package service

import (
	"github.com/marco-almeida/gobank/internal"
	"github.com/sirupsen/logrus"
)

// AccountRepository defines the methods that any Account repository should implement.
type AccountRepository interface {
	// Create hashes the Account password before calling the repository method.
	Create(userID int64) error
	GetAllByUserID(userID, offset, limit int64) ([]internal.Account, error)
	GetByID(userID int64, accountID int64) (internal.Account, error)
	DeleteByID(userID, accountID int64) error
	UpdateBalanceByID(userID int64, accountID int64, balance internal.USD) (internal.Account, error)
}

// Account defines the application service in charge of interacting with Accounts.
type Account struct {
	repo AccountRepository
	log  *logrus.Logger
}

// NewAccount creates a new Account service.
func NewAccount(repo AccountRepository, log *logrus.Logger) *Account {
	return &Account{
		repo: repo,
		log:  log,
	}
}

func (s *Account) Create(userID int64) error {
	return s.repo.Create(userID)
}

func (s *Account) GetAllByUserID(userID, offset, limit int64) ([]internal.Account, error) {
	return s.repo.GetAllByUserID(userID, offset, limit)
}

func (s *Account) GetByID(userID int64, accountID int64) (internal.Account, error) {
	return s.repo.GetByID(userID, accountID)
}

func (s *Account) DeleteByID(userID, accountID int64) error {
	accountToDelete := internal.Account{}
	accountToDelete, err := s.GetByID(userID, accountID)
	if err != nil {
		return err
	}
	if accountToDelete.Balance != 0 {
		return internal.NewErrorf(internal.ErrorCodeInvalidArgument, "account has a balance of %s", accountToDelete.Balance)
	}
	return s.repo.DeleteByID(userID, accountID)
}

func (s *Account) UpdateBalanceByID(userID int64, accountID int64, balance internal.USD) (internal.Account, error) {
	return s.repo.UpdateBalanceByID(userID, accountID, balance)
}
