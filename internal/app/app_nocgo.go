//go:build !cgo

package app

import (
	"context"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/gookit/validate"
	"github.com/knadh/koanf/v2"
	"github.com/robfig/cron/v3"

	"github.com/weavatar/weavatar/pkg/queue"
)

type App struct {
	conf     *koanf.Koanf
	router   *fiber.App
	migrator *gormigrate.Gormigrate
	cron     *cron.Cron
	queue    *queue.Queue
}

func NewApp(conf *koanf.Koanf, router *fiber.App, migrator *gormigrate.Gormigrate, cron *cron.Cron, queue *queue.Queue, _ *validate.Validation) *App {
	return &App{
		conf:     conf,
		router:   router,
		migrator: migrator,
		cron:     cron,
		queue:    queue,
	}
}

func (r *App) Run() error {
	// migrate database
	if err := r.migrator.Migrate(); err != nil {
		return err
	}
	fmt.Println("[DB] database migrated")

	// start cron scheduler
	r.cron.Start()
	fmt.Println("[CRON] cron scheduler started")

	// start queue
	r.queue.Run(context.TODO())

	// run http server
	return r.runServer()
}

// runServer run server
func (r *App) runServer() error {
	fmt.Println("[HTTP] listening and serving on", r.conf.MustString("http.address"))
	return r.router.Listen(r.conf.MustString("http.address"), r.listenConfig())
}

func (r *App) listenConfig() fiber.ListenConfig {
	// prefork not support dual stack
	network := fiber.NetworkTCP
	if r.conf.Bool("http.prefork") {
		network = fiber.NetworkTCP4
	}
	return fiber.ListenConfig{
		ListenerNetwork:       network,
		EnablePrefork:         r.conf.Bool("http.prefork"),
		EnablePrintRoutes:     r.conf.Bool("http.debug"),
		DisableStartupMessage: !r.conf.Bool("http.debug"),
	}
}
