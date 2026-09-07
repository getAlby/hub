package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetSupportedExtensions(t *testing.T) {
	tests := []struct {
		name              string
		methods           []string
		notificationTypes []string
		expected          []string
	}{
		{
			name:     "core methods only always support metadata conventions",
			methods:  []string{GET_INFO_METHOD, GET_BALANCE_METHOD},
			expected: []string{METADATA_EXTENSION},
		},
		{
			name:              "notifications permission advertises notifications extension",
			methods:           []string{GET_INFO_METHOD},
			notificationTypes: []string{"payment_received", "payment_sent"},
			expected:          []string{NOTIFICATIONS_EXTENSION, METADATA_EXTENSION},
		},
		{
			name: "hold invoice methods advertise hold invoice extension",
			methods: []string{
				MAKE_HOLD_INVOICE_METHOD, CANCEL_HOLD_INVOICE_METHOD, SETTLE_HOLD_INVOICE_METHOD,
			},
			expected: []string{HOLD_INVOICES_EXTENSION, METADATA_EXTENSION},
		},
		{
			name:     "partial hold invoice methods do not advertise hold invoice extension",
			methods:  []string{MAKE_HOLD_INVOICE_METHOD, CANCEL_HOLD_INVOICE_METHOD},
			expected: []string{METADATA_EXTENSION},
		},
		{
			name:     "pay_keysend advertises keysend extension",
			methods:  []string{PAY_KEYSEND_METHOD},
			expected: []string{KEYSEND_EXTENSION, METADATA_EXTENSION},
		},
		{
			name:     "multi_pay_keysend advertises keysend extension",
			methods:  []string{MULTI_PAY_KEYSEND_METHOD},
			expected: []string{KEYSEND_EXTENSION, METADATA_EXTENSION},
		},
		{
			name:     "list_transactions advertises transaction history extension",
			methods:  []string{LIST_TRANSACTIONS_METHOD},
			expected: []string{TRANSACTION_HISTORY_EXTENSION, METADATA_EXTENSION},
		},
		{
			name: "all capabilities advertise all extensions in numerical order",
			methods: []string{
				MAKE_HOLD_INVOICE_METHOD, CANCEL_HOLD_INVOICE_METHOD, SETTLE_HOLD_INVOICE_METHOD,
				PAY_KEYSEND_METHOD,
				LIST_TRANSACTIONS_METHOD,
			},
			notificationTypes: []string{"payment_received"},
			expected: []string{
				NOTIFICATIONS_EXTENSION,
				HOLD_INVOICES_EXTENSION,
				KEYSEND_EXTENSION,
				TRANSACTION_HISTORY_EXTENSION,
				METADATA_EXTENSION,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetSupportedExtensions(tt.methods, tt.notificationTypes))
		})
	}
}
