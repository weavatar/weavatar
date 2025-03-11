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
	api := router.Group("/api")
	api.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().Status(fiber.StatusFound).To("https://" + r.conf.MustString("http.domain"))
	})

	api.Get("/avatar", r.avatar.Avatar)
	api.Get("/avatar/:hash", r.avatar.Avatar)

	verifyCode := api.Group("/verify_code")
	verifyCode.Use(middleware.Throttle(5, time.Minute))
	verifyCode.Post("/sms", r.verifyCode.Sms)
	verifyCode.Post("/email", r.verifyCode.Email)

	user := api.Group("/user")
	user.Get("/login", r.user.Login)
	user.Post("/callback", r.user.Callback)
	user.Get("/info", middleware.MustLogin(r.conf), r.user.Get)
	user.Put("/info", middleware.MustLogin(r.conf), r.user.Update)
	user.Post("/logout", middleware.MustLogin(r.conf), r.user.Logout)

	avatars := api.Group("/avatars")
	avatars.Use(middleware.MustLogin(r.conf))
	avatars.Get("/", r.avatar.List)
	avatars.Post("/", r.avatar.Create)
	avatars.Put("/:hash", r.avatar.Update)
	avatars.Delete("/:hash", r.avatar.Delete)
	avatars.Get("/check", r.avatar.Check)
	avatars.Get("/qq", r.avatar.Qq)

	system := api.Group("/system")
	system.Get("count", r.system.Count)
}
