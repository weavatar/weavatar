package biz_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/cache"

	mocksbiz "github.com/weavatar/weavatar/internal/mocks/system/biz"
	"github.com/weavatar/weavatar/internal/system/biz"
)

func TestCount_FetchesYesterdayInChinaAndCaches(t *testing.T) {
	usage := &mocksbiz.Usage{
		FetchFunc: func(context.Context, string, time.Time, time.Time) (uint, error) { return 42, nil },
	}
	uc := newUsecase(usage, &mocksbiz.Avatars{})

	check.Equal(t, uc.Count(t.Context()), 42)
	check.Equal(t, uc.Count(t.Context()), 42)

	fetched := usage.FetchCalls()
	must.Len(t, fetched, 1) // the second count is served from cache
	check.Equal(t, fetched[0].Domain, "weavatar.com")
	start, end := fetched[0].Start, fetched[0].End
	check.Equal(t, end.Sub(start), 24*time.Hour)
	check.Equal(t, end.Format("15:04:05 -0700"), "00:00:00 +0800")
	check.True(t, !time.Now().Before(end) && time.Now().Before(end.Add(24*time.Hour)))
}

func TestCount_FailureReportsZeroUncached(t *testing.T) {
	usage := &mocksbiz.Usage{
		FetchFunc: func(context.Context, string, time.Time, time.Time) (uint, error) { return 0, errors.New("cdn down") },
	}
	uc := newUsecase(usage, &mocksbiz.Avatars{})

	check.Equal(t, uc.Count(t.Context()), 0)
	check.Equal(t, uc.Count(t.Context()), 0)

	check.Len(t, usage.FetchCalls(), 2)
}

func TestRandomAvatars_BuildsURLs(t *testing.T) {
	avatars := &mocksbiz.Avatars{
		RandomHashesFunc: func(context.Context, int) ([]string, error) { return []string{"abc"}, nil },
	}
	uc := newUsecase(&mocksbiz.Usage{}, avatars)

	check.DeepEqual(t, uc.RandomAvatars(t.Context()), []string{"https://weavatar.com/avatar/abc?s=80"})
	check.Equal(t, avatars.RandomHashesCalls()[0].N, 50)
}

func TestRandomAvatars_FailureReportsNone(t *testing.T) {
	avatars := &mocksbiz.Avatars{
		RandomHashesFunc: func(context.Context, int) ([]string, error) { return nil, errors.New("db down") },
	}
	uc := newUsecase(&mocksbiz.Usage{}, avatars)

	got := uc.RandomAvatars(t.Context())

	must.NotNil(t, got)
	check.Len(t, got, 0)
}

func newUsecase(usage *mocksbiz.Usage, avatars *mocksbiz.Avatars) *biz.SystemUsecase {
	return biz.NewSystemUsecase(usage, avatars, cache.NewCache(), "weavatar.com", slog.New(slog.DiscardHandler))
}
