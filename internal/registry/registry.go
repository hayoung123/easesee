package registry

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

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

func (r *Registry) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(r)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (p Project) Validate() error {
	if p.Name == "" {
		return errors.New("name is required")
	}
	if p.Cwd == "" {
		return errors.New("cwd is required")
	}
	if p.Cmd == "" {
		return errors.New("cmd is required")
	}
	return nil
}

func (r *Registry) Find(name string) (int, bool) {
	for i, p := range r.Projects {
		if p.Name == name {
			return i, true
		}
	}
	return -1, false
}

func (r *Registry) Add(p Project) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if _, ok := r.Find(p.Name); ok {
		return fmt.Errorf("project %q already exists", p.Name)
	}
	r.Projects = append(r.Projects, p)
	return nil
}

func (r *Registry) Replace(p Project) error {
	if err := p.Validate(); err != nil {
		return err
	}
	i, ok := r.Find(p.Name)
	if !ok {
		return fmt.Errorf("project %q not found", p.Name)
	}
	r.Projects[i] = p
	return nil
}

func (r *Registry) Remove(name string) error {
	i, ok := r.Find(name)
	if !ok {
		return fmt.Errorf("project %q not found", name)
	}
	r.Projects = append(r.Projects[:i], r.Projects[i+1:]...)
	return nil
}
