package migrations

import (
	_ "embed"
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const appLastSettledTransactionMigration = `ALTER TABLE apps ADD COLUMN last_settled_transaction_at {{ .Timestamp }};`

var appLastSettledTransactionMigrationTmpl = template.Must(template.New("appLastSettledTransactionMigration").Parse(appLastSettledTransactionMigration))

var _202604081200_app_last_settled_transaction = &gormigrate.Migration{
	ID: "202604081200_app_last_settled_transaction",
	Migrate: func(tx *gorm.DB) error {

		err := exec(tx, appLastSettledTransactionMigrationTmpl)
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
