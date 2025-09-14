package service

import (
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/knadh/koanf/v2"
	"github.com/libtnb/cache"

	"github.com/weavatar/weavatar/pkg/cdn"
)

type SystemService struct {
	conf  *koanf.Koanf
	cache cache.Cache
	cdn   *cdn.Cdn
}

func NewSystemService(conf *koanf.Koanf, cache cache.Cache) *SystemService {
	return &SystemService{
		conf:  conf,
		cache: cache,
		cdn:   cdn.NewCdn(conf),
	}
}

func (r *SystemService) Count(c fiber.Ctx) error {
	yesterday := carbon.Now(carbon.PRC).SubDay().StartOfDay()
	today := carbon.Now(carbon.PRC).StartOfDay()
	domain := r.conf.MustString("http.domain")

	// 先判断下有没有缓存
	usage := r.cache.GetInt64("cdn:usage", -1)
	if usage != -1 {
		return Success(c, fiber.Map{
			"usage": usage,
		})
	}

	data, err := r.cdn.GetUsage(domain, yesterday, today)
	if err != nil {
		return Success(c, fiber.Map{
			"usage": 0,
		})
	}

	usage = int64(data)
	ct := time.Duration(carbon.Now(carbon.PRC).EndOfDay().Timestamp() - carbon.Now(carbon.PRC).Timestamp() + 7200)
	_ = r.cache.Put("cdn:usage", usage, ct*time.Second)

	return Success(c, fiber.Map{
		"usage": usage,
	})
}
