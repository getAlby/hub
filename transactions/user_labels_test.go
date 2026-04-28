package transactions

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/datatypes"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/tests"
)

func setupTransactionForUserLabels(t *testing.T, transactionType string, initialMetadata map[string]interface{}) (*transactionsService, *tests.TestService, *db.Transaction) {
	t.Helper()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)

	preimage := tests.MockLNClientTransaction.Preimage
	dbTransaction := &db.Transaction{
		State:          constants.TRANSACTION_STATE_SETTLED,
		Type:           transactionType,
		PaymentRequest: tests.MockLNClientTransaction.Invoice,
		PaymentHash:    tests.MockLNClientTransaction.PaymentHash,
		Preimage:       &preimage,
		AmountMsat:     123000,
	}
	if initialMetadata != nil {
		metadataBytes, err := json.Marshal(initialMetadata)
		require.NoError(t, err)
		dbTransaction.Metadata = datatypes.JSON(metadataBytes)
	}

	require.NoError(t, svc.DB.Create(dbTransaction).Error)

	return NewTransactionsService(svc.DB, svc.EventPublisher), svc, dbTransaction
}

func loadTransactionMetadata(t *testing.T, svc *tests.TestService, id uint) map[string]interface{} {
	t.Helper()

	var dbTransaction db.Transaction
	require.NoError(t, svc.DB.First(&dbTransaction, id).Error)
	if dbTransaction.Metadata == nil {
		return nil
	}

	metadata := map[string]interface{}{}
	require.NoError(t, json.Unmarshal(dbTransaction.Metadata, &metadata))
	return metadata
}

func TestSetTransactionUserLabels_ValidLabels(t *testing.T) {
	transactionsService, svc, dbTransaction := setupTransactionForUserLabels(t, constants.TRANSACTION_TYPE_OUTGOING, nil)
	defer svc.Remove()

	err := transactionsService.SetTransactionUserLabels(context.TODO(), dbTransaction.ID, map[string]string{
		"description":  "top up PPQ.AI",
		"counterparty": "PPQ.AI",
	})
	require.NoError(t, err)

	metadata := loadTransactionMetadata(t, svc, dbTransaction.ID)
	labels, ok := metadata["user_labels"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "top up PPQ.AI", labels["description"])
	assert.Equal(t, "PPQ.AI", labels["counterparty"])
}

func TestSetTransactionUserLabels_IncomingTransaction(t *testing.T) {
	transactionsService, svc, dbTransaction := setupTransactionForUserLabels(t, constants.TRANSACTION_TYPE_INCOMING, nil)
	defer svc.Remove()

	err := transactionsService.SetTransactionUserLabels(context.TODO(), dbTransaction.ID, map[string]string{
		"source": "customer",
	})
	require.NoError(t, err)

	metadata := loadTransactionMetadata(t, svc, dbTransaction.ID)
	labels, ok := metadata["user_labels"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "customer", labels["source"])
}

func TestSetTransactionUserLabels_PreservesExistingMetadata(t *testing.T) {
	transactionsService, svc, dbTransaction := setupTransactionForUserLabels(t, constants.TRANSACTION_TYPE_OUTGOING, map[string]interface{}{
		"comment": "hello",
		"nostr":   map[string]interface{}{"pubkey": "abcdef"},
	})
	defer svc.Remove()

	err := transactionsService.SetTransactionUserLabels(context.TODO(), dbTransaction.ID, map[string]string{
		"account": "sponsoring",
	})
	require.NoError(t, err)

	metadata := loadTransactionMetadata(t, svc, dbTransaction.ID)
	assert.Equal(t, "hello", metadata["comment"])
	assert.NotNil(t, metadata["nostr"])
	labels, ok := metadata["user_labels"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "sponsoring", labels["account"])
}

func TestSetTransactionUserLabels_ClearsLabels(t *testing.T) {
	transactionsService, svc, dbTransaction := setupTransactionForUserLabels(t, constants.TRANSACTION_TYPE_OUTGOING, map[string]interface{}{
		"comment":     "hello",
		"user_labels": map[string]interface{}{"description": "old"},
	})
	defer svc.Remove()

	err := transactionsService.SetTransactionUserLabels(context.TODO(), dbTransaction.ID, map[string]string{})
	require.NoError(t, err)

	metadata := loadTransactionMetadata(t, svc, dbTransaction.ID)
	_, exists := metadata["user_labels"]
	assert.False(t, exists)
	assert.Equal(t, "hello", metadata["comment"])
}

func TestSetTransactionUserLabels_TrimsAndDropsBlankLabels(t *testing.T) {
	transactionsService, svc, dbTransaction := setupTransactionForUserLabels(t, constants.TRANSACTION_TYPE_OUTGOING, nil)
	defer svc.Remove()

	err := transactionsService.SetTransactionUserLabels(context.TODO(), dbTransaction.ID, map[string]string{
		"  account  ": "  sponsoring  ",
		"empty":       "",
		"":            "orphan",
	})
	require.NoError(t, err)

	metadata := loadTransactionMetadata(t, svc, dbTransaction.ID)
	labels, ok := metadata["user_labels"].(map[string]interface{})
	require.True(t, ok)
	assert.Len(t, labels, 1)
	assert.Equal(t, "sponsoring", labels["account"])
}

func TestSetTransactionUserLabels_RejectsOversizedMetadata(t *testing.T) {
	transactionsService, svc, dbTransaction := setupTransactionForUserLabels(t, constants.TRANSACTION_TYPE_OUTGOING, nil)
	defer svc.Remove()

	labels := map[string]string{
		"description": strings.Repeat("a", constants.INVOICE_METADATA_MAX_LENGTH),
	}
	err := transactionsService.SetTransactionUserLabels(context.TODO(), dbTransaction.ID, labels)
	require.Error(t, err)

	encodedMetadata, marshalErr := json.Marshal(map[string]interface{}{
		"user_labels": labels,
	})
	require.NoError(t, marshalErr)

	assert.Equal(t,
		fmt.Sprintf("encoded invoice metadata provided is too large. Limit: %d Received: %d", constants.INVOICE_METADATA_MAX_LENGTH, len(encodedMetadata)),
		err.Error(),
	)
}
