package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// NOTE: This was actually 2025-07-21
var _202508041739_response_events_index = &gormigrate.Migration{
	ID: "202508041739_response_events_index",
	Migrate: func(tx *gorm.DB) error {

		err := tx.Exec("CREATE INDEX idx_response_events_request_id ON response_events(request_id);").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
