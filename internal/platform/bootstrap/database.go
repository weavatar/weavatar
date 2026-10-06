package bootstrap

import (
	"context"
	"log/slog"

	"github.com/go-rio/postgres"
	"github.com/go-rio/rio"

	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/internal/shared/registry"
)

// Data owns the database handle; modules inject the plain *rio.DB.
type Data struct {
	DB *rio.DB
}

// slogHook logs failed statements always, all statements when debug is on.
type slogHook struct {
	log     *slog.Logger
	verbose bool
}

// NewData opens the PostgreSQL handle; it validates the DSN but does not
// connect, so the first query (or /readyz) surfaces connection errors.
func NewData(config *conf.Config, log *slog.Logger) (*Data, func() error, error) {
	db, err := postgres.Open(config.Database.DSN,
		rio.WithQueryHook(newSlogHook(log, config.Database.Debug)),
	)
	if err != nil {
		return nil, nil, err
	}

	sqlDB := db.Unwrap()
	if config.Database.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(config.Database.MaxOpenConns)
	}
	if config.Database.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(config.Database.MaxIdleConns)
	}
	if config.Database.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(config.Database.ConnMaxLifetime)
	}

	return &Data{DB: db}, db.Close, nil
}

// ProvideDB exposes the plain handle for the data layers.
func ProvideDB(data *Data) *rio.DB {
	return data.DB
}

func DatabaseHealthCheck(data *Data) registry.HealthCheck {
	return registry.HealthCheck{Name: "database", Check: data.HealthCheck}
}

func (d *Data) HealthCheck(ctx context.Context) error {
	return d.DB.Unwrap().PingContext(ctx)
}

func (h slogHook) BeforeQuery(ctx context.Context, _ *rio.QueryEvent) context.Context {
	return ctx
}

func (h slogHook) AfterQuery(ctx context.Context, e *rio.QueryEvent) {
	switch {
	case e.Err != nil:
		h.log.ErrorContext(ctx, "query failed",
			slog.String("op", e.Op),
			slog.String("query", e.Query),
			slog.Duration("elapsed", e.Duration),
			slog.Any("err", e.Err),
		)
	case h.verbose:
		h.log.DebugContext(ctx, "query",
			slog.String("op", e.Op),
			slog.String("query", e.Query),
			slog.Int64("rows", e.RowsAffected),
			slog.Duration("elapsed", e.Duration),
		)
	}
}

func newSlogHook(log *slog.Logger, verbose bool) rio.QueryHook {
	return slogHook{log: log, verbose: verbose}
}
