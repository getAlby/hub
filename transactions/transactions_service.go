package transactions

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/db/queries"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
)

type transactionsService struct {
	db             *gorm.DB
	eventPublisher events.EventPublisher
}

type TransactionsService interface {
	events.EventSubscriber
	MakeInvoice(ctx context.Context, amount uint64, description string, descriptionHash string, expiry uint64, metadata map[string]interface{}, lnClient lnclient.LNClient, appId *uint, requestEventId *uint, throughNodePubkey *string) (*Transaction, error)
	LookupTransaction(ctx context.Context, paymentHash string, transactionType *string, lnClient lnclient.LNClient, appId *uint) (*Transaction, error)
	ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaidOutgoing bool, unpaidIncoming bool, transactionType *string, lnClient lnclient.LNClient, appId *uint, forceFilterByAppId bool) (transactions []Transaction, totalCount uint64, err error)
	SendPaymentSync(payReq string, amountMsat *uint64, metadata map[string]interface{}, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error)
	SendKeysend(amount uint64, destination string, customRecords []lnclient.TLVRecord, preimage string, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error)
	MakeHoldInvoice(ctx context.Context, amount uint64, description string, descriptionHash string, expiry uint64, paymentHash string, metadata map[string]interface{}, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error)
	SettleHoldInvoice(ctx context.Context, preimage string, lnClient lnclient.LNClient) (*Transaction, error)
	CancelHoldInvoice(ctx context.Context, paymentHash string, lnClient lnclient.LNClient) error
	SetTransactionMetadata(ctx context.Context, id uint, metadata map[string]interface{}) error
}

const (
	BoostagramTlvType = 7629169
	WhatsatTlvType    = 34349334
	CustomKeyTlvType  = 696969
)

// Prevent races when checking the current balance and creating payment
// transactions from concurrent goroutines.
var balanceValidationLock = &sync.Mutex{}

type Transaction = db.Transaction

type Boostagram struct {
	AppName        string         `json:"app_name"`
	Name           string         `json:"name"`
	Podcast        string         `json:"podcast"`
	URL            string         `json:"url"`
	Episode        StringOrNumber `json:"episode,omitempty"`
	FeedId         StringOrNumber `json:"feedID,omitempty"`
	ItemId         StringOrNumber `json:"itemID,omitempty"`
	Timestamp      int64          `json:"ts,omitempty"`
	Message        string         `json:"message,omitempty"`
	SenderId       StringOrNumber `json:"sender_id"`
	SenderName     string         `json:"sender_name"`
	Time           string         `json:"time"`
	Action         string         `json:"action"`
	ValueMsatTotal int64          `json:"value_msat_total"`
}

type StringOrNumber struct {
	StringData string
	NumberData int64
}

func (sn *StringOrNumber) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &sn.StringData); err == nil {
		return nil
	}

	if err := json.Unmarshal(data, &sn.NumberData); err == nil {
		return nil
	}

	return fmt.Errorf("cannot unmarshal %s into StringOrNumber type", data)
}

func (sn StringOrNumber) String() string {
	if sn.StringData != "" {
		return sn.StringData
	}
	return fmt.Sprintf("%d", sn.NumberData)
}

type notFoundError struct {
}

func NewNotFoundError() error {
	return &notFoundError{}
}

func (err *notFoundError) Error() string {
	return "The transaction requested was not found"
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
	return "Your app does not have enough budget remaining to make this payment. Please review this app in the connections page of your Alby Hub."
}

func NewTransactionsService(db *gorm.DB, eventPublisher events.EventPublisher) *transactionsService {
	return &transactionsService{
		db:             db,
		eventPublisher: eventPublisher,
	}
}

func (svc *transactionsService) MakeInvoice(ctx context.Context, amount uint64, description string, descriptionHash string, expiry uint64, metadata map[string]interface{}, lnClient lnclient.LNClient, appId *uint, requestEventId *uint, throughNodePubkey *string) (*Transaction, error) {
	logger.Logger.WithFields(logrus.Fields{
		"app_id":           appId,
		"request_event_id": requestEventId,
		"amount":           amount,
		"description":      description,
		"description_hash": descriptionHash,
		"expiry":           expiry,
		"metadata":         metadata,
	}).Debug("Making invoice")

	var metadataBytes []byte
	if metadata != nil {
		var err error
		metadataBytes, err = json.Marshal(metadata)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to serialize metadata")
			return nil, err
		}
		if len(metadataBytes) > constants.INVOICE_METADATA_MAX_LENGTH {
			return nil, fmt.Errorf("encoded invoice metadata provided is too large. Limit: %d Received: %d", constants.INVOICE_METADATA_MAX_LENGTH, len(metadataBytes))
		}
	}

	if metadata["app_id"] != nil {
		overwriteAppIdType, ok := metadata["app_id"].(float64)
		if !ok {
			return nil, errors.New("failed to overwrite app ID")
		}
		overwriteAppId := uint(overwriteAppIdType)
		logger.Logger.WithField("app_id", overwriteAppId).Info("Making invoice with overwritten app ID")
		appId = &overwriteAppId
	}

	lnClientTransaction, err := lnClient.MakeInvoice(ctx, int64(amount), description, descriptionHash, int64(expiry), throughNodePubkey)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create transaction")
		return nil, err
	}

	var preimage *string
	if lnClientTransaction.Preimage != "" {
		preimage = &lnClientTransaction.Preimage
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
		Metadata:        datatypes.JSON(metadataBytes),
	}
	err = svc.db.Create(&dbTransaction).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create DB transaction")
		return nil, err
	}
	return &dbTransaction, nil
}

func (svc *transactionsService) MakeHoldInvoice(ctx context.Context, amount uint64, description string, descriptionHash string, expiry uint64, paymentHash string, metadata map[string]interface{}, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error) {
	var err error
	var metadataBytes []byte
	if metadata != nil {
		metadataBytes, err = json.Marshal(metadata)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to serialize metadata")
			return nil, err
		}
		if len(metadataBytes) > constants.INVOICE_METADATA_MAX_LENGTH {
			return nil, fmt.Errorf("encoded invoice metadata provided is too large. Limit: %d Received: %d", constants.INVOICE_METADATA_MAX_LENGTH, len(metadataBytes))
		}
	}

	lnClientTransaction, err := lnClient.MakeHoldInvoice(ctx, int64(amount), description, descriptionHash, int64(expiry), paymentHash)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create hold invoice via LN client")
		return nil, err
	}

	var preimage *string
	if lnClientTransaction.Preimage != "" {
		preimage = &lnClientTransaction.Preimage
	}

	var expiresAt *time.Time
	if lnClientTransaction.ExpiresAt != nil {
		expiresAtValue := time.Unix(*lnClientTransaction.ExpiresAt, 0)
		expiresAt = &expiresAtValue
	}

	dbTransaction := db.Transaction{
		AppId:           appId,
		RequestEventId:  requestEventId,
		Type:            constants.TRANSACTION_TYPE_INCOMING,
		State:           constants.TRANSACTION_STATE_PENDING,
		AmountMsat:      uint64(lnClientTransaction.Amount),
		Description:     description,
		DescriptionHash: descriptionHash,
		PaymentRequest:  lnClientTransaction.Invoice,
		PaymentHash:     lnClientTransaction.PaymentHash,
		ExpiresAt:       expiresAt,
		Preimage:        preimage,
		Metadata:        datatypes.JSON(metadataBytes),
		Hold:            true,
	}
	err = svc.db.Create(&dbTransaction).Error
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create hold invoice DB transaction")
		return nil, err
	}
	return &dbTransaction, nil
}

func (svc *transactionsService) SendPaymentSync(payReq string, amountMsat *uint64, metadata map[string]interface{}, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error) {
	var metadataBytes []byte
	if metadata != nil {
		var err error
		metadataBytes, err = json.Marshal(metadata)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to serialize metadata")
			return nil, err
		}
		if len(metadataBytes) > constants.INVOICE_METADATA_MAX_LENGTH {
			return nil, fmt.Errorf("encoded payment metadata provided is too large. Limit: %d Received: %d", constants.INVOICE_METADATA_MAX_LENGTH, len(metadataBytes))
		}
	}

	payReq = strings.ToLower(payReq)
	paymentRequest, err := decodepay.Decodepay(payReq)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
		}).Errorf("Failed to decode bolt11 invoice: %v", err)

		return nil, err
	}

	if time.Now().After(time.Unix(int64(paymentRequest.CreatedAt+paymentRequest.Expiry), 0)) {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
			"expiry": time.Unix(int64(paymentRequest.CreatedAt+paymentRequest.Expiry), 0),
		}).Errorf("this invoice has expired")

		return nil, errors.New("this invoice has expired")
	}

	selfPayment := false
	if paymentRequest.Payee != "" && paymentRequest.Payee == lnClient.GetPubkey() {
		var incomingTransaction db.Transaction
		result := svc.db.Limit(1).Find(&incomingTransaction, &db.Transaction{
			Type:        constants.TRANSACTION_TYPE_INCOMING,
			PaymentHash: paymentRequest.PaymentHash,
		})
		if result.Error == nil && result.RowsAffected > 0 {
			selfPayment = true
		}
	}

	var dbTransaction db.Transaction

	paymentAmount := uint64(paymentRequest.MSatoshi)
	if amountMsat != nil && paymentRequest.MSatoshi == 0 {
		paymentAmount = *amountMsat
	}

	err = func() error {
		balanceValidationLock.Lock()
		defer balanceValidationLock.Unlock()
		return svc.db.Transaction(func(tx *gorm.DB) error {
			var existingSettledTransaction db.Transaction
			if tx.Limit(1).Find(&existingSettledTransaction, &db.Transaction{
				Type:        constants.TRANSACTION_TYPE_OUTGOING,
				PaymentHash: paymentRequest.PaymentHash,
				State:       constants.TRANSACTION_STATE_SETTLED,
			}).RowsAffected > 0 {
				logger.Logger.WithField("payment_hash", dbTransaction.PaymentHash).Debug("this invoice has already been paid")
				return errors.New("this invoice has already been paid")
			}
			if tx.Limit(1).Find(&existingSettledTransaction, &db.Transaction{
				Type:        constants.TRANSACTION_TYPE_OUTGOING,
				PaymentHash: paymentRequest.PaymentHash,
				State:       constants.TRANSACTION_STATE_PENDING,
			}).RowsAffected > 0 {
				logger.Logger.WithField("payment_hash", dbTransaction.PaymentHash).Debug("this invoice is already being paid")
				return errors.New("there is already a payment pending for this invoice")
			}

			err := svc.validateCanPay(tx, appId, paymentAmount, paymentRequest.Description, selfPayment)
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
				FeeReserveMsat:  CalculateFeeReserveMsat(paymentAmount),
				AmountMsat:      paymentAmount,
				PaymentRequest:  payReq,
				PaymentHash:     paymentRequest.PaymentHash,
				Description:     paymentRequest.Description,
				DescriptionHash: paymentRequest.DescriptionHash,
				ExpiresAt:       expiresAt,
				SelfPayment:     selfPayment,
				Metadata:        datatypes.JSON(metadataBytes),
			}
			err = tx.Create(&dbTransaction).Error
			return err
		})
	}()

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
		}).WithError(err).Error("Failed to create DB transaction")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{
		"app_id":           appId,
		"request_event_id": requestEventId,
		"amount":           paymentAmount,
		"description":      paymentRequest.Description,
		"description_hash": paymentRequest.DescriptionHash,
		"expiry":           paymentRequest.Expiry,
		"self_payment":     selfPayment,
		"metadata":         metadata,
	}).Debug("Initiating payment")

	var response *lnclient.PayInvoiceResponse
	if selfPayment {
		response, err = svc.interceptSelfPayment(paymentRequest.PaymentHash, lnClient)
	} else {
		response, err = lnClient.SendPaymentSync(payReq, amountMsat)
	}

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": payReq,
		}).WithError(err).Error("Failed to send payment")

		svc.db.Transaction(func(tx *gorm.DB) error {
			return svc.markPaymentFailed(tx, &dbTransaction, err.Error())
		})

		return nil, err
	}

	// the payment definitely succeeded
	var settledTransaction *db.Transaction
	err = svc.db.Transaction(func(tx *gorm.DB) error {
		settledTransaction, err = svc.markTransactionSettled(tx, &dbTransaction, response.Preimage, response.Fee, selfPayment)
		return err
	})
	if err != nil {
		return nil, err
	}

	return settledTransaction, nil
}

func (svc *transactionsService) SendKeysend(amount uint64, destination string, customRecords []lnclient.TLVRecord, preimage string, lnClient lnclient.LNClient, appId *uint, requestEventId *uint) (*Transaction, error) {
	if preimage == "" {
		preImageBytes, err := makePreimageHex()
		if err != nil {
			return nil, err
		}
		preimage = hex.EncodeToString(preImageBytes)
	}

	preImageBytes, err := hex.DecodeString(preimage)
	if err != nil || len(preImageBytes) != 32 {
		logger.Logger.WithFields(logrus.Fields{
			"preimage": preimage,
		}).WithError(err).Error("Invalid preimage")
		return nil, err
	}

	paymentHash256 := sha256.New()
	paymentHash256.Write(preImageBytes)
	paymentHashBytes := paymentHash256.Sum(nil)
	paymentHash := hex.EncodeToString(paymentHashBytes)

	metadata := map[string]interface{}{}

	metadata["destination"] = destination

	metadata["tlv_records"] = customRecords
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to serialize transaction metadata")
		return nil, err
	}
	boostagramBytes := svc.getBoostagramBytesFromCustomRecords(customRecords)

	var dbTransaction db.Transaction

	selfPayment := destination == lnClient.GetPubkey()

	err = func() error {
		balanceValidationLock.Lock()
		defer balanceValidationLock.Unlock()
		return svc.db.Transaction(func(tx *gorm.DB) error {
			err := svc.validateCanPay(tx, appId, amount, "", selfPayment)
			if err != nil {
				return err
			}

			dbTransaction = db.Transaction{
				AppId:          appId,
				Description:    svc.getDescriptionFromCustomRecords(customRecords),
				RequestEventId: requestEventId,
				Type:           constants.TRANSACTION_TYPE_OUTGOING,
				State:          constants.TRANSACTION_STATE_PENDING,
				FeeReserveMsat: CalculateFeeReserveMsat(uint64(amount)),
				AmountMsat:     amount,
				Metadata:       datatypes.JSON(metadataBytes),
				Boostagram:     datatypes.JSON(boostagramBytes),
				PaymentHash:    paymentHash,
				Preimage:       &preimage,
				SelfPayment:    selfPayment,
			}
			err = tx.Create(&dbTransaction).Error

			return err
		})
	}()

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"destination": destination,
			"amount":      amount,
		}).WithError(err).Error("Failed to create DB transaction")
		return nil, err
	}

	var payKeysendResponse *lnclient.PayKeysendResponse

	if selfPayment {
		// for keysend self-payments we need to create an incoming payment at the time of the payment
		recipientAppId := svc.getAppIdFromCustomRecords(customRecords, svc.db)
		dbTransaction := db.Transaction{
			AppId:          recipientAppId,
			RequestEventId: nil, // it is related to this request but for a different app
			Type:           constants.TRANSACTION_TYPE_INCOMING,
			State:          constants.TRANSACTION_STATE_PENDING,
			AmountMsat:     amount,
			PaymentHash:    paymentHash,
			Preimage:       &preimage,
			Description:    svc.getDescriptionFromCustomRecords(customRecords),
			Metadata:       datatypes.JSON(metadataBytes),
			Boostagram:     datatypes.JSON(boostagramBytes),
			SelfPayment:    true,
		}
		err = svc.db.Create(&dbTransaction).Error
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to create DB transaction")
			return nil, err
		}

		_, err = svc.interceptSelfPayment(paymentHash, lnClient)
		if err == nil {
			payKeysendResponse = &lnclient.PayKeysendResponse{
				Fee: 0,
			}
		}
	} else {
		payKeysendResponse, err = lnClient.SendKeysend(amount, destination, customRecords, preimage)
	}

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"destination": destination,
			"amount":      amount,
		}).WithError(err).Error("Failed to send payment")

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
	var settledTransaction *db.Transaction
	err = svc.db.Transaction(func(tx *gorm.DB) error {
		settledTransaction, err = svc.markTransactionSettled(tx, &dbTransaction, preimage, payKeysendResponse.Fee, selfPayment)
		return err
	})

	if err != nil {
		return nil, err
	}

	return settledTransaction, nil
}

func (svc *transactionsService) LookupTransaction(ctx context.Context, paymentHash string, transactionType *string, lnClient lnclient.LNClient, appId *uint) (*Transaction, error) {
	transaction := db.Transaction{}

	tx := svc.db

	var isIsolatedApp bool
	if appId != nil {
		err := svc.db.
			Model(&db.App{}).
			Where("id", *appId).
			Pluck("isolated", &isIsolatedApp).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, NewNotFoundError()
			}
			return nil, err
		}
	}

	if isIsolatedApp {
		tx = tx.Where("app_id = ?", *appId)
	}

	if transactionType != nil {
		tx = tx.Where("type = ?", *transactionType)
	}

	// order settled first, otherwise by created date, as there can be multiple outgoing payments
	// for the same payment hash (if you tried to pay an invoice multiple times - e.g. the first time failed)
	result := tx.Order("settled_at desc, created_at desc").Limit(1).Find(&transaction, &db.Transaction{
		// Type:        transactionType,
		PaymentHash: paymentHash,
	})

	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to lookup transaction")
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		logger.Logger.WithFields(logrus.Fields{
			"payment_hash": paymentHash,
			"app_id":       appId,
		}).WithError(result.Error).Error("transaction not found")
		return nil, NewNotFoundError()
	}

	if transaction.State == constants.TRANSACTION_STATE_PENDING {
		svc.checkUnsettledTransaction(ctx, &transaction, lnClient)
	}

	return &transaction, nil
}

func (svc *transactionsService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaidOutgoing bool, unpaidIncoming bool, transactionType *string, lnClient lnclient.LNClient, appId *uint, forceFilterByAppId bool) (transactions []Transaction, totalCount uint64, err error) {
	svc.checkUnsettledTransactions(ctx, lnClient)

	var isIsolatedApp bool
	if appId != nil {
		err := svc.db.
			Model(&db.App{}).
			Where("id", *appId).
			Pluck("isolated", &isIsolatedApp).
			Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, 0, NewNotFoundError()
			}
			return nil, 0, err
		}
	}

	tx := svc.db

	if isIsolatedApp || forceFilterByAppId {
		tx = tx.Where("app_id = ?", *appId)
	}

	if !unpaidOutgoing && !unpaidIncoming {
		tx = tx.Where("state = ?", constants.TRANSACTION_STATE_SETTLED)
	} else if unpaidOutgoing && !unpaidIncoming {
		tx = tx.Where("state = ? OR type = ?", constants.TRANSACTION_STATE_SETTLED, constants.TRANSACTION_TYPE_OUTGOING)
	} else if unpaidIncoming && !unpaidOutgoing {
		tx = tx.Where("state = ? OR type = ?", constants.TRANSACTION_STATE_SETTLED, constants.TRANSACTION_TYPE_INCOMING)
	}

	if transactionType != nil {
		tx = tx.Where("type = ?", *transactionType)
	}

	if from > 0 {
		tx = tx.Where("updated_at >= ?", time.Unix(int64(from), 0))
	}
	if until > 0 {
		tx = tx.Where("updated_at <= ?", time.Unix(int64(until), 0))
	}

	var totalCount64 int64
	result := tx.Model(&db.Transaction{}).Count(&totalCount64)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to count DB transactions")
		return nil, 0, result.Error
	}
	totalCount = uint64(totalCount64)

	tx = tx.Order("updated_at desc")

	if limit > 0 {
		tx = tx.Limit(int(limit))
	}
	if offset > 0 {
		tx = tx.Offset(int(offset))
	}

	result = tx.Find(&transactions)
	if result.Error != nil {
		logger.Logger.WithError(result.Error).Error("Failed to list DB transactions")
		return nil, 0, result.Error
	}

	return transactions, totalCount, nil
}

func (svc *transactionsService) checkUnsettledTransactions(ctx context.Context, lnClient lnclient.LNClient) {
	// Only check unsettled transactions for clients that don't support async events
	// checkUnsettledTransactions does not work for keysend payments!
	if slices.Contains(lnClient.GetSupportedNIP47NotificationTypes(), "payment_received") {
		return
	}

	// check pending payments less than a day old
	transactions := []Transaction{}
	result := svc.db.Where("state = ? AND created_at > ?", constants.TRANSACTION_STATE_PENDING, time.Now().Add(-24*time.Hour)).Find(&transactions)
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
		err = svc.db.Transaction(func(tx *gorm.DB) error {
			_, err = svc.markTransactionSettled(tx, transaction, lnClientTransaction.Preimage, uint64(lnClientTransaction.FeesPaid), false)
			return err
		})

		if err != nil {
			logger.Logger.WithError(err).Error("Failed to mark payment sent when checking unsettled transaction")
		}
	}
}

func (svc *transactionsService) ConsumeEvent(ctx context.Context, event *events.Event, globalProperties map[string]interface{}) {
	switch event.Event {
	case "nwc_lnclient_payment_received":
		lnClientTransaction, ok := event.Properties.(*lnclient.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return
		}

		var dbTransaction db.Transaction
		err := svc.db.Transaction(func(tx *gorm.DB) error {

			result := tx.Limit(1).Find(&dbTransaction, &db.Transaction{
				Type:        constants.TRANSACTION_TYPE_INCOMING,
				PaymentHash: lnClientTransaction.PaymentHash,
			})

			if result.RowsAffected == 0 {
				var appId *uint
				description := lnClientTransaction.Description
				var metadataBytes []byte
				var boostagramBytes []byte
				if lnClientTransaction.Metadata != nil {
					var err error
					metadataBytes, err = json.Marshal(lnClientTransaction.Metadata)
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to serialize transaction metadata")
						return err
					}

					var customRecords []lnclient.TLVRecord
					customRecords, _ = lnClientTransaction.Metadata["tlv_records"].([]lnclient.TLVRecord)
					boostagramBytes = svc.getBoostagramBytesFromCustomRecords(customRecords)
					extractedDescription := svc.getDescriptionFromCustomRecords(customRecords)
					if extractedDescription != "" {
						description = extractedDescription
					}
					// find app by custom key/value records
					appId = svc.getAppIdFromCustomRecords(customRecords, tx)
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
					Description:     description,
					DescriptionHash: lnClientTransaction.DescriptionHash,
					ExpiresAt:       expiresAt,
					Metadata:        datatypes.JSON(metadataBytes),
					Boostagram:      datatypes.JSON(boostagramBytes),
					AppId:           appId,
				}
				err := tx.Create(&dbTransaction).Error
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"payment_hash": lnClientTransaction.PaymentHash,
					}).WithError(err).Error("Failed to create transaction")
					return err
				}
			}

			_, err := svc.markTransactionSettled(tx, &dbTransaction, lnClientTransaction.Preimage, uint64(lnClientTransaction.FeesPaid), false)
			return err
		})

		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": lnClientTransaction.PaymentHash,
			}).WithError(err).Error("Failed to execute DB transaction")
			return
		}

	case "nwc_lnclient_hold_invoice_accepted":
		lnClientTransaction, ok := event.Properties.(*lnclient.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event properties for hold invoice accepted")
			return
		}
		if lnClientTransaction.SettleDeadline == nil {
			logger.Logger.WithField("event", event).Error("Transaction has no settle deadline")
			return
		}
		svc.markHoldInvoiceAccepted(lnClientTransaction.PaymentHash, *lnClientTransaction.SettleDeadline, false)

	case "nwc_lnclient_payment_sent":
		lnClientTransaction, ok := event.Properties.(*lnclient.Transaction)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return
		}

		var dbTransaction db.Transaction
		err := svc.db.Transaction(func(tx *gorm.DB) error {

			// first lookup by pending
			result := tx.Limit(1).Find(&dbTransaction, &db.Transaction{
				Type:        constants.TRANSACTION_TYPE_OUTGOING,
				State:       constants.TRANSACTION_STATE_PENDING,
				PaymentHash: lnClientTransaction.PaymentHash,
			})

			if result.Error != nil {
				return result.Error
			}

			if result.RowsAffected == 0 {
				// if no pending payment was found, lookup by failed, latest updated first
				result := tx.Limit(1).Order("updated_at DESC").Find(&dbTransaction, &db.Transaction{
					Type:        constants.TRANSACTION_TYPE_OUTGOING,
					State:       constants.TRANSACTION_STATE_FAILED,
					PaymentHash: lnClientTransaction.PaymentHash,
				})

				if result.Error != nil {
					return result.Error
				}

				if result.RowsAffected == 0 {
					// Note: payments made from outside cannot be associated with an app
					// for now this is disabled as it only applies to LND, and we do not import LND transactions either.
					logger.Logger.WithField("payment_hash", lnClientTransaction.PaymentHash).Error("failed to mark payment as sent: payment not found")
					return NewNotFoundError()
				}
			}

			_, err := svc.markTransactionSettled(tx, &dbTransaction, lnClientTransaction.Preimage, uint64(lnClientTransaction.FeesPaid), false)
			return err
		})

		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": lnClientTransaction.PaymentHash,
			}).WithError(err).Error("Failed to update transaction")
			return
		}
	case "nwc_lnclient_payment_failed":
		paymentFailedAsyncProperties, ok := event.Properties.(*lnclient.PaymentFailedEventProperties)
		if !ok {
			logger.Logger.WithField("event", event).Error("Failed to cast event")
			return
		}

		lnClientTransaction := paymentFailedAsyncProperties.Transaction

		var dbTransaction db.Transaction
		result := svc.db.Limit(1).Find(&dbTransaction, &db.Transaction{
			Type:        constants.TRANSACTION_TYPE_OUTGOING,
			State:       constants.TRANSACTION_STATE_PENDING,
			PaymentHash: lnClientTransaction.PaymentHash,
		})

		if result.RowsAffected == 0 {
			logger.Logger.WithField("event", event).Error("Failed to find pending outgoing transaction by payment hash")
			return
		}

		svc.db.Transaction(func(tx *gorm.DB) error {
			return svc.markPaymentFailed(tx, &dbTransaction, paymentFailedAsyncProperties.Reason)
		})
	}
}

func (svc *transactionsService) markHoldInvoiceAccepted(paymentHash string, settleDeadline uint32, selfPayment bool) {
	logger.Logger.WithFields(logrus.Fields{
		"paymentHash":  paymentHash,
		"self_payment": selfPayment,
	}).Info("Processing hold invoice accepted event")

	var dbTransaction db.Transaction
	err := svc.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Where("payment_hash = ? AND type = ? AND state = ?", paymentHash, constants.TRANSACTION_TYPE_INCOMING, constants.TRANSACTION_STATE_PENDING).First(&dbTransaction)
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				logger.Logger.WithFields(logrus.Fields{
					"paymentHash": paymentHash,
				}).Warn("No corresponding pending incoming transaction found in DB for accepted hold invoice")
			}
			logger.Logger.WithFields(logrus.Fields{
				"paymentHash": paymentHash,
			}).WithError(result.Error).Error("Failed to query DB for accepted hold invoice")
			return result.Error
		}

		err := tx.Model(&dbTransaction).UpdateColumns(map[string]interface{}{
			"state":           constants.TRANSACTION_STATE_ACCEPTED,
			"self_payment":    selfPayment,
			"settle_deadline": settleDeadline,
		}).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"paymentHash": paymentHash,
				"dbTxID":      dbTransaction.ID,
			}).WithError(err).Error("Failed to update hold invoice state to accepted in DB")
			return err
		}

		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
			"dbTxID":      dbTransaction.ID,
		}).Info("Updated hold invoice state to accepted in DB")

		return nil
	})
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentHash,
		}).WithError(err).Error("Failed DB transaction for hold invoice accepted event")
	} else {
		svc.eventPublisher.Publish(&events.Event{
			Event:      "nwc_hold_invoice_accepted",
			Properties: &dbTransaction,
		})
	}
}

func (svc *transactionsService) interceptSelfPayment(paymentHash string, lnClient lnclient.LNClient) (*lnclient.PayInvoiceResponse, error) {
	logger.Logger.WithField("payment_hash", paymentHash).Debug("Intercepting self payment")
	incomingTransaction := db.Transaction{}
	result := svc.db.Limit(1).Find(&incomingTransaction, &db.Transaction{
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

	if incomingTransaction.Hold {
		return svc.interceptSelfHoldPayment(paymentHash, lnClient)
	}

	if incomingTransaction.Preimage == nil {
		return nil, errors.New("preimage is not set on transaction. Self payments not supported")
	}

	err := svc.db.Transaction(func(tx *gorm.DB) error {
		_, err := svc.markTransactionSettled(tx, &incomingTransaction, *incomingTransaction.Preimage, uint64(0), true)
		return err
	})

	if err != nil {
		return nil, err
	}

	return &lnclient.PayInvoiceResponse{
		Preimage: *incomingTransaction.Preimage,
		Fee:      0,
	}, nil
}

func (svc *transactionsService) interceptSelfHoldPayment(paymentHash string, lnClient lnclient.LNClient) (*lnclient.PayInvoiceResponse, error) {
	settledChannel := make(chan *db.Transaction)
	canceledChannel := make(chan *db.Transaction)

	holdInvoiceUpdatedConsumer := newHoldInvoiceUpdatedConsumer(paymentHash, settledChannel, canceledChannel)

	svc.eventPublisher.RegisterSubscriber(holdInvoiceUpdatedConsumer)
	defer svc.eventPublisher.RemoveSubscriber(holdInvoiceUpdatedConsumer)

	clientInfo, err := lnClient.GetInfo(context.Background())
	if err != nil {
		return nil, errors.New("failed to get client info")
	}
	if clientInfo.BlockHeight == 0 {
		return nil, errors.New("invalid client block height")
	}

	fakeSettleDeadline := clientInfo.BlockHeight + 24

	svc.markHoldInvoiceAccepted(paymentHash, fakeSettleDeadline, true)

	select {
	case settledTransaction := <-settledChannel:
		logger.Logger.WithField("settled_transaction", settledTransaction).Info("self hold payment was settled")
		if settledTransaction.Preimage == nil {
			return nil, errors.New("preimage is not set on self hold payment")
		}

		return &lnclient.PayInvoiceResponse{
			Preimage: *settledTransaction.Preimage,
			Fee:      0,
		}, nil
	case canceledTransaction := <-canceledChannel:
		logger.Logger.WithField("canceled_transaction", canceledTransaction).Info("self hold payment was canceled")
		return nil, lnclient.NewHoldInvoiceCanceledError()
	}
}

func (svc *transactionsService) validateCanPay(tx *gorm.DB, appId *uint, amount uint64, description string, selfPayment bool) error {
	amountWithFeeReserve := amount
	if !selfPayment {
		amountWithFeeReserve += CalculateFeeReserveMsat(amount)
	}

	// ensure balance for isolated apps
	if appId != nil {
		var app db.App
		result := tx.Limit(1).Find(&app, &db.App{
			ID: *appId,
		})
		if result.RowsAffected == 0 {
			return NewNotFoundError()
		}

		var appPermission db.AppPermission
		result = tx.Limit(1).Find(&appPermission, &db.AppPermission{
			AppId: *appId,
			Scope: constants.PAY_INVOICE_SCOPE,
		})
		if result.RowsAffected == 0 {
			return errors.New("app does not have pay_invoice scope")
		}

		if app.Isolated {
			balance := queries.GetIsolatedBalance(tx, appPermission.AppId)

			if int64(amountWithFeeReserve) > balance {
				logger.Logger.WithFields(logrus.Fields{
					"balance":                 balance,
					"self_payment":            selfPayment,
					"amount":                  amount,
					"amount_with_fee_reserve": amountWithFeeReserve,
				}).Debug("Insufficient budget to make payment from isolated app")
				message := NewInsufficientBalanceError().Error()
				if description != "" {
					message += " " + description
				}

				svc.eventPublisher.Publish(&events.Event{
					Event: "nwc_permission_denied",
					Properties: map[string]interface{}{
						"app_name": app.Name,
						"code":     constants.ERROR_INSUFFICIENT_BALANCE,
						"message":  message,
					},
				})
				return NewInsufficientBalanceError()
			}
		}

		if appPermission.MaxAmountSat > 0 {
			budgetUsageSat := queries.GetBudgetUsageSat(tx, &appPermission)
			if int(amountWithFeeReserve/1000) > appPermission.MaxAmountSat-int(budgetUsageSat) {
				message := NewQuotaExceededError().Error()
				if description != "" {
					message += " " + description
				}
				svc.eventPublisher.Publish(&events.Event{
					Event: "nwc_permission_denied",
					Properties: map[string]interface{}{
						"app_name": app.Name,
						"code":     constants.ERROR_QUOTA_EXCEEDED,
						"message":  message,
					},
				})
				return NewQuotaExceededError()
			}
		}
	}

	return nil
}

// max of 1% or 10000 millisats (10 sats)
func CalculateFeeReserveMsat(amountMsat uint64) uint64 {
	return uint64(math.Max(math.Ceil(float64(amountMsat)*0.01), 10000))
}

func makePreimageHex() ([]byte, error) {
	bytes := make([]byte, 32) // 32 bytes * 8 bits/byte = 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func (svc *transactionsService) getBoostagramBytesFromCustomRecords(customRecords []lnclient.TLVRecord) []byte {
	for _, record := range customRecords {
		if record.Type == BoostagramTlvType {
			bytes, err := hex.DecodeString(record.Value)
			if err != nil {
				logger.Logger.WithField("value", record.Value).WithError(err).Error("failed to decode boostagram tlv hex value")
				return nil
			}

			// ensure the boostagram is valid json
			var boostagram Boostagram
			if err := json.Unmarshal(bytes, &boostagram); err != nil {
				logger.Logger.WithField("value", string(bytes)).WithError(err).Error("failed to unmarshal boostagram to json")
				return nil
			}

			return bytes
		}
	}

	return nil
}

func (svc *transactionsService) getDescriptionFromCustomRecords(customRecords []lnclient.TLVRecord) string {
	var description string

	for _, record := range customRecords {
		switch record.Type {
		case BoostagramTlvType:
			bytes, err := hex.DecodeString(record.Value)
			if err != nil {
				continue
			}
			var boostagram Boostagram
			if err := json.Unmarshal(bytes, &boostagram); err != nil {
				continue
			}
			return boostagram.Message

		// TODO: consider adding support for this in LDK
		case WhatsatTlvType:
			bytes, err := hex.DecodeString(record.Value)
			if err == nil {
				description = string(bytes)
			}
		}
	}

	return description
}

func (svc *transactionsService) getAppIdFromCustomRecords(customRecords []lnclient.TLVRecord, tx *gorm.DB) *uint {
	app := db.App{}
	for _, record := range customRecords {
		if record.Type == CustomKeyTlvType {
			decodedString, err := hex.DecodeString(record.Value)
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to parse custom key TLV record as hex")
				continue
			}
			customValue, err := strconv.ParseUint(string(decodedString), 10, 64)
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to parse custom key TLV record as number")
				continue
			}
			err = tx.Take(&app, &db.App{
				ID: uint(customValue),
			}).Error
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to find app by id from custom key TLV record")
				continue
			}
			return &app.ID
		}
	}
	return nil
}

func (svc *transactionsService) SettleHoldInvoice(ctx context.Context, preimage string, lnClient lnclient.LNClient) (*Transaction, error) {
	if len(preimage) != 64 {
		return nil, errors.New("invalid preimage format")
	}
	preimageBytes, err := hex.DecodeString(preimage)
	if err != nil {
		return nil, fmt.Errorf("invalid preimage hex: %w", err)
	}

	paymentHashBytes := sha256.Sum256(preimageBytes)
	paymentHash := hex.EncodeToString(paymentHashBytes[:])

	var dbTransaction db.Transaction
	result := svc.db.Limit(1).Find(&dbTransaction, &db.Transaction{
		Type:        constants.TRANSACTION_TYPE_INCOMING,
		State:       constants.TRANSACTION_STATE_ACCEPTED,
		PaymentHash: paymentHash,
	})

	if result.RowsAffected == 0 {
		logger.Logger.WithField("payment_hash", paymentHash).Error("Failed to find accepted hold invoice")
		return nil, errors.New("failed to find accepted hold invoice")
	}

	if !dbTransaction.SelfPayment {
		err = lnClient.SettleHoldInvoice(ctx, preimage)
	}

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"preimage": preimage,
		}).WithError(err).Error("Failed to settle hold invoice via LN client")
		// Don't mark DB as failed here, as the settle might succeed later or might have already succeeded.
		return nil, err
	}

	var settledTransaction *db.Transaction
	err = svc.db.Transaction(func(tx *gorm.DB) error {
		var err error
		settledTransaction, err = svc.markTransactionSettled(tx, &dbTransaction, preimage, 0, dbTransaction.SelfPayment)
		return err
	})

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"payment_hash": paymentHash,
			"preimage":     preimage,
		}).WithError(err).Error("Failed DB transaction while settling hold invoice")
		return nil, err
	}

	return settledTransaction, nil
}

func (svc *transactionsService) CancelHoldInvoice(ctx context.Context, paymentHash string, lnClient lnclient.LNClient) error {

	var dbTransaction db.Transaction
	result := svc.db.Limit(1).Find(&dbTransaction, &db.Transaction{
		Type:        constants.TRANSACTION_TYPE_INCOMING,
		State:       constants.TRANSACTION_STATE_ACCEPTED,
		PaymentHash: paymentHash,
	})

	if result.RowsAffected == 0 {
		logger.Logger.WithField("payment_hash", paymentHash).Error("Failed to find accepted hold invoice")
		return NewNotFoundError()
	}

	if !dbTransaction.SelfPayment {
		err := lnClient.CancelHoldInvoice(ctx, paymentHash)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": paymentHash,
			}).WithError(err).Error("Failed to cancel hold invoice via LN client")
			// Don't mark DB as failed here, cancellation might have already happened or might succeed later.
			return err
		}
	}

	err := svc.db.Transaction(func(tx *gorm.DB) error {
		var dbTransaction db.Transaction
		result := tx.Limit(1).Find(&dbTransaction, &db.Transaction{
			Type:        constants.TRANSACTION_TYPE_INCOMING,
			State:       constants.TRANSACTION_STATE_ACCEPTED,
			PaymentHash: paymentHash,
		})

		if result.Error != nil {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": paymentHash,
			}).WithError(result.Error).Error("Failed to find accepted hold invoice in DB for cancellation")
			return result.Error
		}
		if result.RowsAffected == 0 {
			logger.Logger.WithFields(logrus.Fields{
				"payment_hash": paymentHash,
			}).Warn("No accepted hold invoice found in DB to mark as failed due to cancellation")
			return NewNotFoundError()
		}

		return svc.markPaymentFailed(tx, &dbTransaction, "Hold invoice was cancelled")
	})

	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"payment_hash": paymentHash,
		}).WithError(err).Error("Failed DB transaction while canceling hold invoice")
		return err
	}

	logger.Logger.WithFields(logrus.Fields{
		"payment_hash": paymentHash,
	}).Info("Marked hold invoice as failed in DB due to cancellation")

	svc.eventPublisher.Publish(&events.Event{
		Event:      "nwc_hold_invoice_canceled",
		Properties: &dbTransaction,
	})

	return nil
}

func (svc *transactionsService) SetTransactionMetadata(ctx context.Context, id uint, metadata map[string]interface{}) error {
	var metadataBytes []byte
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to serialize metadata")
		return err
	}
	if len(metadataBytes) > constants.INVOICE_METADATA_MAX_LENGTH {
		return fmt.Errorf("encoded invoice metadata provided is too large. Limit: %d Received: %d", constants.INVOICE_METADATA_MAX_LENGTH, len(metadataBytes))
	}

	err = svc.db.Model(&db.Transaction{}).Where("id", id).Update("metadata", datatypes.JSON(metadataBytes)).Error
	if err != nil {
		logger.Logger.WithError(err).WithField("metadata", metadata).Error("Failed to update transaction metadata")
		return err
	}

	return nil
}

func (svc *transactionsService) markTransactionSettled(tx *gorm.DB, dbTransaction *db.Transaction, preimage string, fee uint64, selfPayment bool) (*db.Transaction, error) {
	if preimage == "" {
		return nil, errors.New("no preimage in payment")
	}

	if tx.Dialector.Name() == "postgres" {
		// lock based on payment hash to ensure we only mark one transaction as settled
		// (in sqlite transactions are serializable by default)
		transactionsWithPaymentHash := []db.Transaction{}
		tx.Where(&db.Transaction{
			PaymentHash: dbTransaction.PaymentHash,
		}).Clauses(clause.Locking{Strength: "UPDATE"}).Find(&transactionsWithPaymentHash)
	}

	var existingSettledTransaction db.Transaction
	if tx.Limit(1).Find(&existingSettledTransaction, &db.Transaction{
		Type:        dbTransaction.Type,
		PaymentHash: dbTransaction.PaymentHash,
		State:       constants.TRANSACTION_STATE_SETTLED,
	}).RowsAffected > 0 {
		logger.Logger.WithField("payment_hash", dbTransaction.PaymentHash).Debug("payment already marked as sent")
		return &existingSettledTransaction, nil
	}

	now := time.Now()
	err := tx.Model(dbTransaction).Updates(map[string]interface{}{
		"State":          constants.TRANSACTION_STATE_SETTLED,
		"Preimage":       &preimage,
		"FeeMsat":        fee,
		"FeeReserveMsat": 0,
		"SettledAt":      &now,
		"SelfPayment":    selfPayment,
	}).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"payment_hash": dbTransaction.PaymentHash,
		}).WithError(err).Error("Failed to update DB transaction")
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{
		"payment_hash": dbTransaction.PaymentHash,
		"type":         dbTransaction.Type,
	}).Info("Marked transaction as settled")

	event := "nwc_payment_sent"
	if dbTransaction.Type == constants.TRANSACTION_TYPE_INCOMING {
		event = "nwc_payment_received"
	}

	svc.eventPublisher.Publish(&events.Event{
		Event:      event,
		Properties: dbTransaction,
	})

	if dbTransaction.Type == constants.TRANSACTION_TYPE_OUTGOING && dbTransaction.AppId != nil {
		svc.checkBudgetUsage(dbTransaction, tx)
	}

	return dbTransaction, nil
}

func (svc *transactionsService) checkBudgetUsage(dbTransaction *db.Transaction, gormTransaction *gorm.DB) {
	var app db.App
	result := gormTransaction.Limit(1).Find(&app, &db.App{
		ID: *dbTransaction.AppId,
	})
	if result.RowsAffected == 0 {
		logger.Logger.WithField("app_id", dbTransaction.AppId).Error("failed to find app by id")
		return
	}
	if app.Isolated {
		return
	}

	var appPermission db.AppPermission
	result = gormTransaction.Limit(1).Find(&appPermission, &db.AppPermission{
		AppId: app.ID,
		Scope: constants.PAY_INVOICE_SCOPE,
	})
	if result.RowsAffected == 0 {
		logger.Logger.WithField("app_id", dbTransaction.AppId).Error("failed to find pay_invoice scope")
		return
	}

	budgetUsage := queries.GetBudgetUsageSat(gormTransaction, &appPermission)
	warningUsage := uint64(math.Floor(float64(appPermission.MaxAmountSat) * 0.8))
	if budgetUsage >= warningUsage && budgetUsage-dbTransaction.AmountMsat/1000 < warningUsage {
		svc.eventPublisher.Publish(&events.Event{
			Event: "nwc_budget_warning",
			Properties: map[string]interface{}{
				"name": app.Name,
				"id":   app.ID,
			},
		})
	}
}

func (svc *transactionsService) markPaymentFailed(tx *gorm.DB, dbTransaction *db.Transaction, reason string) error {
	var existingTransaction db.Transaction
	result := tx.Limit(1).Find(&existingTransaction, &db.Transaction{
		ID: dbTransaction.ID,
	})

	if result.Error != nil {
		logger.Logger.WithField("payment_hash", dbTransaction.PaymentHash).WithError(result.Error).Error("could not find transaction to mark as failed")
		return result.Error
	}

	if existingTransaction.State == constants.TRANSACTION_STATE_FAILED {
		logger.Logger.WithField("payment_hash", dbTransaction.PaymentHash).Info("payment already marked as failed")
		return nil
	}

	err := tx.Model(dbTransaction).Updates(map[string]interface{}{
		"State":          constants.TRANSACTION_STATE_FAILED,
		"FeeReserveMsat": 0,
		"FailureReason":  reason,
	}).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"payment_hash": dbTransaction.PaymentHash,
		}).WithError(err).Error("Failed to mark transaction as failed")
		return err
	}
	logger.Logger.WithField("payment_hash", dbTransaction.PaymentHash).Info("Marked transaction as failed")

	svc.eventPublisher.Publish(&events.Event{
		Event:      "nwc_payment_failed",
		Properties: dbTransaction,
	})
	return nil
}
