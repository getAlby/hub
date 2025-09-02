package bark

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"

	bindings "github.com/getAlby/hub/bark"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/notifications"
	//bindings "github.com/getAlby/second-hub-go/bark"
)

const barkDB = "bark.sqlite"
const vtxoRefreshInterval = 1 * time.Hour

const nodeCommandNewAddress = "new_address"
const nodeCommandMaintenance = "maintenance"
const nodeCommandArkInfo = "ark_info"
const nodeCommandListVTXOs = "list_vtxos"
const nodeCommandGetBoardingAddress = "get_boarding_address"
const nodeCommandBoard = "board"
const nodeCommandOffboard = "offboard"
const nodeCommandSendOnchain = "send_onchain"
const nodeCommandUnilateralExitAll = "unilateral_exit_all"
const nodeCommandPollExitStatus = "poll_exit_status"
const nodeCommandPayToArkAddress = "pay_to_ark_address"
const nodeCommandListUTXOs = "list_utxos"
const nodeCommandListMovements = "list_movements"
const nodeCommandBolt11Invoice = "bolt11_invoice"
const nodeCommandClaimBolt11Payment = "claim_bolt11_payment"

type BarkService struct {
	wallet         *bindings.Wallet
	eventPublisher events.EventPublisher
	cancel         context.CancelFunc
	wg             *sync.WaitGroup
}

func NewBarkService(ctx context.Context, mnemonic, workdir string, eventPublisher events.EventPublisher) (*BarkService, error) {
	err := os.MkdirAll(workdir, 0755)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to create Bark working dir")
		return nil, err
	}

	var wallet *bindings.Wallet

	dbFilePath := filepath.Join(workdir, barkDB)
	if _, err := os.Stat(dbFilePath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			logger.Logger.WithError(err).Error("Failed to check Bark database file")
			return nil, err
		}

		barkConfig := bindings.Config{
			Network:        "signet",
			AspAddress:     "https://ark.signet.2nd.dev",
			EsploraAddress: "https://esplora.signet.2nd.dev",
		}

		wallet, err = bindings.CreateWallet(dbFilePath, mnemonic, barkConfig)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to create Bark wallet")
			return nil, err
		}
	} else {
		wallet, err = bindings.OpenWallet(dbFilePath, mnemonic)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to open Bark wallet")
			return nil, err
		}
	}

	logger.Logger.Info("Performing wallet maintenance")
	if err := wallet.Maintenance(); err != nil {
		logger.Logger.WithError(err).Error("Failed to perform wallet maintenance")
		return nil, err
	}

	pk, err := wallet.NewAddress()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Ark address")
		return nil, err
	}
	logger.Logger.Info("Ark address: ", pk)

	cctx, cancel := context.WithCancel(ctx)
	wg := &sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()

		ticker := time.NewTicker(vtxoRefreshInterval)
		defer ticker.Stop()

		// Perform initial refresh on startup.
		logger.Logger.Info("Refreshing vtxos")
		if err := wallet.RefreshAll(); err != nil {
			logger.Logger.WithError(err).Error("Failed to refresh vtxos")
		}

		for {
			select {
			case <-ticker.C:
				// TODO: run maintenance here

				logger.Logger.Info("Refreshing vtxos")
				// TODO: only refresh VTXOs close to expiry
				if err := wallet.RefreshAll(); err != nil {
					logger.Logger.WithError(err).Error("Failed to refresh vtxos")
				}
			case <-cctx.Done():
				return
			}
		}
	}()

	return &BarkService{
		wallet:         wallet,
		cancel:         cancel,
		wg:             wg,
		eventPublisher: eventPublisher,
	}, nil
}

func (s *BarkService) Shutdown() error {
	logger.Logger.Info("Shutting down Ark client")

	s.cancel()

	logger.Logger.Info("Waiting for Ark procs to finish")
	s.wg.Wait()

	logger.Logger.Info("Ark shutdown complete")

	return nil
}

func (s *BarkService) SendPaymentSync(ctx context.Context, invoice string, amount *uint64) (*lnclient.PayInvoiceResponse, error) {
	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).WithError(err).Error("Failed to decode bolt11 invoice")

		return nil, err
	}

	// Bark won't allow setting amount on invoices that already have an amount set.
	// Bark also does not support msats, offering rounding up or down to the nearest sat.
	var customAmount *uint64
	if paymentRequest.MSatoshi == 0 && amount != nil {
		if *amount%1000 != 0 {
			amountErr := errors.New("bark only supports amounts in whole sats, not msats")

			logger.Logger.WithFields(logrus.Fields{
				"bolt11": invoice,
				"amount": amount,
			}).Error(amountErr)

			return nil, amountErr
		}

		customAmount = amount
	}

	preimage, err := s.wallet.PayBolt11(invoice, customAmount)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
			"amount": amount,
		}).WithError(err).Error("Failed to pay bolt11 invoice")

		return nil, err
	}

	logger.Logger.WithFields(logrus.Fields{
		"bolt11":   invoice,
		"amount":   amount,
		"preimage": preimage,
	}).Info("Successfully paid bolt11 invoice")

	return &lnclient.PayInvoiceResponse{
		Preimage: preimage,
		Fee:      0,
	}, nil
}

func (s *BarkService) SendKeysend(ctx context.Context, amount uint64, destination string, customRecords []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) GetPubkey() string {
	addr, err := s.wallet.NewAddress()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Ark address")
		return ""
	}

	return addr
}

func (s *BarkService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	addr, err := s.wallet.NewAddress()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Ark address")
		return nil, err
	}

	arkInfo, err := s.wallet.ArkInfo()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Ark info")
		return nil, err
	}

	return &lnclient.NodeInfo{
		Pubkey:  addr,
		Network: arkInfo.Network,
	}, nil
}

func (s *BarkService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error) {

	invoice, err := s.wallet.Bolt11Invoice(uint64(amount / 1000))
	if err != nil {
		logger.Logger.WithError(err).Error("failed to create bolt11 invoice")
		return nil, err
	}

	logger.Logger.WithField("invoice", invoice).Info("created bolt11 invoice")

	transaction, err = invoiceToTransaction(invoice)
	if err != nil {
		logger.Logger.WithError(err).Error("failed to map bolt11 invoice to transaction")
		return nil, err

	}

	barkInvoice, err := s.wallet.LookupInvoice(transaction.PaymentHash)
	if err != nil {
		return nil, err
	}
	transaction.Preimage = barkInvoice.PaymentPreimage

	// FIXME: if Alby Hub is restarted we won't be able to claim previously-created invoices
	go func() {
		err := s.wallet.ClaimBolt11Payment(invoice)
		if err != nil {
			logger.Logger.WithError(err).Error("failed to claim bolt11 payment")
			return
		}

		settledAt := time.Now().Unix()
		transaction.SettledAt = &settledAt

		s.eventPublisher.Publish(&events.Event{
			Event:      "nwc_lnclient_payment_received",
			Properties: transaction,
		})
	}()

	return transaction, nil
}

func (s *BarkService) MakeHoldInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) SettleHoldInvoice(ctx context.Context, preimage string) (err error) {
	return errors.New("not implemented")
}

func (s *BarkService) CancelHoldInvoice(ctx context.Context, paymentHash string) (err error) {
	return errors.New("not implemented")
}

func (s *BarkService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	bdkTransactions := s.wallet.OnchainTransactions()

	transactions := []lnclient.OnchainTransaction{}
	for _, bdkTransaction := range bdkTransactions {
		transactions = append(transactions, lnclient.OnchainTransaction{
			AmountSat:        bdkTransaction.AmountSat,
			CreatedAt:        bdkTransaction.CreatedAt,
			State:            bdkTransaction.State,
			Type:             bdkTransaction.TxType,
			NumConfirmations: bdkTransaction.NumConfirmations,
			TxId:             bdkTransaction.Txid,
		})
	}

	return transactions, nil
}

func (s *BarkService) ListChannels(ctx context.Context) (channels []lnclient.Channel, err error) {
	return []lnclient.Channel{}, nil
}

func (s *BarkService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	addr, err := s.wallet.NewAddress()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Ark public key")
		return nil, err
	}

	return &lnclient.NodeConnectionInfo{
		Pubkey: addr,
	}, nil
}

func (s *BarkService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return &lnclient.NodeStatus{
		IsReady: true,
	}, nil
}

func (s *BarkService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return errors.New("not implemented")
}

func (s *BarkService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	return errors.New("not implemented")
}

func (s *BarkService) DisconnectPeer(ctx context.Context, peerId string) error {
	return errors.New("not implemented")
}

func (s *BarkService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	address, err := s.wallet.OnchainAddress()
	if err != nil {
		return "", err
	}
	return address, nil
}

func (s *BarkService) ResetRouter(key string) error {
	return errors.New("not implemented")
}

func (s *BarkService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	walletBalance, err := s.wallet.WalletBalance()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Bark wallet balance")
		return nil, err
	}

	onchainBalance, err := s.wallet.OnchainBalance()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Bark wallet onchain balance")
		return nil, err
	}

	return &lnclient.OnchainBalanceResponse{
		Spendable:                          int64(onchainBalance.TrustedSpendableSat),
		Total:                              int64(onchainBalance.TotalSat),
		PendingBalancesFromChannelClosures: walletBalance.PendingExitSat,
	}, nil
}

func (s *BarkService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	walletBalance, err := s.wallet.WalletBalance()
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Bark wallet balance")
		return nil, err
	}

	onchainBalance, err := s.GetOnchainBalance(ctx)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to get Bark wallet onchain balance")
		return nil, err
	}

	return &lnclient.BalancesResponse{
		Onchain: *onchainBalance,
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable: int64(walletBalance.SpendableSat * 1000),
		},
	}, nil
}

func (s *BarkService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (txId string, err error) {
	// TODO: support fee rate
	txId, err = s.wallet.SendOnchain(toAddress, amount)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to send onchain transaction")
		return "", err
	}

	return txId, nil
}

func (s *BarkService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return errors.New("not implemented")
}

func (s *BarkService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return errors.New("not implemented")
}

func (s *BarkService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) SignMessage(ctx context.Context, message string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *BarkService) GetStorageDir() (string, error) {
	return "", errors.New("not implemented")
}

func (s *BarkService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, errors.New("not implemented")
}

func (s *BarkService) UpdateLastWalletSyncRequest() {
}

func (s *BarkService) GetSupportedNIP47Methods() []string {
	return []string{"pay_invoice", "get_balance", "get_info", "make_invoice"}
}

func (s *BarkService) GetSupportedNIP47NotificationTypes() []string {
	return []string{
		notifications.PAYMENT_RECEIVED_NOTIFICATION,
		// TODO: notifications.PAYMENT_SENT_NOTIFICATION,
	}
}

func (s *BarkService) MakeOffer(ctx context.Context, description string) (string, error) {
	return "", errors.New("not implemented")
}

func (s *BarkService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return []lnclient.CustomNodeCommandDef{
		{
			Name:        nodeCommandNewAddress,
			Description: "Get new Arc address of the wallet.",
			Args:        nil,
		},
		{
			Name:        nodeCommandMaintenance,
			Description: "Run Bark wallet maintenance.",
			Args:        nil,
		},
		{
			Name:        nodeCommandArkInfo,
			Description: "Get information about the Ark network.",
			Args:        nil,
		},
		{
			Name:        nodeCommandListVTXOs,
			Description: "List VTXOs.",
			Args:        nil,
		},
		{
			Name:        nodeCommandGetBoardingAddress,
			Description: "Get the boarding address for the Ark network.",
			Args:        nil,
		},
		{
			Name:        nodeCommandBoard,
			Description: "Move on-chain bitcoin to off-chain bitcoin",
			Args:        nil,
		},
		{
			Name:        nodeCommandOffboard,
			Description: "Move off-chain bitcoin to on-chain bitcoin",
			Args:        nil,
		},
		{
			Name:        nodeCommandSendOnchain,
			Description: "Send funds onchain.",
			Args: []lnclient.CustomNodeCommandArgDef{
				{
					Name:        "address",
					Description: "Destination onchain address.",
				},
				{
					Name:        "amount",
					Description: "Amount to send, in satoshis.",
				},
			},
		},
		{
			Name:        nodeCommandUnilateralExitAll,
			Description: "Unilateral exit.",
			Args:        nil,
		},
		{
			Name:        nodeCommandPollExitStatus,
			Description: "Poll the status of an exit.",
			Args:        nil,
		},
		{
			Name:        nodeCommandPayToArkAddress,
			Description: "Pay to an Ark address.",
			Args: []lnclient.CustomNodeCommandArgDef{
				{
					Name:        "destination",
					Description: "Destination public key.",
				},
				{
					Name:        "amount",
					Description: "Amount to send, in satoshis.",
				},
			},
		},
		{
			Name:        nodeCommandListUTXOs,
			Description: "List UTXOs.",
			Args:        nil,
		},
		{
			Name:        nodeCommandListMovements,
			Description: "List VTXO movements.",
			Args:        nil,
		},
		{
			Name:        nodeCommandBolt11Invoice,
			Description: "Create new Bolt11 invoice.",
			Args: []lnclient.CustomNodeCommandArgDef{
				{
					Name:        "amount",
					Description: "Invoice amount, in satoshis.",
				},
			},
		},
		{
			Name:        nodeCommandClaimBolt11Payment,
			Description: "Claim Bolt11 payment.",
			Args: []lnclient.CustomNodeCommandArgDef{
				{
					Name:        "invoice",
					Description: "Bolt11 invoice.",
				},
			},
		},
	}
}

func (s *BarkService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	switch command.Name {
	case nodeCommandNewAddress:
		addr, err := s.wallet.NewAddress()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to get Ark public key")
			return nil, err
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"address": addr,
			},
		}, nil
	case nodeCommandMaintenance:
		if err := s.wallet.Maintenance(); err != nil {
			logger.Logger.WithError(err).Error("Failed to perform wallet maintenance")
			return nil, err
		}
		return lnclient.NewCustomNodeCommandResponseEmpty(), nil
	case nodeCommandArkInfo:
		arkInfo, err := s.wallet.ArkInfo()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to get Ark info")
			return nil, err
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"network":              arkInfo.Network,
				"asp_pubkey":           arkInfo.AspPubkey,
				"round_interval_sec":   arkInfo.RoundIntervalSec,
				"nb_round_nonces":      arkInfo.NbRoundNonces,
				"vtxo_exit_delta":      arkInfo.VtxoExitDelta,
				"vtxo_expiry_delta":    arkInfo.VtxoExpiryDelta,
				"max_vtxo_amount_sats": arkInfo.MaxVtxoAmountSats,
			},
		}, nil
	case nodeCommandListVTXOs:
		vtxos, err := s.wallet.Vtxos()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to list VTXOs")
			return nil, err
		}

		respVtxos := make([]map[string]interface{}, 0, len(vtxos))
		for _, vtxo := range vtxos {
			respVtxos = append(respVtxos, convertVtxoToCommandResp(vtxo))
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"vtxos": respVtxos,
			},
		}, nil
	case nodeCommandGetBoardingAddress:
		boardAddress, err := s.wallet.OnchainAddress()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to get boarding address")
			return nil, err
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"boarding_address": boardAddress,
			},
		}, nil
	case nodeCommandBoard:
		err := s.wallet.BoardAll()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to board all on-chain funds")
			return nil, err
		}
		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{},
		}, nil
	case nodeCommandOffboard:
		err := s.wallet.OffboardAll()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to offboard all off-chain funds")
			return nil, err
		}
		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{},
		}, nil
	case nodeCommandSendOnchain:
		var addr string
		var amount uint64
		for _, arg := range command.Args {
			switch arg.Name {
			case "address":
				addr = arg.Value
			case "amount":
				var err error
				amount, err = strconv.ParseUint(arg.Value, 10, 64)
				if err != nil {
					logger.Logger.WithError(err).Error("Failed to parse amount for onchain send")
					return nil, err
				}
			}
		}

		if addr == "" || amount == 0 {
			err := errors.New("address and amount are required for onchain send")
			logger.Logger.WithError(err).Error("Invalid arguments for onchain send")
			return nil, err
		}

		txid, err := s.wallet.SendOnchain(addr, amount)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to send onchain transaction")
			return nil, err
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"txid": txid,
			},
		}, nil
	case nodeCommandUnilateralExitAll:
		if err := s.wallet.ExitAll(); err != nil {
			logger.Logger.WithError(err).Error("Failed to perform unilateral exit")
			return nil, err
		}

		return lnclient.NewCustomNodeCommandResponseEmpty(), nil
	case nodeCommandPollExitStatus:
		exitStatus, err := s.wallet.ExitStatus()
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to poll exit status")
			return nil, err
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"done":   exitStatus.Done,
				"height": exitStatus.Height,
			},
		}, nil
	case nodeCommandPayToArkAddress:
		var destination string
		var amount uint64
		for _, arg := range command.Args {
			switch arg.Name {
			case "destination":
				destination = arg.Value
			case "amount":
				var err error
				amount, err = strconv.ParseUint(arg.Value, 10, 64)
				if err != nil {
					logger.Logger.WithError(err).Error("Failed to parse amount for pay to Ark address")
					return nil, err
				}
			}
		}

		if destination == "" || amount == 0 {
			err := errors.New("destination and amount are required for pay to Ark address")
			logger.Logger.WithError(err).Error("Invalid arguments for pay to Ark address")
			return nil, err
		}

		vtxos, err := s.wallet.Send(destination, amount)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to pay to Ark address")
			return nil, err
		}

		respVtxos := make([]map[string]interface{}, 0, len(vtxos))
		for _, vtxo := range vtxos {
			respVtxos = append(respVtxos, convertVtxoToCommandResp(vtxo))
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"vtxos": respVtxos,
			},
		}, nil
	case nodeCommandListUTXOs:
		utxos := s.wallet.Utxos()

		respUtxos := make([]map[string]interface{}, 0, len(utxos))
		for _, utxo := range utxos {
			respUtxos = append(respUtxos, convertUtxoToCommandResp(utxo))
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"utxos": respUtxos,
			},
		}, nil
	case nodeCommandListMovements:
		movements, err := s.wallet.Movements()
		if err != nil {
			return nil, err
		}

		respMovements := make([]map[string]interface{}, 0, len(movements))
		for _, utxo := range movements {
			respMovements = append(respMovements, convertMovementToCommandResp(utxo))
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"movements": respMovements,
			},
		}, nil
	case nodeCommandBolt11Invoice:
		var amount uint64
		if len(command.Args) == 1 && command.Args[0].Name == "amount" {
			var err error
			amount, err = strconv.ParseUint(command.Args[0].Value, 10, 64)
			if err != nil {
				logger.Logger.WithError(err).Error("Failed to parse amount for pay to Ark address")
				return nil, err
			}
		} else {
			err := errors.New("amount is required to create a Bolt11 invoice")
			logger.Logger.WithError(err).Error("Invalid argument for Bolt11 invoice")
			return nil, err
		}

		invoice, err := s.wallet.Bolt11Invoice(amount)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to create Bolt11 invoice")
			return nil, err
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"invoice": invoice,
			},
		}, nil

	case nodeCommandClaimBolt11Payment:
		var invoice string
		if len(command.Args) == 1 && command.Args[0].Name == "invoice" {
			invoice = command.Args[0].Value
		} else {
			err := errors.New("invoice is required to claim a Bolt11 payment")
			logger.Logger.WithError(err).Error("Invalid argument for claiming Bolt11 payment")
			return nil, err
		}

		err := s.wallet.ClaimBolt11Payment(invoice)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to claim Bolt11 payment")
			return nil, err
		}

		return lnclient.NewCustomNodeCommandResponseEmpty(), nil
	}

	return nil, lnclient.ErrUnknownCustomNodeCommand
}

func convertVtxoToCommandResp(vtxo bindings.Vtxo) map[string]interface{} {
	respVtxo := map[string]interface{}{
		"point":         vtxo.Point,
		"amount_sat":    vtxo.AmountSat,
		"user_pubkey":   vtxo.UserPubkey,
		"asp_pubkey":    vtxo.AspPubkey,
		"expiry_height": vtxo.ExpiryHeight,
		"is_arkoor":     vtxo.IsArkoor,
	}

	return respVtxo
}

func convertUtxoToCommandResp(utxo bindings.Utxo) map[string]interface{} {
	var respUtxo map[string]interface{}

	switch utxo := utxo.(type) {
	case bindings.UtxoLocal:
		respUtxo = map[string]interface{}{
			"type":                "local",
			"txid":                utxo.Outpoint.Txid,
			"vout":                utxo.Outpoint.Vout,
			"amount_sat":          utxo.AmountSat,
			"confirmation_height": utxo.ConfirmationHeight,
		}
	case bindings.UtxoExit:
		respUtxo = map[string]interface{}{
			"type":   "exit",
			"vtxo":   convertVtxoToCommandResp(utxo.Vtxo),
			"height": utxo.Height,
		}
	default:
		respUtxo = map[string]interface{}{
			"type": "<unknown>",
		}
	}

	return respUtxo
}

func convertMovementToCommandResp(movement bindings.Movement) map[string]interface{} {
	var kind string
	switch movement.Kind {
	case bindings.MovementKindBoard:
		kind = "Board"
	case bindings.MovementKindRound:
		kind = "Round"
	case bindings.MovementKindOffboard:
		kind = "Offboard"
	case bindings.MovementKindExit:
		kind = "Exit"
	case bindings.MovementKindArkoorSend:
		kind = "ArkoorSend"
	case bindings.MovementKindArkoorReceive:
		kind = "ArkoorReceive"
	case bindings.MovementKindLightningSend:
		kind = "LightningSend"
	case bindings.MovementKindLightningSendRevocation:
		kind = "LightningSendRevocation"
	case bindings.MovementKindLightningReceive:
		kind = "LightningReceive"
	default:
		kind = "Unknown"
	}

	respMovement := map[string]interface{}{
		"id":                  movement.Id,
		"kind":                kind,
		"amount_sent_sat":     movement.AmountSentSat,
		"amount_received_sat": movement.AmountReceivedSat,
		"fees_sat":            movement.FeesSat,
		"created_at":          movement.CreatedAt,
	}

	return respMovement
}

func invoiceToTransaction(invoice string) (*lnclient.Transaction, error) {
	transactionType := "incoming"

	var expiresAt *int64
	var createdAt int64
	var description string
	var descriptionHash string
	var settledAt *int64
	paymentHash := ""
	metadata := map[string]interface{}{}

	paymentRequest, err := decodepay.Decodepay(strings.ToLower(invoice))
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": invoice,
		}).WithError(err).Error("Failed to decode bolt11 invoice")

		return nil, err
	}
	createdAt = int64(paymentRequest.CreatedAt)
	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
	expiresAt = &expiresAtUnix
	description = paymentRequest.Description
	descriptionHash = paymentRequest.DescriptionHash

	paymentHash = paymentRequest.PaymentHash

	return &lnclient.Transaction{
		Type:            transactionType,
		Preimage:        "",
		PaymentHash:     paymentHash,
		SettledAt:       settledAt,
		Amount:          paymentRequest.MSatoshi,
		Invoice:         invoice,
		CreatedAt:       createdAt,
		Description:     description,
		DescriptionHash: descriptionHash,
		ExpiresAt:       expiresAt,
		Metadata:        metadata,
	}, nil
}
