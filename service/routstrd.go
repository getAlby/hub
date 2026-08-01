package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3"
	decodepay "github.com/nbd-wtf/ln-decodepay"

	"github.com/getAlby/hub/db"
	"github.com/getAlby/hub/db/queries"
	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"gorm.io/datatypes"
)

// RoutstrdService manages the routstrd daemon and its cocod Cashu wallet
// as child processes of Alby Hub. It starts them with the Hub, restarts on
// crash, and stops them gracefully on Hub shutdown.
type RoutstrdService struct {
	svc       *service
	cancelFn  context.CancelFunc
	wg        sync.WaitGroup
	running   bool
	mu        sync.RWMutex
	lastError string
	// nwcFailStreak counts consecutive NWC health failures before reconnect
	nwcFailStreak    int
	lastNwcReconnect time.Time
	// lastAutoRefill gates how often the Hub-side Cashu auto-refill runs
	lastAutoRefill time.Time
	// autoRefillMu serializes auto-refill checks: the 15s supervision ticker
	// and the Start-triggered immediate check run concurrently and would
	// otherwise both pass the cooldown gate and double-refill (hit 2026-08-01)
	autoRefillMu sync.Mutex
	// Auto-refill runtime state, surfaced via the status endpoint so the UI
	// can show real activity (pool balance, last refill, errors) instead of
	// a silent toggle.
	autoRefillLastCheck   time.Time
	autoRefillLastRefill  time.Time
	autoRefillLastAmount  int64
	autoRefillLastBalance int64
	autoRefillLastError   string
}

// AutoRefillStatus is the live state of the Hub-side auto top-up loop.
// The loop always ticks (15s supervision ticker); Enabled (persisted in the
// Routstr app metadata) is the start/stop gate. This struct makes what the
// loop is actually doing visible: pool balance, last check/refill, errors.
type AutoRefillStatus struct {
	Enabled          bool      `json:"enabled"`
	AppID            uint      `json:"appId"`
	Threshold        int64     `json:"threshold"`
	Amount           int64     `json:"amount"`
	CooldownMs       int64     `json:"cooldownMs"`
	PoolBalanceSat   int64     `json:"poolBalanceSat"`
	RoutstrWalletSat int64     `json:"routstrWalletSat"`
	LastCheckAt      time.Time `json:"lastCheckAt,omitempty"`
	LastRefillAt     time.Time `json:"lastRefillAt,omitempty"`
	LastRefillAmount int64     `json:"lastRefillAmount,omitempty"`
	LastError        string    `json:"lastError,omitempty"`
	RoutstrdHealthy  bool      `json:"routstrdHealthy"`
	CocodHealthy     bool      `json:"cocodHealthy"`
}

// AutoRefillConfig mirrors the routstr.autoRefill app metadata: when enabled,
// the Hub tops up the routstrd Cashu wallet from the Routstr app's isolated
// wallet whenever its balance drops below Threshold.
type AutoRefillConfig struct {
	Enabled    bool  `json:"enabled"`
	Threshold  int64 `json:"threshold"`
	Amount     int64 `json:"amount"`
	CooldownMs int64 `json:"cooldownMs"`
}

// NewRoutstrdService creates a new routstrd service manager.
func NewRoutstrdService(svc *service) *RoutstrdService {
	return &RoutstrdService{svc: svc}
}

// Start launches the routstrd and cocod daemons if not already running.
// Non-blocking: supervision runs in a background goroutine.
func (r *RoutstrdService) Start(ctx context.Context) error {
	r.mu.Lock()
	if r.running {
		r.mu.Unlock()
		return fmt.Errorf("routstrd service already running")
	}
	r.running = true
	r.mu.Unlock()

	childCtx, cancel := context.WithCancel(ctx)
	r.mu.Lock()
	r.cancelFn = cancel
	r.mu.Unlock()

	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		r.supervise(childCtx)
	}()

	logger.Logger.Info("Routstrd service started")
	return nil
}

// Stop gracefully shuts down the routstrd daemon (cocod is left running so
// Cashu wallet state is not interrupted by brief Hub restarts).
func (r *RoutstrdService) Stop() {
	r.mu.Lock()
	r.running = false
	cancelFn := r.cancelFn
	r.cancelFn = nil
	r.mu.Unlock()

	if cancelFn != nil {
		cancelFn()
	}

	// Graceful stop of routstrd only — do not stop cocod (wallet)
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", "http://localhost:8008/stop", nil)
	if err == nil {
		resp, err := client.Do(req)
		if err == nil && resp != nil {
			resp.Body.Close()
			logger.Logger.Info("Sent graceful stop to routstrd daemon")
		}
	}

	r.wg.Wait()
	logger.Logger.Info("Routstrd service stopped")
}

// IsRunning returns whether the service is actively managing daemons.
func (r *RoutstrdService) IsRunning() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.running
}

// LastError returns the last error encountered by the service.
func (r *RoutstrdService) LastError() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastError
}

// Status returns the health status of both daemons.
func (r *RoutstrdService) Status() (routstrdHealthy, cocodHealthy bool) {
	routstrdHealthy = r.checkRoutstrdHealth()
	cocodHealthy = r.checkCocodHealth()
	return
}

// GetStatus returns a short status string for the UI.
func (r *RoutstrdService) GetStatus() string {
	if !r.IsRunning() {
		return "stopped"
	}
	routstrdOk, cocodOk := r.Status()
	if routstrdOk && cocodOk {
		return "running"
	}
	if routstrdOk || cocodOk {
		return "degraded"
	}
	return "failed"
}

// GetStatusDetails returns detailed status for API responses.
func (r *RoutstrdService) GetStatusDetails() map[string]interface{} {
	routstrdOk, cocodOk := r.Status()
	return map[string]interface{}{
		"service_running":  r.IsRunning(),
		"routstrd_healthy": routstrdOk,
		"cocod_healthy":    cocodOk,
		"last_error":       r.LastError(),
	}
}

// supervise starts daemons and keeps retrying on failure until ctx is cancelled.
func (r *RoutstrdService) supervise(ctx context.Context) {
	// Immediate attempt
	if err := r.ensureDaemons(ctx); err != nil {
		r.setError(fmt.Sprintf("initial startup failed: %v", err))
		logger.Logger.WithError(err).Error("Failed to start routstrd daemons (will retry)")
		r.svc.eventPublisher.Publish(&events.Event{
			Event: "routstrd_start_failed",
			Properties: map[string]interface{}{
				"error": err.Error(),
			},
		})
	} else {
		r.setError("")
		r.svc.eventPublisher.Publish(&events.Event{Event: "routstrd_started"})
	}

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Only stop routstrd; leave cocod running
			r.stopRoutstrd()
			return
		case <-ticker.C:
			if err := r.ensureDaemons(ctx); err != nil {
				r.setError(fmt.Sprintf("health check failed: %v", err))
				logger.Logger.WithError(err).Warn("Routstrd daemon unhealthy, retrying")
				r.svc.eventPublisher.Publish(&events.Event{
					Event: "routstrd_restarting",
					Properties: map[string]interface{}{
						"error": err.Error(),
					},
				})
			} else {
				if r.LastError() != "" {
					r.setError("")
					logger.Logger.Info("Routstrd daemons healthy again")
				}
				// Daemons up — keep NWC funding source connected (non-fatal)
				r.ensureNwcConnected(ctx)
				// Hub-side Cashu auto-refill (reads routstr.autoRefill app metadata)
				r.checkAutoRefill(ctx)
			}
		}
	}
}

func (r *RoutstrdService) ensureDaemons(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if !r.checkCocodHealth() {
		if err := r.startCocod(); err != nil {
			return fmt.Errorf("failed to start cocod: %w", err)
		}
		// Give cocod a moment after spawn
		if !r.waitFor(ctx, r.checkCocodHealth, 15*time.Second) {
			return fmt.Errorf("cocod did not become healthy within 15s")
		}
	}

	if !r.checkRoutstrdHealth() {
		if err := r.startRoutstrd(); err != nil {
			return fmt.Errorf("failed to start routstrd: %w", err)
		}
		if !r.waitFor(ctx, r.checkRoutstrdHealth, 45*time.Second) {
			return fmt.Errorf("routstrd did not become healthy within 45s")
		}
	}

	return nil
}

func (r *RoutstrdService) waitFor(ctx context.Context, check func() bool, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return false
		}
		if check() {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return check()
}

func (r *RoutstrdService) checkRoutstrdHealth() bool {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get("http://127.0.0.1:8008/health")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// checkCocodHealth prefers socket existence + PID alive; never hangs on CLI ping.
func (r *RoutstrdService) checkCocodHealth() bool {
	home := os.Getenv("HOME")
	socketPath := filepath.Join(home, ".cocod", "cocod.sock")
	if _, err := os.Stat(socketPath); err != nil {
		// Socket missing: daemon is down, or hung on startup (mint rate
		// limiter) — either way the supervisor should try to recover it.
		return false
	}

	pidPath := filepath.Join(home, ".cocod", "cocod.pid")
	if data, err := os.ReadFile(pidPath); err == nil {
		pidStr := strings.TrimSpace(string(data))
		if pid, err := strconv.Atoi(pidStr); err == nil && pid > 0 {
			if r.processExists(pid) {
				return true
			}
			// PID file points at a dead process — reconcile the stale files
			// so the next spawn binds the socket cleanly (startCocod does
			// the same before spawning).
			_ = os.Remove(socketPath)
			_ = os.Remove(pidPath)
			logger.Logger.WithField("pid", pid).Warn("cocod pid file points at a dead process; removed stale socket and pid files")
			return false
		}
	}

	// Socket exists but PID unknown — ambiguous (stale socket from a crash).
	// Treat as unhealthy; startCocod reconciles and respawns.
	return false
}

func (r *RoutstrdService) startCocod() error {
	logger.Logger.Info("Starting cocod daemon...")

	home := os.Getenv("HOME")
	pidPath := filepath.Join(home, ".cocod", "cocod.pid")
	socketPath := filepath.Join(home, ".cocod", "cocod.sock")

	// Reconcile stale state before spawning:
	// - pid file references a dead process → remove pid + stale socket
	// - pid alive but socket missing → hung daemon (mint rate limiter);
	//   kill it so the respawn binds the socket cleanly
	// - no pid file → remove any stale socket
	if data, err := os.ReadFile(pidPath); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 0 {
			if r.processExists(pid) {
				if _, err := os.Stat(socketPath); err != nil {
					logger.Logger.WithField("pid", pid).Warn("cocod pid alive but socket missing (hung); killing and respawning")
					// A hung cocod is stuck walking pending mint operations
					// (mint rate limiter). Clear them so the respawn binds
					// the socket instead of replaying the same hang.
					clearStuckCocodOps(filepath.Join(home, ".cocod", "coco.db"))
					_ = killProcess(pid)
					_ = os.Remove(pidPath)
					_ = os.Remove(socketPath)
				} else {
					logger.Logger.WithField("pid", pid).Info("cocod already running")
					return nil
				}
			} else {
				logger.Logger.WithField("pid", pid).Warn("cocod pid file references a dead process; cleaning stale files")
				_ = os.Remove(pidPath)
				_ = os.Remove(socketPath)
			}
		}
	} else {
		_ = os.Remove(socketPath)
	}

	cocodBin := r.resolveBinary("cocod")
	if cocodBin == "" {
		return fmt.Errorf("cocod binary not found in PATH or ~/.bun/bin")
	}

	cmd := exec.Command(cocodBin, "daemon")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	// Detach so Hub restarts don't kill cocod (no-op on Windows)
	setDetachedProcess(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("spawn cocod: %w", err)
	}
	// Reap in background so we don't leave zombies; process is in its own session
	go func() { _ = cmd.Wait() }()

	logger.Logger.WithField("pid", cmd.Process.Pid).Info("cocod daemon spawned")
	return nil
}

// clearStuckCocodOps deletes pending mint operations from the cocod SQLite
// database at dbPath. A cocod daemon hung on the mint rate limiter is stuck
// replaying stale pending operations at startup (each quote-state check
// round-trips to the mint), so the socket never appears. Clearing them is
// the documented recovery; the ops are unpaid quotes — no ecash is in flight
// for them. The path is a parameter so the recovery is testable.
func clearStuckCocodOps(dbPath string) {
	db, err := sql.Open("sqlite3", dbPath+"?_busy_timeout=5000")
	if err != nil {
		logger.Logger.WithError(err).Warn("cocod recovery: cannot open coco.db")
		return
	}
	defer db.Close()
	res, err := db.Exec("DELETE FROM coco_cashu_mint_operations WHERE state = 'pending'")
	if err != nil {
		logger.Logger.WithError(err).Warn("cocod recovery: cannot clear pending mint operations")
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		logger.Logger.WithField("cleared", n).Info("cocod recovery: cleared stuck pending mint operations")
	}
}

func (r *RoutstrdService) startRoutstrd() error {
	logger.Logger.Info("Starting routstrd daemon...")

	if r.checkRoutstrdHealth() {
		logger.Logger.Info("routstrd already running")
		return nil
	}

	// Prefer the CLI start path (handles lock + health poll)
	routstrdBin := r.resolveBinary("routstrd")
	if routstrdBin != "" {
		cmd := exec.Command(routstrdBin, "start", "--port", "8008")
		cmd.Env = append(os.Environ(),
			"PATH="+filepath.Join(os.Getenv("HOME"), ".bun", "bin")+":"+os.Getenv("PATH"),
			"HOME="+os.Getenv("HOME"),
		)
		cmd.Stdout = nil
		cmd.Stderr = nil
		cmd.Stdin = nil
		setDetachedProcess(cmd)

		if err := cmd.Start(); err != nil {
			logger.Logger.WithError(err).Warn("routstrd start CLI failed, falling back to direct daemon spawn")
		} else {
			go func() { _ = cmd.Wait() }()
			logger.Logger.WithField("pid", cmd.Process.Pid).Info("routstrd start CLI spawned")
			return nil
		}
	}

	// Fallback: direct bun daemon
	home := os.Getenv("HOME")
	bunBin := r.resolveBinary("bun")
	if bunBin == "" {
		return fmt.Errorf("bun binary not found")
	}
	daemonScript := filepath.Join(home, ".bun", "install", "global", "node_modules", "routstrd", "dist", "daemon", "index.js")
	if _, err := os.Stat(daemonScript); err != nil {
		return fmt.Errorf("routstrd daemon script not found: %s", daemonScript)
	}

	cmd := exec.Command(bunBin, "run", daemonScript, "--port", "8008")
	cmd.Env = append(os.Environ(),
		"PATH="+filepath.Join(home, ".bun", "bin")+":"+os.Getenv("PATH"),
		"HOME="+home,
	)
	cmd.Dir = home
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil
	setDetachedProcess(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("spawn routstrd: %w", err)
	}
	go func() { _ = cmd.Wait() }()

	logger.Logger.WithField("pid", cmd.Process.Pid).Info("routstrd daemon spawned")
	return nil
}

func (r *RoutstrdService) stopRoutstrd() {
	client := &http.Client{Timeout: 5 * time.Second}
	req, err := http.NewRequest("POST", "http://127.0.0.1:8008/stop", nil)
	if err != nil {
		return
	}
	resp, err := client.Do(req)
	if err == nil && resp != nil {
		resp.Body.Close()
	}
}

func (r *RoutstrdService) resolveBinary(name string) string {
	// Prefer ~/.bun/bin
	home := os.Getenv("HOME")
	candidates := []string{
		filepath.Join(home, ".bun", "bin", name),
		filepath.Join("/usr/local/bin", name),
		filepath.Join("/usr/bin", name),
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	return ""
}

func (r *RoutstrdService) processExists(pid int) bool {
	if pid <= 0 {
		return false
	}
	return processAlive(pid)
}

func (r *RoutstrdService) setError(msg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastError = msg
}

// ensureNwcConnected checks routstrd NWC status and reconnects from config if needed.
// Never fails ensureDaemons — NWC issues are logged and retried with backoff.
func (r *RoutstrdService) ensureNwcConnected(ctx context.Context) {
	if ctx.Err() != nil || !r.checkRoutstrdHealth() {
		return
	}

	ok := r.checkNwcConnected()
	if ok {
		r.mu.Lock()
		r.nwcFailStreak = 0
		r.mu.Unlock()
		return
	}

	r.mu.Lock()
	r.nwcFailStreak++
	streak := r.nwcFailStreak
	last := r.lastNwcReconnect
	r.mu.Unlock()

	// Require 2 consecutive failures (~30s) before reconnect to avoid relay flapping
	if streak < 2 {
		return
	}
	// Back off at least 2 minutes between reconnect attempts
	if time.Since(last) < 2*time.Minute {
		return
	}

	connStr, err := r.readNwcConnectionString()
	if err != nil || connStr == "" {
		logger.Logger.WithError(err).Debug("NWC unhealthy but no connection string in routstrd config")
		return
	}

	if err := r.reconnectNwc(connStr); err != nil {
		logger.Logger.WithError(err).Warn("Failed to reconnect routstrd NWC")
		r.svc.eventPublisher.Publish(&events.Event{
			Event: "routstrd_nwc_reconnect_failed",
			Properties: map[string]interface{}{
				// Errors can embed raw daemon response bodies — keep
				// published properties short.
				"error": shortError(err),
			},
		})
		return
	}

	r.mu.Lock()
	r.nwcFailStreak = 0
	r.lastNwcReconnect = time.Now()
	r.mu.Unlock()
	logger.Logger.Info("Reconnected routstrd NWC funding source")
	r.svc.eventPublisher.Publish(&events.Event{Event: "routstrd_nwc_reconnected"})
}

func (r *RoutstrdService) checkNwcConnected() bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get("http://127.0.0.1:8008/nwc/status")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return false
	}
	// Daemon wraps as {"output":{...}} or may return flat JSON
	var wrap struct {
		Output json.RawMessage `json:"output"`
	}
	_ = json.Unmarshal(body, &wrap)
	payload := body
	if len(wrap.Output) > 0 {
		payload = wrap.Output
	}
	var status map[string]interface{}
	if err := json.Unmarshal(payload, &status); err != nil {
		return false
	}
	// Accept connected / nwcConnected truthy
	if v, ok := status["connected"].(bool); ok {
		return v
	}
	if v, ok := status["nwcConnected"].(bool); ok {
		return v
	}
	// Some builds expose connectionString presence + walletState
	if ws, ok := status["walletState"].(string); ok && ws != "" && ws != "ERROR" {
		if _, has := status["connectionString"]; has {
			return true
		}
	}
	return false
}

func (r *RoutstrdService) readNwcConnectionString() (string, error) {
	home := os.Getenv("HOME")
	cfgPath := filepath.Join(home, ".routstrd", "config.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return "", err
	}
	var cfg struct {
		Nwc *struct {
			ConnectionString string `json:"connectionString"`
			Mode             string `json:"mode"`
		} `json:"nwc"`
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return "", err
	}
	if cfg.Nwc == nil {
		return "", fmt.Errorf("no nwc block in config")
	}
	return strings.TrimSpace(cfg.Nwc.ConnectionString), nil
}

func (r *RoutstrdService) reconnectNwc(connectionString string) error {
	client := &http.Client{Timeout: 25 * time.Second}
	body, _ := json.Marshal(map[string]string{"connectionString": connectionString})
	resp, err := client.Post("http://127.0.0.1:8008/nwc/connect", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return fmt.Errorf("nwc/connect status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// checkAutoRefill tops up the routstrd Cashu wallet from the Routstr app's
// isolated wallet when its balance drops below the configured threshold.
// Config lives in the Routstr app's metadata (routstr.autoRefill), so it is
// controlled from the Hub UI. All money movement is Hub-direct:
//
//	mint invoice -> Hub pays fromAppId (Routstr app wallet)
//
// Every exit records runtime state (last check, balance, refill, error) so
// the status endpoint can show what the loop is actually doing.
func (r *RoutstrdService) checkAutoRefill(ctx context.Context) {
	// Serialize with other checks (ticker + immediate Start trigger). Without
	// this, two concurrent checks both pass the cooldown gate and refill
	// twice — each refill pays a network fee.
	r.autoRefillMu.Lock()
	defer r.autoRefillMu.Unlock()
	nowT := now()
	if ctx.Err() != nil {
		return
	}
	if !r.checkRoutstrdHealth() {
		r.recordAutoRefill(nowT, 0, 0, time.Time{}, "routstrd daemon unhealthy")
		return
	}

	app := r.findRoutstrApp()
	if app == nil {
		r.recordAutoRefill(nowT, 0, 0, time.Time{}, "no Routstr app found")
		return
	}
	cfg := readAutoRefillConfig(app)
	if cfg == nil || !cfg.Enabled || cfg.Threshold <= 0 || cfg.Amount <= 0 {
		// Stopped (or unconfigured) is a normal state, not an error
		r.recordAutoRefill(nowT, 0, 0, time.Time{}, "")
		return
	}

	r.mu.Lock()
	last := r.lastAutoRefill
	r.mu.Unlock()

	balance, err := r.getCashuWalletBalance()
	if err != nil {
		r.recordAutoRefill(nowT, 0, 0, time.Time{}, fmt.Sprintf("could not read Cashu wallet balance: %v", err))
		logger.Logger.WithError(err).Debug("auto-refill: could not read Cashu wallet balance")
		return
	}
	if !shouldAutoRefill(balance, cfg.Threshold, last, nowT, cfg.CooldownMs) {
		// Healthy (pool above the line) or within the cooldown — nothing to do
		r.recordAutoRefill(nowT, balance, 0, time.Time{}, "")
		return
	}

	logger.Logger.Infof("auto-refill: Cashu balance %d sats < threshold %d, topping up %d sats from Routstr wallet (app %d)", balance, cfg.Threshold, cfg.Amount, app.ID)

	// Create a mint invoice for the refill amount
	invBody, _ := json.Marshal(map[string]int64{"amount": cfg.Amount})
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post("http://127.0.0.1:8008/wallet/receive/bolt11", "application/json", bytes.NewReader(invBody))
	if err != nil {
		r.markAutoRefillAttempted(nowT)
		r.recordAutoRefill(nowT, balance, 0, time.Time{}, fmt.Sprintf("failed to create mint invoice: %v", err))
		logger.Logger.WithError(err).Warn("auto-refill: failed to create mint invoice")
		return
	}
	defer resp.Body.Close()
	var invResp struct {
		Output struct {
			Invoice string `json:"invoice"`
		} `json:"output"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&invResp); err != nil || invResp.Output.Invoice == "" {
		r.markAutoRefillAttempted(nowT)
		r.recordAutoRefill(nowT, balance, 0, time.Time{}, "mint invoice response invalid")
		logger.Logger.WithError(err).Warn("auto-refill: mint invoice response invalid")
		return
	}

	// The daemon is unauthenticated on localhost — verify the returned
	// invoice is exactly for cfg.Amount before paying it. A wrong or hostile
	// response would otherwise drain the Routstr wallet up to its budget.
	payReq, decodeErr := decodepay.Decodepay(invResp.Output.Invoice)
	if decodeErr != nil || validateRefillInvoiceAmount(payReq.MSatoshi, cfg.Amount*1000) != nil {
		r.markAutoRefillAttempted(nowT)
		r.recordAutoRefill(nowT, balance, 0, time.Time{}, "mint invoice amount mismatch")
		logger.Logger.WithFields(map[string]interface{}{
			"expected_msat": cfg.Amount * 1000,
			"actual_msat":   payReq.MSatoshi,
			"decode_error":  decodeErr,
		}).Warn("auto-refill: mint invoice amount does not match requested amount")
		return
	}

	// Pay it from the Routstr app's isolated wallet (Hub-direct, no relay)
	appID := app.ID
	lnClient := r.svc.GetLNClient()
	if lnClient == nil {
		r.markAutoRefillAttempted(nowT)
		r.recordAutoRefill(nowT, balance, 0, time.Time{}, "LN client not started")
		logger.Logger.Warn("auto-refill: LN client not started")
		return
	}
	if _, err := r.svc.GetTransactionsService().SendPaymentSync(invResp.Output.Invoice, nil, nil, lnClient, &appID, nil); err != nil {
		r.markAutoRefillAttempted(nowT)
		r.recordAutoRefill(nowT, balance, 0, time.Time{}, fmt.Sprintf("payment failed: %v", err))
		logger.Logger.WithError(err).Warn("auto-refill: payment failed (Routstr wallet may be low)")
		return
	}
	r.markAutoRefillAttempted(nowT)
	r.recordAutoRefill(nowT, balance, cfg.Amount, nowT, "")
	logger.Logger.Infof("auto-refill: payment initiated for %d sats", cfg.Amount)
}

// recordAutoRefill stores the outcome of the latest auto-refill check for the
// status endpoint. refillAt zero means no refill happened on this check.
func (r *RoutstrdService) recordAutoRefill(check time.Time, balance, amount int64, refillAt time.Time, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.autoRefillLastCheck = check
	r.autoRefillLastBalance = balance
	r.autoRefillLastError = errMsg
	if !refillAt.IsZero() {
		r.autoRefillLastRefill = refillAt
		r.autoRefillLastAmount = amount
	}
}

// GetAutoRefillStatus returns the live state of the auto top-up loop: the
// config the loop is honoring, current pool balance, and last activity.
func (r *RoutstrdService) GetAutoRefillStatus() *AutoRefillStatus {
	status := &AutoRefillStatus{
		RoutstrdHealthy: r.checkRoutstrdHealth(),
		CocodHealthy:    r.checkCocodHealth(),
	}
	app := r.findRoutstrApp()
	if app == nil {
		return status
	}
	cfg := readAutoRefillConfig(app)
	status.AppID = app.ID
	if cfg != nil {
		status.Enabled = cfg.Enabled
		status.Threshold = cfg.Threshold
		status.Amount = cfg.Amount
		status.CooldownMs = cfg.CooldownMs
	}
	status.RoutstrWalletSat = 0
	if balanceMsat, err := queries.GetIsolatedBalanceMsat(r.svc.GetDB(), app.ID); err == nil {
		status.RoutstrWalletSat = balanceMsat / 1000
	}
	if status.RoutstrdHealthy {
		if balance, err := r.getCashuWalletBalance(); err == nil {
			status.PoolBalanceSat = balance
		} else if status.LastError == "" {
			status.LastError = fmt.Sprintf("could not read pool balance: %v", err)
		}
	}
	r.mu.RLock()
	status.LastCheckAt = r.autoRefillLastCheck
	status.LastRefillAt = r.autoRefillLastRefill
	status.LastRefillAmount = r.autoRefillLastAmount
	if r.autoRefillLastError != "" && status.LastError == "" {
		status.LastError = r.autoRefillLastError
	}
	r.mu.RUnlock()
	return status
}

// SetAutoRefillEnabled is the server-side start/stop for auto top-up. It
// flips the routstr.autoRefill.enabled flag in the Routstr app metadata
// (the same read-modify-write the UI performs) and, when starting, runs one
// immediate check so the user sees the loop act right away (cooldown still
// respected to avoid fee-heavy refills). Optional threshold/amount (start
// only) override the stored config atomically — the user's typed values are
// what the loop honors, no blur-save race.
func (r *RoutstrdService) SetAutoRefillEnabled(ctx context.Context, enabled bool, threshold, amount *int64) (*AutoRefillStatus, error) {
	app := r.findRoutstrApp()
	if app == nil {
		return nil, errors.New("no Routstr app found")
	}
	cfg := readAutoRefillConfig(app)
	if cfg == nil {
		cfg = &AutoRefillConfig{Enabled: enabled, Threshold: 500, Amount: 1000, CooldownMs: 5 * 60 * 1000}
	}
	cfg.Enabled = enabled
	if enabled {
		if threshold != nil && *threshold > 0 {
			cfg.Threshold = *threshold
		}
		if amount != nil && *amount > 0 {
			cfg.Amount = *amount
		}
	}
	if err := r.writeAutoRefillConfig(app, cfg); err != nil {
		return nil, err
	}
	if enabled {
		logger.Logger.Info("auto-refill: started (immediate check)")
		// Run the immediate check off the request thread: it talks to the
		// daemon and can take tens of seconds (invoice + payment timeouts).
		go func() {
			checkCtx, cancelCheck := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancelCheck()
			r.checkAutoRefill(checkCtx)
		}()
	} else {
		logger.Logger.Info("auto-refill: stopped")
	}
	return r.GetAutoRefillStatus(), nil
}

// writeAutoRefillConfig persists the routstr.autoRefill block into the app's
// metadata. Read-modify-write: metadata also carries apiKey/clientId/balance.
func (r *RoutstrdService) writeAutoRefillConfig(app *db.App, cfg *AutoRefillConfig) error {
	var meta map[string]interface{}
	if err := json.Unmarshal(app.Metadata, &meta); err != nil {
		return fmt.Errorf("read app metadata: %w", err)
	}
	routstrMeta, _ := meta["routstr"].(map[string]interface{})
	if routstrMeta == nil {
		routstrMeta = map[string]interface{}{}
	}
	routstrMeta["autoRefill"] = cfg
	meta["routstr"] = routstrMeta
	bytes, err := json.Marshal(meta)
	if err != nil {
		return err
	}
	if err := r.svc.GetDB().Model(&db.App{}).Where("id = ?", app.ID).Update("metadata", datatypes.JSON(bytes)).Error; err != nil {
		return fmt.Errorf("write app metadata: %w", err)
	}
	app.Metadata = bytes
	return nil
}

func (r *RoutstrdService) markAutoRefillAttempted(t time.Time) {
	r.mu.Lock()
	r.lastAutoRefill = t
	r.mu.Unlock()
}

func now() time.Time { return time.Now() }

// shortError truncates error strings for event properties so raw daemon
// response bodies never leak into the event bus.
func shortError(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if len(msg) > 300 {
		return msg[:300] + "..."
	}
	return msg
}

// findRoutstrApp returns the Routstr app the auto top-up loop should honor:
// the one with an enabled config, else the one with any autoRefill config
// block (so status/start/stop keep pointing at the app the user configured
// even while stopped), else the first Routstr app.
func (r *RoutstrdService) findRoutstrApp() *db.App {
	if r.svc == nil || r.svc.GetDB() == nil {
		return nil
	}
	var apps []db.App
	if err := r.svc.GetDB().Find(&apps).Error; err != nil {
		return nil
	}
	return selectRoutstrApp(apps)
}

// selectRoutstrApp picks the Routstr app to supervise, in tier order:
//  1. an app with auto-refill enabled and sane values,
//  2. the first app with an explicit autoRefill block ("configured"),
//  3. the first Routstr app (fallback).
//
// Returns nil when no Routstr app exists.
func selectRoutstrApp(apps []db.App) *db.App {
	var fallback *db.App
	var configured *db.App
	for i := range apps {
		var meta map[string]interface{}
		if err := json.Unmarshal(apps[i].Metadata, &meta); err != nil {
			continue
		}
		if id, _ := meta["app_store_app_id"].(string); id == "routstr" {
			if fallback == nil {
				fallback = &apps[i]
			}
			routstrMeta, _ := meta["routstr"].(map[string]interface{})
			_, hasAutoRefillBlock := routstrMeta["autoRefill"].(map[string]interface{})
			// Only apps with an explicit autoRefill block count as
			// "configured": readAutoRefillConfig always returns defaulted
			// values, so it cannot distinguish absent from configured.
			if !hasAutoRefillBlock {
				continue
			}
			if configured == nil {
				configured = &apps[i]
			}
			cfg := readAutoRefillConfig(&apps[i])
			if cfg.Enabled && cfg.Threshold > 0 && cfg.Amount > 0 {
				return &apps[i]
			}
		}
	}
	if configured != nil {
		return configured
	}
	return fallback
}

// readAutoRefillConfig parses routstr.autoRefill from the app metadata.
func readAutoRefillConfig(app *db.App) *AutoRefillConfig {
	var meta map[string]interface{}
	if err := json.Unmarshal(app.Metadata, &meta); err != nil {
		return nil
	}
	routstrMeta, _ := meta["routstr"].(map[string]interface{})
	if routstrMeta == nil {
		return nil
	}
	raw, _ := routstrMeta["autoRefill"].(map[string]interface{})
	// Missing block = unconfigured: report (and, on Start, use) the sane
	// defaults rather than zeros, so status matches the UI inputs and the
	// loop never misreads an unconfigured app as "refill 0 sats".
	cfg := &AutoRefillConfig{
		Enabled:    false,
		Threshold:  500,
		Amount:     1000,
		CooldownMs: 5 * 60 * 1000,
	}
	if raw == nil {
		return cfg
	}
	if v, ok := raw["enabled"].(bool); ok {
		cfg.Enabled = v
	}
	if v, ok := raw["threshold"].(float64); ok {
		cfg.Threshold = int64(v)
	}
	if v, ok := raw["amount"].(float64); ok {
		cfg.Amount = int64(v)
	}
	if v, ok := raw["cooldownMs"].(float64); ok {
		cfg.CooldownMs = int64(v)
	}
	return cfg
}

// shouldAutoRefill reports whether the pool is below the threshold AND the
// cooldown since the last refill attempt has elapsed. A zero/negative
// cooldown falls back to 5 minutes; a zero last-refill time means "never
// refilled" and always passes the cooldown gate.
func shouldAutoRefill(balanceSat, thresholdSat int64, lastRefillAt, nowT time.Time, cooldownMs int64) bool {
	if balanceSat >= thresholdSat {
		return false
	}
	if cooldownMs <= 0 {
		cooldownMs = 5 * 60 * 1000
	}
	return nowT.Sub(lastRefillAt) >= time.Duration(cooldownMs)*time.Millisecond
}

// validateRefillInvoiceAmount verifies the decoded invoice amount matches the
// requested refill amount. Called with the decodepay result at the call site;
// the daemon is unauthenticated on localhost, so a mismatched invoice must
// never be paid from the Routstr wallet.
func validateRefillInvoiceAmount(actualMsat, requestedMsat int64) error {
	if actualMsat != requestedMsat {
		return fmt.Errorf("invoice amount %d msat != requested %d msat", actualMsat, requestedMsat)
	}
	return nil
}

// getCashuWalletBalance returns the total routstrd Cashu wallet balance in sats.
func (r *RoutstrdService) getCashuWalletBalance() (int64, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("http://127.0.0.1:8008/wallet/balance")
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	var wrap struct {
		Output struct {
			Balances map[string]interface{} `json:"balances"`
		} `json:"output"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 64*1024)).Decode(&wrap); err != nil {
		return 0, err
	}
	var total int64
	for _, v := range wrap.Output.Balances {
		switch n := v.(type) {
		case float64:
			total += int64(n)
		case int64:
			total += n
		}
	}
	return total, nil
}
