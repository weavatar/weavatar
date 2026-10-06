// Package biz holds the system module's business logic.
package biz

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/libtnb/cache"

	"github.com/weavatar/weavatar/internal/shared/appinfo"
)

const (
	usageCacheKey = "cdn:usage"
	// usageCacheSlack keeps yesterday's usage cached past midnight, while the
	// CDN may still be settling the new day's figure.
	usageCacheSlack   = 2 * time.Hour
	randomAvatarCount = 50
	randomAvatarSize  = "80"
)

// shanghai falls back to the fixed offset where tzdata is missing; China has
// observed no DST since 1991.
var shanghai = func() *time.Location {
	if loc, err := time.LoadLocation("Asia/Shanghai"); err == nil {
		return loc
	}
	return time.FixedZone("CST", 8*60*60)
}()

// Usage reports how many requests the CDN served for domain in [start, end).
type Usage interface {
	Fetch(ctx context.Context, domain string, start, end time.Time) (uint, error)
}

// Avatars is the slice of the avatar module the home page needs.
type Avatars interface {
	RandomHashes(ctx context.Context, n int) ([]string, error)
}

type SystemUsecase struct {
	usage   Usage
	avatars Avatars
	cache   cache.Cache
	domain  string
	log     *slog.Logger
}

func NewSystemUsecase(usage Usage, avatars Avatars, cache cache.Cache, domain appinfo.Domain, log *slog.Logger) *SystemUsecase {
	return &SystemUsecase{
		usage:   usage,
		avatars: avatars,
		cache:   cache,
		domain:  string(domain),
		log:     log,
	}
}

// Count is yesterday's CDN request count (China time), cached until shortly
// after midnight; 0 when the CDN cannot tell.
func (uc *SystemUsecase) Count(ctx context.Context) int64 {
	if usage := uc.cache.GetInt64(usageCacheKey, -1); usage != -1 {
		return usage
	}

	now := time.Now().In(shanghai)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, shanghai)
	usage, err := uc.usage.Fetch(ctx, uc.domain, today.AddDate(0, 0, -1), today)
	if err != nil {
		uc.log.WarnContext(ctx, "fetch cdn usage failed", slog.Any("err", err))
		return 0
	}

	count := int64(math.MaxInt64)
	if usage < math.MaxInt64 {
		count = int64(usage)
	}
	_ = uc.cache.Put(usageCacheKey, count, today.AddDate(0, 0, 1).Sub(now)+usageCacheSlack)
	return count
}

// RandomAvatars returns URLs of random uploaded avatars, or none on failure.
func (uc *SystemUsecase) RandomAvatars(ctx context.Context) []string {
	hashes, err := uc.avatars.RandomHashes(ctx, randomAvatarCount)
	if err != nil {
		uc.log.WarnContext(ctx, "pick random avatars failed", slog.Any("err", err))
		return []string{}
	}

	urls := make([]string, len(hashes))
	for i, hash := range hashes {
		urls[i] = "https://" + uc.domain + "/avatar/" + hash + "?s=" + randomAvatarSize
	}
	return urls
}
