package migrations

import (
	_ "embed"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

//go:embed initial_migration_postgres.sql
var initialMigrationPostgres string
//go:embed initial_migration_sqlite.sql
var initialMigrationSqlite string

// Initial migration
var _202309271616_initial_migration = &gormigrate.Migration {
	ID: "202309271616_initial_migration",
	Migrate: func(tx *gorm.DB) error {

		// only execute migration if apps table doesn't exist
		err := tx.Exec("Select * from apps").Error;
		if err != nil {
			// find which initial migration should be executed
			var initialMigration string
			if tx.Dialector.Name() == "postgres" {
				initialMigration = initialMigrationPostgres
			} else if tx.Dialector.Name() == "sqlite" {
				initialMigration = initialMigrationSqlite
			} else {
				log.Fatalf("unsupported database type: %s", tx.Dialector.Name())
			}

			err := tx.Exec(initialMigration).Error
			if err != nil {
				return err
			}
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil;
	},
}