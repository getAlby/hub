package tor

import (
	"encoding/base64"
	"errors"
	"fmt"
	"net"
	"net/textproto"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cretz/bine/control"
	tored25519 "github.com/cretz/bine/torutil/ed25519"
	"github.com/sirupsen/logrus"

	"github.com/getAlby/hub/logger"
)

const (
	// DefaultTorControlPort is the default Tor control port
	DefaultTorControlPort = 9051
	// DefaultOnionServicePort is the port exposed on the .onion address
	DefaultOnionServicePort = 9735
	// onionKeyFileName is the file used to persist the onion service private key
	onionKeyFileName = "onion_service_key"
)

// Service manages a Tor hidden service for accepting incoming Lightning peer connections
type Service struct {
	mu           sync.Mutex
	ctrl         *control.Conn
	serviceID    string
	onionAddress string
	running      bool
}

// Config holds the Tor hidden service configuration
type Config struct {
	// TorControlHost is the hostname of the Tor control port (default: 127.0.0.1)
	TorControlHost string
	// TorControlPort is the port number of the Tor control port (default: 9051)
	TorControlPort int
	// TorControlPassword is the optional password for the Tor control port
	TorControlPassword string
	// TargetHost is the host where LDK listens for peer connections (default: 127.0.0.1).
	// In Docker setups where Tor runs in a separate container, this should be the
	// server container's hostname (e.g. the container name on the Docker network).
	TargetHost string
	// LocalPort is the local port that LDK listens on for peer connections
	LocalPort int
	// OnionServicePort is the port to expose on the .onion address (default: 9735)
	OnionServicePort int
	// DataDir is the directory to store the onion service private key
	DataDir string
}

// NewService creates a new Tor hidden service manager
func NewService() *Service {
	return &Service{}
}

// Start creates or restores a Tor v3 onion service that forwards connections
// to the local LDK peer listener. It connects to an existing Tor daemon via
// the control port (as is the case on Umbrel where Tor is a separate service).
func (s *Service) Start(cfg *Config) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return errors.New("tor service is already running")
	}
	s.mu.Unlock()

	controlHost := cfg.TorControlHost
	if controlHost == "" {
		controlHost = "127.0.0.1"
	}
	controlPort := cfg.TorControlPort
	if controlPort == 0 {
		controlPort = DefaultTorControlPort
	}
	onionPort := cfg.OnionServicePort
	if onionPort == 0 {
		onionPort = DefaultOnionServicePort
	}

	controlAddr := fmt.Sprintf("%s:%d", controlHost, controlPort)

	logger.Logger.WithFields(logrus.Fields{
		"control_addr": controlAddr,
		"local_port":   cfg.LocalPort,
		"onion_port":   onionPort,
	}).Info("Connecting to Tor control port")

	// Connect to the Tor control port with retries (Tor may still be bootstrapping)
	var conn net.Conn
	var err error
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		conn, err = net.DialTimeout("tcp", controlAddr, 5*time.Second)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			logger.Logger.WithFields(logrus.Fields{
				"attempt":      i + 1,
				"max_retries":  maxRetries,
				"control_addr": controlAddr,
			}).Debug("Tor control port not ready, retrying...")
			time.Sleep(2 * time.Second)
		}
	}
	if err != nil {
		return fmt.Errorf("failed to connect to Tor control port at %s after %d attempts: %w", controlAddr, maxRetries, err)
	}

	ctrl := control.NewConn(textproto.NewConn(conn))

	// Authenticate with the control port
	if err := ctrl.Authenticate(cfg.TorControlPassword); err != nil {
		ctrl.Close()
		return fmt.Errorf("failed to authenticate with Tor control port: %w", err)
	}

	logger.Logger.Info("Authenticated with Tor control port")

	// Load or generate the onion service ed25519 key
	key, err := loadOrCreateKey(cfg.DataDir)
	if err != nil {
		ctrl.Close()
		return fmt.Errorf("failed to load/create onion key: %w", err)
	}

	// Create the onion service via ADD_ONION
	targetHost := cfg.TargetHost
	if targetHost == "" {
		targetHost = "127.0.0.1"
	}
	targetAddr := fmt.Sprintf("%s:%d", targetHost, cfg.LocalPort)
	resp, err := ctrl.AddOnion(&control.AddOnionRequest{
		Key: key,
		Ports: []*control.KeyVal{
			{Key: fmt.Sprintf("%d", onionPort), Val: targetAddr},
		},
	})
	if err != nil {
		ctrl.Close()
		return fmt.Errorf("failed to create onion service: %w", err)
	}

	onionAddr := resp.ServiceID + ".onion"
	logger.Logger.WithField("onion_address", onionAddr).Info("Tor onion service is active")

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		// Another goroutine started concurrently — clean up orphaned onion service
		_ = ctrl.DelOnion(resp.ServiceID)
		ctrl.Close()
		return errors.New("tor service was started concurrently")
	}
	s.ctrl = ctrl
	s.serviceID = resp.ServiceID
	s.onionAddress = onionAddr
	s.running = true

	return nil
}

// Stop shuts down the Tor onion service and closes the control port connection
func (s *Service) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return nil
	}

	logger.Logger.Info("Stopping Tor onion service")

	var errs []error

	if s.ctrl != nil && s.serviceID != "" {
		if err := s.ctrl.DelOnion(s.serviceID); err != nil {
			errs = append(errs, fmt.Errorf("failed to remove onion service: %w", err))
		}
	}

	if s.ctrl != nil {
		if err := s.ctrl.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close control connection: %w", err))
		}
		s.ctrl = nil
	}

	s.running = false
	s.serviceID = ""
	s.onionAddress = ""

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	logger.Logger.Info("Tor onion service stopped")
	return nil
}

// GetOnionAddress returns the .onion address if the service is running, or empty string
func (s *Service) GetOnionAddress() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.onionAddress
}

// IsRunning returns whether the Tor service is currently running
func (s *Service) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

// loadOrCreateKey loads an existing onion service ed25519 key from disk, or generates
// and persists a new one. Returns a control.Key suitable for AddOnion.
func loadOrCreateKey(dataDir string) (control.Key, error) {
	keyPath := filepath.Join(dataDir, onionKeyFileName)

	// Try to load existing key
	if data, err := os.ReadFile(keyPath); err == nil && len(data) > 0 {
		key, err := control.KeyFromString("ED25519-V3:" + string(data))
		if err != nil {
			return nil, fmt.Errorf("failed to parse stored onion key: %w", err)
		}
		logger.Logger.Info("Loaded existing Tor onion service key")
		return key, nil
	}

	// Generate new key pair
	keyPair, err := tored25519.GenerateKey(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate ed25519 key: %w", err)
	}

	// Serialize for Tor control protocol: the expanded private key as base64
	edKey := &control.ED25519Key{KeyPair: keyPair}
	blob := edKey.Blob()

	// Persist the blob
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create data dir: %w", err)
	}

	if err := os.WriteFile(keyPath, []byte(blob), 0600); err != nil {
		return nil, fmt.Errorf("failed to persist onion key: %w", err)
	}

	// Log the generated onion address
	pubKey := keyPair.PublicKey()
	onionAddr := pubkeyFingerprint(pubKey)
	if onionAddr != "" {
		logger.Logger.WithField("onion_address", onionAddr+".onion").Info("Generated new Tor onion service key")
	}

	return edKey, nil
}

// pubkeyFingerprint returns a truncated base64 fingerprint of a public key for logging.
// The real onion address comes from the Tor control port response (resp.ServiceID).
func pubkeyFingerprint(pubKey tored25519.PublicKey) string {
	if len(pubKey) != 32 {
		return ""
	}
	// V3 onion address = base32(pubkey || checksum || version)
	// This is a simplified derivation; the actual address is returned by Tor
	// in the AddOnion response (ServiceID), so this is just for logging
	return base64.StdEncoding.EncodeToString(pubKey[:8])
}
