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

var SerializeTransactions = false

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
		// avoid SQLITE_BUSY errors with _txlock=IMMEDIATE
		if !strings.Contains(sqliteURI, "_txlock=") {
			sqliteURI = sqliteURI + "?_txlock=IMMEDIATE"
		}
		sqliteConfig := sqlite.Config{
			DriverName: cfg.DriverName,
			DSN:        sqliteURI,
		}
		var err error
		ret, err = newSqliteDB(sqliteConfig, gormConfig)
		if err != nil {
			return nil, err
		}
	}

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
	return strings.HasPrefix(uri, "postgresql://")
}

var txSerializer = make(chan struct{}, 1)

func RunTransaction(db *gorm.DB, txFunc func(tx *gorm.DB) error) error {
	if SerializeTransactions {
		txSerializer <- struct{}{}
		defer func() {
			<-txSerializer
		}()
	}
	return db.Transaction(txFunc)
}
