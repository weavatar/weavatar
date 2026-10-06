// Package data implements the avatar module's ports.
package data

import (
	"context"

	"github.com/go-rio/rio"

	"github.com/weavatar/weavatar/internal/avatar/biz"
)

type txKey struct{}

type txRunner struct {
	db *rio.DB
}

func NewTxRunner(db *rio.DB) biz.TxRunner {
	return &txRunner{db: db}
}

// Run nests as a savepoint when ctx already carries a transaction.
func (r *txRunner) Run(ctx context.Context, fn func(ctx context.Context) error) error {
	return queryer(ctx, r.db).Tx(ctx, func(tx *rio.Tx) error {
		return fn(context.WithValue(ctx, txKey{}, tx))
	})
}

// queryer joins the transaction a txRunner put in ctx.
func queryer(ctx context.Context, db *rio.DB) rio.Queryer {
	if tx, ok := ctx.Value(txKey{}).(*rio.Tx); ok {
		return tx
	}
	return db
}
