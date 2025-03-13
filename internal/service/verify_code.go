package service

import (
	"time"

	"github.com/go-rat/cache"
	"github.com/go-rat/utils/convert"
	"github.com/go-rat/utils/str"
	"github.com/gofiber/fiber/v3"
	"github.com/knadh/koanf/v2"

	"github.com/weavatar/weavatar/internal/http/request"
	"github.com/weavatar/weavatar/pkg/mail"
	"github.com/weavatar/weavatar/pkg/sms"
)

type VerifyCodeService struct {
	conf  *koanf.Koanf
	cache cache.Cache
}

func NewVerifyCodeService(conf *koanf.Koanf, cache cache.Cache) *VerifyCodeService {
	return &VerifyCodeService{
		conf:  conf,
		cache: cache,
	}
}

// Sms
//
//	@Summary	发送短信验证码
//	@Tags		verify_code
//	@Accept		json
//	@Produce	json
//	@Security	Bearer
//	@Param		data	body		request.VerifyCodeSms	true	"request"
//	@Success	200		{object}	SuccessResponse
//	@Router		/verify_code/sms [post]
func (r *VerifyCodeService) Sms(c fiber.Ctx) error {
	req, err := Bind[request.VerifyCodeSms](c)
	if err != nil {
		return Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	if r.cache.Has("code:" + req.UseFor + ":" + req.Phone) {
		return Error(c, fiber.StatusUnprocessableEntity, "请勿频繁发送验证码")
	}

	code := str.RandomN(6)
	if err = r.cache.Put("code:"+convert.CopyString(req.UseFor)+":"+convert.CopyString(req.Phone), code, r.conf.Duration("code.expireTime")*time.Minute); err != nil {
		return ErrorSystem(c)
	}

	if err = sms.New(r.conf).Send(req.Phone, sms.Message{
		Data: map[string]string{"code": code},
	}); err != nil {
		return ErrorSystem(c)
	}

	return Success(c, nil)
}

// Email
//
//	@Summary	发送邮件验证码
//	@Tags		verify_code
//	@Accept		json
//	@Produce	json
//	@Security	Bearer
//	@Param		data	body		request.VerifyCodeEmail	true	"request"
//	@Success	200		{object}	SuccessResponse
//	@Router		/verify_code/email [post]
func (r *VerifyCodeService) Email(c fiber.Ctx) error {
	req, err := Bind[request.VerifyCodeEmail](c)
	if err != nil {
		return Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	if r.cache.Has("code:" + req.UseFor + ":" + req.Email) {
		return Error(c, fiber.StatusUnprocessableEntity, "请勿频繁发送验证码")
	}

	code := str.RandomN(6)
	if err = r.cache.Put("code:"+convert.CopyString(req.UseFor)+":"+convert.CopyString(req.Email), code, r.conf.Duration("code.expireTime")*time.Minute); err != nil {
		return ErrorSystem(c)
	}

	if err = mail.New(
		r.conf.MustString("mail.host"),
		r.conf.MustInt("mail.port"),
		r.conf.MustString("mail.user"),
		r.conf.MustString("mail.password")).
		Send(req.Email, "验证码", mail.CodeTmpl("WeAvatar", code)); err != nil {
		return Error(c, fiber.StatusInternalServerError, "%v", err)
	}

	return Success(c, nil)
}
