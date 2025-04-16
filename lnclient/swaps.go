package lnclient

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/BoltzExchange/boltz-client/v2/pkg/boltz"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

type GetBalancesFn func(context.Context, bool) (*BalancesResponse, error)

func StartAutoSwap(ctx context.Context, balanceThreshold uint64, destination string, getBalancesFn GetBalancesFn) error {
	go func() {
		// TODO: Do we want to check every hour?
		ticker := time.NewTicker(1 * time.Hour)
		for {
			select {
			case <-ticker.C:
				logger.Logger.Info("Checking to see if we can swap")
				balance, err := getBalancesFn(ctx, false)
				if err != nil {
					logger.Logger.WithError(err).Error("Failed to get balance")
					return
				}
				lightningBalance := uint64(balance.Lightning.TotalSpendable)
				if lightningBalance >= balanceThreshold {
					logger.Logger.Info("Initiating swap")
					// TODO: Change this calcuation
					amount := lightningBalance - balanceThreshold
					ReverseSwap(amount, destination)
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return nil
}

func ReverseSwap(amount uint64, destination string) error {
	// TODO: Make these configurable from env or using network env var
	const endpoint = "wss://api.testnet.boltz.exchange/v2/ws"
	var network = boltz.MainNet

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

	boltzApi := &boltz.Api{URL: endpoint}

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
	if err := tree.Init(boltz.CurrencyBtc, false, ourKeys, boltzPubKey); err != nil {
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
			}).Info("Swap created, waiting for invoice to be paid")
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
			if err := boltzWs.Close(); err != nil {
				return err
			}
			break
		}
	}
	return nil
}
