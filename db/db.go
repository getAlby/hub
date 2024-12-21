package db

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"

	"github.com/getAlby/hub/db/migrations"
	"github.com/getAlby/hub/logger"
)

func NewDB(uri string, logDBQueries bool) (*gorm.DB, error) {
	config := &gorm.Config{
		TranslateError: true,
	}
	if logDBQueries {
		config.Logger = gorm_logger.Default.LogMode(gorm_logger.Info)
	}

	if strings.HasPrefix(uri, "postgresql://") {
		return newPostgresDB(uri, config)
	}

	return newSqliteDB(uri, config)
}

func newSqliteDB(uri string, config *gorm.Config) (*gorm.DB, error) {
	// avoid SQLITE_BUSY errors with _txlock=IMMEDIATE
	gormDB, err := gorm.Open(sqlite.Open(uri+"?_txlock=IMMEDIATE"), config)
	if err != nil {
		return nil, err
	}
	err = gormDB.Exec("PRAGMA foreign_keys = ON", nil).Error
	if err != nil {
		return nil, err
	}

	// properly cleanup disk when deleting records
	err = gormDB.Exec("PRAGMA auto_vacuum = FULL", nil).Error
	if err != nil {
		return nil, err
	}

	// avoid SQLITE_BUSY errors with 5 second lock timeout
	err = gormDB.Exec("PRAGMA busy_timeout = 5000", nil).Error
	if err != nil {
		return nil, err
	}

	// enables write-ahead log so that your reads do not block writes and vice-versa.
	err = gormDB.Exec("PRAGMA journal_mode = WAL", nil).Error
	if err != nil {
		return nil, err
	}

	// sqlite will sync less frequently and be more performant, still safe to use because of the enabled WAL mode
	err = gormDB.Exec("PRAGMA synchronous = NORMAL", nil).Error
	if err != nil {
		return nil, err
	}

	// 20MB memory cache
	err = gormDB.Exec("PRAGMA cache_size = -20000", nil).Error
	if err != nil {
		return nil, err
	}

	// moves temporary tables from disk into RAM, speeds up performance a lot
	err = gormDB.Exec("PRAGMA temp_store = memory", nil).Error
	if err != nil {
		return nil, err
	}

	err = migrations.Migrate(gormDB)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to migrate")
		return nil, err
	}

	return gormDB, nil
}

func newPostgresDB(uri string, config *gorm.Config) (*gorm.DB, error) {
	gormDB, err := gorm.Open(postgres.Open(uri), config)
	if err != nil {
		return nil, err
	}

	err = migrations.Migrate(gormDB)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to migrate")
		return nil, err
	}

	return gormDB, nil
}

func Stop(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}

	dbBackend := db.Dialector.Name()

	if dbBackend == "sqlite" {
		err = db.Exec("PRAGMA wal_checkpoint(FULL)", nil).Error
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to execute wal endpoint")
		}
	}

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}
