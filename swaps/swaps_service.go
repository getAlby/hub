package swaps

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
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
}

type SwapsService interface {
	EnableAutoSwaps(ctx context.Context, lnClient lnclient.LNClient) error
	StopAutoSwap()
	ReverseSwap(ctx context.Context, amount uint64, destination string, lnClient lnclient.LNClient) error
}

func NewSwapsService(cfg config.Config, eventPublisher events.EventPublisher, transactionsService transactions.TransactionsService) *swapsService {
	return &swapsService{
		cfg:                 cfg,
		eventPublisher:      eventPublisher,
		transactionsService: transactionsService,
	}
}

func (svc swapsService) EnableAutoSwaps(ctx context.Context, lnClient lnclient.LNClient) error {
	// stop any existing swap process
	svc.StopAutoSwap()

	ctx, cancelFn := context.WithCancel(ctx)
	swapDestination, _ := svc.cfg.Get(config.AutoSwapDestinationKey, "")
	balanceThresholdStr, _ := svc.cfg.Get(config.AutoSwapBalanceThresholdKey, "")

	if swapDestination == "" || balanceThresholdStr == "" {
		cancelFn()
		return errors.New("auto swap not configured")
	}

	parsedBalanceThreshold, err := strconv.ParseUint(balanceThresholdStr, 10, 64)
	if err != nil {
		cancelFn()
		return errors.New("invalid auto swap configuration")
	}

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		for {
			select {
			case <-ticker.C:
				logger.Logger.Info("Checking to see if we can swap")
				balance, err := lnClient.GetBalances(ctx, false)
				if err != nil {
					logger.Logger.WithError(err).Error("Failed to get balance")
					return
				}
				lightningBalance := uint64(balance.Lightning.TotalSpendable)
				balanceThresholdMilliSats := parsedBalanceThreshold * 1000
				if lightningBalance >= balanceThresholdMilliSats {
					// TODO: Change this calcuation
					amount := lightningBalance - balanceThresholdMilliSats
					logger.Logger.WithFields(logrus.Fields{
						"amount":      amount,
						"destination": swapDestination,
					}).Info("Initiating swap")
					// TODO: Should we ourselves add a check that the amount is < 50000
					err := svc.ReverseSwap(ctx, amount/1000, swapDestination, lnClient)
					if err != nil {
						logger.Logger.WithError(err).Error("Failed to swap")
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	svc.cancelFn = cancelFn

	return nil
}

func (svc swapsService) StopAutoSwap() {
	if svc.cancelFn != nil {
		logger.Logger.Info("Stopping swap service...")
		svc.cancelFn()
		logger.Logger.Info("swap service stopped")
	}
}

func (svc swapsService) ReverseSwap(ctx context.Context, amount uint64, destination string, lnClient lnclient.LNClient) error {
	var network, err = boltz.ParseChain(svc.cfg.GetEnv().LDKNetwork)
	if err != nil {
		return err
	}

	ourKeys, err := btcec.NewPrivateKey()
	if err != nil {
		return err
	}

	preimage := make([]byte, 32)
	_, err = rand.Read(preimage)
	if err != nil {
		return err
	}
	preimageHash := sha256.Sum256(preimage)

	boltzApi := &boltz.Api{URL: svc.cfg.GetEnv().BoltzApi}

	swap, err := boltzApi.CreateReverseSwap(boltz.CreateReverseSwapRequest{
		From:           boltz.CurrencyBtc,
		To:             boltz.CurrencyBtc,
		ClaimPublicKey: ourKeys.PubKey().SerializeCompressed(),
		PreimageHash:   preimageHash[:],
		InvoiceAmount:  amount,
	})
	if err != nil {
		return fmt.Errorf("could not create swap: %s", err)
	}

	boltzPubKey, err := btcec.ParsePubKey(swap.RefundPublicKey)
	if err != nil {
		return err
	}

	tree := swap.SwapTree.Deserialize()
	if err := tree.Init(boltz.CurrencyBtc, true, ourKeys, boltzPubKey); err != nil {
		return err
	}

	if err := tree.Check(boltz.ReverseSwap, swap.TimeoutBlockHeight, preimageHash[:]); err != nil {
		return err
	}

	logger.Logger.WithField("swap", swap).Info("Swap created")

	boltzWs := boltzApi.NewWebsocket()
	if err := boltzWs.Connect(); err != nil {
		return fmt.Errorf("could not connect to Boltz websocket: %w", err)
	}

	if err := boltzWs.Subscribe([]string{swap.Id}); err != nil {
		return err
	}

	for update := range boltzWs.Updates {
		parsedStatus := boltz.ParseEvent(update.Status)

		switch parsedStatus {
		case boltz.SwapCreated:
			logger.Logger.WithFields(logrus.Fields{
				"swap":   swap,
				"update": update,
			}).Info("Swap created, paying the invoice")
			metadata := map[string]interface{}{
				"swap": swap,
			}
			_, err := svc.transactionsService.SendPaymentSync(ctx, swap.Invoice, nil, metadata, lnClient, nil, nil)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"swap":   swap,
					"update": update,
				}).Error("Error paying the invoice")
			}
			break

		case boltz.TransactionMempool:
			lockupTransaction, err := boltz.NewTxFromHex(boltz.CurrencyBtc, update.Transaction.Hex, nil)
			if err != nil {
				return err
			}

			vout, _, err := lockupTransaction.FindVout(network, swap.LockupAddress)
			if err != nil {
				return err
			}

			satPerVbyte := float64(2)
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
				satPerVbyte,
				boltzApi,
			)
			if err != nil {
				return fmt.Errorf("could not create claim transaction: %w", err)
			}

			txHex, err := claimTransaction.Serialize()
			if err != nil {
				return fmt.Errorf("could not serialize claim transaction: %w", err)
			}

			txId, err := boltzApi.BroadcastTransaction(boltz.CurrencyBtc, txHex)
			if err != nil {
				return fmt.Errorf("could not broadcast transaction: %w", err)
			}

			logger.Logger.WithField("txId", txId).Info("Transaction broadcasted")
			break

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
			if err := boltzWs.Close(); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
