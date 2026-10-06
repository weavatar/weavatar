//go:build wireinject

// Package verifycode is the verification code module's assembly.
package verifycode

import (
	"github.com/libtnb/wire"

	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/internal/verifycode/biz"
	"github.com/weavatar/weavatar/internal/verifycode/data"
	"github.com/weavatar/weavatar/internal/verifycode/service"
)

var Module = wire.New().
	Provide(data.NewSMSSender).
	Provide(data.NewMailSender).
	Provide(biz.NewCodeUsecase).
	Provide(service.NewVerifyCodeService).
	Multibind[registry.Routes]().
	Contribute[registry.Routes](service.VerifyCodeRoutes).
	Export[registry.Routes]()
