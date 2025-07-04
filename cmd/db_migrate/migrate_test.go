package main

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	test_db "github.com/getAlby/hub/tests/db"
)

type testEnvironment struct {
	source *gorm.DB
	dest   *gorm.DB
}

func (e *testEnvironment) cleanup(t *testing.T) {
	err := db.Stop(e.source)
	require.NoError(t, err)

	err = db.Stop(e.dest)
	require.NoError(t, err)
}

func TestSchemaCheck(t *testing.T) {
	type testCase struct {
		name string
		uri  string
	}

	tc := []testCase{
		{
			name: "schema check sqlite",
			uri:  getTestSqliteURI(0),
		},
	}

	if pgUri := getTestPostgresURI(); pgUri != "" {
		tc = append(tc, testCase{
			name: "schema check postgres",
			uri:  pgUri,
		})
	}

	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			dbConn, err := test_db.NewDBWithURI(t, tt.uri)
			require.NoError(t, err)
			defer db.Stop(dbConn)

			err = checkSchema(dbConn)
			require.NoError(t, err)
		})
	}
}

func TestMigrate(t *testing.T) {
	type testCase struct {
		name      string
		sourceURI string
		destURI   string
	}

	// Test migration between sqlite instances for basic sanity checking.
	tc := []testCase{
		{
			name:      "sqlite to sqlite",
			sourceURI: getTestSqliteURI(0),
			destURI:   getTestSqliteURI(1),
		},
	}

	// Only run Postgres tests if Postgres is configured and its URI is set.
	if getTestPostgresURI() != "" {
		tcPg := []testCase{
			{
				name:      "sqlite to postgres",
				sourceURI: getTestSqliteURI(0),
				destURI:   getTestPostgresURI(),
			},
			{
				name:      "postgres to sqlite",
				sourceURI: getTestPostgresURI(),
				destURI:   getTestSqliteURI(0),
			},
		}

		tc = append(tc, tcPg...)
	}

	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			env, err := setupTest(t, tt.sourceURI, tt.destURI)
			require.NoError(t, err)
			defer env.cleanup(t)

			err = migrateDB(env.source, env.dest)
			require.NoError(t, err)
		})
	}
}

func getTestSqliteURI(dbIndex int) string {
	if uri := os.Getenv("TEST_DB_MIGRATE_SQLITE_URI"); uri != "" {
		return uri
	}

	return fmt.Sprintf("file:testmemdb%d?mode=memory&cache=shared&_txlock=immediate&_foreign_keys=1", dbIndex)
}

func getTestPostgresURI() string {
	return os.Getenv("TEST_DB_MIGRATE_POSTGRES_URI")
}

func setupTest(t *testing.T, sourceURI string, destURI string) (*testEnvironment, error) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	source, err := test_db.NewDBWithURI(t, sourceURI)
	if err != nil {
		t.Fatalf("failed to open source database: %v", err)
	}

	dest, err := test_db.NewDBWithURI(t, destURI)
	if err != nil {
		t.Fatalf("failed to open destination database: %v", err)
	}

	insertMockData(t, source)

	return &testEnvironment{
		source: source,
		dest:   dest,
	}, nil
}

func insertMockData(t *testing.T, tx *gorm.DB) {
	baseTime := time.Date(2025, 01, 15, 8, 0, 0, 0, time.UTC)

	userCfg1 := &db.UserConfig{
		Key:       "Relay",
		Value:     "wss://relay.getalby.com/v1",
		Encrypted: false,
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	create(t, tx, userCfg1)

	app1 := &db.App{
		Name:         "test1",
		Description:  "test1 description",
		AppPubkey:    "2b7dea2866958f17c568cf024e113db7a3baa9c253a9016889196b8d0b11c7ae",
		WalletPubkey: ptr("f766024546ddbdc45db6016714047e34117d5e0d68e51fae06ffca9687783995"),
		CreatedAt:    baseTime,
		UpdatedAt:    baseTime,
		Isolated:     false,
		Metadata:     datatypes.JSON("{}"),
	}
	create(t, tx, app1)

	app1Perm := &db.AppPermission{
		App:           *app1,
		Scope:         "pay_invoice",
		MaxAmountSat:  0,
		BudgetRenewal: "monthly",
		ExpiresAt:     nil,
		CreatedAt:     baseTime,
		UpdatedAt:     baseTime,
	}
	create(t, tx, app1Perm)

	app2 := &db.App{
		Name:         "test2",
		Description:  "test2 description",
		AppPubkey:    "560f31e764f7af64719aba1dfdc0bcb3e681d48bb76265ca939622e1a719fe2a",
		WalletPubkey: ptr("b44c5b3e9c3105b9347cce9f4bbfc899df13c591976fe0f706c1aacd4358020b"),
		CreatedAt:    baseTime,
		UpdatedAt:    baseTime,
		Isolated:     false,
		Metadata:     datatypes.JSON("{}"),
	}
	create(t, tx, app2)

	app2Perm := &db.AppPermission{
		App:           *app2,
		Scope:         "get_info",
		MaxAmountSat:  0,
		BudgetRenewal: "monthly",
		ExpiresAt:     nil,
		CreatedAt:     baseTime,
		UpdatedAt:     baseTime,
	}
	create(t, tx, app2Perm)
}

func create[T any](t *testing.T, tx *gorm.DB, v T) *gorm.DB {
	tx.Create(v)
	require.NoError(t, tx.Error)
	return tx
}

func ptr[T any](v T) *T {
	return &v
}
