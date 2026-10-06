// Package biz holds the verification code module's business logic.
package biz

import (
	"context"
	"time"

	"github.com/libtnb/cache"
	"github.com/libtnb/utils/str"
	"github.com/samber/oops"

	"github.com/weavatar/weavatar/internal/shared/apperr"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
)

// cooldown spaces two codes for the same purpose and target.
const cooldown = time.Minute

// SMSSender delivers a code by text message.
type SMSSender interface {
	Send(ctx context.Context, phone, code string) error
}

// MailSender delivers a code by email.
type MailSender interface {
	Send(ctx context.Context, to, code string) error
}

// CodeUsecase issues the codes the verify_code rule checks: the code for
// purpose p and target t lives under cache key "code:<p>:<t>".
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
	return uc.send(ctx, phone, useFor, uc.sms.Send)
}

func (uc *CodeUsecase) SendEmail(ctx context.Context, email, useFor string) error {
	return uc.send(ctx, email, useFor, uc.mail.Send)
}

func (uc *CodeUsecase) send(ctx context.Context, target, useFor string, deliver func(ctx context.Context, target, code string) error) error {
	key := "code:" + useFor + ":" + target
	cooldownKey := key + ":cd"
	// Add claims the cooldown atomically, so concurrent requests send once;
	// 422 rather than 400 because the frontend shows 422 as a toast
	if !uc.cache.Add(cooldownKey, true, cooldown) {
		return apperr.Unprocessable("verify_code.too_frequent", "请勿频繁发送验证码").
			In("verifycode").Errorf("code %s is cooling down", useFor)
	}

	code := str.RandomN(6)
	if err := uc.cache.Put(key, code, uc.expire); err != nil {
		uc.cache.Forget(cooldownKey)
		return oops.In("verifycode").Wrapf(err, "store code")
	}
	if err := deliver(ctx, target, code); err != nil {
		// a failed delivery must not lock the user out of retrying
		uc.cache.Forget(cooldownKey)
		return err
	}

	return nil
}
