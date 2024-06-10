package migrations

import (
	"github.com/getAlby/nostr-wallet-connect/config"
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, appConfig *config.AppConfig, logger *logrus.Logger) error {

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		_202401191539_initial_migration,
		_202403171120_delete_ldk_payments(appConfig, logger),
		_202404021909_nullable_expires_at,
		_202405302121_store_decrypted_request,
		_202406061259_delete_content,
		_202406071726_vacuum,
	})

	return m.Migrate()
}
