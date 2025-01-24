package cashu

import (
	"context"
	"errors"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/elnosh/gonuts/wallet"
	"github.com/elnosh/gonuts/wallet/storage"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
)

const nodeCommandRestore = "restore"
const exampleCommandWithArg = "example"

type CashuService struct {
	wallet *wallet.Wallet
}

func NewCashuService(workDir string, mintUrl string) (result lnclient.LNClient, err error) {
	if workDir == "" {
		return nil, errors.New("one or more required cashu configuration are missing")
	}
	if mintUrl == "" {
		return nil, errors.New("no mint URL configured")
	}

	// create dir if not exists
	newpath := filepath.Join(workDir)
	_, err = os.Stat(newpath)
	isFirstSetup := err != nil && errors.Is(err, os.ErrNotExist)

	err = os.MkdirAll(newpath, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create cashu working dir: %v", err)
		return nil, err
	}

	logger.Logger.WithField("mintUrl", mintUrl).Info("Setting up cashu wallet")
	config := wallet.Config{WalletPath: newpath, CurrentMintURL: mintUrl}

	wallet, err := wallet.LoadWallet(config)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to load cashu wallet")
		return nil, err
	}

	cs := CashuService{
		wallet: wallet,
	}

	// try to make an invoice to ensure the mint is running
	// TODO: remove once LoadWallet is improved - see https://github.com/elnosh/gonuts/issues/49
	_, err = cs.MakeInvoice(context.Background(), 10000, "", "", 0)
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to load cashu wallet")
		if isFirstSetup {
			// delete wallet to ensure user gets a fresh start when they try a different mint
			// otherwise the wallet freezes on next startup
			fileErr := os.RemoveAll(newpath)
			if fileErr != nil {
				logger.Logger.WithError(fileErr).Error("Failed to remove broken wallet directory")
			}
		}
		return nil, err
	}

	return &cs, nil
}

func (cs *CashuService) Shutdown() error {
	return nil
}

func (cs *CashuService) SendPaymentSync(ctx context.Context, invoice string, amount *uint64) (response *lnclient.PayInvoiceResponse, err error) {
	// TODO: support 0-amount invoices
	if amount != nil {
		return nil, errors.New("0-amount invoices not supported")
	}

	meltResponse, err := cs.wallet.Melt(invoice, cs.wallet.CurrentMint())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to melt invoice")
		return nil, err
	}

	if meltResponse == nil || meltResponse.Preimage == "" {
		return nil, errors.New("no preimage in melt response")
	}

	// TODO: get fee from melt response
	cashuInvoice, err := cs.wallet.GetInvoiceByPaymentRequest(invoice)
	if err != nil {
		logger.Logger.WithField("invoice", invoice).WithError(err).Error("Failed to get invoice after melting")
		return nil, err
	}

	transaction, err := cs.cashuInvoiceToTransaction(cashuInvoice)
	if err != nil {
		logger.Logger.WithField("invoice", invoice).WithError(err).Error("Failed to convert invoice to transaction")
		return nil, err
	}

	fee := uint64(transaction.FeesPaid)

	return &lnclient.PayInvoiceResponse{
		Preimage: meltResponse.Preimage,
		Fee:      fee,
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
	mintResponse, err := cs.wallet.RequestMint(uint64(amount / 1000))
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to mint")
		return nil, err
	}

	paymentRequest, err := decodepay.Decodepay(mintResponse.Request)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"invoice": mintResponse.Request,
		}).WithError(err).Error("Failed to decode bolt11 invoice")
		return nil, err
	}

	return cs.LookupInvoice(ctx, paymentRequest.PaymentHash)
}

func (cs *CashuService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	cashuInvoice := cs.wallet.GetInvoiceByPaymentHash(paymentHash)

	if cashuInvoice == nil {
		logger.Logger.WithField("paymentHash", paymentHash).Error("Failed to lookup payment request by payment hash")
		return nil, errors.New("failed to lookup payment request by payment hash")
	}

	cs.checkInvoice(cashuInvoice)

	transaction, err = cs.cashuInvoiceToTransaction(cashuInvoice)

	return transaction, nil
}

func (cs *CashuService) ListTransactions(ctx context.Context, from, until, limit, offset uint64, unpaid bool, invoiceType string) (transactions []lnclient.Transaction, err error) {
	transactions = []lnclient.Transaction{}

	invoices := cs.wallet.GetAllInvoices()

	for _, invoice := range invoices {
		invoiceCreated := time.UnixMilli(invoice.CreatedAt * 1000)

		if time.Since(invoiceCreated) < 24*time.Hour {
			cs.checkInvoice(&invoice)
		}

		transaction, err := cs.cashuInvoiceToTransaction(&invoice)
		if err != nil {
			continue
		}
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

func (cs *CashuService) RedeemOnchainFunds(ctx context.Context, toAddress string, amount uint64, sendAll bool) (string, error) {
	return "", nil
}

func (cs *CashuService) ResetRouter(key string) error {
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

func (cs *CashuService) GetBalances(ctx context.Context) (*lnclient.BalancesResponse, error) {
	balanceByMints := cs.wallet.GetBalanceByMints()
	totalBalance := uint64(0)

	for _, balance := range balanceByMints {
		totalBalance += balance
	}

	balance := int64(totalBalance * 1000)

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

func (cs *CashuService) cashuInvoiceToTransaction(cashuInvoice *storage.Invoice) (*lnclient.Transaction, error) {
	paymentRequest, err := decodepay.Decodepay(cashuInvoice.PaymentRequest)
	if err != nil {
		logger.Logger.WithFields(logrus.Fields{
			"invoice": cashuInvoice.PaymentRequest,
		}).WithError(err).Error("Failed to decode bolt11 invoice")
		return nil, err
	}

	var settledAt *int64
	if cashuInvoice.SettledAt > 0 {
		settledAt = &cashuInvoice.SettledAt
	}

	var expiresAt *int64

	expiresAtUnix := time.UnixMilli(int64(paymentRequest.CreatedAt) * 1000).Add(time.Duration(paymentRequest.Expiry) * time.Second).Unix()
	expiresAt = &expiresAtUnix
	description := paymentRequest.Description
	descriptionHash := paymentRequest.DescriptionHash

	invoiceType := "outgoing"
	if cashuInvoice.TransactionType == storage.Mint {
		invoiceType = "incoming"
	}

	return &lnclient.Transaction{
		Type:            invoiceType,
		Invoice:         cashuInvoice.PaymentRequest,
		PaymentHash:     paymentRequest.PaymentHash,
		Amount:          paymentRequest.MSatoshi,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		ExpiresAt:       expiresAt,
		Description:     description,
		DescriptionHash: descriptionHash,
		Preimage:        cashuInvoice.Preimage,
		SettledAt:       settledAt,
		FeesPaid:        int64(cashuInvoice.QuoteAmount*1000) - paymentRequest.MSatoshi,
	}, nil
}

func (cs *CashuService) checkInvoice(cashuInvoice *storage.Invoice) {
	if cashuInvoice.TransactionType == storage.Mint && !cashuInvoice.Paid {
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": cashuInvoice.PaymentHash,
		}).Debug("Checking unpaid invoice")

		proofs, err := cs.wallet.MintTokens(cashuInvoice.Id)
		if err != nil {
			logger.Logger.WithFields(logrus.Fields{
				"paymentHash": cashuInvoice.PaymentHash,
			}).WithError(err).Warn("failed to mint")
		}

		if proofs != nil {
			logger.Logger.WithFields(logrus.Fields{
				"paymentHash": cashuInvoice.PaymentHash,
				"amount":      proofs.Amount(),
			}).Info("sats successfully minted")
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
			Name:        exampleCommandWithArg,
			Description: "Example command with argument",
			Args: []lnclient.CustomNodeCommandArgDef{
				{
					Name:        "hello",
					Description: "world",
				},
			},
		},
	}
}

func (cs *CashuService) ExecuteCustomNodeCommand(ctx context.Context, command *lnclient.CustomNodeCommandRequest) (*lnclient.CustomNodeCommandResponse, error) {
	switch command.Name {
	case nodeCommandRestore:
		return cs.executeCommandRestore(ctx)
	case exampleCommandWithArg:
		if len(command.Args) != 1 {
			return nil, errors.New("please provide an argument")
		}

		return &lnclient.CustomNodeCommandResponse{
			RawJson: []byte("{\"" + command.Args[0].Name + "\": \"" + command.Args[0].Value + "\"}"),
		}, nil
	}

	return nil, lnclient.ErrUnknownCustomNodeCommand
}

func (cs *CashuService) executeCommandRestore(ctx context.Context) (*lnclient.CustomNodeCommandResponse, error) {
	// FIXME: needs latest Cashu changes to be merged
	// mnemonic := cs.wallet.Mnemonic()
	// currentMint := cs.wallet.CurrentMint()
	//
	// if err := cs.wallet.Shutdown(); err != nil {
	// 	return nil, err
	// }
	//
	// if err := os.RemoveAll(cs.workDir); err != nil {
	// 	logger.Logger.WithError(err).Error("Failed to remove wallet directory")
	// 	return nil, err
	// }
	//
	// amountRestored, err := wallet.Restore(cs.workDir, mnemonic, []string{currentMint})
	// if err != nil {
	// 	logger.Logger.WithError(err).Error("Failed restore cashu wallet")
	// 	return nil, err
	// }
	//
	// logger.Logger.WithField("amountRestored", amountRestored).Info("Successfully restored cashu wallet")

	return lnclient.NewCustomNodeCommandResponseEmpty(), nil
}
