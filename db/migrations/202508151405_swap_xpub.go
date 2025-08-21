package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// NOTE: This was actually 2025-07-21
var _202508151405_swap_xpub = &gormigrate.Migration{
	ID: "202508151405_swap_xpub",
	Migrate: func(tx *gorm.DB) error {

		err := tx.Exec("ALTER TABLE swaps ADD COLUMN used_xpub BOOLEAN;").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
