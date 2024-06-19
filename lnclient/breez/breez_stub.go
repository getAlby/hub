//go:build skip_breez

package breez

import (
	"github.com/getAlby/nostr-wallet-connect/lnclient"
)

func NewBreezService(mnemonic, apiKey, inviteCode, workDir string) (result lnclient.LNClient, err error) {
	panic("not implemented")
	return nil, nil
}
