package data

import (
	"context"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/pkg/audit"
	"github.com/weavatar/weavatar/pkg/avatars"
	"github.com/weavatar/weavatar/pkg/cdn"
	"github.com/weavatar/weavatar/pkg/qqhash"
)

type fetcher struct {
	gravatarURL string
}

type purger struct {
	cdn *cdn.Cdn
}

func NewFetcher(gravatarURL appinfo.GravatarURL) biz.Fetcher {
	return fetcher{gravatarURL: string(gravatarURL)}
}

func NewPurger(c *cdn.Cdn) biz.Purger {
	return &purger{cdn: c}
}

func NewAuditor(a *audit.Audit) biz.Auditor {
	return a
}

// NewQQHashes accepts empty tables: lookups then miss, disabling the fallback.
func NewQQHashes(tables *qqhash.Tables) biz.QQHashes {
	return tables
}

func (f fetcher) Gravatar(ctx context.Context, hash string) ([]byte, error) {
	return avatars.Gravatar(ctx, f.gravatarURL, hash)
}

func (fetcher) QQ(ctx context.Context, qq string) ([]byte, error) {
	return avatars.Qq(ctx, qq)
}

func (p *purger) Refresh(ctx context.Context, urls []string) error {
	return p.cdn.RefreshUrl(ctx, urls)
}
