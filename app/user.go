package app

import (
	"context"
	"net/mail"
	common "prova/common"
)

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    int64  `json:"phone"`
}

func (u User) Validate() error {

	if u.Name == "" {
		return Errorf(EINVALID, "Name is required")
	}

	if u.Surname == "" {
		return Errorf(EINVALID, "Surname is required")
	}

	if _, err := mail.ParseAddress(u.Email); err != nil {
		return Errorf(EINVALID, "Email is required")
	}

	if u.Password == "" {
		return Errorf(EINVALID, "Password is required")
	}

	if u.Phone == 0 {
		return Errorf(EINVALID, "Phone is required")
	}

	return nil
}

type UserService interface {
	// CreateUser crea un nuovo user
	CreateUser(ctx context.Context, crt UserCreate) (*User, error)
	// DeleteUser elima un user
	DeleteUser(ctx context.Context, id int64) error
	// FindUserByID cerca un User tramite ID
	FindUserByID(ctx context.Context, id int64) (*User, error)
	// UpdateUser aggiorna un User
	UpdateUser(ctx context.Context, id int64, upd UserUpdate) (*User, error)

	FindUsers(ctx context.Context, filter UserFilter) ([]*User, int, error)
}

type UserCreate struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    int64  `json:"phone"`
}

type UserUpdate struct {
	Name     common.Patch[string] `json:"name"`
	Surname  common.Patch[string] `json:"surname"`
	Email    common.Patch[string] `json:"email"`
	Password common.Patch[string] `json:"password"`
}

type UserFilter struct {
	ID    *int64  `json:"id"`
	Email *string `json:"email"`

	Page  int `json:"page"`
	Limit int `json:"limit"`
}