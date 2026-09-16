package transactions

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
	"github.com/stretchr/testify/require"
)

type offerLifecycleStub struct {
	lnclient.LNClient
	waitErr  error
	response *lnclient.PayOfferResponse
	fee      *uint64
}

func (c *offerLifecycleStub) StartOfferPayment(context.Context, string, *uint64, string) (string, error) {
	return "backend-payment-id", nil
}

func (c *offerLifecycleStub) StartOfferPaymentWithFeeLimit(ctx context.Context, offer string, amount *uint64, note string, fee uint64) (string, error) {
	c.fee = &fee
	return c.StartOfferPayment(ctx, offer, amount, note)
}

func (c *offerLifecycleStub) WaitForOfferPayment(context.Context, string) (*lnclient.PayOfferResponse, error) {
	return c.response, c.waitErr
}

func (c *offerLifecycleStub) LookupOfferPayment(context.Context, string) (*lnclient.PayOfferResponse, error) {
	return c.response, c.waitErr
}

func (c *offerLifecycleStub) GetSupportedNIP47NotificationTypes() []string {
	return []string{"payment_received", "payment_sent"}
}

func TestOfferReconcilesAfterLostEvent(t *testing.T) {
	for _, failed := range []bool{false, true} {
		t.Run(fmt.Sprint(failed), func(t *testing.T) {
			svc, err := tests.CreateTestService(t)
			require.NoError(t, err)
			defer svc.Remove()
			client := &offerLifecycleStub{LNClient: svc.LNClient, waitErr: lnclient.ErrOfferPaymentUnknown}
			amount := uint64(1719000)
			service := NewTransactionsService(svc.DB, svc.EventPublisher)
			_, err = service.PayOfferSync(context.Background(), tests.MockOffer, &lnclient.OfferInfo{}, &amount, "attempt", nil, client, nil, nil, nil)
			require.ErrorIs(t, err, lnclient.ErrOfferPaymentUnknown)
			want := constants.TRANSACTION_STATE_SETTLED
			if failed {
				client.waitErr = lnclient.ErrOfferPaymentFailed
				want = constants.TRANSACTION_STATE_FAILED
			} else {
				client.waitErr = nil
				client.response = &lnclient.PayOfferResponse{PaymentHash: "paid-hash", Preimage: "preimage", FeeMsat: 1000}
			}
			rows, _, err := service.ListTransactions(context.Background(), 0, 0, 100, 0, true, true, client, nil, false, nil)
			require.NoError(t, err)
			require.Len(t, rows, 1)
			require.Equal(t, want, rows[0].State)
			require.Zero(t, rows[0].FeeReserveMsat)
		})
	}
}

func TestOfferLifecycleFailureAndUnknown(t *testing.T) {
	for _, tc := range []struct {
		name  string
		err   error
		state string
	}{
		{"failed", errors.New("invoice request expired"), constants.TRANSACTION_STATE_FAILED},
		{"unknown", lnclient.ErrOfferPaymentUnknown, constants.TRANSACTION_STATE_PENDING},
		{"missing response", nil, constants.TRANSACTION_STATE_PENDING},
	} {
		t.Run(tc.name, func(t *testing.T) {
			svc, err := tests.CreateTestService(t)
			require.NoError(t, err)
			defer svc.Remove()
			client := &offerLifecycleStub{LNClient: svc.LNClient, waitErr: tc.err}
			amount, fee := uint64(1719000), uint64(0)
			result, err := NewTransactionsService(svc.DB, svc.EventPublisher).PayOfferSync(context.Background(), tests.MockOffer, &lnclient.OfferInfo{}, &amount, "attempt", nil, client, nil, nil, &fee)
			require.Error(t, err)
			require.Nil(t, result)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			}
			require.NotNil(t, client.fee)
			require.Zero(t, *client.fee)
			var row db.Transaction
			require.NoError(t, svc.DB.Where("ln_client_payment_id = ?", "backend-payment-id").First(&row).Error)
			require.Equal(t, tc.state, row.State)
		})
	}
}
