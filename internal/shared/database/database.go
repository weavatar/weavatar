// Package database carries a transaction through the context, so data
// adapters that take their target from Q join whatever transaction their
// caller opened, even across modules.
package database

import (
	"context"

	"github.com/go-rio/rio"
)

type txKey struct{}

// Runner runs work in one transaction; modules see it through their own
// TxRunner port.
type Runner struct {
	db *rio.DB
}

func NewRunner(db *rio.DB) *Runner {
	return &Runner{db: db}
}

// Q returns ctx's transaction, or db outside one. The application has one
// database, so the transaction is always on db's pool.
func Q(ctx context.Context, db *rio.DB) rio.Queryer {
	if tx, ok := ctx.Value(txKey{}).(*rio.Tx); ok {
		return tx
	}
	return db
}

// Transaction runs fn in a new transaction on db, or in a savepoint when ctx
// already carries one, so a failed nested call rolls back only its own
// writes. fn must use the ctx it gets.
func Transaction(ctx context.Context, db *rio.DB, fn func(ctx context.Context) error) error {
	run := func(tx *rio.Tx) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	}
	if tx, ok := ctx.Value(txKey{}).(*rio.Tx); ok {
		return tx.Tx(ctx, run)
	}
	return db.Tx(ctx, run)
}

func (r *Runner) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return Transaction(ctx, r.db, fn)
}
