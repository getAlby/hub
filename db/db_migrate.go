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
// The destination's ID sequences continue from the source's, so that IDs of
// deleted rows (which e.g. transactions may still reference) are not reused.
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

	logger.Logger.Info("resetting sequences...")
	if err := resetSequences(from, tx); err != nil {
		return fmt.Errorf("failed to reset sequences: %w", err)
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

func resetSequences(from, to *gorm.DB) error {
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
		// user_configs has no AUTOINCREMENT in sqlite, so its IDs are reused anyway
		if to.Dialector.Name() == "sqlite" && req.table == "user_configs" {
			continue
		}

		lastID, err := getLastAssignedID(from, req.table, req.seq)
		if err != nil {
			return fmt.Errorf("failed to get last ID of %q in source: %w", req.table, err)
		}
		if lastID == 0 {
			// no ID was assigned yet (and setval() rejects 0)
			continue
		}

		destinationLastID, err := getLastAssignedID(to, req.table, req.seq)
		if err != nil {
			return fmt.Errorf("failed to get last ID of %q in destination: %w", req.table, err)
		}
		lastID = max(lastID, destinationLastID)

		if err := setLastAssignedID(to, req.table, req.seq, lastID); err != nil {
			return fmt.Errorf("failed to reset ID sequence of %q: %w", req.table, err)
		}
	}

	return nil
}

func getLastAssignedID(db *gorm.DB, table string, seq string) (int64, error) {
	var counter int64
	var err error
	switch db.Dialector.Name() {
	case "sqlite":
		err = db.Raw("SELECT COALESCE(MAX(seq), 0) FROM sqlite_sequence WHERE name = ?;", table).Scan(&counter).Error
	case "postgres":
		// last_value is not handed out yet while is_called is false (e.g. on a fresh sequence)
		err = db.Table(seq).Select("CASE WHEN is_called THEN last_value ELSE last_value - 1 END").Scan(&counter).Error
	default:
		return 0, fmt.Errorf("unsupported database: %q", db.Dialector.Name())
	}
	if err != nil {
		return 0, fmt.Errorf("failed to query ID sequence: %w", err)
	}

	var maxID int64
	if err := db.Table(table).Select("COALESCE(MAX(id), 0)").Scan(&maxID).Error; err != nil {
		return 0, fmt.Errorf("failed to query max ID: %w", err)
	}

	return max(counter, maxID), nil
}

func setLastAssignedID(db *gorm.DB, table string, seq string, lastID int64) error {
	switch db.Dialector.Name() {
	case "sqlite":
		// no upsert: sqlite_sequence has no unique constraint and the row may be missing
		result := db.Exec("UPDATE sqlite_sequence SET seq = ? WHERE name = ?;", lastID, table)
		if result.Error != nil {
			return fmt.Errorf("failed to update sqlite_sequence: %w", result.Error)
		}
		if result.RowsAffected == 0 {
			err := db.Exec("INSERT INTO sqlite_sequence (name, seq) VALUES (?, ?);", table, lastID).Error
			if err != nil {
				return fmt.Errorf("failed to insert into sqlite_sequence: %w", err)
			}
		}
	case "postgres":
		if err := db.Exec("SELECT setval(?, ?);", seq, lastID).Error; err != nil {
			return fmt.Errorf("failed to execute setval(): %w", err)
		}
	default:
		return fmt.Errorf("unsupported database: %q", db.Dialector.Name())
	}

	return nil
}
