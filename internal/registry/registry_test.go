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
