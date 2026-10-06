package bootstrap

import (
	"reflect"
	"strings"

	"github.com/go-rio/rio"
	"github.com/libtnb/cache"
	"github.com/libtnb/validator"
	"github.com/libtnb/validator/translations"

	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/internal/shared/rule"
	"github.com/weavatar/weavatar/pkg/geetest"
)

// NewValidator builds the shared validator; strict required also rejects "",
// 0 and false.
func NewValidator(
	config *conf.Config,
	db *rio.DB,
	c cache.Cache,
	captcha *geetest.Geetest,
) (*validator.Validator, error) {
	opts := []validator.Option{
		validator.WithTagNameFunc(fieldName),
		validator.WithStrictRequired(),
	}
	if messages := localeMessages(config.App.Locale); messages != nil {
		opts = append(opts, validator.WithTranslation(messages))
	}
	opts = append(opts, rule.Options(db, c, captcha, config.App.Debug)...)

	return validator.New(opts...)
}

// fieldName reports fields in error messages by the name the client sent.
func fieldName(field reflect.StructField) string {
	for _, tag := range []string{"form", "json", "query", "uri"} {
		if name, _, _ := strings.Cut(field.Tag.Get(tag), ","); name != "" && name != "-" {
			return name
		}
	}
	return field.Name
}

func localeMessages(locale string) map[string]string {
	switch locale {
	case "zh_Hans", "zh_CN":
		return translations.ZhHans()
	case "zh_Hant", "zh_TW":
		return translations.ZhHant()
	case "ja":
		return translations.Ja()
	case "ko":
		return translations.Ko()
	case "es":
		return translations.Es()
	case "ru":
		return translations.Ru()
	default:
		return nil // built-in English
	}
}
