package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// stores the backend-local payment ID for BOLT-12 offer payments so terminal
// payment events can be matched to the pending transaction across restarts
var _202608281200_transactions_lnclient_payment_id = &gormigrate.Migration{
	ID: "202608281200_transactions_lnclient_payment_id",
	Migrate: func(tx *gorm.DB) error {
		if err := tx.Exec("ALTER TABLE transactions ADD COLUMN ln_client_payment_id TEXT;").Error; err != nil {
			return err
		}
		return tx.Exec("CREATE INDEX IF NOT EXISTS idx_transactions_ln_client_payment_id ON transactions(ln_client_payment_id);").Error
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
