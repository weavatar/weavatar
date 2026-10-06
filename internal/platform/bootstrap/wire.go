//go:build wireinject

package bootstrap

import (
	"log/slog"

	"github.com/go-rio/migrate"
	"github.com/go-rio/rio"
	"github.com/libtnb/cache"
	"github.com/libtnb/utils/jwt"
	"github.com/libtnb/validator"
	"github.com/libtnb/wire"

	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/internal/shared/database"
	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/pkg/audit"
	"github.com/weavatar/weavatar/pkg/cdn"
	"github.com/weavatar/weavatar/pkg/geetest"
	"github.com/weavatar/weavatar/pkg/mail"
	"github.com/weavatar/weavatar/pkg/oauth"
	"github.com/weavatar/weavatar/pkg/qqhash"
	"github.com/weavatar/weavatar/pkg/queue"
	"github.com/weavatar/weavatar/pkg/sms"
)

var Module = wire.New().
	Provide(NewLogger).
	Provide(NewData).
	Provide(ProvideDB).
	Provide(database.NewRunner).
	Provide(NewCache).
	Provide(NewValidator).
	Provide(NewMigrate).
	Provide(NewQueue).
	Provide(NewQqHash).
	Provide(NewJWT).
	Provide(NewCDN).
	Provide(NewAudit).
	Provide(NewSMS).
	Provide(NewMail).
	Provide(NewOAuth).
	Provide(NewGeetest).
	Provide(NewDomain).
	Provide(NewHashDir).
	Provide(NewGravatarURL).
	Provide(NewCodeExpire).
	Provide(NewOAuthClient).
	Multibind[registry.HealthChecks]().
	Contribute[registry.HealthChecks](DatabaseHealthCheck).
	Multibind[registry.Commands]().
	Contribute[registry.Commands](MigrateCommand).
	Export[*slog.Logger]().
	Export[*rio.DB]().
	Export[*database.Runner]().
	Export[*validator.Validator]().
	Export[*migrate.Migrator]().
	Export[cache.Cache]().
	Export[*queue.Queue]().
	Export[*qqhash.Tables]().
	Export[*jwt.JWT]().
	Export[*cdn.Cdn]().
	Export[*audit.Audit]().
	Export[*sms.SMS]().
	Export[*mail.Mail]().
	Export[*oauth.Oauth]().
	Export[*geetest.Geetest]().
	Export[appinfo.Domain]().
	Export[appinfo.HashDir]().
	Export[appinfo.GravatarURL]().
	Export[appinfo.CodeExpire]().
	Export[appinfo.OAuthClient]().
	Export[registry.HealthChecks]().
	Export[registry.Commands]()
