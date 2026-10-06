package biz

import "github.com/weavatar/weavatar/internal/shared/apperr"

func ErrStateExpired() error {
	return apperr.Invalid("user.state_expired", "状态已过期").In("user").New("login state expired")
}
