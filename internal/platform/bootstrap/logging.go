// Package bootstrap provides the boot-time infrastructure providers.
package bootstrap

import (
	"context"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/libtnb/logrotate"

	"github.com/weavatar/weavatar/internal/platform/conf"
)

// NewLogger writes to a rotated file, stderr, or both; stdout stays free for
// command output.
func NewLogger(config *conf.Config) (*slog.Logger, func() error, error) {
	var (
		writers []io.Writer
		closer  = func() error { return nil }
	)

	if config.Log.Output == "file" || config.Log.Output == "both" {
		w, err := logrotate.New(config.Log.Path,
			logrotate.WithMaxSize(100*logrotate.MB),
			logrotate.WithRotateEvery(24*time.Hour),
			logrotate.WithMaxBackups(30),
			logrotate.WithMaxAge(30*logrotate.Day),
			logrotate.WithCompress(),
		)
		if err != nil {
			return nil, nil, err
		}
		writers = append(writers, w)
		closer = func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			return w.Shutdown(ctx)
		}
	}
	if config.Log.Output == "stderr" || config.Log.Output == "both" {
		writers = append(writers, os.Stderr)
	}

	log := slog.New(slog.NewJSONHandler(io.MultiWriter(writers...), &slog.HandlerOptions{
		Level: config.Log.SlogLevel(),
	}))
	slog.SetDefault(log)

	return log, closer, nil
}
