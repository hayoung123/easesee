package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/hayoung123/easesee/internal/registry"
)

func TestRegister_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	regPath := filepath.Join(dir, "registry.yaml")
	if err := registerProject(regPath, registry.Project{
		Name: "myapp",
		Cwd:  "/tmp/myapp",
		Cmd:  "echo hi",
	}, false); err != nil {
		// /tmp/myapp may not exist; we'll create it
		if err2 := os.MkdirAll("/tmp/myapp", 0o755); err2 != nil {
			t.Skip("cannot create /tmp/myapp")
		}
		if err := registerProject(regPath, registry.Project{
			Name: "myapp",
			Cwd:  "/tmp/myapp",
			Cmd:  "echo hi",
		}, false); err != nil {
			t.Fatalf("registerProject: %v", err)
		}
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
	p := registry.Project{Name: "a", Cwd: dir, Cmd: "x"}
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
	if err := registerProject(regPath, registry.Project{Name: "a", Cwd: dir, Cmd: "x"}, false); err != nil {
		t.Fatal(err)
	}
	if err := registerProject(regPath, registry.Project{Name: "a", Cwd: dir, Cmd: "y"}, true); err != nil {
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
