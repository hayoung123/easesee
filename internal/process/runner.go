package process

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/proshy/devs/internal/registry"
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

// Stop sends SIGTERM to the process group. If still alive after timeout, sends SIGKILL.
func Stop(pid int, gracePeriod time.Duration) error {
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return err
	}
	deadline := time.Now().Add(gracePeriod)
	for time.Now().Before(deadline) {
		if syscall.Kill(pid, syscall.Signal(0)) != nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
		return err
	}
	return nil
}

// IsAlive returns true if signal 0 succeeds.
func IsAlive(pid int) bool {
	return syscall.Kill(pid, syscall.Signal(0)) == nil
}
