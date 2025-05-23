package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"

	"github.com/weavatar/weavatar/internal/biz"
)

func init() {
	Migrations = append(Migrations, &gormigrate.Migration{
		ID: "20250310-init",
		Migrate: func(tx *gorm.DB) error {
			return tx.AutoMigrate(
				&biz.User{},
				&biz.Avatar{},
				&biz.App{},
				&biz.AppAvatar{},
				&biz.Image{},
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return tx.Migrator().DropTable(
				&biz.User{},
				&biz.Avatar{},
				&biz.App{},
				&biz.AppAvatar{},
				&biz.Image{},
			)
		},
	})
}
