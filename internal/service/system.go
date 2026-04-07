package service

import (
	"fmt"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/knadh/koanf/v2"
	"github.com/libtnb/cache"
	"gorm.io/gorm"

	"github.com/weavatar/weavatar/internal/biz"
	"github.com/weavatar/weavatar/pkg/cdn"
)

type SystemService struct {
	conf  *koanf.Koanf
	cache cache.Cache
	cdn   *cdn.Cdn
	db    *gorm.DB
}

func NewSystemService(conf *koanf.Koanf, cache cache.Cache, db *gorm.DB) *SystemService {
	return &SystemService{
		conf:  conf,
		cache: cache,
		cdn:   cdn.NewCdn(conf),
		db:    db,
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

func (r *SystemService) RandomAvatars(c fiber.Ctx) error {
	// 随机查询 30 条有头像的记录
	var avatars []biz.Avatar
	if err := r.db.Order("RAND()").Limit(50).Find(&avatars).Error; err != nil {
		return Success(c, fiber.Map{
			"avatars": []string{},
		})
	}

	domain := r.conf.MustString("http.domain")
	urls := make([]string, 0, len(avatars))
	for _, a := range avatars {
		urls = append(urls, fmt.Sprintf("https://%s/avatar/%s?s=80", domain, a.SHA256))
	}

	return Success(c, fiber.Map{
		"avatars": urls,
	})
}
