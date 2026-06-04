//go:build !((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64)) || (windows && amd64))

package bark

import (
	"context"
	"fmt"
	"runtime"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient"
)

// The bark FFI bindings only ship prebuilt native libraries for a subset of
// platforms (notably not 32-bit ARM Linux). On every other platform this stub
// is compiled instead of bark.go so the rest of Alby Hub still builds, and the
// bark backend simply reports that it is unavailable if selected at runtime.

// Config mirrors the real bark.Config so callers compile on all platforms.
type Config struct {
	Network           string
	ServerAddress     string
	EsploraAddress    string
	ServerAccessToken string
}

func NewBarkService(ctx context.Context, eventPublisher events.EventPublisher, workDir, mnemonic string, config Config) (lnclient.LNClient, error) {
	return nil, fmt.Errorf("the bark backend is not supported on %s/%s", runtime.GOOS, runtime.GOARCH)
}
