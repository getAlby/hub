package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Records the ID of the transaction the hub sent to fund a swap in from its
// on-chain balance. This is separate from lockup_tx_id, which is only set once
// the swap provider has seen the transaction.
var _202611191200_swap_deposit_tx_id = &gormigrate.Migration{
	ID: "202611191200_swap_deposit_tx_id",
	Migrate: func(tx *gorm.DB) error {

		err := tx.Exec("ALTER TABLE swaps ADD COLUMN deposit_tx_id TEXT;").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
