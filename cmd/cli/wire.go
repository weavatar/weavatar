//go:build wireinject

package main

import (
	"github.com/google/wire"

	"github.com/weavatar/weavatar/internal/app"
	"github.com/weavatar/weavatar/internal/bootstrap"
	"github.com/weavatar/weavatar/internal/data"
	"github.com/weavatar/weavatar/internal/route"
	"github.com/weavatar/weavatar/internal/service"
)

// initCli init command line.
func initCli() (*app.Cli, error) {
	panic(wire.Build(bootstrap.ProviderSet, route.ProviderSet, service.ProviderSet, data.ProviderSet, app.NewCli))
}
