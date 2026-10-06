//go:build integration

package data_test

import (
	"context"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-rio/migrate"
	"github.com/go-rio/postgres"
	"github.com/go-rio/rio"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/migrations"
	"github.com/weavatar/weavatar/internal/user/biz"
	"github.com/weavatar/weavatar/internal/user/data"
)

func TestIntegrationDeleteKeepsTheRow(t *testing.T) {
	db := migratedFixture(t)
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

// migratedFixture opens TEST_DATABASE_URL on a fresh schema with every migration applied.
func migratedFixture(t *testing.T) *rio.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	admin, err := postgres.Open(dsn)
	must.NoError(t, err)
	schema := "user_" + strconv.Itoa(os.Getpid()) + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)
	_, err = rio.Exec(t.Context(), admin, `CREATE SCHEMA "`+schema+`"`)
	must.NoError(t, err)

	parsed, err := url.Parse(dsn)
	must.NoError(t, err)
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	db, err := postgres.Open(parsed.String())
	must.NoError(t, err)
	t.Cleanup(func() {
		check.NoError(t, db.Close())
		_, err := rio.Exec(context.Background(), admin, `DROP SCHEMA "`+schema+`" CASCADE`)
		check.NoError(t, err)
		check.NoError(t, admin.Close())
	})

	m, err := migrate.New(db.Unwrap(), migrate.Postgres, migrate.WithCollection(migrations.Collection()))
	must.NoError(t, err)
	must.NoError(t, m.Up(t.Context()))
	return db
}
