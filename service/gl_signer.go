package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/getAlby/hub/events"
	"github.com/getAlby/hub/logger"
	"github.com/sirupsen/logrus"
)

// GreenlightSignerService supervises `glcli signer run` while the hub is unlocked.
type GreenlightSignerService struct {
	mu       sync.Mutex
	running  bool
	cancelFn context.CancelFunc
	wg       sync.WaitGroup
	dataDir  string
	network  string
	glcli    string
	pidPath  string
	lastErr  string
	cmd      *exec.Cmd // the live signer process, when spawned

	eventPublisher       events.EventPublisher // nil in external-signer mode
	lastPublishedHealthy bool
}

func NewGreenlightSignerService() *GreenlightSignerService {
	return &GreenlightSignerService{lastPublishedHealthy: false}
}

func (s *GreenlightSignerService) Start(ctx context.Context, dataDir, network, glcliPath string, eventPublisher events.EventPublisher) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	s.dataDir = dataDir
	s.network = network
	if s.network == "" {
		s.network = "bitcoin"
	}
	s.glcli = glcliPath
	s.pidPath = filepath.Join(dataDir, "signer.pid")
	s.eventPublisher = eventPublisher
	s.lastPublishedHealthy = false
	child, cancel := context.WithCancel(ctx)
	s.cancelFn = cancel
	s.mu.Unlock()

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.supervise(child)
	}()
	logger.Logger.Info("Greenlight signer service started")
	return nil
}

func (s *GreenlightSignerService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	cancel := s.cancelFn
	s.cancelFn = nil
	s.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	s.killSigner()
	s.wg.Wait()
	logger.Logger.Info("Greenlight signer service stopped")
}

func (s *GreenlightSignerService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}

func (s *GreenlightSignerService) LastError() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastErr
}

func (s *GreenlightSignerService) setErr(msg string) {
	s.mu.Lock()
	s.lastErr = msg
	s.mu.Unlock()
}

func (s *GreenlightSignerService) supervise(ctx context.Context) {
	if err := s.ensure(ctx); err != nil {
		s.setErr(err.Error())
		logger.Logger.WithError(err).Error("greenlight signer initial start failed (will retry)")
	} else {
		s.setErr("")
	}
	t := time.NewTicker(15 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			s.killSigner()
			return
		case <-t.C:
			if err := s.ensure(ctx); err != nil {
				s.setErr(err.Error())
				s.publishHealth(false, err.Error())
				logger.Logger.WithError(err).Warn("greenlight signer unhealthy, retrying")
			} else if s.LastError() != "" {
				// recovered: error was set previously but ensure() now succeeds
				s.setErr("")
				s.publishHealth(true, "")
				logger.Logger.Info("greenlight signer healthy again")
			}
		}
	}
}

func (s *GreenlightSignerService) publishHealth(healthy bool, lastErr string) {
	if s.eventPublisher == nil {
		return
	}
	s.mu.Lock()
	if s.lastPublishedHealthy == healthy {
		s.mu.Unlock()
		return // no transition
	}
	s.lastPublishedHealthy = healthy
	s.mu.Unlock()
	props := map[string]interface{}{"healthy": healthy}
	if !healthy {
		props["error"] = lastErr
	}
	s.eventPublisher.Publish(&events.Event{
		Event:      "nwc_gl_signer_health",
		Properties: props,
	})
}

func (s *GreenlightSignerService) ensure(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if s.processAlive() {
		return nil
	}
	return s.spawn()
}

func (s *GreenlightSignerService) resolveBinary() (string, error) {
	if s.glcli != "" {
		if p, err := exec.LookPath(s.glcli); err == nil {
			return p, nil
		}
		if _, err := os.Stat(s.glcli); err == nil {
			return s.glcli, nil
		}
	}
	if p, err := exec.LookPath("glcli"); err == nil {
		return p, nil
	}
	for _, p := range []string{"/root/.cargo/bin/glcli", "/usr/local/bin/glcli"} {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("glcli not found")
}

func (s *GreenlightSignerService) spawn() error {
	bin, err := s.resolveBinary()
	if err != nil {
		return err
	}
	backupPath := filepath.Join(s.dataDir, "backup.json")
	args := []string{
		"-d", s.dataDir,
		"-n", s.network,
		"signer", "run",
		"--backup-path", backupPath,
	}
	logPath := filepath.Join(s.dataDir, "signer.log")
	logF, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("open signer log: %w", err)
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdout = logF
	cmd.Stderr = logF
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		_ = logF.Close()
		return fmt.Errorf("spawn glcli signer: %w", err)
	}
	// parent keeps logF open via cmd; reaper closes process, leave file to OS
	pid := cmd.Process.Pid
	_ = os.WriteFile(s.pidPath, []byte(strconv.Itoa(pid)), 0o600)
	s.mu.Lock()
	s.cmd = cmd
	s.mu.Unlock()
	go func() {
		_ = cmd.Wait()
		_ = logF.Close()
		// the process is gone: clear the tracked cmd so processAlive
		// stops reporting it as alive (a stale pid file alone could be
		// reused by an unrelated process)
		s.mu.Lock()
		if s.cmd == cmd {
			s.cmd = nil
		}
		s.mu.Unlock()
	}()
	logger.Logger.WithFields(logrus.Fields{"pid": pid, "data_dir": s.dataDir, "log": logPath}).Info("glcli signer spawned")
	// brief settle
	time.Sleep(800 * time.Millisecond)
	if !s.processAlive() {
		tail, _ := os.ReadFile(logPath)
		return fmt.Errorf("signer exited immediately after spawn: %s", strings.TrimSpace(string(tail)))
	}
	return nil
}

func (s *GreenlightSignerService) processAlive() bool {
	s.mu.Lock()
	cmd := s.cmd
	s.mu.Unlock()
	if cmd != nil && cmd.Process != nil {
		// signal 0 probes liveness of the exact process we spawned —
		// immune to pid-file staleness/pid reuse
		return cmd.Process.Signal(syscall.Signal(0)) == nil
	}
	data, err := os.ReadFile(s.pidPath)
	if err != nil {
		return false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return false
	}
	return syscall.Kill(pid, 0) == nil
}

func (s *GreenlightSignerService) killSigner() {
	data, err := os.ReadFile(s.pidPath)
	if err != nil {
		return
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 0 {
		return
	}
	_ = syscall.Kill(pid, syscall.SIGTERM)
	// wait up to 3s then KILL
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if syscall.Kill(pid, 0) != nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
	_ = os.Remove(s.pidPath)
}
