package transactions

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/db/queries"
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
	LookupTransaction(ctx context.Context, paymentHash string, transactionType *string, lnClient lnclient.LNClient, appId *uint) (*Transaction, error)
	ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, transactionType *string, lnClient lnclient.LNClient, appId *uint) (transactions []Transaction, err error)
	SendPaymentSync(ctx context.Context, payReq string, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error)
	SendKeysend(ctx context.Context, amount uint64, destination string, customRecords []lnclient.TLVRecord, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error)
}

type Transaction = db.Transaction

type notFoundError struct {
}

func NewNotFoundError() error {
	return &notFoundError{}
}

func (err *notFoundError) Error() string {
	return "The transaction requested was not Found"
}

type insufficientBalanceError struct {
}

func NewInsufficientBalanceError() error {
	return &insufficientBalanceError{}
}

func (err *insufficientBalanceError) Error() string {
	return "Insufficient balance remaining to make the requested payment"
}

type quotaExceededError struct {
}

func NewQuotaExceededError() error {
	return &quotaExceededError{}
}

func (err *quotaExceededError) Error() string {
	return "Your wallet has exceeded its spending quota"
}

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

	dbTransaction := db.Transaction{
		AppId:           appId,
		RequestEventId:  requestEventId,
		Type:            lnClientTransaction.Type,
		State:           constants.TRANSACTION_STATE_PENDING,
		AmountMsat:      uint64(lnClientTransaction.Amount),
		Description:     description,
		DescriptionHash: descriptionHash,
		PaymentRequest:  lnClientTransaction.Invoice,
		PaymentHash:     lnClientTransaction.PaymentHash,
		ExpiresAt:       expiresAt,
		Preimage:        preimage,
		Metadata:        metadata,
	}
	err = svc.db.Create(&dbTransaction).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create DB transaction")
		return nil, err
	}
	return &dbTransaction, nil
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

	selfPayment := paymentRequest.Payee != "" && paymentRequest.Payee == lnClient.GetPubkey()

	var dbTransaction db.Transaction

	err = svc.db.Transaction(func(tx *gorm.DB) error {
		err := svc.validateCanPay(tx, appId, uint64(paymentRequest.MSatoshi))
		if err != nil {
			return err
		}

		var expiresAt *time.Time
		if paymentRequest.Expiry > 0 {
			expiresAtValue := time.Now().Add(time.Duration(paymentRequest.Expiry) * time.Second)
			expiresAt = &expiresAtValue
		}
		dbTransaction = db.Transaction{
			AppId:           appId,
			RequestEventId:  requestEventId,
			Type:            constants.TRANSACTION_TYPE_OUTGOING,
			State:           constants.TRANSACTION_STATE_PENDING,
			FeeReserveMsat:  svc.calculateFeeReserveMsat(uint64(paymentRequest.MSatoshi)),
			AmountMsat:      uint64(paymentRequest.MSatoshi),
			PaymentRequest:  payReq,
			PaymentHash:     paymentRequest.PaymentHash,
			Description:     paymentRequest.Description,
			DescriptionHash: paymentRequest.DescriptionHash,
			ExpiresAt:       expiresAt,
			SelfPayment:     selfPayment,
			// Metadata:       metadata,
		}
		err = tx.Create(&dbTransaction).Error
		return err
	})

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
		}).WithError(err).Error("Failed to create DB transaction")
		return nil, err
	}

	var response *lnclient.PayInvoiceResponse
	if selfPayment {
		response, err = svc.interceptSelfPayment(paymentRequest.PaymentHash)
	} else {
		response, err = lnClient.SendPaymentSync(ctx, payReq)
	}

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
		dbErr := svc.db.Model(&dbTransaction).Updates(map[string]interface{}{
			"State":          constants.TRANSACTION_STATE_FAILED,
			"FeeReserveMsat": 0,
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
	dbErr := svc.db.Model(&dbTransaction).Updates(map[string]interface{}{
		"State":          constants.TRANSACTION_STATE_SETTLED,
		"Preimage":       &response.Preimage,
		"FeeMsat":        response.Fee,
		"FeeReserveMsat": 0,
		"SettledAt":      &now,
	}).Error
	if dbErr != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
		}).WithError(dbErr).Error("Failed to update DB transaction")
	}

	// TODO: check the fields are updated here
	return &dbTransaction, nil
}

func (svc *transactionsService) SendKeysend(ctx context.Context, amount uint64, destination string, customRecords []lnclient.TLVRecord, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error) {

	metadata := map[string]interface{}{}

	metadata["destination"] = destination
	metadata["tlv_records"] = customRecords
	metadataBytes, err := json.Marshal(metadata)

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to marshal metadata")
		return nil, err
	}

	var dbTransaction db.Transaction

	err = svc.db.Transaction(func(tx *gorm.DB) error {
		err := svc.validateCanPay(tx, appId, amount)
		if err != nil {
			return err
		}

		// NOTE: transaction is created without payment hash :scream:
		dbTransaction = db.Transaction{
			AppId:          appId,
			RequestEventId: requestEventId,
			Type:           constants.TRANSACTION_TYPE_OUTGOING,
			State:          constants.TRANSACTION_STATE_PENDING,
			FeeReserveMsat: svc.calculateFeeReserveMsat(uint64(amount)),
			AmountMsat:     amount,
			Metadata:       string(metadataBytes),
		}
		err = tx.Create(&dbTransaction).Error

		return err
	})

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
			dbErr := svc.db.Model(&dbTransaction).Updates(&db.Transaction{
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
		dbErr := svc.db.Model(&dbTransaction).Updates(&db.Transaction{
			PaymentHash: paymentHash,
			State:       constants.TRANSACTION_STATE_FAILED,
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

	dbErr := svc.db.Transaction(func(tx *gorm.DB) error {
		// Because we don't know the payment hash in advance when creating the original keysend transaction,
		// there may have been another transaction created with the same payment hash by the nwc_payment_sent event.
		// Delete that one and update the one we already created.
		// This would not be needed if we always can pass the payment hash when doing keysend.
		err := tx.Delete(&db.Transaction{}, "payment_hash = ?", paymentHash).Error
		if err != nil {
			return err
		}

		err = tx.Model(&dbTransaction).Updates(map[string]interface{}{
			"State":          constants.TRANSACTION_STATE_SETTLED,
			"PaymentHash":    paymentHash,
			"Preimage":       &preimage,
			"FeeMsat":        &fee,
			"FeeReserveMsat": 0,
			"SettledAt":      &now,
		}).Error
		return err
	})

	if dbErr != nil {
		logger.Logger.WithFields(logrus.Fields{
			"destination": destination,
			"amount":      amount,
		}).WithError(dbErr).Error("Failed to update DB transaction")
	}

	// TODO: check the fields are updated here
	return &dbTransaction, nil
}

func (svc *transactionsService) LookupTransaction(ctx context.Context, paymentHash string, transactionType *string, lnClient lnclient.LNClient, appId *uint) (*Transaction, error) {
	transaction := db.Transaction{}

	tx := svc.db

	if appId != nil {
		var app db.App
		svc.db.Find(&app, &db.App{
			ID: *appId,
		})
		if app.Isolated {
			tx = tx.Where("app_id == ?", *appId)
		}
	}

	if transactionType != nil {
		tx = tx.Where("type == ?", *transactionType)
	}

	// order settled first, otherwise by created date, as there can be multiple outgoing payments
	// for the same payment hash (if you tried to pay an invoice multiple times - e.g. the first time failed)
	result := tx.Order("settled_at desc, created_at desc").Find(&transaction, &db.Transaction{
		//Type:        transactionType,
		PaymentHash: paymentHash,
	})

	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to lookup DB transaction")
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		logger.Logger.WithFields(logrus.Fields{
			"payment_hash": paymentHash,
			"app_id":       appId,
		}).WithError(result.Error).Error("Failed to lookup DB transaction")
		return nil, NewNotFoundError()
	}

	if transaction.State == constants.TRANSACTION_STATE_PENDING {
		svc.checkUnsettledTransaction(ctx, &transaction, lnClient)
	}

	return &transaction, nil
}

func (svc *transactionsService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, transactionType *string, lnClient lnclient.LNClient, appId *uint) (transactions []Transaction, err error) {
	svc.checkUnsettledTransactions(ctx, lnClient)

	// TODO: add other filtering and pagination
	tx := svc.db

	tx = tx.Order("created_at desc")

	if !unpaid {
		tx = tx.Where("state == ?", constants.TRANSACTION_STATE_SETTLED)
	}

	if transactionType != nil {
		tx = tx.Where("type == ?", *transactionType)
	}

	if from > 0 {
		tx = tx.Where("created_at >= ?", time.Unix(int64(from), 0))
	}
	if until > 0 {
		tx = tx.Where("created_at <= ?", time.Unix(int64(until), 0))
	}

	if limit > 0 {
		tx = tx.Limit(int(limit))
	}
	if offset > 0 {
		tx = tx.Offset(int(limit))
	}

	if appId != nil {
		var app db.App
		svc.db.Find(&app, &db.App{
			ID: *appId,
		})
		if app.Isolated {
			tx = tx.Where("app_id == ?", *appId)
		}
	}

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
	result := svc.db.Where("state == ? AND created_at > ?", constants.TRANSACTION_STATE_PENDING, time.Now().Add(-24*time.Hour)).Find(&transactions)
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
		dbErr := svc.db.Model(transaction).Updates(map[string]interface{}{
			"State":          constants.TRANSACTION_STATE_SETTLED,
			"Preimage":       &lnClientTransaction.Preimage,
			"FeeMsat":        lnClientTransaction.FeesPaid,
			"FeeReserveMsat": 0,
			"SettledAt":      &now,
		}).Error
		if dbErr != nil {
			logger.Logger.WithFields(logrus.Fields{
				"bolt11": transaction.PaymentRequest,
			}).WithError(dbErr).Error("Failed to update DB transaction")
		}
	}
}

func (svc *transactionsService) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	switch event.Event {
	case "nwc_payment_received":
		lnClientTransaction, ok := event.Properties.(*lnclient.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return
		}

		var dbTransaction db.Transaction
		err := svc.db.Transaction(func(tx *gorm.DB) error {

			result := tx.Find(&dbTransaction, &db.Transaction{
				Type:        constants.TRANSACTION_TYPE_INCOMING,
				PaymentHash: lnClientTransaction.PaymentHash,
			})

			if result.RowsAffected == 0 {
				// Note: brand new payments cannot be associated with an app
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
				dbTransaction = db.Transaction{
					Type:            constants.TRANSACTION_TYPE_INCOMING,
					AmountMsat:      uint64(lnClientTransaction.Amount),
					PaymentRequest:  lnClientTransaction.Invoice,
					PaymentHash:     lnClientTransaction.PaymentHash,
					Description:     lnClientTransaction.Description,
					DescriptionHash: lnClientTransaction.DescriptionHash,
					ExpiresAt:       expiresAt,
					Metadata:        metadata,
				}
				err := tx.Create(&dbTransaction).Error
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"payment_hash": lnClientTransaction.PaymentHash,
					}).WithError(err).Error("Failed to create transaction")
					return err
				}
			}

			settledAt := time.Now()

			err := tx.Model(&dbTransaction).Updates(map[string]interface{}{
				"FeeMsat":   lnClientTransaction.FeesPaid,
				"Preimage":  &lnClientTransaction.Preimage,
				"State":     constants.TRANSACTION_STATE_SETTLED,
				"SettledAt": &settledAt,
			}).Error
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"payment_hash": lnClientTransaction.PaymentHash,
				}).WithError(err).Error("Failed to update transaction")
				return err
			}

			return nil
		})

		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": lnClientTransaction.PaymentHash,
			}).WithError(err).Error("Failed to execute DB transaction")
			return
		}
		logger.Logger.WithField("id", dbTransaction.ID).Info("Marked incoming transaction as settled")
	case "nwc_payment_sent":
		lnClientTransaction, ok := event.Properties.(*lnclient.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return
		}

		var dbTransaction db.Transaction
		err := svc.db.Transaction(func(tx *gorm.DB) error {
			result := tx.Find(&dbTransaction, &db.Transaction{
				Type:        constants.TRANSACTION_TYPE_OUTGOING,
				PaymentHash: lnClientTransaction.PaymentHash,
			})

			if result.RowsAffected == 0 {
				// Note: brand new payments cannot be associated with an app
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
				dbTransaction = db.Transaction{
					Type:            constants.TRANSACTION_TYPE_OUTGOING,
					AmountMsat:      uint64(lnClientTransaction.Amount),
					PaymentRequest:  lnClientTransaction.Invoice,
					PaymentHash:     lnClientTransaction.PaymentHash,
					Description:     lnClientTransaction.Description,
					DescriptionHash: lnClientTransaction.DescriptionHash,
					ExpiresAt:       expiresAt,
					Metadata:        metadata,
				}
				err := tx.Create(&dbTransaction).Error
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"payment_hash": lnClientTransaction.PaymentHash,
					}).WithError(err).Error("Failed to create transaction")
					return err
				}
			}

			settledAt := time.Now()
			err := tx.Model(&dbTransaction).Updates(map[string]interface{}{
				"FeeMsat":        lnClientTransaction.FeesPaid,
				"FeeReserveMsat": 0,
				"Preimage":       &lnClientTransaction.Preimage,
				"State":          constants.TRANSACTION_STATE_SETTLED,
				"SettledAt":      &settledAt,
			}).Error
			return err
		})

		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": lnClientTransaction.PaymentHash,
			}).WithError(err).Error("Failed to update transaction")
			return
		}
		logger.Logger.WithField("id", dbTransaction.ID).Info("Marked outgoing transaction as settled")

	case "nwc_payment_failed_async":
		paymentFailedAsyncProperties, ok := event.Properties.(*events.PaymentFailedAsyncProperties)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return
		}

		lnClientTransaction := paymentFailedAsyncProperties.Transaction

		var dbTransaction db.Transaction
		result := svc.db.Find(&dbTransaction, &db.Transaction{
			Type:        constants.TRANSACTION_TYPE_OUTGOING,
			PaymentHash: lnClientTransaction.PaymentHash,
		})

		// Note: this will happen for keysend payments since our transaction entry will not have a payment
		// hash at this point
		if result.RowsAffected == 0 {
			logger.Logger.WithField("event", event).Error("Failed to find outgoing transaction by payment hash")
			return
		}

		err := svc.db.Model(&dbTransaction).Updates(map[string]interface{}{
			"State":          constants.TRANSACTION_STATE_FAILED,
			"FeeReserveMsat": 0,
		}).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": lnClientTransaction.PaymentHash,
			}).WithError(err).Error("Failed to update transaction")
			return
		}
		logger.Logger.WithField("id", dbTransaction.ID).Info("Marked outgoing transaction as failed")
	}
}

func (svc *transactionsService) interceptSelfPayment(paymentHash string) (*lnclient.PayInvoiceResponse, error) {
	// TODO: extract into separate function
	incomingTransaction := db.Transaction{}
	result := svc.db.Find(&incomingTransaction, &db.Transaction{
		Type:        constants.TRANSACTION_TYPE_INCOMING,
		State:       constants.TRANSACTION_STATE_PENDING,
		PaymentHash: paymentHash,
	})
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, NewNotFoundError()
	}
	if incomingTransaction.Preimage == nil {
		return nil, errors.New("preimage is not set on transaction. Self payments not supported")
	}

	// update the incoming transaction
	now := time.Now()
	err := svc.db.Model(&incomingTransaction).Updates(map[string]interface{}{
		"State":       constants.TRANSACTION_STATE_SETTLED,
		"SettledAt":   &now,
		"SelfPayment": true,
	}).Error
	if err != nil {
		return nil, err
	}

	// TODO: publish event for self payment

	return &lnclient.PayInvoiceResponse{
		Preimage: *incomingTransaction.Preimage,
		Fee:      0,
	}, nil
}

func (svc *transactionsService) validateCanPay(tx *gorm.DB, appId *uint, amount uint64) error {
	amountWithFeeReserve := amount + svc.calculateFeeReserveMsat(amount)

	// ensure balance for isolated apps
	if appId != nil {
		var app db.App
		tx.Find(&app, &db.App{
			ID: *appId,
		})
		var appPermission db.AppPermission
		result := tx.Find(&appPermission, &db.AppPermission{
			AppId: *appId,
			Scope: constants.PAY_INVOICE_SCOPE,
		})
		if result.RowsAffected == 0 {
			return errors.New("app does not have pay_invoice scope")
		}

		if app.Isolated {
			balance := queries.GetIsolatedBalance(tx, appPermission.AppId)

			if amountWithFeeReserve > balance {
				return NewInsufficientBalanceError()
			}
		}

		if appPermission.MaxAmountSat > 0 {
			budgetUsageSat := queries.GetBudgetUsageSat(tx, &appPermission)
			if int(amountWithFeeReserve/1000) > appPermission.MaxAmountSat-int(budgetUsageSat) {
				return NewQuotaExceededError()
			}
		}
	}

	return nil
}

// max of 1% or 10000 millisats (10 sats)
func (svc *transactionsService) calculateFeeReserveMsat(amount uint64) uint64 {
	// NOTE: LDK defaults to 1% of the payment amount + 50 sats
	return uint64(math.Max(math.Ceil(float64(amount)*0.01), 10000))
}
