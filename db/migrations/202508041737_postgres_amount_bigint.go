package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202508041737_postgres_amount_bigint = &gormigrate.Migration{
	ID: "202508041737_postgres_amount_bigint",
	Migrate: func(db *gorm.DB) error {

		// sqlite works fine but postgres integers are only
		if db.Dialector.Name() != "postgres" {
			return nil
		}

		if err := db.Transaction(func(tx *gorm.DB) error {

			if err := tx.Exec(`ALTER TABLE transactions
ALTER COLUMN amount_msat TYPE bigint;`).Error; err != nil {
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
