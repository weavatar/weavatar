// Package registry defines the typed collections assembled by wire.
package registry

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/weavatar/weavatar/internal/shared/job"
	"github.com/weavatar/weavatar/internal/shared/transport"
)

// Routes contains the endpoint groups contributed by application modules.
type Routes []transport.Endpoints

// Commands contains the management commands contributed by application modules.
type Commands []*cli.Command

// Jobs contains the scheduled jobs contributed by application modules.
type Jobs []job.Fn

// HealthCheck is one named readiness dependency.
type HealthCheck struct {
	Name  string
	Check func(context.Context) error
}

// HealthChecks contains the readiness dependencies contributed by application modules.
type HealthChecks []HealthCheck
