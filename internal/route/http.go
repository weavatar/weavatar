package route

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/knadh/koanf/v2"

	"github.com/weavatar/weavatar/internal/http/middleware"
	"github.com/weavatar/weavatar/internal/service"
)

type Http struct {
	conf       *koanf.Koanf
	avatar     *service.AvatarService
	verifyCode *service.VerifyCodeService
	user       *service.UserService
	system     *service.SystemService
}

func NewHttp(conf *koanf.Koanf, avatar *service.AvatarService, verifyCode *service.VerifyCodeService, user *service.UserService, system *service.SystemService) *Http {
	return &Http{
		conf:       conf,
		avatar:     avatar,
		verifyCode: verifyCode,
		user:       user,
		system:     system,
	}
}

func (r *Http) Register(router fiber.Router) {
	router.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().Status(fiber.StatusFound).To("https://" + r.conf.MustString("http.domain"))
	})

	router.Get("/avatar", r.avatar.Avatar)
	router.Get("/avatar/:hash", r.avatar.Avatar)

	verifyCode := router.Group("/verify_code")
	verifyCode.Use(middleware.Throttle(5, time.Minute))
	verifyCode.Post("/sms", r.verifyCode.Sms)
	verifyCode.Post("/email", r.verifyCode.Email)

	user := router.Group("/user")
	user.Get("/login", r.user.Login)
	user.Post("/callback", r.user.Callback)
	user.Get("/info", middleware.MustLogin(r.conf), r.user.Get)
	user.Put("/info", middleware.MustLogin(r.conf), r.user.Update)
	user.Post("/logout", middleware.MustLogin(r.conf), r.user.Logout)

	avatars := router.Group("/avatars")
	avatars.Use(middleware.MustLogin(r.conf))
	avatars.Get("/check", r.avatar.Check)
	avatars.Get("/qq", r.avatar.Qq)

	system := router.Group("/system")
	system.Get("count", r.system.Count)
}
