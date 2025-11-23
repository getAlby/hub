package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

var _202505091314_hold_invoices = &gormigrate.Migration{
	ID: "202505091314_hold_invoices",
	Migrate: func(db *gorm.DB) error {

		if err := db.Exec(`
	ALTER TABLE transactions ADD COLUMN hold BOOLEAN;
	ALTER TABLE transactions ADD COLUMN settle_deadline integer;
`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
