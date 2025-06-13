package swaps

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/BoltzExchange/boltz-client/v2/pkg/boltz"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/transactions"
	"github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip39"
)

type swapsService struct {
	autoSwapOutCancelFn context.CancelFunc
	cfg                 config.Config
	eventPublisher      events.EventPublisher
	transactionsService transactions.TransactionsService
	boltzApi            *boltz.Api
}

type SwapsService interface {
	StopAutoSwap()
	EnableAutoSwapOut(ctx context.Context, lnClient lnclient.LNClient) error
	SwapOut(ctx context.Context, amount uint64, destination string, lnClient lnclient.LNClient) (*SwapOutResponse, error)
	SwapIn(ctx context.Context, amount uint64, lnClient lnclient.LNClient) (*SwapInResponse, error)
	CalculateSwapOutFee() (*SwapFees, error)
	CalculateSwapInFee() (*SwapFees, error)
}

const (
	AlbySwapServiceFee = 1.0
)

type FeeRates struct {
	FastestFee  uint64 `json:"fastestFee"`
	HalfHourFee uint64 `json:"halfHourFee"`
	HourFee     uint64 `json:"hourFee"`
	EconomyFee  uint64 `json:"economyFee"`
	MinimumFee  uint64 `json:"minimumFee"`
}

type SwapFees struct {
	AlbyServiceFee  float64 `json:"albyServiceFee"`
	BoltzServiceFee float64 `json:"boltzServiceFee"`
	BoltzNetworkFee uint64  `json:"boltzNetworkFee"`
}

type SwapOutResponse struct {
	TxId        string `json:"txId"`
	SwapId      string `json:"swapId"`
	PaymentHash string `json:"paymentHash"`
}

type SwapInResponse struct {
	OnchainAddress  string `json:"onchainAddress"`
	AmountToDeposit uint64 `json:"amountToDeposit"`
	PaymentHash     string `json:"paymentHash"`
}

func NewSwapsService(cfg config.Config, eventPublisher events.EventPublisher, transactionsService transactions.TransactionsService) SwapsService {
	return &swapsService{
		cfg:                 cfg,
		eventPublisher:      eventPublisher,
		transactionsService: transactionsService,
		boltzApi:            &boltz.Api{URL: cfg.GetEnv().BoltzApi},
	}
}

func (svc *swapsService) StopAutoSwap() {
	if svc.autoSwapOutCancelFn != nil {
		logger.Logger.Info("Stopping auto swap out service...")
		svc.autoSwapOutCancelFn()
		logger.Logger.Info("Auto swap out service stopped")
	}
}

func (svc *swapsService) EnableAutoSwapOut(ctx context.Context, lnClient lnclient.LNClient) error {
	svc.StopAutoSwap()

	ctx, cancelFn := context.WithCancel(ctx)
	swapDestination, _ := svc.cfg.Get(config.AutoSwapDestinationKey, "")
	balanceThresholdStr, _ := svc.cfg.Get(config.AutoSwapBalanceThresholdKey, "")
	amountStr, _ := svc.cfg.Get(config.AutoSwapAmountKey, "")

	if balanceThresholdStr == "" || amountStr == "" {
		cancelFn()
		logger.Logger.Info("Auto swap not configured")
		return nil
	}

	balanceThreshold, err := strconv.ParseUint(balanceThresholdStr, 10, 64)
	if err != nil {
		cancelFn()
		return errors.New("invalid auto swap configuration")
	}

	amount, err := strconv.ParseUint(amountStr, 10, 64)
	if err != nil {
		cancelFn()
		return errors.New("invalid auto swap configuration")
	}

	logger.Logger.Info("Starting auto swap workflow")

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for {
			select {
			case <-ticker.C:
				logger.Logger.Debug("Checking to see if we can swap")
				balance, err := lnClient.GetBalances(ctx, false)
				if err != nil {
					logger.Logger.WithError(err).Error("Failed to get balance")
					return
				}
				lightningBalance := uint64(balance.Lightning.TotalSpendable)
				balanceThresholdMilliSats := balanceThreshold * 1000
				if lightningBalance < balanceThresholdMilliSats {
					logger.Logger.Info("Threshold requirements not met for swap, ignoring")
					return
				}
				logger.Logger.WithFields(logrus.Fields{
					"amount":      amount,
					"destination": swapDestination,
				}).Info("Initiating swap")
				_, err = svc.SwapOut(ctx, amount, swapDestination, lnClient)
				if err != nil {
					logger.Logger.WithError(err).Error("Failed to swap")
				}
			case <-ctx.Done():
				logger.Logger.Info("Stopping auto swap workflow")
				return
			}
		}
	}()

	svc.autoSwapOutCancelFn = cancelFn

	return nil
}

func (svc *swapsService) SwapOut(ctx context.Context, amount uint64, destination string, lnClient lnclient.LNClient) (*SwapOutResponse, error) {
	if destination == "" {
		var err error
		destination, err = svc.cfg.Get(config.OnchainAddressKey, "")
		if err != nil {
			return nil, fmt.Errorf("could not get onchain address from config: %s", err)
		}
	}

	var network, err = boltz.ParseChain(svc.cfg.GetNetwork())
	if err != nil {
		return nil, err
	}

	// TODO: use own keys
	ourKeys, err := btcec.NewPrivateKey()
	if err != nil {
		return nil, err
	}

	preimage := make([]byte, 32)
	_, err = rand.Read(preimage)
	if err != nil {
		return nil, err
	}
	preimageHash := sha256.Sum256(preimage)
	paymentHash := hex.EncodeToString(preimageHash[:])

	reversePairs, err := svc.boltzApi.GetReversePairs()
	if err != nil {
		return nil, fmt.Errorf("could not get reverse pairs: %s", err)
	}

	pair := boltz.Pair{From: boltz.CurrencyBtc, To: boltz.CurrencyBtc}
	pairInfo, err := boltz.FindPair(pair, reversePairs)
	if err != nil {
		return nil, fmt.Errorf("could not find reverse pair: %s", err)
	}

	fees := pairInfo.Fees
	serviceFeePercentage := boltz.Percentage(fees.Percentage)

	serviceFee := boltz.CalculatePercentage(serviceFeePercentage, amount)
	networkFee := fees.MinerFees.Lockup + fees.MinerFees.Claim

	logger.Logger.WithFields(logrus.Fields{
		"serviceFee": serviceFee,
		"networkFee": networkFee,
	}).Info("Calculated fees for swap out")

	albyFee := &boltz.ExtraFees{
		Percentage: AlbySwapServiceFee,
		Id:         "albyServiceFee",
	}

	swap, err := svc.boltzApi.CreateReverseSwap(boltz.CreateReverseSwapRequest{
		From:           boltz.CurrencyBtc,
		To:             boltz.CurrencyBtc,
		ClaimPublicKey: ourKeys.PubKey().SerializeCompressed(),
		PreimageHash:   preimageHash[:],
		InvoiceAmount:  amount,
		Description:    "Boltz swap invoice",
		PairHash:       pairInfo.Hash,
		ReferralId:     "alby",
		ExtraFees:      albyFee,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create swap: %s", err)
	}

	boltzPubKey, err := btcec.ParsePubKey(swap.RefundPublicKey)
	if err != nil {
		return nil, err
	}

	tree := swap.SwapTree.Deserialize()
	if err := tree.Init(boltz.CurrencyBtc, true, ourKeys, boltzPubKey); err != nil {
		return nil, err
	}

	if err := tree.Check(boltz.ReverseSwap, swap.TimeoutBlockHeight, preimageHash[:]); err != nil {
		return nil, err
	}

	logger.Logger.WithField("swap", swap).Info("Swap created")

	txCh := make(chan string, 1)
	errCh := make(chan error, 1)

	boltzWs := svc.boltzApi.NewWebsocket()
	if err := boltzWs.Connect(); err != nil {
		return nil, fmt.Errorf("could not connect to Boltz websocket: %w", err)
	}

	if err := boltzWs.Subscribe([]string{swap.Id}); err != nil {
		_ = boltzWs.Close()
		return nil, err
	}

	go func() {
		defer func() {
			if err := boltzWs.Close(); err != nil {
				logger.Logger.WithError(err).Error("Failed to close boltz websocket")
			}
		}()

		updatesCh := boltzWs.Updates
		paymentErrorCh := make(chan error, 1)

		for {
			select {
			case <-ctx.Done():
				logger.Logger.WithError(ctx.Err()).Error("Reverse swap context cancelled")
				errCh <- ctx.Err()
				return
			case err := <-paymentErrorCh:
				errCh <- err
				return
			case update, ok := <-updatesCh:
				if !ok {
					errCh <- errors.New("boltz websocket closed unexpectedly")
					return
				}

				parsedStatus := boltz.ParseEvent(update.Status)

				switch parsedStatus {
				case boltz.SwapCreated:
					logger.Logger.WithFields(logrus.Fields{
						"swap":   swap,
						"update": update,
					}).Info("Paying the swap invoice")
					go func() {
						// TODO: move to swaps table + review info needed for refund
						metadata := map[string]interface{}{
							"swapId":        swap.Id,
							"onchainAmount": swap.OnchainAmount,
							"refundPubkey":  swap.RefundPublicKey,
						}
						sendPaymentTimeout := int64(3600)
						holdInvoicePayment, err := svc.transactionsService.SendPaymentSync(ctx, swap.Invoice, nil, metadata, lnClient, nil, nil, &sendPaymentTimeout)
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swap":   swap,
								"update": update,
							}).Error("Error paying the swap invoice")
							paymentErrorCh <- err
							return
						}
						logger.Logger.WithField("swapId", swap.Id).Info("Initiated swap invoice payment")
						if holdInvoicePayment.PaymentHash != paymentHash {
							paymentErrorCh <- errors.New("swap hold payment hash mismatch")
							return
						}
					}()
				case boltz.TransactionMempool:
					logger.Logger.WithFields(logrus.Fields{
						"swapId":      swap.Id,
						"transaction": update.Transaction,
					}).Info("Lockup transaction found in mempool")
					txCh <- update.Transaction.Id

					holdInvoicePayment, err := svc.transactionsService.LookupTransaction(ctx, paymentHash, nil, lnClient, nil)
					if err != nil {
						logger.Logger.WithError(err).WithField("payment_hash", paymentHash).Error("Failed to lookup swap hold invoice payment")
						return
					}

					// TODO: move to swaps table + review info needed for refund
					var metadata map[string]interface{}
					jsonErr := json.Unmarshal(holdInvoicePayment.Metadata, &metadata)
					if jsonErr != nil {
						logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
							"payment_hash": paymentHash,
						}).Error("Failed to deserialize transaction metadata")
						return
					}
					metadata["lockupTransactionId"] = update.Transaction.Id
					err = svc.transactionsService.SetTransactionMetadata(ctx, holdInvoicePayment.ID, metadata)
					if err != nil {
						logger.Logger.WithError(err).WithFields(logrus.Fields{
							"payment_hash":          paymentHash,
							"lookup_transaction_id": update.Transaction.Id,
						}).Error("Failed to add lookup transaction id to lightning payment metadata")
						return
					}

				case boltz.TransactionConfirmed:
					logger.Logger.WithFields(logrus.Fields{
						"swapId":      swap.Id,
						"transaction": update.Transaction,
					}).Info("Lockup transaction confirmed in mempool")
					lockupTransaction, err := boltz.NewTxFromHex(boltz.CurrencyBtc, update.Transaction.Hex, nil)
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to build lockup tx from hex")
						return
					}

					vout, _, err := lockupTransaction.FindVout(network, swap.LockupAddress)
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to find lockup address output")
						return
					}

					feeRates, err := svc.getFeeRates()
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to fetch fee rate to create claim transaction")
						return
					}

					claimTransaction, _, err := boltz.ConstructTransaction(
						network,
						boltz.CurrencyBtc,
						[]boltz.OutputDetails{
							{
								SwapId:            swap.Id,
								SwapType:          boltz.ReverseSwap,
								Address:           destination,
								LockupTransaction: lockupTransaction,
								Vout:              vout,
								Preimage:          preimage,
								PrivateKey:        ourKeys,
								SwapTree:          tree,
								Cooperative:       true,
							},
						},
						float64(feeRates.FastestFee),
						svc.boltzApi,
					)
					if err != nil {
						logger.Logger.WithError(err).Error("Could not create claim transaction")
						return
					}

					txHex, err := claimTransaction.Serialize()
					if err != nil {
						logger.Logger.WithError(err).Error("Could not serialize claim transaction")
						return
					}

					// TODO: Replace with LNClient broadcast method to avoid trusting boltz
					txId, err := svc.boltzApi.BroadcastTransaction(boltz.CurrencyBtc, txHex)
					if err != nil {
						logger.Logger.WithError(err).Error("Could not broadcast transaction")
						return
					}

					logger.Logger.WithField("txId", txId).Info("Transaction broadcasted")
				case boltz.InvoiceSettled:
					logger.Logger.WithField("swapId", swap.Id).Info("Swap succeeded")
					svc.eventPublisher.Publish(&events.Event{
						Event: "nwc_swap_succeeded",
						Properties: map[string]interface{}{
							"swapId":        swap.Id,
							"invoice":       swap.Invoice,
							"onchainAmount": swap.OnchainAmount,
							"refundPubkey":  swap.RefundPublicKey,
						},
					})
					return
				}
			}
		}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case txid := <-txCh:
		return &SwapOutResponse{
			TxId:        txid,
			SwapId:      swap.Id,
			PaymentHash: paymentHash,
		}, nil
	}
}

func (svc *swapsService) SwapIn(ctx context.Context, amount uint64, lnClient lnclient.LNClient) (*SwapInResponse, error) {
	amountMSat := amount * 1000
	invoice, err := svc.transactionsService.MakeInvoice(ctx, amountMSat, "Boltz swap in", "", 0, nil, lnClient, nil, nil)
	if err != nil {
		return nil, err
	}

	network, err := boltz.ParseChain(svc.cfg.GetNetwork())
	if err != nil {
		return nil, err
	}

	entropy, err := bip39.NewEntropy(128)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to generate entropy for mnemonic")
		return nil, err
	}
	mnemonic, err := bip39.NewMnemonic(entropy)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to generate mnemonic")
		return nil, err
	}

	// FIXME: use HD key and derivation path: https://github.com/BoltzExchange/boltz-web-app/blob/main/src/utils/rescueFile.ts#L36

	ourKey, _ := btcec.PrivKeyFromBytes(bip39.NewSeed(mnemonic, ""))

	submarinePairs, err := svc.boltzApi.GetSubmarinePairs()
	if err != nil {
		return nil, fmt.Errorf("could not get submarine pairs: %s", err)
	}

	pair := boltz.Pair{From: boltz.CurrencyBtc, To: boltz.CurrencyBtc}
	pairInfo, err := boltz.FindPair(pair, submarinePairs)
	if err != nil {
		return nil, fmt.Errorf("could not find submarine pair: %s", err)
	}

	fees := pairInfo.Fees
	serviceFeePercentage := boltz.Percentage(fees.Percentage)

	serviceFee := boltz.CalculatePercentage(serviceFeePercentage, amount)
	networkFee := fees.MinerFees

	logger.Logger.WithFields(logrus.Fields{
		"serviceFee": serviceFee,
		"networkFee": networkFee,
	}).Info("Calculated fees for swap in")

	albyFee := &boltz.ExtraFees{
		Percentage: AlbySwapServiceFee,
		Id:         "albyServiceFee",
	}

	swap, err := svc.boltzApi.CreateSwap(boltz.CreateSwapRequest{
		From:            boltz.CurrencyBtc,
		To:              boltz.CurrencyBtc,
		RefundPublicKey: ourKey.PubKey().SerializeCompressed(),
		Invoice:         invoice.PaymentRequest,
		PairHash:        pairInfo.Hash,
		ReferralId:      "alby",
		ExtraFees:       albyFee,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create swap: %s", err)
	}

	// TODO: move to swaps table + review info needed for refund
	metadata := map[string]interface{}{
		"swapId":      swap.Id,
		"claimPubkey": swap.ClaimPublicKey,
		"mnemonic":    mnemonic,
		"amount":      swap.ExpectedAmount,
		"address":     swap.Address,
	}

	err = svc.transactionsService.SetTransactionMetadata(ctx, invoice.ID, metadata)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"payment_hash": invoice.PaymentHash,
			"metadata":     metadata,
		}).Error("Failed to add swap metadata to lightning payment")
		return nil, err
	}

	boltzPubKey, err := btcec.ParsePubKey(swap.ClaimPublicKey)
	if err != nil {
		return nil, err
	}

	tree := swap.SwapTree.Deserialize()
	if err := tree.Init(boltz.CurrencyBtc, false, ourKey, boltzPubKey); err != nil {
		return nil, err
	}

	decodedPreimageHash, err := hex.DecodeString(invoice.PaymentHash)
	if err != nil {
		return nil, fmt.Errorf("invalid preimage hash: %v", err)
	}

	if err := tree.Check(boltz.NormalSwap, swap.TimeoutBlockHeight, decodedPreimageHash); err != nil {
		return nil, err
	}

	if err := tree.CheckAddress(swap.Address, network, nil); err != nil {
		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{
		"swap":         swap,
		"payment_hash": invoice.PaymentHash,
	}).Info("Swap created")

	boltzWs := svc.boltzApi.NewWebsocket()
	if err := boltzWs.Connect(); err != nil {
		return nil, fmt.Errorf("could not connect to Boltz websocket: %w", err)
	}

	if err := boltzWs.Subscribe([]string{swap.Id}); err != nil {
		_ = boltzWs.Close()
		return nil, err
	}

	// TODO: add a timeout equivalent to invoice expiry so it doesn't keep waiting forever
	go func() {
		var txHex string

		defer func() {
			if err := boltzWs.Close(); err != nil {
				logger.Logger.WithError(err).Error("Failed to close boltz websocket")
			}
		}()

		updatesCh := boltzWs.Updates

		for {
			select {
			case <-ctx.Done():
				logger.Logger.WithError(ctx.Err()).Error("Submarine swap context cancelled")
				return
			case update, ok := <-updatesCh:
				if !ok {
					// TODO: should we reconnect here?
					logger.Logger.Error("boltz websocket closed unexpectedly")
					return
				}

				parsedStatus := boltz.ParseEvent(update.Status)

				switch parsedStatus {
				case boltz.TransactionMempool:
					logger.Logger.WithFields(logrus.Fields{
						"swapId":      swap.Id,
						"transaction": update.Transaction,
					}).Info("Lockup transaction found in mempool")
					txHex = update.Transaction.Hex
				case boltz.TransactionConfirmed:
					logger.Logger.WithFields(logrus.Fields{
						"swapId":      swap.Id,
						"transaction": update.Transaction,
					}).Info("Lockup transaction confirmed in mempool")
				case boltz.TransactionClaimPending:
					logger.Logger.WithFields(logrus.Fields{
						"swapId":      swap.Id,
						"transaction": update.Transaction,
					}).Info("Sending partial signature to boltz to claim the payment")
					claimDetails, err := svc.boltzApi.GetSwapClaimDetails(swap.Id)
					if err != nil {
						logger.Logger.WithError(err).Error("Could not get claim details from Boltz")
						return
					}

					preimageHash := sha256.Sum256(claimDetails.Preimage)
					if !bytes.Equal(decodedPreimageHash, preimageHash[:]) {
						logger.Logger.WithField("preimage", claimDetails.Preimage).Error("Boltz returned wrong preimage")
						return
					}

					session, _ := boltz.NewSigningSession(tree)
					partial, err := session.Sign(claimDetails.TransactionHash, claimDetails.PubNonce)
					if err != nil {
						logger.Logger.WithError(err).Error("Could not create partial signature")
						return
					}

					if err := svc.boltzApi.SendSwapClaimSignature(swap.Id, partial); err != nil {
						logger.Logger.WithError(err).Error("Could not send partial signature to Boltz")
						return
					}
				case boltz.TransactionClaimed:
					logger.Logger.WithField("swapId", swap.Id).Info("Swap succeeded")
					svc.eventPublisher.Publish(&events.Event{
						Event: "nwc_swap_succeeded",
						Properties: map[string]interface{}{
							"swapType":    "in",
							"swapId":      swap.Id,
							"address":     swap.Address,
							"amount":      swap.ExpectedAmount,
							"claimPubkey": swap.ClaimPublicKey,
						},
					})
					return
				case boltz.TransactionLockupFailed, boltz.InvoiceFailedToPay:
					logger.Logger.WithFields(logrus.Fields{
						"swapId": swap.Id,
						"update": update,
					}).Info("Swap in failed, initiating refund")
					lockupTransaction, err := boltz.NewTxFromHex(boltz.CurrencyBtc, txHex, nil)
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to build lockup tx from hex")
						return
					}
					vout, _, err := lockupTransaction.FindVout(network, swap.Address)
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to find lockup address output")
						return
					}
					feeRates, err := svc.getFeeRates()
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to fetch fee rate to create claim transaction")
						return
					}
					// TODO: generate a new key
					address, err := svc.cfg.Get(config.OnchainAddressKey, "")
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to get on-chain address from config")
						return
					}
					refundTransaction, _, err := boltz.ConstructTransaction(
						network,
						boltz.CurrencyBtc,
						[]boltz.OutputDetails{
							{
								SwapId:             swap.Id,
								SwapType:           boltz.NormalSwap,
								Address:            address,
								LockupTransaction:  lockupTransaction,
								TimeoutBlockHeight: swap.TimeoutBlockHeight,
								Vout:               vout,
								PrivateKey:         ourKey,
								SwapTree:           tree,
								Cooperative:        true,
							},
						},
						float64(feeRates.FastestFee),
						svc.boltzApi,
					)
					if err != nil {
						logger.Logger.WithError(err).Error("Could not create claim transaction")
						return
					}

					txHex, err := refundTransaction.Serialize()
					if err != nil {
						logger.Logger.WithError(err).Error("Could not serialize refund transaction")
						return
					}

					// TODO: Replace with LNClient broadcast method to avoid trusting boltz
					txId, err := svc.boltzApi.BroadcastTransaction(boltz.CurrencyBtc, txHex)
					if err != nil {
						logger.Logger.WithError(err).Error("Could not broadcast transaction")
						return
					}

					logger.Logger.WithField("txId", txId).Info("Transaction broadcasted")
					return
				}
			}
		}
	}()

	return &SwapInResponse{
		OnchainAddress:  swap.Address,
		AmountToDeposit: swap.ExpectedAmount,
		PaymentHash:     invoice.PaymentHash,
	}, nil
}

func (svc *swapsService) CalculateSwapOutFee() (*SwapFees, error) {
	reversePairs, err := svc.boltzApi.GetReversePairs()
	if err != nil {
		return nil, fmt.Errorf("could not get reverse pairs: %s", err)
	}

	pair := boltz.Pair{From: boltz.CurrencyBtc, To: boltz.CurrencyBtc}
	pairInfo, err := boltz.FindPair(pair, reversePairs)
	if err != nil {
		return nil, fmt.Errorf("could not find reverse pair: %s", err)
	}

	fees := pairInfo.Fees
	networkFee := fees.MinerFees.Lockup + fees.MinerFees.Claim

	return &SwapFees{
		AlbyServiceFee:  AlbySwapServiceFee,
		BoltzServiceFee: fees.Percentage,
		BoltzNetworkFee: networkFee,
	}, nil
}

func (svc *swapsService) CalculateSwapInFee() (*SwapFees, error) {
	submarinePairs, err := svc.boltzApi.GetSubmarinePairs()
	if err != nil {
		return nil, fmt.Errorf("could not get reverse pairs: %s", err)
	}

	pair := boltz.Pair{From: boltz.CurrencyBtc, To: boltz.CurrencyBtc}
	pairInfo, err := boltz.FindPair(pair, submarinePairs)
	if err != nil {
		return nil, fmt.Errorf("could not find reverse pair: %s", err)
	}

	fees := pairInfo.Fees

	return &SwapFees{
		AlbyServiceFee:  AlbySwapServiceFee,
		BoltzServiceFee: fees.Percentage,
		BoltzNetworkFee: fees.MinerFees,
	}, nil
}

func (svc *swapsService) getFeeRates() (*FeeRates, error) {
	url := svc.cfg.GetEnv().MempoolApi + "/v1/fees/recommended"
	// force mainnet fees since testnet is so often unavailable
	url = strings.ReplaceAll(url, "testnet/", "")

	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create http request")
		return nil, err
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to send request")
		return nil, err
	}

	if res.StatusCode >= 300 {
		return nil, errors.New("failed to fetch fee rates: unexpected status: " + res.Status)
	}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read fee rates response body")
		return nil, errors.New("failed to read response body")
	}

	var rates FeeRates
	jsonErr := json.Unmarshal(body, &rates)
	if jsonErr != nil {
		logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize fee rates json %s %s", url, string(body))
	}
	return &rates, nil
}
