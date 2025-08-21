package migrations

import (
	_ "embed"
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const forwardsMigration = `
CREATE TABLE forwards(
	id {{ .AutoincrementPrimaryKey }},
	outbound_amount_forwarded_msat bigint,
	total_fee_earned_msat bigint,
	created_at {{ .Timestamp }},
	updated_at {{ .Timestamp }}
);
`

var forwardsMigrationTmpl = template.Must(template.New("forwardsMigration").Parse(forwardsMigration))

var _202508192137_forwards = &gormigrate.Migration{
	ID: "202508192137_forwards",
	Migrate: func(tx *gorm.DB) error {

		if err := exec(tx, forwardsMigrationTmpl); err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
