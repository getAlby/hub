package migrations

import (
	_ "embed"
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const appLastSettledTxMigration = `
ALTER TABLE apps ADD COLUMN last_settled_tx_at {{ .Timestamp }};

CREATE INDEX IF NOT EXISTS idx_transactions_migration 
ON transactions (app_id, state, settled_at);

UPDATE apps
SET last_settled_tx_at = (
	SELECT settled_at
	FROM transactions
	WHERE transactions.app_id = apps.id
		AND transactions.state = 'SETTLED'
		AND transactions.settled_at IS NOT NULL
	ORDER BY transactions.settled_at DESC
	LIMIT 1
)
WHERE EXISTS (
  SELECT 1 FROM transactions
	WHERE transactions.app_id = apps.id 
    AND transactions.state = 'SETTLED'
);

DROP INDEX IF EXISTS idx_transactions_migration;
`

var appLastSettledTxMigrationTmpl = template.Must(template.New("appLastSettledTxMigration").Parse(appLastSettledTxMigration))

var _202604081200_app_last_settled_tx = &gormigrate.Migration{
	ID: "202604081200_app_last_settled_tx",
	Migrate: func(tx *gorm.DB) error {

		err := exec(tx, appLastSettledTxMigrationTmpl)
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
