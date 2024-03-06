package storage

import (
	"database/sql"

	_ "github.com/lib/pq"
	t "github.com/marco-almeida/golang-api-project-layout/internal/types"
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

func (s *PostgresStorage) CreateUser(u *t.RegisterUserRequest) (int64, error) {
	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	// Inserting the user into the database
	lastInsertId := int64(0)
	err = s.db.QueryRow(`INSERT INTO users (first_name, last_name, email, password) VALUES ($1, $2, $3, $4) RETURNING id`, u.FirstName, u.LastName, u.Email, string(hashedPassword)).Scan(&lastInsertId)

	if err != nil {
		return 0, err
	}

	return lastInsertId, nil
	// // Comparing the password with the hash
	// err = bcrypt.CompareHashAndPassword(hashedPassword, passwordBytes)
	// fmt.Println(err) // nil means it is a match
}
