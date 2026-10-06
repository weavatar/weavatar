package server

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/compress"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/etag"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/weavatar/weavatar/internal/platform/conf"
)

// globalMiddlewares leaves out the access log and helmet, whose
// Cross-Origin-Resource-Policy: same-origin would block cross-site embedding.
func globalMiddlewares(config *conf.Config) []fiber.Handler {
	handlers := []fiber.Handler{
		recover.New(recover.Config{
			EnableStackTrace: true,
		}),
	}

	// CORS only when origins are explicitly allowed; empty = same-origin
	if len(config.HTTP.CorsOrigins) > 0 {
		handlers = append(handlers, cors.New(cors.Config{
			AllowOrigins: config.HTTP.CorsOrigins,
		}))
	}

	return append(handlers,
		compress.New(),
		etag.New(),
		requestid.New(),
	)
}
