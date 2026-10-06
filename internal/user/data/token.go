package data

import (
	"github.com/libtnb/utils/jwt"

	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/internal/user/biz"
)

// tokens issues the JWTs that transport.MustLogin accepts.
type tokens struct {
	jwt    *jwt.JWT
	domain string
}

func NewTokens(j *jwt.JWT, domain appinfo.Domain) biz.Tokens {
	return &tokens{jwt: j, domain: string(domain)}
}

func (t *tokens) Issue(userID string) (string, error) {
	return t.jwt.Generate(&jwt.Claims{
		Subject:  userID,
		Audience: []string{t.domain},
		Issuer:   "https://" + t.domain,
	})
}
