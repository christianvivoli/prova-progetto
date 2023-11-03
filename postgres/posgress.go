package postgres

import (
	"context"
	"database/sql"
	"embed"
	"io/fs"
	"sort"
	"time"

	"prova/app"

	"golang.org/x/crypto/bcrypt"

	log "github.com/inconshreveable/log15"
	_ "github.com/lib/pq"
)

//go:embed migration/*.sql
var migrationsFS embed.FS

// Logger is the default logger for the package, should be used only for debug purpose or special cases.
var Logger log.Logger

// HashPassword is defined as var to make easy during tests decrease the bcrypt cost.
var HashPassword = func(password string) ([]byte, error) {
	p, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, app.Errorf(app.EINTERNAL, "Error hashing password: %v", err)
	}
	return p, nil
}

// Tx wraps the SQL Tx object to provide a timestamp at the start of the transaction.
type Tx struct {
	*sql.Tx
	db  *DB
	now time.Time

	_commit   func() error
	_rollback func() error
}

func (t *Tx) Commit() error {
	if err := t._commit(); err != nil {
		return app.Errorf(app.EINTERNAL, "Error committing transaction: %v", err)
	}
	return nil
}

func (t *Tx) Rollback() error {
	if err := t._rollback(); err != nil {
		return app.Errorf(app.EINTERNAL, "Error rolling back transaction: %v", err)
	}
	return nil
}

func (t *Tx) Now() time.Time {
	return t.now
}

// DB represents the database connection.
type DB struct {
	conn   *sql.DB
	ctx    context.Context
	cancel func()

	DSN string

	Now func() time.Time
}

// NewDB returns a new instance of DB with the given DSN.
func NewDB(dsn string) *DB {
	db := &DB{
		DSN: dsn,
		Now: time.Now,
	}
	db.ctx, db.cancel = context.WithCancel(context.Background())
	return db
}

// createMigrationsTable creates the migrations table if it doesn't exist.
func (db *DB) createMigrationsTable() error {
	if _, err := db.conn.Exec("CREATE TABLE IF NOT EXISTS migrations (name VARCHAR(255) PRIMARY KEY);"); err != nil {
		return app.Errorf(app.EINTERNAL, "Error creating migrations table: %v", err)
	}
	return nil
}

// migrate sets up migration tracking and executes pending migration files.
//
// Migration files are embedded in the sqlite/migration folder and are executed
// in lexigraphical order.
//
// Once a migration is run, its name is stored in the 'migrations' table so it
// is not re-executed. Migrations run in a transaction to prevent partial
// migrations
func (db *DB) migrate() error {

	if err := db.createMigrationsTable(); err != nil {
		return err
	}

	names, err := fs.Glob(migrationsFS, "migration/*.sql")
	if err != nil {
		return app.Errorf(app.EINTERNAL, "Error retrieving migrations: %v", err)
	}

	sort.Strings(names)

	tx, err := db.conn.Begin()
	if err != nil {
		return app.Errorf(app.EINTERNAL, "Error creating transaction: %v", err)
	}

	for _, name := range names {
		if err := db.migrateFile(tx, name); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// migrate runs a single migration file within a transaction. On success, the
// migration file name is saved to the "migrations" table to prevent re-running.
func (db *DB) migrateFile(tx *sql.Tx, name string) error {

	var n int
	if err := tx.QueryRow("SELECT COUNT(*) FROM migrations WHERE name = $1", name).Scan(&n); err != nil {
		return app.Errorf(app.EINTERNAL, "Error checking migration: %v", err)
	} else if n != 0 {
		return nil
	}

	if buf, err := fs.ReadFile(migrationsFS, name); err != nil {

		return app.Errorf(app.EINTERNAL, "Error reading migration %s: %v", name, err)

	} else if _, err := tx.Exec(string(buf)); err != nil {

		return app.Errorf(app.EINTERNAL, "Error executing migration %s: %v", name, err)
	}

	if _, err := tx.Exec("INSERT INTO migrations (name) VALUES ($1)", name); err != nil {
		return app.Errorf(app.EINTERNAL, "Error saving migration: %v", err)
	}

	return nil
}


// BeginTx starts a transaction and returns a wrapper Tx type. This type
// provides a reference to the database and a fixed timestamp at the start of
// the transaction. The timestamp allows us to mock time during tests as well.
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {

	txFromContext := app.TxFromContext(ctx)
	if tx, ok := txFromContext.(*Tx); ok {

		// if the transaction has been initialized outside the package it is wrapped and the commit/rollback functions become No-Op delegating the commit/rollback to the layer above.
		return &Tx{
			Tx:        tx.Tx,
			db:        tx.db,
			now:       tx.now,
			_commit:   func() error { return nil },
			_rollback: func() error { return nil },
		}, nil

	} else {

		tx, err := db.conn.BeginTx(ctx, opts)
		if err != nil {
			return nil, app.Errorf(app.EINTERNAL, "Error creating transaction: %v", err)
		}

		// Return wrapper Tx that includes the transaction start time.
		return &Tx{
			Tx:        tx,
			db:        db,
			now:       db.Now().UTC().Truncate(time.Second),
			_commit:   tx.Commit,
			_rollback: tx.Rollback,
		}, nil
	}

}

// Close closes the database connection.
func (db *DB) Close() error {
	db.cancel()
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// MustOpen opens the database and panics on error.
func (db *DB) MustOpen() {
	if err := db.Open(); err != nil {
		panic(err)
	}
}

// Open database connection
func (db *DB) Open() error {

	if db.DSN == "" {
		return app.Errorf(app.EINVALID, "DSN is required")
	}

	var err error

	if db.conn, err = sql.Open("postgres", db.DSN); err != nil {
		return err
	}

	db.conn.SetMaxOpenConns(20)
	db.conn.SetMaxIdleConns(20)
	db.conn.SetConnMaxLifetime(5 * time.Minute)

	if err := db.migrate(); err != nil {
		return err
	}

	return nil
}

// GetRawConn returns the effective postgres.DB of std lib.
func (db *DB) GetRawConn() *sql.DB {
	return db.conn
}
