package process

import (
	"os"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/proshy/devs/internal/registry"
)

func TestStart_DetachedAndCaptures(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "x.log")
	pid, err := Start(registry.Project{
		Name: "test",
		Cwd:  dir,
		Cmd:  "/bin/sh -c 'echo hello; sleep 30'",
	}, logPath)
	if err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() {
		_ = syscall.Kill(-pid, syscall.SIGKILL) // kill process group
	})
	if pid <= 0 {
		t.Errorf("pid = %d", pid)
	}
	time.Sleep(500 * time.Millisecond)
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log read: %v", err)
	}
	if !contains(string(data), "hello") {
		t.Errorf("log = %q, want to contain 'hello'", string(data))
	}
}

func TestStop_GracefulThenForce(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "stop.log")
	pid, err := Start(registry.Project{
		Name: "test", Cwd: dir, Cmd: "/bin/sh -c 'trap : TERM; while true; do sleep 1; done'",
	}, logPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = syscall.Kill(-pid, syscall.SIGKILL) })

	if err := Stop(pid, 500*time.Millisecond); err != nil {
		t.Fatalf("Stop: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	if err := syscall.Kill(pid, syscall.Signal(0)); err == nil {
		t.Errorf("pid %d still alive", pid)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
