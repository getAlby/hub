package ldkserver

import (
	"context"
	"fmt"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	ldkapi "github.com/getAlby/hub/lnclient/ldk-server/grpc/api"
	ldkevents "github.com/getAlby/hub/lnclient/ldk-server/grpc/events"
	ldktypes "github.com/getAlby/hub/lnclient/ldk-server/grpc/types"
	"github.com/getAlby/hub/nip47/models"
)

func TestStartOfferPaymentForwardsFeeLimit(t *testing.T) {
	for _, limit := range []uint64{0, 10000} {
		t.Run(fmt.Sprint(limit), func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, ldkapi.LightningNode_Bolt12Send_FullMethodName, r.URL.Path)
				body, err := io.ReadAll(r.Body)
				require.NoError(t, err)
				payload, err := decodeSingleFrame(body)
				require.NoError(t, err)
				var request ldkapi.Bolt12SendRequest
				require.NoError(t, proto.Unmarshal(payload, &request))
				require.Equal(t, "offer", request.Offer)
				require.EqualValues(t, 1719000, request.GetAmountMsat())
				require.Equal(t, "attempt", request.GetPayerNote())
				require.NotNil(t, request.RouteParameters)
				require.NotNil(t, request.RouteParameters.MaxTotalRoutingFeeMsat)
				require.Equal(t, limit, *request.RouteParameters.MaxTotalRoutingFeeMsat)
				require.EqualValues(t, 10, request.RouteParameters.MaxPathCount)
				response, err := proto.Marshal(&ldkapi.Bolt12SendResponse{PaymentId: "payment-id"})
				require.NoError(t, err)
				w.Header().Set("Grpc-Status", "0")
				_, _ = w.Write(grpcFrame(response))
			}))
			defer server.Close()
			svc := &LDKServerService{baseURL: server.URL, client: server.Client()}
			amount := uint64(1719000)
			id, err := svc.StartOfferPaymentWithFeeLimit(context.Background(), "offer", &amount, "attempt", limit)
			require.NoError(t, err)
			require.Equal(t, "payment-id", id)
		})
	}
}

func TestOfferPaymentResultDistinguishesFailureFromUnknown(t *testing.T) {
	for _, tc := range []struct {
		status ldktypes.PaymentStatus
		want   error
	}{
		{ldktypes.PaymentStatus_PENDING, lnclient.ErrOfferPaymentUnknown},
		{ldktypes.PaymentStatus_FAILED, lnclient.ErrOfferPaymentFailed},
		// Success without proof is not evidence that a payment failed.
		{ldktypes.PaymentStatus_SUCCEEDED, lnclient.ErrOfferPaymentUnknown},
	} {
		payment := &ldktypes.Payment{
			Status: tc.status, Direction: ldktypes.PaymentDirection_OUTBOUND,
			Kind: &ldktypes.PaymentKind{Kind: &ldktypes.PaymentKind_Bolt12Offer{Bolt12Offer: &ldktypes.Bolt12Offer{}}},
		}
		result, err := offerPaymentResult(payment)
		require.Nil(t, result)
		require.ErrorIs(t, err, tc.want)
	}
}

func TestOnchainBalanceExcludesOpenChannels(t *testing.T) {
	for _, usable := range []bool{true, false} {
		resp := &ldkapi.GetBalancesResponse{
			TotalOnchainBalanceSats:        30000,
			SpendableOnchainBalanceSats:    5000,
			TotalAnchorChannelsReserveSats: 25000,
			LightningBalances: []*ldktypes.LightningBalance{{
				BalanceType: &ldktypes.LightningBalance_ClaimableOnChannelClose{
					ClaimableOnChannelClose: &ldktypes.ClaimableOnChannelClose{
						ChannelId: "open", CounterpartyNodeId: "peer", AmountSatoshis: 9756,
					},
				},
			}},
		}
		result := onchainBalanceResponse(resp, []*ldktypes.Channel{{ChannelId: "open", IsUsable: usable}})
		require.EqualValues(t, 30000, result.TotalSat)
		require.EqualValues(t, 5000, result.SpendableSat)
		require.EqualValues(t, 25000, result.ReservedSat)
		require.Zero(t, result.PendingBalancesFromChannelClosuresSat)
		require.Empty(t, result.PendingBalancesDetails)
		require.Empty(t, result.PendingSweepBalancesDetails)
	}
}

func TestOnchainBalancePreservesClosingChannels(t *testing.T) {
	resp := &ldkapi.GetBalancesResponse{
		LightningBalances: []*ldktypes.LightningBalance{
			{BalanceType: &ldktypes.LightningBalance_ClaimableOnChannelClose{
				ClaimableOnChannelClose: &ldktypes.ClaimableOnChannelClose{
					ChannelId: "closing", CounterpartyNodeId: "peer", AmountSatoshis: 9756,
				},
			}},
			{BalanceType: &ldktypes.LightningBalance_ClaimableAwaitingConfirmations{
				ClaimableAwaitingConfirmations: &ldktypes.ClaimableAwaitingConfirmations{
					ChannelId: "confirming", CounterpartyNodeId: "peer", AmountSatoshis: 2000,
				},
			}},
			{BalanceType: &ldktypes.LightningBalance_ContentiousClaimable{
				ContentiousClaimable: &ldktypes.ContentiousClaimable{
					ChannelId: "contentious", CounterpartyNodeId: "peer", AmountSatoshis: 3000,
				},
			}},
		},
	}
	result := onchainBalanceResponse(resp, nil)
	require.EqualValues(t, 14756, result.PendingBalancesFromChannelClosuresSat)
	require.Equal(t, []lnclient.PendingBalanceDetails{
		{ChannelId: "closing", NodeId: "peer", AmountSat: 9756},
		{ChannelId: "confirming", NodeId: "peer", AmountSat: 2000},
		{ChannelId: "contentious", NodeId: "peer", AmountSat: 3000},
	}, result.PendingBalancesDetails)
}

type recordingEventPublisher struct {
	published     []*events.Event
	syncPublished []*events.Event
}

func (p *recordingEventPublisher) RegisterSubscriber(events.EventSubscriber) {}
func (p *recordingEventPublisher) RemoveSubscriber(events.EventSubscriber)   {}
func (p *recordingEventPublisher) SetGlobalProperty(string, interface{})     {}

func (p *recordingEventPublisher) Publish(event *events.Event) {
	p.published = append(p.published, event)
}

func (p *recordingEventPublisher) PublishSync(event *events.Event) {
	p.syncPublished = append(p.syncPublished, event)
}

func TestSupportedNIP47MethodsIncludePayAndReceive(t *testing.T) {
	svc := &LDKServerService{}

	methods := svc.GetSupportedNIP47Methods()

	require.Contains(t, methods, models.PAY_METHOD)
	require.Contains(t, methods, models.RECEIVE_METHOD)
}

func TestHandlePaymentClaimablePublishesDeadlineSynchronously(t *testing.T) {
	publisher := &recordingEventPublisher{}
	svc := &LDKServerService{eventPublisher: publisher}
	claimDeadline := uint32(840_000)
	paymentHash := "0001020304050607080900010203040506070809000102030405060708090001"

	svc.handleEvent(&ldkevents.EventEnvelope{
		Event: &ldkevents.EventEnvelope_PaymentClaimable{
			PaymentClaimable: &ldkevents.PaymentClaimable{
				Payment: &ldktypes.Payment{
					Kind: &ldktypes.PaymentKind{
						Kind: &ldktypes.PaymentKind_Bolt11{
							Bolt11: &ldktypes.Bolt11{Hash: paymentHash},
						},
					},
					Direction: ldktypes.PaymentDirection_INBOUND,
					Status:    ldktypes.PaymentStatus_PENDING,
				},
				ClaimDeadline: &claimDeadline,
			},
		},
	})

	require.Empty(t, publisher.published)
	require.Len(t, publisher.syncPublished, 1)
	require.Equal(t, "nwc_lnclient_hold_invoice_accepted", publisher.syncPublished[0].Event)

	transaction, ok := publisher.syncPublished[0].Properties.(*lnclient.Transaction)
	require.True(t, ok)
	require.Equal(t, paymentHash, transaction.PaymentHash)
	require.Equal(t, &claimDeadline, transaction.SettleDeadline)
}
