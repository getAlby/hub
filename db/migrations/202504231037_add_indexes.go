package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202504231037_add_indexes = &gormigrate.Migration{
	ID: "202504231037_add_indexes",
	Migrate: func(db *gorm.DB) error {

		if err := db.Transaction(func(tx *gorm.DB) error {

			if err := tx.Exec(`
DROP INDEX IF EXISTS idx_transactions_request_event_id;
DROP INDEX IF EXISTS idx_transactions_created_at;
DROP INDEX IF EXISTS idx_transactions_settled_at;
DROP INDEX IF EXISTS idx_transactions_app_id_type_state_created_at_settled_at_payment_hash;
		`).Error; err != nil {
				return err
			}

			if err := tx.Exec(`
CREATE INDEX idx_transactions_state_type ON transactions(state, type);
CREATE INDEX idx_transactions_state_type_updated_at ON transactions(state, type, updated_at);
CREATE INDEX idx_transactions_app_id_state_type_updated_at ON transactions(app_id, state, type, updated_at);
CREATE INDEX idx_transactions_payment_hash_settled_at_created_at ON transactions(payment_hash, settled_at, created_at);
			`).Error; err != nil {
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
