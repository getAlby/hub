package controllers

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/getAlby/go-nostr"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/models"
	"github.com/getAlby/hub/transactions"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

const instructionTypeBolt11 = "bolt11"
const instructionTypeBolt12 = "bolt12"

type payParams struct {
	Payment   string                 `json:"payment"`
	Amount    *uint64                `json:"amount"`
	PayerNote string                 `json:"payer_note"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

type payResult struct {
	TransactionId   string `json:"transaction_id"`
	State           string `json:"state"`
	InstructionType string `json:"instruction_type"`
	Amount          uint64 `json:"amount"`
	FeesPaid        uint64 `json:"fees_paid"`
	PaymentHash     string `json:"payment_hash,omitempty"`
	Preimage        string `json:"preimage,omitempty"`
	CreatedAt       int64  `json:"created_at"`
	SettledAt       *int64 `json:"settled_at,omitempty"`
}

// parsed payment instructions from a BIP-321 URI
type bip321Payment struct {
	bolt11 string
	bolt12 string
	// from the BIP-321 "amount" parameter (BTC), converted to msat
	amountMsat *uint64
}

// HandlePayEvent handles the NWC-321 pay method. It pays one instruction
// from a BIP-321 URI: a BOLT-12 offer (the "lno" URI parameter) when the LN
// backend supports BOLT-12, otherwise the BOLT-11 invoice (the "lightning"
// URI parameter).
func (controller *nip47Controller) HandlePayEvent(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc) {
	payParams := &payParams{}
	resp := decodeRequest(nip47Request, payParams)
	if resp != nil {
		publishResponse(resp, nostr.Tags{})
		return
	}

	publishError := func(nip47Error *models.Error) {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           app.ID,
			"payment":          payParams.Payment,
			"code":             nip47Error.Code,
		}).Error(nip47Error.Message)

		publishResponse(&models.Response{
			ResultType: nip47Request.Method,
			Error:      nip47Error,
		}, nostr.Tags{})
	}

	bip321, nip47Error := parseBip321Uri(payParams.Payment)
	if nip47Error != nil {
		publishError(nip47Error)
		return
	}

	nodeInfo, err := controller.lnClient.GetInfo(ctx)
	if err != nil {
		publishError(&models.Error{
			Code:    constants.ERROR_INTERNAL,
			Message: fmt.Sprintf("Failed to get node info: %s", err.Error()),
		})
		return
	}

	// prefer the BOLT-12 offer if the backend can pay one; per NWC-321 only
	// one instruction may be selected
	if bip321.bolt12 != "" && nodeInfo.SupportsBolt12 {
		controller.payBolt12(ctx, nip47Request, requestEventId, app, publishResponse, publishError, payParams, bip321, nodeInfo.Network)
		return
	}

	if bip321.bolt12 != "" && bip321.bolt11 == "" {
		publishError(&models.Error{
			Code:    constants.ERROR_UNSUPPORTED_PAYMENT_INSTRUCTION,
			Message: "no supported payment instruction found: BOLT-12 payments are not supported by this LN backend",
		})
		return
	}

	controller.payBolt11(ctx, nip47Request, requestEventId, app, publishResponse, publishError, payParams, bip321, nodeInfo.Network)
}

// payBolt11 pays the BOLT-11 instruction of a BIP-321 URI.
func (controller *nip47Controller) payBolt11(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc, publishError func(*models.Error), payParams *payParams, bip321 *bip321Payment, nodeNetwork string) {
	// BOLT-11 does not support payer-provided messages, and per NWC-321 the
	// note must either be delivered or the request rejected before payment
	if payParams.PayerNote != "" {
		publishError(&models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: "payer_note is not supported: only BOLT-12 payments support payer-provided messages",
		})
		return
	}

	bolt11 := strings.ToLower(bip321.bolt11)
	paymentRequest, err := decodepay.Decodepay(bolt11)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           app.ID,
			"bolt11":           bolt11,
		}).WithError(err).Error("Failed to decode bolt11 invoice")

		publishError(&models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: fmt.Sprintf("Failed to decode bolt11 invoice: %s", err.Error()),
		})
		return
	}

	// the invoice must be for the network this node runs on
	expectedPrefix := networkToInvoicePrefix(nodeNetwork)
	if expectedPrefix != "" && !strings.EqualFold(paymentRequest.Currency, expectedPrefix) {
		publishError(&models.Error{
			Code:    constants.ERROR_UNSUPPORTED_NETWORK,
			Message: fmt.Sprintf("the payment instruction is for a different network than the wallet network (%s)", nodeNetwork),
		})
		return
	}

	amountMsat, nip47Error := resolvePayAmount(uint64(paymentRequest.MSatoshi), payParams.Amount, bip321.amountMsat)
	if nip47Error != nil {
		publishError(nip47Error)
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
		"app_id":           app.ID,
		"bolt11":           bolt11,
	}).Info("Sending payment")

	transaction, err := controller.transactionsService.SendPaymentSync(bolt11, amountMsat, payParams.Metadata, controller.lnClient, &app.ID, &requestEventId)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           app.ID,
			"bolt11":           bolt11,
		}).WithError(err).Error("Failed to send payment")

		publishError(mapNip47Error(err))
		return
	}

	controller.publishPayResult(nip47Request, transaction, instructionTypeBolt11, publishResponse)
}

// payBolt12 pays the BOLT-12 offer instruction of a BIP-321 URI.
func (controller *nip47Controller) payBolt12(ctx context.Context, nip47Request *models.Request, requestEventId uint, app *db.App, publishResponse publishFunc, publishError func(*models.Error), payParams *payParams, bip321 *bip321Payment, nodeNetwork string) {
	offer := strings.ToLower(bip321.bolt12)
	offerInfo, err := controller.lnClient.DecodeOffer(ctx, offer)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           app.ID,
			"offer":            offer,
		}).WithError(err).Error("Failed to decode bolt12 offer")

		publishError(&models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: fmt.Sprintf("Failed to decode bolt12 offer: %s", err.Error()),
		})
		return
	}

	if offerInfo.Expired {
		publishError(&models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: "the payment instruction has expired",
		})
		return
	}

	// offers with a chain restriction must match the network this node runs on
	if len(offerInfo.Chains) > 0 && !offerChainsContainNetwork(offerInfo.Chains, nodeNetwork) {
		publishError(&models.Error{
			Code:    constants.ERROR_UNSUPPORTED_NETWORK,
			Message: fmt.Sprintf("the payment instruction is for a different network than the wallet network (%s)", nodeNetwork),
		})
		return
	}

	var offerAmountMsat uint64
	if offerInfo.AmountMsat != nil {
		offerAmountMsat = *offerInfo.AmountMsat
	}
	amountMsat, nip47Error := resolvePayAmount(offerAmountMsat, payParams.Amount, bip321.amountMsat)
	if nip47Error != nil {
		publishError(nip47Error)
		return
	}

	logger.Logger.WithFields(logrus.Fields{
		"request_event_id": requestEventId,
		"app_id":           app.ID,
		"offer":            offer,
	}).Info("Sending BOLT-12 payment")

	transaction, err := controller.transactionsService.PayOfferSync(ctx, offer, offerInfo, amountMsat, payParams.PayerNote, payParams.Metadata, controller.lnClient, &app.ID, &requestEventId)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"request_event_id": requestEventId,
			"app_id":           app.ID,
			"offer":            offer,
		}).WithError(err).Error("Failed to pay BOLT-12 offer")

		publishError(mapNip47Error(err))
		return
	}

	controller.publishPayResult(nip47Request, transaction, instructionTypeBolt12, publishResponse)
}

// publishPayResult publishes the NWC-321 pay result for a settled transaction.
// If the backend did not return a payment hash, the DB transaction ID is used
// as the wallet-scoped transaction identifier.
func (controller *nip47Controller) publishPayResult(nip47Request *models.Request, transaction *transactions.Transaction, instructionType string, publishResponse publishFunc) {
	transactionId := transaction.PaymentHash
	if transactionId == "" {
		transactionId = strconv.FormatUint(uint64(transaction.ID), 10)
	}

	var settledAt *int64
	if transaction.SettledAt != nil {
		settledAtUnix := transaction.SettledAt.Unix()
		settledAt = &settledAtUnix
	}

	preimage := ""
	if transaction.Preimage != nil {
		preimage = *transaction.Preimage
	}

	publishResponse(&models.Response{
		ResultType: nip47Request.Method,
		Result: payResult{
			TransactionId:   transactionId,
			State:           strings.ToLower(transaction.State),
			InstructionType: instructionType,
			Amount:          transaction.AmountMsat,
			FeesPaid:        transaction.FeeMsat,
			PaymentHash:     transaction.PaymentHash,
			Preimage:        preimage,
			CreatedAt:       transaction.CreatedAt.Unix(),
			SettledAt:       settledAt,
		},
	}, nostr.Tags{})
}

// parseBip321Uri parses a BIP-321 URI and returns the BOLT-11 invoice from its
// "lightning" parameter and the optional "amount" parameter, or a NIP-47 error
// if the URI is invalid or contains no supported payment instruction.
func parseBip321Uri(payment string) (*bip321Payment, *models.Error) {
	parsed, err := url.Parse(payment)
	if err != nil || !strings.EqualFold(parsed.Scheme, "bitcoin") {
		return nil, &models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: "payment must be a valid BIP-321 URI",
		}
	}

	query, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return nil, &models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: "payment must be a valid BIP-321 URI",
		}
	}

	result := &bip321Payment{}
	// BIP-321 URIs may be fully uppercased (e.g. for QR codes)
	for key, values := range query {
		// per BIP-321, a URI with an unknown required ("req-" prefixed)
		// parameter must be considered invalid. This includes "req-pop",
		// as we cannot open proof-of-payment callbacks. The optional "pop"
		// and other optional parameters we do not understand are ignored,
		// as BIP-321 permits.
		if strings.HasPrefix(strings.ToLower(key), "req-") {
			return nil, &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: fmt.Sprintf("unsupported required parameter in BIP-321 URI: %s", key),
			}
		}
		if strings.EqualFold(key, "lightning") && len(values) > 0 && values[0] != "" {
			result.bolt11 = values[0]
		}
		if strings.EqualFold(key, "lno") && len(values) > 0 && values[0] != "" {
			result.bolt12 = values[0]
		}
		if strings.EqualFold(key, "amount") && len(values) > 0 && values[0] != "" {
			amountMsat, err := parseBtcAmountToMsat(values[0])
			if err != nil {
				return nil, &models.Error{
					Code:    constants.ERROR_BAD_REQUEST,
					Message: fmt.Sprintf("invalid amount parameter in BIP-321 URI: %s", values[0]),
				}
			}
			result.amountMsat = &amountMsat
		}
	}

	if result.bolt11 == "" && result.bolt12 == "" {
		return nil, &models.Error{
			Code:    constants.ERROR_UNSUPPORTED_PAYMENT_INSTRUCTION,
			Message: "no supported payment instruction found",
		}
	}

	return result, nil
}

// resolvePayAmount validates the invoice amount against the amounts provided
// in the request params and the BIP-321 URI, per NWC-321: "The wallet service
// MUST reject conflicting or invalid amounts before payment." It returns the
// amount to pay for a zero-amount invoice, or nil if the invoice has one.
func resolvePayAmount(invoiceMsat uint64, paramAmountMsat *uint64, uriAmountMsat *uint64) (*uint64, *models.Error) {
	if paramAmountMsat != nil && *paramAmountMsat == 0 {
		return nil, &models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: "amount must be greater than 0",
		}
	}
	if uriAmountMsat != nil && *uriAmountMsat == 0 {
		return nil, &models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: "amount parameter in BIP-321 URI must be greater than 0",
		}
	}

	if invoiceMsat > 0 {
		if paramAmountMsat != nil && *paramAmountMsat != invoiceMsat {
			return nil, &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: "amount conflicts with the amount of the payment instruction",
			}
		}
		if uriAmountMsat != nil && *uriAmountMsat != invoiceMsat {
			return nil, &models.Error{
				Code:    constants.ERROR_BAD_REQUEST,
				Message: "amount parameter in BIP-321 URI conflicts with the amount of the payment instruction",
			}
		}
		return nil, nil
	}

	if paramAmountMsat != nil && uriAmountMsat != nil && *paramAmountMsat != *uriAmountMsat {
		return nil, &models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: "amount conflicts with the amount parameter in the BIP-321 URI",
		}
	}
	amountMsat := paramAmountMsat
	if amountMsat == nil {
		amountMsat = uriAmountMsat
	}
	if amountMsat == nil {
		return nil, &models.Error{
			Code:    constants.ERROR_BAD_REQUEST,
			Message: "amount is required when the payment instruction has no amount",
		}
	}
	return amountMsat, nil
}

// parseBtcAmountToMsat converts a BIP-321 decimal BTC amount (e.g. "0.00000123")
// to millisatoshis.
func parseBtcAmountToMsat(value string) (uint64, error) {
	intPart, fracPart, _ := strings.Cut(value, ".")
	if intPart == "" && fracPart == "" {
		return 0, fmt.Errorf("empty amount")
	}
	if intPart == "" {
		intPart = "0"
	}
	// millisatoshi precision is 11 decimal places of a bitcoin
	if len(fracPart) > 11 {
		return 0, fmt.Errorf("too many decimal places")
	}
	fracPart = fracPart + strings.Repeat("0", 11-len(fracPart))

	whole, err := strconv.ParseUint(intPart, 10, 64)
	if err != nil {
		return 0, err
	}
	if whole > 21_000_000 {
		return 0, fmt.Errorf("amount too large")
	}
	frac, err := strconv.ParseUint(fracPart, 10, 64)
	if err != nil {
		return 0, err
	}
	return whole*100_000_000_000 + frac, nil
}

// offerChainsContainNetwork returns true if an offer's chain restriction
// includes the network this node runs on.
func offerChainsContainNetwork(offerChains []string, nodeNetwork string) bool {
	for _, chain := range offerChains {
		if strings.EqualFold(chain, nodeNetwork) {
			return true
		}
	}
	return false
}

// networkToInvoicePrefix maps an LNClient network name to the BOLT-11 invoice
// human-readable-part network prefix. Returns "" for unknown network names.
func networkToInvoicePrefix(network string) string {
	switch strings.ToLower(network) {
	case "bitcoin", "mainnet":
		return "bc"
	case "testnet", "testnet3", "testnet4":
		return "tb"
	case "signet", "mutinynet":
		return "tbs"
	case "regtest":
		return "bcrt"
	}
	return ""
}
