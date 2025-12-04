package ark

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	boltz "github.com/ArkLabsHQ/fulmine/pkg/boltz"
	decodepay "github.com/nbd-wtf/ln-decodepay"

	//swap "github.com/ArkLabsHQ/fulmine/pkg/swap"
	arksdk "github.com/arkade-os/go-sdk"
	"github.com/arkade-os/go-sdk/client"
	grpcclient "github.com/arkade-os/go-sdk/client/grpc"
	indexer "github.com/arkade-os/go-sdk/indexer"
	indexerTransport "github.com/arkade-os/go-sdk/indexer/grpc"
	"github.com/arkade-os/go-sdk/store"
	"github.com/arkade-os/go-sdk/types"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/lnclient/ark/swap"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/nip47/notifications"
	"github.com/sirupsen/logrus"
	"github.com/tyler-smith/go-bip32"
	"github.com/tyler-smith/go-bip39"
)

const nodeCommandReceive = "receive"
const nodeCommandSend = "send"
const nodeCommandListVtxos = "list_vtxos"
const nodeCommandRefreshVtxos = "refresh_vtxos"

const (
	//serverUrl  = "https://mutinynet.arkade.sh"
	serverUrl  = "https://arkade.computer"
	clientType = arksdk.GrpcClient
	walletType = arksdk.SingleKeyWallet
)

type ArkService struct {
	workDir        string
	arkClient      arksdk.ArkClient
	boltzSvc       *boltz.Api
	grpcClient     client.TransportClient
	indexerClient  indexer.Indexer
	pubkey         *btcec.PublicKey
	eventPublisher events.EventPublisher
}

// Experimental Mutinynet Ark client
// currently supports receiving and sending off-chain via custom node commands
// TODO:
// - lightning payments via boltz swaps
// - on-chain receive and send
func NewArkService(ctx context.Context, cfg config.Config, eventPublisher events.EventPublisher, workDir, mnemonic, unlockPassword string) (result lnclient.LNClient, err error) {
	if workDir == "" || mnemonic == "" || unlockPassword == "" {
		return nil, errors.New("one or more required ark configuration are missing")
	}

	seed, err := seedFromMnemonic(mnemonic)
	if err != nil {
		return nil, fmt.Errorf("failed to generate seed from mnemonic: %s", err)
	}

	privKeyBytes, err := hex.DecodeString(seed)
	if err != nil {
		return nil, err
	}
	_, pubkey := btcec.PrivKeyFromBytes(privKeyBytes)

	var arkClient arksdk.ArkClient
	_, err = os.Stat(workDir)

	isNewWallet := err != nil && errors.Is(err, os.ErrNotExist)

	appDataStore, err := store.NewStore(store.Config{
		ConfigStoreType: types.FileStore,
		// AppDataStoreType: types.KVStore,
		BaseDir: workDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to setup app data store: %s", err)
	}

	if isNewWallet {

		arkClient, err = arksdk.NewArkClient(appDataStore)
		if err != nil {
			return nil, fmt.Errorf("failed to create new ark client: %s", err)
		}

		err = arkClient.Init(ctx, arksdk.InitArgs{
			WalletType:          walletType,
			ClientType:          clientType,
			ServerUrl:           serverUrl,
			Password:            unlockPassword,
			Seed:                seed,
			WithTransactionFeed: false, // true crashes
		})

		if err != nil {
			return nil, fmt.Errorf("failed to initialize wallet: %s", err)
		}
	} else {
		arkClient, err = arksdk.LoadArkClient(appDataStore)
		if err != nil {
			return nil, fmt.Errorf("failed to load ark client: %s", err)
		}
	}

	err = arkClient.Unlock(ctx, unlockPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to unlock ark client: %s", err)
	}

	boltzSvc := &boltz.Api{URL: "https://api.ark.boltz.exchange", WSURL: "wss://api.ark.boltz.exchange"}
	grpcClient, err := grpcclient.NewClient(serverUrl)
	if err != nil {
		return nil, err
	}

	indexerClient, err := indexerTransport.NewClient(serverUrl)
	if err != nil {
		return nil, err
	}

	svc := ArkService{
		workDir:        workDir,
		arkClient:      arkClient,
		boltzSvc:       boltzSvc,
		grpcClient:     grpcClient,
		indexerClient:  indexerClient,
		pubkey:         pubkey,
		eventPublisher: eventPublisher,
	}

	return &svc, nil
}

func seedFromMnemonic(mnemonic string) (string, error) {
	// copied from ark-node
	seed := bip39.NewSeed(mnemonic, "")
	key, err := bip32.NewMasterKey(seed)
	if err != nil {
		return "", err
	}

	// TODO: validate this path
	derivationPath := []uint32{
		bip32.FirstHardenedChild + 44,
		bip32.FirstHardenedChild + 1237,
		bip32.FirstHardenedChild + 0,
		0,
		0,
	}

	next := key
	for _, idx := range derivationPath {
		var err error
		if next, err = next.NewChildKey(idx); err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(next.Key), nil
}

func (svc *ArkService) Shutdown() error {
	defer func() {
		// ensure the app does not crash if ark shutdown fails
		if r := recover(); r != nil {
			logger.Logger.WithField("r", r).Error("Failed to cleanly shut down ark backend")
		}
	}()

	svc.arkClient.Stop()
	return nil
}

func (svc *ArkService) SendPaymentSync(invoice string, amount *uint64) (response *lnclient.PayInvoiceResponse, err error) {
	swapTimeout := uint32(15) // seconds
	swapHandler := swap.NewSwapHandler(
		svc.arkClient, svc.grpcClient, svc.indexerClient, svc.boltzSvc, svc.pubkey, swapTimeout,
	)

	unilateralRefund := func(swapData swap.Swap) error {
		//err := s.scheduleSwapRefund(swapData.Id, *swapData.Opts)
		//return err
		return errors.New("TODO implement refund")
	}

	// TODO: use svc context
	swapDetails, err := swapHandler.PayInvoice(context.Background(), invoice, unilateralRefund)
	if err != nil {
		return nil, err
	}

	// TODO: this should be returned by swapHandler / fulmine
	preimage := ""
	func() {
		res, err :=
			svc.boltzSvc.Client.Get(fmt.Sprintf("%s/v2/swap/submarine/%s/preimage", svc.boltzSvc.URL, swapDetails.Id))
		if err != nil {
			logger.Logger.WithError(err).Error("failed to get submarine swap preimage")
			return
		}

		body, err := io.ReadAll(res.Body)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to read submarine preimage response body")
			return
		}

		if res.StatusCode >= 300 {
			logger.Logger.WithFields(logrus.Fields{
				"body":        string(body),
				"status_code": res.StatusCode,
			}).Error("submarine preimage endpoint returned non-success code")
			return
		}

		type SubmarinePreimage struct {
			Preimage string `json:"preimage"`
		}

		submarinePreimage := &SubmarinePreimage{}
		err = json.Unmarshal(body, submarinePreimage)
		if err != nil {
			logger.Logger.WithError(err).Error("Failed to decode submarine preimage API response")
			return
		}
		preimage = submarinePreimage.Preimage
	}()

	return &lnclient.PayInvoiceResponse{
		Preimage: preimage,
		Fee:      0, // TODO: add fee
	}, nil
}

/*
		func (s *Service) scheduleSwapRefund(swapId string, opts vhtlc.Opts) (err error) {
	vHTLC := domain.NewVhtlc(opts)

	unilateral := func() {
		txid, err := s.refundVHTLC(context.Background(), swapId, false, vHTLC)
		if err != nil {
			log.WithError(err).Error("failed to refund vhtlc")
			return
		}

		swapData := domain.Swap{
			Id:         swapId,
			Status:     domain.SwapFailed,
			RedeemTxId: txid,
		}

		if err := s.dbSvc.Swap().Update(context.Background(), swapData); err != nil {
			log.WithError(err).Error("failed to add payment data to db")
		}

		log.Infof("vhtlc refunded %s", txid)
	}

	vtxos, err := s.getVHTLCFunds(context.Background(), []domain.Vhtlc{vHTLC})
	if err != nil {
		log.WithError(err).Error("failed to check vhtlc status")
		return err
	}
	if len(vtxos) == 0 {
		return fmt.Errorf("vhtlc %s not found or already spent", opts.PreimageHash)
	}

	refundLT := opts.RefundLocktime

	if refundLT.IsSeconds() {
		at := time.Unix(int64(refundLT), 0)
		if err := s.schedulerSvc.ScheduleRefundAtTime(at, unilateral); err != nil {
			return err
		}
		log.Debugf("scheduled unilateral refund of swap %s at %s", swapId, at.Format(time.RFC3339))
	} else {
		if err := s.schedulerSvc.ScheduleRefundAtHeight(uint32(refundLT), unilateral); err != nil {
			return err
		}
		log.Debugf("scheduled unilateral refund of swap %s at block height %d", swapId, refundLT)
	}

	return nil
}
*/

func (svc *ArkService) SendKeysend(amount uint64, destination string, custom_records []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	return nil, errors.New("keysend not supported")
}

func (svc *ArkService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, throughNodePubkey *string) (transaction *lnclient.Transaction, err error) {

	swapTimeout := uint32(15) // seconds, seems this is for paying only?
	swapHandler := swap.NewSwapHandler(
		svc.arkClient, svc.grpcClient, svc.indexerClient, svc.boltzSvc, svc.pubkey, swapTimeout,
	)

	postProcess := func(swapData swap.Swap) error {
		if swapData.Status != swap.SwapSuccess {
			logger.Logger.WithField("swap_data", swapData).Error("Swap has unexpected status")
			svc.eventPublisher.Publish(&events.Event{
				Event: "nwc_lnclient_payment_failed",
				Properties: &lnclient.PaymentFailedEventProperties{
					Transaction: transaction,
					Reason:      fmt.Sprintf("Unexpected swap status: %d", swapData.Status),
				},
			})
			return nil
		}

		svc.eventPublisher.Publish(&events.Event{
			Event:      "nwc_lnclient_payment_received",
			Properties: transaction,
		})

		logger.Logger.WithField("swap_data", swapData).Info("Swap success")
		return nil
	}

	makePreimageHex := func() ([]byte, error) {
		bytes := make([]byte, 32) // 32 bytes * 8 bits/byte = 256 bits
		_, err := rand.Read(bytes)
		if err != nil {
			return nil, err
		}
		return bytes, nil
	}

	preImageBytes, err := makePreimageHex()
	if err != nil {
		return nil, err
	}
	preimage := hex.EncodeToString(preImageBytes)

	// TODO: use official swap package once custom preimage can be passed
	swapDetails, err := swapHandler.ReverseSwap(ctx, uint64(amount/1000), preImageBytes, postProcess)
	if err != nil {
		return nil, err
	}

	paymentRequest, err := decodepay.Decodepay(strings.ToLower(swapDetails.Invoice))
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"bolt11": swapDetails.Invoice,
		}).WithError(err).Error("Failed to decode bolt11 invoice")

		return nil, err
	}
	createdAt := int64(paymentRequest.CreatedAt)
	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
	expiresAt := &expiresAtUnix
	description = paymentRequest.Description
	descriptionHash = paymentRequest.DescriptionHash

	transaction = &lnclient.Transaction{
		Invoice:         swapDetails.Invoice,
		Type:            constants.TRANSACTION_TYPE_INCOMING,
		PaymentHash:     hex.EncodeToString(swapDetails.PreimageHash),
		Amount:          amount,
		CreatedAt:       createdAt,
		Preimage:        preimage,
		ExpiresAt:       expiresAt,
		Description:     description,
		DescriptionHash: descriptionHash,
	}
	return transaction, nil
}

func (svc *ArkService) MakeOffer(ctx context.Context, description string) (string, error) {
	return "", errors.New("not supported")
}

func (svc *ArkService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("TODO")
}

func (svc *ArkService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	return nil, errors.New("TODO")
}

func (svc *ArkService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	return &lnclient.NodeInfo{}, nil
}

func (svc *ArkService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	return nil, nil
}

func (svc *ArkService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	return &lnclient.NodeConnectionInfo{}, nil
}

func (svc *ArkService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return nil
}

func (svc *ArkService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, nil
}

func (svc *ArkService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, nil
}

func (svc *ArkService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	return "", nil
}

func (svc *ArkService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	return nil, errors.ErrUnsupported
}

func (svc *ArkService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (string, error) {
	return "", errors.ErrUnsupported
}

func (svc *ArkService) ResetRouter(key string) error {
	return errors.ErrUnsupported
}

func (svc *ArkService) SignMessage(ctx context.Context, message string) (string, error) {
	return "", errors.ErrUnsupported
}

func (svc *ArkService) DisconnectPeer(ctx context.Context, peerId string) error {
	return errors.ErrUnsupported
}

func (svc *ArkService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, errors.ErrUnsupported
}
func (svc *ArkService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return nil, errors.ErrUnsupported
}

func (svc *ArkService) GetStorageDir() (string, error) {
	return "", errors.ErrUnsupported
}
func (svc *ArkService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, errors.ErrUnsupported
}
func (svc *ArkService) UpdateLastWalletSyncRequest() {}

func (svc *ArkService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return &lnclient.NodeStatus{
		IsReady: true,
	}, nil
}

func (svc *ArkService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return errors.ErrUnsupported
}

func (svc *ArkService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return errors.ErrUnsupported
}

func (svc *ArkService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	return errors.ErrUnsupported
}

func (svc *ArkService) MakeHoldInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return nil, errors.ErrUnsupported
}

func (svc *ArkService) SettleHoldInvoice(ctx context.Context, preimage string) (err error) {
	return errors.ErrUnsupported
}

func (svc *ArkService) CancelHoldInvoice(ctx context.Context, paymentHash string) (err error) {
	return errors.ErrUnsupported
}

func (svc *ArkService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	return nil, errors.ErrUnsupported
}

func (svc *ArkService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	balance, err := svc.arkClient.Balance(ctx, false)
	if err != nil {
		return nil, err
	}
	balanceMsat := int64(balance.OffchainBalance.Total) * 1000

	return &lnclient.BalancesResponse{
		Onchain: lnclient.OnchainBalanceResponse{
			Spendable: int64(balance.OnchainBalance.SpendableAmount),
			// TODO: read locked amounts from balance.OnchainBalance.LockedAmount
			Total: int64(balance.OnchainBalance.SpendableAmount),
		},
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable:       balanceMsat,
			TotalReceivable:      0,
			NextMaxSpendable:     balanceMsat,
			NextMaxReceivable:    0,
			NextMaxSpendableMPP:  balanceMsat,
			NextMaxReceivableMPP: 0,
		},
	}, nil
}

func (svc *ArkService) GetSupportedNIP47Methods() []string {
	//return []string{"pay_invoice", "get_balance", "get_budget", "get_info", "make_invoice", "lookup_invoice", "list_transactions", "multi_pay_invoice"}
	return []string{"get_balance", "get_budget", "get_info"}
}

func (svc *ArkService) GetSupportedNIP47NotificationTypes() []string {
	return []string{
		notifications.PAYMENT_RECEIVED_NOTIFICATION,
	}
}

func (svc *ArkService) GetPubkey() string {
	return ""
}

func (svc *ArkService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return []lnclient.CustomNodeCommandDef{
		{
			Name:        nodeCommandReceive,
			Description: "Return addresses to receive funds, either on-chain or natively on Ark.",
			Args:        nil,
		},
		{
			Name:        nodeCommandListVtxos,
			Description: "List Ark VTXOs",
			Args:        nil,
		},
		{
			Name:        nodeCommandRefreshVtxos,
			Description: "Refresh Ark VTXOs",
			Args:        nil,
		},
		{
			Name:        nodeCommandSend,
			Description: "Send funds natively on Ark.",
			Args: []lnclient.CustomNodeCommandArgDef{
				{
					Name:        "address",
					Description: "off-chain ark address",
				},
				{
					Name:        "amount",
					Description: "amount to send in sats",
				},
			},
		},
	}
}

func (svc *ArkService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	switch command.Name {
	case nodeCommandReceive:
		onchainAddress, offchainAddress, boardingAddress, err := svc.arkClient.Receive(ctx)
		if err != nil {
			return nil, err
		}
		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"boardingAddress": boardingAddress,
				"offchainAddress": offchainAddress,
				"onchainAddress":  onchainAddress,
			},
		}, nil
	case nodeCommandListVtxos:
		spendable, spent, err := svc.arkClient.ListVtxos(ctx)
		if err != nil {
			return nil, err
		}
		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"spendable": spendable,
				"spent":     spent,
			},
		}, nil
	case nodeCommandRefreshVtxos:
		roundId, err := svc.arkClient.Settle(ctx)
		if err != nil {
			return nil, err
		}
		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"roundId": roundId,
			},
		}, nil
	case nodeCommandSend:
		var address string
		var amount uint64
		var err error
		for i := range command.Args {
			switch command.Args[i].Name {
			case "address":
				address = command.Args[i].Value
			case "amount":
				amount, err = strconv.ParseUint(string(command.Args[i].Value), 10, 64)
			}
		}
		if err != nil {
			return nil, err
		}

		receivers := []types.Receiver{
			{To: address, Amount: amount},
		}

		redeemTxid, err := svc.arkClient.SendOffChain(ctx, false, receivers)

		if err != nil {
			return nil, err
		}

		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"redeemTxid": redeemTxid,
			},
		}, nil
	}

	return nil, lnclient.ErrUnknownCustomNodeCommand
}
