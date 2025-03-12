package rule

import (
	"github.com/knadh/koanf/v2"

	"github.com/weavatar/weavatar/pkg/geetest"
)

// Geetest 验证极验票据是否有效
// Geetest verify whether the Geetest ticket is valid
type Geetest struct {
	conf *koanf.Koanf
}

func NewGeetest(conf *koanf.Koanf) *Geetest {
	return &Geetest{conf: conf}
}

func (r *Geetest) Passes(val any, options ...any) bool {
	if r.conf.Bool("app.debug") {
		return true
	}

	ticket, ok := val.(geetest.Ticket)
	if !ok {
		return false
	}

	verify, err := geetest.NewGeetest(r.conf.MustString("geetest.id"), r.conf.MustString("geetest.key")).Verify(ticket)
	if err != nil {
		return false
	}

	return verify
}
