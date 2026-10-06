package rule_test

import (
	"context"
	"errors"
	"testing"

	"github.com/libtnb/assert/must"
	"github.com/libtnb/cache"
	"github.com/libtnb/validator"
	"github.com/libtnb/validator/translations"

	"github.com/weavatar/weavatar/internal/shared/rule"
	"github.com/weavatar/weavatar/pkg/geetest"
)

func newValidator(t *testing.T, c cache.Cache, verifier rule.Verifier, skip bool) *validator.Validator {
	t.Helper()
	opts := rule.Options(nil, c, verifier, skip)
	opts = append(opts, validator.WithTranslation(translations.ZhHans()))
	v, err := validator.New(opts...)
	must.NoError(t, err)
	return v
}

// validate runs rules over data and returns the first message, "" on success.
func validate(t *testing.T, v *validator.Validator, data map[string]any, rules map[string]string) (string, error) {
	t.Helper()
	vd, err := v.Map(data, rules)
	must.NoError(t, err)
	err = vd.Validate(t.Context())
	if err == nil {
		return "", nil
	}
	if fields, ok := validator.AsErrors(err); ok {
		return fields.One(), nil
	}
	return "", err
}

type allRules struct {
	Login      string         `json:"login" validate:"required && exists:users,phone,email"`
	NewPhone   string         `json:"new_phone" validate:"required && cn_mobile && not_exists:users,phone"`
	VerifyCode string         `json:"verify_code" validate:"required && verify_code:new_phone,update_phone,true"`
	Captcha    geetest.Ticket `json:"captcha" validate:"required && geetest"`
}

func TestOptionsCompileEveryRule(t *testing.T) {
	v := newValidator(t, nil, nil, true)
	must.NoError(t, v.Check[allRules]())
}

func TestCNMobile(t *testing.T) {
	v := newValidator(t, nil, nil, true)
	for phone, want := range map[string]string{
		"13800138000":    "",
		"19912345678":    "",
		"":               "", // presence is required's job
		"12800138000":    "phone 必须是有效的手机号",
		"1380013800":     "phone 必须是有效的手机号",
		"138001380000":   "phone 必须是有效的手机号",
		"+8613800138000": "phone 必须是有效的手机号",
	} {
		msg, err := validate(t, v, map[string]any{"phone": phone}, map[string]string{"phone": "cn_mobile"})
		must.NoError(t, err)
		must.Equal(t, msg, want, must.Msgf("phone %q", phone))
	}
}

func TestVerifyCode(t *testing.T) {
	c := cache.NewCache()
	v := newValidator(t, c, nil, true)
	data := func(login, code string) map[string]any {
		return map[string]any{"login": login, "verify_code": code}
	}
	keep := map[string]string{"verify_code": "verify_code:login,register"}
	forget := map[string]string{"verify_code": "verify_code:login,register,true"}

	must.NoError(t, c.Put("code:register:user@example.com", "123456", 0))

	msg, err := validate(t, v, data("user@example.com", "123456"), keep)
	must.NoError(t, err)
	must.Equal(t, msg, "")
	must.True(t, c.Has("code:register:user@example.com"), must.Msgf("code kept without the clear flag"))

	for name, tc := range map[string]map[string]any{
		"wrong code":     data("user@example.com", "654321"),
		"empty code":     data("user@example.com", ""),
		"other target":   data("other@example.com", "123456"),
		"empty target":   data("", "123456"),
		"missing target": {"verify_code": "123456"},
	} {
		msg, err := validate(t, v, tc, keep)
		must.NoError(t, err)
		must.Equal(t, msg, "verify_code 验证码错误", must.Msgf("%s", name))
	}

	msg, err = validate(t, v, data("user@example.com", "123456"), map[string]string{"verify_code": "verify_code:login,login"})
	must.NoError(t, err)
	must.Equal(t, msg, "verify_code 验证码错误", must.Msgf("a code is bound to its purpose"))

	msg, err = validate(t, v, data("user@example.com", "123456"), forget)
	must.NoError(t, err)
	must.Equal(t, msg, "")
	must.False(t, c.Has("code:register:user@example.com"), must.Msgf("code cleared after a match"))

	msg, err = validate(t, v, data("user@example.com", "123456"), keep)
	must.NoError(t, err)
	must.Equal(t, msg, "verify_code 验证码错误")
}

func TestVerifyCodeWithoutCacheFails(t *testing.T) {
	v := newValidator(t, nil, nil, true)
	msg, err := validate(t, v, map[string]any{"login": "a", "verify_code": "1"},
		map[string]string{"verify_code": "verify_code:login,login"})
	must.NoError(t, err)
	must.Equal(t, msg, "verify_code 验证码错误")
}

type fakeVerifier struct {
	passed bool
	err    error
	calls  int
}

func (f *fakeVerifier) Verify(context.Context, geetest.Ticket) (bool, error) {
	f.calls++
	return f.passed, f.err
}

func TestGeetest(t *testing.T) {
	ticket := geetest.Ticket{LotNumber: "lot", CaptchaOutput: "out", PassToken: "pass", GenTime: "1"}
	rules := map[string]string{"captcha": "geetest"}
	const failed = "验证码校验失败（更换设备环境或刷新重试）"

	t.Run("skip admits anything", func(t *testing.T) {
		msg, err := validate(t, newValidator(t, nil, nil, true), map[string]any{"captcha": geetest.Ticket{}}, rules)
		must.NoError(t, err)
		must.Equal(t, msg, "")
	})

	t.Run("verified ticket", func(t *testing.T) {
		verifier := &fakeVerifier{passed: true}
		msg, err := validate(t, newValidator(t, nil, verifier, false), map[string]any{"captcha": ticket}, rules)
		must.NoError(t, err)
		must.Equal(t, msg, "")
		must.Equal(t, verifier.calls, 1)
	})

	t.Run("rejected ticket", func(t *testing.T) {
		verifier := &fakeVerifier{passed: false}
		msg, err := validate(t, newValidator(t, nil, verifier, false), map[string]any{"captcha": ticket}, rules)
		must.NoError(t, err)
		must.Equal(t, msg, failed)
	})

	t.Run("network error fails the field", func(t *testing.T) {
		verifier := &fakeVerifier{err: errors.New("timeout")}
		msg, err := validate(t, newValidator(t, nil, verifier, false), map[string]any{"captcha": ticket}, rules)
		must.NoError(t, err)
		must.Equal(t, msg, failed)
	})

	t.Run("empty ticket fails without a call", func(t *testing.T) {
		verifier := &fakeVerifier{passed: true}
		msg, err := validate(t, newValidator(t, nil, verifier, false), map[string]any{"captcha": geetest.Ticket{}}, rules)
		must.NoError(t, err)
		must.Equal(t, msg, failed)
		must.Equal(t, verifier.calls, 0)
	})

	t.Run("missing verifier is an evaluation error", func(t *testing.T) {
		_, err := validate(t, newValidator(t, nil, nil, false), map[string]any{"captcha": ticket}, rules)
		must.ErrorIs(t, err, validator.ErrRuleEvaluation)
	})
}

func TestExistsCheckArgs(t *testing.T) {
	v := newValidator(t, nil, nil, true)
	for _, expr := range []string{
		"exists:users,phone",
		"exists:users,phone,email",
		"not_exists:users,phone",
		"not_exists:users,phone,email",
	} {
		_, err := v.Map(map[string]any{"x": "v"}, map[string]string{"x": expr})
		must.NoError(t, err, must.Msgf("%s", expr))
	}
	for _, expr := range []string{
		"exists",
		"exists:users",
		"exists:Users,phone",
		"exists:users,1phone",
		"not_exists:users",
		"not_exists:users,phone-col",
		"not_exists:users,\"phone OR 1\"",
	} {
		_, err := v.Map(map[string]any{"x": "v"}, map[string]string{"x": expr})
		must.Error(t, err, must.Msgf("%s", expr))
	}
}

func TestExistsSkipsEmptyAndNeedsDatabase(t *testing.T) {
	v := newValidator(t, nil, nil, true)
	for _, expr := range []string{"exists:users,phone", "not_exists:users,email"} {
		msg, err := validate(t, v, map[string]any{"x": ""}, map[string]string{"x": expr})
		must.NoError(t, err)
		must.Equal(t, msg, "")

		_, err = validate(t, v, map[string]any{"x": "value"}, map[string]string{"x": expr})
		must.ErrorIs(t, err, validator.ErrRuleEvaluation)
	}
}

func TestVerifyCodeCheckArgs(t *testing.T) {
	v := newValidator(t, nil, nil, true)
	for expr, ok := range map[string]bool{
		"verify_code:phone,login":        true,
		"verify_code:phone,login,true":   true,
		"verify_code:phone,login,false":  true,
		"verify_code":                    false,
		"verify_code:phone":              false,
		"verify_code:phone,login,maybe":  false,
		"verify_code:phone,login,true,x": false,
	} {
		_, err := v.Map(map[string]any{"x": "v"}, map[string]string{"x": expr})
		if ok {
			must.NoError(t, err, must.Msgf("%s", expr))
		} else {
			must.Error(t, err, must.Msgf("%s", expr))
		}
	}
}
