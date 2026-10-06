package data

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/weavatar/weavatar/internal/avatar/biz"
)

// segment admits the hashes, QQ numbers, app IDs and kinds that become path
// elements, so request values cannot climb out of the storage root.
var segment = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// store lays files out as <root>/<dirs...>/<key[:2]>/<key>.
type store struct {
	root string
}

// NewStore keeps files under ./storage, relative to the working directory.
func NewStore() biz.Store {
	return NewStoreAt("storage")
}

func NewStoreAt(root string) biz.Store {
	return &store{root: root}
}

func (s *store) WriteAvatar(sha256 string, img []byte) error {
	return s.write(img, sha256, "upload", "default")
}

func (s *store) ReadAvatar(sha256 string) ([]byte, error) {
	return s.read(sha256, "upload", "default")
}

func (s *store) RemoveAvatar(sha256 string) error {
	path, err := s.path(sha256, "upload", "default")
	if err != nil {
		return err
	}
	if err = os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (s *store) ReadAppAvatar(appID, sha256 string) ([]byte, error) {
	return s.read(sha256, "upload", "app", appID)
}

func (s *store) ReadCache(kind, key string) ([]byte, time.Time, bool) {
	path, err := s.path(key, "cache", kind)
	if err != nil {
		return nil, time.Time{}, false
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, time.Time{}, false
	}
	img, err := os.ReadFile(path) //nolint:gosec // path built from validated segments
	if err != nil {
		return nil, time.Time{}, false
	}
	return img, info.ModTime(), true
}

func (s *store) WriteCache(kind, key string, img []byte) error {
	return s.write(img, key, "cache", kind)
}

func (s *store) WriteChecker(imgHash string, img []byte) error {
	return s.write(img, imgHash, "checker")
}

// PurgeCache skips what it cannot remove and reports the first failure. The
// walk is rooted so a symlink swapped in mid-walk cannot redirect a removal.
func (s *store) PurgeCache(ctx context.Context, olderThan time.Duration) error {
	root, err := os.OpenRoot(filepath.Join(s.root, "cache"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil // nothing cached yet
	}
	if err != nil {
		return fmt.Errorf("purge avatar cache: %w", err)
	}
	defer func() { _ = root.Close() }()

	deadline := time.Now().Add(-olderThan)
	var (
		failed int
		first  error
	)
	fail := func(err error) {
		if failed++; first == nil {
			first = err
		}
	}

	err = fs.WalkDir(root.FS(), ".", func(path string, d fs.DirEntry, err error) error {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) { // vanished mid-walk
				fail(err)
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			if !errors.Is(err, fs.ErrNotExist) {
				fail(err)
			}
			return nil
		}
		if info.ModTime().Before(deadline) {
			if err = root.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
				fail(err)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("purge avatar cache: %w", err)
	}
	if first != nil {
		return fmt.Errorf("purge avatar cache: %d entries failed, first: %w", failed, first)
	}
	return nil
}

func (s *store) path(key string, dirs ...string) (string, error) {
	if len(key) < 2 || !segment.MatchString(key) {
		return "", fmt.Errorf("store: invalid key %q", key)
	}
	for _, dir := range dirs {
		if !segment.MatchString(dir) {
			return "", fmt.Errorf("store: invalid directory %q", dir)
		}
	}

	elems := make([]string, 0, len(dirs)+3)
	elems = append(elems, s.root)
	elems = append(elems, dirs...)
	elems = append(elems, key[:2], key)
	return filepath.Join(elems...), nil
}

func (s *store) read(key string, dirs ...string) ([]byte, error) {
	path, err := s.path(key, dirs...)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(path) //nolint:gosec // path built from validated segments
}

// write replaces the file atomically, so concurrent readers never see a
// partial image.
func (s *store) write(img []byte, key string, dirs ...string) error {
	path, err := s.path(key, dirs...)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err = os.MkdirAll(dir, 0o755); err != nil { //nolint:gosec // avatars are public images
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }() // no-op after the rename

	if _, err = tmp.Write(img); err != nil {
		_ = tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	if err = os.Chmod(tmp.Name(), 0o644); err != nil { //nolint:gosec // avatars are public images
		return err
	}
	return os.Rename(tmp.Name(), path)
}
