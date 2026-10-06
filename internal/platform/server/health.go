package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/internal/shared/transport"
)

type healthResult struct {
	name string
	err  error
}

// HealthRoutes serves the probes; they stay out of the OpenAPI docs.
func HealthRoutes(checks registry.HealthChecks) transport.Endpoints {
	return transport.Endpoints{
		{Method: fiber.MethodGet, Path: "/healthz", Handler: func(c fiber.Ctx) error {
			return c.SendString("ok")
		}},
		{Method: fiber.MethodGet, Path: "/readyz", Handler: func(c fiber.Ctx) error {
			if err := checkReadiness(c.Context(), checks, 5*time.Second); err != nil {
				return transport.Error(c, fiber.StatusServiceUnavailable, "%v", err)
			}
			return c.SendString("ok")
		}},
	}
}

// RootRoutes sends visitors of the bare API host to the website.
func RootRoutes(config *conf.Config) transport.Endpoints {
	home := func(c fiber.Ctx) error {
		return c.Redirect().Status(fiber.StatusFound).To("https://" + config.HTTP.Domain)
	}
	return transport.Endpoints{
		{Method: fiber.MethodGet, Path: "/", Handler: home},
		{Method: fiber.MethodGet, Path: "/api", Handler: home},
	}
}

func checkReadiness(parent context.Context, checks registry.HealthChecks, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	results := make(chan healthResult, len(checks))
	for _, check := range checks {
		go func() {
			results <- healthResult{name: check.Name, err: check.Check(ctx)}
		}()
	}

	for range checks {
		select {
		case result := <-results:
			if result.err != nil {
				cancel()
				return fmt.Errorf("%s unavailable", result.name)
			}
		case <-ctx.Done():
			return errors.New("readiness checks timed out")
		}
	}
	return nil
}
