package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"gorm.io/gorm"

	"github.com/getAlby/hub/db"
)

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
