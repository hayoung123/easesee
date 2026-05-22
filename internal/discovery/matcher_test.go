package discovery

import (
	"testing"

	"github.com/proshy/easesee/internal/registry"
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
