package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/go-rat/fiber-skeleton/internal/biz"
)

func init() {
	Migrations = append(Migrations, &gormigrate.Migration{
		ID: "20240812-init",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(
				&biz.User{},
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(
				&biz.User{},
			)
		},
	})
}
