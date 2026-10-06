package service

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/weavatar/weavatar/internal/shared/transport"
)

func VerifyCodeRoutes(code *VerifyCodeService) transport.Endpoints {
	// one handler on both endpoints: they share the five-per-minute budget
	throttle := []fiber.Handler{transport.Throttle(5, time.Minute)}

	return transport.Endpoints{
		{Method: fiber.MethodPost, Path: "/api/verify_code/sms", Handler: code.Sms, Middlewares: throttle,
			Summary: "发送短信验证码", Tags: []string{"verify_code"},
			Document: transport.Describe[VerifyCodeSms, transport.Envelope[any]](http.StatusOK)},
		{Method: fiber.MethodPost, Path: "/api/verify_code/email", Handler: code.Email, Middlewares: throttle,
			Summary: "发送邮件验证码", Tags: []string{"verify_code"},
			Document: transport.Describe[VerifyCodeEmail, transport.Envelope[any]](http.StatusOK)},
	}
}
