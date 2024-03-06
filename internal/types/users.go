package types

import (
	"errors"
	"reflect"
	"time"
)

type RegisterUserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

func ValidateRegisterUserRequest(r *RegisterUserRequest) error {
	// iterate over struct fields
	val := reflect.ValueOf(r).Elem()
	for i := 0; i < val.NumField(); i++ {
		// if attribute value is empty, return error
		if val.Field(i).Interface() == "" {
			return errors.New(val.Type().Field(i).Tag.Get("json") + " is required")
		}
	}
	return nil
}

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
