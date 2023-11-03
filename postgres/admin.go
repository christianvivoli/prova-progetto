package postgres

import (
	"context"
	"fmt"
	"prova/app"
	"strings"

	"prova/postgres/query"
)

var _ app.AdminService = (*AdminService)(nil)

type AdminService struct {
	db *DB
}

func NewAdminService(db *DB) *AdminService {
	return &AdminService{db: db}
}

// CreateAdmin implements app.AdminService.
func (s *AdminService) CreateAdmin(ctx context.Context, crt app.AdminCreate) (*app.Admin, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	admin, err := createAdmin(ctx, tx, crt)
	if err != nil {
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, err
	}

	return admin, nil
}

// DeleteAdmin implements app.AdminService.
func (s *AdminService) DeleteAdmin(ctx context.Context, id int64) error {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteAdmin(ctx, tx, id); err != nil {
		return err
	} else if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

// FindAdminByID implements app.AdminService.
func (s *AdminService) FindAdminByID(ctx context.Context, id int64) (*app.Admin, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	admin, err := findAdminByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (s *AdminService) UpdateAdmin(ctx context.Context, id int64, upd app.AdminUpdate) (*app.Admin, error) {

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	admin, err := updateAdmin(ctx, tx, id, upd)
	if err != nil {
		return nil, err
	} else if err := tx.Commit(); err != nil {
		return nil, err
	}

	return admin, nil
}

func createAdmin(ctx context.Context, tx *Tx, crt app.AdminCreate) (*app.Admin, error) {

	bcryptedPassword, err := HashPassword(crt.Password)
	if err != nil {
		return nil, err
	}

	admin := &app.Admin{
		Name:     crt.Name,
		Email:    crt.Email,
		Surname:  crt.Surname,
		Password: string(bcryptedPassword),
		Active:   true,
	}

	if err := admin.Validate(); err != nil {
		return nil, err
	}

	if err := tx.QueryRowContext(ctx, `
		INSERT INTO admin (name, surname, email, password, active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, admin.Name, admin.Email, admin.Password, admin.Active, admin.Surname).Scan(&admin.ID); err != nil {
		return nil, app.Errorf(app.EINTERNAL, "Error creating admin: %v", err)
	}

	return admin, nil
}

// deleteAdmin elimina un amministratore.
func deleteAdmin(ctx context.Context, tx *Tx, id int64) error {

	admin, err := findAdminByID(ctx, tx, id)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
		DELETE FROM admin
		WHERE id = $1
	`, admin.ID); err != nil {
		return app.Errorf(app.EINTERNAL, "Error deleting admin: %v", err)
	}

	return nil
}

// findAdminByID cerca un amministratore per ID.
func findAdminByID(ctx context.Context, tx *Tx, id int64) (*app.Admin, error) {

	a, _, err := findAdmins(ctx, tx, app.AdminFilter{ID: &id})
	if err != nil {
		return nil, err
	} else if len(a) == 0 {
		return nil, app.Errorf(app.ENOTFOUND, "Admin not found")
	}

	return a[0], nil
}

// findAdmins cerca gli amministratori, restituisce il numero totale di risultati al netto della paginazione.
func findAdmins(ctx context.Context, tx *Tx, filter app.AdminFilter) (_ []*app.Admin, n int, err error) {

	where, args := []string{"1 = 1"}, []any{}

	counterParameter := 1

	if v := filter.ID; v != nil {
		where, args = append(where, fmt.Sprintf("admin.id = $%d", counterParameter)), append(args, *v)
		counterParameter += 1
	}

	if v := filter.Email; v != nil {
		where, args = append(where, fmt.Sprintf("admin.email = $%d", counterParameter)), append(args, *v)
		counterParameter += 1
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT
			admin.id,
			admin.name,
			admin.surname,
			admin.email,
			admin.password,
			admin.active,
			COUNT(*) OVER() AS total_count
		FROM admin
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY admin.id DESC
	`+query.FormatLimitPage(filter.Limit, filter.Page), args...)
	if err != nil {
		return nil, 0, app.Errorf(app.EINTERNAL, "Error querying admin: %v", err)
	}
	defer rows.Close()

	admins := []*app.Admin{}

	for rows.Next() {

		var admin app.Admin

		if err := rows.Scan(
			&admin.ID,
			&admin.Name,
			&admin.Surname,
			&admin.Email,
			&admin.Password,
			&admin.Active,
		); err != nil {
			return nil, 0, app.Errorf(app.EINTERNAL, "Error scanning admins: %v", err)
		}

		admins = append(admins, &admin)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, app.Errorf(app.EINTERNAL, "Error iterating admins: %v", err)
	}

	return admins, n, nil
}

// updateAdmin aggiorna un amministratore.
func updateAdmin(ctx context.Context, tx *Tx, id int64, upd app.AdminUpdate) (*app.Admin, error) {

	admin, err := findAdminByID(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if v := upd.Name; v.Set {
		admin.Name = v.Value
	}

	if v := upd.Surname; v.Set {
		admin.Surname = v.Value
	}

	if v := upd.Email; v.Set {

		a, count, err := findAdmins(ctx, tx, app.AdminFilter{Email: &v.Value})
		if err != nil {
			return nil, err
		} else if count > 0 && a[0].ID != admin.ID {
			return nil, app.Errorf(app.ECONFLICT, "Email already in use")
		}

		admin.Email = v.Value
	}
	if v := upd.Password; v.Set {
		bcryptedPassword, err := HashPassword(v.Value)
		if err != nil {
			return nil, err
		}
		admin.Password = string(bcryptedPassword)
	}

	if err := admin.Validate(); err != nil {
		return nil, err
	}

	if _, err := tx.ExecContext(ctx, `
		UPDATE admin SET
			name = $2,
			surname = $3,
			email = $4,
			password = $5,
		WHERE id = $1
	`, admin.ID, admin.Name, admin.Surname, admin.Email, admin.Password); err != nil {
		return nil, app.Errorf(app.EINTERNAL, "Error updating admin: %v", err)
	}

	return admin, nil
}
