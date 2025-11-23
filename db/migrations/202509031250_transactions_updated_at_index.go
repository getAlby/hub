package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// speeds up this query:
// SELECT * FROM transactions WHERE state = 'SETTLED' OR type = 'outgoing' ORDER BY updated_at DESC LIMIT 20;
var _202509031250_transactions_updated_at_index = &gormigrate.Migration{
	ID: "202509031250_transactions_updated_at_index",
	Migrate: func(tx *gorm.DB) error {

		err := tx.Exec("CREATE INDEX IF NOT EXISTS idx_transactions_updated_at ON transactions(updated_at);").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
