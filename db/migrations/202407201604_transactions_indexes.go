package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// This migration
// - Sets NULL fees to 0 instead
// - Adds indexes that should speed up queries related to transactions
// - (basic indexes)
// - index for budget query
// - index for isolated balance queries
// - index for list_transactions / lookup transaciton queries

var _202407201604_transactions_indexes = &gormigrate.Migration{
	ID: "202407201604_transactions_indexes",
	Migrate: func(db *gorm.DB) error {

		if err := db.Transaction(func(tx *gorm.DB) error {

			// make transaction fees non-nullable
			err := tx.Exec(`
UPDATE transactions set fee_msat = 0 where fee_msat is NULL;
UPDATE transactions set fee_reserve_msat = 0 where fee_reserve_msat is NULL;
		`).Error
			if err != nil {
				return err
			}

			// basic transaction indexes
			err = tx.Exec(`
CREATE INDEX idx_transactions_app_id ON transactions(app_id);
CREATE INDEX idx_transactions_request_event_id ON transactions(request_event_id);
CREATE INDEX idx_transactions_state ON transactions(state);
CREATE INDEX idx_transactions_type ON transactions(type);
CREATE INDEX idx_transactions_payment_hash ON transactions(payment_hash);
CREATE INDEX idx_transactions_created_at ON transactions(created_at);
CREATE INDEX idx_transactions_settled_at ON transactions(settled_at);
		`).Error
			if err != nil {
				return err
			}

			// budgets
			err = tx.Exec(`
CREATE INDEX idx_transactions_app_id_type_state_created_at ON transactions(app_id, type, state, created_at);
		`).Error
			if err != nil {
				return err
			}

			// isolated balance
			err = tx.Exec(`
CREATE INDEX idx_transactions_app_id_type_state ON transactions(app_id, type, state);
		`).Error
			if err != nil {
				return err
			}

			// list transactions / lookup transaction variations
			err = tx.Exec(`
CREATE INDEX idx_transactions_app_id_type_state_created_at_settled_at_payment_hash ON transactions(app_id, type, state, created_at, settled_at, payment_hash);
		`).Error
			if err != nil {
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
