package migrations

import (
	"fmt"
	"strings"
	"text/template"

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
		_202408291715_app_metadata,
		_202410141503_add_wallet_pubkey,
		_202412212345_fix_types,
		_202504231037_add_indexes,
		_202505091314_hold_invoices,
		_202506170342_swaps,
		_202508041712_delete_non_cascade_deleted_records,
		_202508041737_postgres_amount_bigint,
		_202508041738_app_last_used,
		_202508041739_response_events_index,
	})

	return m.Migrate()
}

type sqlDialectDef struct {
	Timestamp               string
	AutoincrementPrimaryKey string
	DropTableCascade        string
}

var sqlDialectSqlite = sqlDialectDef{
	Timestamp:               "datetime",
	AutoincrementPrimaryKey: "INTEGER PRIMARY KEY AUTOINCREMENT",
	DropTableCascade:        "",
}

var sqlDialectPostgres = sqlDialectDef{
	Timestamp:               "timestamptz",
	AutoincrementPrimaryKey: "SERIAL PRIMARY KEY",
	DropTableCascade:        "CASCADE",
}

func getDialect(tx *gorm.DB) *sqlDialectDef {
	switch tx.Dialector.Name() {
	case "postgres":
		return &sqlDialectPostgres
	default:
		return &sqlDialectSqlite
	}
}

func exec(tx *gorm.DB, templ *template.Template) error {
	dialect := getDialect(tx)
	var buf strings.Builder
	err := templ.Execute(&buf, dialect)
	if err != nil {
		panic(fmt.Sprintf("failed to render SQL template: %v", err))
	}

	return tx.Exec(buf.String()).Error
}
