package storage

import (
	"database/sql"

	_ "github.com/lib/pq"
	t "github.com/marco-almeida/gobank/internal/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type PostgresStorage struct {
	log *logrus.Logger
	db  *sql.DB
}

func NewPostgresStorage(connStr string, log *logrus.Logger) *PostgresStorage {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	log.Info("Connected to Postgres")

	return &PostgresStorage{log: log, db: db}
}

func (s *PostgresStorage) Init() error {
	err := s.createUsersTable()
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) createUsersTable() error {
	_, err := s.db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id BIGSERIAL PRIMARY KEY NOT NULL,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	)`)

	if err != nil {
		return err
	}

	_, err = s.db.Exec(`CREATE TABLE IF NOT EXISTS accounts (
		id BIGSERIAL PRIMARY KEY NOT NULL,
		user_id BIGINT NOT NULL,
		balance BIGINT NOT NULL DEFAULT 0,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	)`)

	return err
}

func (s *PostgresStorage) CreateUser(u *t.RegisterUserRequest) error {
	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Inserting the user into the database
	_, err = s.db.Exec(`INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4)`, u.FirstName, u.LastName, u.Email, string(hashedPassword))

	return err
}

func (s *PostgresStorage) GetAllUsers() ([]t.User, error) {
	rows, err := s.db.Query(`SELECT id, first_name, last_name, email, created_at FROM users`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []t.User{}
	for rows.Next() {
		var u t.User
		err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *PostgresStorage) DeleteUserByID(id int64) error {
	_, err := s.db.Exec(`DELETE FROM users WHERE id = $1`, id)
	return err
}

func (s *PostgresStorage) GetUserByEmail(email string) (t.User, error) {
	var u t.User
	err := s.db.QueryRow(`SELECT id, email, password FROM users WHERE email = $1`, email).Scan(&u.ID, &u.Email, &u.Password)
	if err != nil {
		return t.User{}, err
	}
	return u, nil
}

func (s *PostgresStorage) UpdateUserByID(id int64, u *t.RegisterUserRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`UPDATE users SET first_name = $1, last_name = $2, email = $3, password = $4 WHERE id = $5`, u.FirstName, u.LastName, u.Email, string(hashedPassword), id)
	return err
}

func (s *PostgresStorage) PartialUpdateUserByID(id int64, u *t.RegisterUserRequest) error {
	// only update fields that are not empty
	if u.FirstName != "" {
		_, err := s.db.Exec(`UPDATE users SET first_name = $1 WHERE id = $2`, u.FirstName, id)
		if err != nil {
			return err
		}
	}

	if u.LastName != "" {
		_, err := s.db.Exec(`UPDATE users SET last_name = $1 WHERE id = $2`, u.LastName, id)
		if err != nil {
			return err
		}
	}

	if u.Email != "" {
		_, err := s.db.Exec(`UPDATE users SET email = $1 WHERE id = $2`, u.Email, id)
		if err != nil {
			return err
		}
	}

	if u.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = s.db.Exec(`UPDATE users SET password = $1 WHERE id = $2`, string(hashedPassword), id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *PostgresStorage) GetUserByID(id int64) (t.User, error) {
	var u t.User
	err := s.db.QueryRow(`SELECT id, first_name, last_name, email, created_at FROM users WHERE id = $1`, id).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.CreatedAt)
	if err != nil {
		return t.User{}, err
	}
	return u, nil
}

// accounts

func (s *PostgresStorage) CreateAccount(userID int64) error {
	_, err := s.db.Exec(`INSERT INTO accounts (user_id) VALUES ($1)`, userID)
	return err
}

func (s *PostgresStorage) GetAllAccountsByUserID(userID int64) ([]t.Account, error) {
	rows, err := s.db.Query(`SELECT id, user_id, balance, created_at FROM accounts WHERE user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := []t.Account{}
	for rows.Next() {
		var a t.Account
		err := rows.Scan(&a.ID, &a.UserID, &a.Balance, &a.CreatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func (s *PostgresStorage) GetAccountByID(userID int64, accountID int64) (t.Account, error) {
	var a t.Account
	err := s.db.QueryRow(`SELECT id, user_id, balance, created_at FROM accounts WHERE user_id = $1 AND id = $2`, userID, accountID).Scan(&a.ID, &a.UserID, &a.Balance, &a.CreatedAt)
	if err != nil {
		return t.Account{}, err
	}
	return a, nil
}

func (s *PostgresStorage) DeleteAccountByID(userID int64, accountID int64) error {
	// first check if account exists and its balance is 0
	var balance t.USD
	err := s.db.QueryRow(`SELECT balance FROM accounts WHERE user_id = $1 AND id = $2`, userID, accountID).Scan(&balance)
	if err != nil {
		return err
	}
	if balance != 0 {
		return t.ErrZeroBalance
	}

	_, err = s.db.Exec(`DELETE FROM accounts WHERE user_id = $1 AND id = $2`, userID, accountID)
	return err
}

func (s *PostgresStorage) UpdateAccountBalanceByID(userID int64, accountID int64, balance t.USD) (t.USD, error) {
	var newBalance t.USD
	err := s.db.QueryRow(`UPDATE accounts SET balance = balance + $1 WHERE user_id = $2 AND id = $3 RETURNING balance`, balance, userID, accountID).Scan(&newBalance)
	if err != nil {
		return 0, err
	}
	return newBalance, nil
}
