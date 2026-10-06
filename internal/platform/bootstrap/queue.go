package bootstrap

import (
	"log/slog"

	"github.com/weavatar/weavatar/pkg/queue"
)

// NewQueue builds the in-process job queue; the app starts and stops it.
func NewQueue(log *slog.Logger) *queue.Queue {
	return queue.New(1000, log)
}
