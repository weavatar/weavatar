package app

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gookit/validate"
	"github.com/knadh/koanf/v2"
	"github.com/robfig/cron/v3"
)

type App struct {
	conf     *koanf.Koanf
	router   *fiber.App
	migrator *gormigrate.Gormigrate
	cron     *cron.Cron
}

func NewApp(conf *koanf.Koanf, router *fiber.App, migrator *gormigrate.Gormigrate, cron *cron.Cron, _ *validate.Validation) *App {
	return &App{
		conf:     conf,
		router:   router,
		migrator: migrator,
		cron:     cron,
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

	// run http server
	return r.runServer()
}

// runServer run server
func (r *App) runServer() error {
	fmt.Println("[HTTP] listening and serving on", r.conf.MustString("http.address"))
	return r.router.Listen(r.conf.MustString("http.address"))
}
