package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// This migration fixes column types for Postgres compatibility.
//
// First, an autoincrement sequence is created for user_configs.id (this works
// in sqlite, because an integer primary key becomes "autoincrement"
// automatically).
//
// Second, user_configs.encrypted is converted into a boolean; otherwise,
// it fails to de/serialize from/to the Go model's `Encrypted bool` field.
// Again, this happens to work in sqlite due to the way booleans are handled
// (they're just an alias for numeric).
var _202412212345_fix_types = &gormigrate.Migration{
	ID: "20241221234500_fix_types",
	Migrate: func(tx *gorm.DB) error {
		if tx.Dialector.Name() == "postgres" {
			err := tx.Exec(`
CREATE SEQUENCE user_configs_id_seq;
ALTER TABLE user_configs ALTER COLUMN id SET DEFAULT nextval('user_configs_id_seq');
ALTER SEQUENCE user_configs_id_seq OWNED BY user_configs.id;
ALTER TABLE user_configs ALTER COLUMN encrypted SET DATA TYPE BOOLEAN USING encrypted::integer::boolean; 
`).Error
			if err != nil {
				return err
			}
		}
		return nil
	},
	Rollback: func(tx *gorm.DB) error {
		return nil
	},
}
