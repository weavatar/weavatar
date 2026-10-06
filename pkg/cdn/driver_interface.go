package cdn

import (
	"context"
	"time"

	"github.com/imroc/req/v3"
)

const (
	// Layouts some driver APIs require: a numeric zone offset even for UTC.
	iso8601Layout      = "2006-01-02T15:04:05-07:00"
	iso8601MilliLayout = "2006-01-02T15:04:05.999-07:00"

	requestTimeout = 10 * time.Second
)

type Driver interface {
	RefreshUrl(ctx context.Context, urls []string) error
	RefreshPath(ctx context.Context, paths []string) error
	// GetUsage counts domain's requests in [startTime, endTime).
	GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error)
}

// newClient is called once per driver; a req client is safe for concurrent
// requests while its settings stay unchanged.
func newClient() *req.Client {
	return req.C().SetTimeout(requestTimeout)
}
