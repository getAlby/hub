package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// This migration
// - Replaces the old payments table with a new transactions table
// - Adds new properties to app_permissions
//   - balance_type string - isolated | full
//   - visibility string - isolated | full
//
// - Renames max amount on app permissions to be clear its in sats
var _202407012100_transactions = &gormigrate.Migration{
	ID: "202407012100_transactions",
	Migrate: func(tx *gorm.DB) error {

		if err := tx.Exec(`
CREATE TABLE transactions(
	id integer,
	app_id integer,
	request_event_id integer,
	type text,
	state text,
	payment_request text,
	preimage text,
	payment_hash text,
	description text,
	description_hash text,
	amount_msat integer,
	fee_msat integer,
	fee_reserve_msat integer,
	created_at datetime,
	updated_at datetime,
	expires_at datetime,
	settled_at datetime,
	metadata text,
	PRIMARY KEY (id)
);

DROP TABLE payments;

ALTER TABLE app_permissions ADD balance_type string;
ALTER TABLE app_permissions ADD visibility string;

UPDATE app_permissions set balance_type = "full";
UPDATE app_permissions set visibility = "full";

ALTER TABLE app_permissions RENAME COLUMN max_amount TO max_amount_sat;

`).Error; err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
