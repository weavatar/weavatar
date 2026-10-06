// Package registry defines the typed collections assembled by wire.
package registry

import (
	"context"

	"github.com/urfave/cli/v3"

	"github.com/weavatar/weavatar/internal/shared/job"
	"github.com/weavatar/weavatar/internal/shared/transport"
)

// Routes collects the modules' endpoint groups.
type Routes []transport.Endpoints

// Commands collects the modules' management commands.
type Commands []*cli.Command

// Jobs collects the modules' scheduled jobs.
type Jobs []job.Fn

// HealthCheck is one named readiness dependency.
type HealthCheck struct {
	Name  string
	Check func(context.Context) error
}

// HealthChecks collects the modules' readiness dependencies.
type HealthChecks []HealthCheck

// UserCleanup removes one module's data about a user inside the account
// deletion transaction; Run must use the ctx it gets.
type UserCleanup struct {
	Name string
	Run  func(ctx context.Context, userID string) error
}

// UserCleanups collects the modules' account deletion cleanups.
type UserCleanups []UserCleanup
