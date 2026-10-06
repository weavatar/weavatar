//go:build integration

package database_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/go-rio/postgres"
	"github.com/go-rio/rio"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/shared/database"
)

func TestIntegrationRunCommitsOrRollsBack(t *testing.T) {
	db := itemsFixture(t)
	runner := database.NewRunner(db)
	failure := errors.New("failure")

	must.NoError(t, runner.Run(t.Context(), func(ctx context.Context) error {
		_, ok := database.Q(ctx, db).(*rio.Tx)
		check.True(t, ok)
		return insertItem(ctx, db, 1, "committed")
	}))
	err := runner.Run(t.Context(), func(ctx context.Context) error {
		check.NoError(t, insertItem(ctx, db, 2, "rolled back"))
		return failure
	})

	check.ErrorIs(t, err, failure)
	check.DeepEqual(t, itemNames(t, db), []string{"committed"})
}

func TestIntegrationNestedRunIsASavepoint(t *testing.T) {
	db := itemsFixture(t)
	runner := database.NewRunner(db)
	failure := errors.New("inner failure")

	must.NoError(t, runner.Run(t.Context(), func(ctx context.Context) error {
		check.NoError(t, insertItem(ctx, db, 1, "outer"))
		err := runner.Run(ctx, func(ctx context.Context) error {
			check.NoError(t, insertItem(ctx, db, 2, "inner"))
			return failure
		})
		check.ErrorIs(t, err, failure)
		return insertItem(ctx, db, 3, "after the savepoint")
	}))

	check.DeepEqual(t, itemNames(t, db), []string{"outer", "after the savepoint"})
}

func TestIntegrationOuterFailureUndoesNestedWork(t *testing.T) {
	db := itemsFixture(t)
	runner := database.NewRunner(db)

	err := runner.Run(t.Context(), func(ctx context.Context) error {
		check.NoError(t, runner.Run(ctx, func(ctx context.Context) error {
			return insertItem(ctx, db, 1, "inner")
		}))
		return errors.New("outer failure")
	})

	check.Error(t, err)
	check.Len(t, itemNames(t, db), 0)
}

// itemsFixture opens TEST_DATABASE_URL on a fresh schema holding an items table.
func itemsFixture(t *testing.T) *rio.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	admin, err := postgres.Open(dsn)
	must.NoError(t, err)
	schema := "tx_" + strconv.Itoa(os.Getpid()) + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)
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

	_, err = rio.Exec(t.Context(), db, `CREATE TABLE items (id bigint PRIMARY KEY, name text NOT NULL)`)
	must.NoError(t, err)
	return db
}

func insertItem(ctx context.Context, db *rio.DB, id int64, name string) error {
	_, err := rio.Exec(ctx, database.Q(ctx, db), `INSERT INTO items (id, name) VALUES (?, ?)`, id, name)
	return err
}

func itemNames(t *testing.T, db *rio.DB) []string {
	t.Helper()
	rows, err := rio.Raw[struct{ Name string }](`SELECT name FROM items ORDER BY id`).All(t.Context(), db)
	must.NoError(t, err)
	names := make([]string, len(rows))
	for i, row := range rows {
		names[i] = row.Name
	}
	return names
}
