package model

import (
	"time"
)

type User struct {
	ID        int64     `json:"id"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"-"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
