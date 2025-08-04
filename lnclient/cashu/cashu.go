package cashu

import (
	"context"
	"errors"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/elnosh/gonuts/cashu/nuts/nut04"
	"github.com/elnosh/gonuts/cashu/nuts/nut05"
	"github.com/elnosh/gonuts/wallet"
	"github.com/elnosh/gonuts/wallet/storage"
	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"
)

const nodeCommandRestore = "restore"
const nodeCommandCheckMnemonic = "checkmnemonic"
const nodeCommandResetWallet = "reset"

type CashuService struct {
	wallet               *wallet.Wallet
	workDir              string
	hasDifferentMnemonic bool
}

func NewCashuService(cfg config.Config, workDir, mnemonic, mintUrl string) (result lnclient.LNClient, err error) {
	if workDir == "" {
		return nil, errors.New("one or more required cashu configuration are missing")
	}
	if mintUrl == "" {
		return nil, errors.New("no mint URL configured")
	}

	_, err = os.Stat(workDir)
	isFirstSetup := err != nil && errors.Is(err, os.ErrNotExist)

	if isFirstSetup {
		// make the cashu wallet use the Alby Hub provided mnemonic
		wallet.Restore(workDir, mnemonic, []string{mintUrl})
	}

	logger.Logger.WithField("mintUrl", mintUrl).Info("Setting up cashu wallet")
	config := wallet.Config{WalletPath: workDir, CurrentMintURL: mintUrl}

	cashuWallet, err := wallet.LoadWallet(config)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to load cashu wallet")
		return nil, err
	}

	cs := CashuService{
		wallet:  cashuWallet,
		workDir: workDir,
	}

	if cs.wallet.Mnemonic() != mnemonic {
		logger.Logger.Warn("Cashu is not using Alby Hub mnemonic!")
		cs.hasDifferentMnemonic = true
	}

	return &cs, nil
}

func (cs *CashuService) Shutdown() error {
	return cs.wallet.Shutdown()
}

func (cs *CashuService) SendPaymentSync(ctx context.Context, invoice string, amount *uint64, timeoutSeconds *int64) (response *lnclient.PayInvoiceResponse, err error) {
	// TODO: support 0-amount invoices
	if amount != nil {
		return nil, errors.New("0-amount invoices not supported")
	}

	meltQuoteResponse, err := cs.wallet.RequestMeltQuote(invoice, cs.wallet.CurrentMint())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to request melt quote")
		return nil, err
	}

	meltResponse, err := cs.wallet.Melt(meltQuoteResponse.Quote)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to melt invoice")
		return nil, err
	}

	if meltResponse == nil || meltResponse.Preimage == "" {
		return nil, errors.New("no preimage in melt response")
	}
	fee := meltResponse.FeeReserve - meltResponse.Change.Amount()

	return &lnclient.PayInvoiceResponse{
		Preimage: meltResponse.Preimage,
		Fee:      fee * 1000,
	}, nil
}

func (cs *CashuService) SendKeysend(ctx context.Context, amount uint64, destination string, custom_records []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	return nil, errors.New("keysend not supported")
}

func (cs *CashuService) MakeInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64) (transaction *lnclient.Transaction, err error) {
	// TODO: support expiry
	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}
	mintResponse, err := cs.wallet.RequestMint(uint64(amount/1000), cs.wallet.CurrentMint())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to mint")
		return nil, err
	}

	mintQuote := cs.wallet.GetMintQuoteById(mintResponse.Quote)
	return cs.cashuMintQuoteToTransaction(mintQuote), nil
}

func (cs *CashuService) MakeHoldInvoice(ctx context.Context, amount int64, description string, descriptionHash string, expiry int64, paymentHash string) (transaction *lnclient.Transaction, err error) {
	return nil, errors.New("not implemented")
}

func (cs *CashuService) SettleHoldInvoice(ctx context.Context, preimage string) (err error) {
	return errors.New("not implemented")
}

func (cs *CashuService) CancelHoldInvoice(ctx context.Context, paymentHash string) (err error) {
	return errors.New("not implemented")
}

func (cs *CashuService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	mintQuote := cs.getMintQuoteByPaymentHash(paymentHash)
	if mintQuote != nil {
		cs.checkIncomingPayment(mintQuote)
		transaction = cs.cashuMintQuoteToTransaction(mintQuote)
		return transaction, nil
	}

	meltQuote := cs.getMeltQuoteByPaymentHash(paymentHash)
	if meltQuote != nil {
		cs.checkOutgoingPayment(meltQuote)
		transaction = cs.cashuMeltQuoteToTransaction(meltQuote)
		return transaction, nil
	}

	logger.Logger.WithField("paymentHash", paymentHash).Error("Failed to lookup payment request by payment hash")
	return nil, errors.New("failed to lookup payment request by payment hash")
}

func (cs *CashuService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	mintQuotes := cs.wallet.GetMintQuotes()
	meltQuotes := cs.wallet.GetMeltQuotes()
	transactions = make([]lnclient.Transaction, 0, len(mintQuotes)+len(meltQuotes))

	for _, mintQuote := range mintQuotes {
		invoiceCreated := time.UnixMilli(mintQuote.CreatedAt * 1000)
		if time.Since(invoiceCreated) < 24*time.Hour && mintQuote.State != nut04.Paid {
			cs.checkIncomingPayment(&mintQuote)
		}

		transaction := cs.cashuMintQuoteToTransaction(&mintQuote)
		if transaction.SettledAt == nil {
			continue
		}
		transactions = append(transactions, *transaction)
	}

	for _, meltQuote := range meltQuotes {
		transaction := cs.cashuMeltQuoteToTransaction(&meltQuote)
		if transaction.SettledAt == nil {
			continue
		}
		transactions = append(transactions, *transaction)
	}

	// sort by created date descending
	sort.SliceStable(transactions, func(i, j int) bool {
		return transactions[i].CreatedAt > transactions[j].CreatedAt
	})

	return transactions, nil
}

func (cs *CashuService) GetInfo(ctx context.Context) (info *lnclient.NodeInfo, err error) {
	return &lnclient.NodeInfo{
		Alias:       "NWC (Cashu)",
		Color:       "#897FFF",
		Pubkey:      "",
		Network:     "bitcoin",
		BlockHeight: 0,
		BlockHash:   "",
	}, nil
}

func (cs *CashuService) ListChannels(ctx context.Context) ([]lnclient.Channel, error) {
	return nil, nil
}

func (cs *CashuService) GetNodeConnectionInfo(ctx context.Context) (nodeConnectionInfo *lnclient.NodeConnectionInfo, err error) {
	return &lnclient.NodeConnectionInfo{}, nil
}

func (cs *CashuService) ConnectPeer(ctx context.Context, connectPeerRequest *lnclient.ConnectPeerRequest) error {
	return nil
}

func (cs *CashuService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, nil
}

func (cs *CashuService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) (*lnclient.CloseChannelResponse, error) {
	return nil, nil
}

func (cs *CashuService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	return "", nil
}

func (cs *CashuService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	return &lnclient.OnchainBalanceResponse{
		Spendable: 0,
		Total:     0,
	}, nil
}

func (cs *CashuService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, feeRate *uint64, sendAll bool) (string, error) {
	return "", nil
}

func (cs *CashuService) ResetRouter(key string) error {
	mnemonic := cs.wallet.Mnemonic()
	currentMint := cs.wallet.CurrentMint()

	if err := cs.wallet.Shutdown(); err != nil {
		return err
	}

	if err := os.RemoveAll(cs.workDir); err != nil {
		logger.Logger.WithError(err).Error("Failed to remove wallet directory")
		return err
	}

	amountRestored, err := wallet.Restore(cs.workDir, mnemonic, []string{currentMint})
	if err != nil {
		logger.Logger.WithError(err).Error("Failed restore cashu wallet")
		return err
	}

	logger.Logger.WithField("amountRestored", amountRestored).Info("Successfully restored cashu wallet")
	return nil
}

func (cs *CashuService) SignMessage(ctx context.Context, message string) (string, error) {
	return "", nil
}

func (cs *CashuService) DisconnectPeer(ctx context.Context, peerId string) error {
	return nil
}

func (cs *CashuService) ListPeers(ctx context.Context) ([]lnclient.PeerDetails, error) {
	return nil, nil
}
func (cs *CashuService) GetLogOutput(ctx context.Context, maxLen int) ([]byte, error) {
	return nil, nil
}

func (cs *CashuService) GetStorageDir() (string, error) {
	return "", nil
}
func (cs *CashuService) GetNetworkGraph(ctx context.Context, nodeIds []string) (lnclient.NetworkGraphResponse, error) {
	return nil, nil
}
func (cs *CashuService) UpdateLastWalletSyncRequest() {}

func (cs *CashuService) GetNodeStatus(ctx context.Context) (nodeStatus *lnclient.NodeStatus, err error) {
	return &lnclient.NodeStatus{
		IsReady: true,
	}, nil
}

func (cs *CashuService) SendPaymentProbes(ctx context.Context, invoice string) error {
	return nil
}

func (cs *CashuService) SendSpontaneousPaymentProbes(ctx context.Context, amountMsat uint64, nodeId string) error {
	return nil
}

func (cs *CashuService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	return nil
}

func (cs *CashuService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	cashuBalance := cs.wallet.GetBalance()
	balance := int64(cashuBalance * 1000)

	return &lnclient.BalancesResponse{
		Onchain: lnclient.OnchainBalanceResponse{
			Spendable: 0,
			Total:     0,
		},
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendable:       balance,
			TotalReceivable:      0,
			NextMaxSpendable:     balance,
			NextMaxReceivable:    0,
			NextMaxSpendableMPP:  balance,
			NextMaxReceivableMPP: 0,
		},
	}, nil
}

func (cs *CashuService) cashuMintQuoteToTransaction(mintQuote *storage.MintQuote) *lnclient.Transaction {
	// note: if a mint quote exists, then the payment request is already valid
	paymentRequest, _ := decodepay.Decodepay(mintQuote.PaymentRequest)
	var settledAt *int64
	if mintQuote.SettledAt > 0 {
		settledAt = &mintQuote.SettledAt
	}

	var expiresAt *int64

	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
	expiresAt = &expiresAtUnix
	description := paymentRequest.Description
	descriptionHash := paymentRequest.DescriptionHash

	return &lnclient.Transaction{
		Type:        constants.TRANSACTION_TYPE_INCOMING,
		Invoice:     mintQuote.PaymentRequest,
		PaymentHash: paymentRequest.PaymentHash,
		// note: setting dummy preimage so that it gets marked as settled
		Preimage:        paymentRequest.PaymentHash,
		Amount:          paymentRequest.MSatoshi,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		ExpiresAt:       expiresAt,
		Description:     description,
		DescriptionHash: descriptionHash,
		SettledAt:       settledAt,
	}
}

func (cs *CashuService) cashuMeltQuoteToTransaction(meltQuote *storage.MeltQuote) *lnclient.Transaction {
	// note: if a melt quote exists, then the payment request is already valid
	paymentRequest, _ := decodepay.Decodepay(meltQuote.PaymentRequest)

	var settledAt *int64
	if meltQuote.SettledAt > 0 {
		settledAt = &meltQuote.SettledAt
	}

	var expiresAt *int64

	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
	expiresAt = &expiresAtUnix
	description := paymentRequest.Description
	descriptionHash := paymentRequest.DescriptionHash

	return &lnclient.Transaction{
		Type:            constants.TRANSACTION_TYPE_OUTGOING,
		Invoice:         meltQuote.PaymentRequest,
		PaymentHash:     paymentRequest.PaymentHash,
		Amount:          paymentRequest.MSatoshi,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		ExpiresAt:       expiresAt,
		Description:     description,
		DescriptionHash: descriptionHash,
		Preimage:        meltQuote.Preimage,
		SettledAt:       settledAt,
		FeesPaid:        int64(meltQuote.FeeReserve * 1000),
	}
}

func (cs *CashuService) getMintQuoteByPaymentHash(paymentHash string) *storage.MintQuote {
	mintQuotes := cs.wallet.GetMintQuotes()

	for _, mintQuote := range mintQuotes {
		bolt11, err := decodepay.Decodepay(mintQuote.PaymentRequest)
		if err != nil {
			return nil
		}
		if bolt11.PaymentHash == paymentHash {
			return &mintQuote
		}
	}

	return nil
}

func (cs *CashuService) getMeltQuoteByPaymentHash(paymentHash string) *storage.MeltQuote {
	meltQuotes := cs.wallet.GetMeltQuotes()

	for _, meltQuote := range meltQuotes {
		bolt11, err := decodepay.Decodepay(meltQuote.PaymentRequest)
		if err != nil {
			return nil
		}
		if bolt11.PaymentHash == paymentHash {
			return &meltQuote
		}
	}

	return nil
}

func (cs *CashuService) checkIncomingPayment(mintQuote *storage.MintQuote) {
	bolt11, _ := decodepay.Decodepay(mintQuote.PaymentRequest)

	if mintQuote.State != nut04.Paid {
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": bolt11.PaymentHash,
		}).Debug("Checking unpaid invoice")

		mintQuoteState, err := cs.wallet.MintQuoteState(mintQuote.QuoteId)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"paymentHash": bolt11.PaymentHash,
			}).WithError(err).Warn("failed to check invoice state")
			return
		}

		if mintQuoteState.State == nut04.Paid {
			amountMinted, err := cs.wallet.MintTokens(mintQuote.QuoteId)
			if err != nil {
				logger.Logger.WithFields(logrus.Fields{
					"paymentHash": bolt11.PaymentHash,
				}).WithError(err).Warn("failed to mint")
			}
			if amountMinted > 0 {
				logger.Logger.WithFields(logrus.Fields{
					"paymentHash": bolt11.PaymentHash,
					"amount":      amountMinted,
				}).Info("sats successfully minted")
			}
		}
	}
}

func (cs *CashuService) checkOutgoingPayment(meltQuote *storage.MeltQuote) {
	bolt11, _ := decodepay.Decodepay(meltQuote.PaymentRequest)

	if meltQuote.State != nut05.Paid {
		_, err := cs.wallet.CheckMeltQuoteState(meltQuote.QuoteId)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"paymentHash": bolt11.PaymentHash,
			}).WithError(err).Warn("failed to check invoice state")
		}

	}

}

func (cs *CashuService) GetSupportedNIP47Methods() []string {
	return []string{"pay_invoice", "get_balance", "get_budget", "get_info", "make_invoice", "lookup_invoice", "list_transactions", "multi_pay_invoice"}
}

func (cs *CashuService) GetSupportedNIP47NotificationTypes() []string {
	return []string{}
}

func (svc *CashuService) GetPubkey() string {
	return ""
}

func (cs *CashuService) GetCustomNodeCommandDefinitions() []lnclient.CustomNodeCommandDef {
	return []lnclient.CustomNodeCommandDef{
		{
			Name:        nodeCommandRestore,
			Description: "Restore cashu tokens after the wallet had a stuck payment.",
			Args:        nil,
		},
		{
			Name:        nodeCommandCheckMnemonic,
			Description: "Check if your cashu wallet uses the same mnemonic as Alby Hub.",
			Args:        nil,
		},
		{
			Name:        nodeCommandResetWallet,
			Description: "Completely resets your cashu wallet. Only do this if you have no funds.",
			Args:        nil,
		},
	}
}

func (cs *CashuService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	switch command.Name {
	case nodeCommandRestore:
		return cs.executeCommandRestore()
	case nodeCommandResetWallet:
		return cs.executeCommandResetWallet()
	case nodeCommandCheckMnemonic:
		return &lnclient.CustomNodeCommandResponse{
			Response: map[string]interface{}{
				"matches": !cs.hasDifferentMnemonic,
			},
		}, nil
	}

	return nil, lnclient.ErrUnknownCustomNodeCommand
}

func (cs *CashuService) executeCommandRestore() (*lnclient.CustomNodeCommandResponse, error) {
	mnemonic := cs.wallet.Mnemonic()
	currentMintUrl := cs.wallet.CurrentMint()

	if err := cs.wallet.Shutdown(); err != nil {
		return nil, err
	}

	if err := os.Rename(cs.workDir, cs.workDir+strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
		logger.Logger.WithError(err).Error("Failed to rename wallet directory")
		return nil, err
	}

	amountRestored, err := wallet.Restore(cs.workDir, mnemonic, []string{currentMintUrl})
	if err != nil {
		logger.Logger.WithError(err).Error("Failed restore cashu wallet")
		return nil, err
	}

	logger.Logger.WithField("amountRestored", amountRestored).Info("Successfully restored cashu wallet")

	config := wallet.Config{WalletPath: cs.workDir, CurrentMintURL: currentMintUrl}
	cashuWallet, err := wallet.LoadWallet(config)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to load cashu wallet")
		return nil, err
	}

	cs.wallet = cashuWallet

	return &lnclient.CustomNodeCommandResponse{
		Response: map[string]interface{}{
			"amountRestored": amountRestored,
			"message":        "Restore successful.",
		},
	}, nil
}

func (cs *CashuService) executeCommandResetWallet() (*lnclient.CustomNodeCommandResponse, error) {
	if err := cs.wallet.Shutdown(); err != nil {
		return nil, err
	}

	if err := os.Rename(cs.workDir, cs.workDir+strconv.FormatInt(time.Now().Unix(), 10)); err != nil {
		logger.Logger.WithError(err).Error("Failed to rename wallet directory")
		return nil, err
	}

	go func() {
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}()

	return &lnclient.CustomNodeCommandResponse{
		Response: map[string]interface{}{
			"message": "Reset successful. Your hub will shutdown in 10 seconds...",
		},
	}, nil
}

func (svc *CashuService) MakeOffer(ctx context.Context, description string) (string, error) {
	return "", errors.New("not supported")
}

func (cs *CashuService) ListOnchainTransactions(ctx context.Context) ([]lnclient.OnchainTransaction, error) {
	return nil, errors.ErrUnsupported
}
