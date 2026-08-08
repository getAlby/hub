package greenlight

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/getAlby/hub/events"
)

func TestCheckNodeHealth_Healthy(t *testing.T) {
	node := newMockNode()
	svc, cleanup := newTestService(t, node)
	defer cleanup()

	healthy, err := svc.checkNodeHealth(svc.ctx)
	if err != nil {
		t.Fatalf("expected healthy node: %v", err)
	}
	if !healthy {
		t.Fatal("expected healthy=true")
	}
}

func TestCheckNodeHealth_GetinfoFailure(t *testing.T) {
	node := newMockNode()
	svc, cleanup := newTestService(t, node)
	defer cleanup()

	node.mu.Lock()
	node.getinfoErr = errors.New("node unreachable")
	node.mu.Unlock()

	healthy, err := svc.checkNodeHealth(svc.ctx)
	if healthy {
		t.Fatal("expected healthy=false")
	}
	if err == nil {
		t.Fatal("expected an error for an unreachable node")
	}
}

func TestCheckNodeHealth_PumpStalled(t *testing.T) {
	node := newMockNode()
	svc, cleanup := newTestService(t, node)
	defer cleanup()

	// simulate a WaitAnyInvoice call outstanding way past its 30s server
	// timeout — the hsmd-wedge symptom: node answers Getinfo but the
	// invoice path stopped processing
	svc.healthMtx.Lock()
	svc.inPumpCall = true
	svc.lastPumpCallStart = time.Now().Add(-3 * time.Minute)
	svc.healthMtx.Unlock()

	healthy, err := svc.checkNodeHealth(svc.ctx)
	if healthy {
		t.Fatal("expected healthy=false for a stalled invoice pump")
	}
	if err == nil {
		t.Fatal("expected a pump-stall error")
	}
}

func TestGetNodeStatus_ServesWatchdogCache(t *testing.T) {
	node := newMockNode()
	svc, cleanup := newTestService(t, node)
	defer cleanup()

	// healthy state
	healthy, err := svc.runHealthCheck(svc.ctx)
	if err != nil || !healthy {
		t.Fatalf("expected healthy check: %v", err)
	}

	status, err := svc.GetNodeStatus(svc.ctx)
	if err != nil {
		t.Fatalf("GetNodeStatus failed: %v", err)
	}
	if !status.IsReady {
		t.Fatal("expected IsReady=true for a healthy node")
	}
	health, ok := status.InternalNodeStatus.(nodeHealthStatus)
	if !ok || !health.Healthy {
		t.Fatalf("expected healthy internal status, got %#v", status.InternalNodeStatus)
	}

	// node goes down: the API must keep answering from the cache
	node.mu.Lock()
	node.getinfoErr = errors.New("node frozen")
	node.mu.Unlock()
	healthy, err = svc.runHealthCheck(svc.ctx)
	if healthy || err == nil {
		t.Fatal("expected the check to fail now")
	}

	status, err = svc.GetNodeStatus(svc.ctx)
	if err != nil {
		t.Fatalf("GetNodeStatus must not fail on a wedged node: %v", err)
	}
	health, ok = status.InternalNodeStatus.(nodeHealthStatus)
	if !ok || health.Healthy {
		t.Fatalf("expected unhealthy internal status, got %#v", status.InternalNodeStatus)
	}
	if health.LastError == "" {
		t.Fatal("expected the last error to be surfaced")
	}
}

func TestWatchdog_PublishesUnhealthyEvent(t *testing.T) {
	node := newMockNode()
	node.mu.Lock()
	node.getinfoErr = errors.New("node frozen")
	node.mu.Unlock()

	client, conn, cleanup := startMockNode(t, node)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer cleanup()

	capPub := &capturePublisher{}
	svc := &GreenlightService{
		ctx:                    ctx,
		cancel:                 cancel,
		client:                 client,
		conn:                   conn,
		eventPublisher:         capPub,
		config:                 Config{},
		workDir:                t.TempDir(),
		healthCheckInterval:    20 * time.Millisecond,
		healthCheckTimeout:     200 * time.Millisecond,
		healthFailureThreshold: 2,
		pumpStallThreshold:     90 * time.Second,
	}
	svc.startHealthWatchdog(ctx)

	ev := capPub.waitForEvent(nodeHealthEvent, 5*time.Second)
	if ev == nil {
		t.Fatal("expected the unhealthy node-health event")
	}
	props, ok := ev.Properties.(map[string]interface{})
	if !ok || props["healthy"] != false {
		t.Fatalf("expected unhealthy=true event properties, got %#v", ev.Properties)
	}

	// recovery: the mock heals, the watchdog must publish healthy again
	node.mu.Lock()
	node.getinfoErr = nil
	node.mu.Unlock()

	ev = capPub.waitForEvent(nodeHealthEvent, 5*time.Second)
	if ev == nil {
		t.Fatal("expected the healthy node-health event after recovery")
	}
	props, ok = ev.Properties.(map[string]interface{})
	if !ok || props["healthy"] != true {
		t.Fatalf("expected healthy=true event properties, got %#v", ev.Properties)
	}
}

var _ events.EventPublisher = (*capturePublisher)(nil)
