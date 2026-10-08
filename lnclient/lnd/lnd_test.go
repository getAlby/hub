package lnd

import (
	"testing"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/stretchr/testify/assert"
)

func TestLndInvoiceToTransaction_Amount(t *testing.T) {
	// a settled invoice reports what was paid, which can exceed the invoice amount
	settled := lndInvoiceToTransaction(&lnrpc.Invoice{
		State:       lnrpc.Invoice_SETTLED,
		ValueMsat:   20_000_000,
		AmtPaidMsat: 31_000_000,
	})
	assert.Equal(t, int64(31_000_000), settled.AmountMsat)

	// an amountless invoice is only known by what was paid
	amountless := lndInvoiceToTransaction(&lnrpc.Invoice{
		State:       lnrpc.Invoice_SETTLED,
		AmtPaidMsat: 5_000,
	})
	assert.Equal(t, int64(5_000), amountless.AmountMsat)

	// an unpaid invoice keeps its invoice amount
	open := lndInvoiceToTransaction(&lnrpc.Invoice{
		State:     lnrpc.Invoice_OPEN,
		ValueMsat: 20_000_000,
	})
	assert.Equal(t, int64(20_000_000), open.AmountMsat)
}
