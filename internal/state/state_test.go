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
