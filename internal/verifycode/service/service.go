// Package service adapts HTTP to the verification code usecase.
package service

import (
	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/internal/shared/transport"
	"github.com/weavatar/weavatar/internal/verifycode/biz"
)

type VerifyCodeService struct {
	code     *biz.CodeUsecase
	validate *validator.Validator
}

func NewVerifyCodeService(code *biz.CodeUsecase, validate *validator.Validator) *VerifyCodeService {
	return &VerifyCodeService{
		code:     code,
		validate: validate,
	}
}

func (r *VerifyCodeService) Sms(c fiber.Ctx) error {
	req, err := transport.Bind[VerifyCodeSms](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	if err = r.code.SendSMS(c.Context(), req.Phone, req.UseFor); err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success[any](c, nil)
}

func (r *VerifyCodeService) Email(c fiber.Ctx) error {
	req, err := transport.Bind[VerifyCodeEmail](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	if err = r.code.SendEmail(c.Context(), req.Email, req.UseFor); err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success[any](c, nil)
}
