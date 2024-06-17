package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// TODO: delete the whole migration
// Delete LDK payments that were not migrated to the new LDK format (PaymentKind)
// this has now been removed as it only affects old test builds that should have been updated by now
// in case someone encounters it, they can run the commands on their DB to fix the issue.
var _202403171120_delete_ldk_payments = &gormigrate.Migration{
	ID: "202403171120_delete_ldk_payments",
	Migrate: func(tx *gorm.DB) error {

		/*ldkDbPath := filepath.Join(appConfig.Workdir, "ldk", "storage", "ldk_node_data.sqlite")
		if _, err := os.Stat(ldkDbPath); errors.Is(err, os.ErrNotExist) {
			logger.Logger.Info("No LDK database, skipping migration")
			return nil
		}
		ldkDb, err := sql.Open("sqlite", ldkDbPath)
		if err != nil {
			return err
		}
		result, err := ldkDb.Exec(`update ldk_node_data set primary_namespace="payments_bkp" where primary_namespace == "payments"`)
		if err != nil {
			return err
		}
		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return err
		}
		logger.Logger.WithFields(logrus.Fields{
			"rowsAffected": rowsAffected,
		}).Info("Removed incompatible payments from LDK database")

		return err*/
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
