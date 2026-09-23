package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

type esploraLookup struct {
	serverURL string
	client    http.Client
}

func newEsploraLookup(serverURL string) *esploraLookup {
	return &esploraLookup{
		serverURL: serverURL,
		client: http.Client{
			Timeout: requestTimeout,
		},
	}
}

func (e *esploraLookup) AddressHasTransactions(ctx context.Context, address string) (bool, error) {
	url := e.serverURL + "/address/" + address + "/txs"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}
	res, err := e.client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to get address transactions: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		logger.Logger.WithFields(logrus.Fields{
			"url":         url,
			"status_code": res.StatusCode,
			"body":        string(body),
		}).Error("Esplora endpoint returned non-success code")
		return false, fmt.Errorf("esplora endpoint returned non-success code: %s", string(body))
	}

	var transactions []json.RawMessage
	if err := json.Unmarshal(body, &transactions); err != nil {
		return false, fmt.Errorf("failed to deserialize esplora address txs response: %w", err)
	}

	return len(transactions) > 0, nil
}

func (e *esploraLookup) Close() {
	e.client.CloseIdleConnections()
}

var _ AddressLookup = (*esploraLookup)(nil)
