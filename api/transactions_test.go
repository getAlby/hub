package api

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/tests/mocks"
	"github.com/getAlby/hub/transactions"
)

func setupAPIWithTransaction(t *testing.T, initialMetadata map[string]interface{}) (*api, *tests.TestService, *db.Transaction) {
	t.Helper()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)

	mockSvc := mocks.NewMockService(t)
	mockSvc.On("GetLNClient").Return(svc.LNClient).Maybe()
	mockSvc.On("GetTransactionsService").Return(transactions.NewTransactionsService(svc.DB, svc.EventPublisher)).Maybe()

	preimage := tests.MockLNClientTransaction.Preimage
	dbTx := &db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           constants.TRANSACTION_TYPE_OUTGOING,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &preimage,
		AmountMsat:     123000,
	}
	if initialMetadata != nil {
		bytes, err := json.Marshal(initialMetadata)
		require.NoError(t, err)
		dbTx.Metadata = datatypes.JSON(bytes)
	}
	require.NoError(t, svc.DB.Create(dbTx).Error)

	return &api{svc: mockSvc}, svc, dbTx
}

func loadMetadata(t *testing.T, svc *tests.TestService, paymentHash string) map[string]interface{} {
	t.Helper()
	var stored db.Transaction
	require.NoError(t, svc.DB.Where(&db.Transaction{PaymentHash: paymentHash}).First(&stored).Error)
	if stored.Metadata == nil {
		return nil
	}
	out := map[string]interface{}{}
	require.NoError(t, json.Unmarshal(stored.Metadata, &out))
	return out
}

func TestSetTransactionUserLabels_Set(t *testing.T) {
	a, svc, dbTx := setupAPIWithTransaction(t, nil)
	defer svc.Remove()

	err := a.SetTransactionUserLabels(context.TODO(), dbTx.PaymentHash, map[string]string{
		"description":  "top up PPQ.AI",
		"counterparty": "PPQ.AI",
	})
	require.NoError(t, err)

	stored := loadMetadata(t, svc, dbTx.PaymentHash)
	labels, ok := stored["user_label"].(map[string]interface{})
	require.True(t, ok, "user_label should be set")
	assert.Equal(t, "top up PPQ.AI", labels["description"])
	assert.Equal(t, "PPQ.AI", labels["counterparty"])
}

func TestSetTransactionUserLabels_PreservesOtherKeys(t *testing.T) {
	a, svc, dbTx := setupAPIWithTransaction(t, map[string]interface{}{
		"comment": "hello",
		"nostr":   map[string]interface{}{"pubkey": "abcdef"},
	})
	defer svc.Remove()

	err := a.SetTransactionUserLabels(context.TODO(), dbTx.PaymentHash, map[string]string{
		"account": "sponsoring",
	})
	require.NoError(t, err)

	stored := loadMetadata(t, svc, dbTx.PaymentHash)
	assert.Equal(t, "hello", stored["comment"])
	assert.NotNil(t, stored["nostr"])
	labels, ok := stored["user_label"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "sponsoring", labels["account"])
}

func TestSetTransactionUserLabels_ClearRemovesKey(t *testing.T) {
	a, svc, dbTx := setupAPIWithTransaction(t, map[string]interface{}{
		"comment":    "hello",
		"user_label": map[string]interface{}{"description": "old"},
	})
	defer svc.Remove()

	err := a.SetTransactionUserLabels(context.TODO(), dbTx.PaymentHash, map[string]string{})
	require.NoError(t, err)

	stored := loadMetadata(t, svc, dbTx.PaymentHash)
	_, exists := stored["user_label"]
	assert.False(t, exists, "user_label should be removed when labels are empty")
	assert.Equal(t, "hello", stored["comment"])
}

func TestSetTransactionUserLabels_TrimsAndDropsBlankRows(t *testing.T) {
	a, svc, dbTx := setupAPIWithTransaction(t, nil)
	defer svc.Remove()

	err := a.SetTransactionUserLabels(context.TODO(), dbTx.PaymentHash, map[string]string{
		"  account  ": "  sponsoring  ",
		"empty":       "",
		"":            "orphan",
	})
	require.NoError(t, err)

	stored := loadMetadata(t, svc, dbTx.PaymentHash)
	labels, ok := stored["user_label"].(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, labels, 1)
	assert.Equal(t, "sponsoring", labels["account"])
}

func TestSetTransactionUserLabels_RejectsLongKey(t *testing.T) {
	a, svc, dbTx := setupAPIWithTransaction(t, nil)
	defer svc.Remove()

	err := a.SetTransactionUserLabels(context.TODO(), dbTx.PaymentHash, map[string]string{
		strings.Repeat("k", userLabelKeyMaxLength+1): "v",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "key too long")
}

func TestSetTransactionUserLabels_RejectsLongValue(t *testing.T) {
	a, svc, dbTx := setupAPIWithTransaction(t, nil)
	defer svc.Remove()

	err := a.SetTransactionUserLabels(context.TODO(), dbTx.PaymentHash, map[string]string{
		"description": strings.Repeat("v", userLabelValueMaxLength+1),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "value too long")
}
