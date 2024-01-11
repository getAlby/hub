package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		_202309271616_initial_migration,
		_202309271617_fix_preimage_null,
		_202309271618_add_payment_sum_index,
		_202401092201_add_events_id_index,
	})

	return m.Migrate()
}
