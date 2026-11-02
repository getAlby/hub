package main

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/db/queries"
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

type testCase struct {
	name      string
	sourceURI string
	destURI   string
}

func getTestCases() []testCase {
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

	return tc
}

func TestMigrate(t *testing.T) {
	for _, tt := range getTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			env, err := setupTest(t, tt.sourceURI, tt.destURI)
			require.NoError(t, err)
			defer env.cleanup(t)

			err = db.MigrateDB(env.source, env.dest)
			require.NoError(t, err)

			requireCount[db.App](t, env.dest, 2)
			requireCount[db.AppPermission](t, env.dest, 2)
			requireCount[db.RequestEvent](t, env.dest, 1)
			requireCount[db.ResponseEvent](t, env.dest, 1)
			requireCount[db.Transaction](t, env.dest, 1)
			requireCount[db.Swap](t, env.dest, 1)
			requireCount[db.Forward](t, env.dest, 1)
			requireCount[db.UserConfig](t, env.dest, 1)
		})
	}
}

func TestMigrateRemovesDeletedAppsFromTransactions(t *testing.T) {
	for _, tt := range getTestCases() {
		t.Run(tt.name, func(t *testing.T) {
			env, err := setupTest(t, tt.sourceURI, tt.destURI)
			require.NoError(t, err)
			defer env.cleanup(t)

			deletedApp := &db.App{
				Name:      "deleted",
				AppPubkey: "b923bc264d8d0a680f81599ea5b4e0145734564ccb314eb230b9916eeb5de170",
				Isolated:  true,
				Metadata:  datatypes.JSON("{}"),
			}
			create(t, env.source, deletedApp)
			transaction := &db.Transaction{
				AppId:       &deletedApp.ID,
				Type:        constants.TRANSACTION_TYPE_INCOMING,
				State:       constants.TRANSACTION_STATE_SETTLED,
				AmountMsat:  5_000_000,
				PaymentHash: "13a165e20d710ebd58cb0f41529cbd3a32454445beb54d1f169115e69dd6313d",
			}
			create(t, env.source, transaction)
			require.NoError(t, env.source.Delete(deletedApp).Error)

			err = db.MigrateDB(env.source, env.dest)
			require.NoError(t, err)

			var migratedTransaction db.Transaction
			require.NoError(t, env.dest.First(&migratedTransaction, transaction.ID).Error)
			assert.Nil(t, migratedTransaction.AppId)

			// the new app may get the deleted app's ID
			newApp := &db.App{
				Name:      "new",
				AppPubkey: "8d2c60ccd6f70721dd86573a062c65e06263f0e9e057ba8e5015e53d4b704870",
				Isolated:  true,
				Metadata:  datatypes.JSON("{}"),
			}
			create(t, env.dest, newApp)

			balance, err := queries.GetIsolatedBalanceMsat(env.dest, newApp.ID)
			require.NoError(t, err)
			assert.Zero(t, balance, "new app must not inherit the deleted app's balance")
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
		Value:     "wss://relay.getalby.com",
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

	requestEvent1 := &db.RequestEvent{
		AppId:       &app1.ID,
		NostrId:     "a35a1ca6d1a06e08a509f2c8fe3edb2ba10811d030e2f6f3239e9f21203ac954",
		ContentData: "{}",
		Method:      "pay_invoice",
		State:       "executed",
		CreatedAt:   baseTime,
		UpdatedAt:   baseTime,
	}
	create(t, tx, requestEvent1)

	responseEvent1 := &db.ResponseEvent{
		NostrId:   "e30d55d0e4f0d5391a1a1379f1d8b7d38ad02b3554b06ca993aa8790a3153f61",
		RequestId: requestEvent1.ID,
		State:     "confirmed",
		RepliedAt: baseTime,
		CreatedAt: baseTime,
		UpdatedAt: baseTime,
	}
	create(t, tx, responseEvent1)

	transaction1 := &db.Transaction{
		AppId:          &app1.ID,
		RequestEventId: &requestEvent1.ID,
		Type:           "outgoing",
		State:          "settled",
		AmountMsat:     21000,
		FeeMsat:        1000,
		PaymentRequest: "lnbc210n1invoice",
		PaymentHash:    "13d9764a54269fa4d5f4e7c410f4ffdbc839bbeaa2fcbb96343ca502f0c86e34",
		Description:    "test transaction",
		Preimage:       ptr("2c1ee1b464b1a1a147debe0ac0c8ce4b615f9bfa64d12a25c1c4d10ea45a5b02"),
		CreatedAt:      baseTime,
		UpdatedAt:      baseTime,
		SettledAt:      &baseTime,
		Metadata:       datatypes.JSON("{}"),
		Boostagram:     datatypes.JSON("{}"),
	}
	create(t, tx, transaction1)

	swap1 := &db.Swap{
		SwapId:             "swap1",
		Type:               "out",
		State:              "success",
		Invoice:            "lnbc210n1swapinvoice",
		SendAmountSat:      21000,
		ReceiveAmountSat:   20000,
		Preimage:           "35a3f1a7a06a41b9ba3a1b1a8ff852e5085b3b593f8ba4677a35a1ca6d1a06e0",
		PaymentHash:        "e6b1a1379f1d8b7d38ad02b3554b06ca993aa8790a3153f61e30d55d0e4f0d53",
		DestinationAddress: "bc1qtest",
		LockupAddress:      "bc1qlockup",
		LockupTxId:         "lockuptx",
		ClaimTxId:          "claimtx",
		AutoSwap:           false,
		TimeoutBlockHeight: 900000,
		BoltzPubkey:        "02d1a06e08a509f2c8fe3edb2ba10811d030e2f6f3239e9f21203ac954a35a1c",
		SwapTree:           datatypes.JSON("{}"),
		CreatedAt:          baseTime,
		UpdatedAt:          baseTime,
	}
	create(t, tx, swap1)

	forward1 := &db.Forward{
		OutboundAmountForwardedMsat: 1000000,
		TotalFeeEarnedMsat:          1000,
		CreatedAt:                   baseTime,
		UpdatedAt:                   baseTime,
	}
	create(t, tx, forward1)
}

func requireCount[T any](t *testing.T, tx *gorm.DB, expected int64) {
	var count int64
	var model T
	require.NoError(t, tx.Model(&model).Count(&count).Error)
	require.Equal(t, expected, count)
}

func create[T any](t *testing.T, tx *gorm.DB, v T) *gorm.DB {
	require.NoError(t, tx.Create(v).Error)
	return tx
}

func ptr[T any](v T) *T {
	return &v
}
