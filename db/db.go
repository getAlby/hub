package db

import (
	"fmt"

	"github.com/getAlby/hub/db/migrations"
	"github.com/getAlby/hub/logger"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func NewDB(uri string) (*gorm.DB, error) {
	// var sqlDb *sql.DB
	gormDB, err := gorm.Open(sqlite.Open(uri), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = gormDB.Exec("PRAGMA foreign_keys=ON;").Error
	if err != nil {
		return nil, err
	}
	err = gormDB.Exec("PRAGMA auto_vacuum=FULL;").Error
	if err != nil {
		return nil, err
	}

	// sqlDb, err = DB.DB()
	// if err != nil {
	// 	return err
	// }
	// this causes errors when concurrently saving DB entries and otherwise requires mutexes
	// sqlDb.SetMaxOpenConns(1)

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

	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	return nil
}
