package registry

import (
	"errors"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

type Project struct {
	Name string            `yaml:"name"`
	Cwd  string            `yaml:"cwd"`
	Cmd  string            `yaml:"cmd"`
	Env  map[string]string `yaml:"env,omitempty"`
}

type Registry struct {
	Version  int       `yaml:"version"`
	Projects []Project `yaml:"projects"`
}

// Load reads the registry file. Missing file returns an empty registry (no error).
func Load(path string) (*Registry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &Registry{Version: 1}, nil
		}
		return nil, err
	}
	var r Registry
	if err := yaml.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	if r.Version == 0 {
		r.Version = 1
	}
	return &r, nil
}
