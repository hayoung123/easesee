package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/hayoung123/easesee/internal/registry"
)

// Start launches the project's cmd in cwd, detached from the parent process.
// stdout/stderr go to logPath. Returns the new PID.
func Start(p registry.Project, logPath string) (int, error) {
	if p.Cmd == "" {
		return 0, errors.New("cmd is empty")
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return 0, err
	}
	logF, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return 0, err
	}

	cmd := exec.Command("/bin/sh", "-c", p.Cmd)
	cmd.Dir = p.Cwd
	cmd.Stdout = logF
	cmd.Stderr = logF
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
	cmd.Env = os.Environ()
	for k, v := range p.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	_, _ = fmt.Fprintf(logF, "\n--- devs start %s ---\n", time.Now().Format(time.RFC3339))

	if err := cmd.Start(); err != nil {
		_ = logF.Close()
		return 0, err
	}
	go func() { _ = cmd.Wait() }()
	return cmd.Process.Pid, nil
}

// Stop sends SIGTERM to the process, then SIGKILL after gracePeriod if still alive.
// Signals both the process group (covers setsid-spawned trees we own) and the
// direct PID (covers externally-started servers whose PGID differs from PID).
// After the tracked pid dies, SIGKILL is sent to the whole process group so that
// child processes (e.g. vite/node spawned by pnpm) release their ports immediately
// rather than continuing graceful shutdown while the caller assumes the port is free.
func Stop(pid int, gracePeriod time.Duration) error {
	pgid, pgidErr := syscall.Getpgid(pid)
	signal(pid, syscall.SIGTERM)
	deadline := time.Now().Add(gracePeriod)
	for time.Now().Before(deadline) {
		if !IsAlive(pid) {
			// Parent is dead but children (e.g. vite) may still hold the port.
			// SIGKILL the whole group so ports are released before we return.
			if pgidErr == nil && pgid > 1 {
				_ = syscall.Kill(-pgid, syscall.SIGKILL)
			}
			time.Sleep(50 * time.Millisecond)
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	signal(pid, syscall.SIGKILL)
	return nil
}

// signal sends sig to both the process group and the process itself,
// ignoring ESRCH (already gone) errors.
func signal(pid int, sig syscall.Signal) {
	if pgid, err := syscall.Getpgid(pid); err == nil && pgid > 1 {
		_ = syscall.Kill(-pgid, sig)
	}
	_ = syscall.Kill(pid, sig)
}

// IsAlive returns true if signal 0 succeeds.
func IsAlive(pid int) bool {
	return syscall.Kill(pid, syscall.Signal(0)) == nil
}
