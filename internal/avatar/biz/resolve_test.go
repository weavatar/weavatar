package biz_test

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/png"
	"log/slog"
	"testing"
	"time"

	"github.com/go-rio/rio"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/cache"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	mocksbiz "github.com/weavatar/weavatar/internal/mocks/avatar/biz"
	"github.com/weavatar/weavatar/pkg/embed"
	"github.com/weavatar/weavatar/pkg/imaging"
	"github.com/weavatar/weavatar/pkg/queue"
)

const (
	sha256Hash = "8d6c4ec2a9bd2a6b9e3d3c3b8f0ea7c8b2d4a1f0e9c8b7a6d5c4b3a2f1e0d9c8"
	md5Hash    = "0bc83cb571cd1c50ba6f3e8a78ef1346"
)

var errUpstream = errors.New("upstream down")

// deps holds every port of the usecase; a nil Func panics when called, which
// pins down the calls a rule must not make.
type deps struct {
	repo    *mocksbiz.AvatarRepo
	images  *mocksbiz.ImageRepo
	tx      *mocksbiz.TxRunner
	users   *mocksbiz.Users
	store   *mocksbiz.Store
	fetcher *mocksbiz.Fetcher
	qq      *mocksbiz.QQHashes
	gen     *mocksbiz.Generator
	purger  *mocksbiz.Purger
	auditor *mocksbiz.Auditor
	queue   *queue.Queue
	cache   cache.Cache
}

func newUsecase(t *testing.T) (*biz.AvatarUsecase, *deps) {
	t.Helper()

	d := &deps{
		repo:    &mocksbiz.AvatarRepo{},
		images:  &mocksbiz.ImageRepo{},
		tx:      &mocksbiz.TxRunner{RunFunc: func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }},
		users:   &mocksbiz.Users{},
		store:   &mocksbiz.Store{},
		fetcher: &mocksbiz.Fetcher{},
		qq:      &mocksbiz.QQHashes{},
		gen:     &mocksbiz.Generator{},
		purger:  &mocksbiz.Purger{}, // purges must go through the queue
		auditor: &mocksbiz.Auditor{},
		queue:   queue.New(10, slog.New(slog.DiscardHandler)), // never started: jobs stay countable
		cache:   cache.NewCache(),
	}
	t.Cleanup(func() { _ = d.queue.Stop(context.Background()) })
	uc := biz.NewAvatarUsecase(d.repo, d.images, d.tx, d.users, d.store, d.fetcher, d.qq, d.gen,
		d.purger, d.auditor, d.queue, d.cache, "weavatar.com", slog.New(slog.DiscardHandler))

	return uc, d
}

func TestResolve_ServesUploadFirst(t *testing.T) {
	uc, d := newUsecase(t)
	updated := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	d.repo.FindForServeFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		return &biz.Avatar{SHA256: sha256Hash, UserID: "u1", UpdatedAt: updated}, nil
	}
	d.store.ReadAvatarFunc = func(string) ([]byte, error) { return pngOf(t, 100), nil }
	d.audited(false)

	res, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: sha256Hash, Ext: "png", Size: 60})

	must.NoError(t, err)
	check.Equal(t, res.From, biz.FromWeAvatar)
	check.Equal(t, res.LastModified, updated)
	check.Equal(t, sizeOf(t, res.Image), 60)
	check.Len(t, d.fetcher.GravatarCalls(), 0)
}

func TestResolve_PrefersAppOverride(t *testing.T) {
	uc, d := newUsecase(t)
	appUpdated := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)
	d.repo.FindForServeFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		avatar := &biz.Avatar{SHA256: sha256Hash}
		avatar.AppSHA256.Set(&biz.AppAvatar{AppID: "app1", UpdatedAt: appUpdated})
		avatar.AppMD5.Set(nil)
		return avatar, nil
	}
	d.store.ReadAppAvatarFunc = func(string, string) ([]byte, error) { return pngOf(t, 50), nil }
	d.audited(false)

	res, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: sha256Hash, AppID: "app1", Ext: "png", Size: 50})

	must.NoError(t, err)
	check.Equal(t, res.LastModified, appUpdated)
	read := d.store.ReadAppAvatarCalls()
	must.Len(t, read, 1)
	check.Equal(t, read[0].AppID, "app1")
	check.Equal(t, read[0].Sha256, sha256Hash)
}

func TestResolve_FallsBackToGravatarAndCachesIt(t *testing.T) {
	uc, d := newUsecase(t)
	d.noUpload()
	d.noCache()
	d.fetcher.GravatarFunc = func(context.Context, string) ([]byte, error) { return pngOf(t, 80), nil }
	d.audited(false)

	res, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: md5Hash, Ext: "png", Size: 80})

	must.NoError(t, err)
	check.Equal(t, res.From, biz.FromGravatar)
	written := d.store.WriteCacheCalls()
	must.Len(t, written, 1)
	check.Equal(t, written[0].Kind, "gravatar")
	check.Equal(t, written[0].Key, md5Hash)
}

func TestResolve_ServesFreshCacheWithoutFetching(t *testing.T) {
	uc, d := newUsecase(t)
	d.noUpload()
	cachedAt := time.Now().Add(-time.Hour)
	d.store.ReadCacheFunc = func(string, string) ([]byte, time.Time, bool) { return pngOf(t, 80), cachedAt, true }
	d.audited(false)

	res, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: md5Hash, Ext: "png", Size: 80})

	must.NoError(t, err)
	check.Equal(t, res.From, biz.FromGravatar)
	check.Equal(t, res.LastModified, cachedAt)
}

func TestResolve_RefetchesStaleCache(t *testing.T) {
	uc, d := newUsecase(t)
	d.noUpload()
	d.store.ReadCacheFunc = func(string, string) ([]byte, time.Time, bool) {
		return pngOf(t, 80), time.Now().Add(-biz.CacheTTL - time.Minute), true
	}
	d.store.WriteCacheFunc = func(string, string, []byte) error { return nil }
	d.fetcher.GravatarFunc = func(context.Context, string) ([]byte, error) { return pngOf(t, 80), nil }
	d.audited(false)

	_, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: md5Hash, Ext: "png", Size: 80})

	must.NoError(t, err)
	check.Len(t, d.fetcher.GravatarCalls(), 1)
}

func TestResolve_FallsBackToQQWithoutBanCheck(t *testing.T) {
	uc, d := newUsecase(t)
	d.noUpload()
	d.noCache()
	d.fetcher.GravatarFunc = func(context.Context, string) ([]byte, error) { return nil, errUpstream }
	d.qq.LookupFunc = func(string) (uint32, bool) { return 10001, true }
	d.fetcher.QQFunc = func(context.Context, string) ([]byte, error) { return pngOf(t, 100), nil }
	// images.FindFunc stays nil: a ban check would panic

	res, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: md5Hash, Ext: "png", Size: 80})

	must.NoError(t, err)
	check.Equal(t, res.From, biz.FromQQ)
	fetched := d.fetcher.QQCalls()
	must.Len(t, fetched, 1)
	check.Equal(t, fetched[0].Qq, "10001")
}

func TestResolve_SwapsBannedImage(t *testing.T) {
	uc, d := newUsecase(t)
	d.noUpload()
	d.noCache()
	d.fetcher.GravatarFunc = func(context.Context, string) ([]byte, error) { return pngOf(t, 80), nil }
	d.audited(true)

	res, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: md5Hash, Ext: "png", Size: 80})

	must.NoError(t, err)
	ban, err := embed.DefaultFS.ReadFile("default/ban.png")
	must.NoError(t, err)
	img, _, err := imaging.Decode(ban)
	must.NoError(t, err)
	want, err := imaging.Encode(imaging.Resize(img, 80, 80), "png")
	must.NoError(t, err)
	check.Equal(t, res.From, biz.FromGravatar)
	check.True(t, bytes.Equal(res.Image, want))
}

func TestResolve_QueuesOneAuditForUnauditedImage(t *testing.T) {
	uc, d := newUsecase(t)
	d.noUpload()
	d.noCache()
	d.fetcher.GravatarFunc = func(context.Context, string) ([]byte, error) { return pngOf(t, 80), nil }
	d.images.FindFunc = func(context.Context, string) (*biz.Image, error) { return nil, rio.ErrNotFound }

	for range 2 {
		res, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: md5Hash, Ext: "png", Size: 80})
		must.NoError(t, err)
		check.Equal(t, res.From, biz.FromGravatar)
	}

	check.Equal(t, d.queue.Len(), 1)
}

func TestResolve_DefaultWhenNothingFound(t *testing.T) {
	tests := []struct {
		name     string
		dflt     string
		notFound bool
		redirect string
	}{
		{name: "404", dflt: "404", notFound: true},
		{name: "url redirects", dflt: "https://example.com/a.png", redirect: "https://example.com/a.png"},
		{name: "generated", dflt: "identicon"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, d := newUsecase(t)
			d.noUpload()
			d.noCache()
			d.fetcher.GravatarFunc = func(context.Context, string) ([]byte, error) { return nil, errUpstream }
			d.qq.LookupFunc = func(string) (uint32, bool) { return 0, false }
			d.gen.GenerateFunc = func(string, string, int, string) ([]byte, error) { return pngOf(t, 120), nil }

			res, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: md5Hash, Ext: "png", Size: 64, Default: tt.dflt})

			must.NoError(t, err)
			check.Equal(t, res.NotFound, tt.notFound)
			check.Equal(t, res.Redirect, tt.redirect)
			if tt.notFound || tt.redirect != "" {
				check.Len(t, d.gen.GenerateCalls(), 0)
				return
			}
			check.Equal(t, res.From, biz.FromWeAvatar)
			check.Equal(t, sizeOf(t, res.Image), 64)
			generated := d.gen.GenerateCalls()
			must.Len(t, generated, 1)
			check.Equal(t, generated[0].Kind, "identicon")
			check.Equal(t, generated[0].Seed, md5Hash)
			check.Equal(t, generated[0].Size, 64)
		})
	}
}

func TestResolve_ForceSkipsLookups(t *testing.T) {
	uc, d := newUsecase(t)
	// repo, store and fetcher funcs stay nil: any lookup would panic
	d.gen.GenerateFunc = func(string, string, int, string) ([]byte, error) { return pngOf(t, 80), nil }

	res, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: md5Hash, Ext: "png", Size: 80, Force: true, Default: "mp"})

	must.NoError(t, err)
	check.Equal(t, res.From, biz.FromWeAvatar)
	check.Equal(t, d.gen.GenerateCalls()[0].Kind, "mp")
}

func TestResolve_InvalidHashDrawsDefaultWithStableSeed(t *testing.T) {
	uc, d := newUsecase(t)
	d.gen.GenerateFunc = func(string, string, int, string) ([]byte, error) { return pngOf(t, 80), nil }

	_, err := uc.Resolve(t.Context(), biz.ResolveRequest{Ext: "png", Size: 80, Default: "color"})

	must.NoError(t, err)
	check.Equal(t, d.gen.GenerateCalls()[0].Seed, "weavatar")
}

func TestResolve_InvalidHashSkipsOwnerLookup(t *testing.T) {
	uc, d := newUsecase(t)
	// repo.FindForServeFunc stays nil: querying a malformed hash would panic
	d.gen.GenerateFunc = func(string, string, int, string) ([]byte, error) { return pngOf(t, 80), nil }

	_, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: "not-a-hash", Ext: "png", Size: 80, Default: "initials"})

	must.NoError(t, err)
	check.Equal(t, d.gen.GenerateCalls()[0].Kind, "initials")
}

func TestResolve_InitialsFromOwnerNickname(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.FindForServeFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		return &biz.Avatar{SHA256: sha256Hash, UserID: "u1"}, nil
	}
	d.users.NicknameFunc = func(context.Context, string) (string, error) { return "耗子", nil }
	d.gen.GenerateFunc = func(string, string, int, string) ([]byte, error) { return pngOf(t, 80), nil }

	_, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: sha256Hash, Ext: "png", Size: 80, Force: true, Default: "initials"})

	must.NoError(t, err)
	check.Equal(t, d.users.NicknameCalls()[0].UserID, "u1")
	check.Equal(t, d.gen.GenerateCalls()[0].Text, "耗")
}

func TestResolve_InitialsPreferRequestName(t *testing.T) {
	uc, d := newUsecase(t)
	d.gen.GenerateFunc = func(string, string, int, string) ([]byte, error) { return pngOf(t, 80), nil }

	_, err := uc.Resolve(t.Context(), biz.ResolveRequest{Hash: sha256Hash, Ext: "png", Size: 80, Force: true, Default: "initials", Name: "alice"})

	must.NoError(t, err)
	check.Equal(t, d.gen.GenerateCalls()[0].Text, "a")
}

// noUpload makes every hash miss the WeAvatar uploads.
func (d *deps) noUpload() {
	d.repo.FindForServeFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		return nil, rio.ErrNotFound
	}
}

// noCache makes every fetch miss the disk cache and caches it successfully.
func (d *deps) noCache() {
	d.store.ReadCacheFunc = func(string, string) ([]byte, time.Time, bool) { return nil, time.Time{}, false }
	d.store.WriteCacheFunc = func(string, string, []byte) error { return nil }
}

func (d *deps) audited(banned bool) {
	d.images.FindFunc = func(_ context.Context, hash string) (*biz.Image, error) {
		return &biz.Image{Hash: hash, Banned: banned}, nil
	}
}

func pngOf(t *testing.T, size int) []byte {
	t.Helper()
	var buf bytes.Buffer
	must.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size))))
	return buf.Bytes()
}

func sizeOf(t *testing.T, img []byte) int {
	t.Helper()
	cfg, err := png.DecodeConfig(bytes.NewReader(img))
	must.NoError(t, err)
	check.Equal(t, cfg.Width, cfg.Height)
	return cfg.Width
}

// receive guards a channel read so a job that never runs fails the test
// instead of hanging it.
func receive[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for a queued job")
		var zero T
		return zero
	}
}
