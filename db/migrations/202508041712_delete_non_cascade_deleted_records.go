package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// NOTE: This was actually 2025-07-08
var _202508041712_delete_non_cascade_deleted_records = &gormigrate.Migration{
	ID: "202508041712_delete_non_cascade_deleted_records",
	Migrate: func(db *gorm.DB) error {

		// the following tables have ON DELETE CASCADE
		// which was not being applied due to PRAGMA foreign_keys = ON;
		// not applying to all DB connections:
		// - app_permissions -> apps
		// - request_events -> apps
		// - response_events -> request_events
		// however we will only delete app permissions.
		// excess request and response events will be removed in batches by a background task.

		if err := db.Exec(`DELETE FROM app_permissions
WHERE app_id NOT IN (SELECT id FROM apps)`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
