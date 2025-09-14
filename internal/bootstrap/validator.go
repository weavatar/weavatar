package bootstrap

import (
	"github.com/gookit/validate"
	"github.com/gookit/validate/locales/zhcn"
	"github.com/knadh/koanf/v2"
	"github.com/libtnb/cache"
	"gorm.io/gorm"

	"github.com/weavatar/weavatar/internal/http/rule"
)

// NewValidator just for register global rules
func NewValidator(conf *koanf.Koanf, db *gorm.DB, cache cache.Cache) *validate.Validation {
	zhcn.RegisterGlobal()
	validate.Config(func(opt *validate.GlobalOption) {
		opt.StopOnError = false
		opt.SkipOnEmpty = true
		opt.FieldTag = "form"
	})

	// register global rules
	rule.GlobalRules(conf, db, cache)

	return validate.NewEmpty()
}
