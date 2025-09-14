package bootstrap

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/knadh/koanf/v2"
	"github.com/libtnb/utils/env"
	sloggorm "github.com/orandin/slog-gorm"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/weavatar/weavatar/internal/migration"
)

func NewDB(conf *koanf.Koanf, log *slog.Logger) (*gorm.DB, error) {
	var dsn string
	if env.IsWindows() {
		dsn = fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true&interpolateParams=true&allowAllFiles=%t",
			conf.MustString("database.user"),
			conf.MustString("database.password"),
			conf.MustString("database.host"),
			conf.MustInt("database.port"),
			conf.MustString("database.name"),
			conf.Bool("database.import"),
		)
	} else {
		dsn = fmt.Sprintf(
			"%s:%s@unix(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&multiStatements=true&interpolateParams=true&allowAllFiles=%t",
			conf.MustString("database.user"),
			conf.MustString("database.password"),
			conf.MustString("database.socket"),
			conf.MustString("database.name"),
			conf.Bool("database.import"),
		)
	}
	logOptions := []sloggorm.Option{
		sloggorm.WithHandler(log.Handler()),
	}
	if conf.Bool("database.debug") {
		logOptions = append(logOptions, sloggorm.WithTraceAll())
	}
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                                   sloggorm.New(logOptions...),
		SkipDefaultTransaction:                   true,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(20)
	sqlDB.SetMaxOpenConns(500)
	sqlDB.SetConnMaxLifetime(time.Hour)
	sqlDB.SetConnMaxIdleTime(time.Hour)

	return db, nil
}

func NewMigrate(db *gorm.DB) *gormigrate.Gormigrate {
	return gormigrate.New(db, &gormigrate.Options{
		ValidateUnknownMigrations: true,
	}, migration.Migrations)
}
