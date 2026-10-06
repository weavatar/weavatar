package bootstrap

import (
	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/pkg/audit"
	"github.com/weavatar/weavatar/pkg/cdn"
	"github.com/weavatar/weavatar/pkg/geetest"
	"github.com/weavatar/weavatar/pkg/mail"
	"github.com/weavatar/weavatar/pkg/oauth"
	"github.com/weavatar/weavatar/pkg/sms"
)

func NewCDN(config *conf.Config) (*cdn.Cdn, error) {
	return cdn.New(config.CDN)
}

func NewAudit(config *conf.Config) (*audit.Audit, error) {
	return audit.New(config.Audit)
}

// NewSMS builds the SMS sender; templates show the code expiry.
func NewSMS(config *conf.Config) *sms.SMS {
	return sms.New(config.SMS, config.Code.Expire)
}

func NewMail(config *conf.Config) *mail.Mail {
	return mail.New(config.Mail.Host, config.Mail.Port, config.Mail.User, config.Mail.Password)
}

func NewOAuth(config *conf.Config) *oauth.Oauth {
	return oauth.NewOauth(config.OAuth.ClientID, config.OAuth.ClientSecret, config.OAuth.BaseURL)
}

func NewGeetest(config *conf.Config) *geetest.Geetest {
	return geetest.NewGeetest(config.Geetest.ID, config.Geetest.Key)
}
