package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202406301207_rename_request_methods = &gormigrate.Migration{
	ID: "202406301207_rename_request_methods",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec("ALTER TABLE app_permissions RENAME request_method TO scope").Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
