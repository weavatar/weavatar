package rule

import (
	"github.com/go-rat/cache"
	"github.com/gookit/validate"
	"github.com/spf13/cast"
)

// VerifyCode 验证码
// VerifyCode verify code
type VerifyCode struct {
	cache cache.Cache
}

func NewVerifyCode(cache cache.Cache) *VerifyCode {
	return &VerifyCode{cache: cache}
}

func (r *VerifyCode) Passes(data validate.DataFace, val any, options ...any) bool {
	if len(options) < 2 {
		return false
	}

	fieldName := options[0].(string) // 字段名称，如 phone
	useFor := options[1].(string)    // 验证码类型，如 register
	needClear := false               // 是否清除验证码
	if len(options) > 2 {
		needClear = cast.ToBool(options[2])
	}

	field, exist := data.Get(fieldName)
	if !exist {
		return false
	}
	if !r.cache.Has("code:" + useFor + ":" + cast.ToString(field)) {
		return false
	}
	if r.cache.Get("code:"+useFor+":"+cast.ToString(field)) != val {
		return false
	}

	if needClear {
		r.cache.Forget("code:" + useFor + ":" + cast.ToString(field))
	}

	return true
}
