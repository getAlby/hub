package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// This migration adds boostagram column to transactions
var _202408061737_add_boostagrams_and_use_json = &gormigrate.Migration{
	ID: "202408061737_add_boostagrams_and_use_json",
	Migrate: func(db *gorm.DB) error {
		err := db.Transaction(func(tx *gorm.DB) error {
			return tx.Exec(`
			ALTER TABLE transactions ADD COLUMN boostagram JSONB;
			ALTER TABLE transactions ADD COLUMN metadata_temp JSONB;
			UPDATE transactions SET metadata_temp = "";
			UPDATE transactions SET metadata_temp = json(metadata) where metadata != "";
			ALTER TABLE transactions DROP COLUMN metadata;
			ALTER TABLE transactions RENAME COLUMN metadata_temp TO metadata;
		`).Error
		})

		return err
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
