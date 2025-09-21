package sms

import (
	"sync"
	"time"

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
	cache    = sync.Map{}
)

type SMS struct {
	aliyun  Driver
	tencent Driver
}

func New(conf *koanf.Koanf) *SMS {
	once.Do(func() {
		instance = &SMS{
			aliyun: &Aliyun{
				accessKeyId:     conf.MustString("sms.aliyun.accessKeyId"),
				accessKeySecret: conf.MustString("sms.aliyun.accessKeySecret"),
				signName:        conf.MustString("sms.aliyun.signName"),
				templateCode:    conf.MustString("sms.aliyun.templateCode"),
				expireTime:      conf.MustString("code.expireTime"),
			},
			tencent: &Tencent{
				secretId:   conf.MustString("sms.tencent.secretId"),
				secretKey:  conf.MustString("sms.tencent.secretKey"),
				signName:   conf.MustString("sms.tencent.signName"),
				templateId: conf.MustString("sms.tencent.templateId"),
				sdkAppId:   conf.MustString("sms.tencent.sdkAppId"),
				expireTime: conf.MustString("code.expireTime"),
			},
		}
		// 启动GC定时器
		go startCacheGC()
	})

	return instance
}

// Send 发送短信
func (s *SMS) Send(phone string, message Message) error {
	if _, ok := cache.Load(phone); ok {
		// 缓存中存在 = 2分钟半内发送过 = 可能被傻逼运营商拦截了
		// 直接用阿里云重发
		if err := s.aliyun.Send(phone, message); err != nil {
			return err
		}
		cache.Delete(phone) // 发送后删除缓存
		return nil
	}

	// 首次发送，默认用腾讯云
	if err := s.tencent.Send(phone, message); err != nil {
		// 腾讯云接口报错，用阿里云兜底
		if err = s.aliyun.Send(phone, message); err != nil {
			return err
		}
	}

	cache.Store(phone, time.Now().Unix()) // 记录发送时间
	return nil
}

// startCacheGC 启动缓存垃圾回收定时器
func startCacheGC() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		gcCache()
	}
}

// gcCache 清理过期缓存
func gcCache() {
	now := time.Now().Unix()
	cache.Range(func(key, value any) bool {
		if timestamp, ok := value.(int64); ok {
			// 如果超过2分钟则删除
			if now-timestamp >= 120 {
				cache.Delete(key)
			}
		}
		return true
	})
}
