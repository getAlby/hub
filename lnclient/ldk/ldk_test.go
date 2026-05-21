package ldk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/tests"
)

func TestGetVssNodeIdentifier(t *testing.T) {
	mnemonic := "thought turkey ask pottery head say catalog desk pledge elbow naive mimic"
	expectedVssNodeIdentifier := "751636"

	svc, err := tests.CreateTestServiceWithMnemonic(t, mnemonic, "123")
	require.NoError(t, err)
	defer svc.Remove()

	vssNodeIdentifier, err := GetVssNodeIdentifier(svc.Keys)
	require.NoError(t, err)

	assert.Equal(t, expectedVssNodeIdentifier, vssNodeIdentifier)
}
func TestGetVssNodeIdentifier2(t *testing.T) {
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	expectedVssNodeIdentifier := "770256"

	svc, err := tests.CreateTestServiceWithMnemonic(t, mnemonic, "123")
	require.NoError(t, err)
	defer svc.Remove()

	vssNodeIdentifier, err := GetVssNodeIdentifier(svc.Keys)
	require.NoError(t, err)

	assert.Equal(t, expectedVssNodeIdentifier, vssNodeIdentifier)
}

func TestSanitizeChainEndpointForBitcoind(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		rpcPort  string
		expected string
	}{
		{
			name:     "adds configured port to host",
			endpoint: "127.0.0.1",
			rpcPort:  "8332",
			expected: "127.0.0.1:8332",
		},
		{
			name:     "preserves endpoint port",
			endpoint: "127.0.0.1:18443",
			rpcPort:  "8332",
			expected: "127.0.0.1:18443",
		},
		{
			name:     "formats ipv6 host",
			endpoint: "[2001:db8::1]",
			rpcPort:  "8332",
			expected: "[2001:db8::1]:8332",
		},
		{
			name:     "strips credentials from url-shaped input",
			endpoint: "user:pass@[2001:db8::1]:18443",
			rpcPort:  "8332",
			expected: "[2001:db8::1]:18443",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeChainEndpoint(tt.endpoint, tt.rpcPort))
		})
	}
}

func TestSanitizeChainEndpointForURLs(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "keeps bare electrum endpoint unchanged",
			endpoint: "electrum.example.com:50002",
			expected: "electrum.example.com:50002",
		},
		{
			name:     "strips url credentials",
			endpoint: "ssl://user:pass@electrum.example.com:50002",
			expected: "ssl://electrum.example.com:50002",
		},
		{
			name:     "strips esplora credentials",
			endpoint: "https://user:pass@esplora.example.com/api",
			expected: "https://esplora.example.com/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeChainEndpoint(tt.endpoint, ""))
		})
	}
}
