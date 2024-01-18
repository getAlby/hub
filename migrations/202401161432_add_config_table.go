package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Create a composite index to improve performance of finding the latest nostr event for an app
var _202401161432_add_config_table = &gormigrate.Migration{
	ID: "202401161432_add_config_table",
	Migrate: func(tx *gorm.DB) error {
		return tx.Exec(`
CREATE TABLE configs (
	id int NOT NULL PRIMARY KEY,
	ln_backend_type TEXT,
	lnd_address TEXT,
	lnd_cert_file TEXT,
	lnd_cert_hex TEXT,
	lnd_macaroon_file TEXT,
	lnd_macaroon_hex TEXT,
	breez_mnemonic TEXT,
	greenlight_invite_code TEXT
);`).Error
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
