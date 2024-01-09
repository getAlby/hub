package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Create a composite index to improve performance of summing payments in the current budget period
var _202401092201_add_events_id_index = &gormigrate.Migration{
	ID: "202401092201_add_events_id_index",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec("CREATE INDEX idx_nostr_events_app_id_and_id ON nostr_events(app_id, id)").Error
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
