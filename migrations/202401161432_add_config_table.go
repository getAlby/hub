package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Create a composite index to improve performance of finding the latest nostr event for an app
var _202401161432_add_config_table = &gormigrate.Migration{
	ID: "202401161432_add_config_table",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
CREATE TABLE config_entries (
	key TEXT PRIMARY KEY,
	value TEXT
);`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
