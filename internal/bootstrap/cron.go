package bootstrap

import (
	"log/slog"

	"github.com/knadh/koanf/v2"
	"github.com/robfig/cron/v3"

	"github.com/weavatar/weavatar/internal/cronjob"
	pkgcron "github.com/weavatar/weavatar/pkg/cron"
)

func NewCron(conf *koanf.Koanf, log *slog.Logger, jobs *cronjob.Jobs) (*cron.Cron, error) {
	logger := pkgcron.NewLogger(log, conf.Bool("app.debug"))

	c := cron.New(
		cron.WithParser(cron.NewParser(
			cron.SecondOptional|cron.Minute|cron.Hour|cron.Dom|cron.Month|cron.Dow|cron.Descriptor,
		)),
		cron.WithLogger(logger),
		cron.WithChain(cron.Recover(logger), cron.SkipIfStillRunning(logger)),
	)
	if err := jobs.Register(c); err != nil {
		return nil, err
	}

	return c, nil
}
