package swaps

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/BoltzExchange/boltz-client/v2/pkg/boltz"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/transactions"
	"github.com/sirupsen/logrus"
)

type swapsService struct {
	cancelFn            context.CancelFunc
	cfg                 config.Config
	eventPublisher      events.EventPublisher
	transactionsService transactions.TransactionsService
	boltzApi            *boltz.Api
}

type SwapsService interface {
	EnableAutoSwap(ctx context.Context, lnClient lnclient.LNClient) error
	StopAutoSwap()
	CalculateFee() (*SwapFees, error)
	ReverseSwap(ctx context.Context, amount uint64, destination string, lnClient lnclient.LNClient) (string, error)
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

func NewSwapsService(cfg config.Config, eventPublisher events.EventPublisher, transactionsService transactions.TransactionsService) *swapsService {
	return &swapsService{
		cfg:                 cfg,
		eventPublisher:      eventPublisher,
		transactionsService: transactionsService,
		boltzApi:            &boltz.Api{URL: cfg.GetEnv().BoltzApi},
	}
}

func (svc *swapsService) EnableAutoSwap(ctx context.Context, lnClient lnclient.LNClient) error {
	// stop any existing swap process
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

	parsedBalanceThreshold, err := strconv.ParseUint(balanceThresholdStr, 10, 64)
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
				balanceThresholdMilliSats := parsedBalanceThreshold * 1000
				if lightningBalance >= balanceThresholdMilliSats {
					logger.Logger.WithFields(logrus.Fields{
						"amount":      amount,
						"destination": swapDestination,
					}).Info("Initiating swap")
					_, err := svc.ReverseSwap(ctx, amount, swapDestination, lnClient)
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to swap")
					}
				} else {
					logger.Logger.Info("Threshold requirements not met for swap, ignoring")
				}
			case <-ctx.Done():
				logger.Logger.Info("Stopping auto swap workflow")
				return
			}
		}
	}()

	svc.cancelFn = cancelFn

	return nil
}

func (svc *swapsService) StopAutoSwap() {
	if svc.cancelFn != nil {
		logger.Logger.Info("Stopping swap service...")
		svc.cancelFn()
		logger.Logger.Info("swap service stopped")
	}
}

func (svc *swapsService) ReverseSwap(ctx context.Context, amount uint64, destination string, lnClient lnclient.LNClient) (string, error) {
	if destination == "" {
		var err error
		destination, err = svc.cfg.Get(config.OnchainAddressKey, "")
		if err != nil {
			return "", fmt.Errorf("could not get onchain address from config: %s", err)
		}
	}

	var network, err = boltz.ParseChain(svc.cfg.GetNetwork())
	if err != nil {
		return "", err
	}

	ourKeys, err := btcec.NewPrivateKey()
	if err != nil {
		return "", err
	}

	preimage := make([]byte, 32)
	_, err = rand.Read(preimage)
	if err != nil {
		return "", err
	}
	preimageHash := sha256.Sum256(preimage)

	reversePairs, err := svc.boltzApi.GetReversePairs()
	if err != nil {
		return "", fmt.Errorf("could not get reverse pairs: %s", err)
	}

	pair := boltz.Pair{From: boltz.CurrencyBtc, To: boltz.CurrencyBtc}
	pairInfo, err := boltz.FindPair(pair, reversePairs)
	if err != nil {
		return "", fmt.Errorf("could not find reverse pair: %s", err)
	}

	fees := pairInfo.Fees
	serviceFeePercentage := boltz.Percentage(fees.Percentage)

	serviceFee := boltz.CalculatePercentage(serviceFeePercentage, amount)
	networkFee := fees.MinerFees.Lockup + fees.MinerFees.Claim

	logger.Logger.WithFields(logrus.Fields{
		"serviceFee": serviceFee,
		"networkFee": networkFee,
	}).Info("Calculated fees for swap")

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
		return "", fmt.Errorf("could not create swap: %s", err)
	}

	boltzPubKey, err := btcec.ParsePubKey(swap.RefundPublicKey)
	if err != nil {
		return "", err
	}

	tree := swap.SwapTree.Deserialize()
	if err := tree.Init(boltz.CurrencyBtc, true, ourKeys, boltzPubKey); err != nil {
		return "", err
	}

	if err := tree.Check(boltz.ReverseSwap, swap.TimeoutBlockHeight, preimageHash[:]); err != nil {
		return "", err
	}

	logger.Logger.WithField("swap", swap).Info("Swap created")

	txCh := make(chan string, 1)
	errCh := make(chan error, 1)

	boltzWs := svc.boltzApi.NewWebsocket()
	if err := boltzWs.Connect(); err != nil {
		return "", fmt.Errorf("could not connect to Boltz websocket: %w", err)
	}

	if err := boltzWs.Subscribe([]string{swap.Id}); err != nil {
		_ = boltzWs.Close()
		return "", err
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
					err := lnClient.SendPaymentProbes(ctx, swap.Invoice)
					if err != nil {
						logger.Logger.WithField("swapId", swap.Id).Info("Couldn't probe invoice payment, terminating swap")
						errCh <- err
						return
					}
					go func() {
						metadata := map[string]interface{}{
							"swapId":        swap.Id,
							"onchainAmount": swap.OnchainAmount,
							"refundPubkey":  swap.RefundPublicKey,
						}
						sendPaymentTimeout := int64(3600)
						_, err := svc.transactionsService.SendPaymentSync(ctx, swap.Invoice, nil, metadata, lnClient, nil, nil, &sendPaymentTimeout)
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swap":   swap,
								"update": update,
							}).Error("Error paying the swap invoice")
							paymentErrorCh <- err
							return
						}
						logger.Logger.WithField("swapId", swap.Id).Info("Initiated swap invoice payment")
					}()
				case boltz.TransactionMempool:
					logger.Logger.WithFields(logrus.Fields{
						"swapId":      swap.Id,
						"transaction": update.Transaction,
					}).Info("Lockup transaction found in mempool")
					txCh <- update.Transaction.Id
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
						logger.Logger.WithError(err).Error("Could not create claim transaction")
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

					// TODO: Replace with LNClient method
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
		return "", ctx.Err()
	case err := <-errCh:
		return "", err
	case txid := <-txCh:
		return txid, nil
	}
}

func (svc *swapsService) CalculateFee() (*SwapFees, error) {
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

func (svc *swapsService) getFeeRates() (*FeeRates, error) {
	url := svc.cfg.GetEnv().MempoolApi + "/v1/fees/recommended"

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

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return nil, errors.New("failed to read response body")
	}

	var rates FeeRates
	jsonErr := json.Unmarshal(body, &rates)
	if jsonErr != nil {
		logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return nil, fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}
	return &rates, nil
}
