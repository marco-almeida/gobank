package service

import (
	"github.com/marco-almeida/gobank/internal"
	"github.com/marco-almeida/gobank/internal/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository defines the methods that any User repository should implement.
type UserRepository interface {
	// Create hashes the user password before calling the repository method.
	Create(*model.User) error
	GetAll(limit, offset int64) ([]model.User, error)
	DeleteByID(int64) error
	GetByEmail(string) (model.User, error)
	UpdateByID(int64, *model.User) error
	PartialUpdateByID(int64, *model.User) error
	GetByID(int64) (model.User, error)
}

// User defines the application service in charge of interacting with Users.
type User struct {
	repo UserRepository
	log  *logrus.Logger
}

// NewUsers creates a new User service.
func NewUser(repo UserRepository, log *logrus.Logger) *User {
	return &User{
		repo: repo,
		log:  log,
	}
}

func (s *User) GetAll(limit, offset int64) ([]model.User, error) {
	return s.repo.GetAll(limit, offset)
}

func (s *User) Get(id int64) (model.User, error) {
	return s.repo.GetByID(id)
}

func (s *User) Create(u model.User) error {
	// Hashing the password with the default cost of 10
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return internal.WrapErrorf(err, internal.ErrorCodeUnknown, "failed to hash password")
	}
	u.Password = string(hashedPassword)
	return s.repo.Create(&u)
}

func (s *User) Delete(id int64) error {
	return s.repo.DeleteByID(id)
}

func (s *User) Update(id int64, u model.User) (model.User, error) {
	return u, s.repo.UpdateByID(id, &u)
}

func (s *User) PartialUpdate(id int64, u model.User) (model.User, error) {
	return u, s.repo.PartialUpdateByID(id, &u)
}
