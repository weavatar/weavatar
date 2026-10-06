//go:build wireinject

// Package avatar is the avatar module's assembly.
package avatar

import (
	"github.com/libtnb/wire"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/avatar/data"
	"github.com/weavatar/weavatar/internal/avatar/service"
	"github.com/weavatar/weavatar/internal/shared/registry"
)

var Module = wire.New().
	Provide(data.NewAvatarRepo).
	Provide(data.NewImageRepo).
	Provide(data.NewTxRunner).
	Provide(data.NewUsers).
	Provide(data.NewStore).
	Provide(data.NewFetcher).
	Provide(data.NewQQHashes).
	Provide(data.NewGenerator).
	Provide(data.NewPurger).
	Provide(data.NewAuditor).
	Provide(biz.NewAvatarUsecase).
	Provide(service.NewAvatarService).
	Multibind[registry.Routes]().
	Contribute[registry.Routes](service.AvatarRoutes).
	Multibind[registry.Commands]().
	Contribute[registry.Commands](service.HashCommand).
	Multibind[registry.Jobs]().
	Contribute[registry.Jobs](service.PurgeExpiredCacheJob).
	Multibind[registry.UserCleanups]().
	Contribute[registry.UserCleanups](data.NewUserCleanup).
	Export[*biz.AvatarUsecase]().
	Export[registry.Routes]().
	Export[registry.Commands]().
	Export[registry.Jobs]().
	Export[registry.UserCleanups]()
