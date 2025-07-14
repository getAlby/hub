package migrations

import (
	_ "embed"
	"text/template"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

const appLastUsedMigration = `ALTER TABLE apps ADD COLUMN last_used {{ .Timestamp }};`

var appLastUsedMigrationTmpl = template.Must(template.New("appLastUsedMigration").Parse(appLastUsedMigration))

// NOTE: This was actually 2025-07-14
var _202508041738_app_last_used = &gormigrate.Migration{
	ID: "202508041738_app_last_used",
	Migrate: func(tx *gorm.DB) error {

		err := exec(tx, appLastUsedMigrationTmpl)
		if err != nil {
			return err
		}

		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
