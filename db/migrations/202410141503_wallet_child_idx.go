package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202410141503_wallet_child_idx = &gormigrate.Migration{
	ID: "202410141503_wallet_child_idx",
	Migrate: func(tx *gorm.DB) error {

		if err := tx.Exec(`
	ALTER TABLE apps ADD COLUMN wallet_child_idx INTEGER;
	CREATE UNIQUE INDEX idx_wallet_child_idx ON apps (wallet_child_idx);
`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
