package sms

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
	"time"
)

// A resend within resendWindow suggests the carrier blocked the last message,
// so it goes through Tencent.
const resendWindow = 2 * time.Minute

type Config struct {
	Aliyun  AliyunConfig  `koanf:"aliyun"`
	Tencent TencentConfig `koanf:"tencent"`
}

type AliyunConfig struct {
	AccessKeyID     string `koanf:"access_key_id"`
	AccessKeySecret string `koanf:"access_key_secret"`
	SignName        string `koanf:"sign_name"`
	TemplateCode    string `koanf:"template_code"`
}

type TencentConfig struct {
	SecretID   string `koanf:"secret_id"`
	SecretKey  string `koanf:"secret_key"`
	SignName   string `koanf:"sign_name"`
	TemplateID string `koanf:"template_id"`
	SDKAppID   string `koanf:"sdk_app_id"`
}

type Message struct {
	Data    map[string]string
	Content string
}

type SMS struct {
	aliyun  Driver
	tencent Driver
	now     func() time.Time

	mu   sync.Mutex
	sent map[string]time.Time // phone -> last successful send
}

// New fills codeExpire into the Tencent template as whole minutes.
func New(config Config, codeExpire time.Duration) *SMS {
	return &SMS{
		aliyun: &Aliyun{
			accessKeyId:     config.Aliyun.AccessKeyID,
			accessKeySecret: config.Aliyun.AccessKeySecret,
			signName:        config.Aliyun.SignName,
			templateCode:    config.Aliyun.TemplateCode,
		},
		tencent: &Tencent{
			secretId:   config.Tencent.SecretID,
			secretKey:  config.Tencent.SecretKey,
			signName:   config.Tencent.SignName,
			templateId: config.Tencent.TemplateID,
			sdkAppId:   config.Tencent.SDKAppID,
			expireTime: strconv.Itoa(int(codeExpire.Minutes())),
		},
		now:  time.Now,
		sent: make(map[string]time.Time),
	}
}

// Send tries Aliyun then Tencent, but a resend within resendWindow goes
// straight to Tencent.
func (s *SMS) Send(ctx context.Context, phone string, message Message) error {
	if s.recentlySent(phone) {
		if err := s.tencent.Send(ctx, phone, message); err != nil {
			return err
		}
		s.forget(phone)
		return nil
	}

	if aliErr := s.aliyun.Send(ctx, phone, message); aliErr != nil {
		if err := s.tencent.Send(ctx, phone, message); err != nil {
			return errors.Join(aliErr, err)
		}
	}

	s.remember(phone)
	return nil
}

// recentlySent also drops expired entries, keeping the map bounded.
func (s *SMS) recentlySent(phone string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now()
	for k, at := range s.sent {
		if now.Sub(at) >= resendWindow {
			delete(s.sent, k)
		}
	}

	_, ok := s.sent[phone]
	return ok
}

func (s *SMS) remember(phone string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// the map outlives the request, whose buffers phone may alias
	s.sent[strings.Clone(phone)] = s.now()
}

func (s *SMS) forget(phone string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sent, phone)
}
