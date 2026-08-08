package greenlight

import (
	"context"
	"fmt"
	"time"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/lnclient/cln/clngrpc"
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

// Node health watchdog.
//
// Greenlight nodes run on Blockstream infrastructure: the hub cannot
// restart a wedged node the way it can restart an embedded LDK node. We
// proved (Blockstream/greenlight#739) that a signmessage call freezes the
// production VLS signer and wedges the hsmd queue until Blockstream
// intervenes — and a wedged node silently stops processing payments.
//
// The watchdog detects both failure classes without piling more RPCs onto
// a frozen node:
//   - reachability: a Getinfo with a short deadline catches a dead node or
//     a network split;
//   - invoice-path liveness: the WaitAnyInvoice pump is the incoming-money
//     path. A single pump call outstanding well past its 30s server-side
//     timeout means the node stopped processing (the hsmd wedge symptom).
//
// State is cached and served from GetNodeStatus, so the hub's node-status
// API keeps answering (with an "unhealthy" verdict) even when the node
// itself is frozen — a live Getinfo there would hang the API.

const (
	defaultHealthCheckInterval     = 30 * time.Second
	defaultHealthCheckTimeout      = 10 * time.Second
	defaultHealthFailureThreshold  = 3
	defaultPumpStallThreshold      = 90 * time.Second
	nodeHealthEvent                = "nwc_gl_node_health"
)

// nodeHealthStatus is what GetNodeStatus exposes in InternalNodeStatus
// (additive; the interface field is interface{}).
type nodeHealthStatus struct {
	Healthy     bool   `json:"healthy"`
	LastCheckAt int64  `json:"last_check_at"` // unix seconds
	LastError   string `json:"last_error,omitempty"`
}

func (g *GreenlightService) startHealthWatchdog(ctx context.Context) {
	go func() {
		// immediate first check so GetNodeStatus is meaningful right away
		g.runHealthCheck(ctx)
		ticker := time.NewTicker(g.healthCheckInterval)
		defer ticker.Stop()

		failures := 0
		wasHealthy := true
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				healthy, err := g.runHealthCheck(ctx)
				if healthy {
					failures = 0
					if !wasHealthy {
						logger.Logger.Info("greenlight node health restored")
						g.eventPublisher.Publish(&events.Event{
							Event:      nodeHealthEvent,
							Properties: map[string]interface{}{"healthy": true},
						})
						wasHealthy = true
					}
					continue
				}
				failures++
				if failures >= g.healthFailureThreshold && wasHealthy {
					logger.Logger.WithFields(logrus.Fields{
						"last_error": err,
						"failures":   failures,
					}).Error("greenlight node is unhealthy: RPC or invoice path stalled (possible frozen node — see runbook; Blockstream may need to restart it)")
					g.eventPublisher.Publish(&events.Event{
						Event: nodeHealthEvent,
						Properties: map[string]interface{}{
							"healthy": false,
							"error":   err,
						},
					})
					wasHealthy = false
				}
			}
		}
	}()
}

// runHealthCheck performs one health probe and caches the result. Returns
// (healthy, error-string-or-nil).
func (g *GreenlightService) runHealthCheck(ctx context.Context) (bool, error) {
	healthy, err := g.checkNodeHealth(ctx)

	g.healthMtx.Lock()
	g.nodeHealthy = healthy
	g.lastHealthCheck = time.Now()
	if err != nil {
		g.lastHealthError = err.Error()
	} else {
		g.lastHealthError = ""
	}
	g.healthMtx.Unlock()

	return healthy, err
}

func (g *GreenlightService) checkNodeHealth(ctx context.Context) (bool, error) {
	// defensive defaults for directly-constructed services (tests)
	timeout := g.healthCheckTimeout
	if timeout <= 0 {
		timeout = defaultHealthCheckTimeout
	}
	stallThreshold := g.pumpStallThreshold
	if stallThreshold <= 0 {
		stallThreshold = defaultPumpStallThreshold
	}

	// 1. reachability (catches node down / network split)
	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	info, err := g.client.Getinfo(checkCtx, &clngrpc.GetinfoRequest{})
	cancel()
	if err != nil {
		return false, fmt.Errorf("getinfo failed: %w", err)
	}

	g.healthMtx.Lock()
	g.lastInfo = info
	g.healthMtx.Unlock()

	// 2. invoice-path liveness: an outstanding pump call well past the 30s
	// server timeout means the node stopped processing (hsmd wedge)
	g.healthMtx.RLock()
	inCall := g.inPumpCall
	callStart := g.lastPumpCallStart
	g.healthMtx.RUnlock()
	if inCall && !callStart.IsZero() {
		stalled := time.Since(callStart)
		if stalled > stallThreshold {
			return false, fmt.Errorf("invoice pump stalled: waitanyinvoice outstanding for %s", stalled.Round(time.Second))
		}
	}
	return true, nil
}

func (g *GreenlightService) markPumpCallStart() {
	g.healthMtx.Lock()
	defer g.healthMtx.Unlock()
	g.lastPumpCallStart = time.Now()
	g.inPumpCall = true
}

func (g *GreenlightService) markPumpCallDone() {
	g.healthMtx.Lock()
	defer g.healthMtx.Unlock()
	g.inPumpCall = false
}
