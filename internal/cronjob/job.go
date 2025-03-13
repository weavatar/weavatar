package cronjob

import (
	"log/slog"

	"github.com/google/wire"
	"github.com/robfig/cron/v3"
)

var ProviderSet = wire.NewSet(NewJobs)

type Jobs struct {
	log *slog.Logger
}

func NewJobs(log *slog.Logger) *Jobs {
	return &Jobs{
		log: log,
	}
}

func (r *Jobs) Register(c *cron.Cron) error {
	if _, err := c.AddJob("0 * * * *", NewUpdateExpiredAvatar(r.log)); err != nil {
		return err
	}

	return nil
}
