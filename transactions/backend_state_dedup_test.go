package transactions

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// backendLookupMockLn wraps tests.MockLn with an authoritative backend payment
// status lookup and counts SendPaymentSync calls, so tests can prove the hub
// never re-pays an invoice the node has already settled or is still paying.
type backendLookupMockLn struct {
	*tests.MockLn

	mu       sync.Mutex
	sends    int
	lookupFn func(paymentHash string) (*lnclient.PaymentStatusResponse, error)
}

func (m *backendLookupMockLn) SendPaymentSync(payReq string, amountMsat *uint64) (*lnclient.PayInvoiceResponse, error) {
	m.mu.Lock()
	m.sends++
	m.mu.Unlock()
	return m.MockLn.SendPaymentSync(payReq, amountMsat)
}

func (m *backendLookupMockLn) LookupPaymentStatus(paymentHash string) (*lnclient.PaymentStatusResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.lookupFn(paymentHash)
}

func (m *backendLookupMockLn) sendCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.sends
}

func newBackendLookupMockLn(svc *tests.TestService, fn func(string) (*lnclient.PaymentStatusResponse, error)) *backendLookupMockLn {
	return &backendLookupMockLn{
		MockLn:   svc.LNClient.(*tests.MockLn),
		lookupFn: fn,
	}
}

func completeStatus() *lnclient.PaymentStatusResponse {
	return &lnclient.PaymentStatusResponse{
		Status:   lnclient.PAYMENT_STATUS_COMPLETE,
		Preimage: "abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234abcd1234",
		FeeMsat:  2000,
	}
}

func createOutgoingRow(svc *tests.TestService, state string, feeReserveMsat uint64) {
	svc.DB.Create(&db.Transaction{
		State:          state,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		AmountMsat:     123000,
		FeeReserveMsat: feeReserveMsat,
	})
}

func findRows(svc *tests.TestService) []db.Transaction {
	rows := []db.Transaction{}
	svc.DB.Find(&rows)
	return rows
}

// The sync response was lost and the row marked FAILED, but the node actually
// settled the payment. A retry must reconcile from backend state, not re-pay.
func TestDedup_BackendComplete_WithFailedRow_SettlesWithoutRepay(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	createOutgoingRow(svc, constants.TRANSACTION_STATE_FAILED, 0)

	mock := newBackendLookupMockLn(svc, func(string) (*lnclient.PaymentStatusResponse, error) {
		return completeStatus(), nil
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, mock, nil, nil)

	assert.NoError(t, err)
	require.NotNil(t, transaction)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	require.NotNil(t, transaction.Preimage)
	assert.Equal(t, completeStatus().Preimage, *transaction.Preimage)
	assert.Equal(t, uint64(2000), transaction.FeeMsat)
	assert.Zero(t, transaction.FeeReserveMsat)
	assert.Equal(t, 0, mock.sendCount())

	rows := findRows(svc)
	require.Equal(t, 1, len(rows))
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, rows[0].State)
}

// The settlement happened while the hub row was still PENDING (async event
// lost). A retry must reconcile to SETTLED without a second backend payment.
func TestDedup_BackendComplete_WithPendingRow_SettlesWithoutRepay(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	createOutgoingRow(svc, constants.TRANSACTION_STATE_PENDING, 10000)

	mock := newBackendLookupMockLn(svc, func(string) (*lnclient.PaymentStatusResponse, error) {
		return completeStatus(), nil
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, mock, nil, nil)

	assert.NoError(t, err)
	require.NotNil(t, transaction)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, 0, mock.sendCount())

	rows := findRows(svc)
	require.Equal(t, 1, len(rows))
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, rows[0].State)
}

// The DB row says FAILED (sync error path) but the node's payment group is
// still in progress. This is the groupid retry-collision: a new send must NOT
// be initiated.
func TestDedup_BackendPending_WithStaleFailedRow_BlocksRepay(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	createOutgoingRow(svc, constants.TRANSACTION_STATE_FAILED, 0)

	mock := newBackendLookupMockLn(svc, func(string) (*lnclient.PaymentStatusResponse, error) {
		return &lnclient.PaymentStatusResponse{Status: lnclient.PAYMENT_STATUS_PENDING}, nil
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, mock, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, 0, mock.sendCount())

	rows := findRows(svc)
	require.Equal(t, 1, len(rows))
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, rows[0].State)
}

// Backend also reports no in-flight payment without a DB row: nothing may be
// started against a still-pending node payment even if the hub lost its row.
func TestDedup_BackendPending_WithoutRow_BlocksPayment(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mock := newBackendLookupMockLn(svc, func(string) (*lnclient.PaymentStatusResponse, error) {
		return &lnclient.PaymentStatusResponse{Status: lnclient.PAYMENT_STATUS_PENDING}, nil
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, mock, nil, nil)

	assert.Error(t, err)
	assert.Nil(t, transaction)
	assert.Equal(t, 0, mock.sendCount())
	assert.Equal(t, 0, len(findRows(svc)))
}

// Backend reports the prior attempt genuinely failed: the legacy retry path
// proceeds.
func TestDedup_BackendFailed_ProceedsWithNewPayment(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	createOutgoingRow(svc, constants.TRANSACTION_STATE_FAILED, 0)

	mock := newBackendLookupMockLn(svc, func(string) (*lnclient.PaymentStatusResponse, error) {
		return &lnclient.PaymentStatusResponse{Status: lnclient.PAYMENT_STATUS_FAILED}, nil
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, mock, nil, nil)

	assert.NoError(t, err)
	require.NotNil(t, transaction)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, 1, mock.sendCount())

	rows := findRows(svc)
	assert.Equal(t, 2, len(rows))
}

// Unknown backend state falls through to the existing DB guard behaviour.
func TestDedup_BackendUnknown_ProceedsWithLegacyGuard(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mock := newBackendLookupMockLn(svc, func(string) (*lnclient.PaymentStatusResponse, error) {
		return &lnclient.PaymentStatusResponse{Status: lnclient.PAYMENT_STATUS_UNKNOWN}, nil
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, mock, nil, nil)

	assert.NoError(t, err)
	require.NotNil(t, transaction)
	assert.Equal(t, 1, mock.sendCount())
}

// A lookup RPC failure must not block the legacy path (backstop stays the DB
// guard plus the node's own checks).
func TestDedup_LookupError_ProceedsWithLegacyGuard(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mock := newBackendLookupMockLn(svc, func(string) (*lnclient.PaymentStatusResponse, error) {
		return nil, errors.New("rpc timeout")
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, mock, nil, nil)

	assert.NoError(t, err)
	require.NotNil(t, transaction)
	assert.Equal(t, 1, mock.sendCount())
}

// Two concurrent callers for the same invoice must result in exactly one
// backend payment.
func TestDedup_ConcurrentCallers_SingleBackendPayment(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	delay := 200 * time.Millisecond
	svc.LNClient.(*tests.MockLn).PaymentDelay = &delay

	mock := newBackendLookupMockLn(svc, func(string) (*lnclient.PaymentStatusResponse, error) {
		return &lnclient.PaymentStatusResponse{Status: lnclient.PAYMENT_STATUS_UNKNOWN}, nil
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	var wg sync.WaitGroup
	results := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, mock, nil, nil)
			results[idx] = err
		}(i)
	}
	wg.Wait()

	successes := 0
	for _, err := range results {
		if err == nil {
			successes++
		}
	}
	assert.Equal(t, 1, successes, "exactly one caller should get a settled payment")
	assert.Equal(t, 1, mock.sendCount(), "the backend must receive exactly one payment")

	rows := findRows(svc)
	require.Equal(t, 1, len(rows))
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, rows[0].State)
}

// Backends without the optional lookup capability keep the old behavior.
func TestDedup_BackendWithoutLookupCapability_LegacyBehavior(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	createOutgoingRow(svc, constants.TRANSACTION_STATE_FAILED, 0)

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.SendPaymentSync(tests.MockLNClientTransaction.Invoice, nil, nil, svc.LNClient, nil, nil)

	assert.NoError(t, err)
	require.NotNil(t, transaction)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, transaction.State)
	assert.Equal(t, 2, len(findRows(svc)))
}
