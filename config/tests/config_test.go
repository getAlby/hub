package test

import (
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/getAlby/hub/config"
	"github.com/getAlby/hub/tests"
)

func TestCheckUnlockPasswordCache_InvalidSecond(t *testing.T) {
	unlockPassword := "123"
	svc, err := tests.CreateTestServiceWithMnemonic(t, "", unlockPassword)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.ChangeUnlockPassword("", unlockPassword)
	require.NoError(t, err)
	err = svc.Cfg.SaveUnlockPasswordCheck(unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("UnlockPasswordCheck", unlockPassword)
	require.NoError(t, err)
	require.Equal(t, "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT", value)

	assert.True(t, svc.Cfg.CheckUnlockPassword(unlockPassword))
	assert.False(t, svc.Cfg.CheckUnlockPassword(unlockPassword+"1"))
	assert.False(t, svc.Cfg.CheckUnlockPassword(""))
}
func TestCheckUnlockPasswordCache_InvalidFirst(t *testing.T) {
	unlockPassword := "123"
	svc, err := tests.CreateTestServiceWithMnemonic(t, "", unlockPassword)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.ChangeUnlockPassword("", unlockPassword)
	require.NoError(t, err)
	err = svc.Cfg.SaveUnlockPasswordCheck(unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("UnlockPasswordCheck", unlockPassword)
	require.NoError(t, err)
	require.Equal(t, "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT", value)

	value, err = svc.Cfg.Get("UnlockPasswordCheck", unlockPassword+"1")
	require.Error(t, err)

	assert.False(t, svc.Cfg.CheckUnlockPassword(""))
	assert.False(t, svc.Cfg.CheckUnlockPassword(unlockPassword+"1"))
	assert.True(t, svc.Cfg.CheckUnlockPassword(unlockPassword))
}

func TestCheckUnlockPassword_ChangePassword(t *testing.T) {
	unlockPassword := "123"
	svc, err := tests.CreateTestServiceWithMnemonic(t, "", unlockPassword)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.ChangeUnlockPassword("", unlockPassword)
	require.NoError(t, err)
	err = svc.Cfg.SaveUnlockPasswordCheck(unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("UnlockPasswordCheck", unlockPassword)
	require.NoError(t, err)
	require.Equal(t, "THIS STRING SHOULD MATCH IF PASSWORD IS CORRECT", value)

	assert.True(t, svc.Cfg.CheckUnlockPassword(unlockPassword))

	newUnlockPassword := "1234"

	err = svc.Cfg.ChangeUnlockPassword(unlockPassword, newUnlockPassword)
	require.NoError(t, err)

	assert.False(t, svc.Cfg.CheckUnlockPassword(unlockPassword))
	assert.True(t, svc.Cfg.CheckUnlockPassword(newUnlockPassword))
	// test caching
	assert.False(t, svc.Cfg.CheckUnlockPassword(unlockPassword))
	assert.True(t, svc.Cfg.CheckUnlockPassword(newUnlockPassword))
}

func TestSetIgnore_NoEncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.SetIgnore("key", "value", "")
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value", value)

	err = svc.Cfg.SetIgnore("key", "value2", "")
	require.NoError(t, err)

	// value should not be updated
	updatedValue, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value", updatedValue)
}

func TestSetIgnore_EncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	unlockPassword := "123"

	err = svc.Cfg.SetIgnore("key", "value", unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value", value)

	invalidValue, err := svc.Cfg.Get("key", unlockPassword+"1")
	assert.Error(t, err)
	assert.Equal(t, "", invalidValue)

	err = svc.Cfg.SetIgnore("key", "value2", unlockPassword)
	require.NoError(t, err)

	// value should not be updated
	updatedValue, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value", updatedValue)
}

func TestSetUpdate_NoEncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.SetUpdate("key", "value", "")
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value", value)

	err = svc.Cfg.SetUpdate("key", "value2", "")
	require.NoError(t, err)

	// value should be updated
	updatedValue, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value2", updatedValue)
}

func TestSetUpdate_EncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	unlockPassword := "123"

	err = svc.Cfg.SetUpdate("key", "value", unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value", value)

	invalidValue, err := svc.Cfg.Get("key", unlockPassword+"1")
	assert.Error(t, err)
	assert.Equal(t, "", invalidValue)

	err = svc.Cfg.SetUpdate("key", "value2", unlockPassword)
	require.NoError(t, err)

	// value should be updated
	updatedValue, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value2", updatedValue)
}

func TestSetUpdate_NoEncryptionKeyToEncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	err = svc.Cfg.SetUpdate("key", "value", "")
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value", value)

	unlockPassword := "123"

	err = svc.Cfg.SetUpdate("key", "value2", unlockPassword)
	require.NoError(t, err)

	// value should be updated
	updatedValue, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value2", updatedValue)
}

func TestSetUpdate_EncryptionKeyToNoEncryptionKey(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	unlockPassword := "123"

	err = svc.Cfg.SetUpdate("key", "value", unlockPassword)
	require.NoError(t, err)

	value, err := svc.Cfg.Get("key", unlockPassword)
	assert.Equal(t, "value", value)

	err = svc.Cfg.SetUpdate("key", "value2", "")
	require.NoError(t, err)

	// value should be updated
	updatedValue, err := svc.Cfg.Get("key", "")
	assert.Equal(t, "value2", updatedValue)
}

func TestJWTSecret_GeneratedOnUnlock(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	cfg, err := config.NewConfig(&config.AppConfig{}, svc.DB)
	require.NoError(t, err)

	err = cfg.ChangeUnlockPassword("", "123")
	require.NoError(t, err)

	err = cfg.Unlock("123")
	require.NoError(t, err)

	jwtSecret, err := cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, jwtSecret)

	encryptedSecret, err := cfg.Get("JWTSecret", "")
	require.NoError(t, err)
	decryptedSecret, err := cfg.Get("JWTSecret", "123")
	require.NoError(t, err)
	assert.NotEqual(t, encryptedSecret, decryptedSecret)

	// unlock again without doing anything, ensure the same JWT secret is returned
	err = cfg.Unlock("123")
	require.NoError(t, err)
	jwtSecret2, err := cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.Equal(t, jwtSecret, jwtSecret2)
}

func TestJWTSecret_ChangePassword(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	cfg, err := config.NewConfig(&config.AppConfig{}, svc.DB)
	require.NoError(t, err)

	err = cfg.ChangeUnlockPassword("", "123")
	require.NoError(t, err)

	err = cfg.Unlock("123")
	require.NoError(t, err)

	jwtSecret, err := cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, jwtSecret)

	err = cfg.ChangeUnlockPassword("", "1234")
	require.NoError(t, err)

	newJwtSecret, err := cfg.GetJWTSecret()
	require.ErrorContains(t, err, "unlock")

	err = cfg.Unlock("1234")
	require.NoError(t, err)

	// a new JWT secret must be generated after password change
	newJwtSecret, err = cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, newJwtSecret)
	assert.NotEqual(t, newJwtSecret, jwtSecret)
}

func TestJWTSecret_ReplaceUnencryptedSecretOnUnlock(t *testing.T) {
	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	cfg, err := config.NewConfig(&config.AppConfig{}, svc.DB)
	require.NoError(t, err)

	err = cfg.ChangeUnlockPassword("", "123")
	require.NoError(t, err)

	// simulate a hub that had an unencrypted JWT secret
	oldJwtSecret := "dummy secret"
	err = svc.Cfg.SetUpdate("JWTSecret", oldJwtSecret, "")
	require.NoError(t, err)

	err = cfg.Unlock("123")
	require.NoError(t, err)

	jwtSecret, err := cfg.GetJWTSecret()
	require.NoError(t, err)
	assert.NotEmpty(t, jwtSecret)
	assert.NotEqual(t, jwtSecret, oldJwtSecret)

	// ensure it is saved to DB
	jwtSecretFromCfg, err := cfg.Get("JWTSecret", "123")
	require.NoError(t, err)
	assert.Equal(t, jwtSecret, jwtSecretFromCfg)
}

func TestValidateChainSource(t *testing.T) {
	// Setup a Mock Esplora Server (Returns 200 OK)
	esploraServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer esploraServer.Close()

	// Setup a Mock "Bad" Server (Returns 500 Error)
	badServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer badServer.Close()

	// Setup a Mock Electrum Server (TCP Listener)
	// Listen on a random available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()
	validElectrumAddr := listener.Addr().String()

	svc, err := tests.CreateTestService(t)
	require.NoError(t, err)
	defer svc.Remove()

	tests := []struct {
		name          string
		backendType   string
		url           string
		shouldError   bool
		errorContains string // substring to check in error message
	}{
		// --- Esplora Cases ---
		{
			name:        "Esplora - Valid HTTP",
			backendType: "esplora",
			url:         esploraServer.URL,
			shouldError: false,
		},
		{
			name:        "Esplora - Valid HTTPS (Mocked)",
			backendType: "esplora",
			url:         esploraServer.URL,
			shouldError: false,
		},
		{
			name:          "Esplora - Invalid Prefix",
			backendType:   "esplora",
			url:           "ftp://example.com",
			shouldError:   true,
			errorContains: "must start with http:// or https://",
		},
		{
			name:          "Esplora - Connection Refused",
			backendType:   "esplora",
			url:           "http://127.0.0.1:54321", // assuming this port is closed
			shouldError:   true,
			errorContains: "failed to connect",
		},
		{
			name:          "Esplora - Server Error (Non-200)",
			backendType:   "esplora",
			url:           badServer.URL,
			shouldError:   true,
			errorContains: "server returned error code 500",
		},

		// --- Electrum Cases ---
		{
			name:        "Electrum - Valid TCP",
			backendType: "electrum",
			url:         "tcp://" + validElectrumAddr,
			shouldError: false,
		},
		{
			name:        "Electrum - Valid SSL (Prefix check only)",
			backendType: "electrum",
			// We use the same listener. logic trims 'ssl://' and dials.
			// Since our mock is just a raw TCP listener, it will accept the connection even if we call it "ssl".
			url:         "ssl://" + validElectrumAddr,
			shouldError: false,
		},
		{
			name:          "Electrum - Invalid Prefix",
			backendType:   "electrum",
			url:           "http://example.com",
			shouldError:   true,
			errorContains: "must start with ssl:// or tcp://",
		},
		{
			name:          "Electrum - Missing Port",
			backendType:   "electrum",
			url:           "tcp://127.0.0.1",
			shouldError:   true,
			errorContains: "missing a port number",
		},
		{
			name:          "Electrum - Connection Refused",
			backendType:   "electrum",
			url:           "tcp://127.0.0.1:54321", // assuming this port is closed
			shouldError:   true,
			errorContains: "could not connect",
		},

		// --- General Cases ---
		{
			name:          "Unsupported Backend",
			backendType:   "unknown_backend",
			url:           "http://localhost",
			shouldError:   true,
			errorContains: "unsupported backend type",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := svc.Cfg.ValidateChainSource(tc.backendType, tc.url)

			if tc.shouldError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.errorContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
