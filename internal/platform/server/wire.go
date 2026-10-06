//go:build wireinject

package server

import (
	"github.com/libtnb/wire"

	"github.com/weavatar/weavatar/internal/shared/registry"
)

var Module = wire.New().
	Multibind[registry.Routes]().
	Contribute[registry.Routes](HealthRoutes).
	Contribute[registry.Routes](RootRoutes).
	Export[registry.Routes]()
