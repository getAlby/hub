package transactions

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/constants"
	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/lnclient"
	"github.com/getAlby/hub/tests"
)

func TestSelfHoldPaymentSettled(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	preimage := tests.MockLNClientHoldTransaction.Preimage
	paymentHash := tests.MockLNClientHoldTransaction.PaymentHash

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.MakeHoldInvoice(ctx, 1000, "Hold payment test", "", 0, paymentHash, nil, svc.LNClient, nil, nil)
	require.NoError(t, err)
	assert.True(t, transaction.Hold)
	// use the pubkey from the decoded tests.MockLNClientHoldTransaction invoice
	svc.LNClient.(*tests.MockLn).Pubkey = "038a73de75fdc3c7ec951a5e0b0fa95c5cd353bd7ca72df2086aa228c1f92819a5"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := transactionsService.SendPaymentSync(transaction.PaymentRequest, nil, nil, svc.LNClient, nil, nil)
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, result.State)
		assert.Equal(t, true, result.SelfPayment)
		assert.Equal(t, false, result.Hold)
	}()

	// wait for payment to be accepted
	time.Sleep(10 * time.Millisecond)
	settledTransaction, err := transactionsService.SettleHoldInvoice(ctx, preimage, svc.LNClient)
	assert.NoError(t, err)
	require.NotNil(t, settledTransaction)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, settledTransaction.State)
	assert.Equal(t, true, settledTransaction.SelfPayment)
	assert.Equal(t, true, settledTransaction.Hold)
	wg.Wait()
}
func TestSelfHoldPaymentCanceled(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	paymentHash := tests.MockLNClientHoldTransaction.PaymentHash

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	transaction, err := transactionsService.MakeHoldInvoice(ctx, 1000, "Hold payment test", "", 0, paymentHash, nil, svc.LNClient, nil, nil)
	require.NoError(t, err)
	assert.True(t, transaction.Hold)
	// use the pubkey from the decoded tests.MockLNClientHoldTransaction invoice
	svc.LNClient.(*tests.MockLn).Pubkey = "038a73de75fdc3c7ec951a5e0b0fa95c5cd353bd7ca72df2086aa228c1f92819a5"

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := transactionsService.SendPaymentSync(transaction.PaymentRequest, nil, nil, svc.LNClient, nil, nil)
		assert.ErrorIs(t, err, lnclient.NewHoldInvoiceCanceledError())
		assert.Nil(t, result)

		outgoingTransactionType := constants.TRANSACTION_TYPE_OUTGOING
		failedOutgoingTransaction, err := transactionsService.LookupTransaction(ctx, paymentHash, &outgoingTransactionType, svc.LNClient, nil)
		assert.Nil(t, err)
		require.NotNil(t, failedOutgoingTransaction)
		assert.Equal(t, constants.TRANSACTION_STATE_FAILED, failedOutgoingTransaction.State)
		assert.Equal(t, true, failedOutgoingTransaction.SelfPayment)
		assert.Equal(t, false, failedOutgoingTransaction.Hold)
	}()

	// wait for payment to be accepted
	time.Sleep(10 * time.Millisecond)
	err = transactionsService.CancelHoldInvoice(ctx, paymentHash, svc.LNClient)
	assert.NoError(t, err)

	incomingTransactionType := constants.TRANSACTION_TYPE_INCOMING
	updatedHoldTransaction, err := transactionsService.LookupTransaction(ctx, paymentHash, &incomingTransactionType, svc.LNClient, nil)
	assert.Nil(t, err)
	assert.Equal(t, constants.TRANSACTION_STATE_FAILED, updatedHoldTransaction.State)
	assert.Equal(t, true, updatedHoldTransaction.SelfPayment)
	assert.Equal(t, true, updatedHoldTransaction.Hold)
	wg.Wait()
}

func TestWrappedInvoice(t *testing.T) {
	ctx := context.TODO()

	svc, err := tests.CreateTestService(t)
	svc.Cfg.SetUpdate("LNBackendType", config.LDKBackendType, "")
	require.NoError(t, err)
	defer svc.Remove()

	// use the pubkey from Bob's invoice to activate self payments
	svc.LNClient.(*tests.MockLn).Pubkey = "03b6b08ea1f2b70f2260886c9b28fed115b833f18c227abdbf0e8d0629d42b430c"

	transactionsService := NewTransactionsService(svc.DB, svc.EventPublisher)
	// FIXME: undo this change and generate new invoices with 10y expiry
	transactionsService.allowExpiredInvoices = true

	// Charlie creates invoice with payment hash
	// Bob also creates invoice with payment hash, but it's a HOLD invoice one.

	// Create 3 isolated apps: Charlie (invoice creator), Bob (wrapper), Alice (payer)
	charlieApp, _, err := svc.AppsService.CreateApp("Charlie", "", 0, "", nil, []string{constants.MAKE_INVOICE_SCOPE}, true, nil)
	require.NoError(t, err)
	require.NotNil(t, charlieApp)

	bobApp, _, err := svc.AppsService.CreateApp("Bob", "", 0, "", nil, []string{constants.MAKE_INVOICE_SCOPE, constants.PAY_INVOICE_SCOPE}, true, nil)
	require.NoError(t, err)
	require.NotNil(t, bobApp)

	aliceApp, _, err := svc.AppsService.CreateApp("Alice", "", 0, "", nil, []string{constants.PAY_INVOICE_SCOPE}, true, nil)
	require.NoError(t, err)
	require.NotNil(t, aliceApp)

	// created with sandbox.albylabs.com
	// Charlie's 1000 sat invoice
	mockCharlieInvoice := &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         "lnbc10u1p5hk29tdpq2pshjmt9de6zqmmxyqcnqvpsypekzarnnp4qwmtpr4p72ms7gnq3pkfk2876y2msvl33s3840dlp6xsv2w59dpscpp5jfng0kh5ea3gdx6m3fgg9ze8fy223e9qwugpr5pnmhlfnn7r8qlqsp5z8myf87llx9qw8nyhun7hmkhmysn3zn0hdr7afgpm0y6sq4uhrfs9qyysgqcqzp2xqyz5vqysqe04e5jnqnumxwnarneum0x200hag3txc9yxs2y3uvu9mwd39jw0j8sefyqul3e2f6mjr5uh2uk87fsm8s8hlqf7xyv4kwlvl4r6gqzcxtew",
		Description:     "mock hold invoice",
		DescriptionHash: "",
		Preimage:        "92e1454361a292de1d76e1bf3a2aeceee12e5efa3cd7de496a378de36e628056",
		PaymentHash:     "926687daf4cf62869b5b8a50828b274914a8e4a0771011d033ddfe99cfc3383e",
		Amount:          1000_000,
	}

	// Bob's 1100 sat invoice lnbc11u1p5hk2dwdq5fphkcepqd9h8vmmfvdjsnp4qwmtpr4p72ms7gnq3pkfk2876y2msvl33s3840dlp6xsv2w59dpscpp5jfng0kh5ea3gdx6m3fgg9ze8fy223e9qwugpr5pnmhlfnn7r8qlqsp58wvher465xxt3aca7p3t9fgfxrf5kvztq9dkeuef4uk7kvhjmaaq9qyysgqcqzp2xqyz5vqk7sl3ldtt7l98df5h2kd2wfe55lup0vddmxdtkpjarklxj46zrvshgf7rrue2vs6qkaukse5fg024ex6z5cwuqp6qthq7zflxcazgdgq0elt53
	// (same payment hash)

	mockBobInvoice := &lnclient.Transaction{
		Type:            "incoming",
		Invoice:         "lnbc11u1p5hk2dwdq5fphkcepqd9h8vmmfvdjsnp4qwmtpr4p72ms7gnq3pkfk2876y2msvl33s3840dlp6xsv2w59dpscpp5jfng0kh5ea3gdx6m3fgg9ze8fy223e9qwugpr5pnmhlfnn7r8qlqsp58wvher465xxt3aca7p3t9fgfxrf5kvztq9dkeuef4uk7kvhjmaaq9qyysgqcqzp2xqyz5vqk7sl3ldtt7l98df5h2kd2wfe55lup0vddmxdtkpjarklxj46zrvshgf7rrue2vs6qkaukse5fg024ex6z5cwuqp6qthq7zflxcazgdgq0elt53",
		Description:     "mock hold invoice",
		DescriptionHash: "",
		Preimage:        "",
		PaymentHash:     "926687daf4cf62869b5b8a50828b274914a8e4a0771011d033ddfe99cfc3383e",
		Amount:          1100_000,
	}

	svc.LNClient.(*tests.MockLn).MakeInvoiceResponses = []*lnclient.Transaction{
		mockCharlieInvoice,
		mockBobInvoice,
	}
	svc.LNClient.(*tests.MockLn).MakeInvoiceErrors = []error{nil, nil}

	var preimages = []string{tests.MockLNClientHoldTransaction.Preimage, tests.MockLNClientHoldTransaction.Preimage}
	svc.LNClient.(*tests.MockLn).PayInvoiceResponses = []*lnclient.PayInvoiceResponse{{
		Preimage: preimages[0],
	}, {
		Preimage: preimages[1],
	}}
	svc.LNClient.(*tests.MockLn).PayInvoiceErrors = []error{nil, nil}

	// Charlie creates a standard invoice for 1000 sats
	charlieInvoice, err := transactionsService.MakeInvoice(ctx, 1000, "Charlie invoice", "", 0, nil, svc.LNClient, &charlieApp.ID, nil, nil)
	require.NoError(t, err)
	require.False(t, charlieInvoice.Hold)
	require.Equal(t, mockCharlieInvoice.Invoice, charlieInvoice.PaymentRequest)

	// Bob creates a wrapped invoice with the same payment hash but higher amount (1100 sats)
	// Bob acts as an intermediary, adding a fee of 100 sats
	bobWrappedInvoice, err := transactionsService.MakeHoldInvoice(ctx, 1100, "Bob wrapped invoice", "", 0, charlieInvoice.PaymentHash, nil, svc.LNClient, &bobApp.ID, nil)
	require.NoError(t, err)
	require.True(t, bobWrappedInvoice.Hold)
	require.Equal(t, mockBobInvoice.Invoice, bobWrappedInvoice.PaymentRequest)

	// Top up Alice's wallet
	aliceFundingTx := db.Transaction{
		AppId:          &aliceApp.ID,
		RequestEventId: nil,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		FeeReserveMsat: uint64(0),
		AmountMsat:     10000_000,
		PaymentRequest: "lnbc100u1p5hkvrndpz2pshjmt9de6zqmmxyqcnqvpsxqs8xct5wvnp4qwmtpr4p72ms7gnq3pkfk2876y2msvl33s3840dlp6xsv2w59dpscpp5hpd6h7t023cf3q8d06y9slqcnkydffgzun9th5vjm62nsw8wssgqsp5dfddw9ezn93u7g9xmzh4q74kmwxlf0gxgx8c5e4cuu2ce3eapmgq9qyysgqcqzp2xqyz5vqp0t02p3882uhqsz0qf56jgy6mrf2523tudqnf5d2f6f83ud3krd9tu4zkd4yzwyc74acprnvz2853yf9lc89n90sy3r0lvckyy59racq40428t",
		PaymentHash:    "b85babf96f54709880ed7e88587c189d88d4a502e4cabbd192de953838ee8410",
		SelfPayment:    true,
	}
	svc.DB.Save(&aliceFundingTx)

	// Top up Bob's wallet
	bobFundingTx := db.Transaction{
		AppId:          &bobApp.ID,
		RequestEventId: nil,
		Type:           constants.TRANSACTION_TYPE_INCOMING,
		State:          constants.TRANSACTION_STATE_SETTLED,
		FeeReserveMsat: uint64(0),
		AmountMsat:     10000_000,
		PaymentRequest: "lnbc100u1p5hkvy7dpz2pshjmt9de6zqmmxyqcnqvpsxqs8xct5wvnp4qwmtpr4p72ms7gnq3pkfk2876y2msvl33s3840dlp6xsv2w59dpscpp526ulrmuxlgr9zmn56etnkr8eear5xyt2f3h5vecyfe9fwwmjgakqsp5ld7hwc4m6dy3zhy94lr24gklnqyuh7c2xe3wue7qjnlqf59l36rs9qyysgqcqzp2xqyz5vqpj7xg854g2wvdqxcdvt5pjucw0ljxckey3cm82tpx4nlgr8lgs6qv6jwvmnamurguxvwyxt3eft653zeqv7s2gq3gzag925mr796yhqqfaczhq",
		PaymentHash:    "56b9f1ef86fa06516e74d6573b0cf9cf4743116a4c6f4667044e4a973b72476c",
		SelfPayment:    true,
	}
	svc.DB.Save(&bobFundingTx)

	// Alice pays Bob's wrapped invoice
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		result, err := transactionsService.SendPaymentSync(bobWrappedInvoice.PaymentRequest, nil, nil, svc.LNClient, &aliceApp.ID, nil)
		assert.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, result.State)
	}()

	// TODO: rather than wait, listen for the event
	// Wait for Alice's payment to be accepted
	time.Sleep(10 * time.Millisecond)

	// Bob pays Charlie's invoice to get the preimage
	result, err := transactionsService.SendPaymentSync(charlieInvoice.PaymentRequest, nil, nil, svc.LNClient, &bobApp.ID, nil)
	assert.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, result.State)

	// TODO: expect Alice's payment is still held here

	// Bob settles Alice's invoice using the preimage from Charlie
	// TODO: allow passing a payment request
	settledAliceInvoice, err := transactionsService.SettleHoldInvoice(ctx, *result.Preimage, svc.LNClient)
	assert.NoError(t, err)
	require.NotNil(t, settledAliceInvoice)
	assert.Equal(t, constants.TRANSACTION_STATE_SETTLED, settledAliceInvoice.State)
	assert.Equal(t, true, settledAliceInvoice.SelfPayment)
	assert.Equal(t, true, settledAliceInvoice.Hold)

	wg.Wait()
}
