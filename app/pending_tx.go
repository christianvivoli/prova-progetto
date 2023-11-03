package app

import (
	"context"
	"time"
)

// PendingTx rapresent a return value that holds the transaction.
type PendingTx interface {
	// Commit commits the transaction.
	Commit() error
	// Rollback rollbacks the transaction.
	Rollback() error
	// Now returns the current time.
	Now() time.Time
}

// BeginTx defines a func that initializes a transaction, also returns a context that hold the returning trasaction, useful when you need to achieve a transaction isolation on multi layer.
type BeginTx func(ctx context.Context) (PendingTx, context.Context, error)
