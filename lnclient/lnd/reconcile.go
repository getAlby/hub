package lnd

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
)

// reconcileLookback bounds the startup reconciliation window. It mirrors the
// 24h cap used by the transactions service when polling unsettled payments,
// so a restart only replays recent terminal activity instead of full history.
const reconcileLookback = 24 * time.Hour

const reconcilePageSize = 100

type paymentsLister interface {
	ListPayments(ctx context.Context, req *lnrpc.ListPaymentsRequest, options ...grpc.CallOption) (*lnrpc.ListPaymentsResponse, error)
}

type invoicesLister interface {
	ListInvoices(ctx context.Context, req *lnrpc.ListInvoiceRequest, options ...grpc.CallOption) (*lnrpc.ListInvoiceResponse, error)
}

type eventPublisher interface {
	Publish(event *events.Event)
}

// reconcileMissedPayments replays terminal LND payments and invoices that
// settled while the hub was offline and the notification streams were down.
// It publishes the same events the subscriptions publish, so the transactions
// service deduplicates them through its existing pending-first lookups.
func (svc *LNDService) reconcileMissedPayments(ctx context.Context) error {
	return reconcileMissedEvents(ctx, svc.client, svc.client, svc.eventPublisher, time.Now().Add(-reconcileLookback))
}

func reconcileMissedEvents(ctx context.Context, payments paymentsLister, invoices invoicesLister, publisher eventPublisher, since time.Time) error {
	return errors.Join(
		publishTerminalPayments(ctx, payments, publisher, since),
		publishSettledInvoices(ctx, invoices, publisher, since),
	)
}

func publishTerminalPayments(ctx context.Context, payments paymentsLister, publisher eventPublisher, since time.Time) error {
	var indexOffset uint64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		resp, err := payments.ListPayments(ctx, &lnrpc.ListPaymentsRequest{
			IncludeIncomplete: false,
			IndexOffset:       indexOffset,
			MaxPayments:       reconcilePageSize,
			Reversed:          false,
			CreationDateStart: uint64(since.Unix()),
			OmitHops:          true,
		})
		if err != nil {
			return fmt.Errorf("list payments for reconciliation: %w", err)
		}
		for _, payment := range resp.Payments {
			switch payment.Status {
			case lnrpc.Payment_SUCCEEDED:
				transaction, err := lndPaymentToTransaction(payment)
				if err != nil {
					continue
				}
				logger.Logger.WithField("payment_hash", transaction.PaymentHash).Info("Reconciling missed settled payment")
				publisher.Publish(&events.Event{
					Event:      "nwc_lnclient_payment_sent",
					Properties: transaction,
				})
			case lnrpc.Payment_FAILED:
				transaction, err := lndPaymentToTransaction(payment)
				if err != nil {
					continue
				}
				logger.Logger.WithField("payment_hash", transaction.PaymentHash).Info("Reconciling missed failed payment")
				publisher.Publish(&events.Event{
					Event: "nwc_lnclient_payment_failed",
					Properties: &lnclient.PaymentFailedEventProperties{
						Transaction: transaction,
						Reason:      payment.FailureReason.String(),
					},
				})
			default:
				continue
			}
		}
		if len(resp.Payments) < reconcilePageSize || resp.LastIndexOffset == indexOffset {
			return nil
		}
		indexOffset = resp.LastIndexOffset
	}
}

func publishSettledInvoices(ctx context.Context, invoices invoicesLister, publisher eventPublisher, since time.Time) error {
	var indexOffset uint64
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		// lnd has no server-side settlement-time filter: CreationDateStart
		// filters by creation date, so settled invoices are filtered by
		// SettleDate here after scanning all invoices.
		resp, err := invoices.ListInvoices(ctx, &lnrpc.ListInvoiceRequest{
			IndexOffset:    indexOffset,
			NumMaxInvoices: reconcilePageSize,
			Reversed:       false,
		})
		if err != nil {
			return fmt.Errorf("list invoices for reconciliation: %w", err)
		}
		for _, invoice := range resp.Invoices {
			if invoice.State != lnrpc.Invoice_SETTLED || invoice.SettleDate < since.Unix() {
				continue
			}
			transaction := lndInvoiceToTransaction(invoice)
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": transaction.PaymentHash,
			}).Info("Reconciling missed settled invoice")
			publisher.Publish(&events.Event{
				Event:      "nwc_lnclient_payment_received",
				Properties: transaction,
			})
		}
		if len(resp.Invoices) < reconcilePageSize || resp.LastIndexOffset == indexOffset {
			return nil
		}
		indexOffset = resp.LastIndexOffset
	}
}
