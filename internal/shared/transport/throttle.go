package transport

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
)

// Throttle limits each client IP to tokens requests per interval. Every call
// owns a separate budget: reuse one handler on several endpoints to share it.
func Throttle(tokens uint64, interval time.Duration) fiber.Handler {
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   tokens,
		Interval: interval,
	})
	if err != nil {
		panic(fmt.Errorf("create throttle store: %w", err))
	}

	return func(c fiber.Ctx) error {
		// the store outlives the request, whose buffers the IP may alias
		limit, remaining, reset, ok, err := store.Take(c.Context(), strings.Clone(c.IP()))
		if err != nil {
			return ErrorSystem(c)
		}

		resetTime := time.Unix(0, int64(reset)).UTC().Format(time.RFC1123) //nolint:gosec // nanosecond timestamp fits int64

		c.Set(httplimit.HeaderRateLimitLimit, strconv.FormatUint(limit, 10))
		c.Set(httplimit.HeaderRateLimitRemaining, strconv.FormatUint(remaining, 10))
		c.Set(httplimit.HeaderRateLimitReset, resetTime)

		if !ok {
			c.Set(httplimit.HeaderRetryAfter, resetTime)
			return Error(c, fiber.StatusTooManyRequests, "请求过于频繁")
		}

		return c.Next()
	}
}
