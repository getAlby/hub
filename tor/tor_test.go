package tor

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"

	tored25519 "github.com/cretz/bine/torutil/ed25519"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/logger"
)

func init() {
	logger.Init("4")
}

func TestLoadOrCreateKey_GeneratesAndPersistsKey(t *testing.T) {
	tmpDir := t.TempDir()

	// First call should generate a new key
	key1, err := loadOrCreateKey(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, key1)

	// Key file should exist
	keyPath := filepath.Join(tmpDir, onionKeyFileName)
	_, err = os.Stat(keyPath)
	require.NoError(t, err, "key file should exist after generation")

	// Second call should load the same key
	key2, err := loadOrCreateKey(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, key2)

	// Both keys should produce the same blob (same key)
	assert.Equal(t, key1.Blob(), key2.Blob(), "loaded key should match generated key")
}

func TestLoadOrCreateKey_DifferentDirsProduceDifferentKeys(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	key1, err := loadOrCreateKey(tmpDir1)
	require.NoError(t, err)

	key2, err := loadOrCreateKey(tmpDir2)
	require.NoError(t, err)

	assert.NotEqual(t, key1.Blob(), key2.Blob(), "keys from different directories should be different")
}

func TestLoadOrCreateKey_CorruptKeyFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, onionKeyFileName)

	// Write corrupt data
	err := os.WriteFile(keyPath, []byte("not-a-valid-key-blob"), 0600)
	require.NoError(t, err)

	_, err = loadOrCreateKey(tmpDir)
	assert.Error(t, err, "should fail with corrupt key data")
	assert.Contains(t, err.Error(), "parse stored onion key")
}

func TestLoadOrCreateKey_EmptyKeyFile(t *testing.T) {
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, onionKeyFileName)

	// Write empty file — should trigger new key generation (len(data) > 0 check)
	err := os.WriteFile(keyPath, []byte{}, 0600)
	require.NoError(t, err)

	key, err := loadOrCreateKey(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, key, "should generate new key when file is empty")
}

func TestLoadOrCreateKey_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := loadOrCreateKey(tmpDir)
	require.NoError(t, err)

	keyPath := filepath.Join(tmpDir, onionKeyFileName)
	info, err := os.Stat(keyPath)
	require.NoError(t, err)

	// Key file should be owner-only read/write (0600)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm(), "key file should have 0600 permissions")
}

func TestLoadOrCreateKey_CreatesDataDir(t *testing.T) {
	tmpDir := t.TempDir()
	nestedDir := filepath.Join(tmpDir, "nested", "deep", "dir")

	key, err := loadOrCreateKey(nestedDir)
	require.NoError(t, err)
	require.NotNil(t, key)

	// Nested dirs should have been created
	_, err = os.Stat(nestedDir)
	require.NoError(t, err, "nested data directory should be created")
}

func TestNewService(t *testing.T) {
	svc := NewService()
	require.NotNil(t, svc)
	assert.False(t, svc.IsRunning(), "new service should not be running")
	assert.Empty(t, svc.GetOnionAddress(), "new service should have empty onion address")
}

func TestService_StopWhenNotRunning(t *testing.T) {
	svc := NewService()
	// Stopping a non-running service should not error
	err := svc.Stop()
	assert.NoError(t, err, "stop should not error when not running")
}

func TestService_StopClearsState(t *testing.T) {
	svc := NewService()

	// Manually set internal state to simulate a running service
	svc.mu.Lock()
	svc.running = true
	svc.serviceID = "test-service-id"
	svc.onionAddress = "testaddress.onion"
	svc.ctrl = nil // no real control connection
	svc.mu.Unlock()

	err := svc.Stop()
	assert.NoError(t, err)
	assert.False(t, svc.IsRunning(), "should not be running after stop")
	assert.Empty(t, svc.GetOnionAddress(), "onion address should be cleared after stop")
}

func TestService_StartAlreadyRunning(t *testing.T) {
	svc := NewService()

	// Simulate running state
	svc.mu.Lock()
	svc.running = true
	svc.mu.Unlock()

	err := svc.Start(&Config{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already running")
}

func TestService_StartBadAuth(t *testing.T) {
	// Spin up a TCP listener that accepts connections but doesn't speak Tor control protocol
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	// Accept one connection and close it immediately (simulates auth failure)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	svc := NewService()
	cfg := &Config{
		TorControlHost: "127.0.0.1",
		TorControlPort: addr.Port,
		DataDir:        t.TempDir(),
	}

	err = svc.Start(cfg)
	assert.Error(t, err, "should fail with bad auth")
	assert.False(t, svc.IsRunning())
}

func TestConfig_Defaults(t *testing.T) {
	cfg := &Config{}

	// These are used inside Start() — verify the zero-value defaults that Start applies
	assert.Equal(t, "", cfg.TorControlHost, "default control host should be empty (Start fills 127.0.0.1)")
	assert.Equal(t, 0, cfg.TorControlPort, "default control port should be 0 (Start fills 9051)")
	assert.Equal(t, 0, cfg.OnionServicePort, "default onion port should be 0 (Start fills 9735)")
	assert.Equal(t, "", cfg.TargetHost, "default target host should be empty (Start fills 127.0.0.1)")
}

func TestService_ConcurrentAccess(t *testing.T) {
	svc := NewService()

	// Simulate running state
	svc.mu.Lock()
	svc.running = true
	svc.onionAddress = "testconcurrent.onion"
	svc.mu.Unlock()

	var wg sync.WaitGroup
	// Concurrent reads should not race
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = svc.IsRunning()
		}()
		go func() {
			defer wg.Done()
			_ = svc.GetOnionAddress()
		}()
	}
	wg.Wait()
}

func TestPubkeyFingerprint_ValidKey(t *testing.T) {
	// 32-byte key should produce a non-empty base64 result
	pubKey := make(tored25519.PublicKey, 32)
	for i := range pubKey {
		pubKey[i] = byte(i)
	}
	result := pubkeyFingerprint(pubKey)
	assert.NotEmpty(t, result)

	// Should be valid base64
	_, err := base64.StdEncoding.DecodeString(result)
	assert.NoError(t, err, "result should be valid base64")
}

func TestPubkeyFingerprint_InvalidKeyLength(t *testing.T) {
	// Too short
	shortKey := make(tored25519.PublicKey, 16)
	assert.Empty(t, pubkeyFingerprint(shortKey), "should return empty for short key")

	// Too long
	longKey := make(tored25519.PublicKey, 64)
	assert.Empty(t, pubkeyFingerprint(longKey), "should return empty for long key")

	// Empty
	assert.Empty(t, pubkeyFingerprint(nil), "should return empty for nil key")
}

func TestPubkeyFingerprint_DifferentKeysProduceDifferentAddresses(t *testing.T) {
	key1 := make(tored25519.PublicKey, 32)
	key2 := make(tored25519.PublicKey, 32)
	rand.Read(key1)
	rand.Read(key2)

	addr1 := pubkeyFingerprint(key1)
	addr2 := pubkeyFingerprint(key2)
	assert.NotEqual(t, addr1, addr2, "different keys should produce different addresses")
}

func TestPubkeyFingerprint_Deterministic(t *testing.T) {
	key := make(tored25519.PublicKey, 32)
	for i := range key {
		key[i] = byte(i * 7)
	}

	result1 := pubkeyFingerprint(key)
	result2 := pubkeyFingerprint(key)
	assert.Equal(t, result1, result2, "same key should always produce same address")
}

func TestConstants(t *testing.T) {
	assert.Equal(t, 9051, DefaultTorControlPort)
	assert.Equal(t, 9735, DefaultOnionServicePort)
	assert.Equal(t, "onion_service_key", onionKeyFileName)
}

func TestGetOnionAddress_ReturnsSetValue(t *testing.T) {
	svc := NewService()

	svc.mu.Lock()
	svc.onionAddress = "abc123.onion"
	svc.mu.Unlock()

	assert.Equal(t, "abc123.onion", svc.GetOnionAddress())
}

func TestIsRunning_ReflectsState(t *testing.T) {
	svc := NewService()
	assert.False(t, svc.IsRunning())

	svc.mu.Lock()
	svc.running = true
	svc.mu.Unlock()
	assert.True(t, svc.IsRunning())

	svc.mu.Lock()
	svc.running = false
	svc.mu.Unlock()
	assert.False(t, svc.IsRunning())
}

func TestService_DoubleStop(t *testing.T) {
	svc := NewService()

	// Simulate running state
	svc.mu.Lock()
	svc.running = true
	svc.ctrl = nil
	svc.mu.Unlock()

	err := svc.Stop()
	assert.NoError(t, err)

	// Second stop should also be fine
	err = svc.Stop()
	assert.NoError(t, err)
}

func TestLoadOrCreateKey_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()

	// Generate key
	key1, err := loadOrCreateKey(tmpDir)
	require.NoError(t, err)
	blob1 := key1.Blob()

	// Load 10 more times — should always return the same key
	for i := 0; i < 10; i++ {
		key, err := loadOrCreateKey(tmpDir)
		require.NoError(t, err)
		assert.Equal(t, blob1, key.Blob(), fmt.Sprintf("iteration %d should return same key", i))
	}
}
