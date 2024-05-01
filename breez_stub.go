//go:build skip_breez

package main

import (
	"github.com/sirupsen/logrus"

	"github.com/getAlby/nostr-wallet-connect/models/lnclient"
)

func NewBreezService(logger *logrus.Logger, mnemonic, apiKey, inviteCode, workDir string) (result lnclient.LNClient, err error) {
	panic("not implemented")
	return nil, nil
}
