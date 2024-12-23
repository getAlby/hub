package migrations

import (
	_ "embed"
	"testing"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Delete the content column from both request and response events table
var _202406061259_delete_content = &gormigrate.Migration{
	ID: "202406061259_remove_content_columns",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec("ALTER TABLE response_events DROP COLUMN content").Error; err != nil {
			return err
		}

		if err := tx.Exec("ALTER TABLE request_events DROP COLUMN content").Error; err != nil {
			return err
		}

		// Cannot run when testing with txdb: VACUUM must be run outside of transaction.
		if !testing.Testing() {
			if err := tx.Exec("VACUUM").Error; err != nil {
				return err
			}
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
