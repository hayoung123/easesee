package config

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPaths(t *testing.T) {
	p := New()
	if !strings.HasSuffix(p.ConfigDir, filepath.Join(".config", "easesee")) {
		t.Errorf("ConfigDir = %q, want suffix .config/easesee", p.ConfigDir)
	}
	if !strings.HasSuffix(p.StateDir, filepath.Join(".local", "state", "easesee")) {
		t.Errorf("StateDir = %q, want suffix .local/state/easesee", p.StateDir)
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
