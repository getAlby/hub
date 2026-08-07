package transactions

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"gorm.io/datatypes"
)

func TestListTransactions_Paid(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, false, false, svc.LNClient, nil, false, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), totalCount)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, uint64(123000), incomingTransactions[0].AmountMsat)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransactions[0].State)
	assert.Equal(t, tests.MockLNClientTransaction.Preimage, *incomingTransactions[0].Preimage)
	assert.Zero(t, incomingTransactions[0].FeeReserveMsat)
}

func TestListTransactions_UnpaidIncoming(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now(),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-1 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, false, true, svc.LNClient, nil, false, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(3), totalCount)
	assert.Equal(t, 3, len(incomingTransactions))
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, incomingTransactions[0].State)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, incomingTransactions[1].State)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, incomingTransactions[2].State)
	for _, transaction := range incomingTransactions {
		assert.Equal(t, constants.TRANSACTION_TYPE_INCOMING, transaction.Type)
	}
}

func TestListTransactions_UnpaidOutgoing(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now(),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-1 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	outgoingTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, true, false, svc.LNClient, nil, false, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(3), totalCount)
	assert.Equal(t, 3, len(outgoingTransactions))
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, outgoingTransactions[0].State)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, outgoingTransactions[1].State)
	assert.Equal(t, constants.TRANSACTION_STATE_PENDING, outgoingTransactions[2].State)
	for _, transaction := range outgoingTransactions {
		assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, transaction.Type)
	}
}

func TestListTransactions_Unpaid(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now(),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-1 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		UpdatedAt:      time.Now().Add(-2 * time.Second),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	outgoingTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, true, true, svc.LNClient, nil, false, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(5), totalCount)
	assert.Equal(t, 5, len(outgoingTransactions))
}

func TestListTransactions_Limit(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(1 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 1, 0, false, false, svc.LNClient, nil, false, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), totalCount)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "first", incomingTransactions[0].Description)
}

func TestListTransactions_Offset(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(3 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(2 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "third",
		UpdatedAt:      time.Now().Add(1 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "fourth",
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 1, 2, false, false, svc.LNClient, nil, false, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(4), totalCount)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "third", incomingTransactions[0].Description)
}

func TestListTransactions_MinAmount(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     1000,
		Description:    "small",
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     10000,
		Description:    "large",
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	minAmountMsat := uint64(5000)

	filteredTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, false, false, svc.LNClient, nil, false, &ListTransactionsFilters{
		MinAmountMsat: &minAmountMsat,
	})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), totalCount)
	require.Len(t, filteredTransactions, 1)
	assert.Equal(t, "large", filteredTransactions[0].Description)
}

func TestListTransactions_HideFailed(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "settled",
		UpdatedAt:      time.Now().Add(2 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_FAILED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "failed",
		UpdatedAt:      time.Now().Add(1 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "pending",
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	filteredTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, true, false, svc.LNClient, nil, false, &ListTransactionsFilters{
		HideFailed: true,
	})
	assert.NoError(t, err)
	assert.Equal(t, uint64(2), totalCount)
	require.Len(t, filteredTransactions, 2)
	assert.Equal(t, "settled", filteredTransactions[0].Description)
	assert.Equal(t, "pending", filteredTransactions[1].Description)
}

func TestListTransactions_Type(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "received",
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "sent",
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transactionType := constants.TRANSACTION_TYPE_OUTGOING

	filteredTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, false, false, svc.LNClient, nil, false, &ListTransactionsFilters{
		Type: &transactionType,
	})
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), totalCount)
	require.Len(t, filteredTransactions, 1)
	assert.Equal(t, "sent", filteredTransactions[0].Description)
}

func TestListTransactions_Search(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: "lnbc1coffee",
		PaymentHash:    "3086c621ecbef1ba99446fca8f484e2dbef77b28ee76a94ab8bb8b0e7f60a0f1",
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "Coffee shop",
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "Zap",
		Metadata:       datatypes.JSON(`{"user_labels":{"category":"Drinks"}}`),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: "lnbc1discount",
		PaymentHash:    "af88b1571c1a0b2b1e8c05bf74e6a2f6b3a4f4a2be27077b1c5f5e2e4f6a8b9c",
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "50% discount",
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	for _, testCase := range []struct {
		searchTerm           string
		expectedDescriptions []string
	}{
		{"COFFEE", []string{"Coffee shop"}},
		// exact payment hash
		{tests.MockLNClientTransaction.PaymentHash, []string{"Zap"}},
		// full invoice is decoded and matched by its payment hash
		{tests.MockLNClientTransaction.Invoice, []string{"Zap"}},
		// invoices are not matched by substring
		{"lnbc1disc", []string{}},
		// partial payment hashes are not matched
		{tests.MockLNClientTransaction.PaymentHash[:32], []string{}},
		{"drinks", []string{"Zap"}},
		{"category", []string{"Zap"}},
		{"50%", []string{"50% discount"}},
		{"nonexistent", []string{}},
	} {
		filteredTransactions, totalCount, err := transactionsService.ListTransactions(ctx, 0, 0, 0, 0, false, false, svc.LNClient, nil, false, &ListTransactionsFilters{
			SearchTerm: testCase.searchTerm,
		})
		assert.NoError(t, err, "search: %s", testCase.searchTerm)
		assert.Equal(t, uint64(len(testCase.expectedDescriptions)), totalCount, "search: %s", testCase.searchTerm)
		require.Len(t, filteredTransactions, len(testCase.expectedDescriptions), "search: %s", testCase.searchTerm)
		for i, description := range testCase.expectedDescriptions {
			assert.Equal(t, description, filteredTransactions[i].Description, "search: %s", testCase.searchTerm)
		}
	}
}

func TestListTransactions_FromUntil(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(10 * time.Minute),
		CreatedAt:      time.Now().Add(10 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "third",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, totalCount, err := transactionsService.ListTransactions(ctx, uint64(time.Now().Add(4*time.Minute).Unix()), uint64(time.Now().Add(6*time.Minute).Unix()), 0, 0, false, false, svc.LNClient, nil, false, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), totalCount)
	assert.Equal(t, 1, len(incomingTransactions))
	assert.Equal(t, "second", incomingTransactions[0].Description)
}

func TestListTransactions_FromUntilUnpaidOutgoing(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(10 * time.Minute),
		CreatedAt:      time.Now().Add(10 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "third",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, totalCount, err := transactionsService.ListTransactions(ctx, uint64(time.Now().Add(4*time.Minute).Unix()), uint64(time.Now().Add(6*time.Minute).Unix()), 0, 0, true, false, svc.LNClient, nil, false, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), totalCount)
	assert.Equal(t, "second", incomingTransactions[0].Description)
	assert.Equal(t, constants.TRANSACTION_TYPE_OUTGOING, incomingTransactions[0].Type)
}

func TestListTransactions_FromUntilUnpaidIncoming(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	mockPreimage := tests.MockLNClientTransaction.Preimage
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "first",
		UpdatedAt:      time.Now().Add(10 * time.Minute),
		CreatedAt:      time.Now().Add(10 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "second",
		UpdatedAt:      time.Now().Add(5 * time.Minute),
		CreatedAt:      time.Now().Add(5 * time.Minute),
	})
	svc.DB.Create(&db.Transaction{
		State:          constants.TRANSACTION_STATE_PENDING,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &mockPreimage,
		AmountMsat:     123000,
		Description:    "third",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	})

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)

	incomingTransactions, totalCount, err := transactionsService.ListTransactions(ctx, uint64(time.Now().Add(4*time.Minute).Unix()), uint64(time.Now().Add(6*time.Minute).Unix()), 0, 0, false, true, svc.LNClient, nil, false, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), totalCount)
	assert.Equal(t, "second", incomingTransactions[0].Description)
	assert.Equal(t, constants.TRANSACTION_TYPE_INCOMING, incomingTransactions[0].Type)
}
