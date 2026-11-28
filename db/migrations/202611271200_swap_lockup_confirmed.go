package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Records that the lockup transaction of a swap in has confirmed, so the
// status screen can show that the swap is now waiting for the lightning payment.
var _202611271200_swap_lockup_confirmed = &gormigrate.Migration{
	ID: "202611271200_swap_lockup_confirmed",
	Migrate: func(tx *gorm.DB) error {

		err := tx.Exec("ALTER TABLE swaps ADD COLUMN lockup_confirmed BOOLEAN;").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
