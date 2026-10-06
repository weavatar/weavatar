package bootstrap

import (
	"log/slog"

	"github.com/libtnb/cron"
	"github.com/libtnb/cron/wrap"

	"github.com/weavatar/weavatar/internal/shared/registry"
)

// NewCron builds the scheduler with every module's job registered. Specs take
// an optional leading seconds field.
func NewCron(log *slog.Logger, jobs registry.Jobs) (*cron.Cron, error) {
	c, err := cron.New(
		cron.WithLogger(log),
		cron.WithSecondsField(),
		cron.WithChain(wrap.Recover(wrap.WithLogger(log)), wrap.SkipIfRunning()),
	)
	if err != nil {
		return nil, err
	}
	if err := registerJobs(jobs, c); err != nil {
		return nil, err
	}

	return c, nil
}

func registerJobs(jobs registry.Jobs, c *cron.Cron) error {
	for _, apply := range jobs {
		if err := apply(c); err != nil {
			return err
		}
	}

	return nil
}
