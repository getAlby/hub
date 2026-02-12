package queries

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
)

func TestGetIsolatedBalance_PendingNoOverflow(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	app.Isolated = true
	svc.DB.Save(&app)

	paymentAmount := uint64(1000) // 1 sat

	tx := db.Transaction{
		AppId:          &app.ID,
		RequestEventId: nil,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_PENDING,
		FeeReserveMsat: uint64(10000),
		AmountMsat:     paymentAmount,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		SelfPayment:    true,
	}
	svc.DB.Save(&tx)

	balance, err := GetIsolatedBalance(svc.DB, app.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(-11000), balance)
}

func TestGetIsolatedBalance_SettledNoOverflow(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	app, _, err := tests.CreateApp(svc)
	assert.NoError(t, err)
	app.Isolated = true
	svc.DB.Save(&app)

	paymentAmount := uint64(1000) // 1 sat

	tx := db.Transaction{
		AppId:          &app.ID,
		RequestEventId: nil,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		FeeReserveMsat: uint64(0),
		AmountMsat:     paymentAmount,
		PaymentRequest: tests.MockInvoice,
		PaymentHash:    tests.MockPaymentHash,
		SelfPayment:    true,
	}
	svc.DB.Save(&tx)

	balance, err := GetIsolatedBalance(svc.DB, app.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(-1000), balance)
}
