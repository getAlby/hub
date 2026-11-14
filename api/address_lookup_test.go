package api

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
	"github.com/getAlby/hub/tests/mocks"
)

// bitcoind has no address index, so the receive address is rotated after any
// incoming on-chain payment since it was issued
func TestGetUnusedOnchainAddress_Bitcoind(t *testing.T) {
	ctx := context.TODO()

	testSvc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer testSvc.Remove()

	cfg, err := config.NewConfig(&config.AppConfig{
		Workdir:            ".test",
		LDKBitcoindRpcHost: "127.0.0.1",
		LDKBitcoindRpcPort: "1",
	}, testSvc.DB)
	require.NoError(t, err)

	issuedAt := uint64(time.Now().Add(-time.Hour).Unix())
	require.NoError(t, cfg.SetUpdate(config.OnchainAddressKey, "bcrt1qcurrent", ""))
	require.NoError(t, cfg.SetUpdate(config.OnchainAddressIssuedAtKey, strconv.FormatUint(issuedAt, 10), ""))

	lnClient := mocks.NewMockLNClient(t)
	svc := mocks.NewMockService(t)
	svc.On("GetLNClient").Return(lnClient)
	theAPI := &api{cfg: cfg, svc: svc}

	// incoming payment before the address was issued, outgoing payment after
	lnClient.On("ListOnchainTransactions", mock.Anything).Return([]lnclient.OnchainTransaction{
		{Type: "incoming", CreatedAt: issuedAt - 60},
		{Type: "outgoing", CreatedAt: issuedAt + 60},
	}, nil).Once()

	address, err := theAPI.GetUnusedOnchainAddress(ctx)
	require.NoError(t, err)
	assert.Equal(t, "bcrt1qcurrent", address)

	// incoming payment after the address was issued
	lnClient.On("ListOnchainTransactions", mock.Anything).Return([]lnclient.OnchainTransaction{
		{Type: "incoming", CreatedAt: issuedAt + 60},
	}, nil).Once()
	lnClient.On("GetNewOnchainAddress", mock.Anything).Return("bcrt1qnext", nil).Once()

	address, err = theAPI.GetUnusedOnchainAddress(ctx)
	require.NoError(t, err)
	assert.Equal(t, "bcrt1qnext", address)

	// the new address's issue time is saved, so the earlier payment no longer rotates it
	newIssuedAtStr, err := cfg.Get(config.OnchainAddressIssuedAtKey, "")
	require.NoError(t, err)
	newIssuedAt, err := strconv.ParseUint(newIssuedAtStr, 10, 64)
	require.NoError(t, err)
	assert.Greater(t, newIssuedAt, issuedAt+60)
}

// addresses issued before the issue time was saved are rotated once if the
// wallet has ever received an on-chain payment
func TestCheckIncomingOnchainPaymentSinceAddressIssued_NoIssueTime(t *testing.T) {
	testSvc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer testSvc.Remove()

	lnClient := mocks.NewMockLNClient(t)
	svc := mocks.NewMockService(t)
	svc.On("GetLNClient").Return(lnClient)
	theAPI := &api{cfg: testSvc.Cfg, svc: svc}

	lnClient.On("ListOnchainTransactions", mock.Anything).Return([]lnclient.OnchainTransaction{
		{Type: "incoming", CreatedAt: 1_700_000_000},
	}, nil).Once()

	received, err := theAPI.checkIncomingOnchainPaymentSinceAddressIssued(context.TODO())
	require.NoError(t, err)
	assert.True(t, received)
}
