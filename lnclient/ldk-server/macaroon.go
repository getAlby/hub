package ldkserver

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path"
	"strings"
	"time"

	"gopkg.in/macaroon.v2"
)

// ldk-server authenticates each request with a v2 macaroon sent hex-encoded in the
// "macaroon" header. The configured macaroon is reusable and kept private; every request
// gets a copy with a final "request = <unix-seconds> <RpcMethod> <body-sha256>" caveat
// that binds it to the method, exact gRPC body and time.
// See ldk-server docs/request-binding.md.
const (
	macaroonHeader         = "macaroon"
	requestCaveatPrefix    = "request = "
	maxMacaroonBytes       = 4096
	maxMacaroonCaveats     = 32
	maxRequestMethodBytes  = 128
	maxRequestCaveatLength = len(requestCaveatPrefix) + 20 + 1 + maxRequestMethodBytes + 1 + 64
)

func parseReusableMacaroon(macaroonHex string) (*macaroon.Macaroon, error) {
	macaroonHex = strings.TrimSpace(macaroonHex)
	if len(macaroonHex) > maxMacaroonBytes*2 {
		return nil, errors.New("macaroon exceeds size limit")
	}
	macBytes, err := hex.DecodeString(macaroonHex)
	if err != nil {
		return nil, fmt.Errorf("macaroon must be hex-encoded: %w", err)
	}
	if len(macBytes) == 0 || macBytes[0] != 2 {
		return nil, errors.New("ldk-server requires a v2 macaroon")
	}
	mac := &macaroon.Macaroon{}
	if err := mac.UnmarshalBinary(macBytes); err != nil {
		return nil, fmt.Errorf("failed to parse macaroon: %w", err)
	}
	caveats := mac.Caveats()
	for _, caveat := range caveats {
		if caveat.VerificationId != nil || caveat.Location != "" {
			return nil, errors.New("ldk-server macaroons do not support third-party caveats")
		}
		if strings.HasPrefix(string(caveat.Id), requestCaveatPrefix) {
			return nil, errors.New("macaroon is already bound to a request; use the reusable macaroon (e.g. admin.macaroon)")
		}
	}
	// Reserve room for the per-request proof (one caveat, plus tag, 2 length bytes and end marker).
	if len(caveats) >= maxMacaroonCaveats || len(macBytes)+maxRequestCaveatLength+4 > maxMacaroonBytes {
		return nil, errors.New("macaroon has no room for a request binding caveat")
	}
	return mac, nil
}

// requestMethodName converts a full gRPC method path ("/api.LightningNode/GetNodeInfo")
// into the short RPC name ("GetNodeInfo") used in the request caveat.
func requestMethodName(fullMethod string) string {
	return path.Base(fullMethod)
}

func requestCaveat(method string, framedBody []byte, timestamp time.Time) string {
	bodyHash := sha256.Sum256(framedBody)
	return fmt.Sprintf("%s%d %s %s", requestCaveatPrefix, timestamp.Unix(), method, hex.EncodeToString(bodyHash[:]))
}

// bindMacaroonToRequest returns the hex-encoded token for a single request.
// framedBody must be the exact bytes sent: the 5-byte gRPC frame header followed by the protobuf.
func bindMacaroonToRequest(reusable *macaroon.Macaroon, fullMethod string, framedBody []byte, timestamp time.Time) (string, error) {
	mac := reusable.Clone()
	if err := mac.AddFirstPartyCaveat([]byte(requestCaveat(requestMethodName(fullMethod), framedBody, timestamp))); err != nil {
		return "", fmt.Errorf("failed to add request caveat: %w", err)
	}
	macBytes, err := mac.MarshalBinary()
	if err != nil {
		return "", fmt.Errorf("failed to serialize macaroon: %w", err)
	}
	return hex.EncodeToString(macBytes), nil
}
