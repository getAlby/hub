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
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/transactions"
	"github.com/sirupsen/logrus"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Swap = db.Swap

type swapsService struct {
	autoSwapOutCancelFn context.CancelFunc
	db                  *gorm.DB
	ctx                 context.Context
	lnClient            lnclient.LNClient
	cfg                 config.Config
	keys                keys.Keys
	eventPublisher      events.EventPublisher
	transactionsService transactions.TransactionsService
	boltzApi            *boltz.Api
}

type SwapsService interface {
	StopAutoSwapOut()
	EnableAutoSwapOut() error
	SwapOut(amount uint64, destination string, autoSwap bool) (*SwapResponse, error)
	SwapIn(amount uint64, autoSwap bool) (*SwapResponse, error)
	CalculateSwapOutFee() (*SwapFees, error)
	CalculateSwapInFee() (*SwapFees, error)
	RefundSwap(swapId, address string) error
	GetSwap(swapId string) (*Swap, error)
	ListSwaps() ([]Swap, error)
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

type TxStatusInfo struct {
	Confirmed   bool   `json:"confirmed"`
	BlockHeight uint32 `json:"block_height"`
	BlockHash   string `json:"block_hash"`
	BlockTime   uint64 `json:"block_time"`
}

type MempoolTx struct {
	TxId   string       `json:"txid"`
	Status TxStatusInfo `json:"status"`
}

type SwapFees struct {
	AlbyServiceFee  float64 `json:"albyServiceFee"`
	BoltzServiceFee float64 `json:"boltzServiceFee"`
	BoltzNetworkFee uint64  `json:"boltzNetworkFee"`
}

type SwapResponse struct {
	SwapId      string `json:"swapId"`
	PaymentHash string `json:"paymentHash"`
}

// TODO: Subscribe to boltz for all pending swaps to update
func NewSwapsService(ctx context.Context, db *gorm.DB, cfg config.Config, keys keys.Keys, eventPublisher events.EventPublisher,
	lnClient lnclient.LNClient, transactionsService transactions.TransactionsService) SwapsService {
	return &swapsService{
		ctx:                 ctx,
		cfg:                 cfg,
		db:                  db,
		keys:                keys,
		eventPublisher:      eventPublisher,
		transactionsService: transactionsService,
		lnClient:            lnClient,
		boltzApi:            &boltz.Api{URL: cfg.GetEnv().BoltzApi},
	}
}

func (svc *swapsService) StopAutoSwapOut() {
	if svc.autoSwapOutCancelFn != nil {
		logger.Logger.Info("Stopping auto swap out service...")
		svc.autoSwapOutCancelFn()
		logger.Logger.Info("Auto swap out service stopped")
	}
}

func (svc *swapsService) EnableAutoSwapOut() error {
	svc.StopAutoSwapOut()

	ctx, cancelFn := context.WithCancel(svc.ctx)
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
				balance, err := svc.lnClient.GetBalances(ctx, false)
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
				_, err = svc.SwapOut(amount, swapDestination, true)
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

func (svc *swapsService) SwapOut(amount uint64, destination string, autoSwap bool) (*SwapResponse, error) {
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

	dbSwap := db.Swap{
		Type:               constants.SWAP_TYPE_OUT,
		State:              constants.SWAP_STATE_PENDING,
		SendAmount:         amount,
		DestinationAddress: destination,
		PaymentHash:        paymentHash,
		AutoSwap:           autoSwap,
	}

	var tree *boltz.SwapTree
	var ourKeys *btcec.PrivateKey
	var swap *boltz.CreateReverseSwapResponse

	err = svc.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&dbSwap).Error
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				err := tx.Model(&dbSwap).Update("state", constants.SWAP_STATE_FAILED).Error
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"dbSwapID":    dbSwap.ID,
						"paymentHash": paymentHash,
					}).Error("Failed to mark swap as failed")
				}
			}
		}()

		ourKeys, err = svc.keys.GetSwapKey(dbSwap.ID)
		if err != nil {
			return fmt.Errorf("error generating swap child private key: %w", err)
		}

		swap, err = svc.boltzApi.CreateReverseSwap(boltz.CreateReverseSwapRequest{
			From:           boltz.CurrencyBtc,
			To:             boltz.CurrencyBtc,
			ClaimPublicKey: ourKeys.PubKey().SerializeCompressed(),
			PreimageHash:   preimageHash[:],
			InvoiceAmount:  amount,
			Description:    "Boltz swap out",
			PairHash:       pairInfo.Hash,
			ReferralId:     "alby",
			ExtraFees:      albyFee,
		})

		if err != nil {
			return fmt.Errorf("could not create swap: %s", err)
		}

		boltzPubKey, err := btcec.ParsePubKey(swap.RefundPublicKey)
		if err != nil {
			return err
		}

		tree = swap.SwapTree.Deserialize()
		if err := tree.Init(boltz.CurrencyBtc, true, ourKeys, boltzPubKey); err != nil {
			return err
		}

		if err := tree.Check(boltz.ReverseSwap, swap.TimeoutBlockHeight, preimageHash[:]); err != nil {
			return err
		}

		swapTreeJson, err := json.Marshal(swap.SwapTree)
		if err != nil {
			return err
		}

		err = tx.Model(&dbSwap).Updates(&db.Swap{
			SwapId:      swap.Id,
			BoltzPubkey: hex.EncodeToString(swap.RefundPublicKey),
			SwapTree:    datatypes.JSON(swapTreeJson),
		}).Error
		if err != nil {
			return err
		}

		// commit transaction
		return nil
	})

	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"dbSwapId":    dbSwap.ID,
			"swapId":      swap.Id,
			"paymentHash": paymentHash,
		}).Error("Failed to save swap")
		return nil, err
	}

	logger.Logger.WithField("swapId", swap.Id).Info("Swap created")

	boltzWs := svc.boltzApi.NewWebsocket()
	if err := boltzWs.Connect(); err != nil {
		svc.markSwapState(&dbSwap, constants.SWAP_STATE_FAILED)
		return nil, fmt.Errorf("could not connect to Boltz websocket: %w", err)
	}

	if err := boltzWs.Subscribe([]string{swap.Id}); err != nil {
		_ = boltzWs.Close()
		svc.markSwapState(&dbSwap, constants.SWAP_STATE_FAILED)
		return nil, err
	}

	go func() {
		var swapState string
		claimTicker := time.NewTicker(10 * time.Second)

		defer func() {
			if swapState == "" {
				swapState = constants.SWAP_STATE_FAILED
			}
			svc.markSwapState(&dbSwap, swapState)
			claimTicker.Stop()
			if err := boltzWs.Close(); err != nil {
				logger.Logger.WithError(err).WithFields(logrus.Fields{
					"swapId": swap.Id,
				}).Error("Failed to close boltz websocket")
			}
		}()

		paymentErrorCh := make(chan error, 1)

		for {
			updatesCh := boltzWs.Updates
			for {
				select {
				case <-svc.ctx.Done():
					logger.Logger.WithError(svc.ctx.Err()).WithFields(logrus.Fields{
						"swapId": swap.Id,
					}).Error("Swap out context cancelled")
					return
				case err := <-paymentErrorCh:
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"swapId": swap.Id,
					}).Error("Failed to pay hold invoice, terminating swap out...")
					return
				case <-claimTicker.C:
					if dbSwap.ClaimTxId != "" {
						tx, err := svc.getMempoolTx(dbSwap.ClaimTxId)
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId":    dbSwap.SwapId,
								"claimTxId": dbSwap.ClaimTxId,
							}).Debug("Claim poll failed; will retry")
							break
						}
						if tx.Status.Confirmed {
							swapState = constants.SWAP_STATE_SUCCESS
							logger.Logger.WithField("swapId", dbSwap.SwapId).Info("Swap succeeded")
							svc.eventPublisher.Publish(&events.Event{
								Event: "nwc_swap_succeeded",
								Properties: map[string]interface{}{
									"swapType":    constants.SWAP_TYPE_OUT,
									"swapId":      dbSwap.SwapId,
									"amount":      dbSwap.ReceivedAmount,
									"destination": dbSwap.DestinationAddress,
								},
							})
							return
						}
					}
				case update, ok := <-updatesCh:
					if !ok {
						logger.Logger.WithField("swapId", swap.Id).Error("Boltz websocket closed unexpectedly, reconnecting...")
						if err := boltzWs.Connect(); err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Reconnection to Boltz websocket failed")
							return
						}
						if err := boltzWs.Subscribe([]string{swap.Id}); err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Resubscribe after reconnect failed")
							return
						}
						break
					}

					parsedStatus := boltz.ParseEvent(update.Status)

					switch parsedStatus {
					case boltz.SwapCreated:
						logger.Logger.WithField("swapId", swap.Id).Info("Paying the swap invoice")
						go func() {
							metadata := map[string]interface{}{
								"swap_id": swap.Id,
							}
							sendPaymentTimeout := int64(3600)
							holdInvoicePayment, err := svc.transactionsService.SendPaymentSync(svc.ctx, swap.Invoice, nil, metadata, svc.lnClient, nil, nil, &sendPaymentTimeout)
							if err != nil {
								logger.Logger.WithError(err).WithFields(logrus.Fields{
									"swapId": swap.Id,
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
							"swapId":     swap.Id,
							"lockupTxId": update.Transaction.Id,
						}).Info("Lockup transaction found in mempool")
						err = svc.db.Model(&dbSwap).Update("lockup_tx_id", update.Transaction.Id).Error
						if err != nil {
							logger.Logger.WithFields(logrus.Fields{
								"swapId":     swap.Id,
								"lockupTxId": update.Transaction.Id,
							}).WithError(err).Error("Failed to save lockup txid to swap")
						}
					case boltz.TransactionConfirmed:
						logger.Logger.WithFields(logrus.Fields{
							"swapId":     swap.Id,
							"lockupTxId": dbSwap.LockupTxId,
						}).Info("Lockup transaction confirmed in mempool")
						lockupTransaction, err := boltz.NewTxFromHex(boltz.CurrencyBtc, update.Transaction.Hex, nil)
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Failed to build lockup tx from hex")
							return
						}

						vout, _, err := lockupTransaction.FindVout(network, swap.LockupAddress)
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Failed to find lockup address output")
							return
						}

						feeRates, err := svc.getFeeRates()
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Failed to fetch fee rate to create claim transaction")
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
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Could not create claim transaction")
							return
						}

						vout, _, _ = claimTransaction.FindVout(network, destination)
						claimAmount, _ := claimTransaction.VoutValue(vout)

						txHex, err := claimTransaction.Serialize()
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Could not serialize claim transaction")
							return
						}

						// TODO: Replace with LNClient broadcast method to avoid trusting boltz
						claimTxId, err := svc.boltzApi.BroadcastTransaction(boltz.CurrencyBtc, txHex)
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Could not broadcast transaction")
							return
						}

						logger.Logger.WithFields(logrus.Fields{
							"swapId":    swap.Id,
							"claimTxId": claimTxId,
						}).Info("Claim transaction broadcasted")

						err = svc.db.Model(&dbSwap).Updates(&db.Swap{
							ClaimTxId:      claimTxId,
							ReceivedAmount: claimAmount,
						}).Error
						if err != nil {
							logger.Logger.WithFields(logrus.Fields{
								"swapId":      swap.Id,
								"claimTxId":   claimTxId,
								"claimAmount": claimAmount,
							}).WithError(err).Error("Failed to save claim info to swap")
							return
						}
					case boltz.TransactionFailed, boltz.SwapExpired:
						logger.Logger.WithFields(logrus.Fields{
							"swapId": swap.Id,
							"reason": update.Status,
						}).Info("Swap out failed, HTLC is cancelled")
						return
					}
				}
			}
		}
	}()

	return &SwapResponse{
		SwapId:      swap.Id,
		PaymentHash: paymentHash,
	}, nil
}

func (svc *swapsService) SwapIn(amount uint64, autoSwap bool) (*SwapResponse, error) {
	amountMSat := amount * 1000
	invoice, err := svc.transactionsService.MakeInvoice(svc.ctx, amountMSat, "Boltz swap in", "", 0, nil, svc.lnClient, nil, nil)
	if err != nil {
		return nil, err
	}

	decodedPreimageHash, err := hex.DecodeString(invoice.PaymentHash)
	if err != nil {
		return nil, fmt.Errorf("invalid preimage hash: %v", err)
	}

	network, err := boltz.ParseChain(svc.cfg.GetNetwork())
	if err != nil {
		return nil, err
	}

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

	dbSwap := db.Swap{
		Type:        constants.SWAP_TYPE_IN,
		State:       constants.SWAP_STATE_PENDING,
		PaymentHash: invoice.PaymentHash,
		AutoSwap:    autoSwap,
	}

	var tree *boltz.SwapTree
	var ourKeys *btcec.PrivateKey
	var swap *boltz.CreateSwapResponse

	err = svc.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&dbSwap).Error
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				err := tx.Model(&dbSwap).Update("state", constants.SWAP_STATE_FAILED).Error
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"dbSwapID":    dbSwap.ID,
						"paymentHash": invoice.PaymentHash,
					}).Error("Failed to mark swap as failed")
				}
			}
		}()

		ourKeys, err = svc.keys.GetSwapKey(dbSwap.ID)
		if err != nil {
			return fmt.Errorf("error generating swap child private key: %w", err)
		}

		swap, err = svc.boltzApi.CreateSwap(boltz.CreateSwapRequest{
			From:            boltz.CurrencyBtc,
			To:              boltz.CurrencyBtc,
			RefundPublicKey: ourKeys.PubKey().SerializeCompressed(),
			Invoice:         invoice.PaymentRequest,
			PairHash:        pairInfo.Hash,
			ReferralId:      "alby",
			ExtraFees:       albyFee,
		})
		if err != nil {
			return fmt.Errorf("could not create swap: %s", err)
		}

		boltzPubKey, err := btcec.ParsePubKey(swap.ClaimPublicKey)
		if err != nil {
			return err
		}

		tree = swap.SwapTree.Deserialize()
		if err := tree.Init(boltz.CurrencyBtc, false, ourKeys, boltzPubKey); err != nil {
			return err
		}

		if err := tree.Check(boltz.NormalSwap, swap.TimeoutBlockHeight, decodedPreimageHash); err != nil {
			return err
		}

		if err := tree.CheckAddress(swap.Address, network, nil); err != nil {
			return err
		}

		swapTreeJson, err := json.Marshal(swap.SwapTree)
		if err != nil {
			return err
		}

		err = tx.Model(&dbSwap).Updates(&db.Swap{
			SwapId:             swap.Id,
			SendAmount:         swap.ExpectedAmount,
			DestinationAddress: swap.Address,
			BoltzPubkey:        hex.EncodeToString(swap.ClaimPublicKey),
			SwapTree:           datatypes.JSON(swapTreeJson),
		}).Error
		if err != nil {
			return err
		}

		// commit transaction
		return nil
	})

	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"paymentHash": invoice.PaymentHash,
		}).Error("Failed to save swap")
		return nil, err
	}

	metadata := map[string]interface{}{
		"swap_id": swap.Id,
	}
	err = svc.transactionsService.SetTransactionMetadata(svc.ctx, invoice.ID, metadata)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId":      swap.Id,
			"paymentHash": invoice.PaymentHash,
			"metadata":    metadata,
		}).Error("Failed to add swap metadata to lightning payment")
	}

	logger.Logger.WithField("swapId", swap.Id).Info("Swap created")

	boltzWs := svc.boltzApi.NewWebsocket()
	if err := boltzWs.Connect(); err != nil {
		svc.markSwapState(&dbSwap, constants.SWAP_STATE_FAILED)
		return nil, fmt.Errorf("could not connect to Boltz websocket: %w", err)
	}

	if err := boltzWs.Subscribe([]string{swap.Id}); err != nil {
		_ = boltzWs.Close()
		svc.markSwapState(&dbSwap, constants.SWAP_STATE_FAILED)
		return nil, err
	}

	go func() {
		var swapState string
		defer func() {
			if swapState == "" {
				swapState = constants.SWAP_STATE_FAILED
			}
			svc.markSwapState(&dbSwap, swapState)
			if err := boltzWs.Close(); err != nil {
				logger.Logger.WithError(err).WithFields(logrus.Fields{
					"swapId": swap.Id,
				}).Error("Failed to close boltz websocket")
			}
		}()

		for {
			updatesCh := boltzWs.Updates

			for {
				select {
				case <-svc.ctx.Done():
					logger.Logger.WithError(svc.ctx.Err()).WithFields(logrus.Fields{
						"swapId": swap.Id,
					}).Error("Swap in context cancelled")
					return
				case update, ok := <-updatesCh:
					if !ok {
						logger.Logger.WithField("swapId", swap.Id).Error("Boltz websocket closed unexpectedly, reconnecting...")
						if err := boltzWs.Connect(); err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Reconnection to Boltz websocket failed")
							return
						}
						if err := boltzWs.Subscribe([]string{swap.Id}); err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Resubscribe after reconnect failed")
							return
						}
						break
					}

					parsedStatus := boltz.ParseEvent(update.Status)

					switch parsedStatus {
					case boltz.TransactionMempool:
						logger.Logger.WithFields(logrus.Fields{
							"swapId":     swap.Id,
							"lockupTxId": update.Transaction.Id,
						}).Info("Lockup transaction found in mempool")
						err = svc.db.Model(&dbSwap).Update("lockup_tx_id", update.Transaction.Id).Error
						if err != nil {
							logger.Logger.WithFields(logrus.Fields{
								"swapId":     swap.Id,
								"lockupTxId": update.Transaction.Id,
							}).WithError(err).Error("Failed to save lockup txid to swap")
						}
					case boltz.TransactionConfirmed:
						logger.Logger.WithFields(logrus.Fields{
							"swapId":     swap.Id,
							"lockupTxId": dbSwap.LockupTxId,
						}).Info("Lockup transaction confirmed in mempool")
					case boltz.TransactionClaimPending:
						// this is not a mandatory step as boltz can still claim the locked up funds via the script path
						logger.Logger.WithFields(logrus.Fields{
							"swapId":      swap.Id,
							"transaction": update.Transaction,
						}).Info("Sending partial signature to boltz to claim the payment")
						claimDetails, err := svc.boltzApi.GetSwapClaimDetails(swap.Id)
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Could not get claim details from Boltz")
							return
						}

						preimageHash := sha256.Sum256(claimDetails.Preimage)
						if !bytes.Equal(decodedPreimageHash, preimageHash[:]) {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId":   swap.Id,
								"preimage": claimDetails.Preimage,
							}).Error("Boltz returned wrong preimage")
							return
						}

						session, _ := boltz.NewSigningSession(tree)
						partial, err := session.Sign(claimDetails.TransactionHash, claimDetails.PubNonce)
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Could not create partial signature")
							return
						}

						if err := svc.boltzApi.SendSwapClaimSignature(swap.Id, partial); err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Could not send partial signature to Boltz")
							return
						}
					case boltz.TransactionClaimed:
						swapState = constants.SWAP_STATE_SUCCESS
						err = svc.db.Model(&dbSwap).Updates(&db.Swap{
							ReceivedAmount: amount,
						}).Error
						if err != nil {
							logger.Logger.WithFields(logrus.Fields{
								"swapId":         swap.Id,
								"receivedAmount": amount,
							}).WithError(err).Error("Failed to save received amount to swap")
							return
						}
						logger.Logger.WithField("swapId", swap.Id).Info("Swap succeeded")
						svc.eventPublisher.Publish(&events.Event{
							Event: "nwc_swap_succeeded",
							Properties: map[string]interface{}{
								"swapType": constants.SWAP_TYPE_IN,
								"swapId":   swap.Id,
								"amount":   amount,
							},
						})
						return
					case boltz.TransactionLockupFailed, boltz.InvoiceFailedToPay, boltz.SwapExpired:
						logger.Logger.WithFields(logrus.Fields{
							"swapId": swap.Id,
							"reason": update.Status,
						}).Error("Swap in failed, initiating refund")

						err := svc.RefundSwap(swap.Id, "")
						if err != nil {
							logger.Logger.WithError(err).WithFields(logrus.Fields{
								"swapId": swap.Id,
							}).Error("Could not process refund")
						} else {
							swapState = constants.SWAP_STATE_REFUNDED
						}
						return
					}
				}
			}
		}
	}()

	return &SwapResponse{
		SwapId:      swap.Id,
		PaymentHash: invoice.PaymentHash,
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

func (svc *swapsService) markSwapState(dbSwap *db.Swap, state string) {
	if svc.db.Limit(1).Find(dbSwap, &db.Swap{
		SwapId: dbSwap.SwapId,
		State:  state,
	}).RowsAffected > 0 {
		logger.Logger.WithField("swapId", dbSwap.SwapId).Debugf("swap already marked as %s", state)
		return
	}

	dbErr := svc.db.Model(dbSwap).Update("state", state).Error
	if dbErr != nil {
		logger.Logger.WithError(dbErr).WithField("swapId", dbSwap.SwapId).Error("Failed to update swap state")
	}
}

func (svc *swapsService) RefundSwap(swapId, address string) error {
	var swap db.Swap
	err := svc.db.Limit(1).Find(&swap, &db.Swap{
		SwapId: swapId,
	}).Error
	if err != nil {
		logger.Logger.WithField("swapId", swapId).WithError(err).Error("Could not find swap to process refund")
		return err
	}

	if swap.Type != constants.SWAP_TYPE_IN {
		return fmt.Errorf("cannot process refund for swap id: %s", swapId)
	}

	if swap.ClaimTxId != "" {
		return fmt.Errorf("refund already processed with claim txid: %s", swap.ClaimTxId)
	}

	network, err := boltz.ParseChain(svc.cfg.GetNetwork())
	if err != nil {
		return err
	}

	swapTransactionResp, err := svc.boltzApi.GetSwapTransaction(swapId)
	if err != nil {
		logger.Logger.WithField("swapId", swapId).WithError(err).Error("Failed to get lockup tx from swap id")
		return err
	}

	if swap.LockupTxId == "" {
		err = svc.db.Model(&swap).Update("lockup_tx_id", swapTransactionResp.Id).Error
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"swapId":     swapId,
				"lockupTxId": swapTransactionResp.Id,
			}).WithError(err).Error("Failed to save lockup txid to swap")
			return err
		}
	}

	ourKeys, err := svc.keys.GetSwapKey(swap.ID)
	if err != nil {
		return fmt.Errorf("error generating swap child private key: %w", err)
	}

	var serializedTree boltz.SerializedTree
	if err := json.Unmarshal(swap.SwapTree, &serializedTree); err != nil {
		return err
	}

	boltzPubkeyBytes, err := hex.DecodeString(swap.BoltzPubkey)
	if err != nil {
		return fmt.Errorf("invalid boltz pubkey: %v", err)
	}

	boltzPubKey, err := btcec.ParsePubKey(boltzPubkeyBytes)
	if err != nil {
		return err
	}

	decodedPreimageHash, err := hex.DecodeString(swap.PaymentHash)
	if err != nil {
		return fmt.Errorf("invalid preimage hash: %v", err)
	}

	tree := serializedTree.Deserialize()
	if err := tree.Init(boltz.CurrencyBtc, false, ourKeys, boltzPubKey); err != nil {
		return err
	}

	if err := tree.Check(boltz.NormalSwap, swapTransactionResp.TimeoutBlockHeight, decodedPreimageHash); err != nil {
		return err
	}

	if err := tree.CheckAddress(swap.DestinationAddress, network, nil); err != nil {
		return err
	}

	lockupTransaction, err := boltz.NewTxFromHex(boltz.CurrencyBtc, swapTransactionResp.Hex, nil)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to build lockup tx from hex")
		return err
	}
	vout, _, err := lockupTransaction.FindVout(network, swap.DestinationAddress)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to find lockup address output")
		return err
	}
	feeRates, err := svc.getFeeRates()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to fetch fee rate to create claim transaction")
		return err
	}

	if address == "" {
		address, err = svc.cfg.Get(config.OnchainAddressKey, "")
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to get on-chain address from config")
			return err
		}
	}

	refundTransaction, _, err := boltz.ConstructTransaction(
		network,
		boltz.CurrencyBtc,
		[]boltz.OutputDetails{
			{
				SwapId:             swapId,
				SwapType:           boltz.NormalSwap,
				Address:            address,
				LockupTransaction:  lockupTransaction,
				TimeoutBlockHeight: swapTransactionResp.TimeoutBlockHeight,
				Vout:               vout,
				PrivateKey:         ourKeys,
				SwapTree:           tree,
				Cooperative:        true,
			},
		},
		float64(feeRates.FastestFee),
		svc.boltzApi,
	)
	if err != nil {
		logger.Logger.WithError(err).Error("Could not create claim transaction")
		return err
	}

	vout, _, _ = refundTransaction.FindVout(network, address)
	refundAmount, _ := refundTransaction.VoutValue(vout)

	txHex, err := refundTransaction.Serialize()
	if err != nil {
		logger.Logger.WithError(err).Error("Could not serialize refund transaction")
		return err
	}

	// TODO: Replace with LNClient broadcast method to avoid trusting boltz
	claimTxId, err := svc.boltzApi.BroadcastTransaction(boltz.CurrencyBtc, txHex)
	if err != nil {
		logger.Logger.WithError(err).Error("Could not broadcast transaction")
		return err
	}

	logger.Logger.WithFields(logrus.Fields{
		"swapId":    swapId,
		"claimTxId": claimTxId,
	}).Info("Claim transaction broadcasted for refund")

	err = svc.db.Model(&swap).Updates(&db.Swap{
		ClaimTxId:      claimTxId,
		ReceivedAmount: refundAmount,
		State:          constants.SWAP_STATE_REFUNDED,
	}).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"swapId":    swapId,
			"claimTxId": claimTxId,
		}).WithError(err).Error("Failed to save claim txid to swap")
		return err
	}

	return nil
}

func (svc *swapsService) GetSwap(swapId string) (*Swap, error) {
	var swap db.Swap
	err := svc.db.Limit(1).Find(&swap, &db.Swap{
		SwapId: swapId,
	}).Error

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get swap")
		return nil, err
	}

	return &swap, nil
}

func (svc *swapsService) ListSwaps() ([]Swap, error) {
	var swaps []db.Swap
	err := svc.db.Find(&swaps).Error

	if err != nil {
		logger.Logger.WithError(err).Error("Failed to list swaps")
		return nil, err
	}

	return swaps, nil
}

func (svc *swapsService) getMempoolTx(txId string) (*MempoolTx, error) {
	var transaction MempoolTx
	endpoint := fmt.Sprintf("/tx/%s", txId)
	if err := svc.requestMempoolApi(endpoint, &transaction); err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (svc *swapsService) getFeeRates() (*FeeRates, error) {
	var rates FeeRates
	if err := svc.requestMempoolApi("/v1/fees/recommended", &rates); err != nil {
		return nil, err
	}
	return &rates, nil
}

func (svc *swapsService) requestMempoolApi(endpoint string, result interface{}) error {
	url := svc.cfg.GetEnv().MempoolApi + endpoint
	url = strings.ReplaceAll(url, "testnet/", "")

	client := http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to create http request")
		return err
	}

	res, err := client.Do(req)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to send request")
		return err
	}

	defer res.Body.Close()

	body, readErr := io.ReadAll(res.Body)
	if readErr != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to read response body")
		return errors.New("failed to read response body")
	}

	jsonErr := json.Unmarshal(body, &result)
	if jsonErr != nil {
		logger.Logger.WithError(jsonErr).WithFields(logrus.Fields{
			"url": url,
		}).Error("Failed to deserialize json")
		return fmt.Errorf("failed to deserialize json %s %s", url, string(body))
	}
	return nil
}
