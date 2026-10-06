//go:build wireinject

// Package system is the system module's assembly.
package system

import (
	"github.com/libtnb/wire"

	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/internal/system/biz"
	"github.com/weavatar/weavatar/internal/system/data"
	"github.com/weavatar/weavatar/internal/system/service"
)

var Module = wire.New().
	Provide(data.NewUsage).
	Provide(data.NewAvatars).
	Provide(biz.NewSystemUsecase).
	Provide(service.NewSystemService).
	Multibind[registry.Routes]().
	Contribute[registry.Routes](service.SystemRoutes).
	Export[registry.Routes]()
