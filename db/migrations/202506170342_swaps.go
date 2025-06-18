package migrations

import (
	_ "embed"
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const swapsMigration = `
CREATE TABLE swaps(
	id {{ .AutoincrementPrimaryKey }},
	swap_id text UNIQUE,
	type text,
	state text,
	address text,
	amount_sent integer,
	amount_received integer,
	boltz_pubkey text,
	payment_hash text,
	destination text,
	lockup_tx_id text,
	claim_tx_id text,
	auto_swap boolean,
	created_at {{ .Timestamp }},
	updated_at {{ .Timestamp }}
);
`

var swapsMigrationTmpl = template.Must(template.New("swapsMigration").Parse(swapsMigration))

var _202506170342_swaps = &gormigrate.Migration{
	ID: "202506170342_swaps",
	Migrate: func(tx *gorm.DB) error {

		if err := exec(tx, swapsMigrationTmpl); err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
