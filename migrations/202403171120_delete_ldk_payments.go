package migrations

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"

	"database/sql"

	"github.com/getAlby/nostr-wallet-connect/models/config"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Delete LDK payments that were not migrated to the new LDK format (PaymentKind)
func _202403171120_delete_ldk_payments(appConfig *config.AppConfig, logger *logrus.Logger) *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202403171120_delete_ldk_payments",
		Migrate: func(tx *gorm.DB) error {

			ldkDbPath := filepath.Join(appConfig.Workdir, "ldk", "storage", "ldk_node_data.sqlite")
			if _, err := os.Stat(ldkDbPath); errors.Is(err, os.ErrNotExist) {
				logger.Info("No LDK database, skipping migration")
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
			logger.WithFields(logrus.Fields{
				"rowsAffected": rowsAffected,
			}).Info("Removed incompatible payments from LDK database")

			return err
		},
		Rollback: func(tx *gorm.DB) error {
			return nil
		},
	}
}
