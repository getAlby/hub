package migrations

import (
	_ "embed"
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

//go:embed 202401191539_initial_migration.sql.tmpl
var initialMigration string

var initialMigrationTmpl = template.Must(template.New("initial_migration").Parse(initialMigration))

// Initial migration
var _202401191539_initial_migration = &gormigrate.Migration{
	ID: "202401191539_initial_migration",
	Migrate: func(tx *gorm.DB) error {
		return exec(tx, initialMigrationTmpl)
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
