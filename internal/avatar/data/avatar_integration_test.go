//go:build integration

package data_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-rio/migrate"
	"github.com/go-rio/postgres"
	"github.com/go-rio/rio"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/avatar/data"
	"github.com/weavatar/weavatar/internal/migrations"
)

// fakeClock stamps CreatedAt/UpdatedAt deterministically.
type fakeClock struct {
	now time.Time
}

func (c *fakeClock) Now() time.Time { return c.now }

func TestIntegrationAvatarRepo_CRUD(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	db := newTestDB(t, clock)
	repo := data.NewAvatarRepo(db)
	ctx := t.Context()

	must.NoError(t, repo.Create(ctx, avatarOf("a", "u1")))
	clock.now = clock.now.Add(time.Minute)
	must.NoError(t, repo.Create(ctx, avatarOf("b", "u1")))
	must.NoError(t, repo.Create(ctx, avatarOf("c", "u2")))
	must.ErrorIs(t, repo.Create(ctx, avatarOf("a", "u2")), biz.ErrAvatarExists)

	list, total, err := repo.List(ctx, "u1", 1, 10)
	must.NoError(t, err)
	check.Equal(t, total, 2)
	must.Len(t, list, 2)
	check.Equal(t, list[0].Raw, "b") // newest first

	all, err := repo.ListAllByUser(ctx, "u1")
	must.NoError(t, err)
	check.Len(t, all, 2)
	none, err := repo.ListAllByUser(ctx, "u3")
	must.NoError(t, err)
	check.Len(t, none, 0)

	found, err := repo.Find(ctx, "u1", "a-md5")
	must.NoError(t, err)
	check.Equal(t, found.SHA256, "a-sha256")
	_, err = repo.Find(ctx, "u2", "a-md5")
	must.ErrorIs(t, err, rio.ErrNotFound)

	exists, err := repo.ExistsByRaw(ctx, "c")
	must.NoError(t, err)
	check.True(t, exists)

	clock.now = clock.now.Add(time.Minute)
	must.NoError(t, repo.Touch(ctx, found))
	touched, err := repo.Find(ctx, "u1", "a-sha256")
	must.NoError(t, err)
	check.True(t, touched.UpdatedAt.Equal(clock.now))
	check.True(t, touched.CreatedAt.Equal(found.CreatedAt))

	random, err := repo.RandomHashes(ctx, 2)
	must.NoError(t, err)
	must.Len(t, random, 2)
	check.True(t, strings.HasSuffix(random[0], "-sha256"))

	must.NoError(t, repo.Delete(ctx, found))
	exists, err = repo.ExistsByRaw(ctx, "a")
	must.NoError(t, err)
	check.False(t, exists)
}

func TestIntegrationAvatarRepo_FindForServeLoadsAppOverride(t *testing.T) {
	db := newTestDB(t, &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	repo := data.NewAvatarRepo(db)
	ctx := t.Context()

	avatar := avatarOf("a", "u1")
	must.NoError(t, repo.Create(ctx, avatar))
	must.NoError(t, rio.Insert(ctx, db, &biz.AppAvatar{AvatarSHA256: avatar.SHA256, AvatarMD5: avatar.MD5, AppID: "app1"}))

	served, err := repo.FindForServe(ctx, avatar.MD5, "app1")
	must.NoError(t, err)
	must.NotNil(t, served.AppSHA256.Row())
	check.Equal(t, served.AppSHA256.Row().AppID, "app1")
	must.NotNil(t, served.AppMD5.Row())

	other, err := repo.FindForServe(ctx, avatar.SHA256, "app2")
	must.NoError(t, err)
	check.Nil(t, other.AppSHA256.Row())
	check.Nil(t, other.AppMD5.Row())

	plain, err := repo.FindForServe(ctx, avatar.SHA256, "")
	must.NoError(t, err)
	check.True(t, plain.AppSHA256.Loaded())
	check.Nil(t, plain.AppSHA256.Row())

	_, err = repo.FindForServe(ctx, "missing", "")
	must.ErrorIs(t, err, rio.ErrNotFound)
}

func TestIntegrationTxRunner_RollsBackOnError(t *testing.T) {
	db := newTestDB(t, &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	repo := data.NewAvatarRepo(db)
	tx := data.NewTxRunner(db)
	ctx := t.Context()
	errWrite := errors.New("file write failed")

	var createErr error
	err := tx.Run(ctx, func(ctx context.Context) error {
		createErr = repo.Create(ctx, avatarOf("a", "u1"))
		return errWrite
	})

	must.NoError(t, createErr)
	must.ErrorIs(t, err, errWrite)
	exists, err := repo.ExistsByRaw(ctx, "a")
	must.NoError(t, err)
	check.False(t, exists)
}

func TestIntegrationImageRepo(t *testing.T) {
	db := newTestDB(t, &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	images := data.NewImageRepo(db)
	ctx := t.Context()

	_, err := images.Find(ctx, "h")
	must.ErrorIs(t, err, rio.ErrNotFound)

	must.NoError(t, images.Create(ctx, &biz.Image{Hash: "h", Banned: true, Remark: "porn"}))
	image, err := images.Find(ctx, "h")
	must.NoError(t, err)
	check.True(t, image.Banned)
	check.Equal(t, image.Remark, "porn")
}

// newTestDB migrates TEST_DATABASE_URL (PostgreSQL) and empties the avatar
// tables; rows are stamped by clock.
func newTestDB(t *testing.T, clock *fakeClock) *rio.DB {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL is not set")
	}
	db, err := postgres.Open(dsn, rio.WithClock(clock.Now))
	must.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	m, err := migrate.New(db.Unwrap(), migrate.Postgres, migrate.WithCollection(migrations.Collection()))
	must.NoError(t, err)
	must.NoError(t, m.Up(t.Context()))
	_, err = rio.Exec(t.Context(), db, "TRUNCATE avatars, app_avatars, images")
	must.NoError(t, err)

	return db
}

func avatarOf(raw, userID string) *biz.Avatar {
	return &biz.Avatar{SHA256: raw + "-sha256", MD5: raw + "-md5", Raw: raw, UserID: userID}
}
