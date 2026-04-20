package api

import (
	"testing"
	"time"

	"github.com/getAlby/hub/swaps"
	"github.com/stretchr/testify/require"
)

func TestPayInvoiceRequestResolvedAmountMsat(t *testing.T) {
	deprecated := uint64(1000)
	explicit := uint64(2000)

	require.Nil(t, (*PayInvoiceRequest)(nil).ResolvedAmountMsat())
	require.Equal(t, &deprecated, (&PayInvoiceRequest{Amount: &deprecated}).ResolvedAmountMsat())
	require.Equal(t, &explicit, (&PayInvoiceRequest{Amount: &deprecated, AmountMsat: &explicit}).ResolvedAmountMsat())
}

func TestMakeInvoiceRequestResolvedAmountMsat(t *testing.T) {
	explicit := uint64(2000)

	require.Equal(t, uint64(0), (*MakeInvoiceRequest)(nil).ResolvedAmountMsat())
	require.Equal(t, uint64(1000), (&MakeInvoiceRequest{Amount: 1000}).ResolvedAmountMsat())
	require.Equal(t, explicit, (&MakeInvoiceRequest{Amount: 1000, AmountMsat: &explicit}).ResolvedAmountMsat())
}

func TestSendSpontaneousPaymentProbesRequestResolvedAmountMsat(t *testing.T) {
	require.Equal(t, uint64(0), (*SendSpontaneousPaymentProbesRequest)(nil).ResolvedAmountMsat())
	require.Equal(t, uint64(1000), (&SendSpontaneousPaymentProbesRequest{Amount: 1000}).ResolvedAmountMsat())
	require.Equal(t, uint64(2000), (&SendSpontaneousPaymentProbesRequest{Amount: 1000, AmountMsat: 2000}).ResolvedAmountMsat())
}

func TestLSPOrderRequestResolvedAmountSat(t *testing.T) {
	require.Equal(t, uint64(0), (*LSPOrderRequest)(nil).ResolvedAmountSat())
	require.Equal(t, uint64(1000), (&LSPOrderRequest{Amount: 1000}).ResolvedAmountSat())
	require.Equal(t, uint64(2000), (&LSPOrderRequest{Amount: 1000, AmountSat: 2000}).ResolvedAmountSat())
}

func TestRedeemOnchainFundsRequestResolvedFields(t *testing.T) {
	deprecatedFeeRate := uint64(5)
	explicitFeeRate := uint64(10)

	require.Equal(t, uint64(0), (*RedeemOnchainFundsRequest)(nil).ResolvedAmountSat())
	require.Nil(t, (*RedeemOnchainFundsRequest)(nil).ResolvedFeeRateSatPerVbyte())

	req := &RedeemOnchainFundsRequest{
		Amount:  1000,
		FeeRate: &deprecatedFeeRate,
	}
	require.Equal(t, uint64(1000), req.ResolvedAmountSat())
	require.Equal(t, &deprecatedFeeRate, req.ResolvedFeeRateSatPerVbyte())

	req.AmountSat = 2000
	req.FeeRateSatPerVbyte = &explicitFeeRate
	require.Equal(t, uint64(2000), req.ResolvedAmountSat())
	require.Equal(t, &explicitFeeRate, req.ResolvedFeeRateSatPerVbyte())
}

func TestInitiateSwapRequestResolvedSwapAmountSat(t *testing.T) {
	require.Equal(t, uint64(0), (*InitiateSwapRequest)(nil).ResolvedSwapAmountSat())
	require.Equal(t, uint64(1000), (&InitiateSwapRequest{SwapAmount: 1000}).ResolvedSwapAmountSat())
	require.Equal(t, uint64(2000), (&InitiateSwapRequest{SwapAmount: 1000, SwapAmountSat: 2000}).ResolvedSwapAmountSat())
}

func TestEnableAutoSwapRequestResolvedFields(t *testing.T) {
	require.Equal(t, uint64(0), (*EnableAutoSwapRequest)(nil).ResolvedBalanceThresholdSat())
	require.Equal(t, uint64(0), (*EnableAutoSwapRequest)(nil).ResolvedSwapAmountSat())

	req := &EnableAutoSwapRequest{
		BalanceThreshold: 1000,
		SwapAmount:       2000,
	}
	require.Equal(t, uint64(1000), req.ResolvedBalanceThresholdSat())
	require.Equal(t, uint64(2000), req.ResolvedSwapAmountSat())

	req.BalanceThresholdSat = 3000
	req.SwapAmountSat = 4000
	require.Equal(t, uint64(3000), req.ResolvedBalanceThresholdSat())
	require.Equal(t, uint64(4000), req.ResolvedSwapAmountSat())
}

func TestToApiSwapUsesSatFieldsAndKeepsDeprecatedAliases(t *testing.T) {
	now := time.Now()
	dbSwap := &swaps.Swap{
		SwapId:             "swap-id",
		Type:               "out",
		State:              "pending",
		Invoice:            "invoice",
		SendAmountSat:      123,
		ReceiveAmountSat:   456,
		PaymentHash:        "payment-hash",
		DestinationAddress: "destination",
		RefundAddress:      "refund",
		LockupAddress:      "lockup",
		LockupTxId:         "lockup-tx",
		ClaimTxId:          "claim-tx",
		AutoSwap:           true,
		BoltzPubkey:        "boltz-pubkey",
		CreatedAt:          now,
		UpdatedAt:          now,
		UsedXpub:           true,
	}

	apiSwap := toApiSwap(dbSwap)
	require.Equal(t, uint64(123), apiSwap.SendAmount)
	require.Equal(t, uint64(123), apiSwap.SendAmountSat)
	require.Equal(t, uint64(456), apiSwap.ReceiveAmount)
	require.Equal(t, uint64(456), apiSwap.ReceiveAmountSat)
}
