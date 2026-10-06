package rule

import (
	"regexp"

	"github.com/libtnb/validator"
)

var _ validator.Rule = (*CNMobile)(nil)

var mobile = regexp.MustCompile(`^1[3-9]\d{9}$`)

// CNMobile passes a mainland China mobile number. Empty values pass;
// presence is required's job.
type CNMobile struct{}

func NewCNMobile() *CNMobile {
	return &CNMobile{}
}

func (r *CNMobile) Signature() string { return "cn_mobile" }

func (r *CNMobile) Message() string { return "{field} 必须是有效的手机号" }

func (r *CNMobile) Passes(f *validator.Field) bool {
	if validator.IsEmptyValue(f.Reflect()) {
		return true
	}
	s, ok := f.Value[string]()
	return ok && mobile.MatchString(s)
}
