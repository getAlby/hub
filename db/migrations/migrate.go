package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func Migrate(gormDB *gorm.DB) error {

	m := gormigrate.New(gormDB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		_202401191539_initial_migration,
		_202403171120_delete_ldk_payments,
		_202404021909_nullable_expires_at,
		_202405302121_store_decrypted_request,
		_202406061259_delete_content,
		_202406071726_vacuum,
		_202406301207_rename_request_methods,
		_202407012100_transactions,
		_202407151352_autoincrement,
		_202407201604_transactions_indexes,
		_202407262257_remove_invalid_scopes,
		_202408061737_add_boostagrams_and_use_json,
		_202408191242_transaction_failure_reason,
		_202409021648_channels,
	})

	return m.Migrate()
}
