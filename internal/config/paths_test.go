package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPaths(t *testing.T) {
	p := New()
	if !strings.HasSuffix(p.ConfigDir, filepath.Join(".config", "devs")) {
		t.Errorf("ConfigDir = %q, want suffix .config/devs", p.ConfigDir)
	}
	if !strings.HasSuffix(p.StateDir, filepath.Join(".local", "state", "devs")) {
		t.Errorf("StateDir = %q, want suffix .local/state/devs", p.StateDir)
	}
	if filepath.Base(p.RegistryFile) != "registry.yaml" {
		t.Errorf("RegistryFile = %q, want registry.yaml", p.RegistryFile)
	}
	if filepath.Base(p.StateFile) != "state.json" {
		t.Errorf("StateFile = %q, want state.json", p.StateFile)
	}
	if filepath.Base(p.LockFile) != "lock" {
		t.Errorf("LockFile = %q, want lock", p.LockFile)
	}
	if filepath.Base(p.LogsDir) != "logs" {
		t.Errorf("LogsDir = %q, want logs", p.LogsDir)
	}
}
