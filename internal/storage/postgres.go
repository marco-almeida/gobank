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
