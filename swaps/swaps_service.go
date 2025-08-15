package swaps

import (
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
	"sync"
	"time"

	"github.com/BoltzExchange/boltz-client/v2/pkg/boltz"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service/keys"
	"github.com/getAlby/hub/transactions"
	decodepay "github.com/nbd-wtf/ln-decodepay"
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
	boltzWs             *boltz.Websocket
	swapListeners       map[string]chan boltz.SwapUpdate
	swapListenersLock   sync.Mutex
}

type SwapsService interface {
	StopAutoSwapOut()
	EnableAutoSwapOut() error
	SwapOut(amount uint64, destination string, autoSwap bool, usedXpubDerivation bool) (*SwapResponse, error)
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

func NewSwapsService(ctx context.Context, db *gorm.DB, cfg config.Config, keys keys.Keys, eventPublisher events.EventPublisher,
	lnClient lnclient.LNClient, transactionsService transactions.TransactionsService) SwapsService {
	boltzApi := &boltz.Api{URL: cfg.GetEnv().BoltzApi}
	boltzWs := boltzApi.NewWebsocket()

	for {
		err := boltzWs.Connect()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to connect to boltz websocket, retrying in 2s...")
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	logger.Logger.Info("Connected to boltz websocket")

	svc := &swapsService{
		ctx:                 ctx,
		cfg:                 cfg,
		db:                  db,
		keys:                keys,
		eventPublisher:      eventPublisher,
		transactionsService: transactionsService,
		lnClient:            lnClient,
		boltzApi:            boltzApi,
		boltzWs:             boltzWs,
		swapListeners:       make(map[string]chan boltz.SwapUpdate),
	}

	go func() {
		for {
			update, ok := <-svc.boltzWs.Updates
			if !ok {
				logger.Logger.Error("Received error from boltz websocket")
				continue
			}

			svc.swapListenersLock.Lock()
			ch, ok := svc.swapListeners[update.Id]
			svc.swapListenersLock.Unlock()
			if ok {
				ch <- update
			} else {
				logger.Logger.WithField("swap_id", update.Id).Error("Failed to receive update from boltz")
			}
		}
	}()

	err := svc.EnableAutoSwapOut()
	if err != nil {
		logger.Logger.WithError(err).Error("Couldn't enable auto swaps")
	}

	go svc.subscribePendingSwaps()

	return svc
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
		for {
			select {
			case <-time.After(1 * time.Hour):
				logger.Logger.Debug("Checking to see if we can swap")
				balance, err := svc.lnClient.GetBalances(ctx, false)
				if err != nil {
					logger.Logger.WithError(err).Error("Failed to get balance")
					continue
				}
				lightningBalance := uint64(balance.Lightning.TotalSpendable)
				balanceThresholdMilliSats := balanceThreshold * 1000
				if lightningBalance < balanceThresholdMilliSats {
					logger.Logger.Info("Threshold requirements not met for swap, ignoring")
					continue
				}

				actualDestination := swapDestination
				var usedXpubDerivation bool
				if swapDestination != "" {
					if err := svc.validateXpub(swapDestination); err == nil {
						actualDestination, err = svc.getNextUnusedAddressFromXpub()
						if err != nil {
							logger.Logger.WithError(err).Error("Failed to get next address from xpub")
							continue
						}
						usedXpubDerivation = true
					}
				}

				logger.Logger.WithFields(logrus.Fields{
					"amount":      amount,
					"destination": actualDestination,
				}).Info("Initiating swap")
				_, err = svc.SwapOut(amount, actualDestination, true, usedXpubDerivation)
				if err != nil {
					logger.Logger.WithError(err).Error("Failed to initiate swap")
					continue
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

func (svc *swapsService) SwapOut(amount uint64, destination string, autoSwap bool, usedXpubDerivation bool) (*SwapResponse, error) {
	if destination == "" {
		var err error
		destination, err = svc.lnClient.GetNewOnchainAddress(svc.ctx)
		if err != nil {
			return nil, fmt.Errorf("could not get onchain address from config: %s", err)
		}
	}

	preimage := make([]byte, 32)
	_, err := rand.Read(preimage)
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
		Preimage:           hex.EncodeToString(preimage),
		AutoSwap:           autoSwap,
		UsedXpub:           usedXpubDerivation,
	}

	var ourKeys *btcec.PrivateKey
	var swap *boltz.CreateReverseSwapResponse

	defer func() {
		if err != nil && dbSwap.ID != 0 {
			svc.markSwapState(&dbSwap, constants.SWAP_STATE_FAILED)
		}
	}()

	err = svc.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&dbSwap).Error
		if err != nil {
			return err
		}

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
			Description:    "Lightning to on-chain swap",
			PairHash:       pairInfo.Hash,
			ReferralId:     "alby",
			ExtraFees:      albyFee,
		})

		if err != nil {
			return fmt.Errorf("could not create swap: %s", err)
		}

		swapTreeJson, err := json.Marshal(swap.SwapTree)
		if err != nil {
			return err
		}

		err = tx.Model(&dbSwap).Updates(&db.Swap{
			SwapId:             swap.Id,
			Invoice:            swap.Invoice,
			LockupAddress:      swap.LockupAddress,
			TimeoutBlockHeight: swap.TimeoutBlockHeight,
			BoltzPubkey:        hex.EncodeToString(swap.RefundPublicKey),
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
			"paymentHash": paymentHash,
		}).Error("Failed to save swap")
		return nil, err
	}

	logger.Logger.WithField("swapId", swap.Id).Info("Swap created")

	if autoSwap {
		// block until the swap finishes to ensure we can't do multiple concurrent auto swaps
		svc.startSwapOutListener(&dbSwap)
	} else {
		// run in parallel as we need to return the swap ID in the HTTP response
		go svc.startSwapOutListener(&dbSwap)
	}

	return &SwapResponse{
		SwapId:      swap.Id,
		PaymentHash: paymentHash,
	}, nil
}

func (svc *swapsService) SwapIn(amount uint64, autoSwap bool) (*SwapResponse, error) {
	amountMSat := amount * 1000
	invoice, err := svc.transactionsService.MakeInvoice(svc.ctx, amountMSat, "On-chain to lightning swap", "", 0, nil, svc.lnClient, nil, nil)
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
		Invoice:     invoice.PaymentRequest,
		PaymentHash: invoice.PaymentHash,
		AutoSwap:    autoSwap,
	}

	var ourKeys *btcec.PrivateKey
	var swap *boltz.CreateSwapResponse

	defer func() {
		if err != nil && dbSwap.ID != 0 {
			svc.markSwapState(&dbSwap, constants.SWAP_STATE_FAILED)
		}
	}()

	err = svc.db.Transaction(func(tx *gorm.DB) error {
		err := tx.Save(&dbSwap).Error
		if err != nil {
			return err
		}

		defer func() {
			if err != nil {
				svc.markSwapState(&dbSwap, constants.SWAP_STATE_FAILED)
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

		swapTreeJson, err := json.Marshal(swap.SwapTree)
		if err != nil {
			return err
		}

		err = tx.Model(&dbSwap).Updates(&db.Swap{
			SwapId:             swap.Id,
			SendAmount:         swap.ExpectedAmount,
			LockupAddress:      swap.Address,
			TimeoutBlockHeight: swap.TimeoutBlockHeight,
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
		return nil, err
	}

	logger.Logger.WithField("swapId", swap.Id).Info("Swap created")

	go svc.startSwapInListener(&dbSwap)

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

	dbErr := svc.db.Model(dbSwap).Updates(&db.Swap{
		State: state,
	}).Error
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
		return errors.New("only On-chain -> Lightning swaps can be refunded")
	}

	if swap.ClaimTxId != "" {
		return fmt.Errorf("refund already processed with claim txid: %s", swap.ClaimTxId)
	}

	network, err := boltz.ParseChain(svc.cfg.GetNetwork())
	if err != nil {
		return err
	}

	// Fetch raw hex to construct the lockup transaction
	swapTransactionResp, err := svc.boltzApi.GetSwapTransaction(swapId)
	if err != nil {
		logger.Logger.WithField("swapId", swapId).WithError(err).Error("Failed to get lockup tx from swap id")
		return err
	}

	if swap.LockupTxId == "" {
		err = svc.db.Model(&swap).Updates(&db.Swap{
			LockupTxId: swapTransactionResp.Id,
		}).Error
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

	if err := tree.Check(boltz.NormalSwap, swap.TimeoutBlockHeight, decodedPreimageHash); err != nil {
		return err
	}

	if err := tree.CheckAddress(swap.LockupAddress, network, nil); err != nil {
		return err
	}

	lockupTransaction, err := boltz.NewTxFromHex(boltz.CurrencyBtc, swapTransactionResp.Hex, nil)
	if err != nil {
		logger.Logger.WithField("swapId", swapId).WithError(err).Error("Failed to build lockup tx from hex")
		return err
	}
	vout, _, err := lockupTransaction.FindVout(network, swap.LockupAddress)
	if err != nil {
		logger.Logger.WithField("swapId", swapId).WithError(err).Error("Failed to find lockup address output")
		return err
	}
	feeRates, err := svc.getFeeRates()
	if err != nil {
		logger.Logger.WithField("swapId", swapId).WithError(err).Error("Failed to fetch fee rate to create claim transaction")
		return err
	}

	if address == "" {
		address, err = svc.lnClient.GetNewOnchainAddress(svc.ctx)
		if err != nil {
			logger.Logger.WithField("swapId", swapId).WithError(err).Error("Failed to get new on-chain address from config")
			return err
		}
	}

	err = svc.db.Model(&swap).Updates(&db.Swap{
		RefundAddress: address,
	}).Error
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"swapId":        swapId,
			"refundAddress": address,
		}).WithError(err).Error("Failed to save refund address to swap")
		return err
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
		logger.Logger.WithField("swapId", swapId).WithError(err).Error("Could not create claim transaction")
		return err
	}

	vout, _, _ = refundTransaction.FindVout(network, address)
	refundAmount, _ := refundTransaction.VoutValue(vout)

	txHex, err := refundTransaction.Serialize()
	if err != nil {
		logger.Logger.WithField("swapId", swapId).WithError(err).Error("Could not serialize refund transaction")
		return err
	}

	// TODO: Replace with LNClient broadcast method to avoid trusting boltz
	claimTxId, err := svc.boltzApi.BroadcastTransaction(boltz.CurrencyBtc, txHex)
	if err != nil {
		logger.Logger.WithField("swapId", swapId).WithError(err).Error("Could not broadcast transaction")
		return err
	}

	logger.Logger.WithFields(logrus.Fields{
		"swapId":    swapId,
		"claimTxId": claimTxId,
	}).Info("Claim transaction broadcasted for refund")

	err = svc.db.Model(&swap).Updates(&db.Swap{
		ClaimTxId:     claimTxId,
		ReceiveAmount: refundAmount,
		State:         constants.SWAP_STATE_REFUNDED,
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

func (svc *swapsService) subscribePendingSwaps() {
	var swaps []db.Swap
	if err := svc.db.Where("state = ?", constants.SWAP_STATE_PENDING).Find(&swaps).Error; err != nil {
		logger.Logger.WithError(err).Error("failed to load pending swaps")
		return
	}
	if len(swaps) == 0 {
		return
	}

	logger.Logger.WithField("count", len(swaps)).Info("Resuming pending swaps...")

	ids := make([]string, len(swaps))
	for i, s := range swaps {
		ids[i] = s.SwapId
	}

	for _, swap := range swaps {
		switch swap.Type {
		case constants.SWAP_TYPE_IN:
			go svc.startSwapInListener(&swap)
		case constants.SWAP_TYPE_OUT:
			go svc.startSwapOutListener(&swap)
		}
	}
}

func (svc *swapsService) startSwapInListener(swap *db.Swap) {
	for {
		err := svc.boltzWs.Subscribe([]string{swap.SwapId})
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to subscribe to boltz websocket, retrying in 2s...")
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	logger.Logger.WithField("swapId", swap.SwapId).Info("Subscribed to boltz websocket")

	updateCh := make(chan boltz.SwapUpdate)
	svc.swapListenersLock.Lock()
	svc.swapListeners[swap.SwapId] = updateCh
	svc.swapListenersLock.Unlock()

	var err error
	defer func() {
		svc.swapListenersLock.Lock()
		delete(svc.swapListeners, swap.SwapId)
		svc.swapListenersLock.Unlock()
		svc.boltzWs.Unsubscribe(swap.SwapId)
		if err != nil {
			svc.markSwapState(swap, constants.SWAP_STATE_FAILED)
		}
	}()

	var network *boltz.Network
	network, err = boltz.ParseChain(svc.cfg.GetNetwork())
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to parse network")
		return
	}

	var ourKeys *btcec.PrivateKey
	ourKeys, err = svc.keys.GetSwapKey(swap.ID)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to generate swap child private key")
		return
	}

	var serializedTree boltz.SerializedTree
	if err = json.Unmarshal(swap.SwapTree, &serializedTree); err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to unmarshal swap tree")
		return
	}

	boltzPubkeyBytes, _ := hex.DecodeString(swap.BoltzPubkey)

	var boltzPubKey *btcec.PublicKey
	boltzPubKey, err = btcec.ParsePubKey(boltzPubkeyBytes)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to parse boltz pubkey")
		return
	}

	decodedPreimageHash, _ := hex.DecodeString(swap.PaymentHash)

	tree := serializedTree.Deserialize()
	if err = tree.Init(boltz.CurrencyBtc, false, ourKeys, boltzPubKey); err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to initialize swap tree")
		return
	}

	if err = tree.Check(boltz.NormalSwap, swap.TimeoutBlockHeight, decodedPreimageHash); err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to check swap tree")
		return
	}

	if err = tree.CheckAddress(swap.LockupAddress, network, nil); err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to check address")
		return
	}

	paymentRequest, _ := decodepay.Decodepay(swap.Invoice)
	amount := uint64(paymentRequest.MSatoshi / 1000)

	for {
		select {
		case <-svc.ctx.Done():
			logger.Logger.WithError(svc.ctx.Err()).WithFields(logrus.Fields{
				"swapId": swap.SwapId,
			}).Error("Swap in context cancelled")
			return
		case update, ok := <-updateCh:
			if !ok {
				logger.Logger.WithField("swap_id", update.Id).Error("Failed to receive update from boltz")
				continue
			}
			if update.Id != swap.SwapId {
				continue
			}
			switch boltz.ParseEvent(update.Status) {
			case boltz.TransactionMempool:
				logger.Logger.WithFields(logrus.Fields{
					"swapId":     swap.SwapId,
					"lockupTxId": update.Transaction.Id,
				}).Info("Lockup transaction found in mempool")
				err = svc.db.Model(swap).Updates(&db.Swap{
					LockupTxId: update.Transaction.Id,
				}).Error
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"swapId":     swap.SwapId,
						"lockupTxId": update.Transaction.Id,
					}).WithError(err).Error("Failed to save lockup txid to swap")
					return
				}
			case boltz.TransactionConfirmed:
				logger.Logger.WithFields(logrus.Fields{
					"swapId":     swap.SwapId,
					"lockupTxId": swap.LockupTxId,
				}).Info("Lockup transaction confirmed in mempool")
			case boltz.TransactionClaimed:
				svc.markSwapState(swap, constants.SWAP_STATE_SUCCESS)
				err = svc.db.Model(swap).Updates(&db.Swap{
					ReceiveAmount: amount,
				}).Error
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"swapId":        swap.SwapId,
						"receiveAmount": amount,
					}).WithError(err).Error("Failed to save received amount to swap")
					return
				}
				logger.Logger.WithField("swapId", swap.SwapId).Info("Swap succeeded")
				svc.eventPublisher.Publish(&events.Event{
					Event: "nwc_swap_succeeded",
					Properties: map[string]interface{}{
						"swapType": constants.SWAP_TYPE_IN,
					},
				})
				return
			case boltz.TransactionLockupFailed, boltz.InvoiceFailedToPay, boltz.SwapExpired:
				logger.Logger.WithFields(logrus.Fields{
					"swapId": swap.SwapId,
					"reason": update.Status,
				}).Error("Swap in failed, initiating refund")

				err = svc.RefundSwap(swap.SwapId, "")
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"swapId": swap.SwapId,
					}).Error("Could not process refund")
				}
				return
			}
		}
	}
}

func (svc *swapsService) startSwapOutListener(swap *db.Swap) {
	for {
		err := svc.boltzWs.Subscribe([]string{swap.SwapId})
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to subscribe to boltz websocket, retrying in 2s...")
			time.Sleep(2 * time.Second)
			continue
		}
		break
	}

	logger.Logger.WithField("swapId", swap.SwapId).Info("Subscribed to boltz websocket")

	updateCh := make(chan boltz.SwapUpdate)
	svc.swapListenersLock.Lock()
	svc.swapListeners[swap.SwapId] = updateCh
	svc.swapListenersLock.Unlock()

	var err error
	defer func() {
		svc.swapListenersLock.Lock()
		delete(svc.swapListeners, swap.SwapId)
		svc.swapListenersLock.Unlock()
		svc.boltzWs.Unsubscribe(swap.SwapId)
		if err != nil {
			svc.markSwapState(swap, constants.SWAP_STATE_FAILED)
		}
	}()

	var network *boltz.Network
	network, err = boltz.ParseChain(svc.cfg.GetNetwork())
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to parse network")
		return
	}

	var ourKeys *btcec.PrivateKey
	ourKeys, err = svc.keys.GetSwapKey(swap.ID)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to generate swap child private key")
		return
	}

	var serializedTree boltz.SerializedTree
	if err = json.Unmarshal(swap.SwapTree, &serializedTree); err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to unmarshal swap tree")
		return
	}

	boltzPubkeyBytes, _ := hex.DecodeString(swap.BoltzPubkey)

	var boltzPubKey *btcec.PublicKey
	boltzPubKey, err = btcec.ParsePubKey(boltzPubkeyBytes)
	if err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to parse boltz pubkey")
		return
	}

	preimageBytes, _ := hex.DecodeString(swap.Preimage)
	preimageHash := sha256.Sum256(preimageBytes)

	tree := serializedTree.Deserialize()
	if err = tree.Init(boltz.CurrencyBtc, true, ourKeys, boltzPubKey); err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to initialize swap tree")
		return
	}

	if err = tree.Check(boltz.ReverseSwap, swap.TimeoutBlockHeight, preimageHash[:]); err != nil {
		logger.Logger.WithError(err).WithFields(logrus.Fields{
			"swapId": swap.SwapId,
		}).Error("Failed to check swap tree")
		return
	}

	claimTicker := time.NewTicker(10 * time.Second)
	defer claimTicker.Stop()

	paymentErrorCh := make(chan error, 1)

	for {
		select {
		case <-svc.ctx.Done():
			logger.Logger.WithError(svc.ctx.Err()).WithFields(logrus.Fields{
				"swapId": swap.SwapId,
			}).Error("Swap out context cancelled")
			return
		case err = <-paymentErrorCh:
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"swapId": swap.SwapId,
			}).Error("Failed to pay hold invoice, terminating swap out...")
			return
		case <-claimTicker.C:
			if swap.ClaimTxId != "" {
				tx, err := svc.getMempoolTx(swap.ClaimTxId)
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"swapId":    swap.SwapId,
						"claimTxId": swap.ClaimTxId,
					}).Debug("Claim poll failed; will retry")
					break
				}
				if tx.Status.Confirmed {
					svc.markSwapState(swap, constants.SWAP_STATE_SUCCESS)
					logger.Logger.WithField("swapId", swap.SwapId).Info("Swap succeeded")
					if swap.UsedXpub {
						svc.bumpAutoswapXpubIndex(swap.ID)
					}
					svc.eventPublisher.Publish(&events.Event{
						Event: "nwc_swap_succeeded",
						Properties: map[string]interface{}{
							"swapType": constants.SWAP_TYPE_OUT,
						},
					})
					return
				}
			}
		case update, ok := <-updateCh:
			if !ok {
				logger.Logger.WithField("swap_id", update.Id).Error("Failed to receive update from boltz")
				continue
			}
			if update.Id != swap.SwapId {
				continue
			}
			switch boltz.ParseEvent(update.Status) {
			case boltz.SwapCreated:
				logger.Logger.WithField("swapId", swap.SwapId).Info("Paying the swap invoice")
				go func() {
					_, err := svc.transactionsService.LookupTransaction(svc.ctx, swap.PaymentHash, nil, svc.lnClient, nil)
					if err == transactions.NewNotFoundError() {
						logger.Logger.WithField("swapId", swap.SwapId).Info("Already initiated swap invoice payment")
						return
					}
					metadata := map[string]interface{}{
						"swap_id": swap.SwapId,
					}
					logger.Logger.WithField("swapId", swap.SwapId).Info("Initiating swap invoice payment")
					_, err = svc.transactionsService.SendPaymentSync(svc.ctx, swap.Invoice, nil, metadata, svc.lnClient, nil, nil)
					if err != nil {
						logger.Logger.WithError(err).WithFields(logrus.Fields{
							"swapId": swap.SwapId,
						}).Error("Error paying the swap invoice")
						paymentErrorCh <- err
						return
					}
				}()
			case boltz.TransactionMempool:
				logger.Logger.WithFields(logrus.Fields{
					"swapId":     swap.SwapId,
					"lockupTxId": update.Transaction.Id,
				}).Info("Lockup transaction found in mempool")
				err = svc.db.Model(swap).Updates(&db.Swap{
					LockupTxId: update.Transaction.Id,
				}).Error
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"swapId":     swap.SwapId,
						"lockupTxId": update.Transaction.Id,
					}).WithError(err).Error("Failed to save lockup txid to swap")
					return
				}
			case boltz.TransactionConfirmed:
				logger.Logger.WithFields(logrus.Fields{
					"swapId":     swap.SwapId,
					"lockupTxId": swap.LockupTxId,
				}).Info("Lockup transaction confirmed in mempool")

				var lockupTransaction boltz.Transaction
				lockupTransaction, err = boltz.NewTxFromHex(boltz.CurrencyBtc, update.Transaction.Hex, nil)
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"swapId": swap.SwapId,
					}).Error("Failed to build lockup tx from hex")
					return
				}

				var vout uint32
				vout, _, err = lockupTransaction.FindVout(network, swap.LockupAddress)
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"swapId": swap.SwapId,
					}).Error("Failed to find lockup address output")
					return
				}

				var feeRates *FeeRates
				feeRates, err = svc.getFeeRates()
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"swapId": swap.SwapId,
					}).Error("Failed to fetch fee rate to create claim transaction")
					return
				}

				var claimTransaction boltz.Transaction
				claimTransaction, _, err = boltz.ConstructTransaction(
					network,
					boltz.CurrencyBtc,
					[]boltz.OutputDetails{
						{
							SwapId:            swap.SwapId,
							SwapType:          boltz.ReverseSwap,
							Address:           swap.DestinationAddress,
							LockupTransaction: lockupTransaction,
							Vout:              vout,
							Preimage:          preimageBytes,
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
						"swapId": swap.SwapId,
					}).Error("Could not create claim transaction")
					return
				}

				vout, _, _ = claimTransaction.FindVout(network, swap.DestinationAddress)
				claimAmount, _ := claimTransaction.VoutValue(vout)

				var txHex string
				txHex, err = claimTransaction.Serialize()
				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"swapId": swap.SwapId,
					}).Error("Could not serialize claim transaction")
					return
				}

				var claimTxId string
				for attempt := 1; attempt <= 5; attempt++ {
					// TODO: Replace with LNClient broadcast method to avoid trusting boltz
					claimTxId, err = svc.boltzApi.BroadcastTransaction(boltz.CurrencyBtc, txHex)
					if err != nil {
						logger.Logger.WithError(err).WithFields(logrus.Fields{
							"swapId":  swap.SwapId,
							"attempt": attempt,
						}).Warn("Failed to broadcast transaction, retrying")
						time.Sleep(1 * time.Second)
						continue
					}
					break
				}

				if err != nil {
					logger.Logger.WithError(err).WithFields(logrus.Fields{
						"swapId": swap.SwapId,
					}).Error("Could not broadcast transaction")
					return
				}

				logger.Logger.WithFields(logrus.Fields{
					"swapId":    swap.SwapId,
					"claimTxId": claimTxId,
				}).Info("Claim transaction broadcasted")

				err = svc.db.Model(swap).Updates(&db.Swap{
					ClaimTxId:     claimTxId,
					ReceiveAmount: claimAmount,
				}).Error
				if err != nil {
					logger.Logger.WithFields(logrus.Fields{
						"swapId":      swap.SwapId,
						"claimTxId":   claimTxId,
						"claimAmount": claimAmount,
					}).WithError(err).Error("Failed to save claim info to swap")
					return
				}
			case boltz.TransactionFailed, boltz.SwapExpired:
				logger.Logger.WithFields(logrus.Fields{
					"swapId": swap.SwapId,
					"reason": update.Status,
				}).Error("Swap out failed, HTLC is cancelled")
				err = errors.New(update.Status)
				return
			}
		}
	}
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
	for attempt := 1; attempt <= 10; attempt++ {
		err := svc.doMempoolRequest(endpoint, result)
		if err != nil {
			logger.Logger.WithError(err).WithFields(logrus.Fields{
				"attempt":  attempt,
				"endpoint": endpoint,
			}).Error("Mempool API request failed, retrying")
			time.Sleep(1 * time.Second)
			continue
		}
		return nil
	}

	return fmt.Errorf("ran out of attempts to request %s", endpoint)
}

func (svc *swapsService) doMempoolRequest(endpoint string, result interface{}) error {
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
	req = req.WithContext(svc.ctx)
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

func (svc *swapsService) bumpAutoswapXpubIndex(swapId uint) {
	indexStr, err := svc.cfg.Get(config.AutoSwapXpubIndexStart, "")
	if err != nil {
		logger.Logger.Error("failed to get auto swap xpub index")
		return
	}
	if indexStr == "" {
		indexStr = "0"
	}
	index, err := strconv.ParseUint(indexStr, 10, 32)
	if err != nil {
		logger.Logger.Error("failed to parse auto swap xpub index")
		return
	}

	err = svc.cfg.SetUpdate(config.AutoSwapXpubIndexStart, strconv.FormatUint(uint64(index+1), 10), "")
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to update auto swap xpub index")
	}
	logger.Logger.WithFields(logrus.Fields{
		"swapId":    swapId,
		"nextIndex": index + 1,
	}).Info("Updated xpub index start for swap address")
}

func (svc *swapsService) deriveAddressFromXpub(xpub string, index uint32) (string, error) {
	var netParams *chaincfg.Params
	switch svc.cfg.GetNetwork() {
	case "bitcoin", "mainnet":
		netParams = &chaincfg.MainNetParams
	case "testnet":
		netParams = &chaincfg.TestNet3Params
	case "regtest":
		netParams = &chaincfg.RegressionNetParams
	case "signet":
		netParams = &chaincfg.SigNetParams
	default:
		return "", fmt.Errorf("unsupported network: %s", svc.cfg.GetNetwork())
	}

	extPubKey, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return "", fmt.Errorf("failed to parse xpub: %w", err)
	}

	externalChain, err := extPubKey.Derive(0)
	if err != nil {
		return "", fmt.Errorf("failed to derive external chain: %w", err)
	}

	addressKey, err := externalChain.Derive(index)
	if err != nil {
		return "", fmt.Errorf("failed to derive address key at index %d: %w", index, err)
	}

	pubKey, err := addressKey.ECPubKey()
	if err != nil {
		return "", fmt.Errorf("failed to get public key: %w", err)
	}

	pubKeyHash := btcutil.Hash160(pubKey.SerializeCompressed())
	address, err := btcutil.NewAddressWitnessPubKeyHash(pubKeyHash, netParams)
	if err != nil {
		return "", fmt.Errorf("failed to create address: %w", err)
	}

	return address.EncodeAddress(), nil
}

func (svc *swapsService) checkAddressHasTransactions(address string, esploraApiRequester func(endpoint string) (interface{}, error)) (bool, error) {
	response, err := esploraApiRequester("/address/" + address + "/txs")
	if err != nil {
		return false, fmt.Errorf("failed to get address transactions: %w", err)
	}

	transactions, ok := response.([]interface{})
	if !ok {
		return false, fmt.Errorf("unexpected response format from esplora API")
	}

	return len(transactions) > 0, nil
}

func (svc *swapsService) getNextUnusedAddressFromXpub() (string, error) {
	destination, _ := svc.cfg.Get(config.AutoSwapDestinationKey, "")
	if destination == "" {
		return "", errors.New("no destination configured")
	}

	if err := svc.validateXpub(destination); err != nil {
		return "", errors.New("destination is not a valid XPUB")
	}

	indexStr, err := svc.cfg.Get(config.AutoSwapXpubIndexStart, "")
	if err != nil {
		return "", err
	}
	if indexStr == "" {
		indexStr = "0"
	}
	index, err := strconv.ParseUint(indexStr, 10, 32)
	if err != nil {
		return "", err
	}

	esploraApiRequester := func(endpoint string) (interface{}, error) {
		url := svc.cfg.GetEnv().LDKEsploraServer + endpoint

		client := http.Client{
			Timeout: time.Second * 10,
		}

		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, err
		}
		req = req.WithContext(svc.ctx)
		res, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		var jsonContent interface{}
		err = json.Unmarshal(body, &jsonContent)
		if err != nil {
			return nil, err
		}
		return jsonContent, nil
	}

	const addressLookAheadLimit = 100

	for i := uint32(index); i < uint32(index)+addressLookAheadLimit; i++ {
		address, err := svc.deriveAddressFromXpub(destination, i)
		if err != nil {
			return "", fmt.Errorf("failed to derive address at index %d: %w", i, err)
		}

		hasTransactions, err := svc.checkAddressHasTransactions(address, esploraApiRequester)
		if err != nil {
			return "", fmt.Errorf("failed to check address for transactions at index %d: %w", i, err)
		}

		if !hasTransactions {
			return address, nil
		}
	}

	return "", fmt.Errorf("could not find unused address within %d addresses starting from index %d", addressLookAheadLimit, index)
}

func (svc *swapsService) validateXpub(xpub string) error {
	_, err := hdkeychain.NewKeyFromString(xpub)
	if err != nil {
		return fmt.Errorf("invalid xpub: %w", err)
	}
	return nil
}
