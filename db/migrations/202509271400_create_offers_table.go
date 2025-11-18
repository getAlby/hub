package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// _202509271400_create_offers_table creates the offers table to store BOLT12 offer details
var _202509271400_create_offers_table = &gormigrate.Migration{
	ID: "202509271400_create_offers_table",
	Migrate: func(tx *gorm.DB) error {
		// Create the offers table
		err := tx.Exec(`
			CREATE TABLE IF NOT EXISTS offers (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				offer_id TEXT NOT NULL UNIQUE,
				offer_string TEXT NOT NULL,
				description TEXT,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);
		`).Error
		if err != nil {
			return err
		}

		// Create indexes
		err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_offers_offer_id ON offers(offer_id);").Error
		if err != nil {
			return err
		}

		err = tx.Exec("CREATE INDEX IF NOT EXISTS idx_offers_created_at ON offers(created_at);").Error
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		err := tx.Exec("DROP INDEX IF EXISTS idx_offers_created_at;").Error
		if err != nil {
			return err
		}

		err = tx.Exec("DROP INDEX IF EXISTS idx_offers_offer_id;").Error
		if err != nil {
			return err
		}

		return tx.Exec("DROP TABLE IF EXISTS offers;").Error
	},
}
