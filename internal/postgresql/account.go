package postgres

import (
	"database/sql"
	"errors"

	"github.com/marco-almeida/gobank/internal"
)

// Account represents the repository used for interacting with Account records.
type Account struct {
	db *sql.DB
}

// NewAccount instantiates the Account repository.
func NewAccount(db *sql.DB) *Account {
	return &Account{
		db: db,
	}
}

func (s *Account) Init() error {
	return s.createTable()
}

func (s *Account) createTable() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS accounts (
		id BIGSERIAL PRIMARY KEY NOT NULL,
		user_id BIGINT NOT NULL,
		balance BIGINT NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)

	return err
}

func (s *Account) Create(userID int64) error {
	_, err := s.db.Exec(`INSERT INTO accounts (user_id) VALUES ($1)`, userID)
	return err
}

func (s *Account) GetAllByUserID(userID, offset, limit int64) ([]internal.Account, error) {
	rows, err := s.db.Query(`SELECT id, user_id, balance, created_at FROM accounts WHERE user_id = $1 OFFSET $2 LIMIT $3`, userID, offset, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []internal.Account
	for rows.Next() {
		var a internal.Account
		err := rows.Scan(&a.ID, &a.UserID, &a.Balance, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (s *Account) GetByID(userID int64, accountID int64) (internal.Account, error) {
	var a internal.Account
	err := s.db.QueryRow(`SELECT id, user_id, balance, created_at FROM accounts WHERE user_id = $1 AND id = $2`, userID, accountID).Scan(&a.ID, &a.UserID, &a.Balance, &a.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return internal.Account{}, internal.WrapErrorf(err, internal.ErrorCodeNotFound, "account not found")
		}
		return internal.Account{}, err
	}
	return a, nil
}

func (s *Account) DeleteByID(userID, accountID int64) error {
	// first check if account exists and its balance is 0
	var balance internal.USD
	err := s.db.QueryRow(`SELECT balance FROM accounts WHERE user_id = $1 AND id = $2`, userID, accountID).Scan(&balance)
	if err != nil {
		return err
	}
	if balance != 0 {
		return internal.ErrZeroBalance
	}

	_, err = s.db.Exec(`DELETE FROM accounts WHERE user_id = $1 AND id = $2`, userID, accountID)
	return err
}

func (s *Account) UpdateBalanceByID(userID int64, accountID int64, balance internal.USD) (internal.Account, error) {
	var updatedAccount internal.Account
	err := s.db.QueryRow(`UPDATE accounts SET balance = balance + $1 WHERE user_id = $2 AND id = $3 RETURNING id, balance`, balance, userID, accountID).Scan(&updatedAccount.ID, &updatedAccount.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return updatedAccount, internal.WrapErrorf(err, internal.ErrorCodeNotFound, "account not found")
		}
		return updatedAccount, err
	}
	return updatedAccount, nil
}
