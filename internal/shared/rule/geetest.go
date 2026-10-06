package rule

import (
	"context"
	"errors"

	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/pkg/geetest"
)

var _ validator.FallibleRule = (*Geetest)(nil)

// Verifier checks a geetest ticket; *geetest.Geetest implements it.
type Verifier interface {
	Verify(ctx context.Context, ticket geetest.Ticket) (bool, error)
}

// Geetest passes when the captcha ticket verifies. Unlike most rules an empty
// ticket fails: the captcha is a security check, not a format.
type Geetest struct {
	verifier Verifier
	skip     bool
}

// NewGeetest builds the rule; skip (app.debug) admits every ticket.
func NewGeetest(verifier Verifier, skip bool) *Geetest {
	return &Geetest{verifier: verifier, skip: skip}
}

func (r *Geetest) Signature() string { return "geetest" }

func (r *Geetest) Message() string {
	return "验证码校验失败（更换设备环境或刷新重试）"
}

func (r *Geetest) Validate(f *validator.Field) (bool, error) {
	if r.skip {
		return true, nil
	}
	ticket, ok := f.Value[geetest.Ticket]()
	if !ok || ticket == (geetest.Ticket{}) {
		return false, nil
	}
	if r.verifier == nil {
		return false, errors.New("rule: no geetest verifier configured")
	}

	// an unreachable geetest service counts as a failed captcha, not an
	// evaluation error, so the client sees the captcha message
	passed, err := r.verifier.Verify(f.Context(), ticket)
	return passed && err == nil, nil
}
