package biz

import "github.com/weavatar/weavatar/internal/shared/apperr"

func ErrStateExpired() error {
	return apperr.Invalid("user.state_expired", "状态已过期").In("user").New("login state expired")
}

func ErrDeletionIdentityMismatch() error {
	return apperr.Forbidden("user.deletion_identity_mismatch", "授权的账号与当前账号不一致").
		In("user").New("deletion authorized by another account")
}
