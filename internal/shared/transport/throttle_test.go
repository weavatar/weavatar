package transport_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/assert/must"
	"github.com/sethvargo/go-limiter/httplimit"

	"github.com/weavatar/weavatar/internal/shared/transport"
)

func TestThrottleSharesBudgetAcrossEndpoints(t *testing.T) {
	throttle := transport.Throttle(2, time.Minute)
	app := fiber.New()
	ok := func(c fiber.Ctx) error { return c.SendString("ok") }
	app.Post("/sms", throttle, ok)
	app.Post("/email", throttle, ok)

	statuses := make([]int, 0, 3)
	for _, path := range []string{"/sms", "/email", "/sms"} {
		resp, err := app.Test(httptest.NewRequest(fiber.MethodPost, path, nil))
		must.NoError(t, err)
		_ = resp.Body.Close()
		statuses = append(statuses, resp.StatusCode)
		must.NotEmpty(t, resp.Header.Get(httplimit.HeaderRateLimitLimit))
		if resp.StatusCode == fiber.StatusTooManyRequests {
			must.NotEmpty(t, resp.Header.Get(httplimit.HeaderRetryAfter))
		}
	}

	must.DeepEqual(t, statuses, []int{fiber.StatusOK, fiber.StatusOK, fiber.StatusTooManyRequests})
}
