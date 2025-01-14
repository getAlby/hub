package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202405302121_store_decrypted_request = &gormigrate.Migration{
	ID: "_202405302121_store_decrypted_request",
	Migrate: func(tx *gorm.DB) error {

		err := tx.Exec(`ALTER TABLE request_events ADD COLUMN method TEXT;`).Error
		if err != nil {
			return err
		}

		err = tx.Exec(`CREATE INDEX "idx_request_events_method" ON "request_events"("method");`).Error
		if err != nil {
			return err
		}

		err = tx.Exec(`ALTER TABLE request_events ADD COLUMN content_data TEXT`).Error
		return err
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
