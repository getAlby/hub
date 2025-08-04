package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/getAlby/hub/db"
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
}

func main() {
	var fromDSN, toDSN string

	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	flag.StringVar(&fromDSN, "from", "", "source DSN")
	flag.StringVar(&toDSN, "to", "", "destination DSN")

	flag.Parse()

	if fromDSN == "" || toDSN == "" {
		flag.Usage()
		logger.Logger.Error("missing DSN")
		os.Exit(1)
	}

	stopDB := func(d *gorm.DB) {
		if err := db.Stop(d); err != nil {
			logger.Logger.WithError(err).Error("failed to close database")
		}
	}

	logger.Logger.Info("opening source DB...")
	fromDB, err := db.NewDB(fromDSN, false)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to open source database")
		os.Exit(1)
	}
	defer stopDB(fromDB)

	logger.Logger.Info("opening destination DB...")
	toDB, err := db.NewDB(toDSN, false)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to open destination database")
		os.Exit(1)
	}
	defer stopDB(toDB)

	// Migrations are applied to both the source and the target DB, so
	// schemas should be equal at this point.
	err = checkSchema(fromDB)
	if err != nil {
		logger.Logger.WithError(err).Error("database schema check failed; the migration tool may be outdated")
		os.Exit(1)
	}

	// Check if VSS is enabled in the source database
	var vssConfig db.UserConfig
	result := fromDB.Where("key = ?", "LdkVssEnabled").First(&vssConfig)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			logger.Logger.Error("LdkVssEnabled config not found in source DB. Migration will not proceed.")
		} else {
			logger.Logger.WithError(result.Error).Error("failed to query LdkVssEnabled config from source DB")
		}
		os.Exit(1)
	}

	if vssConfig.Value != "true" {
		logger.Logger.Error("VSS is not enabled in the source DB (LdkVssEnabled is not 'true'). Migration will not proceed.")
		os.Exit(1)
	}
	logger.Logger.Info("LdkVssEnabled check passed.")

	logger.Logger.Info("migrating...")
	err = migrateDB(fromDB, toDB)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to migrate database")
		os.Exit(1)
	}

	logger.Logger.Info("migration complete")
}

func migrateDB(from, to *gorm.DB) error {
	tx := to.Begin()
	defer tx.Rollback()

	if err := tx.Error; err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	// Table migration order matters: referenced tables must be migrated
	// before referencing tables.

	logger.Logger.Info("migrating apps...")
	if err := migrateTable[db.App](from, tx); err != nil {
		return fmt.Errorf("failed to migrate apps: %w", err)
	}

	logger.Logger.Info("migrating app_permissions...")
	if err := migrateTable[db.AppPermission](from, tx); err != nil {
		return fmt.Errorf("failed to migrate app_permissions: %w", err)
	}

	logger.Logger.Info("migrating request_events...")
	if err := migrateTable[db.RequestEvent](from, tx); err != nil {
		return fmt.Errorf("failed to migrate request_events: %w", err)
	}

	logger.Logger.Info("migrating response_events...")
	if err := migrateTable[db.ResponseEvent](from, tx); err != nil {
		return fmt.Errorf("failed to migrate response_events: %w", err)
	}

	logger.Logger.Info("migrating transactions...")
	if err := migrateTable[db.Transaction](from, tx); err != nil {
		return fmt.Errorf("failed to migrate transactions: %w", err)
	}

	logger.Logger.Info("migrating user_configs...")
	if err := migrateTable[db.UserConfig](from, tx); err != nil {
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
