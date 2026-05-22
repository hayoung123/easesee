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
		t.Fatalf("got %d listeners, want 2: %+v", len(got), got)
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
