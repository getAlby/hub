package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// NOTE: This was actually 2025-07-08
var _202508041737_postgres_amount_bigint = &gormigrate.Migration{
	ID: "202508041737_postgres_amount_bigint",
	Migrate: func(db *gorm.DB) error {

		// sqlite works fine but postgres integers are only 4 bytes
		// amounts are in msats (= max ~2.1M sats)
		if db.Dialector.Name() != "postgres" {
			return nil
		}

		if err := db.Exec(`ALTER TABLE transactions
ALTER COLUMN amount_msat TYPE bigint;`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
