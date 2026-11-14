package chain

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEsploraGetTransactionHex(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/tx/known/hex" {
			_, _ = w.Write([]byte("0200000001abcd\n"))
			return
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("Transaction not found"))
	}))
	defer server.Close()

	lookup := newEsploraLookup(server.URL)
	defer lookup.Close()

	txHex, err := lookup.GetTransactionHex(context.Background(), "known")
	require.NoError(t, err)
	assert.Equal(t, "0200000001abcd", txHex)

	_, err = lookup.GetTransactionHex(context.Background(), "unknown")
	assert.Error(t, err)
}
