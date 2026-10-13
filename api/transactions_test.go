package api

import (
	"context"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/tests/mocks"
	"github.com/getAlby/hub/transactions"
)

func TestCreateInvoice_ToApp(t *testing.T) {
	ctx := context.TODO()

	testSvc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer testSvc.Remove()

	app, _, err := tests.CreateApp(testSvc)
	require.NoError(t, err)

	svc := mocks.NewMockService(t)
	svc.On("GetLNClient").Return(testSvc.LNClient)
	svc.On("GetTransactionsService").Return(transactions.NewTransactionsService(testSvc.DB, testSvc.EventPublisher))

	theAPI := &api{
		appsSvc: testSvc.AppsService,
		svc:     svc,
	}

	transaction, err := theAPI.CreateInvoice(ctx, 1000, "Hello world", &app.ID)

	require.NoError(t, err)
	require.NotNil(t, transaction.AppId)
	assert.Equal(t, app.ID, *transaction.AppId)
}

func TestCreateInvoice_ToAppNotFound(t *testing.T) {
	ctx := context.TODO()

	testSvc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer testSvc.Remove()

	svc := mocks.NewMockService(t)
	svc.On("GetLNClient").Return(testSvc.LNClient)

	theAPI := &api{
		appsSvc: testSvc.AppsService,
		svc:     svc,
	}

	missingAppId := uint(999)
	transaction, err := theAPI.CreateInvoice(ctx, 1000, "Hello world", &missingAppId)

	assert.Nil(t, transaction)
	require.Error(t, err)
	assert.Equal(t, "app does not exist", err.Error())
}

func TestParseListTransactionsFilters(t *testing.T) {
	minAmountMsat := uint64(1000_000)
	outgoing := "outgoing"

	filters, err := ParseListTransactionsFilters(url.Values{
		"type":         {"outgoing"},
		"minAmountSat": {"1000"},
		"hideFailed":   {"true"},
		"search":       {" coffee "},
	})
	require.NoError(t, err)
	assert.Equal(t, ListTransactionsFilters{
		Type:          &outgoing,
		MinAmountMsat: &minAmountMsat,
		HideFailed:    true,
		SearchTerm:    "coffee",
	}, filters)

	filters, err = ParseListTransactionsFilters(url.Values{})
	require.NoError(t, err)
	assert.Equal(t, ListTransactionsFilters{}, filters)

	for _, invalidQuery := range []url.Values{
		{"type": {"sideways"}},
		{"minAmountSat": {"abc"}},
		{"minAmountSat": {"-1"}},
		{"minAmountSat": {"0"}},
		{"minAmountSat": {"18446744073709551615"}},
		{"hideFailed": {"maybe"}},
	} {
		_, err = ParseListTransactionsFilters(invalidQuery)
		assert.Error(t, err, "query: %v", invalidQuery)
	}
}
