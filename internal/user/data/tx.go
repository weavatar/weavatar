package data

import (
	"github.com/go-rio/rio"

	"github.com/weavatar/weavatar/internal/shared/database"
	"github.com/weavatar/weavatar/internal/user/biz"
)

func NewTxRunner(db *rio.DB) biz.TxRunner {
	return database.NewRunner(db)
}
