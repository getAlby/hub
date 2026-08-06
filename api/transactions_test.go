package api

import (
	"context"
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
