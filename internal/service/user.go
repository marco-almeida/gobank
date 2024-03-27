package service

import (
	"github.com/lib/pq"
	"github.com/marco-almeida/gobank/internal/model"
	"github.com/sirupsen/logrus"
)

// UserRepository defines the methods that any User repository should implement.
type UserRepository interface {
	Create(u *model.User) error
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

func (s *User) Create(user model.User) error {
	err := s.repo.Create(&user)

	if err != nil {
		// h.svc.log.Infof("Error creating user: %v", err)
		// check if error of type duplicate key
		pgErr, ok := err.(*pq.Error)
		if !ok {
			// WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Error creating user"})
			return err
		}
		if ok && pgErr.Code == "23505" {
			// WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: "email address is already in use"})
			return err
		}

		// WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: "Error creating user"})
		return err

	}
	return err
}

func (s *User) Delete(id int64) error {
	return s.repo.DeleteByID(id)
}

func (s *User) Update(id int64, user model.User) (model.User, error) {
	return user, s.repo.UpdateByID(id, &user)
}

func (s *User) PartialUpdate(id int64, user model.User) (model.User, error) {
	return user, s.repo.PartialUpdateByID(id, &user)
}
