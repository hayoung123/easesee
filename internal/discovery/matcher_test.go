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

// Once pnpm/yarn hand off to the real dev server, the live listener is a child
// (e.g. vite) whose cmdline is vite's own — the project name appears only
// inside the path, NOT as a whitespace-separated token. The name matcher must
// still attribute it, otherwise servers started outside the dashboard (or whose
// recorded PID has gone stale) show as OFF even though they're listening.
func TestMatch_ChildListenerWithProjectNameInPath(t *testing.T) {
	root := "/home/u/order-platform-client"
	projects := []registry.Project{
		{Name: "order-history", Cwd: root, Cmd: "pnpm order-history dev"},
		{Name: "ordersheet", Cwd: root, Cmd: "pnpm order dev"},
	}
	listeners := []Listener{{PID: 42, Cmd: "node", Port: 5173}}
	proc := func(pid int) ProcInfo {
		return ProcInfo{
			PID:     42,
			Cwd:     root + "/apps/order-history",
			Cmdline: "node /home/u/order-platform-client/apps/order-history/node_modules/.bin/../vite/bin/vite.js --mode beta",
		}
	}
	matches := Match(projects, listeners, proc)
	if m, ok := matches["order-history"]; !ok || m.Port != 5173 {
		t.Errorf("order-history should match its child vite listener: %+v", matches)
	}
	// Segment-level matching must not regress to substring: "order" (ordersheet)
	// must not match the "order-history" path segment.
	if _, ok := matches["ordersheet"]; ok {
		t.Errorf("ordersheet must NOT match (segment 'order' != 'order-history'): %+v", matches)
	}
}

// Some projects register a scoped pnpm filter (`pnpm --filter @app/direct-farm
// start`); the distinctive token is `@app/direct-farm`, but the live vite path
// carries only the bare segment `direct-farm`. The matcher must reconcile the
// two via the registered token's last path segment.
func TestMatch_ScopedFilterTokenMatchesPathSegment(t *testing.T) {
	root := "/home/u/service-front"
	projects := []registry.Project{
		{Name: "direct-farm", Cwd: root, Cmd: "pnpm --filter @app/direct-farm start"},
		{Name: "hypermarket", Cwd: root, Cmd: "pnpm --filter @app/hypermarket dev"},
	}
	listeners := []Listener{{PID: 7, Cmd: "node", Port: 3000}}
	proc := func(pid int) ProcInfo {
		return ProcInfo{
			PID:     7,
			Cwd:     root + "/apps/direct-farm",
			Cmdline: "node /home/u/service-front/apps/direct-farm/node_modules/.bin/../vite/bin/vite.js",
		}
	}
	matches := Match(projects, listeners, proc)
	if m, ok := matches["direct-farm"]; !ok || m.Port != 3000 {
		t.Errorf("direct-farm should match via @app/direct-farm -> direct-farm: %+v", matches)
	}
	if _, ok := matches["hypermarket"]; ok {
		t.Errorf("hypermarket must NOT match: %+v", matches)
	}
}
