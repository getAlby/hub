package queries

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
)

func TestGetBudgetUsage_IncludesPendingAndSettledOutgoing(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		BudgetRenewal: constants.BUDGET_RENEWAL_NEVER,
	}

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_PENDING,
		AmountMsat:     50000,
		FeeMsat:        1000,
		FeeReserveMsat: 2000,
	}).Error)

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		AmountMsat:     25000,
		FeeMsat:        500,
		FeeReserveMsat: 500,
	}).Error)

	budgetUsage, err := GetBudgetUsage(svc.DB, appPermission)
	require.NoError(t, err)
	assert.Equal(t, uint64(79000), budgetUsage)
}

func TestGetBudgetUsage_ExcludesWrongStateTypeAndApp(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	otherApp, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		BudgetRenewal: constants.BUDGET_RENEWAL_NEVER,
	}

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		AmountMsat:     20000,
		FeeMsat:        1000,
		FeeReserveMsat: 0,
	}).Error)

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		AmountMsat:     90000,
		FeeMsat:        0,
		FeeReserveMsat: 0,
	}).Error)

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_FAILED,
		AmountMsat:     90000,
		FeeMsat:        0,
		FeeReserveMsat: 0,
	}).Error)

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &otherApp.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		AmountMsat:     90000,
		FeeMsat:        0,
		FeeReserveMsat: 0,
	}).Error)

	budgetUsage, err := GetBudgetUsage(svc.DB, appPermission)
	require.NoError(t, err)
	assert.Equal(t, uint64(21000), budgetUsage)
}

func TestGetBudgetUsage_BudgetWindow(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	appPermissionDaily := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		BudgetRenewal: constants.BUDGET_RENEWAL_DAILY,
	}

	dailyStart := getStartOfBudget(constants.BUDGET_RENEWAL_DAILY)

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		AmountMsat:     20000,
		FeeMsat:        1000,
		FeeReserveMsat: 0,
		CreatedAt:      dailyStart.Add(1 * time.Minute),
	}).Error)

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_PENDING,
		AmountMsat:     20000,
		FeeMsat:        0,
		FeeReserveMsat: 1000,
		CreatedAt:      dailyStart.Add(2 * time.Hour),
	}).Error)

	budgetUsage, err := GetBudgetUsage(svc.DB, appPermissionDaily)
	require.NoError(t, err)
	assert.Equal(t, uint64(42000), budgetUsage)

	require.NoError(t, svc.DB.Where("app_id = ?", app.ID).Delete(&db.Transaction{}).Error)

	appPermissionWeekly := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		BudgetRenewal: constants.BUDGET_RENEWAL_WEEKLY,
	}

	weeklyStart := getStartOfBudget(constants.BUDGET_RENEWAL_WEEKLY)

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		AmountMsat:     10000,
		FeeMsat:        1000,
		FeeReserveMsat: 0,
		CreatedAt:      weeklyStart.Add(30 * time.Minute),
	}).Error)

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_PENDING,
		AmountMsat:     7000,
		FeeMsat:        0,
		FeeReserveMsat: 0,
		CreatedAt:      weeklyStart.Add(-1 * time.Hour),
	}).Error)

	budgetUsage, err = GetBudgetUsage(svc.DB, appPermissionWeekly)
	require.NoError(t, err)
	assert.Equal(t, uint64(11000), budgetUsage)
}

func TestGetBudgetUsage_BudgetWindowNever(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	require.NoError(t, err)

	appPermission := &db.AppPermission{
		AppId:         app.ID,
		App:           *app,
		Scope:         constants.PAY_INVOICE_SCOPE,
		BudgetRenewal: constants.BUDGET_RENEWAL_NEVER,
	}

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		AmountMsat:     5000,
		FeeMsat:        1000,
		FeeReserveMsat: 0,
		CreatedAt:      time.Now().AddDate(-2, 0, 0),
	}).Error)

	require.NoError(t, svc.DB.Create(&db.Transaction{
		AppId:          &app.ID,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_PENDING,
		AmountMsat:     4000,
		FeeMsat:        0,
		FeeReserveMsat: 1000,
		CreatedAt:      time.Now().AddDate(-1, 0, 0),
	}).Error)

	budgetUsage, err := GetBudgetUsage(svc.DB, appPermission)
	require.NoError(t, err)
	assert.Equal(t, uint64(11000), budgetUsage)
}
