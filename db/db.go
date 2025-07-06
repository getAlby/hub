package db

import (
	"fmt"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gorm_logger "gorm.io/gorm/logger"

	"github.com/getAlby/hub/db/migrations"
	sqlite_wrapper "github.com/getAlby/hub/db/sqlite-wrapper"
	"github.com/getAlby/hub/logger"
)

type Config struct {
	URI        string
	LogQueries bool
	DriverName string
}

func NewDB(uri string, logDBQueries bool) (*gorm.DB, error) {
	return NewDBWithConfig(&Config{
		URI:        uri,
		LogQueries: logDBQueries,
		DriverName: "",
	})
}

func NewDBWithConfig(cfg *Config) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		TranslateError: true,
	}
	if cfg.LogQueries {
		gormConfig.Logger = gorm_logger.Default.LogMode(gorm_logger.Info)
	}

	var ret *gorm.DB

	if IsPostgresURI(cfg.URI) {
		pgConfig := postgres.Config{
			DriverName: cfg.DriverName,
			DSN:        cfg.URI,
		}
		var err error
		ret, err = newPostgresDB(pgConfig, gormConfig)
		if err != nil {
			return nil, err
		}
	} else {
		sqliteURI := cfg.URI

		// apply pragma if we're not running the tests
		if !strings.Contains(sqliteURI, "?mode=memory") {
			// see https://github.com/mattn/go-sqlite3?tab=readme-ov-file#connection-string
			// _txlock: avoid SQLITE_BUSY errors with _txlock=immediate
			// _auto_vacuum: properly cleanup disk when deleting records with auto_vacuum=1
			// _busy_timeout: avoid SQLITE_BUSY errors with 5 second lock timeout
			// _journal_mode: enables write-ahead log so that your reads do not block writes and vice-versa.
			// _synchronous: sqlite will sync less frequently and be more performant, still safe to use because of the enabled WAL mode
			// _cache_size: 20MB memory cache
			sqliteURI = sqliteURI + "?_txlock=immediate&_foreign_keys=1&_auto_vacuum=1&_busy_timeout=5000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=-20000"
		}

		driverName := sqlite_wrapper.Sqlite3WrapperDriverName
		if cfg.DriverName != "" {
			driverName = cfg.DriverName
		}

		sqliteConfig := sqlite.Config{
			DriverName: driverName,
			DSN:        sqliteURI,
		}

		var err error
		ret, err = newSqliteDB(sqliteConfig, gormConfig)
		if err != nil {
			return nil, err
		}
	}

	logger.Logger.WithField("db_backend", ret.Dialector.Name()).Debug("loaded database")

	err := migrations.Migrate(ret)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to migrate")
		return nil, err
	}

	return ret, nil
}

func newSqliteDB(sqliteConfig sqlite.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	gormDB, err := gorm.Open(sqlite.New(sqliteConfig), gormConfig)
	if err != nil {
		return nil, err
	}

	return gormDB, nil
}

func newPostgresDB(pgConfig postgres.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	gormDB, err := gorm.Open(postgres.New(pgConfig), gormConfig)
	if err != nil {
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
	logger.Logger.WithField("db_backend", dbBackend).Debug("shutting down database")
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

func IsPostgresURI(uri string) bool {
	return strings.HasPrefix(uri, "postgresql://") ||
		strings.HasPrefix(uri, "postgres://") // Schema used by the "testdb" package.
}
