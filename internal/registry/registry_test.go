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
