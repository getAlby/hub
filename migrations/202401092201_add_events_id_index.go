package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Create a composite index to improve performance of finding the latest nostr event for an app
var _202401092201_add_events_id_index = &gormigrate.Migration{
	ID: "202401092201_add_events_id_index",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec("CREATE INDEX IF NOT EXISTS idx_nostr_events_app_id_and_id ON nostr_events(app_id, id)").Error
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
