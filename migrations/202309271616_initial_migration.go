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

var initialMigrations = map[string]string {
	"postgres": initialMigrationPostgres,
	"sqlite": initialMigrationSqlite,
}

// Initial migration
var _202309271616_initial_migration = &gormigrate.Migration {
	ID: "202309271616_initial_migration",
	Migrate: func(tx *gorm.DB) error {
		// only execute migration if apps table doesn't exist
		err := tx.Exec("SELECT * FROM apps").Error;
		if err != nil {
			// find which initial migration should be executed
			initialMigration := initialMigrations[tx.Dialector.Name()]
			if initialMigration == "" {
				log.Fatalf("unsupported database type: %s", tx.Dialector.Name())
			}

			return tx.Exec(initialMigration).Error
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil;
	},
}