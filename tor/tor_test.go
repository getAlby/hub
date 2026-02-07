package tor

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getAlby/hub/logger"
)

func init() {
	logger.Init("4")
}

func TestLoadOrCreateKey_GeneratesAndPersistsKey(t *testing.T) {
	tmpDir := t.TempDir()

	// First call should generate a new key
	key1, err := loadOrCreateKey(tmpDir)
	if err != nil {
		t.Fatalf("failed to create key: %v", err)
	}
	if key1 == nil {
		t.Fatal("expected non-nil key")
	}

	// Key file should exist
	keyPath := filepath.Join(tmpDir, onionKeyFileName)
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		t.Fatal("key file should exist after generation")
	}

	// Second call should load the same key
	key2, err := loadOrCreateKey(tmpDir)
	if err != nil {
		t.Fatalf("failed to load key: %v", err)
	}
	if key2 == nil {
		t.Fatal("expected non-nil key on reload")
	}

	// Both keys should produce the same blob (same key)
	if key1.Blob() != key2.Blob() {
		t.Error("loaded key should match generated key")
	}
}

func TestLoadOrCreateKey_DifferentDirsProduceDifferentKeys(t *testing.T) {
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	key1, err := loadOrCreateKey(tmpDir1)
	if err != nil {
		t.Fatalf("failed to create key1: %v", err)
	}

	key2, err := loadOrCreateKey(tmpDir2)
	if err != nil {
		t.Fatalf("failed to create key2: %v", err)
	}

	if key1.Blob() == key2.Blob() {
		t.Error("keys from different directories should be different")
	}
}

func TestNewService(t *testing.T) {
	svc := NewService()
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
	if svc.IsRunning() {
		t.Error("new service should not be running")
	}
	if svc.GetOnionAddress() != "" {
		t.Error("new service should have empty onion address")
	}
}

func TestService_StopWhenNotRunning(t *testing.T) {
	svc := NewService()
	// Stopping a non-running service should not error
	if err := svc.Stop(); err != nil {
		t.Errorf("stop should not error when not running: %v", err)
	}
}
