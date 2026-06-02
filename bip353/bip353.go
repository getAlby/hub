package bip353

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// dohResolverURL is a DNS-over-HTTPS resolver that performs DNSSEC validation
// and exposes the authenticated-data (AD) bit in its JSON response. BIP-353
// requires the payment instructions to be DNSSEC-signed, so we rely on the
// resolver to validate the chain and reject any answer that is not authenticated.
// It is a package variable so tests can point it at a mock server.
var dohResolverURL = "https://dns.google/resolve"

// dohResponse is the JSON shape returned by RFC 8484 JSON DoH resolvers
// (dns.google, cloudflare-dns.com).
type dohResponse struct {
	Status int  `json:"Status"`
	AD     bool `json:"AD"`
	Answer []struct {
		Type int    `json:"type"`
		Data string `json:"data"`
	} `json:"Answer"`
}

const dnsTypeTXT = 16

// LookupOffer resolves a BIP-353 human-readable address (e.g. "₿user@domain"
// or "user@domain") to the BOLT-12 offer (lno...) advertised in its DNSSEC-signed
// TXT record. It returns an error if the record is missing, not DNSSEC-validated,
// or does not contain a BOLT-12 offer.
func LookupOffer(ctx context.Context, address string) (string, error) {
	user, domain, err := parseAddress(address)
	if err != nil {
		return "", err
	}

	name := fmt.Sprintf("%s.user._bitcoin-payment.%s", user, domain)

	resp, err := query(ctx, name)
	if err != nil {
		return "", err
	}

	// Status 0 is NOERROR. Anything else (e.g. NXDOMAIN) means there is no
	// BIP-353 record for this address.
	if resp.Status != 0 {
		return "", errors.New("no BIP-353 payment record found for this address")
	}

	// BIP-353 mandates DNSSEC. Without a validated (authenticated) answer we
	// cannot trust the offer, so we refuse to pay it.
	if !resp.AD {
		return "", errors.New("BIP-353 record is not DNSSEC-validated")
	}

	uri := ""
	for _, answer := range resp.Answer {
		if answer.Type != dnsTypeTXT {
			continue
		}
		data := normalizeTXT(answer.Data)
		if strings.HasPrefix(strings.ToLower(data), "bitcoin:") {
			if uri != "" {
				// Per BIP-353 there must be exactly one bitcoin: URI.
				return "", errors.New("multiple BIP-353 payment records found for this address")
			}
			uri = data
		}
	}

	if uri == "" {
		return "", errors.New("no BIP-353 payment record found for this address")
	}

	return extractOffer(uri)
}

// parseAddress splits a BIP-353 address into its user and domain parts,
// stripping the optional ₿ (U+20BF) prefix.
func parseAddress(address string) (user string, domain string, err error) {
	address = strings.TrimSpace(address)
	address = strings.TrimPrefix(address, "₿")
	address = strings.TrimPrefix(address, "@") // tolerate a stray leading @

	parts := strings.Split(address, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", errors.New("invalid BIP-353 address")
	}

	// The user label is a single DNS label and must not contain dots.
	if strings.Contains(parts[0], ".") {
		return "", "", errors.New("invalid BIP-353 address")
	}

	return parts[0], parts[1], nil
}

func query(ctx context.Context, name string) (*dohResponse, error) {
	reqURL, err := url.Parse(dohResolverURL)
	if err != nil {
		return nil, err
	}
	q := reqURL.Query()
	q.Set("name", name)
	q.Set("type", "TXT")
	q.Set("do", "true") // request DNSSEC data
	reqURL.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/dns-json")

	httpResp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.New("failed to resolve BIP-353 address")
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DNS resolver returned status %d", httpResp.StatusCode)
	}

	var resp dohResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode DNS response: %w", err)
	}

	return &resp, nil
}

// normalizeTXT cleans up a TXT record value returned by a JSON DoH resolver.
// Long TXT records are split into multiple quoted character-strings which the
// resolver may return joined as `"foo" "bar"`; we strip the quotes and join them
// back into a single string.
func normalizeTXT(data string) string {
	if strings.HasPrefix(data, "\"") {
		data = strings.ReplaceAll(data, "\" \"", "")
		data = strings.Trim(data, "\"")
	}
	return data
}

// extractOffer parses a BIP-21 URI and returns its BOLT-12 offer (lno parameter).
func extractOffer(uri string) (string, error) {
	question := strings.Index(uri, "?")
	if question == -1 {
		return "", errors.New("BIP-353 address does not contain a BOLT-12 offer")
	}

	params, err := url.ParseQuery(uri[question+1:])
	if err != nil {
		return "", errors.New("failed to parse BIP-353 payment instructions")
	}

	offer := strings.TrimSpace(params.Get("lno"))
	if offer == "" {
		return "", errors.New("BIP-353 address does not contain a BOLT-12 offer")
	}

	return offer, nil
}
