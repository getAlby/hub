package greenlight

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/getAlby/hub/logger"
	"github.com/tyler-smith/go-bip39"
)

//go:embed extract_creds.py
var embeddedExtractCreds string

const (
	seedFileName        = "hsm_secret"
	credentialsFileName = "credentials.gfs"
	deviceCredsDirName  = "device-creds"
)

// sanitizeOutput strips the mnemonic from glcli output (e.g. "New seed
// generated from mnemonic: ...") before logging or returning in errors.
var reMnemonic = regexp.MustCompile(`mnemonic:\s*\\S.*`) // rest of mnemonic line

// MnemonicToSeed32 matches gl-cli: mnemonic.to_seed("")[0..32], 12 words only.
func MnemonicToSeed32(mnemonic string) ([]byte, error) {
	mnemonic = strings.TrimSpace(mnemonic)
	if !bip39.IsMnemonicValid(mnemonic) {
		return nil, fmt.Errorf("invalid bip39 mnemonic")
	}
	if words := strings.Fields(mnemonic); len(words) != 12 {
		return nil, fmt.Errorf("greenlight requires a 12-word mnemonic, got %d", len(words))
	}
	full := bip39.NewSeed(mnemonic, "")
	seed := make([]byte, 32)
	copy(seed, full[:32])
	return seed, nil
}

// WriteSeedFile writes raw 32-byte seed as hsm_secret (0600).
func WriteSeedFile(dataDir string, seed []byte) error {
	if len(seed) != 32 {
		return fmt.Errorf("seed must be 32 bytes, got %d", len(seed))
	}
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dataDir, seedFileName), seed, 0o600)
}

func resolveGlcli(glcliPath string) (string, error) {
	if glcliPath == "" {
		glcliPath = "glcli"
	}
	if bin, err := exec.LookPath(glcliPath); err == nil {
		return bin, nil
	}
	for _, p := range []string{"/root/.cargo/bin/glcli", "/usr/local/bin/glcli"} {
		if _, e := os.Stat(p); e == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("glcli not found (set GREENLIGHT_GLCLI_PATH)")
}

// ensureSeedFile writes the seed derived from the mnemonic ONLY when no
// hsm_secret exists yet. The existing file is authoritative: the node's
// identity derives from it, and overwriting it with a different mnemonic
// would silently desync the signer from the node (every signing request
// would be rejected). Recovery works because a fresh data dir has no file.
func ensureSeedFile(dataDir string, seed []byte) error {
	seedPath := filepath.Join(dataDir, seedFileName)
	if _, err := os.Stat(seedPath); err == nil {
		return nil
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("stat hsm_secret: %w", err)
	}
	return WriteSeedFile(dataDir, seed)
}

// EnsureProvisioned writes seed from mnemonic, registers/recovers via glcli if needed,
// extracts device PEMs, and returns (deviceCredsDir, nodeURI).
func EnsureProvisioned(dataDir, network, glcliPath, nobodyCrt, nobodyKey, mnemonic, extractScript string) (credsDir, nodeURI string, err error) {
	if err := os.MkdirAll(dataDir, 0o700); err != nil {
		return "", "", err
	}

	seed, err := MnemonicToSeed32(mnemonic)
	if err != nil {
		return "", "", err
	}
	if err := ensureSeedFile(dataDir, seed); err != nil {
		return "", "", fmt.Errorf("write hsm_secret: %w", err)
	}

	credsBlob := filepath.Join(dataDir, credentialsFileName)
	needRegister := true
	if st, err := os.Stat(credsBlob); err == nil && st.Size() > 0 {
		needRegister = false
	}

	bin, err := resolveGlcli(glcliPath)
	if err != nil {
		return "", "", err
	}

	if nobodyCrt == "" {
		nobodyCrt = os.Getenv("GL_NOBODY_CRT")
	}
	if nobodyKey == "" {
		nobodyKey = os.Getenv("GL_NOBODY_KEY")
	}

	run := func(args ...string) (string, error) {
		cmd := exec.Command(bin, args...)
		env := os.Environ()
		if nobodyCrt != "" {
			// guard against env-injection: reject paths with newlines
			if strings.ContainsAny(nobodyCrt, "\n\r") || strings.ContainsAny(nobodyKey, "\n\r") {
				return "", fmt.Errorf("GL_NOBODY_CRT/KEY must not contain newlines")
			}
			env = append(env, "GL_NOBODY_CRT="+nobodyCrt)
		}
		if nobodyKey != "" {
			env = append(env, "GL_NOBODY_KEY="+nobodyKey)
		}
		cmd.Env = env
		out, err := cmd.CombinedOutput()
		s := string(out)
		// strip seed/mnemonic from output before logging or including in errors
		sanitized := reMnemonic.ReplaceAllString(s, "mnemonic: [redacted]")
		if err != nil {
			return s, fmt.Errorf("%w: %s", err, strings.TrimSpace(sanitized))
		}
		return s, nil
	}

	if needRegister {
		if nobodyCrt == "" || nobodyKey == "" {
			return "", "", fmt.Errorf("GL_NOBODY_CRT and GL_NOBODY_KEY required to register a greenlight node")
		}
		logger.Logger.Info("Registering greenlight node via glcli")
		out, regErr := run("-d", dataDir, "-n", network, "scheduler", "register")
		if regErr != nil {
			logger.Logger.WithError(regErr).Warn("register failed, trying recover")
			out2, recErr := run("-d", dataDir, "-n", network, "scheduler", "recover")
			if recErr != nil {
				return "", "", fmt.Errorf("register failed (%v); recover failed: %w\nregister: %s\nrecover: %s", regErr, recErr, out, out2)
			}
			logger.Logger.Info("Recovered greenlight device credentials")
		} else {
			logger.Logger.WithField("out", strings.TrimSpace(out)).Info("Registered greenlight node")
		}
	}

	if _, err := os.Stat(credsBlob); err != nil {
		return "", "", fmt.Errorf("credentials.gfs missing in %s", dataDir)
	}

	// schedule (wake)
	if _, err := run("-d", dataDir, "-n", network, "scheduler", "schedule"); err != nil {
		logger.Logger.WithError(err).Warn("glcli schedule failed (node domain may still wake on connect)")
	}

	credsDir = filepath.Join(dataDir, deviceCredsDirName)
	if extractScript != "" {
		if _, err := os.Stat(extractScript); err != nil {
			// caller-supplied path missing: fall back to the embedded copy
			logger.Logger.WithField("path", extractScript).Warn("extract_creds.py path not found, using embedded copy")
			extractScript = ""
		}
	}
	if extractScript == "" {
		// deployment-independent: write the embedded script next to the node
		// data (no CWD or machine-specific path assumptions).
		extractScript = filepath.Join(dataDir, "extract_creds.py")
		if err := os.WriteFile(extractScript, []byte(embeddedExtractCreds), 0o600); err != nil {
			return "", "", fmt.Errorf("write embedded extract_creds: %w", err)
		}
	}
	// resolve python3 via LookPath (not plain "python3" from PATH) for exec
	python3, err := exec.LookPath("python3")
	if err != nil {
		return "", "", fmt.Errorf("python3 not found in PATH: %w", err)
	}
	cmd := exec.Command(python3, extractScript, credsBlob, credsDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", fmt.Errorf("extract_creds failed: %w: %s", err, string(out))
	}
	logger.Logger.Info(strings.TrimSpace(string(out)))

	// parse GREENLIGHT_NODE_URI=... from extract output
	re := regexp.MustCompile(`GREENLIGHT_NODE_URI=(\S+)`)
	if m := re.FindSubmatch(out); len(m) == 2 {
		nodeURI = string(m[1])
	}
	if nodeURI == "" {
		re2 := regexp.MustCompile(`node domain:\s*(\S+)`)
		if m := re2.FindSubmatch(out); len(m) == 2 {
			nodeURI = string(m[1]) + ":443"
		}
	}
	if nodeURI == "" {
		return "", "", fmt.Errorf("could not derive node URI from extract_creds output")
	}

	// Required PEMs
	for _, f := range []string{"ca.pem", "client.pem", "client-key.pem"} {
		if _, err := os.Stat(filepath.Join(credsDir, f)); err != nil {
			return "", "", fmt.Errorf("missing %s after extract", f)
		}
	}
	return credsDir, nodeURI, nil
}
