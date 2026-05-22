package discovery

import (
	"testing"

	"github.com/hayoung123/easesee/internal/registry"
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

// Regression: in a pnpm monorepo, registering both `order-history` (cmd:
// `pnpm order-history dev`) and `ordersheet` (cmd: `pnpm order dev`) under
// the same root cwd used to light up BOTH rows when only order-history was
// actually running, because the substring "order" appears inside
// "order-history" in the live cmdline.
func TestMatch_SiblingTokensDoNotFalsePositive(t *testing.T) {
	root := "/home/u/order-platform-client"
	projects := []registry.Project{
		{Name: "order-history", Cwd: root, Cmd: "pnpm order-history dev"},
		{Name: "ordersheet", Cwd: root, Cmd: "pnpm order dev"},
		{Name: "membership", Cwd: root, Cmd: "pnpm membership dev"},
	}
	listeners := []Listener{{PID: 42, Cmd: "node", Port: 5173}}
	proc := func(pid int) ProcInfo {
		return ProcInfo{
			PID:     42,
			Cwd:     root + "/apps/order-history",
			Cmdline: "node /usr/bin/pnpm order-history dev",
		}
	}
	matches := Match(projects, listeners, proc)
	if m, ok := matches["order-history"]; !ok || m.Port != 5173 {
		t.Errorf("order-history should match: %+v", matches)
	}
	if _, ok := matches["ordersheet"]; ok {
		t.Errorf("ordersheet should NOT match (token 'order' must not substring-match 'order-history'): %+v", matches)
	}
	if _, ok := matches["membership"]; ok {
		t.Errorf("membership should NOT match: %+v", matches)
	}
}
