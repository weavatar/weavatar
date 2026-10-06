// Package database carries a transaction in ctx so data adapters in any
// module join their caller's transaction.
package database

import (
	"context"

	"github.com/go-rio/rio"
)

type txKey struct{}

// Runner backs every module's TxRunner port.
type Runner struct {
	db *rio.DB
}

func NewRunner(db *rio.DB) *Runner {
	return &Runner{db: db}
}

// Q returns the transaction in ctx, else db; with a single database both
// share db's pool.
func Q(ctx context.Context, db *rio.DB) rio.Queryer {
	if tx, ok := ctx.Value(txKey{}).(*rio.Tx); ok {
		return tx
	}
	return db
}

// Transaction runs fn in a new transaction, or in a savepoint when ctx
// already carries one; fn must use the ctx it gets.
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
