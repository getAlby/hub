package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// This migration removes old app permissions for request methods (now we use scopes)
var _202408191242_transaction_failure_reason = &gormigrate.Migration{
	ID: "202408191242_transaction_failure_reason",
	Migrate: func(tx *gorm.DB) error {

		if err := tx.Exec(`
	ALTER TABLE transactions ADD failure_reason text;
`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
