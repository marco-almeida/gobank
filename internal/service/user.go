package service

import (
	"github.com/lib/pq"
	"github.com/marco-almeida/gobank/internal/storage"
	"github.com/marco-almeida/gobank/internal/types"
	"github.com/sirupsen/logrus"
)

// UsersService ...
type UsersService interface {
	GetAll(limit, offset int64) ([]types.User, error)
	Get(id int64) (types.User, error)
	Create(user types.User) error
	Delete(id int64) error
	Update(id int64, user types.User) (types.User, error)
	PartialUpdate(id int64, user types.User) (types.User, error)
}

type usersService struct {
	store storage.UserStore
	log   *logrus.Logger
}

func NewUsers(store storage.UserStore, log *logrus.Logger) UsersService {
	return &usersService{
		store: store,
		log:   log,
	}
}

func (s *usersService) GetAll(limit, offset int64) ([]types.User, error) {
	return s.store.GetAll(limit, offset)
}

func (s *usersService) Get(id int64) (types.User, error) {
	return s.store.GetByID(id)
}

func (s *usersService) Create(user types.User) error {
	err := s.store.Create(&user)

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

func (s *usersService) Delete(id int64) error {
	return s.store.DeleteByID(id)
}

func (s *usersService) Update(id int64, user types.User) (types.User, error) {
	return user, s.store.UpdateByID(id, &user)
}

func (s *usersService) PartialUpdate(id int64, user types.User) (types.User, error) {
	return user, s.store.PartialUpdateByID(id, &user)
}
