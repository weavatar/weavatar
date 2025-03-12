package rule

import (
	"github.com/go-rat/cache"
	"github.com/gookit/validate"
	"github.com/knadh/koanf/v2"
	"gorm.io/gorm"
)

func GlobalRules(conf *koanf.Koanf, db *gorm.DB, cache cache.Cache) {
	validate.AddValidators(validate.M{
		"exists":     NewExists(db).Passes,
		"notExists":  NewNotExists(db).Passes,
		"geetest":    NewGeetest(conf).Passes,
		"verifyCode": NewVerifyCode(cache).Passes,
	})
	validate.AddGlobalMessages(map[string]string{
		"exists":     "{field} 不存在",
		"notExists":  "{field} 已存在",
		"geetest":    "验证码校验失败（更换设备环境或刷新重试）",
		"verifyCode": "{field} 验证码错误",
	})
}
