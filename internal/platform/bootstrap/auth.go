package bootstrap

import (
	"time"

	"github.com/libtnb/utils/jwt"

	"github.com/weavatar/weavatar/internal/platform/conf"
)

// NewJWT builds the login token signer and parser; tokens live one hour.
func NewJWT(config *conf.Config) *jwt.JWT {
	return jwt.NewJWT(config.App.Key, time.Hour)
}
