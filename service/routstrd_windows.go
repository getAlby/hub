//go:build windows

package service

import (
	"os/exec"

	"golang.org/x/sys/windows"
)

// setDetachedProcess is a no-op on Windows: there is no Setsid attribute,
// and the daemons (cocod, routstrd) are Unix binaries not supported on
// Windows. This file exists so the package compiles for the Wails
// windows/amd64 build.
func setDetachedProcess(cmd *exec.Cmd) {}

// processAlive checks process existence via OpenProcess. FindProcess alone
// is not usable on Windows (it always succeeds), and Signal(0) is
// unimplemented.
func processAlive(pid int) bool {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	windows.CloseHandle(h)
	return true
}

// killProcess terminates a process via TerminateProcess. No-op-safe: the
// daemons are Unix binaries, so this path is only reachable on Unix.
func killProcess(pid int) error {
	h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		return err
	}
	defer windows.CloseHandle(h)
	return windows.TerminateProcess(h, 1)
}
