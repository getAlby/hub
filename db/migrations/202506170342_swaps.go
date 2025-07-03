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
	invoice text,
	send_amount integer,
	receive_amount integer,
	destination_address text,
	refund_address text,
	lockup_address text,
	boltz_pubkey text,
	preimage text,
	payment_hash text,
	lockup_tx_id text,
	claim_tx_id text,
	auto_swap boolean,
	timeout_block_height integer,
	swap_tree json,
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
