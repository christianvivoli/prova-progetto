package app

import (
	"context"
	"net/mail"
	common "prova/common"
)

type Admin struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Active   bool   `json:"admin"`
}

func (a Admin) Validate() error {

	if a.Name == "" {
		return Errorf(EINVALID, "Name is required")
	}

	if a.Surname == "" {
		return Errorf(EINVALID, "Surname is required")
	}

	if _, err := mail.ParseAddress(a.Email); err != nil {
		return Errorf(EINVALID, "Email is invalid")
	}

	if a.Password == "" {
		return Errorf(EINVALID, "Password is required")
	}

	return nil
}

type AdminService interface {
	// CreateAdmin crea un nuovo amministratore.
	CreateAdmin(ctx context.Context, crt AdminCreate) (*Admin, error)
	// DeleteAdmin elimina un amministratore.
	DeleteAdmin(ctx context.Context, id int64) error
	// FindAdminByID cerca un amministratore per ID.
	FindAdminByID(ctx context.Context, id int64) (*Admin, error)
	// UpdateAdmin aggiorna un amministratore.
	UpdateAdmin(ctx context.Context, id int64, upd AdminUpdate) (*Admin, error)
}

type AdminCreate struct {
	Name     string `json:"name"`
	Surname  string `json:"surname"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AdminUpdate struct {
	Name     common.Patch[string] `json:"name"`
	Surname  common.Patch[string] `json:"surname"`
	Email    common.Patch[string] `json:"email"`
	Password common.Patch[string] `json:"password"`
}

type AdminFilter struct {
	ID    *int64  `json:"id"`
	Email *string `json:"email"`

	Page  int `json:"page"`
	Limit int `json:"limit"`
}
