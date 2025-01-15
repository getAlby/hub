package migrations

import (
	_ "embed"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// VACUUM to finish the update the vacuum mode to auto_vacuum
// See https://sqlite.org/pragma.html
// "The database connection can be changed between full and incremental autovacuum mode at any time.
// However, changing from "none" to "full" or "incremental" can only occur when the database is new
// (no tables have yet been created) or by running the VACUUM command."
var _202406071726_vacuum = &gormigrate.Migration{
	ID: "202406071726_vacuum",
	Migrate: func(tx *gorm.DB) error {
		// Disabled for now: not used.
		// Cannot run when testing with txdb: VACUUM must be run outside of transaction.
		// if !testing.Testing() {
		//	if err := tx.Exec("VACUUM").Error; err != nil {
		//		return err
		//	}
		// }

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
