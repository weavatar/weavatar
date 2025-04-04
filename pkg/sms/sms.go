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
		switch conf.MustString("audit.driver") {
		case "aliyun":
			instance = &SMS{
				driver: &Aliyun{
					accessKeyId:     conf.MustString("sms.aliyun.accessKeyId"),
					accessKeySecret: conf.MustString("sms.aliyun.accessKeySecret"),
					signName:        conf.MustString("sms.aliyun.signName"),
					templateCode:    conf.MustString("sms.aliyun.templateCode"),
					expireTime:      conf.MustString("code.expireTime"),
				},
			}
		case "tencent":
			instance = &SMS{
				driver: &Tencent{
					secretId:   conf.MustString("sms.tencent.secretId"),
					secretKey:  conf.MustString("sms.tencent.secretKey"),
					signName:   conf.MustString("sms.tencent.signName"),
					templateId: conf.MustString("sms.tencent.templateId"),
					sdkAppId:   conf.MustString("sms.tencent.sdkAppId"),
					expireTime: conf.MustString("code.expireTime"),
				},
			}
		default:
			panic("failed to initialize sms, unsupported driver: " + conf.MustString("sms.driver"))
		}
	})

	return instance
}

// Send 发送短信
func (s *SMS) Send(phone string, message Message) error {
	return s.driver.Send(phone, message)
}
