package service

import (
	"github.com/libtnb/cron"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/shared/job"
)

// PurgeExpiredCacheJob drops stale Gravatar/QQ caches hourly.
func PurgeExpiredCacheJob(avatar *biz.AvatarUsecase) job.Fn {
	return func(c *cron.Cron) error {
		_, err := c.Add("0 0 * * * *", cron.JobFunc(avatar.PurgeExpiredCache), cron.WithName("avatar.purge_expired_cache"))
		return err
	}
}
