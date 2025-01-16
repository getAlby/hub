package db

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"gorm.io/gorm"

	"github.com/getAlby/hub/db"
)

const defaultTestDB = "test.db"

func GetTestDatabaseURI() string {
	ret := os.Getenv("TEST_DATABASE_URI")
	if ret == "" {
		// TODO: use in-memory DB, or a temporary file
		return defaultTestDB
	}

	return ret
}

func NewDB(t *testing.T) (*gorm.DB, error) {
	return NewDBWithURI(t, GetTestDatabaseURI())
}

func NewDBWithURI(t *testing.T, uri string) (*gorm.DB, error) {
	if db.IsPostgresURI(uri) {
		parsedURI, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to parse postgres DB URI: %w", err)
		}

		var user, password string
		if userInfo := parsedURI.User; userInfo != nil {
			user = userInfo.Username()
			password, _ = userInfo.Password()
		}

		dbName := strings.TrimPrefix(parsedURI.Path, "/")

		config := pgtestdb.Custom(t, pgtestdb.Config{
			DriverName: "pgx",
			Host:       parsedURI.Hostname(),
			Port:       parsedURI.Port(),
			User:       user,
			Password:   password,
			Database:   dbName,
		}, pgtestdb.NoopMigrator{})

		uri = config.URL()
	}

	return db.NewDBWithConfig(&db.Config{
		URI:        uri,
		LogQueries: true,
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