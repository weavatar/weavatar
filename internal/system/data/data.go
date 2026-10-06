// Package data adapts the system module's ports.
package data

import (
	"context"
	"time"

	avatarbiz "github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/system/biz"
	"github.com/weavatar/weavatar/pkg/cdn"
)

type usage struct {
	cdn *cdn.Cdn
}

// avatars adapts the avatar module's usecase to the biz.Avatars port.
type avatars struct {
	uc *avatarbiz.AvatarUsecase
}

func NewUsage(c *cdn.Cdn) biz.Usage {
	return &usage{cdn: c}
}

func NewAvatars(uc *avatarbiz.AvatarUsecase) biz.Avatars {
	return &avatars{uc: uc}
}

func (u *usage) Fetch(ctx context.Context, domain string, start, end time.Time) (uint, error) {
	return u.cdn.GetUsage(ctx, domain, start, end)
}

func (a *avatars) RandomHashes(ctx context.Context, n int) ([]string, error) {
	return a.uc.Random(ctx, n)
}
