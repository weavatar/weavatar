package app

import (
	"context"
	_ "expvar" // registers /debug/vars on the default mux
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof" //nolint:gosec // private debug listener only
	"time"

	"github.com/go-rio/migrate"
	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/cron"
	"github.com/libtnb/graceful"

	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/pkg/queue"
)

type App struct {
	conf     *conf.Config
	router   *fiber.App
	migrator *migrate.Migrator
	cron     *cron.Cron
	queue    *queue.Queue
}

// fiberServer adapts *fiber.App to graceful.Server.
type fiberServer struct {
	app  *fiber.App
	conf *conf.Config
}

func NewApp(
	config *conf.Config,
	router *fiber.App,
	migrator *migrate.Migrator,
	scheduler *cron.Cron,
	jobs *queue.Queue,
) *App {
	return &App{
		conf:     config,
		router:   router,
		migrator: migrator,
		cron:     scheduler,
		queue:    jobs,
	}
}

// Run migrates the database and serves until ctx is cancelled; SIGHUP
// hot-upgrades. HTTP stops before the queue so in-flight requests can still
// enqueue.
func (r *App) Run(ctx context.Context) error {
	if err := r.migrator.Up(ctx); err != nil {
		return err
	}
	fmt.Println("[DB] database migrated")

	g := graceful.New(
		graceful.WithUpgrade(),
		graceful.WithShutdownTimeout(30*time.Second),
	)
	// no Handler: serves http.DefaultServeMux, where pprof and expvar register
	if addr := r.conf.HTTP.DebugAddress; addr != "" {
		g.Listen("debug", addr, &http.Server{ReadHeaderTimeout: 10 * time.Second})
	}
	g.Add("queue", r.queue.Start, r.queue.Stop)
	g.Add("cron", r.cron.Start, r.cron.Stop)
	g.Listen("http", r.conf.HTTP.Address, fiberServer{app: r.router, conf: r.conf})

	fmt.Println("[HTTP] listening and serving on", r.conf.HTTP.Address)
	return g.Run(ctx)
}

func (s fiberServer) Serve(ln net.Listener) error {
	return s.app.Listener(ln, fiber.ListenConfig{
		EnablePrintRoutes:     s.conf.HTTP.Debug,
		DisableStartupMessage: !s.conf.HTTP.Debug,
	})
}

func (s fiberServer) Shutdown(ctx context.Context) error {
	return s.app.ShutdownWithContext(ctx)
}
