package transport

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/utils/jwt"
)

// UserIDKey is the fiber.Locals key MustLogin stores the user ID under.
const UserIDKey = "user_id"

// MustLogin admits requests carrying a valid "Authorization: Bearer <jwt>"
// and stores the token subject as the user ID; others get a 401.
func MustLogin(parser *jwt.JWT) fiber.Handler {
	return func(c fiber.Ctx) error {
		token, ok := BearerToken(c.Get(fiber.HeaderAuthorization))
		if !ok {
			return Error(c, fiber.StatusUnauthorized, "未登录")
		}

		claims, err := parser.Parse(token)
		if err != nil {
			// a well-signed token outside its validity window
			if errors.Is(err, jwt.ErrInvalidClaims) {
				return Error(c, fiber.StatusUnauthorized, "登录已过期")
			}
			return Error(c, fiber.StatusUnauthorized, "未登录")
		}
		if claims.Subject == "" {
			return Error(c, fiber.StatusUnauthorized, "未登录")
		}

		fiber.Locals[string](c, UserIDKey, claims.Subject)
		return c.Next()
	}
}

// UserID returns the user ID stored by MustLogin, or "" outside it.
func UserID(c fiber.Ctx) string {
	return fiber.Locals[string](c, UserIDKey)
}

// BearerToken extracts the token from an "Authorization: Bearer <token>"
// header value; the scheme is case-insensitive.
func BearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(header, " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}
	token = strings.TrimSpace(token)
	return token, token != ""
}
