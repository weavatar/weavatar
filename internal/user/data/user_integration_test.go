//go:build integration

package data_test

import (
	"os"
	"testing"

	"github.com/go-rio/migrate"
	"github.com/go-rio/postgres"
	"github.com/go-rio/rio"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/migrations"
	"github.com/weavatar/weavatar/internal/user/biz"
	"github.com/weavatar/weavatar/internal/user/data"
)

func TestIntegrationUserRepo_DeleteKeepsTheRow(t *testing.T) {
	db := newTestDB(t)
	repo := data.NewUserRepo(db)
	ctx := t.Context()
	user := &biz.User{ID: "u1", OpenID: "open-1", UnionID: "union-1", Nickname: "alice", Avatar: "https://a/1.png", RealName: true}
	must.NoError(t, repo.Create(ctx, user))

	must.NoError(t, repo.Delete(ctx, user))

	_, err := repo.Find(ctx, "u1")
	check.ErrorIs(t, err, rio.ErrNotFound)
	_, err = repo.FindByUnionID(ctx, "union-1")
	check.ErrorIs(t, err, rio.ErrNotFound)

	// the row stays, every column intact, only stamped
	row, err := rio.From[biz.User]().WithTrashed().Where("id = ?").Must().First(ctx, db, "u1")
	must.NoError(t, err)
	check.NotNil(t, row.DeletedAt)
	check.Equal(t, row.UnionID, "union-1")
	check.Equal(t, row.Nickname, "alice")
	check.True(t, row.RealName)

	// the identity may register again
	must.NoError(t, repo.Create(ctx, &biz.User{ID: "u2", OpenID: "open-1", UnionID: "union-1", Nickname: "alice again"}))
}

// newTestDB migrates TEST_DATABASE_URL (PostgreSQL) and empties the users table.
func newTestDB(t *testing.T) *rio.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	db, err := postgres.Open(dsn)
	must.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	m, err := migrate.New(db.Unwrap(), migrate.Postgres, migrate.WithCollection(migrations.Collection()))
	must.NoError(t, err)
	must.NoError(t, m.Up(t.Context()))
	_, err = rio.Exec(t.Context(), db, "TRUNCATE users")
	must.NoError(t, err)

	return db
}
