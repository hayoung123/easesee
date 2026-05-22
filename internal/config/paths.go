package config

import (
	"os"
	"path/filepath"
)

type Paths struct {
	ConfigDir    string
	StateDir     string
	RegistryFile string
	StateFile    string
	LockFile     string
	LogsDir      string
}

func New() Paths {
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".config", "devs")
	stateDir := filepath.Join(home, ".local", "state", "devs")
	return Paths{
		ConfigDir:    configDir,
		StateDir:     stateDir,
		RegistryFile: filepath.Join(configDir, "registry.yaml"),
		StateFile:    filepath.Join(stateDir, "state.json"),
		LockFile:     filepath.Join(stateDir, "lock"),
		LogsDir:      filepath.Join(stateDir, "logs"),
	}
}

// EnsureDirs creates ConfigDir, StateDir, and LogsDir if missing.
func (p Paths) EnsureDirs() error {
	for _, d := range []string{p.ConfigDir, p.StateDir, p.LogsDir} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
	}
	return nil
}
