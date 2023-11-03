package postgres

import (
	"context"
	"fmt"
	"prova/app"
	"prova/postgres/query"
	"strings"
)

var _ app.UserService = (*UserService)(nil)

type UserService struct {
	db *DB
}

// DeleteUser implements app.UserService.
func (s *UserService) DeleteUser(ctx context.Context, id int64) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteUser(ctx, tx, id); err != nil {
		return err
	} else if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// FindUserByID implements app.UserService.
func (s *UserService) FindUserByID(ctx context.Context, id int64) (*app.User, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return nil, err
	} else if err := attachUserAssociations(ctx, tx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// FindUsers implements app.UserService.
func (s *UserService) FindUsers(ctx context.Context, filter app.UserFilter) ([]*app.User, int, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	return findUsers(ctx, tx, filter)
}

// UpdateUser implements app.UserService.
func (s *UserService) UpdateUser(ctx context.Context, id int64, upd app.UserUpdate) (*app.User, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := updateUser(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	} else if err := attachUserAssociations(ctx, tx, user); err != nil {
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func NewUserService(db *DB) *UserService {
	return &UserService{db: db}
}

func (u *UserService) CreateUser(ctx context.Context, crt app.UserCreate) (*app.User, error) {

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	user, err := createUser(ctx, tx, crt)
	if err != nil {
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil
}

func createUser(ctx context.Context, tx *Tx, crt app.UserCreate) (*app.User, error) {

	bcryptedPassword, err := HashPassword(crt.Password)
	if err != nil {
		return nil, err
	}

	user := &app.User{
		Name:     crt.Name,
		Email:    crt.Email,
		Surname:  crt.Surname,
		Password: string(bcryptedPassword),
		Phone:    crt.Phone,
	}

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO users (name, surname, email, password, phone)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, user.Name, user.Surname, user.Email, user.Password, user.Phone).Scan(&user.ID); err != nil {
		return nil, app.Errorf(app.EINTERNAL, "Error creating user: %v", err)
	}

	return user, nil
}

// deleteAdmin elimina un amministratore.
func deleteUser(ctx context.Context, tx *Tx, id int64) error {

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM user
		WHERE id = $1
	`, user.ID); err != nil {
		return app.Errorf(app.EINTERNAL, "Error deleting user: %v", err)
	}

	return nil
}

// findAdminByID cerca un amministratore per ID.
func findUserByID(ctx context.Context, tx *Tx, id int64) (*app.User, error) {

	a, _, err := findUsers(ctx, tx, app.UserFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(a) == 0 {
		return nil, app.Errorf(app.ENOTFOUND, "User not found")
	}

	return a[0], nil
}

// findAdmins cerca gli amministratori, restituisce il numero totale di risultati al netto della paginazione.
func findUsers(ctx context.Context, tx *Tx, filter app.UserFilter) (_ []*app.User, n int, err error) {

	where, args := []string{"1 = 1"}, []any{}

	counterParameter := 1

	if v := filter.ID; v != nil {
		where, args = append(where, fmt.Sprintf("users.id = $%d", counterParameter)), append(args, *v)
		counterParameter += 1
	}

	if v := filter.Email; v != nil {
		where, args = append(where, fmt.Sprintf("users.email = $%d", counterParameter)), append(args, *v)
		counterParameter += 1
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			users.id,
			users.name,
			users.surname,
			users.email,
			users.password,
			users.phone,
			COUNT(*) OVER() AS total_count
		FROM users
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY users.id DESC
	`+query.FormatLimitPage(filter.Limit, filter.Page), args...)
	if err != nil {
		return nil, 0, app.Errorf(app.EINTERNAL, "Error querying user: %v", err)
	}
	defer rows.Close()

	users := []*app.User{}

	for rows.Next() {

		var user app.User

		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Surname,
			&user.Email,
			&user.Password,
			&user.Phone,
			&n,
		); err != nil {
			return nil, 0, app.Errorf(app.EINTERNAL, "Error scanning user: %v", err)
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, app.Errorf(app.EINTERNAL, "Error iterating user: %v", err)
	}

	return users, n, nil
}

// updateAdmin aggiorna un amministratore.
func updateUser(ctx context.Context, tx *Tx, id int64, upd app.UserUpdate) (*app.User, error) {

	user, err := findUserByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if v := upd.Name; v.Set {
		user.Name = v.Value
	}

	if v := upd.Surname; v.Set {
		user.Name = v.Value
	}

	if v := upd.Email; v.Set {

		a, count, err := findAdmins(ctx, tx, app.AdminFilter{Email: &v.Value})
		if err != nil {
			return nil, err
		} else if count > 0 && a[0].ID != user.ID {
			return nil, app.Errorf(app.ECONFLICT, "Email already in use")
		}

		user.Email = v.Value
	}
	if v := upd.Password; v.Set {
		bcryptedPassword, err := HashPassword(v.Value)
		if err != nil {
			return nil, err
		}
		user.Password = string(bcryptedPassword)
	}

	// if v := upd.Phone; v.Set {
	// 	user.Name = v.Value
	// }

	if err := user.Validate(); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE admins SET
			name = $2,
			surname = $6,
			email = $3,
			password = $4,
			phone = $5,
		WHERE id = $1
	`, user.ID, user.Name, user.Email, user.Password, user.Phone, user.Surname); err != nil {
		return nil, app.Errorf(app.EINTERNAL, "Error updating user: %v", err)
	}

	return user, nil
}

func attachUserAssociations(ctx context.Context, tx *Tx, user *app.User) (err error) {
	return nil
}
