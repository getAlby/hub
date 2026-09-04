package ldkserver

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	ldkevents "github.com/getAlby/hub/lnclient/ldk-server/grpc/events"
	ldktypes "github.com/getAlby/hub/lnclient/ldk-server/grpc/types"
)

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
