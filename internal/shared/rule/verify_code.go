package rule

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"strconv"

	"github.com/libtnb/cache"
	"github.com/libtnb/validator"
)

var _ validator.Rule = (*VerifyCode)(nil)

// VerifyCode passes when the value equals the code sent to a sibling field:
// verify_code:phone,register checks cache key "code:register:<phone>". A third
// argument "true" deletes the code once it matches, so it cannot be replayed;
// the rule is then not side-effect free, use it once per request. An empty
// value fails.
type VerifyCode struct {
	cache cache.Cache
}

func NewVerifyCode(c cache.Cache) *VerifyCode {
	return &VerifyCode{cache: c}
}

func (r *VerifyCode) Signature() string { return "verify_code" }

func (r *VerifyCode) Message() string { return "{field} 验证码错误" }

func (r *VerifyCode) CheckArgs(args []string) error {
	if len(args) < 2 || len(args) > 3 {
		return errors.New("want a sibling field, a purpose and an optional clear flag")
	}
	if args[0] == "" || args[1] == "" {
		return errors.New("sibling field and purpose must not be empty")
	}
	if len(args) == 3 {
		if _, err := strconv.ParseBool(args[2]); err != nil {
			return fmt.Errorf("clear flag %q is not a boolean", args[2])
		}
	}
	return nil
}

func (r *VerifyCode) Passes(f *validator.Field) bool {
	code, ok := f.Value[string]()
	if !ok || code == "" || r.cache == nil {
		return false
	}
	args := f.Attrs()
	target, ok := f.Sibling[string](args[0])
	if !ok || target == "" {
		return false
	}

	key := "code:" + args[1] + ":" + target
	stored := r.cache.GetString(key)
	if stored == "" || subtle.ConstantTimeCompare([]byte(stored), []byte(code)) != 1 {
		return false
	}

	if len(args) == 3 {
		if forget, _ := strconv.ParseBool(args[2]); forget {
			r.cache.Forget(key)
		}
	}
	return true
}
