# devs Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build `devs`, a Go-based TUI for managing locally registered dev servers with agent-friendly registration and survival across dashboard restarts.

**Architecture:** Single Go binary with cobra subcommands. Default invocation launches a Bubbletea TUI. `devs register` adds projects to a YAML registry. The TUI reads the registry, polls `lsof + ps` every 2s to detect running listeners, matches them to registry entries by cwd, and lets the user toggle start/stop. Started processes are detached via `setsid` so they survive dashboard exit. State of dashboard-spawned PIDs is persisted to JSON for re-attach.

**Tech Stack:** Go 1.22+, Bubbletea + Bubbles + Lipgloss (TUI), cobra (CLI), yaml.v3 (registry), stdlib `os/exec` + `syscall` (process management). External commands: `lsof`, `ps`, `git`.

---

## File Structure

```
devs/
├── cmd/devs/main.go
├── internal/
│   ├── config/
│   │   ├── paths.go              # ~/.config/devs, ~/.local/state/devs paths
│   │   └── paths_test.go
│   ├── registry/
│   │   ├── registry.go           # types, Load, Save, Add, Remove
│   │   └── registry_test.go
│   ├── state/
│   │   ├── state.go              # RuntimeState, Load, Save, Lock
│   │   └── state_test.go
│   ├── discovery/
│   │   ├── lsof.go               # ListListeners
│   │   ├── ps.go                 # GetProcInfo
│   │   ├── matcher.go            # MatchProjects
│   │   └── matcher_test.go
│   ├── process/
│   │   ├── runner.go             # Start (detached), Stop, Kill
│   │   └── runner_test.go
│   ├── git/
│   │   └── git.go                # Branch, IsDirty
│   ├── cli/
│   │   ├── root.go               # cobra root + Execute
│   │   ├── register.go           # devs register
│   │   └── ls.go                 # devs ls
│   └── tui/
│       ├── keys.go               # key bindings
│       ├── styles.go             # lipgloss styles
│       ├── app.go                # top-level model
│       ├── table.go              # table view
│       ├── log.go                # log pane
│       └── form.go               # add form
├── skills/
│   ├── devs-register/SKILL.md
│   └── devs-help/SKILL.md
├── docs/superpowers/
│   ├── specs/2026-05-22-devs-design.md
│   └── plans/2026-05-22-devs-implementation.md
├── Makefile
├── README.md
├── INSTALL.md
├── go.mod
└── .gitignore
```

Module path: `github.com/proshy/devs`. (Adjust if final repo URL differs; replace in all imports.)

---

## Task 1: Module Init + .gitignore

**Files:**
- Create: `go.mod`
- Create: `.gitignore`

- [ ] **Step 1: Init Go module**

Run from repo root (`~/Desktop/devs`):
```bash
go mod init github.com/proshy/devs
```
Expected: creates `go.mod` with `module github.com/proshy/devs` and Go version.

- [ ] **Step 2: Add .gitignore**

Create `.gitignore`:
```
# Build artifacts
/bin/
/devs

# Editor
.vscode/
.idea/
*.swp
.DS_Store

# Logs (from running tool, in case of accidental local state)
*.log

# Go
*.test
*.out
coverage.txt
```

- [ ] **Step 3: Commit**

```bash
git add go.mod .gitignore
git commit -m "chore: init Go module and gitignore"
```

---

## Task 2: Config Paths

**Files:**
- Create: `internal/config/paths.go`
- Create: `internal/config/paths_test.go`

- [ ] **Step 1: Write failing test**

`internal/config/paths_test.go`:
```go
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
```

- [ ] **Step 2: Run, see fail**

```bash
go test ./internal/config/
```
Expected: FAIL (package undefined).

- [ ] **Step 3: Implement**

`internal/config/paths.go`:
```go
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
```

- [ ] **Step 4: Run, see pass**

```bash
go test ./internal/config/
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/config/
git commit -m "feat(config): add XDG-like path resolver"
```

---

## Task 3: Registry — Types + Load

**Files:**
- Create: `internal/registry/registry.go`
- Create: `internal/registry/registry_test.go`

- [ ] **Step 1: Add yaml dependency**

```bash
go get gopkg.in/yaml.v3
```

- [ ] **Step 2: Write failing test**

`internal/registry/registry_test.go`:
```go
package registry

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.yaml")
	content := `version: 1
projects:
  - name: order-history
    cwd: ~/Desktop/order-platform-client
    cmd: pnpm order-history dev
  - name: food
    cwd: ~/Desktop/food
    cmd: pnpm dev
    env:
      NODE_ENV: development
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	r, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(r.Projects) != 2 {
		t.Fatalf("Projects len = %d, want 2", len(r.Projects))
	}
	if r.Projects[0].Name != "order-history" {
		t.Errorf("Projects[0].Name = %q", r.Projects[0].Name)
	}
	if r.Projects[1].Env["NODE_ENV"] != "development" {
		t.Errorf("Projects[1].Env[NODE_ENV] = %q", r.Projects[1].Env["NODE_ENV"])
	}
}

func TestLoad_MissingFile(t *testing.T) {
	r, err := Load(filepath.Join(t.TempDir(), "nope.yaml"))
	if err != nil {
		t.Fatalf("Load on missing should return empty registry, got err: %v", err)
	}
	if len(r.Projects) != 0 {
		t.Errorf("missing registry should be empty, got %d", len(r.Projects))
	}
}
```

- [ ] **Step 3: Run, see fail**

```bash
go test ./internal/registry/
```
Expected: FAIL (package undefined).

- [ ] **Step 4: Implement Load**

`internal/registry/registry.go`:
```go
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
```

- [ ] **Step 5: Run, see pass**

```bash
go test ./internal/registry/
```
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add go.mod go.sum internal/registry/
git commit -m "feat(registry): add Project/Registry types and Load"
```

---

## Task 4: Registry — Save + Add + Validate

**Files:**
- Modify: `internal/registry/registry.go`
- Modify: `internal/registry/registry_test.go`

- [ ] **Step 1: Add tests**

Append to `internal/registry/registry_test.go`:
```go
func TestSaveLoad_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "registry.yaml")
	r := &Registry{Version: 1, Projects: []Project{{Name: "a", Cwd: "~/a", Cmd: "echo a"}}}
	if err := r.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	r2, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(r2.Projects) != 1 || r2.Projects[0].Name != "a" {
		t.Errorf("roundtrip mismatch: %+v", r2)
	}
}

func TestAdd_Duplicate(t *testing.T) {
	r := &Registry{Version: 1}
	if err := r.Add(Project{Name: "a", Cwd: "/x", Cmd: "x"}); err != nil {
		t.Fatalf("first Add: %v", err)
	}
	err := r.Add(Project{Name: "a", Cwd: "/y", Cmd: "y"})
	if err == nil {
		t.Errorf("duplicate Add should error")
	}
}

func TestAdd_Validation(t *testing.T) {
	r := &Registry{Version: 1}
	cases := []Project{
		{Name: "", Cwd: "/x", Cmd: "x"},
		{Name: "a", Cwd: "", Cmd: "x"},
		{Name: "a", Cwd: "/x", Cmd: ""},
	}
	for i, p := range cases {
		if err := r.Add(p); err == nil {
			t.Errorf("case %d: expected validation error, got nil", i)
		}
	}
}

func TestRemove(t *testing.T) {
	r := &Registry{Version: 1, Projects: []Project{{Name: "a", Cwd: "/x", Cmd: "x"}}}
	if err := r.Remove("a"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if len(r.Projects) != 0 {
		t.Errorf("Remove did not delete")
	}
	if err := r.Remove("a"); err == nil {
		t.Errorf("Remove missing should error")
	}
}
```

- [ ] **Step 2: Run, see fail**

```bash
go test ./internal/registry/
```
Expected: FAIL (Save/Add/Remove undefined).

- [ ] **Step 3: Implement Save/Add/Remove/validation**

Append to `internal/registry/registry.go`:
```go
import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// (keep existing imports — show full file)

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
```

Final file should have all imports merged correctly (single `import` block at top).

- [ ] **Step 4: Run, see pass**

```bash
go test ./internal/registry/
```
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/registry/
git commit -m "feat(registry): add Save/Add/Replace/Remove with validation"
```

---

## Task 5: CLI Skeleton (Cobra Root + main)

**Files:**
- Create: `cmd/devs/main.go`
- Create: `internal/cli/root.go`

- [ ] **Step 1: Add cobra dependency**

```bash
go get github.com/spf13/cobra
```

- [ ] **Step 2: Implement root command**

`internal/cli/root.go`:
```go
package cli

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "devs",
	Short: "devs — local dev server dashboard",
	Long:  "A TUI for managing registered local dev servers. Run with no args to launch the dashboard.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// TUI launch wired up in a later task.
		cmd.Println("(TUI not yet implemented — use `devs register --help`)")
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}
```

- [ ] **Step 3: Implement main**

`cmd/devs/main.go`:
```go
package main

import (
	"fmt"
	"os"

	"github.com/proshy/devs/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
```

- [ ] **Step 4: Verify build**

```bash
go build -o bin/devs ./cmd/devs
./bin/devs --help
```
Expected: cobra help text shows `devs` description.

- [ ] **Step 5: Commit**

```bash
git add cmd/devs/ internal/cli/ go.mod go.sum
git commit -m "feat(cli): add cobra skeleton and main entry"
```

---

## Task 6: `devs register` Subcommand

**Files:**
- Create: `internal/cli/register.go`
- Create: `internal/cli/register_test.go`

- [ ] **Step 1: Write test**

`internal/cli/register_test.go`:
```go
package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/proshy/devs/internal/registry"
)

func TestRegister_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.yaml")
	if err := registerProject(regPath, registry.Project{
		Name: "myapp",
		Cwd:  "/tmp/myapp",
		Cmd:  "echo hi",
	}, false); err != nil {
		t.Fatalf("registerProject: %v", err)
	}
	r, err := registry.Load(regPath)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(r.Projects) != 1 || r.Projects[0].Name != "myapp" {
		t.Errorf("registry contents: %+v", r.Projects)
	}
}

func TestRegister_DuplicateWithoutForce(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.yaml")
	p := registry.Project{Name: "a", Cwd: "/x", Cmd: "x"}
	if err := registerProject(regPath, p, false); err != nil {
		t.Fatal(err)
	}
	if err := registerProject(regPath, p, false); err == nil {
		t.Errorf("expected duplicate error")
	}
}

func TestRegister_DuplicateWithForce(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.yaml")
	if err := registerProject(regPath, registry.Project{Name: "a", Cwd: "/x", Cmd: "x"}, false); err != nil {
		t.Fatal(err)
	}
	if err := registerProject(regPath, registry.Project{Name: "a", Cwd: "/y", Cmd: "y"}, true); err != nil {
		t.Fatalf("force replace failed: %v", err)
	}
	r, _ := registry.Load(regPath)
	if r.Projects[0].Cmd != "y" {
		t.Errorf("force did not replace, got cmd=%q", r.Projects[0].Cmd)
	}
}

func TestRegister_CwdMustExist(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.yaml")
	err := registerProject(regPath, registry.Project{
		Name: "a", Cwd: filepath.Join(dir, "does-not-exist"), Cmd: "x",
	}, false)
	if err == nil {
		t.Errorf("expected error for non-existent cwd")
	}
}

func TestRegister_ExpandTilde(t *testing.T) {
	home, _ := os.UserHomeDir()
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.yaml")
	if err := registerProject(regPath, registry.Project{
		Name: "a", Cwd: "~", Cmd: "x",
	}, false); err != nil {
		t.Fatalf("register: %v", err)
	}
	r, _ := registry.Load(regPath)
	if r.Projects[0].Cwd != home {
		t.Errorf("tilde not expanded, got %q want %q", r.Projects[0].Cwd, home)
	}
}
```

- [ ] **Step 2: Run, see fail**

```bash
go test ./internal/cli/
```
Expected: FAIL.

- [ ] **Step 3: Implement registerProject + cobra subcommand**

`internal/cli/register.go`:
```go
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/proshy/devs/internal/config"
	"github.com/proshy/devs/internal/registry"
)

var registerFlags struct {
	Name  string
	Cwd   string
	Cmd   string
	Force bool
}

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a project in the devs registry",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := config.New()
		if err := paths.EnsureDirs(); err != nil {
			return err
		}
		return registerProject(paths.RegistryFile, registry.Project{
			Name: registerFlags.Name,
			Cwd:  registerFlags.Cwd,
			Cmd:  registerFlags.Cmd,
		}, registerFlags.Force)
	},
}

func init() {
	registerCmd.Flags().StringVar(&registerFlags.Name, "name", "", "project name (required)")
	registerCmd.Flags().StringVar(&registerFlags.Cwd, "cwd", "", "working directory (required)")
	registerCmd.Flags().StringVar(&registerFlags.Cmd, "cmd", "", "command to start dev server (required)")
	registerCmd.Flags().BoolVar(&registerFlags.Force, "force", false, "replace if name already exists")
	_ = registerCmd.MarkFlagRequired("name")
	_ = registerCmd.MarkFlagRequired("cwd")
	_ = registerCmd.MarkFlagRequired("cmd")
	rootCmd.AddCommand(registerCmd)
}

func registerProject(registryPath string, p registry.Project, force bool) error {
	cwd, err := expandPath(p.Cwd)
	if err != nil {
		return err
	}
	if _, err := os.Stat(cwd); err != nil {
		return fmt.Errorf("cwd does not exist: %s", cwd)
	}
	p.Cwd = cwd

	r, err := registry.Load(registryPath)
	if err != nil {
		return err
	}
	if force {
		if _, ok := r.Find(p.Name); ok {
			if err := r.Replace(p); err != nil {
				return err
			}
		} else if err := r.Add(p); err != nil {
			return err
		}
	} else {
		if err := r.Add(p); err != nil {
			return err
		}
	}
	if err := r.Save(registryPath); err != nil {
		return err
	}
	fmt.Printf("registered %q (cwd=%s, cmd=%q)\n", p.Name, p.Cwd, p.Cmd)
	return nil
}

func expandPath(p string) (string, error) {
	if strings.HasPrefix(p, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~"))
	}
	return filepath.Abs(p)
}
```

- [ ] **Step 4: Run, see pass**

```bash
go test ./internal/cli/
```
Expected: PASS (5 tests).

- [ ] **Step 5: Smoke test the binary**

```bash
go build -o bin/devs ./cmd/devs
HOME_BACKUP=$HOME
export HOME=$(mktemp -d)
./bin/devs register --name demo --cwd "$HOME" --cmd "sleep 1"
cat "$HOME/.config/devs/registry.yaml"
export HOME=$HOME_BACKUP
```
Expected: stdout `registered "demo" ...`; yaml file shows the demo project.

- [ ] **Step 6: Commit**

```bash
git add internal/cli/
git commit -m "feat(cli): add devs register subcommand"
```

---

## Task 7: `devs ls` Subcommand

**Files:**
- Create: `internal/cli/ls.go`

- [ ] **Step 1: Implement**

`internal/cli/ls.go`:
```go
package cli

import (
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/proshy/devs/internal/config"
	"github.com/proshy/devs/internal/registry"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List registered projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		paths := config.New()
		r, err := registry.Load(paths.RegistryFile)
		if err != nil {
			return err
		}
		if len(r.Projects) == 0 {
			fmt.Println("no projects registered")
			return nil
		}
		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "NAME\tCWD\tCMD")
		for _, p := range r.Projects {
			fmt.Fprintf(w, "%s\t%s\t%s\n", p.Name, p.Cwd, p.Cmd)
		}
		return w.Flush()
	},
}

func init() {
	rootCmd.AddCommand(lsCmd)
}
```

- [ ] **Step 2: Smoke test**

```bash
go build -o bin/devs ./cmd/devs
./bin/devs ls
```
Expected: shows your real registered projects (or "no projects registered" if empty).

- [ ] **Step 3: Commit**

```bash
git add internal/cli/ls.go
git commit -m "feat(cli): add devs ls subcommand"
```

---

## Task 8: State Store + Lock

**Files:**
- Create: `internal/state/state.go`
- Create: `internal/state/state_test.go`

- [ ] **Step 1: Write test**

`internal/state/state_test.go`:
```go
package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveLoad(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "state.json")
	s := &State{Managed: map[string]Managed{
		"order-history": {PID: 12345, StartedAt: time.Unix(1700000000, 0), LogPath: "/tmp/x.log"},
	}}
	if err := s.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	s2, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	got := s2.Managed["order-history"]
	if got.PID != 12345 || got.LogPath != "/tmp/x.log" {
		t.Errorf("roundtrip mismatch: %+v", got)
	}
}

func TestLoad_MissingReturnsEmpty(t *testing.T) {
	s, err := Load(filepath.Join(t.TempDir(), "nope"))
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(s.Managed) != 0 {
		t.Errorf("expected empty, got %d", len(s.Managed))
	}
}

func TestAcquireLock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lock")
	rel, err := AcquireLock(path)
	if err != nil {
		t.Fatalf("AcquireLock: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("lock file not created")
	}
	if err := rel(); err != nil {
		t.Fatalf("release: %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("lock file not removed")
	}
}

func TestAcquireLock_DoubleHeld(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "lock")
	rel, err := AcquireLock(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = rel() })

	if _, err := AcquireLock(path); err == nil {
		t.Errorf("second AcquireLock should fail while held")
	}
}
```

- [ ] **Step 2: Run, see fail**

```bash
go test ./internal/state/
```
Expected: FAIL.

- [ ] **Step 3: Implement**

`internal/state/state.go`:
```go
package state

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
	"time"
)

type Managed struct {
	PID       int       `json:"pid"`
	StartedAt time.Time `json:"started_at"`
	LogPath   string    `json:"log_path"`
}

type State struct {
	Managed map[string]Managed `json:"managed"`
}

func Load(path string) (*State, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return &State{Managed: map[string]Managed{}}, nil
		}
		return nil, err
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		// Corrupt: back up and start fresh.
		_ = os.Rename(path, path+".bak")
		return &State{Managed: map[string]Managed{}}, nil
	}
	if s.Managed == nil {
		s.Managed = map[string]Managed{}
	}
	return &s, nil
}

func (s *State) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// AcquireLock writes current PID into path. Returns a release func.
// Fails if path exists and the recorded PID is still alive.
func AcquireLock(path string) (func() error, error) {
	if data, err := os.ReadFile(path); err == nil {
		if pid, err := strconv.Atoi(string(data)); err == nil && pidAlive(pid) {
			return nil, fmt.Errorf("another devs instance is running (pid=%d)", pid)
		}
		// stale lock — overwrite
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		return nil, err
	}
	return func() error { return os.Remove(path) }, nil
}

func pidAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds; use signal 0 to check.
	return p.Signal(syscall.Signal(0)) == nil
}
```

- [ ] **Step 4: Run, see pass**

```bash
go test ./internal/state/
```
Expected: PASS (4 tests).

- [ ] **Step 5: Commit**

```bash
git add internal/state/
git commit -m "feat(state): add runtime state store and single-instance lock"
```

---

## Task 9: Discovery — lsof Wrapper

**Files:**
- Create: `internal/discovery/lsof.go`
- Create: `internal/discovery/lsof_test.go`

- [ ] **Step 1: Capture sample lsof output**

```bash
lsof -nP -iTCP -sTCP:LISTEN -F pcn > internal/discovery/testdata/lsof_sample.txt 2>/dev/null || true
mkdir -p internal/discovery/testdata
lsof -nP -iTCP -sTCP:LISTEN -F pcn > internal/discovery/testdata/lsof_sample.txt
head -20 internal/discovery/testdata/lsof_sample.txt
```
(If `head` shows no `p`/`c`/`n` prefixed lines, your lsof version differs — pass `-F` differently. Adjust `Parse` accordingly.)

- [ ] **Step 2: Write test (parser-only, using fixture)**

`internal/discovery/lsof_test.go`:
```go
package discovery

import (
	"strings"
	"testing"
)

func TestParseLsof(t *testing.T) {
	input := strings.Join([]string{
		"p123",
		"cnode",
		"PTCP",
		"n*:5173",
		"p123",
		"PTCP",
		"n*:5173",
		"p456",
		"cnode",
		"PTCP",
		"n127.0.0.1:4992",
	}, "\n") + "\n"
	got := parseLsof(input)
	if len(got) != 2 {
		t.Fatalf("got %d listeners, want 2: %+v", got, got)
	}
	m := map[int]Listener{}
	for _, l := range got {
		m[l.PID] = l
	}
	if m[123].Port != 5173 || m[123].Cmd != "node" {
		t.Errorf("pid 123 = %+v", m[123])
	}
	if m[456].Port != 4992 {
		t.Errorf("pid 456 = %+v", m[456])
	}
}
```

- [ ] **Step 3: Run, see fail**

```bash
go test ./internal/discovery/
```
Expected: FAIL.

- [ ] **Step 4: Implement**

`internal/discovery/lsof.go`:
```go
package discovery

import (
	"os/exec"
	"strconv"
	"strings"
)

type Listener struct {
	PID  int
	Cmd  string
	Port int
}

// ListListeners shells out to lsof and returns TCP LISTEN entries (deduped per PID+port).
func ListListeners() ([]Listener, error) {
	out, err := exec.Command("lsof", "-nP", "-iTCP", "-sTCP:LISTEN", "-F", "pcn").Output()
	if err != nil {
		return nil, err
	}
	return parseLsof(string(out)), nil
}

func parseLsof(s string) []Listener {
	var pid int
	var cmd string
	seen := map[string]bool{}
	var out []Listener
	for _, line := range strings.Split(s, "\n") {
		if len(line) < 2 {
			continue
		}
		switch line[0] {
		case 'p':
			pid, _ = strconv.Atoi(line[1:])
			cmd = ""
		case 'c':
			cmd = line[1:]
		case 'n':
			// strip host: "*:5173", "127.0.0.1:5173", "[::1]:5173"
			n := line[1:]
			if i := strings.LastIndex(n, ":"); i >= 0 {
				n = n[i+1:]
			}
			port, err := strconv.Atoi(n)
			if err != nil {
				continue
			}
			key := strconv.Itoa(pid) + ":" + strconv.Itoa(port)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, Listener{PID: pid, Cmd: cmd, Port: port})
		}
	}
	return out
}
```

- [ ] **Step 5: Run, see pass**

```bash
go test ./internal/discovery/
```
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/discovery/lsof.go internal/discovery/lsof_test.go
git commit -m "feat(discovery): add lsof wrapper for LISTEN ports"
```

---

## Task 10: Discovery — ps Wrapper (cwd + cmdline)

**Files:**
- Create: `internal/discovery/ps.go`

- [ ] **Step 1: Implement**

`internal/discovery/ps.go`:
```go
package discovery

import (
	"bufio"
	"os/exec"
	"strconv"
	"strings"
)

type ProcInfo struct {
	PID     int
	Cwd     string
	Cmdline string
}

// GetCwd returns the working directory of a process via lsof.
func GetCwd(pid int) string {
	out, err := exec.Command("lsof", "-a", "-p", strconv.Itoa(pid), "-d", "cwd", "-Fn").Output()
	if err != nil {
		return ""
	}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	for sc.Scan() {
		line := sc.Text()
		if len(line) > 1 && line[0] == 'n' {
			return line[1:]
		}
	}
	return ""
}

// GetCmdline returns the full command line via ps.
func GetCmdline(pid int) string {
	out, err := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func GetProcInfo(pid int) ProcInfo {
	return ProcInfo{PID: pid, Cwd: GetCwd(pid), Cmdline: GetCmdline(pid)}
}
```

- [ ] **Step 2: Smoke test**

```bash
cat > /tmp/pstest.go <<'EOF'
package main
import (
  "fmt"
  "os"
  "github.com/proshy/devs/internal/discovery"
)
func main() {
  pid := os.Getpid()
  info := discovery.GetProcInfo(pid)
  fmt.Printf("self: pid=%d cwd=%q cmd=%q\n", info.PID, info.Cwd, info.Cmdline)
}
EOF
go run /tmp/pstest.go
rm /tmp/pstest.go
```
Expected: prints own PID with cwd ending in `Desktop/devs` and cmdline containing `/tmp/pstest`.

- [ ] **Step 3: Commit**

```bash
git add internal/discovery/ps.go
git commit -m "feat(discovery): add ps/lsof helpers for cwd and cmdline"
```

---

## Task 11: Discovery — Matcher

**Files:**
- Create: `internal/discovery/matcher.go`
- Create: `internal/discovery/matcher_test.go`

- [ ] **Step 1: Write test**

`internal/discovery/matcher_test.go`:
```go
package discovery

import (
	"testing"

	"github.com/proshy/devs/internal/registry"
)

func TestMatch_ByCwdAndCmdToken(t *testing.T) {
	projects := []registry.Project{
		{Name: "order-history", Cwd: "/home/u/order-platform-client", Cmd: "pnpm order-history dev"},
		{Name: "food", Cwd: "/home/u/food", Cmd: "pnpm dev"},
	}
	listeners := []Listener{
		{PID: 100, Cmd: "node", Port: 5173},
		{PID: 200, Cmd: "node", Port: 5174},
	}
	proc := func(pid int) ProcInfo {
		switch pid {
		case 100:
			return ProcInfo{PID: 100, Cwd: "/home/u/order-platform-client/apps/order-history", Cmdline: "node /usr/bin/pnpm order-history dev"}
		case 200:
			return ProcInfo{PID: 200, Cwd: "/home/u/food", Cmdline: "node /usr/bin/pnpm dev"}
		}
		return ProcInfo{}
	}

	matches := Match(projects, listeners, proc)
	if m, ok := matches["order-history"]; !ok || m.Port != 5173 {
		t.Errorf("order-history match wrong: %+v", matches)
	}
	if m, ok := matches["food"]; !ok || m.Port != 5174 {
		t.Errorf("food match wrong: %+v", matches)
	}
}

func TestMatch_NoMatchWhenCwdDifferent(t *testing.T) {
	projects := []registry.Project{
		{Name: "a", Cwd: "/x", Cmd: "pnpm dev"},
	}
	listeners := []Listener{{PID: 1, Cmd: "node", Port: 5000}}
	proc := func(pid int) ProcInfo {
		return ProcInfo{PID: 1, Cwd: "/y", Cmdline: "node pnpm dev"}
	}
	matches := Match(projects, listeners, proc)
	if _, ok := matches["a"]; ok {
		t.Errorf("expected no match")
	}
}
```

- [ ] **Step 2: Run, see fail**

```bash
go test ./internal/discovery/
```
Expected: FAIL.

- [ ] **Step 3: Implement**

`internal/discovery/matcher.go`:
```go
package discovery

import (
	"strings"

	"github.com/proshy/devs/internal/registry"
)

type Match struct {
	PID  int
	Port int
}

// Match returns name -> Match for projects whose cwd contains a LISTEN PID with matching cmd tokens.
// procFn is injected so tests can stub ps/lsof lookups.
func Match(projects []registry.Project, listeners []Listener, procFn func(pid int) ProcInfo) map[string]Match {
	out := map[string]Match{}
	for _, l := range listeners {
		info := procFn(l.PID)
		for _, p := range projects {
			if _, ok := out[p.Name]; ok {
				continue
			}
			if !strings.HasPrefix(info.Cwd, p.Cwd) {
				continue
			}
			if !cmdTokenOverlap(p.Cmd, info.Cmdline) {
				continue
			}
			out[p.Name] = Match{PID: l.PID, Port: l.Port}
		}
	}
	return out
}

// cmdTokenOverlap returns true if the registered cmd's distinctive tokens appear in the live cmdline.
// "Distinctive" = tokens of length >= 3 that aren't trivial words.
func cmdTokenOverlap(registered, live string) bool {
	tokens := strings.Fields(registered)
	for _, t := range tokens {
		if len(t) < 3 {
			continue
		}
		if t == "run" || t == "dev" || t == "start" || t == "pnpm" || t == "npm" || t == "yarn" {
			continue
		}
		if strings.Contains(live, t) {
			return true
		}
	}
	// fallback: if no distinctive token, accept first non-trivial token match
	for _, t := range tokens {
		if strings.Contains(live, t) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: Run, see pass**

```bash
go test ./internal/discovery/
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/discovery/matcher.go internal/discovery/matcher_test.go
git commit -m "feat(discovery): match registry projects to live listeners by cwd+cmd"
```

---

## Task 12: Process Runner — Start

**Files:**
- Create: `internal/process/runner.go`
- Create: `internal/process/runner_test.go`

- [ ] **Step 1: Write test**

`internal/process/runner_test.go`:
```go
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
	// give it a moment to write the log
	time.Sleep(200 * time.Millisecond)
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("log read: %v", err)
	}
	if string(data) == "" || string(data)[:5] != "hello" {
		t.Errorf("log = %q, want starts with 'hello'", string(data))
	}
}
```

- [ ] **Step 2: Run, see fail**

```bash
go test ./internal/process/
```
Expected: FAIL.

- [ ] **Step 3: Implement Start**

`internal/process/runner.go`:
```go
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

	// Use /bin/sh so users can write shell-style cmds with pipes/env, matching what they'd
	// type in a terminal.
	cmd := exec.Command("/bin/sh", "-c", p.Cmd)
	cmd.Dir = p.Cwd
	cmd.Stdout = logF
	cmd.Stderr = logF
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true, // new session → dev server survives our exit
	}
	cmd.Env = os.Environ()
	for k, v := range p.Env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Write a banner so tail viewers can see when each session started.
	_, _ = fmt.Fprintf(logF, "\n--- devs start %s ---\n", time.Now().Format(time.RFC3339))

	if err := cmd.Start(); err != nil {
		_ = logF.Close()
		return 0, err
	}
	// Release child so we don't keep it as a zombie.
	go func() { _ = cmd.Wait() }()
	return cmd.Process.Pid, nil
}
```

- [ ] **Step 4: Run, see pass**

```bash
go test ./internal/process/
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/process/
git commit -m "feat(process): add detached Start with log capture"
```

---

## Task 13: Process Runner — Stop

**Files:**
- Modify: `internal/process/runner.go`
- Modify: `internal/process/runner_test.go`

- [ ] **Step 1: Append test**

Append to `internal/process/runner_test.go`:
```go
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
	// after Stop returns, process should be gone
	time.Sleep(100 * time.Millisecond)
	if err := syscall.Kill(pid, syscall.Signal(0)); err == nil {
		t.Errorf("pid %d still alive", pid)
	}
}
```

- [ ] **Step 2: Run, see fail**

```bash
go test ./internal/process/
```
Expected: FAIL.

- [ ] **Step 3: Implement Stop**

Append to `internal/process/runner.go`:
```go
// Stop sends SIGTERM to the process group. If still alive after timeout, sends SIGKILL.
func Stop(pid int, gracePeriod time.Duration) error {
	// Negative PID = process group (because we used Setsid).
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		if errors.Is(err, syscall.ESRCH) {
			return nil
		}
		return err
	}
	deadline := time.Now().Add(gracePeriod)
	for time.Now().Before(deadline) {
		if syscall.Kill(pid, syscall.Signal(0)) != nil {
			return nil // gone
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
```

- [ ] **Step 4: Run, see pass**

```bash
go test ./internal/process/
```
Expected: PASS.

- [ ] **Step 5: Commit**

```bash
git add internal/process/
git commit -m "feat(process): add graceful Stop and IsAlive"
```

---

## Task 14: Git Enricher

**Files:**
- Create: `internal/git/git.go`

- [ ] **Step 1: Implement**

`internal/git/git.go`:
```go
package git

import (
	"os/exec"
	"strings"
)

// Branch returns the current branch in cwd, or "" if not a git repo.
func Branch(cwd string) string {
	out, err := exec.Command("git", "-C", cwd, "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// IsDirty returns true if the working tree has uncommitted changes.
func IsDirty(cwd string) bool {
	out, err := exec.Command("git", "-C", cwd, "status", "--porcelain").Output()
	if err != nil {
		return false
	}
	return len(strings.TrimSpace(string(out))) > 0
}
```

- [ ] **Step 2: Smoke test**

```bash
cat > /tmp/gtest.go <<'EOF'
package main
import (
	"fmt"
	"github.com/proshy/devs/internal/git"
)
func main() {
	fmt.Println("branch:", git.Branch("."))
	fmt.Println("dirty:", git.IsDirty("."))
}
EOF
go run /tmp/gtest.go
rm /tmp/gtest.go
```
Expected: prints `branch: main` (or your branch) and `dirty: true` (since the plan is uncommitted at this moment) or `false` after commit.

- [ ] **Step 3: Commit**

```bash
git add internal/git/
git commit -m "feat(git): add Branch and IsDirty helpers"
```

---

## Task 15: TUI Keys + Styles

**Files:**
- Create: `internal/tui/keys.go`
- Create: `internal/tui/styles.go`

- [ ] **Step 1: Add Bubbletea/Lipgloss deps**

```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/bubbles/table
go get github.com/charmbracelet/lipgloss
```

- [ ] **Step 2: Implement keys**

`internal/tui/keys.go`:
```go
package tui

import "github.com/charmbracelet/bubbles/key"

type keymap struct {
	Up      key.Binding
	Down    key.Binding
	Toggle  key.Binding
	Start   key.Binding
	Stop    key.Binding
	Restart key.Binding
	Log     key.Binding
	Add     key.Binding
	Edit    key.Binding
	Refresh key.Binding
	Quit    key.Binding
	QuitAll key.Binding
}

func defaultKeys() keymap {
	return keymap{
		Up:      key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "up")),
		Down:    key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "down")),
		Toggle:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "toggle")),
		Start:   key.NewBinding(key.WithKeys("s"), key.WithHelp("s", "start")),
		Stop:    key.NewBinding(key.WithKeys("x"), key.WithHelp("x", "stop")),
		Restart: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "restart")),
		Log:     key.NewBinding(key.WithKeys("l"), key.WithHelp("l", "log")),
		Add:     key.NewBinding(key.WithKeys("a"), key.WithHelp("a", "add")),
		Edit:    key.NewBinding(key.WithKeys("e"), key.WithHelp("e", "edit registry")),
		Refresh: key.NewBinding(key.WithKeys("R"), key.WithHelp("R", "refresh")),
		Quit:    key.NewBinding(key.WithKeys("q"), key.WithHelp("q", "quit")),
		QuitAll: key.NewBinding(key.WithKeys("Q"), key.WithHelp("Q", "quit+kill all")),
	}
}
```

- [ ] **Step 3: Implement styles**

`internal/tui/styles.go`:
```go
package tui

import "github.com/charmbracelet/lipgloss"

var (
	stateOn  = lipgloss.NewStyle().Foreground(lipgloss.Color("#10b981")).SetString("● ON ")
	stateOff = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280")).SetString("○ OFF")
	dirty    = lipgloss.NewStyle().Foreground(lipgloss.Color("#f59e0b")).SetString("★")
	header   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#3b82f6"))
	help     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6b7280"))
)
```

- [ ] **Step 4: Build check**

```bash
go build ./internal/tui/...
```
Expected: no errors.

- [ ] **Step 5: Commit**

```bash
git add internal/tui/ go.mod go.sum
git commit -m "feat(tui): add keymap and lipgloss styles"
```

---

## Task 16: TUI App Skeleton

**Files:**
- Create: `internal/tui/app.go`

- [ ] **Step 1: Implement minimal Bubbletea model**

`internal/tui/app.go`:
```go
package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/proshy/devs/internal/config"
)

type model struct {
	paths  config.Paths
	keys   keymap
	status string
	quit   bool
}

func New() *model {
	return &model{
		paths: config.New(),
		keys:  defaultKeys(),
	}
}

func (m *model) Init() tea.Cmd { return nil }

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, m.keys.Quit, m.keys.QuitAll) {
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *model) View() string {
	return fmt.Sprintf("%s\n\n%s\n\n%s",
		header.Render("devs (skeleton)"),
		"press q to quit",
		help.Render("q: quit  Q: quit+kill all"),
	)
}

// Run starts the program.
func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```

Add missing import at top: `"github.com/charmbracelet/bubbles/key"` — full imports:
```go
import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/proshy/devs/internal/config"
)
```

- [ ] **Step 2: Wire TUI into root command**

Edit `internal/cli/root.go` — replace the `RunE` body:
```go
RunE: func(cmd *cobra.Command, args []string) error {
    return tui.Run()
},
```
Add import `"github.com/proshy/devs/internal/tui"`.

- [ ] **Step 3: Smoke test**

```bash
go build -o bin/devs ./cmd/devs
./bin/devs
```
Expected: alt-screen TUI with "devs (skeleton)" and help. `q` exits.

- [ ] **Step 4: Commit**

```bash
git add internal/tui/ internal/cli/root.go
git commit -m "feat(tui): minimal Bubbletea app with quit handling"
```

---

## Task 17: TUI Table View (registry-only, static)

**Files:**
- Create: `internal/tui/table.go`
- Modify: `internal/tui/app.go`

- [ ] **Step 1: Build table component**

`internal/tui/table.go`:
```go
package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/lipgloss"

	"github.com/proshy/devs/internal/discovery"
	"github.com/proshy/devs/internal/git"
	"github.com/proshy/devs/internal/registry"
)

func newTable() table.Model {
	cols := []table.Column{
		{Title: "NAME", Width: 18},
		{Title: "STATE", Width: 6},
		{Title: "PORT", Width: 6},
		{Title: "BRANCH", Width: 16},
		{Title: "CMD", Width: 40},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithHeight(15),
		table.WithFocused(true),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.Bold(true).Foreground(lipgloss.Color("#3b82f6"))
	s.Selected = s.Selected.Foreground(lipgloss.Color("#fff")).Background(lipgloss.Color("#1e3a8a"))
	t.SetStyles(s)
	return t
}

// rowFor builds a table row for a project, given a match (may be empty).
func rowFor(p registry.Project, m discovery.Match) table.Row {
	state := stateOff.String()
	port := "—"
	if m.PID != 0 {
		state = stateOn.String()
		port = fmt.Sprintf("%d", m.Port)
	}
	branch := git.Branch(p.Cwd)
	if branch == "" {
		branch = "—"
	} else if git.IsDirty(p.Cwd) {
		branch += dirty.String()
	}
	cmd := p.Cmd
	if len(cmd) > 40 {
		cmd = cmd[:37] + "…"
	}
	return table.Row{p.Name, state, port, branch, cmd}
}
```

- [ ] **Step 2: Replace app.go with table-enabled model**

`internal/tui/app.go`:
```go
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/proshy/devs/internal/config"
	"github.com/proshy/devs/internal/discovery"
	"github.com/proshy/devs/internal/registry"
)

type model struct {
	paths   config.Paths
	keys    keymap
	reg     *registry.Registry
	tbl     table.Model
	matches map[string]discovery.Match
	err     error
	status  string
}

func New() *model {
	m := &model{
		paths:   config.New(),
		keys:    defaultKeys(),
		tbl:     newTable(),
		matches: map[string]discovery.Match{},
	}
	m.reloadRegistry()
	m.refresh()
	return m
}

func (m *model) reloadRegistry() {
	r, err := registry.Load(m.paths.RegistryFile)
	if err != nil {
		m.err = err
		return
	}
	m.reg = r
}

func (m *model) refresh() {
	if m.reg == nil {
		return
	}
	listeners, err := discovery.ListListeners()
	if err != nil {
		m.status = "discovery error: " + err.Error()
	} else {
		m.matches = discovery.Match(m.reg.Projects, listeners, discovery.GetProcInfo)
	}
	var rows []table.Row
	for _, p := range m.reg.Projects {
		rows = append(rows, rowFor(p, m.matches[p.Name]))
	}
	m.tbl.SetRows(rows)
}

func (m *model) Init() tea.Cmd { return tea.Batch(tickCmd()) }

type tickMsg struct{}

func tickCmd() tea.Cmd {
	return tea.Tick(2_000_000_000, func(_ any) tea.Msg { return tickMsg{} })
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.QuitAll):
			return m, tea.Quit
		case key.Matches(msg, m.keys.Refresh):
			m.refresh()
		}
	case tickMsg:
		m.refresh()
		return m, tickCmd()
	}
	var cmd tea.Cmd
	m.tbl, cmd = m.tbl.Update(msg)
	return m, cmd
}

func (m *model) View() string {
	var b strings.Builder
	b.WriteString(header.Render(" devs "))
	b.WriteString("\n\n")
	b.WriteString(m.tbl.View())
	b.WriteString("\n")
	if m.status != "" {
		b.WriteString(help.Render(m.status))
		b.WriteString("\n")
	}
	b.WriteString(help.Render(" enter:toggle  s:start  x:stop  r:restart  l:log  a:add  e:edit  R:refresh  q:quit "))
	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(fmt.Sprintf("ERROR: %v", m.err))
	}
	return b.String()
}

func Run() error {
	p := tea.NewProgram(New(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
```

Note: replace `2_000_000_000` with `2*time.Second` and add `"time"` import. Imports needed (final list):
```go
import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/proshy/devs/internal/config"
	"github.com/proshy/devs/internal/discovery"
	"github.com/proshy/devs/internal/registry"
)
```
And `tickCmd`:
```go
func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(_ time.Time) tea.Msg { return tickMsg{} })
}
```

- [ ] **Step 3: Smoke test**

```bash
go build -o bin/devs ./cmd/devs
./bin/devs ls          # verify registry not empty (register one if so)
./bin/devs             # launch TUI, verify table shows registered projects with ON/OFF
```
Expected: TUI shows registered projects. Manually start one of them in another terminal — within 2s the dashboard flips its row to `● ON`.

- [ ] **Step 4: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): table view with 2s discovery refresh"
```

---

## Task 18: TUI — Start/Stop Actions

**Files:**
- Modify: `internal/tui/app.go`

- [ ] **Step 1: Add state store + process runner integration**

Add fields and methods to `model` in `internal/tui/app.go`:

```go
import (
    // existing imports + add:
    "path/filepath"

    "github.com/proshy/devs/internal/process"
    "github.com/proshy/devs/internal/state"
)

// inside model struct, add:
//   st *state.State
//   release func() error  (lock releaser)
```

In `New()`, after `m.reloadRegistry()`:
```go
release, err := state.AcquireLock(m.paths.LockFile)
if err != nil {
    m.err = err
    return m
}
m.release = release
s, _ := state.Load(m.paths.StateFile)
m.st = s
```

Add method:
```go
func (m *model) startSelected() {
    if m.reg == nil || m.tbl.Cursor() >= len(m.reg.Projects) {
        return
    }
    p := m.reg.Projects[m.tbl.Cursor()]
    if _, ok := m.matches[p.Name]; ok {
        m.status = p.Name + " already running"
        return
    }
    logPath := filepath.Join(m.paths.LogsDir, p.Name+".log")
    pid, err := process.Start(p, logPath)
    if err != nil {
        m.status = "start failed: " + err.Error()
        return
    }
    if m.st.Managed == nil {
        m.st.Managed = map[string]state.Managed{}
    }
    m.st.Managed[p.Name] = state.Managed{PID: pid, StartedAt: timeNow(), LogPath: logPath}
    _ = m.st.Save(m.paths.StateFile)
    m.status = fmt.Sprintf("started %s (pid=%d)", p.Name, pid)
    m.refresh()
}

func (m *model) stopSelected() {
    if m.reg == nil || m.tbl.Cursor() >= len(m.reg.Projects) {
        return
    }
    p := m.reg.Projects[m.tbl.Cursor()]
    match, ok := m.matches[p.Name]
    if !ok {
        m.status = p.Name + " is not running"
        return
    }
    if err := process.Stop(match.PID, 3*time.Second); err != nil {
        m.status = "stop failed: " + err.Error()
        return
    }
    delete(m.st.Managed, p.Name)
    _ = m.st.Save(m.paths.StateFile)
    m.status = "stopped " + p.Name
    m.refresh()
}

func timeNow() time.Time { return time.Now() }
```

In `Update()`, add cases inside `tea.KeyMsg`:
```go
case key.Matches(msg, m.keys.Start), key.Matches(msg, m.keys.Toggle):
    if _, on := m.matches[selectedName(m)]; on {
        m.stopSelected()
    } else {
        m.startSelected()
    }
case key.Matches(msg, m.keys.Stop):
    m.stopSelected()
case key.Matches(msg, m.keys.Restart):
    m.stopSelected()
    time.Sleep(200 * time.Millisecond)
    m.startSelected()
```

Helper:
```go
func selectedName(m *model) string {
    if m.reg == nil || m.tbl.Cursor() >= len(m.reg.Projects) {
        return ""
    }
    return m.reg.Projects[m.tbl.Cursor()].Name
}
```

In `Update()` handle `tea.Quit` cleanup: when `QuitAll` is pressed, kill all managed PIDs before quit. Update:
```go
case key.Matches(msg, m.keys.QuitAll):
    for _, mg := range m.st.Managed {
        _ = process.Stop(mg.PID, 1*time.Second)
    }
    if m.release != nil { _ = m.release() }
    return m, tea.Quit
case key.Matches(msg, m.keys.Quit):
    if m.release != nil { _ = m.release() }
    return m, tea.Quit
```

- [ ] **Step 2: Smoke test**

```bash
go build -o bin/devs ./cmd/devs
# register a test project that prints to stdout
./bin/devs register --name dev-smoke --cwd "$HOME" --cmd "/bin/sh -c 'while true; do echo tick; sleep 1; done'" --force
./bin/devs
# In TUI: cursor on dev-smoke, press enter → row flips to ON
# In another terminal: lsof -i | grep <pid>; ps -p <pid>
# Press enter again → row flips to OFF
# Press Q → all managed killed
```

- [ ] **Step 3: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): start/stop/restart with state persistence + cleanup"
```

---

## Task 19: TUI — Log Pane

**Files:**
- Create: `internal/tui/log.go`
- Modify: `internal/tui/app.go`

- [ ] **Step 1: Implement log pane**

`internal/tui/log.go`:
```go
package tui

import (
	"bufio"
	"os"

	"github.com/charmbracelet/bubbles/viewport"
)

type logPane struct {
	vp     viewport.Model
	path   string
	visible bool
}

func newLog() logPane {
	vp := viewport.New(80, 10)
	return logPane{vp: vp}
}

// load tails up to last N lines from path.
func (l *logPane) load(path string) {
	l.path = path
	if path == "" {
		l.vp.SetContent("(no log)")
		return
	}
	f, err := os.Open(path)
	if err != nil {
		l.vp.SetContent("(log not available: " + err.Error() + ")")
		return
	}
	defer f.Close()
	const max = 200
	lines := make([]string, 0, max)
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
		if len(lines) > max {
			lines = lines[1:]
		}
	}
	var content string
	for _, ln := range lines {
		content += ln + "\n"
	}
	l.vp.SetContent(content)
	l.vp.GotoBottom()
}
```

- [ ] **Step 2: Wire into app**

In `internal/tui/app.go`:
- Add `log logPane` field
- In `New()`: `m.log = newLog()`
- In `Update()`:
```go
case key.Matches(msg, m.keys.Log):
    m.log.visible = !m.log.visible
    if m.log.visible {
        name := selectedName(m)
        if mg, ok := m.st.Managed[name]; ok {
            m.log.load(mg.LogPath)
        } else {
            m.log.load("")
        }
    }
```
- In `View()`: append log pane when visible:
```go
if m.log.visible {
    b.WriteString("\n")
    b.WriteString(header.Render(" log "))
    b.WriteString("\n")
    b.WriteString(m.log.vp.View())
}
```

- [ ] **Step 3: Smoke test**

```bash
go build -o bin/devs ./cmd/devs
./bin/devs
# In TUI: start dev-smoke (registered earlier), press l → see "tick" lines
```

- [ ] **Step 4: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): add log pane toggle (last 200 lines)"
```

---

## Task 20: TUI — Add Form

**Files:**
- Create: `internal/tui/form.go`
- Modify: `internal/tui/app.go`

- [ ] **Step 1: Implement form**

`internal/tui/form.go`:
```go
package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type addForm struct {
	visible bool
	inputs  []textinput.Model
	idx     int
}

func newAddForm() addForm {
	mk := func(placeholder string) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.Width = 40
		return ti
	}
	inputs := []textinput.Model{mk("name"), mk("cwd"), mk("cmd")}
	inputs[0].Focus()
	return addForm{inputs: inputs}
}

func (f *addForm) next() {
	f.inputs[f.idx].Blur()
	f.idx = (f.idx + 1) % len(f.inputs)
	f.inputs[f.idx].Focus()
}

func (f *addForm) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	f.inputs[f.idx], cmd = f.inputs[f.idx].Update(msg)
	return cmd
}

func (f *addForm) values() (name, cwd, cmd string) {
	return f.inputs[0].Value(), f.inputs[1].Value(), f.inputs[2].Value()
}

func (f addForm) View() string {
	return "─ Add project ─\n" +
		" name: " + f.inputs[0].View() + "\n" +
		" cwd:  " + f.inputs[1].View() + "\n" +
		" cmd:  " + f.inputs[2].View() + "\n" +
		" tab:next  enter:save  esc:cancel"
}
```

- [ ] **Step 2: Wire into app**

In `internal/tui/app.go`:
- Add field `form addForm`
- In `New()`: `m.form = newAddForm()`
- In `Update()`:
```go
case key.Matches(msg, m.keys.Add):
    m.form.visible = true
    return m, nil
```
- Add early branch when form visible:
```go
if m.form.visible {
    if k, ok := msg.(tea.KeyMsg); ok {
        switch k.String() {
        case "esc":
            m.form.visible = false
            return m, nil
        case "tab":
            m.form.next()
            return m, nil
        case "enter":
            name, cwd, cmd := m.form.values()
            if err := registerInline(m.paths.RegistryFile, name, cwd, cmd); err != nil {
                m.status = "add failed: " + err.Error()
            } else {
                m.status = "added " + name
                m.form.visible = false
                m.reloadRegistry()
                m.refresh()
            }
            return m, nil
        }
    }
    cmd := m.form.Update(msg)
    return m, cmd
}
```
- In `View()`, when `m.form.visible`, replace body with form view.

- [ ] **Step 3: Add helper**

Append to `internal/tui/app.go`:
```go
func registerInline(regPath, name, cwd, cmd string) error {
	// Reuse the cli registration logic via a tiny copy to avoid import cycles.
	r, err := registry.Load(regPath)
	if err != nil {
		return err
	}
	if err := r.Add(registry.Project{Name: name, Cwd: cwd, Cmd: cmd}); err != nil {
		return err
	}
	return r.Save(regPath)
}
```

- [ ] **Step 4: Add textinput dep**

```bash
go get github.com/charmbracelet/bubbles/textinput
```

- [ ] **Step 5: Smoke test**

```bash
go build -o bin/devs ./cmd/devs
./bin/devs
# Press a → form. Fill name/cwd/cmd, press enter → row appears.
```

- [ ] **Step 6: Commit**

```bash
git add internal/tui/ go.mod go.sum
git commit -m "feat(tui): inline add-project form"
```

---

## Task 21: TUI — Edit Registry via $EDITOR

**Files:**
- Modify: `internal/tui/app.go`

- [ ] **Step 1: Implement**

Add to `Update()`:
```go
case key.Matches(msg, m.keys.Edit):
    return m, openEditor(m.paths.RegistryFile)
```

Append helper:
```go
func openEditor(path string) tea.Cmd {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi"
	}
	c := exec.Command(editor, path)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return reloadMsg{err: err}
	})
}

type reloadMsg struct{ err error }
```

In `Update()`, handle `reloadMsg`:
```go
case reloadMsg:
    if msg.err != nil {
        m.status = "editor error: " + msg.err.Error()
    }
    m.reloadRegistry()
    m.refresh()
```

Add imports: `"os"`, `"os/exec"`.

- [ ] **Step 2: Smoke test**

```bash
go build -o bin/devs ./cmd/devs
EDITOR=nano ./bin/devs
# In TUI press e → opens registry.yaml in nano. Save and exit → table reflects edits.
```

- [ ] **Step 3: Commit**

```bash
git add internal/tui/
git commit -m "feat(tui): open registry in $EDITOR with e key"
```

---

## Task 22: Makefile

**Files:**
- Create: `Makefile`

- [ ] **Step 1: Write Makefile**

`Makefile`:
```makefile
BIN_DIR ?= $(HOME)/.local/bin
SKILLS_DIR ?= $(HOME)/.claude/skills
REPO_DIR := $(shell pwd)

.PHONY: build install install-skills uninstall test clean

build:
	@mkdir -p bin
	go build -o bin/devs ./cmd/devs

install: build
	@mkdir -p $(BIN_DIR)
	install -m 0755 bin/devs $(BIN_DIR)/devs
	@echo "installed → $(BIN_DIR)/devs"

install-skills:
	@mkdir -p $(SKILLS_DIR)
	@for d in devs-register devs-help; do \
		rm -rf $(SKILLS_DIR)/$$d; \
		ln -s $(REPO_DIR)/skills/$$d $(SKILLS_DIR)/$$d; \
		echo "linked → $(SKILLS_DIR)/$$d"; \
	done

uninstall:
	rm -f $(BIN_DIR)/devs
	rm -rf $(SKILLS_DIR)/devs-register $(SKILLS_DIR)/devs-help
	@echo "removed binary and skill links (state and registry preserved)"

test:
	go test ./...

clean:
	rm -rf bin
```

- [ ] **Step 2: Verify targets**

```bash
make build
ls bin/devs
make test
```
Expected: build succeeds, tests pass.

- [ ] **Step 3: Commit**

```bash
git add Makefile
git commit -m "build: add Makefile (build, install, install-skills, test, clean)"
```

---

## Task 23: README + INSTALL

**Files:**
- Create: `README.md`
- Create: `INSTALL.md`

- [ ] **Step 1: Write README**

`README.md`:
```markdown
# devs

A terminal dashboard for managing locally registered dev servers — see what's running where, start/stop with one keystroke, and let agents do the registration for you.

## Why

If you juggle multiple dev servers across cmux/tmux panes you lose track of which is running where. `devs` shows them all in one place, recognizes servers started in other panes, and survives its own restart so your work isn't tied to its uptime.

## Features

- **One panel, all your dev servers** — see ports, branches, run state
- **Survives restart** — dashboard exit doesn't kill your servers
- **Detects external** — a server started in another pane shows up automatically (matched by cwd)
- **Agent-friendly** — `/devs-register` skill lets your agent register projects from project context with one sentence
- **Single Go binary** — no daemon

## Quick start

See [INSTALL.md](./INSTALL.md) for step-by-step installation (agent-friendly).

```bash
make install        # → ~/.local/bin/devs
make install-skills # → ~/.claude/skills/devs-{register,help}

# Register a project (manually or via the skill)
devs register --name my-app --cwd ~/path/to/repo --cmd "pnpm dev"

# Launch the dashboard
devs
```

## Keys

| Key | Action |
|-----|--------|
| ↑/↓ or j/k | Move cursor |
| enter | Toggle running/stopped |
| s | Start | 
| x | Stop |
| r | Restart |
| l | Toggle log pane |
| a | Add project (inline form) |
| e | Edit `registry.yaml` in $EDITOR |
| R | Manual refresh |
| q | Quit (servers keep running) |
| Q | Quit + kill all dashboard-spawned |

## Files

- Registry: `~/.config/devs/registry.yaml`
- Runtime state: `~/.local/state/devs/state.json`
- Logs: `~/.local/state/devs/logs/<name>.log`

## Skills

- `~/.claude/skills/devs-register` — agent skill: "서버 등록해줘" → scans project, calls `devs register`
- `~/.claude/skills/devs-help` — usage reference for agents

## Design

See [docs/superpowers/specs/2026-05-22-devs-design.md](./docs/superpowers/specs/2026-05-22-devs-design.md).

## License

MIT (or your choice).
```

- [ ] **Step 2: Write INSTALL.md**

`INSTALL.md`:
```markdown
# Installing devs

This guide is written so an AI agent (Claude Code, etc.) can follow it step-by-step. Each step has a verification command — if it fails, fix that before continuing.

## 1. Prerequisites

### Go 1.22+

```bash
go version
```
Expected: `go version go1.22.x` or higher.

Install if missing:
```bash
brew install go
```

### git, lsof, ps

These ship with macOS by default:
```bash
which git lsof ps
```
Expected: three paths printed.

## 2. Clone

```bash
git clone https://github.com/proshy/devs.git ~/.local/share/devs
cd ~/.local/share/devs
```

## 3. Build + Install

```bash
make install
```
Expected: `installed → /Users/<you>/.local/bin/devs`

Verify:
```bash
which devs
devs --help
```
Expected: cobra help text.

If `which devs` fails, ensure `~/.local/bin` is in `PATH`:
```bash
echo $PATH | tr ':' '\n' | grep local/bin
```

## 4. Install Skills (optional but recommended)

```bash
make install-skills
```
Expected: two symlinks created in `~/.claude/skills/`.

Verify:
```bash
ls -la ~/.claude/skills/ | grep devs
```

## 5. First Registration

In a project root that has a dev server:
```bash
cd ~/path/to/your-project
devs register --name your-project --cwd "$PWD" --cmd "pnpm dev"
```

Or, with the skill installed, just tell your agent:
> 서버 등록해줘

The agent will read `package.json` (or pyproject/build.gradle/etc), pick the right dev script, and call `devs register` for you.

## 6. Launch

```bash
devs
```
Expected: TUI shows your registered project as `○ OFF`. Press `s` to start it. After 2 seconds the row flips to `● ON`.

## Uninstall

```bash
cd ~/.local/share/devs
make uninstall
```
Registry and state files are preserved; remove them manually if desired:
```bash
rm -rf ~/.config/devs ~/.local/state/devs
```
```

- [ ] **Step 3: Commit**

```bash
git add README.md INSTALL.md
git commit -m "docs: add README and agent-friendly INSTALL guide"
```

---

## Task 24: Skill — `devs-register`

**Files:**
- Create: `skills/devs-register/SKILL.md`

- [ ] **Step 1: Write skill**

`skills/devs-register/SKILL.md`:
```markdown
---
name: devs-register
description: Register a project in the devs local-server dashboard. Trigger when the user mentions "devs 등록", "서버 등록", "dev server 등록", or similar requests to add a project to the local dev dashboard.
---

# devs-register

Helps the user add a local project to the `devs` dashboard by detecting how its dev server is run and calling `devs register`.

## Preconditions

Before running, verify:

```bash
which devs
```

If missing, direct the user to [INSTALL.md](https://github.com/proshy/devs/blob/main/INSTALL.md).

## Flow

1. **Find project root** — starting from `$PWD` (or a path the user gave you), walk up until you find one of:
   - `package.json`
   - `pyproject.toml`
   - `build.gradle`, `build.gradle.kts`, `pom.xml`
   - `go.mod`
   - `Cargo.toml`
   - `Makefile`
   - `Procfile`

2. **Identify dev candidates** — based on what you found:

   **Node** (`package.json`):
   - Look at `scripts` field. Common dev names: `dev`, `start`, `serve`, `start:dev`, `dev:server`.
   - For monorepos (`pnpm-workspace.yaml`, `lerna.json`, `nx.json`, workspaces in package.json): check each package's `scripts` and prefer the one matching the project root path semantics. Ask the user if multiple candidates.

   **Python** (`pyproject.toml`):
   - `[project.scripts]` entries
   - Common: `uvicorn app:app --reload`, `flask run`, `python manage.py runserver`

   **JVM** (`build.gradle*`):
   - `./gradlew bootRun`, `./gradlew run`
   - For Maven: `mvn spring-boot:run`

   **Go** (`go.mod`):
   - `go run ./cmd/<name>` if `cmd/*` exists
   - `make dev` if Makefile has dev target

   **Procfile**:
   - First line, or specific named process

3. **If multiple candidates**, ask the user once with a numbered list. If just one obvious candidate, proceed directly.

4. **Suggest a name** — basename of the project root, or for monorepos the workspace name. Ask the user once if unsure.

5. **Call `devs register`**:
   ```bash
   devs register --name <NAME> --cwd "<ABSOLUTE_PATH>" --cmd "<DEV_COMMAND>"
   ```
   If duplicate name error, ask if user wants `--force` to replace.

6. **Confirm with `devs ls`**:
   ```bash
   devs ls
   ```
   Verify the new entry appears.

7. **Tell the user**: "등록 완료. `devs` 실행해서 `s` 또는 `enter`로 시작."

## Examples

### Example 1: pnpm monorepo

User says: "서버 등록해줘" while in `~/Desktop/order-platform-client`.

You:
1. Find `package.json` at root with `"order-history": "pnpm --filter order-history"` etc.
2. Notice `apps/order-history/` is a workspace with its own `dev` script.
3. Ask: "어떤 걸 등록할까요? (1) order-history, (2) ordersheet, (3) membership"
4. User picks 1.
5. Run: `devs register --name order-history --cwd ~/Desktop/order-platform-client --cmd "pnpm order-history dev"`

### Example 2: Single Python service

User in `~/projects/api`. `pyproject.toml` has no scripts, but `main.py` runs `uvicorn`.

You:
1. Read `main.py`, see `uvicorn.run(...)`.
2. Suggest cmd: `uvicorn main:app --reload --port 8000`
3. Run: `devs register --name api --cwd ~/projects/api --cmd "uvicorn main:app --reload --port 8000"`

## Don't

- Don't invent dev commands not present in the project. If you can't find one, ask the user.
- Don't register with absolute paths to executables that may not exist in the user's `PATH`. Prefer the form the user would type (`pnpm dev`, not `/Users/foo/.nvm/.../bin/pnpm dev`).
- Don't bulk-register without asking. One project at a time.
```

- [ ] **Step 2: Commit**

```bash
mkdir -p skills/devs-register
git add skills/devs-register/SKILL.md
git commit -m "skill: add devs-register for agent-driven project registration"
```

---

## Task 25: Skill — `devs-help`

**Files:**
- Create: `skills/devs-help/SKILL.md`

- [ ] **Step 1: Write skill**

`skills/devs-help/SKILL.md`:
```markdown
---
name: devs-help
description: Reference for using the `devs` local-server dashboard. Trigger when the user asks how to use devs, how to start/stop/view a registered dev server via devs, or anything about the devs dashboard.
---

# devs-help

Quick reference for the `devs` TUI dashboard.

## Launching

```bash
devs
```

Launches the TUI. If no projects are registered yet, the table is empty — use the `/devs-register` skill or `devs register --help` to add one.

## Inside the TUI

| Key | Action |
|-----|--------|
| ↑/↓ or j/k | Move cursor between rows |
| `enter` | Toggle the highlighted server (start if OFF, stop if ON) |
| `s` | Start |
| `x` | Stop (SIGTERM → SIGKILL after 3s) |
| `r` | Restart |
| `l` | Toggle log pane (last 200 lines of stdout) |
| `a` | Open inline Add form (name / cwd / cmd) |
| `e` | Open registry YAML in `$EDITOR` (reloads on save) |
| `R` (shift+r) | Manual refresh |
| `q` | Quit dashboard (servers keep running) |
| `Q` (shift+q) | Quit AND kill all dashboard-spawned servers |

## Files to know

| Path | Purpose |
|------|---------|
| `~/.config/devs/registry.yaml` | Registered projects. Edit manually or via TUI/skill. |
| `~/.local/state/devs/state.json` | Tracks PIDs the dashboard started. Don't edit. |
| `~/.local/state/devs/logs/<name>.log` | Captured stdout for dashboard-spawned servers. |

## Common scenarios

### "Why is my row stuck on OFF when the server is actually running?"

The matcher matches by **cwd** (the registered cwd must be a prefix of the live process's cwd) and a **distinctive token in the command** (e.g. `order-history` in `pnpm order-history dev`).

- Verify cwd: `lsof -a -p <pid> -d cwd`
- Verify command: `ps -p <pid> -o command=`

Fix: edit registry so `cwd` and `cmd` match the live invocation.

### "Server won't stop"

The dashboard sends SIGTERM to the process **group** (because we spawned with `setsid`). If a server traps SIGTERM or forks daemons we don't know about:

```bash
# Find rogue children
pgrep -P <pid>
# Hard kill the group
kill -9 -<pgid>
```

### "Dashboard says lock held"

```bash
cat ~/.local/state/devs/lock         # see which PID is supposedly running
ps -p <pid>                          # check if alive
rm ~/.local/state/devs/lock          # remove if stale
```

## Don't

- Don't edit `state.json` while `devs` is running.
- Don't kill the dashboard process with SIGKILL — `q` cleanly releases the lock; SIGKILL leaves it.
- Don't register a project with cwd inside a node_modules folder — match will be fragile.
```

- [ ] **Step 2: Commit**

```bash
mkdir -p skills/devs-help
git add skills/devs-help/SKILL.md
git commit -m "skill: add devs-help reference"
```

---

## Task 26: End-to-End Smoke Verification

**Files:**
- None (verification only)

- [ ] **Step 1: Full install + register + run cycle**

```bash
cd ~/Desktop/devs
make uninstall        # clean slate (binary + skills)
make install
make install-skills
which devs            # → ~/.local/bin/devs
ls -la ~/.claude/skills/ | grep devs

# Register two simple test servers
devs register --name smoke-a --cwd "$HOME" --cmd "/bin/sh -c 'while true; do echo a; sleep 1; done'" --force
devs register --name smoke-b --cwd "$HOME" --cmd "/bin/sh -c 'while true; do echo b; sleep 1; done'" --force

devs ls               # both shown
```

- [ ] **Step 2: TUI exercises**

Launch `devs`. Verify:
1. Both projects shown as `○ OFF`
2. Press `s` on smoke-a → row flips to `● ON` within 2s (but port column may show — since sleep doesn't open a port; for real port test register against a real dev server)
3. Press `l` → log pane shows `a\na\n...` lines
4. Press `r` → restart works
5. Press `Q` → both managed processes killed
6. Re-launch `devs` → both rows back to `○ OFF`

- [ ] **Step 3: Cleanup test data**

```bash
devs ls                          # confirm smoke-a, smoke-b
# Remove via editor:
$EDITOR ~/.config/devs/registry.yaml   # delete smoke-* entries, save
devs ls                          # confirm gone
```

- [ ] **Step 4: Tag v0.1.0**

```bash
git tag -a v0.1.0 -m "v0.1.0: initial release"
```

(Push when ready: `git remote add origin …` + `git push --tags`.)

- [ ] **Step 5: Commit final state**

```bash
git status     # should be clean
```

---

## Self-Review

Mental checklist after writing this plan:

1. **Spec coverage**: every section of the spec maps to a task:
   - Registry types/CRUD → T3, T4
   - `devs register` CLI → T6
   - State store + lock → T8
   - lsof/ps/matcher → T9, T10, T11
   - Process runner (detached + stop) → T12, T13
   - Git enricher → T14
   - TUI table/start-stop/log/form/edit → T16-T21
   - Single-instance lock integration → T18 (in TUI init)
   - Distribution (Makefile + INSTALL) → T22, T23
   - Skills → T24, T25
   - E2E → T26
   ✓

2. **No placeholders**: every code block is concrete Go/YAML/Makefile. No "TODO" or "TBD" left.

3. **Type consistency**: `registry.Project`, `state.Managed`, `discovery.Match`, `process.Start/Stop` consistent across tasks where referenced.

4. **One concern fixed in self-review**: Task 17 (table view) initially required state store integration, but state is added in Task 18 (start/stop). T17 is fine without state — it only reads registry + discovery. ✓
