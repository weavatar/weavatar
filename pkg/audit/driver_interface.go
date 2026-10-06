package audit

import "context"

type Driver interface {
	Check(ctx context.Context, url string) (bool, string, error)
}
