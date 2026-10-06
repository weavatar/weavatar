//go:build wireinject

// Package user is the user module's assembly.
package user

import (
	"github.com/libtnb/wire"

	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/internal/user/biz"
	"github.com/weavatar/weavatar/internal/user/data"
	"github.com/weavatar/weavatar/internal/user/service"
)

var Module = wire.New().
	Provide(data.NewUserRepo).
	Provide(data.NewOAuthProvider).
	Provide(data.NewTokens).
	Provide(data.NewTxRunner).
	Provide(biz.NewUserUsecase).
	Provide(biz.NewDeletionUsecase).
	Provide(service.NewUserService).
	Multibind[registry.Routes]().
	Contribute[registry.Routes](service.UserRoutes).
	Export[*biz.UserUsecase]().
	Export[registry.Routes]()
