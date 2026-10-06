// Package biz holds the verification code module's business logic.
package biz

import (
	"context"
	"time"

	"github.com/libtnb/cache"
	"github.com/libtnb/utils/str"
	"github.com/samber/oops"

	"github.com/weavatar/weavatar/internal/shared/appinfo"
)

// cooldown is the minimum gap between two codes for one target and purpose.
const cooldown = time.Minute

// SMSSender delivers a code by text message.
type SMSSender interface {
	Send(ctx context.Context, phone, code string) error
}

// MailSender delivers a code by email.
type MailSender interface {
	Send(ctx context.Context, to, code string) error
}

// CodeUsecase sends verification codes. Codes live in the shared cache under
// "code:<use_for>:<target>", where the verify_code validation rule reads them.
type CodeUsecase struct {
	cache  cache.Cache
	sms    SMSSender
	mail   MailSender
	expire time.Duration
}

func NewCodeUsecase(c cache.Cache, sms SMSSender, mail MailSender, expire appinfo.CodeExpire) *CodeUsecase {
	return &CodeUsecase{
		cache:  c,
		sms:    sms,
		mail:   mail,
		expire: time.Duration(expire),
	}
}

func (uc *CodeUsecase) SendSMS(ctx context.Context, phone, useFor string) error {
	return uc.send(ctx, useFor, phone, uc.sms.Send)
}

func (uc *CodeUsecase) SendEmail(ctx context.Context, email, useFor string) error {
	return uc.send(ctx, useFor, email, uc.mail.Send)
}

func (uc *CodeUsecase) send(
	ctx context.Context,
	useFor, target string,
	deliver func(ctx context.Context, target, code string) error,
) error {
	key := "code:" + useFor + ":" + target

	// claiming the cooldown atomically keeps concurrent requests from both
	// sending; it is released again if delivery fails
	cdKey := key + ":cd"
	if !uc.cache.Add(cdKey, 1, cooldown) {
		return ErrTooFrequent()
	}

	code := str.RandomN(6)
	if err := uc.cache.Put(key, code, uc.expire); err != nil {
		uc.cache.Forget(cdKey)
		return oops.In("verifycode").Wrapf(err, "store code")
	}

	if err := deliver(ctx, target, code); err != nil {
		uc.cache.Forget(cdKey)
		return err
	}

	return nil
}
