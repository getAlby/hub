package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateRefillInvoiceAmountMatch(t *testing.T) {
	assert.NoError(t, validateRefillInvoiceAmount(1000_000, 1000_000))
	assert.NoError(t, validateRefillInvoiceAmount(0, 0))
}

func TestValidateRefillInvoiceAmountMismatch(t *testing.T) {
	err := validateRefillInvoiceAmount(1000_000, 999_000)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "1000000")
	assert.Contains(t, err.Error(), "999000")
}

func TestValidateRefillInvoiceAmountOffByOneMsat(t *testing.T) {
	// Even a 1-msat mismatch must be rejected (exact-amount guarantee).
	assert.Error(t, validateRefillInvoiceAmount(1000_001, 1000_000))
	assert.Error(t, validateRefillInvoiceAmount(999_999, 1000_000))
}

func TestValidateRefillInvoiceAmountZeroInvoice(t *testing.T) {
	// A decoded invoice reporting 0 msat is never valid for a refill.
	assert.Error(t, validateRefillInvoiceAmount(0, 1000_000))
}
