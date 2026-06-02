package bip353

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const mockOffer = "lno1zr5qyugqgskrk70kqmuq7v3dnr2fnmhukps9n8hut48vkqpqnskt2svsq"

func TestParseAddress(t *testing.T) {
	tests := []struct {
		name       string
		address    string
		wantUser   string
		wantDomain string
		wantErr    bool
	}{
		{"plain", "alice@example.com", "alice", "example.com", false},
		{"bitcoin prefix", "₿alice@example.com", "alice", "example.com", false},
		{"whitespace", "  alice@example.com  ", "alice", "example.com", false},
		{"missing domain", "alice@", "", "", true},
		{"missing user", "@example.com", "", "", true},
		{"no at", "alice.example.com", "", "", true},
		{"dotted user", "al.ice@example.com", "", "", true},
		{"empty", "", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, domain, err := parseAddress(tt.address)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantUser, user)
			assert.Equal(t, tt.wantDomain, domain)
		})
	}
}

func TestExtractOffer(t *testing.T) {
	tests := []struct {
		name    string
		uri     string
		want    string
		wantErr bool
	}{
		{"offer only", "bitcoin:?lno=" + mockOffer, mockOffer, false},
		{"with onchain", "bitcoin:bc1qexample?lno=" + mockOffer, mockOffer, false},
		{"no offer", "bitcoin:bc1qexample?amount=1", "", true},
		{"no query", "bitcoin:bc1qexample", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offer, err := extractOffer(tt.uri)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, offer)
		})
	}
}

func TestNormalizeTXT(t *testing.T) {
	assert.Equal(t, "bitcoin:?lno=abc", normalizeTXT("bitcoin:?lno=abc"))
	assert.Equal(t, "bitcoin:?lno=abcdef", normalizeTXT("\"bitcoin:?lno=abc\" \"def\""))
	assert.Equal(t, "bitcoin:?lno=abc", normalizeTXT("\"bitcoin:?lno=abc\""))
}

func mockResolver(t *testing.T, body string, status int) func() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "alice.user._bitcoin-payment.example.com", r.URL.Query().Get("name"))
		assert.Equal(t, "TXT", r.URL.Query().Get("type"))
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	original := dohResolverURL
	dohResolverURL = server.URL
	return func() {
		dohResolverURL = original
		server.Close()
	}
}

func TestLookupOffer(t *testing.T) {
	t.Run("resolves DNSSEC-validated offer", func(t *testing.T) {
		body := `{"Status":0,"AD":true,"Answer":[{"type":16,"data":"bitcoin:bc1qexample?lno=` + mockOffer + `"}]}`
		defer mockResolver(t, body, http.StatusOK)()

		offer, err := LookupOffer(context.Background(), "₿alice@example.com")
		require.NoError(t, err)
		assert.Equal(t, mockOffer, offer)
	})

	t.Run("rejects non-DNSSEC-validated record", func(t *testing.T) {
		body := `{"Status":0,"AD":false,"Answer":[{"type":16,"data":"bitcoin:?lno=` + mockOffer + `"}]}`
		defer mockResolver(t, body, http.StatusOK)()

		_, err := LookupOffer(context.Background(), "alice@example.com")
		assert.ErrorContains(t, err, "DNSSEC")
	})

	t.Run("errors when no record found", func(t *testing.T) {
		body := `{"Status":3,"AD":false,"Answer":[]}`
		defer mockResolver(t, body, http.StatusOK)()

		_, err := LookupOffer(context.Background(), "alice@example.com")
		assert.ErrorContains(t, err, "no BIP-353 payment record")
	})

	t.Run("errors when record has no offer", func(t *testing.T) {
		body := `{"Status":0,"AD":true,"Answer":[{"type":16,"data":"bitcoin:bc1qexample?amount=1"}]}`
		defer mockResolver(t, body, http.StatusOK)()

		_, err := LookupOffer(context.Background(), "alice@example.com")
		assert.ErrorContains(t, err, "BOLT-12 offer")
	})

	t.Run("ignores non-bitcoin TXT records", func(t *testing.T) {
		body := `{"Status":0,"AD":true,"Answer":[{"type":16,"data":"some other txt record"},{"type":16,"data":"bitcoin:?lno=` + mockOffer + `"}]}`
		defer mockResolver(t, body, http.StatusOK)()

		offer, err := LookupOffer(context.Background(), "alice@example.com")
		require.NoError(t, err)
		assert.Equal(t, mockOffer, offer)
	})
}
