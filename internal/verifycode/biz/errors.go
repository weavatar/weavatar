package biz

import "github.com/weavatar/weavatar/internal/shared/apperr"

// ErrTooFrequent is 422 rather than 400 because the frontend shows 422 as a toast.
func ErrTooFrequent() error {
	return apperr.Unprocessable("verify_code.too_frequent", "请勿频繁发送验证码").In("verifycode").New("code sent within the cooldown")
}
