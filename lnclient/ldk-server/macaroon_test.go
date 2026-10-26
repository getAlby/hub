package ldkserver

import (
	"encoding/hex"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gopkg.in/macaroon.v2"

	ldkapi "github.com/getAlby/hub/lnclient/ldk-server/grpc/api"
)

// Reference vectors from ldk-server-macaroons/tests/data/macaroons-v2.txt
// (root key "test-key", identifier "test-id").
const (
	vectorReusableMacaroon = "0201000207746573742d69640002177065726d697373696f6e73203d206e6f64653a7265616400000620fcd6db26ff34eca91e70ec1822ebb0e5ebb0ba3968cdb6544b900d9142a57659"
	vectorBoundMacaroon    = "0201000207746573742d69640002177065726d697373696f6e73203d206e6f64653a7265616400026172657175657374203d2031383030303030303030204765744e6f6465496e666f2038383535353038616164653136656335373364323165366134383564666430613736323430383563316131346235656364643634383564653063363833396134000006206553a4b4f42f654b01b195823008752c41d0cff8247b77fbb77e4860debcc7e2"
)

func TestBindMacaroonToRequest_ReferenceVector(t *testing.T) {
	mac, err := parseReusableMacaroon(vectorReusableMacaroon)
	require.NoError(t, err)

	emptyFrame := grpcFrame(nil)
	require.Equal(t, make([]byte, 5), emptyFrame)

	token, err := bindMacaroonToRequest(mac, ldkapi.LightningNode_GetNodeInfo_FullMethodName, emptyFrame, time.Unix(1800000000, 0))
	require.NoError(t, err)

	got := decodeMacaroon(t, token)
	want := decodeMacaroon(t, vectorBoundMacaroon)
	require.Equal(t, want.Id(), got.Id())
	require.Equal(t, want.Caveats(), got.Caveats())
	require.Equal(t, want.Signature(), got.Signature())
	require.Equal(t,
		"request = 1800000000 GetNodeInfo 8855508aade16ec573d21e6a485dfd0a7624085c1a14b5ecdd6485de0c6839a4",
		string(got.Caveats()[len(got.Caveats())-1].Id))

	// The reusable macaroon must not be modified.
	require.Len(t, mac.Caveats(), 1)
}

func TestParseReusableMacaroon_RejectsBoundMacaroon(t *testing.T) {
	_, err := parseReusableMacaroon(vectorBoundMacaroon)
	require.ErrorContains(t, err, "already bound")
}

func TestParseReusableMacaroon_RejectsInvalid(t *testing.T) {
	_, err := parseReusableMacaroon("not-hex")
	require.Error(t, err)

	_, err = parseReusableMacaroon("")
	require.Error(t, err)
}

func decodeMacaroon(t *testing.T, tokenHex string) *macaroon.Macaroon {
	t.Helper()
	b, err := hex.DecodeString(tokenHex)
	require.NoError(t, err)
	mac := &macaroon.Macaroon{}
	require.NoError(t, mac.UnmarshalBinary(b))
	return mac
}

func testMacaroon(t *testing.T) *macaroon.Macaroon {
	t.Helper()
	mac, err := parseReusableMacaroon(vectorReusableMacaroon)
	require.NoError(t, err)
	return mac
}

func parseRequestCaveatTimestamp(t *testing.T, caveat string) int64 {
	t.Helper()
	fields := strings.Fields(strings.TrimPrefix(caveat, requestCaveatPrefix))
	require.Len(t, fields, 3)
	ts, err := strconv.ParseInt(fields[0], 10, 64)
	require.NoError(t, err)
	return ts
}
