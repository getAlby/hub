package cashu

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"sync"
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

const nodeCommandRecover = "recover"

// how long SendPaymentSync waits for a PENDING melt before giving up
const meltPendingTimeout = 60 * time.Second
const meltPendingPollInterval = 2 * time.Second

// how often the background loop claims paid mint quotes and settles
// in-flight melts
const reconcileInterval = 30 * time.Second
const maxMintQuotesPerReconcile = 25

type CashuService struct {
	wallet               *wallet.Wallet
	workDir              string
	hasDifferentMnemonic bool

	// mu serializes every operation that selects proofs or derives
	// outputs. gonuts is not safe for concurrent use: two requests at once
	// (UI polling plus a payment, or two lookups) derived the same NUT-13
	// outputs and the mint rejected the second with "outputs are pending"
	// or "outputs already signed".
	mu sync.Mutex

	stopReconcile chan struct{}
	stopOnce      sync.Once
	// closed after the startup reconcile finished
	startupReconciled chan struct{}
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

	cs.stopReconcile = make(chan struct{})
	cs.startupReconciled = make(chan struct{})
	go cs.reconcileLoop()

	return &cs, nil
}

func (cs *CashuService) Shutdown() error {
	cs.stopOnce.Do(func() { close(cs.stopReconcile) })
	cs.mu.Lock()
	defer cs.mu.Unlock()
	return cs.wallet.Shutdown()
}

// reconcileLoop recovers state left by a crash, a timeout or an older build
// right after startup, then keeps claiming paid invoices so incoming
// payments are credited even when nothing polls the hub.
func (cs *CashuService) reconcileLoop() {
	cs.reconcile()
	close(cs.startupReconciled)
	ticker := time.NewTicker(reconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-cs.stopReconcile:
			return
		case <-ticker.C:
			cs.reconcile()
		}
	}
}

type reconcileResult struct {
	MeltQuotesResolved int    `json:"meltQuotesResolved"`
	MintQuotesClaimed  int    `json:"mintQuotesClaimed"`
	AmountClaimedSat   uint64 `json:"amountClaimedSat"`
	PendingBalanceSat  uint64 `json:"pendingBalanceSat"`
	BalanceSat         uint64 `json:"balanceSat"`
}

func (cs *CashuService) reconcile() reconcileResult {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var result reconcileResult
	resolved, err := cs.wallet.ReconcileMeltQuotes()
	if err != nil {
		logger.Logger.WithError(err).Warn("cashu: failed to check some in-flight melt quotes")
	}
	result.MeltQuotesResolved = resolved

	now := uint64(time.Now().Unix())
	quotes := cs.wallet.MintQuotesToClaim()
	// newest first, and a bounded number per pass, so a hub with many open
	// invoices does not hit the mint's rate limit
	sort.Slice(quotes, func(i, j int) bool { return quotes[i].CreatedAt > quotes[j].CreatedAt })
	checked := 0
	for _, q := range quotes {
		// the mint refuses to issue an expired quote, even a paid one, so
		// only active quotes are worth a request
		if q.QuoteExpiry != 0 && q.QuoteExpiry < now {
			continue
		}
		if checked >= maxMintQuotesPerReconcile {
			break
		}
		checked++
		minted := cs.claimMintQuote(q.QuoteId)
		if minted > 0 {
			result.MintQuotesClaimed++
			result.AmountClaimedSat += minted
		}
	}
	result.PendingBalanceSat = cs.wallet.PendingBalance()
	result.BalanceSat = cs.wallet.GetBalance()
	if result.MeltQuotesResolved > 0 || result.MintQuotesClaimed > 0 {
		logger.Logger.WithField("result", result).Info("cashu: reconciled wallet state")
	}
	return result
}

// claimMintQuote checks a mint quote and mints its ecash if the invoice is
// paid but not issued. Callers must hold cs.mu.
func (cs *CashuService) claimMintQuote(quoteId string) uint64 {
	quote := cs.wallet.GetMintQuoteById(quoteId)
	if quote == nil || quote.State == nut04.Issued {
		return 0
	}
	fields := logrus.Fields{"quoteId": quoteId}
	if bolt11, err := decodepay.Decodepay(quote.PaymentRequest); err == nil {
		fields["paymentHash"] = bolt11.PaymentHash
	}

	state, err := cs.wallet.MintQuoteState(quoteId)
	if err != nil {
		logger.Logger.WithFields(fields).WithError(err).Warn("failed to check invoice state")
		return 0
	}
	// The previous check skipped quotes stored as PAID. When minting a paid
	// quote failed once, the quote stayed PAID forever and the payment was
	// never credited. Retry until the mint says ISSUED.
	if state.State != nut04.Paid {
		return 0
	}
	amountMinted, err := cs.wallet.MintTokens(quoteId)
	if err != nil {
		logger.Logger.WithFields(fields).WithError(err).Warn("failed to mint")
		return 0
	}
	fields["amount"] = amountMinted
	logger.Logger.WithFields(fields).Info("sats successfully minted")
	return amountMinted
}

func (cs *CashuService) SendPaymentSync(invoice string, amountMsat *uint64) (response *lnclient.PayInvoiceResponse, err error) {
	// TODO: support 0-amount invoices
	if amountMsat != nil {
		return nil, errors.New("0-amount invoices not supported")
	}

	paymentRequest, err := decodepay.Decodepay(invoice)
	if err != nil {
		return nil, fmt.Errorf("invalid invoice: %w", err)
	}

	cs.mu.Lock()
	balanceBefore := cs.wallet.GetBalance()

	meltQuoteResponse, err := cs.wallet.RequestMeltQuote(invoice, cs.wallet.CurrentMint())
	if err != nil {
		cs.mu.Unlock()
		logger.Logger.WithError(err).Error("Failed to request melt quote")
		return nil, err
	}
	quoteId := meltQuoteResponse.Quote

	meltResponse, err := cs.wallet.Melt(quoteId)
	if err != nil && !errors.Is(err, wallet.ErrMeltOutcomeUnknown) {
		cs.mu.Unlock()
		logger.Logger.WithError(err).Error("Failed to melt invoice")
		return nil, err
	}
	state := nut05.Pending
	if meltResponse != nil {
		state = meltResponse.State
	}
	// read the balance before releasing the lock, so a concurrent incoming
	// mint cannot skew the fee computed below
	balanceAfter := cs.wallet.GetBalance()
	cs.mu.Unlock()

	// PENDING, or a request that failed without proving the payment failed:
	// poll the quote instead of reporting a failure that may not be true.
	deadline := time.Now().Add(meltPendingTimeout)
	for state == nut05.Pending && time.Now().Before(deadline) {
		time.Sleep(meltPendingPollInterval)
		cs.mu.Lock()
		resp, checkErr := cs.wallet.CheckMeltQuoteState(quoteId)
		cs.mu.Unlock()
		if checkErr != nil {
			logger.Logger.WithError(checkErr).Warn("failed to check melt quote state")
			continue
		}
		state = resp.State
		if state != nut05.Pending {
			cs.mu.Lock()
			balanceAfter = cs.wallet.GetBalance()
			cs.mu.Unlock()
		}
	}

	switch state {
	case nut05.Paid:
	case nut05.Unpaid:
		return nil, errors.New("payment failed: the mint reports the invoice unpaid")
	default:
		// Still in flight. The proofs stay pending and the background
		// reconcile settles them, but the hub has no pending result for
		// SendPaymentSync.
		return nil, errors.New("payment still pending at the mint; the hub will settle it when the mint resolves it")
	}

	quote := cs.wallet.GetMeltQuoteById(quoteId)
	if quote == nil {
		return nil, errors.New("melt quote disappeared after payment")
	}

	preimage := quote.Preimage
	if preimage == "" {
		// NUT-05 allows a PAID melt without a preimage. Nutshell does that
		// when payer and payee use the same mint, because it settles the
		// payment internally and no lightning payment happens. The old
		// code reported these as failed although the ecash was spent and
		// the payee was paid. The hub needs a non-empty preimage to record
		// a settled payment, so use the payment hash. That is the same
		// placeholder this backend already uses for incoming payments. It
		// is not a proof of payment.
		logger.Logger.WithFields(logrus.Fields{
			"paymentHash": paymentRequest.PaymentHash,
			"quoteId":     quoteId,
		}).Warn("cashu: melt PAID without preimage (internal settlement on the mint); recording payment hash as placeholder")
		preimage = paymentRequest.PaymentHash
	}

	// Fee = everything that left the wallet beyond the invoice amount:
	// lightning fee (reserve minus change) plus the keyset input fees
	// (input_fee_ppk) on the melt and on any swap that prepared the
	// inputs. The old code ignored input fees.
	var feeSat uint64
	amountSat := uint64(paymentRequest.MSatoshi / 1000)
	if balanceBefore >= balanceAfter+amountSat {
		feeSat = balanceBefore - balanceAfter - amountSat
	} else if quote.FeeReserve >= quote.ChangeAmount {
		feeSat = quote.FeeReserve - quote.ChangeAmount
	}

	return &lnclient.PayInvoiceResponse{
		Preimage: preimage,
		FeeMsat:  feeSat * 1000,
	}, nil
}

func (cs *CashuService) SendKeysend(amountMsat uint64, destination string, custom_records []lnclient.TLVRecord, preimage string) (*lnclient.PayKeysendResponse, error) {
	return nil, errors.New("keysend not supported")
}

func (cs *CashuService) MakeInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, throughNodePubkey *string) (transaction *lnclient.Transaction, err error) {
	// TODO: support expiry
	if expiry == 0 {
		expiry = lnclient.DEFAULT_INVOICE_EXPIRY
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	mintResponse, err := cs.wallet.RequestMint(uint64(amountMsat/1000), cs.wallet.CurrentMint())
	if err != nil {
		logger.Logger.WithError(err).Error("Failed to mint")
		return nil, err
	}

	mintQuote := cs.wallet.GetMintQuoteById(mintResponse.Quote)
	return cs.cashuMintQuoteToTransaction(mintQuote), nil
}

func (cs *CashuService) MakeHoldInvoice(ctx context.Context, amountMsat int64, description string, descriptionHash string, expiry int64, paymentHash string, minCltvExpiryDelta *uint64) (transaction *lnclient.Transaction, err error) {
	_ = minCltvExpiryDelta
	return nil, errors.New("not implemented")
}

func (cs *CashuService) SettleHoldInvoice(ctx context.Context, preimage string) (err error) {
	return errors.New("not implemented")
}

func (cs *CashuService) CancelHoldInvoice(ctx context.Context, paymentHash string) (err error) {
	return errors.New("not implemented")
}

func (cs *CashuService) LookupInvoice(ctx context.Context, paymentHash string) (transaction *lnclient.Transaction, err error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	mintQuote := cs.getMintQuoteByPaymentHash(paymentHash)
	if mintQuote != nil {
		cs.claimMintQuote(mintQuote.QuoteId)
		// re-read: the copy above predates the mint, so the old code
		// reported a just-credited invoice as unsettled until the next poll
		if refreshed := cs.wallet.GetMintQuoteById(mintQuote.QuoteId); refreshed != nil {
			mintQuote = refreshed
		}
		transaction = cs.cashuMintQuoteToTransaction(mintQuote)
		return transaction, nil
	}

	meltQuote := cs.getMeltQuoteByPaymentHash(paymentHash)
	if meltQuote != nil {
		cs.checkOutgoingPayment(meltQuote)
		if refreshed := cs.wallet.GetMeltQuoteById(meltQuote.QuoteId); refreshed != nil {
			meltQuote = refreshed
		}
		transaction = cs.cashuMeltQuoteToTransaction(meltQuote)
		return transaction, nil
	}

	logger.Logger.WithField("paymentHash", paymentHash).Error("Failed to lookup payment request by payment hash")
	return nil, errors.New("failed to lookup payment request by payment hash")
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
	return errors.New("not implemented")
}

func (cs *CashuService) OpenChannel(ctx context.Context, openChannelRequest *lnclient.OpenChannelRequest) (*lnclient.OpenChannelResponse, error) {
	return nil, errors.New("not implemented")
}

func (cs *CashuService) CloseChannel(ctx context.Context, closeChannelRequest *lnclient.CloseChannelRequest) error {
	return errors.New("not implemented")
}

func (cs *CashuService) GetNewOnchainAddress(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (cs *CashuService) GetUnusedOnchainAddress(ctx context.Context) (string, error) {
	return "", errors.New("not implemented")
}

func (cs *CashuService) GetOnchainBalance(ctx context.Context) (*lnclient.OnchainBalanceResponse, error) {
	return &lnclient.OnchainBalanceResponse{}, nil
}

func (cs *CashuService) RedeemOnchainFunds(ctx context.Context, toAddress string, amountSat uint64, feeRate *uint64, sendAll bool) (string, error) {
	return "", errors.New("not implemented")
}

func (cs *CashuService) ResetRouter(key string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
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
	return "", errors.New("not implemented")
}

func (cs *CashuService) DisconnectPeer(ctx context.Context, peerId string) error {
	return errors.New("not implemented")
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

func (cs *CashuService) UpdateChannel(ctx context.Context, updateChannelRequest *lnclient.UpdateChannelRequest) error {
	return errors.New("not implemented")
}

func (cs *CashuService) GetBalances(ctx context.Context, includeInactiveChannels bool) (*lnclient.BalancesResponse, error) {
	cs.mu.Lock()
	cashuBalance := cs.wallet.GetBalance()
	cs.mu.Unlock()
	balance := int64(cashuBalance * 1000)

	return &lnclient.BalancesResponse{
		Onchain: lnclient.OnchainBalanceResponse{
			PendingBalancesDetails:      []lnclient.PendingBalanceDetails{},
			PendingSweepBalancesDetails: []lnclient.PendingBalanceDetails{}},
		Lightning: lnclient.LightningBalanceResponse{
			TotalSpendableMsat:      balance,
			NextMaxSpendableMsat:    balance,
			NextMaxSpendableMPPMsat: balance,
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
		AmountMsat:      paymentRequest.MSatoshi,
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

	preimage := meltQuote.Preimage
	if preimage == "" && meltQuote.State == nut05.Paid {
		// internal settlement: see SendPaymentSync
		preimage = paymentRequest.PaymentHash
	}
	var fee uint64
	if meltQuote.FeeReserve >= meltQuote.ChangeAmount {
		fee = meltQuote.FeeReserve - meltQuote.ChangeAmount
	}

	return &lnclient.Transaction{
		Type:            constants.TRANSACTION_TYPE_OUTGOING,
		Invoice:         meltQuote.PaymentRequest,
		PaymentHash:     paymentRequest.PaymentHash,
		AmountMsat:      paymentRequest.MSatoshi,
		CreatedAt:       int64(paymentRequest.CreatedAt),
		ExpiresAt:       expiresAt,
		Description:     description,
		DescriptionHash: descriptionHash,
		Preimage:        preimage,
		SettledAt:       settledAt,
		FeesPaidMsat:    int64(fee * 1000),
	}
}

func (cs *CashuService) getMintQuoteByPaymentHash(paymentHash string) *storage.MintQuote {
	mintQuotes := cs.wallet.GetMintQuotes()

	for _, mintQuote := range mintQuotes {
		bolt11, err := decodepay.Decodepay(mintQuote.PaymentRequest)
		if err != nil {
			// one undecodable quote must not hide every other quote
			continue
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
			// one undecodable quote must not hide every other quote
			continue
		}
		if bolt11.PaymentHash == paymentHash {
			return &meltQuote
		}
	}

	return nil
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
			Name:        nodeCommandRecover,
			Description: "Settle stuck payments: check in-flight payments with the mint, release proofs of failed ones and claim paid invoices. Safe to run any time; does not touch the wallet directory.",
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
	case nodeCommandRecover:
		result := cs.reconcile()
		return &lnclient.CustomNodeCommandResponse{
			Response: result,
		}, nil
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
	cs.mu.Lock()
	defer cs.mu.Unlock()
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
	cs.stopOnce.Do(func() { close(cs.stopReconcile) })
	cs.mu.Lock()
	defer cs.mu.Unlock()
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
