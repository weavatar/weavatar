//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/weavatar/weavatar/internal/app"
	"github.com/weavatar/weavatar/internal/bootstrap"
	"github.com/weavatar/weavatar/internal/data"
	"github.com/weavatar/weavatar/internal/http/middleware"
	"github.com/weavatar/weavatar/internal/route"
	"github.com/weavatar/weavatar/internal/service"
)

// initApp init application.
func initApp() (*app.App, error) {
	panic(wire.Build(bootstrap.ProviderSet, middleware.ProviderSet, route.ProviderSet, service.ProviderSet, data.ProviderSet, app.NewApp))
}
