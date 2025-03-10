package route

import (
	"github.com/gofiber/fiber/v3"

	"github.com/weavatar/weavatar/internal/service"
)

type Http struct {
	user *service.UserService
}

func NewHttp(user *service.UserService) *Http {
	return &Http{
		user: user,
	}
}

func (r *Http) Register(router fiber.Router) {
	router.Get("/", func(c fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	router.Get("/users", r.user.List)
}
