package middleware

import (
	"time"

	"github.com/go-rat/utils/jwt"
	"github.com/gofiber/fiber/v3"
	"github.com/knadh/koanf/v2"
)

func MustLogin(conf *koanf.Koanf) fiber.Handler {
	parser := jwt.NewJWT(conf.MustString("app.key"), time.Hour)
	return func(c fiber.Ctx) error {
		token := c.Get("Authorization") // Bearer token
		if len(token) < 7 {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"msg": "未登录",
			})
		}

		claims, err := parser.Parse(token[7:])
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"msg": "未登录",
			})
		}

		if !claims.IsValidAt(time.Now()) {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"msg": "登录已过期",
			})
		}

		fiber.Locals[string](c, "user_id", claims.Subject)

		return c.Next()
	}
}
