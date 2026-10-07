package server

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gofiber/contrib/monitor"
	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/validator"
	"github.com/libtnb/validator/contrib/openapi"

	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/internal/shared/transport"
)

func NewRouter(
	config *conf.Config,
	validate *validator.Validator,
	version Version,
	routes registry.Routes,
) (*fiber.App, error) {
	r := fiber.New(fiber.Config{
		AppName:           config.App.Name,
		BodyLimit:         config.HTTP.BodyLimit << 10,
		ReadBufferSize:    config.HTTP.HeaderLimit,
		ReadTimeout:       config.HTTP.ReadTimeout,
		WriteTimeout:      config.HTTP.WriteTimeout,
		IdleTimeout:       config.HTTP.IdleTimeout,
		ReduceMemoryUsage: config.HTTP.ReduceMemoryUsage,
		// Fiber v3 reads ProxyHeader only from trusted peers, so trust every
		// address: the service is reachable only through nginx.
		ProxyHeader:        config.HTTP.ProxyHeader,
		TrustProxy:         config.HTTP.ProxyHeader != "",
		TrustProxyConfig:   fiber.TrustProxyConfig{Proxies: []string{"0.0.0.0/0", "::/0"}},
		EnableIPValidation: true,
		// every framework-level error (404, 405, 413, panics) leaves as JSON
		ErrorHandler: errorHandler,
		JSONEncoder:  json.Marshal,
		JSONDecoder:  json.Unmarshal,
	})

	for _, handler := range globalMiddlewares(config) {
		r.Use(handler)
	}

	HTTP(routes, r)

	// the live metrics page; Chart.js comes from a mirror reachable from China
	r.Get("/api/monitor", monitor.New(monitor.Config{
		Title:      "WeAvatar Monitor",
		ChartJSURL: "https://fastly.jsdelivr.net/npm/chart.js@2.9/dist/Chart.bundle.min.js",
	}))

	if config.HTTP.Docs {
		spec, err := SpecJSON(config.App.Name, version, validate, routes)
		if err != nil {
			return nil, err
		}
		docs := openapi.DocsHTML(config.App.Name, "/openapi.json")
		r.Get("/openapi.json", func(c fiber.Ctx) error {
			c.Type("json")
			return c.Send(spec)
		})
		r.Get("/docs", func(c fiber.Ctx) error {
			c.Type("html")
			return c.Send(docs)
		})
	}

	return r, nil
}

// errorHandler is the single error exit, answering in the usual envelope;
// 5xx details are logged, not sent.
func errorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := errors.AsType[*fiber.Error](err); ok {
		code = e.Code
	}

	if code >= fiber.StatusInternalServerError {
		slog.ErrorContext(c.Context(), "unhandled error",
			slog.String("method", c.Method()),
			slog.String("path", c.Path()),
			slog.Any("err", err),
		)
		return transport.Error(c, code, "%s", http.StatusText(code))
	}

	return transport.Error(c, code, "%s", err.Error())
}
