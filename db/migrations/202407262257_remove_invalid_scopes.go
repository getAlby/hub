package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// This migration removes old app permissions for request methods (now we use scopes)
var _202407262257_remove_invalid_scopes = &gormigrate.Migration{
	ID: "202407262257_remove_invalid_scopes",
	Migrate: func(tx *gorm.DB) error {

		if err := tx.Exec(`
delete from app_permissions where scope = 'pay_keysend';
delete from app_permissions where scope = 'multi_pay_keysend';
delete from app_permissions where scope = 'multi_pay_invoice';
`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
