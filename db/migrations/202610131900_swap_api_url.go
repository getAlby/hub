package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Records the swap API a swap was created with, so pending swaps and refunds
// keep using it if the swap provider changes. Existing swaps were created with
// the public Boltz API.
var _202610131900_swap_api_url = &gormigrate.Migration{
	ID: "202610131900_swap_api_url",
	Migrate: func(tx *gorm.DB) error {

		err := tx.Exec("ALTER TABLE swaps ADD COLUMN api_url TEXT NOT NULL DEFAULT 'https://api.boltz.exchange';").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
