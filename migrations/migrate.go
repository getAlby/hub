package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) {

	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		_202309271616_initial_migration,
		_202309271617_fix_preimage_null,
		_202309271618_add_payment_sum_index,
	})

	err := m.Migrate()
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	log.Println("Any pending migrations ran successfully")
}