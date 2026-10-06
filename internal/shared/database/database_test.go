package database_test

import (
	"testing"

	"github.com/go-rio/postgres"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/shared/database"
)

func TestQOutsideATransactionIsThePool(t *testing.T) {
	// Open validates the DSN without connecting
	db, err := postgres.Open("postgres://localhost/weavatar")
	must.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	check.True(t, database.Q(t.Context(), db) == db)
}
