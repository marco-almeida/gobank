package storage

import (
	t "github.com/marco-almeida/gobank/internal/types"
)

type Storer interface {
	// users

	CreateUser(u *t.RegisterUserRequest) error
	GetAllUsers() ([]t.User, error)
	DeleteUserByID(int64) error
	GetUserByEmail(string) (t.User, error)
	UpdateUserByID(int64, *t.RegisterUserRequest) error
	PartialUpdateUserByID(int64, *t.RegisterUserRequest) error
	GetUserByID(int64) (t.User, error)

	// bank accounts

	CreateAccount(userID int64) error
	GetAllAccountsByUserID(userID int64) ([]t.Account, error)
	GetAccountByID(userID int64, accountID int64) (t.Account, error)
	// DeleteAccountByID(accountID int64) error
	// UpdateAccountBalanceByID(accountID int64, balance int64) error
}
