package bootstrap

import (
	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
)

func NewDomain(config *conf.Config) appinfo.Domain {
	return appinfo.Domain(config.HTTP.Domain)
}

func NewHashDir(config *conf.Config) appinfo.HashDir {
	return appinfo.HashDir(config.Hash.Dir)
}

func NewGravatarURL(config *conf.Config) appinfo.GravatarURL {
	return appinfo.GravatarURL(config.Gravatar.URL)
}

func NewCodeExpire(config *conf.Config) appinfo.CodeExpire {
	return appinfo.CodeExpire(config.Code.Expire)
}

func NewOAuthClient(config *conf.Config) appinfo.OAuthClient {
	return appinfo.OAuthClient{
		BaseURL:  config.OAuth.BaseURL,
		ClientID: config.OAuth.ClientID,
	}
}
