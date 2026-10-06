package service_test

import (
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/internal/shared/rule"
	"github.com/weavatar/weavatar/internal/verifycode/service"
	"github.com/weavatar/weavatar/pkg/geetest"
)

var ticket = geetest.Ticket{LotNumber: "lot", CaptchaOutput: "out", PassToken: "pass", GenTime: "1"}

// TestCheckRules catches invalid validate tags at test time.
func TestCheckRules(t *testing.T) {
	v := newValidator(t)

	check.NoError(t, v.Check[service.VerifyCodeSms]())
	check.NoError(t, v.Check[service.VerifyCodeEmail]())
}

func TestVerifyCodeSmsRules(t *testing.T) {
	v := newValidator(t)

	cases := []struct {
		name string
		req  service.VerifyCodeSms
		want bool
	}{
		{"valid", service.VerifyCodeSms{Phone: "13800138000", UseFor: "avatar", Captcha: ticket}, true},
		{"missing phone", service.VerifyCodeSms{UseFor: "avatar", Captcha: ticket}, false},
		{"not a mobile", service.VerifyCodeSms{Phone: "12345678901", UseFor: "avatar", Captcha: ticket}, false},
		{"unknown purpose", service.VerifyCodeSms{Phone: "13800138000", UseFor: "login", Captcha: ticket}, false},
		{"missing captcha", service.VerifyCodeSms{Phone: "13800138000", UseFor: "avatar"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := v.Valid(t.Context(), &tc.req)
			must.NoError(t, err)
			check.Equal(t, ok, tc.want)
		})
	}
}

func TestVerifyCodeEmailRules(t *testing.T) {
	v := newValidator(t)

	cases := []struct {
		name string
		req  service.VerifyCodeEmail
		want bool
	}{
		{"valid", service.VerifyCodeEmail{Email: "a@weavatar.com", UseFor: "avatar", Captcha: ticket}, true},
		{"missing email", service.VerifyCodeEmail{UseFor: "avatar", Captcha: ticket}, false},
		{"not an email", service.VerifyCodeEmail{Email: "weavatar.com", UseFor: "avatar", Captcha: ticket}, false},
		{"missing purpose", service.VerifyCodeEmail{Email: "a@weavatar.com", Captcha: ticket}, false},
		{"missing captcha", service.VerifyCodeEmail{Email: "a@weavatar.com", UseFor: "avatar"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := v.Valid(t.Context(), &tc.req)
			must.NoError(t, err)
			check.Equal(t, ok, tc.want)
		})
	}
}

// newValidator mirrors the production validator's rule set and strictness;
// geetest is skipped as with app.debug.
func newValidator(t *testing.T) *validator.Validator {
	t.Helper()

	opts := append([]validator.Option{validator.WithStrictRequired()}, rule.Options(nil, nil, nil, true)...)
	v, err := validator.New(opts...)
	must.NoError(t, err)
	return v
}
