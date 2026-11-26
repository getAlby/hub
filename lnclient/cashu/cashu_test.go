package cashu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"crypto/rand"
	"crypto/sha256"
	"os"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/btcec/v2/ecdsa"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/elnosh/gonuts/cashu/nuts/nut04"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/logger"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/zpay32"
	decodepay "github.com/nbd-wtf/ln-decodepay"
	"github.com/stretchr/testify/require"
	"github.com/tyler-smith/go-bip39"
)

// These tests run against a real Nutshell mint with the FakeWallet backend
// (tested with Nutshell 0.21.0 and input_fee_ppk 100, like the Mutinynet mint).
// gonuts does not support v2 keyset IDs yet, so VERSION=0.19.0 makes a fresh
// Nutshell database generate a v1 keyset:
//
//	VERSION=0.19.0 MINT_BACKEND_BOLT11_SAT=FakeWallet MINT_INPUT_FEE_PPK=100 FAKEWALLET_BRR=True mint
//	CASHU_TEST_MINT_URL=http://127.0.0.1:3339 go test ./lnclient/cashu/
//
// They are skipped when CASHU_TEST_MINT_URL is not set.

// mintProxy fronts the test mint. It can strip melt preimages
// (what Nutshell does for same-mint payments), fail mint requests, and it
// records every blinded output so tests can assert none is sent twice.
type mintProxy struct {
	target string

	mu             sync.Mutex
	stripPreimage  bool
	failMintTimes  int
	seen           map[string]bool
	duplicateCount int
}

func (p *mintProxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	body, _ := io.ReadAll(req.Body)
	path := req.URL.Path
	isMint := req.Method == http.MethodPost && path == "/v1/mint/bolt11"
	isMelt := (req.Method == http.MethodPost && path == "/v1/melt/bolt11") ||
		(req.Method == http.MethodGet && strings.HasPrefix(path, "/v1/melt/quote/bolt11/"))

	p.mu.Lock()
	if req.Method == http.MethodPost {
		var payload struct {
			Outputs []struct {
				B string `json:"B_"`
			} `json:"outputs"`
		}
		_ = json.Unmarshal(body, &payload)
		if path != "/v1/restore" {
			for _, o := range payload.Outputs {
				if p.seen[o.B] {
					p.duplicateCount++
				}
				p.seen[o.B] = true
			}
		}
	}
	if isMint && p.failMintTimes > 0 {
		p.failMintTimes--
		p.mu.Unlock()
		rw.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(rw, `{"detail":"outputs are pending","code":11004}`)
		return
	}
	strip := p.stripPreimage
	p.mu.Unlock()

	fwd, _ := http.NewRequest(req.Method, p.target+req.URL.RequestURI(), bytes.NewReader(body))
	fwd.Header = req.Header.Clone()
	fwd.Header.Del("Accept-Encoding") // let the client decompress
	resp, err := http.DefaultClient.Do(fwd)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	if strip && isMelt && resp.StatusCode == http.StatusOK {
		var m map[string]any
		if json.Unmarshal(respBody, &m) == nil {
			m["payment_preimage"] = nil
			respBody, _ = json.Marshal(m)
		}
	}
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(resp.StatusCode)
	rw.Write(respBody)
}

func (p *mintProxy) duplicates() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.duplicateCount
}

type hubEnv struct {
	cs      *CashuService
	proxy   *mintProxy
	mintURL string
}

func setupHubEnv(t *testing.T) *hubEnv {
	t.Helper()
	target := os.Getenv("CASHU_TEST_MINT_URL")
	if target == "" {
		t.Skip("CASHU_TEST_MINT_URL not set (needs a Nutshell FakeWallet mint)")
	}
	logger.Init("3")
	dir := t.TempDir()

	proxy := &mintProxy{target: target, seen: map[string]bool{}}
	srv := httptest.NewServer(proxy)
	t.Cleanup(srv.Close)

	entropy, err := bip39.NewEntropy(128)
	require.NoError(t, err)
	mnemonic, err := bip39.NewMnemonic(entropy)
	require.NoError(t, err)

	client, err := NewCashuService(nil, filepath.Join(dir, "cashu"), mnemonic, srv.URL)
	require.NoError(t, err)
	cs := client.(*CashuService)
	t.Cleanup(func() { cs.Shutdown() })
	<-cs.startupReconciled
	return &hubEnv{cs: cs, proxy: proxy, mintURL: srv.URL}
}

// fund mints ecash through the hub's own receive path
func (e *hubEnv) fund(t *testing.T, sat int64) {
	t.Helper()
	tx, err := e.cs.MakeInvoice(context.Background(), sat*1000, "fund", "", 0, nil)
	require.NoError(t, err)
	got, err := e.cs.LookupInvoice(context.Background(), tx.PaymentHash)
	require.NoError(t, err)
	require.NotNil(t, got.SettledAt, "fake backend invoices are paid at once")
}

func (e *hubEnv) internalInvoice(t *testing.T, sat uint64) string {
	t.Helper()
	resp, err := http.Post(e.mintURL+"/v1/mint/quote/bolt11", "application/json",
		strings.NewReader(fmt.Sprintf(`{"amount":%d,"unit":"sat"}`, sat)))
	require.NoError(t, err)
	defer resp.Body.Close()
	var q nut04.PostMintQuoteBolt11Response
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&q))
	return q.Request
}

func (e *hubEnv) balanceSat(t *testing.T) int64 {
	b, err := e.cs.GetBalances(context.Background(), false)
	require.NoError(t, err)
	return b.Lightning.TotalSpendableMsat / 1000
}

// Bug 1: a same-mint payment (PAID, no preimage) was reported as failed
// although the ecash was spent and the payee was paid.
func TestSendPaymentSync_InternalSettlementWithoutPreimage(t *testing.T) {
	e := setupHubEnv(t)
	e.proxy.stripPreimage = true
	e.fund(t, 300)

	before := e.balanceSat(t)
	invoice := e.internalInvoice(t, 21)
	resp, err := e.cs.SendPaymentSync(invoice, nil)
	require.NoError(t, err, "PAID without preimage is a successful payment")

	bolt11, err := decodepay.Decodepay(invoice)
	require.NoError(t, err)
	require.Equal(t, bolt11.PaymentHash, resp.Preimage, "payment hash is the documented placeholder")

	after := e.balanceSat(t)
	spent := before - after
	require.Equal(t, int64(21)+int64(resp.FeeMsat/1000), spent,
		"reported fee must account for everything that left the wallet (input fees included)")
	require.Greater(t, resp.FeeMsat, uint64(0), "input_fee_ppk=100 makes the melt cost at least 1 sat")

	// the outgoing lookup agrees
	tx, err := e.cs.LookupInvoice(context.Background(), bolt11.PaymentHash)
	require.NoError(t, err)
	require.NotNil(t, tx.SettledAt)
	require.Equal(t, bolt11.PaymentHash, tx.Preimage)

	// Bug 2: later payments must keep working
	for i := 0; i < 3; i++ {
		_, err := e.cs.SendPaymentSync(e.internalInvoice(t, 5), nil)
		require.NoError(t, err, "follow-up payment %d", i)
	}
	require.Zero(t, e.proxy.duplicates())
}

// Bug 3: a paid invoice whose first mint attempt failed stayed pending
// forever, because the quote was stored as PAID and PAID quotes were skipped.
func TestLookupInvoice_RetriesPaidButUnissuedQuote(t *testing.T) {
	e := setupHubEnv(t)
	// fail more times than gonuts retries internally, so the first lookup
	// really fails
	e.proxy.failMintTimes = 5
	tx, err := e.cs.MakeInvoice(context.Background(), 50_000, "receive", "", 0, nil)
	require.NoError(t, err)

	first, err := e.cs.LookupInvoice(context.Background(), tx.PaymentHash)
	require.NoError(t, err)
	require.Nil(t, first.SettledAt)
	require.Equal(t, int64(0), e.balanceSat(t))

	second, err := e.cs.LookupInvoice(context.Background(), tx.PaymentHash)
	require.NoError(t, err)
	require.NotNil(t, second.SettledAt, "paid quote must be minted on the next check")
	require.Equal(t, int64(50), e.balanceSat(t))
}

// The first lookup after payment reports the invoice settled (the old
// code returned the pre-mint copy of the quote).
func TestLookupInvoice_SettledInSameCall(t *testing.T) {
	e := setupHubEnv(t)
	tx, err := e.cs.MakeInvoice(context.Background(), 10_000, "receive", "", 0, nil)
	require.NoError(t, err)
	got, err := e.cs.LookupInvoice(context.Background(), tx.PaymentHash)
	require.NoError(t, err)
	require.NotNil(t, got.SettledAt)
}

// The background reconcile (and the "recover" command) credits paid
// invoices without any lookup.
func TestRecoverCommandClaimsPaidInvoices(t *testing.T) {
	e := setupHubEnv(t)
	e.cs.mu.Lock() // keep the background loop from claiming first
	_, err := e.cs.wallet.RequestMint(77, e.cs.wallet.CurrentMint())
	e.cs.mu.Unlock()
	require.NoError(t, err)

	res, err := e.cs.ExecuteCustomNodeCommand(context.Background(), &lnclient.CustomNodeCommandRequest{Name: nodeCommandRecover})
	require.NoError(t, err)
	r := res.Response.(reconcileResult)
	require.Equal(t, 1, r.MintQuotesClaimed)
	require.Equal(t, uint64(77), r.AmountClaimedSat)
	require.Equal(t, int64(77), e.balanceSat(t))
}

// UI polling, NWC lookups and payments run concurrently in the hub. Without
// serialization they derived the same outputs ("outputs are pending").
func TestConcurrentOperationsDoNotReuseOutputs(t *testing.T) {
	e := setupHubEnv(t)
	e.proxy.stripPreimage = true
	e.fund(t, 400)

	var hashes []string
	for i := 0; i < 6; i++ {
		tx, err := e.cs.MakeInvoice(context.Background(), int64(3+i)*1000, "r", "", 0, nil)
		require.NoError(t, err)
		hashes = append(hashes, tx.PaymentHash)
	}

	var wg sync.WaitGroup
	errs := make(chan error, 64)
	for _, h := range hashes {
		for k := 0; k < 3; k++ {
			wg.Add(1)
			go func(h string) {
				defer wg.Done()
				if _, err := e.cs.LookupInvoice(context.Background(), h); err != nil {
					errs <- err
				}
			}(h)
		}
	}
	for i := 0; i < 5; i++ {
		inv := e.internalInvoice(t, uint64(4+i))
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := e.cs.SendPaymentSync(inv, nil); err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}
	require.Zero(t, e.proxy.duplicates(), "no blinded output may be sent twice")
	for _, h := range hashes {
		tx, err := e.cs.LookupInvoice(context.Background(), h)
		require.NoError(t, err)
		require.NotNil(t, tx.SettledAt)
	}
}

// invoice from a node that is not the mint (FakeWallet pays any invoice,
// fee reserve 2 sat, actual fee 1 sat)
func externalInvoice(t *testing.T, sat uint64) string {
	t.Helper()
	preimage := make([]byte, 32)
	_, err := rand.Read(preimage)
	require.NoError(t, err)
	hash := sha256.Sum256(preimage)
	key, err := btcec.NewPrivateKey()
	require.NoError(t, err)
	inv, err := zpay32.NewInvoice(&chaincfg.SigNetParams, hash, time.Now(),
		zpay32.Amount(lnwire.MilliSatoshi(sat*1000)), zpay32.Description("external"),
		zpay32.PaymentAddr([32]byte(preimage)))
	require.NoError(t, err)
	enc, err := inv.Encode(zpay32.MessageSigner{SignCompact: func(msg []byte) ([]byte, error) {
		return ecdsa.SignCompact(key, chainhash.HashB(msg), true), nil
	}})
	require.NoError(t, err)
	return enc
}

// External payment with a lightning fee reserve: unused reserve comes back as
// NUT-08 change, and the fee the hub records equals what left the wallet.
func TestSendPaymentSync_ExternalWithChange(t *testing.T) {
	e := setupHubEnv(t)
	e.fund(t, 400)
	for _, amt := range []uint64{100, 37, 5} {
		before := e.balanceSat(t)
		inv := externalInvoice(t, amt)
		resp, err := e.cs.SendPaymentSync(inv, nil)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Preimage)
		spent := before - e.balanceSat(t)
		require.Equal(t, int64(amt)+int64(resp.FeeMsat/1000), spent)

		bolt11, _ := decodepay.Decodepay(inv)
		q := e.cs.getMeltQuoteByPaymentHash(bolt11.PaymentHash)
		require.NotNil(t, q)
		require.Greater(t, q.FeeReserve, uint64(0))
		require.Greater(t, q.ChangeAmount, uint64(0), "unused fee reserve must come back as change")
	}
	require.Zero(t, e.proxy.duplicates())
}
