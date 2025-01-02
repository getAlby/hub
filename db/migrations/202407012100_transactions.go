package migrations

import (
	_ "embed"
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const transactionsMigration = `
CREATE TABLE transactions(
	id {{ .AutoincrementPrimaryKey }},
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
	created_at {{ .Timestamp }},
	updated_at {{ .Timestamp }},
	expires_at {{ .Timestamp }},
	settled_at {{ .Timestamp }},
	metadata text,
	self_payment boolean
);

DROP TABLE payments;

ALTER TABLE apps ADD isolated boolean;
UPDATE apps set isolated = false;

ALTER TABLE app_permissions RENAME COLUMN max_amount TO max_amount_sat;
`

var transactionsMigrationTmpl = template.Must(template.New("transactionsMigration").Parse(transactionsMigration))

// This migration
// - Replaces the old payments table with a new transactions table
// - Adds new properties to apps
//   - isolated boolean
//
// - Renames max amount on app permissions to be clear its in sats
var _202407012100_transactions = &gormigrate.Migration{
	ID: "202407012100_transactions",
	Migrate: func(tx *gorm.DB) error {

		if err := exec(tx, transactionsMigrationTmpl); err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
