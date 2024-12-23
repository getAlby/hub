package db

import (
	"os"
	"testing"

	"github.com/DATA-DOG/go-txdb"
	"gorm.io/gorm"

	"github.com/getAlby/hub/db"
)

const defaultTestDB = "test.db"
const testDriver = "txdb"

func init() {
	if testing.Testing() {
		uri := GetTestDatabaseURI()
		if db.IsPostgresURI(uri) {
			txdb.Register(testDriver, "pgx", uri)
		}
	}
}

func GetTestDatabaseURI() string {
	ret := os.Getenv("TEST_DATABASE_URI")
	if ret == "" {
		// TODO: use in-memory DB, or a temporary file
		return defaultTestDB
	}

	return ret
}

func NewDB() (*gorm.DB, error) {
	uri := GetTestDatabaseURI()
	driverName := ""
	if db.IsPostgresURI(uri) {
		driverName = testDriver
	}

	return db.NewDBWithConfig(&db.Config{
		URI:        uri,
		LogQueries: true,
		DriverName: driverName,
	})
}

func CloseDB(d *gorm.DB) {
	if err := db.Stop(d); err != nil {
		panic("failed to close database: " + err.Error())
	}

	if GetTestDatabaseURI() == defaultTestDB {
		os.Remove(defaultTestDB)
	}
}
