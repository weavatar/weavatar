// Package rule holds the custom validator rules.
package rule

import (
	"github.com/go-rio/rio"
	"github.com/libtnb/cache"
	"github.com/libtnb/validator"
)

// Options registers every custom rule. Tests that only compile tags
// (validator.Check) may leave every dependency nil and skip geetest.
func Options(db *rio.DB, c cache.Cache, verifier Verifier, skipGeetest bool) []validator.Option {
	return []validator.Option{
		validator.WithRules(
			NewVerifyCode(c),
			NewCNMobile(),
		),
		validator.WithFallibleRules(
			NewExists(db),
			NewNotExists(db),
			NewGeetest(verifier, skipGeetest),
		),
	}
}
