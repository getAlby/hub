package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202408291715_app_metadata = &gormigrate.Migration{
	ID: "202408291715_app_metadata",
	Migrate: func(tx *gorm.DB) error {

		if err := tx.Exec(`
	ALTER TABLE apps ADD COLUMN metadata JSON;
`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
