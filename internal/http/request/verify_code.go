package request

import (
	"github.com/weavatar/weavatar/pkg/geetest"
)

type VerifyCodeSms struct {
	Phone  string `json:"phone" form:"phone" validate:"required|isCnMobile"`
	UseFor string `json:"use_for" form:"use_for" validate:"required|in:avatar"`

	Captcha geetest.Ticket `json:"captcha" form:"captcha" validate:"required|geetest"`
}

type VerifyCodeEmail struct {
	Email  string `json:"email" form:"email" validate:"required|email"`
	UseFor string `json:"use_for" form:"use_for" validate:"required|in:avatar"`

	Captcha geetest.Ticket `json:"captcha" form:"captcha" validate:"required|geetest"`
}
