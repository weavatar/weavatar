package sms

import (
	"sync"

	"github.com/knadh/koanf/v2"
)

// Message 短信内容
type Message struct {
	Data    map[string]string
	Content string
}

var (
	instance *SMS
	once     sync.Once
)

type SMS struct {
	driver Driver
}

func New(conf *koanf.Koanf) *SMS {
	once.Do(func() {
		instance = &SMS{
			driver: &Tencent{
				SecretId:   conf.MustString("sms.secretId"),
				SecretKey:  conf.MustString("sms.secretKey"),
				SignName:   conf.MustString("sms.signName"),
				TemplateId: conf.MustString("sms.templateId"),
				SdkAppId:   conf.MustString("sms.sdkAppId"),
				ExpireTime: conf.MustString("code.expireTime"),
			},
		}
	})

	return instance
}

// Send 发送短信
func (s *SMS) Send(phone string, message Message) error {
	return s.driver.Send(phone, message)
}
