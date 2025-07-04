package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202508041712_delete_non_cascade_deleted_records = &gormigrate.Migration{
	ID: "202508041712_delete_non_cascade_deleted_records",
	Migrate: func(db *gorm.DB) error {

		// the following tables have ON DELETE CASCADE
		// which was not being applied due to PRAGMA foreign_keys = ON;
		// not applying to all DB connections:
		// - app_permissions -> apps
		// - request_events -> apps
		// - response_events -> request_events

		if err := db.Transaction(func(tx *gorm.DB) error {

			if err := tx.Exec(`DELETE FROM app_permissions
WHERE app_id NOT IN (SELECT id FROM apps);`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`DELETE FROM request_events
WHERE app_id NOT IN (SELECT id FROM apps);`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`DELETE FROM response_events
WHERE request_id NOT IN (SELECT id FROM request_events);`).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
