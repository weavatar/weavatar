//go:build wireinject

package app

import (
	"github.com/libtnb/wire"

	"github.com/weavatar/weavatar/internal/avatar"
	"github.com/weavatar/weavatar/internal/migrations"
	"github.com/weavatar/weavatar/internal/platform/bootstrap"
	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/internal/platform/server"
	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/internal/system"
	"github.com/weavatar/weavatar/internal/user"
	"github.com/weavatar/weavatar/internal/verifycode"
)

var ApplicationModule = wire.New().
	Multibind[registry.Routes]().
	Multibind[registry.Commands]().
	Multibind[registry.Jobs]().
	Multibind[registry.HealthChecks]().
	Include(
		bootstrap.Module,
		server.Module,
		user.Module,
		avatar.Module,
		verifycode.Module,
		system.Module,
	).
	Provide(conf.Load).
	Provide(migrations.Collection).
	Provide(server.NewVersion).
	Provide(bootstrap.NewCron).
	Provide(server.NewRouter).
	Provide(newRootCommand).
	Provide(NewApp).
	Provide(NewCli)

var InitializeApp = ApplicationModule.Injector[func(string) (*App, func() error, error)]()

var InitializeCLI = ApplicationModule.Injector[func() (*Cli, func() error, error)]()
