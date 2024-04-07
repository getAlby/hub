package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Delete LDK payments that were not migrated to the new LDK format (PaymentKind)
// TODO: delete this sometime in the future (only affects current testers)
var _202404021909_nullable_expires_at = &gormigrate.Migration{
	ID: "202404021909_nullable_expires_at",
	Migrate: func(tx *gorm.DB) error {

		err := tx.Exec(`update app_permissions set expires_at = NULL where expires_at = "0001-01-01 00:00:00+00:00"`).Error

		return err
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
