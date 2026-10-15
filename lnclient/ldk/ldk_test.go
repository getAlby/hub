package ldk

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/getAlby/ldk-node-go/ldk_node"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
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

func makeLsps2OpeningFeeParams(minFeeMsat uint64, proportional uint32, minPaymentSizeMsat uint64, maxPaymentSizeMsat uint64) ldk_node.Lsps2OpeningFeeParams {
	return ldk_node.Lsps2OpeningFeeParams{
		MinFeeMsat:           minFeeMsat,
		Proportional:         proportional,
		ValidUntil:           "2035-01-01T00:00:00Z",
		MinLifetime:          4032,
		MaxClientToSelfDelay: 2016,
		MinPaymentSizeMsat:   minPaymentSizeMsat,
		MaxPaymentSizeMsat:   maxPaymentSizeMsat,
		Promise:              "promise",
	}
}

func TestComputeLsps2MaxTotalOpeningFeeMsat(t *testing.T) {
	t.Run("proportional fee above minimum fee", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			// 0.5% of 10M msat = 50k msat > 10k msat minimum
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
		}
		assert.Equal(t, uint64(50_000), computeLsps2MaxTotalOpeningFeeMsat(10_000_000, menu))
	})

	t.Run("minimum fee above proportional fee", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			// 0.5% of 1M msat = 5k msat < 10k msat minimum
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
		}
		assert.Equal(t, uint64(10_000), computeLsps2MaxTotalOpeningFeeMsat(1_000_000, menu))
	})

	t.Run("highest fee across menu entries", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
			makeLsps2OpeningFeeParams(10_000, 20_000, 1_000_000, 100_000_000),
		}
		assert.Equal(t, uint64(200_000), computeLsps2MaxTotalOpeningFeeMsat(10_000_000, menu))
	})

	t.Run("entries not covering the payment size are skipped", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
			// covers larger payments only, would otherwise win with 2%
			makeLsps2OpeningFeeParams(10_000, 20_000, 20_000_000, 100_000_000),
		}
		assert.Equal(t, uint64(50_000), computeLsps2MaxTotalOpeningFeeMsat(10_000_000, menu))
	})

	t.Run("menu fee above ceiling is clamped to ceiling", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			// 20% of 100M msat = 20M msat, above the 10% / 10M msat ceiling
			makeLsps2OpeningFeeParams(5_000_000, 200_000, 1_000_000, 1_000_000_000),
		}
		assert.Equal(t, uint64(10_000_000), computeLsps2MaxTotalOpeningFeeMsat(100_000_000, menu))
	})

	t.Run("base ceiling applies when no entry covers the payment size", func(t *testing.T) {
		menu := []ldk_node.Lsps2OpeningFeeParams{
			makeLsps2OpeningFeeParams(10_000, 5_000, 1_000_000, 100_000_000),
		}
		// 10% of 200M msat = 20M msat
		assert.Equal(t, uint64(20_000_000), computeLsps2MaxTotalOpeningFeeMsat(200_000_000, menu))
	})

	t.Run("base ceiling applies on empty menu", func(t *testing.T) {
		assert.Equal(t, uint64(5_000_000), computeLsps2MaxTotalOpeningFeeMsat(10_000_000, nil))
	})
}

func TestComputeLsps2MinPaymentSizeMsat(t *testing.T) {
	t.Run("minimum fee below ceiling base", func(t *testing.T) {
		// 1000 sat minimum fee: smallest usable payment nets 1 sat above the fee
		params := makeLsps2OpeningFeeParams(1_000_000, 10_000, 1_000, 100_000_000_000)
		minPaymentSizeMsat, ok := computeLsps2MinPaymentSizeMsat(params)
		require.True(t, ok)
		assert.Equal(t, uint64(1_001_000), minPaymentSizeMsat)
	})

	t.Run("minimum fee above ceiling base", func(t *testing.T) {
		// 8000 sat minimum fee exceeds the 5000 sat ceiling base, so the
		// smallest payment is where the fee equals 10% of the payment
		params := makeLsps2OpeningFeeParams(8_000_000, 10_000, 1_000, 100_000_000_000)
		minPaymentSizeMsat, ok := computeLsps2MinPaymentSizeMsat(params)
		require.True(t, ok)
		assert.Equal(t, uint64(80_000_000), minPaymentSizeMsat)
	})

	t.Run("proportional fee above ceiling percentage never fits", func(t *testing.T) {
		// 30% proportional fee with a minimum fee above the ceiling base can
		// never satisfy the 10% ceiling
		params := makeLsps2OpeningFeeParams(6_000_000, 300_000, 1_000, 1_000_000_000)
		_, ok := computeLsps2MinPaymentSizeMsat(params)
		assert.False(t, ok)
	})
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

// writeChannelBackup saves a static channel backup containing the given
// channels to the backup directory.
func writeChannelBackup(t *testing.T, backupDir string, fileName string, channelIds ...string) {
	backup := events.StaticChannelsBackupEvent{}
	for _, channelId := range channelIds {
		backup.Channels = append(backup.Channels, events.ChannelBackup{ChannelID: channelId})
	}
	backupBytes, err := json.Marshal(backup)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(backupDir, fileName), backupBytes, 0600))
}

func TestLatestChannelBackupIsMissingChannels(t *testing.T) {
	logger.Init(strconv.Itoa(int(logrus.DebugLevel)))

	t.Run("no channels", func(t *testing.T) {
		assert.False(t, latestChannelBackupIsMissingChannels(filepath.Join(t.TempDir(), "static_channel_backups"), []string{}))
	})

	t.Run("missing backup directory", func(t *testing.T) {
		assert.True(t, latestChannelBackupIsMissingChannels(filepath.Join(t.TempDir(), "static_channel_backups"), []string{"a"}))
	})

	t.Run("empty backup directory", func(t *testing.T) {
		assert.True(t, latestChannelBackupIsMissingChannels(t.TempDir(), []string{"a"}))
	})

	t.Run("only non-backup files", func(t *testing.T) {
		backupDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(backupDir, "notes.txt"), []byte("x"), 0600))
		assert.True(t, latestChannelBackupIsMissingChannels(backupDir, []string{"a"}))
	})

	t.Run("latest backup contains all channels", func(t *testing.T) {
		backupDir := t.TempDir()
		writeChannelBackup(t, backupDir, "2024-01-02T03-04-05.json", "a", "b", "closed")
		assert.False(t, latestChannelBackupIsMissingChannels(backupDir, []string{"a", "b"}))
	})

	t.Run("latest backup is missing a channel", func(t *testing.T) {
		backupDir := t.TempDir()
		writeChannelBackup(t, backupDir, "2024-01-02T03-04-05.json", "a")
		assert.True(t, latestChannelBackupIsMissingChannels(backupDir, []string{"a", "b"}))
	})

	t.Run("only the latest backup is checked", func(t *testing.T) {
		backupDir := t.TempDir()
		writeChannelBackup(t, backupDir, "2024-01-02T03-04-05.json", "a", "b")
		writeChannelBackup(t, backupDir, "2024-02-01T00-00-00.json", "a")
		assert.True(t, latestChannelBackupIsMissingChannels(backupDir, []string{"a", "b"}))

		writeChannelBackup(t, backupDir, "2024-03-01T00-00-00.json", "a", "b")
		assert.False(t, latestChannelBackupIsMissingChannels(backupDir, []string{"a", "b"}))
	})

	t.Run("invalid latest backup", func(t *testing.T) {
		backupDir := t.TempDir()
		require.NoError(t, os.WriteFile(filepath.Join(backupDir, "2024-01-02T03-04-05.json"), []byte("not json"), 0600))
		assert.True(t, latestChannelBackupIsMissingChannels(backupDir, []string{"a"}))
	})

	t.Run("glob characters in path", func(t *testing.T) {
		backupDir := filepath.Join(t.TempDir(), "work[dir")
		require.NoError(t, os.MkdirAll(backupDir, 0700))
		assert.True(t, latestChannelBackupIsMissingChannels(backupDir, []string{"a"}))

		writeChannelBackup(t, backupDir, "2024-01-02T03-04-05.json", "a")
		assert.False(t, latestChannelBackupIsMissingChannels(backupDir, []string{"a"}))
	})

	t.Run("unreadable backup path", func(t *testing.T) {
		// a file instead of a directory: neither missing nor listable
		backupPath := filepath.Join(t.TempDir(), "static_channel_backups")
		require.NoError(t, os.WriteFile(backupPath, []byte("x"), 0600))
		assert.False(t, latestChannelBackupIsMissingChannels(backupPath, []string{"a"}))
	})
}
