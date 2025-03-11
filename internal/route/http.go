package route

import (
	"github.com/gofiber/fiber/v3"

	"github.com/weavatar/weavatar/internal/service"
)

type Http struct {
	avatar *service.AvatarService
	user   *service.UserService
}

func NewHttp(avatar *service.AvatarService, user *service.UserService) *Http {
	return &Http{
		avatar: avatar,
		user:   user,
	}
}

func (r *Http) Register(router fiber.Router) {
	router.Get("/", func(c fiber.Ctx) error {
		return c.Redirect().Status(fiber.StatusFound).To("https://weavatar.com/")
	})

	router.Get("/avatar", r.avatar.Avatar)
	router.Get("/avatar/:hash", r.avatar.Avatar)

	router.Get("/users", r.user.List)
}
