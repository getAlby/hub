//go:build !windows

package service

import (
	"os/exec"
	"syscall"
)

// setDetachedProcess runs the command in its own session so Hub restarts
// don't kill the daemon. Unix-only: Windows has no Setsid.
func setDetachedProcess(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// processAlive reports whether a process with the given PID exists.
// Signal 0 checks existence without killing.
func processAlive(pid int) bool {
	return syscall.Kill(pid, 0) == nil
}

// killProcess terminates a process (used to recover a hung daemon).
func killProcess(pid int) error {
	return syscall.Kill(pid, syscall.SIGKILL)
}
