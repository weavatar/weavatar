package data_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/avatar/data"
)

const hash = "0bc83cb571cd1c50ba6f3e8a78ef1346"

func TestStoreWritesReadsAndRemovesAvatars(t *testing.T) {
	root := t.TempDir()
	s := data.NewStoreAt(root)

	must.NoError(t, s.WriteAvatar(hash, []byte("img")))
	_, err := os.Stat(filepath.Join(root, "upload", "default", "0b", hash))
	must.NoError(t, err)
	got, err := s.ReadAvatar(hash)
	must.NoError(t, err)
	check.Equal(t, string(got), "img")

	must.NoError(t, s.RemoveAvatar(hash))
	_, err = s.ReadAvatar(hash)
	must.Error(t, err)
	check.NoError(t, s.RemoveAvatar(hash)) // already gone
}

func TestStoreReadsAppAvatar(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "upload", "app", "app1", "0b", hash)
	must.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	must.NoError(t, os.WriteFile(path, []byte("app"), 0o644))

	got, err := data.NewStoreAt(root).ReadAppAvatar("app1", hash)

	must.NoError(t, err)
	check.Equal(t, string(got), "app")
}

func TestStoreCacheReportsModTime(t *testing.T) {
	s := data.NewStoreAt(t.TempDir())
	_, _, ok := s.ReadCache("qq", "10001")
	check.False(t, ok)

	must.NoError(t, s.WriteCache("qq", "10001", []byte("img")))
	img, modTime, ok := s.ReadCache("qq", "10001")

	must.True(t, ok)
	check.Equal(t, string(img), "img")
	check.True(t, time.Since(modTime) < time.Minute)
}

func TestStoreRejectsPathEscapes(t *testing.T) {
	s := data.NewStoreAt(t.TempDir())

	check.Error(t, s.WriteCache("../x", hash, nil))
	check.Error(t, s.WriteChecker("a", nil))
	_, err := s.ReadAppAvatar("../../etc", hash)
	check.Error(t, err)
	_, err = s.ReadAvatar("../" + hash)
	check.Error(t, err)
}

func TestStorePurgeCacheKeepsFreshFiles(t *testing.T) {
	root := t.TempDir()
	s := data.NewStoreAt(root)
	must.NoError(t, s.WriteCache("gravatar", hash, []byte("old")))
	must.NoError(t, s.WriteCache("qq", "10001", []byte("new")))
	old := filepath.Join(root, "cache", "gravatar", "0b", hash)
	must.NoError(t, os.Chtimes(old, time.Now(), time.Now().Add(-15*24*time.Hour)))

	must.NoError(t, s.PurgeCache(t.Context(), 14*24*time.Hour))

	_, err := os.Stat(old)
	check.True(t, os.IsNotExist(err))
	_, _, ok := s.ReadCache("qq", "10001")
	check.True(t, ok)
}

func TestStorePurgeWithoutCacheDirSucceeds(t *testing.T) {
	check.NoError(t, data.NewStoreAt(t.TempDir()).PurgeCache(t.Context(), time.Hour))
}
