package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202410141503_add_wallet_pubkey = &gormigrate.Migration{
	ID: "202410141503_add_wallet_pubkey",
	Migrate: func(tx *gorm.DB) error {

		if err := tx.Exec(`
	ALTER TABLE apps ADD COLUMN wallet_pubkey TEXT;
`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
