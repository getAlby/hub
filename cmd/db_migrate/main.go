package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"

	"gorm.io/gorm"

	"github.com/getAlby/hub/db"
)

var expectedTables = []string{
	"apps",
	"app_permissions",
	"request_events",
	"response_events",
	"transactions",
	"user_configs",
	"migrations",
}

func main() {
	var fromDSN, toDSN string

	flag.StringVar(&fromDSN, "from", "", "source DSN")
	flag.StringVar(&toDSN, "to", "", "destination DSN")

	flag.Parse()

	if fromDSN == "" || toDSN == "" {
		flag.Usage()
		slog.Error("missing DSN")
		os.Exit(1)
	}

	stopDB := func(d *gorm.DB) {
		if err := db.Stop(d); err != nil {
			slog.Error("failed to close database", "error", err)
		}
	}

	slog.Info("opening source DB...")
	fromDB, err := db.NewDB(fromDSN, false)
	if err != nil {
		slog.Error("failed to open source database", "error", err)
		os.Exit(1)
	}
	defer stopDB(fromDB)

	slog.Info("opening destination DB...")
	toDB, err := db.NewDB(toDSN, false)
	if err != nil {
		slog.Error("failed to open destination database", "error", err)
		os.Exit(1)
	}
	defer stopDB(toDB)

	// Migrations are applied to both the source and the target DB, so
	// schemas should be equal at this point.
	err = checkSchema(fromDB)
	if err != nil {
		slog.Error("database schema check failed; the migration tool may be outdated", "error", err)
		os.Exit(1)
	}

	slog.Info("migrating...")
	err = migrateDB(fromDB, toDB)
	if err != nil {
		slog.Error("failed to migrate database", "error", err)
		os.Exit(1)
	}

	slog.Info("migration complete")
}

func migrateDB(from, to *gorm.DB) error {
	tx := to.Begin()
	defer tx.Rollback()

	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Table migration order matters: referenced tables must be migrated
	// before referencing tables.

	slog.Info("migrating apps...")
	if err := migrateTable[db.App](from, to); err != nil {
		return fmt.Errorf("failed to migrate apps: %w", err)
	}

	slog.Info("migrating app_permissions...")
	if err := migrateTable[db.AppPermission](from, to); err != nil {
		return fmt.Errorf("failed to migrate app_permissions: %w", err)
	}

	slog.Info("migrating request_events...")
	if err := migrateTable[db.RequestEvent](from, to); err != nil {
		return fmt.Errorf("failed to migrate request_events: %w", err)
	}

	slog.Info("migrating response_events...")
	if err := migrateTable[db.ResponseEvent](from, to); err != nil {
		return fmt.Errorf("failed to migrate response_events: %w", err)
	}

	slog.Info("migrating transactions...")
	if err := migrateTable[db.Transaction](from, to); err != nil {
		return fmt.Errorf("failed to migrate transactions: %w", err)
	}

	slog.Info("migrating user_configs...")
	if err := migrateTable[db.UserConfig](from, to); err != nil {
		return fmt.Errorf("failed to migrate user_configs: %w", err)
	}

	if to.Dialector.Name() == "postgres" {
		slog.Info("resetting sequences...")
		if err := resetSequences(to); err != nil {
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

	if err := to.Create(data).Error; err != nil {
		return fmt.Errorf("failed to insert data: %w", err)
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
			slog.Error("failed to close rows", "error", err)
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
