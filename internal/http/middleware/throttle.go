package middleware

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
)

// Throttle 限流器
func Throttle(tokens uint64, interval time.Duration) fiber.Handler {
	store, err := memorystore.New(&memorystore.Config{
		Tokens:   tokens,
		Interval: interval,
	})
	if err != nil {
		log.Fatalf("failed to create throttle memorystore: %v", err)
	}

	return func(c fiber.Ctx) error {
		// Take from the store.
		limit, remaining, reset, ok, err := store.Take(c, c.IP())
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"msg": "系统内部错误",
			})
		}

		resetTime := time.Unix(0, int64(reset)).UTC().Format(time.RFC1123)

		// Set headers (we do this regardless of whether the request is permitted).
		c.Set(httplimit.HeaderRateLimitLimit, strconv.FormatUint(limit, 10))
		c.Set(httplimit.HeaderRateLimitRemaining, strconv.FormatUint(remaining, 10))
		c.Set(httplimit.HeaderRateLimitReset, resetTime)

		// Fail if there were no tokens remaining.
		if !ok {
			c.Set(httplimit.HeaderRetryAfter, resetTime)
			return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
				"msg": "请求过于频繁",
			})
		}

		return c.Next()
	}
}
