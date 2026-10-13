package api

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/macaroon.v2"
)

// generateTestCert returns a self-signed certificate PEM block and its
// matching EC private key PEM block.
func generateTestCert(t *testing.T) (certPEM []byte, keyPEM []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	require.NoError(t, err)

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Unix(0, 0),
		NotAfter:     time.Unix(1<<31, 0),
	}
	der, err := x509.CreateCertificate(crand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})

	keyDER, err := x509.MarshalPKCS8PrivateKey(key)
	require.NoError(t, err)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER})

	return certPEM, keyPEM
}

func TestReadAndCanonicalizeLNDCert(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t)
	dir := t.TempDir()

	t.Run("valid certificate", func(t *testing.T) {
		path := filepath.Join(dir, "tls.cert")
		require.NoError(t, os.WriteFile(path, certPEM, 0600))

		got, err := readAndCanonicalizeLNDCert(path)
		require.NoError(t, err)

		raw, err := hex.DecodeString(got)
		require.NoError(t, err)
		require.True(t, x509.NewCertPool().AppendCertsFromPEM(raw))
	})

	t.Run("bundled private key is stripped", func(t *testing.T) {
		path := filepath.Join(dir, "bundle.pem")
		require.NoError(t, os.WriteFile(path, append(append([]byte{}, certPEM...), keyPEM...), 0600))

		got, err := readAndCanonicalizeLNDCert(path)
		require.NoError(t, err)

		raw, err := hex.DecodeString(got)
		require.NoError(t, err)
		// Only the CERTIFICATE block must survive - the private key must not
		// be persisted.
		require.NotContains(t, string(raw), "PRIVATE KEY")
		require.Contains(t, string(raw), "CERTIFICATE")
	})

	t.Run("arbitrary non-cert file is rejected", func(t *testing.T) {
		path := filepath.Join(dir, "secret.txt")
		require.NoError(t, os.WriteFile(path, []byte("root:x:0:0:root:/root:/bin/bash\n"), 0600))

		_, err := readAndCanonicalizeLNDCert(path)
		require.Error(t, err)
	})

	t.Run("missing file is rejected", func(t *testing.T) {
		_, err := readAndCanonicalizeLNDCert(filepath.Join(dir, "does-not-exist"))
		require.Error(t, err)
	})
}

func TestReadAndCanonicalizeLNDMacaroon(t *testing.T) {
	dir := t.TempDir()

	t.Run("valid macaroon", func(t *testing.T) {
		mac, err := macaroon.New([]byte("root-key"), []byte("id"), "location", macaroon.LatestVersion)
		require.NoError(t, err)
		raw, err := mac.MarshalBinary()
		require.NoError(t, err)

		path := filepath.Join(dir, "admin.macaroon")
		require.NoError(t, os.WriteFile(path, raw, 0600))

		got, err := readAndCanonicalizeLNDMacaroon(path)
		require.NoError(t, err)

		gotRaw, err := hex.DecodeString(got)
		require.NoError(t, err)
		roundTrip := &macaroon.Macaroon{}
		require.NoError(t, roundTrip.UnmarshalBinary(gotRaw))
	})

	t.Run("arbitrary non-macaroon file is rejected", func(t *testing.T) {
		path := filepath.Join(dir, "id_rsa")
		require.NoError(t, os.WriteFile(path, []byte("-----BEGIN OPENSSH PRIVATE KEY-----\nsecret\n"), 0600))

		_, err := readAndCanonicalizeLNDMacaroon(path)
		require.Error(t, err)
	})

	t.Run("missing file is rejected", func(t *testing.T) {
		_, err := readAndCanonicalizeLNDMacaroon(filepath.Join(dir, "does-not-exist"))
		require.Error(t, err)
	})
}

func TestValidateCLNLightningDir(t *testing.T) {
	certPEM, keyPEM := generateTestCert(t)

	writeCLNDir := func(t *testing.T, dir string) {
		t.Helper()
		require.NoError(t, os.WriteFile(filepath.Join(dir, "ca.pem"), certPEM, 0600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "client.pem"), certPEM, 0600))
		require.NoError(t, os.WriteFile(filepath.Join(dir, "client-key.pem"), keyPEM, 0600))
	}

	t.Run("valid directory", func(t *testing.T) {
		dir := t.TempDir()
		writeCLNDir(t, dir)
		require.NoError(t, validateCLNLightningDir(dir, false))
	})

	t.Run("valid directory with hold subdirectory", func(t *testing.T) {
		dir := t.TempDir()
		writeCLNDir(t, dir)
		holdDir := filepath.Join(dir, "hold")
		require.NoError(t, os.Mkdir(holdDir, 0700))
		writeCLNDir(t, holdDir)
		require.NoError(t, validateCLNLightningDir(dir, true))
	})

	t.Run("hold requested but subdirectory missing", func(t *testing.T) {
		dir := t.TempDir()
		writeCLNDir(t, dir)
		require.Error(t, validateCLNLightningDir(dir, true))
	})

	t.Run("arbitrary directory is rejected", func(t *testing.T) {
		require.Error(t, validateCLNLightningDir(t.TempDir(), false))
	})
}
