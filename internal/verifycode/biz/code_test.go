package biz_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/cache"
	"github.com/samber/oops"

	mocksbiz "github.com/weavatar/weavatar/internal/mocks/verifycode/biz"
	"github.com/weavatar/weavatar/internal/shared/apperr"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/internal/verifycode/biz"
)

const expire = 5 * time.Minute

var sixDigits = regexp.MustCompile(`^\d{6}$`)

// ttlCache records the lifetime each key is written with.
type ttlCache struct {
	cache.Cache
	ttls map[string]time.Duration
}

func (c *ttlCache) Put(key string, value any, ttl time.Duration) error {
	c.ttls[key] = ttl
	return c.Cache.Put(key, value, ttl)
}

func (c *ttlCache) Add(key string, value any, ttl time.Duration) bool {
	added := c.Cache.Add(key, value, ttl)
	if added {
		c.ttls[key] = ttl
	}
	return added
}

func TestCodeUsecase_SendSMS(t *testing.T) {
	uc, sms, _, c := newUsecase()
	sms.SendFunc = okSender

	must.NoError(t, uc.SendSMS(t.Context(), "13800138000", "avatar"))

	sent := sms.SendCalls()
	must.Len(t, sent, 1)
	check.Equal(t, sent[0].Phone, "13800138000")
	check.True(t, sixDigits.MatchString(sent[0].Code), sent[0].Code)
	// the verify_code rule reads the code from this key
	check.Equal(t, c.GetString("code:avatar:13800138000"), sent[0].Code)
	check.Equal(t, c.ttls["code:avatar:13800138000"], expire)
	check.Equal(t, c.ttls["code:avatar:13800138000:cd"], time.Minute)
}

func TestCodeUsecase_SendEmail(t *testing.T) {
	uc, _, mail, c := newUsecase()
	mail.SendFunc = okSender

	must.NoError(t, uc.SendEmail(t.Context(), "a@weavatar.com", "avatar"))

	sent := mail.SendCalls()
	must.Len(t, sent, 1)
	check.Equal(t, sent[0].To, "a@weavatar.com")
	check.Equal(t, c.GetString("code:avatar:a@weavatar.com"), sent[0].Code)
}

func TestCodeUsecase_Cooldown(t *testing.T) {
	uc, sms, _, _ := newUsecase()
	sms.SendFunc = okSender
	must.NoError(t, uc.SendSMS(t.Context(), "13800138000", "avatar"))

	err := uc.SendSMS(t.Context(), "13800138000", "avatar")

	must.Error(t, err)
	check.Equal(t, apperr.KindOf(err), apperr.KindUnprocessable)
	check.Equal(t, apperr.CodeOf(err), "verify_code.too_frequent")
	check.Equal(t, oops.GetPublic(err, ""), "请勿频繁发送验证码")
	check.Len(t, sms.SendCalls(), 1)

	// the cooldown is per target and per purpose
	must.NoError(t, uc.SendSMS(t.Context(), "13900139000", "avatar"))
	must.NoError(t, uc.SendSMS(t.Context(), "13800138000", "other"))
}

func TestCodeUsecase_FailedDeliveryReleasesCooldown(t *testing.T) {
	uc, sms, _, _ := newUsecase()
	boom := errors.New("gateway down")
	sms.SendFunc = func(context.Context, string, string) error { return boom }

	must.ErrorIs(t, uc.SendSMS(t.Context(), "13800138000", "avatar"), boom)

	sms.SendFunc = okSender
	must.NoError(t, uc.SendSMS(t.Context(), "13800138000", "avatar"))
}

// newUsecase leaves the mock funcs nil, so an unexpected send panics the test.
func newUsecase() (*biz.CodeUsecase, *mocksbiz.SMSSender, *mocksbiz.MailSender, *ttlCache) {
	sms := &mocksbiz.SMSSender{}
	mail := &mocksbiz.MailSender{}
	c := &ttlCache{Cache: cache.NewCache(cache.WithCleanupInterval(0)), ttls: map[string]time.Duration{}}
	return biz.NewCodeUsecase(c, sms, mail, appinfo.CodeExpire(expire)), sms, mail, c
}

func okSender(context.Context, string, string) error { return nil }
