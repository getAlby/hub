package db

import (
	"fmt"
	"slices"

	"gorm.io/gorm"

	"github.com/getAlby/hub/logger"
)

var expectedTables = []string{
	"apps",
	"app_permissions",
	"request_events",
	"response_events",
	"transactions",
	"swaps",
	"user_configs",
	"migrations",
	"forwards",
}

// MigrateDB copies all rows from one database to another. Both databases
// must have an up-to-date schema (they are checked against expectedTables).
// Orphaned request and response events are deleted from the source database
// before copying, as they would violate foreign key constraints in the
// destination database.
func MigrateDB(from, to *gorm.DB) error {
	if err := checkSchema(from); err != nil {
		return fmt.Errorf("source database schema check failed: %w", err)
	}

	if err := checkSchema(to); err != nil {
		return fmt.Errorf("destination database schema check failed: %w", err)
	}

	// NOTE: we assume that excess request events have already been cleaned up due to the background task
	// and only a maximum of ~1000 remain.
	logger.Logger.Info("Deleting orphaned request events.")
	err := from.Exec("DELETE FROM request_events WHERE app_id NOT IN (SELECT id FROM apps);").Error
	if err != nil {
		return fmt.Errorf("failed to delete orphaned request events: %w", err)
	}

	// NOTE: we assume that excess response events have already been cleaned up due to the background task
	// and only a maximum of ~1000 remain.
	logger.Logger.Info("Deleting orphaned response events.")
	err = from.Exec("DELETE FROM response_events WHERE request_id NOT IN (SELECT id FROM request_events);").Error
	if err != nil {
		return fmt.Errorf("failed to delete orphaned response events: %w", err)
	}

	tx := to.Begin()
	defer tx.Rollback()

	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Table migration order matters: referenced tables must be migrated
	// before referencing tables.

	logger.Logger.Info("migrating apps...")
	if err := migrateTable[App](from, tx); err != nil {
		return fmt.Errorf("failed to migrate apps: %w", err)
	}

	logger.Logger.Info("migrating app_permissions...")
	if err := migrateTable[AppPermission](from, tx); err != nil {
		return fmt.Errorf("failed to migrate app_permissions: %w", err)
	}

	logger.Logger.Info("migrating request_events...")
	if err := migrateTable[RequestEvent](from, tx); err != nil {
		return fmt.Errorf("failed to migrate request_events: %w", err)
	}

	logger.Logger.Info("migrating response_events...")
	if err := migrateTable[ResponseEvent](from, tx); err != nil {
		return fmt.Errorf("failed to migrate response_events: %w", err)
	}

	logger.Logger.Info("migrating transactions...")
	if err := migrateTable[Transaction](from, tx); err != nil {
		return fmt.Errorf("failed to migrate transactions: %w", err)
	}

	logger.Logger.Info("migrating swaps...")
	if err := migrateTable[Swap](from, tx); err != nil {
		return fmt.Errorf("failed to migrate swaps: %w", err)
	}

	logger.Logger.Info("migrating forwards...")
	if err := migrateTable[Forward](from, tx); err != nil {
		return fmt.Errorf("failed to migrate forwards: %w", err)
	}

	logger.Logger.Info("migrating user_configs...")
	if err := migrateTable[UserConfig](from, tx); err != nil {
		return fmt.Errorf("failed to migrate user_configs: %w", err)
	}

	if to.Dialector.Name() == "postgres" {
		logger.Logger.Info("resetting sequences...")
		if err := resetSequences(tx); err != nil {
			return fmt.Errorf("failed to reset sequences: %w", err)
		}
	}

	tx.Commit()
	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func migrateTable[T any](from, to *gorm.DB) error {
	var data []T
	if err := from.Find(&data).Error; err != nil {
		return fmt.Errorf("failed to fetch data: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	// to avoid "failed to migrate transactions: failed to insert data: extended protocol limited to 65535 parameters"
	// see https://stackoverflow.com/questions/77372430/extended-protocol-limited-to-65535-parameters-golang-gorm
	// max statements is 65535
	// but it's the number of records * columns
	// to be safe, using a lower value of 1000.
	// this will fail if any table has more than 65 columns, which I doubt we will have
	max := 1000
	for i := 0; i < len(data); i += max {
		j := min(i+max, len(data))

		if err := to.Create(data[i:j]).Error; err != nil {
			return fmt.Errorf("failed to insert data: %w", err)
		}
	}

	return nil
}

func checkSchema(db *gorm.DB) error {
	tables, err := listTables(db)
	if err != nil {
		return fmt.Errorf("failed to list database tables: %w", err)
	}

	for _, table := range expectedTables {
		if !slices.Contains(tables, table) {
			return fmt.Errorf("table missing from the database: %q", table)
		}
	}

	for _, table := range tables {
		if !slices.Contains(expectedTables, table) {
			return fmt.Errorf("unexpected table found in the database: %q", table)
		}
	}

	return nil
}

func listTables(db *gorm.DB) ([]string, error) {
	var query string

	switch db.Dialector.Name() {
	case "sqlite":
		query = "SELECT name FROM sqlite_master WHERE type='table'  AND name NOT LIKE 'sqlite_%';"
	case "postgres":
		query = "SELECT tablename FROM pg_tables WHERE schemaname = 'public';"
	default:
		return nil, fmt.Errorf("unsupported database: %q", db.Dialector.Name())
	}

	rows, err := db.Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query table names: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			logger.Logger.WithError(err).Error("failed to close rows")
		}
	}()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

func resetSequences(db *gorm.DB) error {
	type resetReq struct {
		table string
		seq   string
	}

	resetReqs := []resetReq{
		{"apps", "apps_2_id_seq"},
		{"app_permissions", "app_permissions_2_id_seq"},
		{"request_events", "request_events_id_seq"},
		{"response_events", "response_events_id_seq"},
		{"transactions", "transactions_id_seq"},
		{"swaps", "swaps_id_seq"},
		{"forwards", "forwards_id_seq"},
		{"user_configs", "user_configs_id_seq"},
	}

	for _, req := range resetReqs {
		if err := resetPostgresSequence(db, req.table, req.seq); err != nil {
			return fmt.Errorf("failed to reset sequence %q for %q: %w", req.seq, req.table, err)
		}
	}

	return nil
}

func resetPostgresSequence(db *gorm.DB, table string, seq string) error {
	query := fmt.Sprintf("SELECT setval('%s', (SELECT MAX(id) FROM %s));", seq, table)
	if err := db.Exec(query).Error; err != nil {
		return fmt.Errorf("failed to execute setval(): %w", err)
	}

	return nil
}
