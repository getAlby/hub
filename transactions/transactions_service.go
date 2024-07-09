package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"time"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type transactionsService struct {
	db *gorm.DB
}

type TransactionsService interface {
	events.EventSubscriber
	MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error)
	LookupTransaction(ctx context.Context, paymentHash string, lnClient lnclient.LNClient, appId *uint) (*Transaction, error)
	ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string, lnClient lnclient.LNClient) (transactions []Transaction, err error)
	SendPaymentSync(ctx context.Context, payReq string, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error)
	SendKeysend(ctx context.Context, amount uint64, destination string, customRecords []lnclient.TLVRecord, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error)
}

type Transaction = db.Transaction

const (
	TRANSACTION_TYPE_INCOMING = "incoming"
	TRANSACTION_TYPE_OUTGOING = "outgoing"

	TRANSACTION_STATE_PENDING = "PENDING"
	TRANSACTION_STATE_SETTLED = "SETTLED"
	TRANSACTION_STATE_FAILED  = "FAILED"
)

func NewTransactionsService(db *gorm.DB) *transactionsService {
	return &transactionsService{
		db: db,
	}
}

func (svc *transactionsService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error) {
	lnClientTransaction, err := lnClient.MakeInvoice(ctx, amount, description, descriptionHash, expiry)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create transaction")
		return nil, err
	}

	var preimage *string
	if lnClientTransaction.Preimage != "" {
		preimage = &lnClientTransaction.Preimage
	}

	var metadata string
	if lnClientTransaction.Metadata != nil {
		metadataBytes, err := json.Marshal(lnClientTransaction.Metadata)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to serialize transaction metadata")
			return nil, err
		}
		metadata = string(metadataBytes)
	}

	var expiresAt *time.Time
	if lnClientTransaction.ExpiresAt != nil {
		expiresAtValue := time.Unix(*lnClientTransaction.ExpiresAt, 0)
		expiresAt = &expiresAtValue
	}

	dbTransaction := &db.Transaction{
		AppId:           appId,
		RequestEventId:  requestEventId,
		Type:            lnClientTransaction.Type,
		State:           TRANSACTION_STATE_PENDING,
		Amount:          uint64(lnClientTransaction.Amount),
		Description:     description,
		DescriptionHash: descriptionHash,
		PaymentRequest:  lnClientTransaction.Invoice,
		PaymentHash:     lnClientTransaction.PaymentHash,
		ExpiresAt:       expiresAt,
		Preimage:        preimage,
		Metadata:        metadata,
	}
	err = svc.db.Create(dbTransaction).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create DB transaction")
		return nil, err
	}
	return dbTransaction, nil
}

func (svc *transactionsService) SendPaymentSync(ctx context.Context, payReq string, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error) {
	payReq = strings.ToLower(payReq)
	paymentRequest, err := decodepay.Decodepay(payReq)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		return nil, err
	}

	// TODO: in transaction, ensure budget
	var expiresAt *time.Time
	if paymentRequest.Expiry > 0 {
		expiresAtValue := time.Now().Add(time.Duration(paymentRequest.Expiry) * time.Second)
		expiresAt = &expiresAtValue
	}
	dbTransaction := &db.Transaction{
		AppId:           appId,
		RequestEventId:  requestEventId,
		Type:            TRANSACTION_TYPE_OUTGOING,
		State:           TRANSACTION_STATE_PENDING,
		Amount:          uint64(paymentRequest.MSatoshi),
		PaymentRequest:  payReq,
		PaymentHash:     paymentRequest.PaymentHash,
		Description:     paymentRequest.Description,
		DescriptionHash: paymentRequest.DescriptionHash,
		ExpiresAt:       expiresAt,
		// Metadata:       metadata,
	}
	err = svc.db.Create(dbTransaction).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
		}).WithError(err).Error("Failed to create DB transaction")
		return nil, err
	}

	response, err := lnClient.SendPaymentSync(ctx, payReq)

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
		}).WithError(err).Error("Failed to send payment")

		if errors.Is(err, lnclient.NewTimeoutError()) {
			logger.Logger.WithFields(logrus.Fields{
				"bolt11": payReq,
			}).WithError(err).Error("Timed out waiting for payment to be sent. It may still succeed. Skipping update of transaction status")
			// we cannot update the payment to failed as it still might succeed.
			// we'll need to check the status of it later
			return nil, err
		}

		// As the LNClient did not return a timeout error, we assume the payment definitely failed
		dbErr := svc.db.Model(dbTransaction).Updates(&db.Transaction{
			State: TRANSACTION_STATE_FAILED,
		}).Error
		if dbErr != nil {
			logger.Logger.WithFields(logrus.Fields{
				"bolt11": payReq,
			}).WithError(dbErr).Error("Failed to update DB transaction")
		}

		return nil, err
	}

	// the payment definitely succeeded
	now := time.Now()
	dbErr := svc.db.Model(dbTransaction).Updates(&db.Transaction{
		State:     TRANSACTION_STATE_SETTLED,
		Preimage:  &response.Preimage,
		Fee:       response.Fee,
		SettledAt: &now,
	}).Error
	if dbErr != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
		}).WithError(dbErr).Error("Failed to update DB transaction")
	}

	// TODO: check the fields are updated here
	return dbTransaction, nil
}

func (svc *transactionsService) SendKeysend(ctx context.Context, amount uint64, destination string, customRecords []lnclient.TLVRecord, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error) {
	// TODO: in transaction, ensure budget

	metadata := map[string]interface{}{}

	metadata["destination"] = destination
	metadata["tlv_records"] = customRecords
	metadataBytes, err := json.Marshal(metadata)

	// NOTE: transaction is created without payment hash :scream:
	dbTransaction := &db.Transaction{
		AppId:          appId,
		RequestEventId: requestEventId,
		Type:           TRANSACTION_TYPE_OUTGOING,
		State:          TRANSACTION_STATE_PENDING,
		Amount:         amount,
		Metadata:       string(metadataBytes),
	}
	err = svc.db.Create(dbTransaction).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"destination": destination,
			"amount":      amount,
		}).WithError(err).Error("Failed to create DB transaction")
		return nil, err
	}

	paymentHash, preimage, fee, err := lnClient.SendKeysend(ctx, amount, destination, customRecords)

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"destination": destination,
			"amount":      amount,
		}).WithError(err).Error("Failed to send payment")

		if errors.Is(err, lnclient.NewTimeoutError()) {

			logger.Logger.WithFields(logrus.Fields{
				"destination": destination,
				"amount":      amount,
			}).WithError(err).Error("Timed out waiting for payment to be sent. It may still succeed. Skipping update of transaction status")

			// we cannot update the payment to failed as it still might succeed.
			// we'll need to check the status of it later
			// but we have the payment hash now, so save it on the transaction
			dbErr := svc.db.Model(dbTransaction).Updates(&db.Transaction{
				PaymentHash: paymentHash,
			}).Error
			if dbErr != nil {
				logger.Logger.WithFields(logrus.Fields{
					"destination": destination,
					"amount":      amount,
				}).WithError(dbErr).Error("Failed to update DB transaction")
			}
			return nil, err
		}

		// As the LNClient did not return a timeout error, we assume the payment definitely failed
		dbErr := svc.db.Model(dbTransaction).Updates(&db.Transaction{
			PaymentHash: paymentHash,
			State:       TRANSACTION_STATE_FAILED,
		}).Error
		if dbErr != nil {
			logger.Logger.WithFields(logrus.Fields{
				"destination": destination,
				"amount":      amount,
			}).WithError(dbErr).Error("Failed to update DB transaction")
		}

		return nil, err
	}

	// the payment definitely succeeded
	now := time.Now()
	dbErr := svc.db.Model(dbTransaction).Updates(&db.Transaction{
		State:       TRANSACTION_STATE_SETTLED,
		PaymentHash: paymentHash,
		Preimage:    &preimage,
		Fee:         &fee,
		SettledAt:   &now,
	}).Error
	if dbErr != nil {
		logger.Logger.WithFields(logrus.Fields{
			"destination": destination,
			"amount":      amount,
		}).WithError(dbErr).Error("Failed to update DB transaction")
	}

	// TODO: check the fields are updated here
	return dbTransaction, nil
}

func (svc *transactionsService) LookupTransaction(ctx context.Context, paymentHash string, lnClient lnclient.LNClient, appId *uint) (*Transaction, error) {
	transaction := db.Transaction{}

	// FIXME: this is currently not unique
	result := svc.db.Find(&transaction, &db.Transaction{
		//Type:        transactionType,
		PaymentHash: paymentHash,
		AppId:       appId,
	})

	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to lookup DB transaction")
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("transaction not found")
	}

	if transaction.State == TRANSACTION_STATE_PENDING {
		svc.checkUnsettledTransaction(ctx, &transaction, lnClient)
	}

	return &transaction, nil
}

func (svc *transactionsService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, transactionType string, lnClient lnclient.LNClient) (transactions []Transaction, err error) {
	svc.checkUnsettledTransactions(ctx, lnClient)

	// TODO: add other filtering and pagination
	tx := svc.db
	if !unpaid {
		tx = tx.Where("state == ?", TRANSACTION_STATE_SETTLED)
	}

	tx = tx.Order("created_at desc")

	if limit != 0 {
		tx = tx.Limit(int(limit))
	}
	result := tx.Find(&transactions)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to list DB transactions")
		return nil, result.Error
	}

	return transactions, nil
}

func (svc *transactionsService) checkUnsettledTransactions(ctx context.Context, lnClient lnclient.LNClient) {
	// Only check unsettled transactions for clients that don't support async events
	// checkUnsettledTransactions does not work for keysend payments!
	if slices.Contains(lnClient.GetSupportedNIP47NotificationTypes(), "payment_received") {
		return
	}

	// check pending payments less than a day old
	transactions := []Transaction{}
	result := svc.db.Where("state == ? AND created_at > ?", TRANSACTION_STATE_PENDING, time.Now().Add(-24*time.Hour)).Find(&transactions)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to list DB transactions")
		return
	}
	for _, transaction := range transactions {
		svc.checkUnsettledTransaction(ctx, &transaction, lnClient)
	}
}
func (svc *transactionsService) checkUnsettledTransaction(ctx context.Context, transaction *db.Transaction, lnClient lnclient.LNClient) {
	if slices.Contains(lnClient.GetSupportedNIP47NotificationTypes(), "payment_received") {
		return
	}

	lnClientTransaction, err := lnClient.LookupInvoice(ctx, transaction.PaymentHash)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": transaction.PaymentRequest,
		}).WithError(err).Error("Failed to check transaction")
		return
	}
	// update transaction state
	if lnClientTransaction.SettledAt != nil {
		// the payment definitely succeeded
		now := time.Now()
		fee := uint64(lnClientTransaction.FeesPaid)
		dbErr := svc.db.Model(transaction).Updates(&db.Transaction{
			State:     TRANSACTION_STATE_SETTLED,
			Preimage:  &lnClientTransaction.Preimage,
			Fee:       &fee,
			SettledAt: &now,
		}).Error
		if dbErr != nil {
			logger.Logger.WithFields(logrus.Fields{
				"bolt11": transaction.PaymentRequest,
			}).WithError(dbErr).Error("Failed to update DB transaction")
		}
	}
}

func (svc *transactionsService) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) error {
	switch event.Event {
	case "nwc_payment_received":
		lnClientTransaction, ok := event.Properties.(*lnclient.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return errors.New("failed to cast event")
		}

		err := svc.db.Transaction(func(tx *gorm.DB) error {
			var dbTransaction *db.Transaction

			result := tx.Find(dbTransaction, &db.Transaction{
				Type:        TRANSACTION_TYPE_INCOMING,
				PaymentHash: lnClientTransaction.PaymentHash,
			})

			if result.RowsAffected == 0 {
				// Note: brand new payments (keysend only) cannot be associated with an app
				var metadata string
				if lnClientTransaction.Metadata != nil {
					metadataBytes, err := json.Marshal(lnClientTransaction.Metadata)
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to serialize transaction metadata")
						return err
					}
					metadata = string(metadataBytes)
				}
				var expiresAt *time.Time
				if lnClientTransaction.ExpiresAt != nil {
					expiresAtValue := time.Unix(*lnClientTransaction.ExpiresAt, 0)
					expiresAt = &expiresAtValue
				}
				dbTransaction = &db.Transaction{
					Type:            TRANSACTION_TYPE_INCOMING,
					Amount:          uint64(lnClientTransaction.Amount),
					PaymentRequest:  lnClientTransaction.Invoice,
					PaymentHash:     lnClientTransaction.PaymentHash,
					Description:     lnClientTransaction.Description,
					DescriptionHash: lnClientTransaction.DescriptionHash,
					ExpiresAt:       expiresAt,
					Metadata:        metadata,
				}
				err := tx.Create(dbTransaction).Error
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"payment_hash": lnClientTransaction.PaymentHash,
					}).WithError(err).Error("Failed to create transaction")
					return err
				}
			}

			settledAt := time.Now()
			fee := uint64(lnClientTransaction.FeesPaid)

			err := tx.Model(dbTransaction).Updates(&db.Transaction{
				Fee:       &fee,
				Preimage:  &lnClientTransaction.Preimage,
				State:     TRANSACTION_STATE_SETTLED,
				SettledAt: &settledAt,
			}).Error
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"payment_hash": lnClientTransaction.PaymentHash,
				}).WithError(err).Error("Failed to update transaction")
				return err
			}

			logger.Logger.WithField("id", dbTransaction.ID).Info("Marked incoming transaction as settled")
			return nil
		})

		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": lnClientTransaction.PaymentHash,
			}).WithError(err).Error("Failed to execute DB transaction")
			return err
		}
	case "nwc_payment_sent":
		lnClientTransaction, ok := event.Properties.(*lnclient.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return errors.New("failed to cast event")
		}

		var dbTransaction *db.Transaction
		result := svc.db.Find(&dbTransaction, &db.Transaction{
			Type:        TRANSACTION_TYPE_OUTGOING,
			PaymentHash: lnClientTransaction.PaymentHash,
		})

		if result.RowsAffected == 0 {
			logger.Logger.WithField("event", event).Error("Failed to find outgoing transaction by payment hash")
			return errors.New("could not find outgoing transaction by payment hash")
		}

		settledAt := time.Now()
		fee := uint64(lnClientTransaction.FeesPaid)
		err := svc.db.Model(dbTransaction).Updates(&db.Transaction{
			Fee:       &fee,
			Preimage:  &lnClientTransaction.Preimage,
			State:     TRANSACTION_STATE_SETTLED,
			SettledAt: &settledAt,
		}).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": lnClientTransaction.PaymentHash,
			}).WithError(err).Error("Failed to update transaction")
			return err
		}
		logger.Logger.WithField("id", dbTransaction.ID).Info("Marked outgoing transaction as settled")
		// TODO: support nwc_payment_failed
	}

	return nil
}
