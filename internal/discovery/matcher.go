package discovery

import (
	"strings"

	"github.com/hayoung123/easesee/internal/registry"
)

type MatchResult struct {
	PID  int
	Port int
}

// Match returns name -> MatchResult for projects whose cwd contains a LISTEN PID with matching cmd tokens.
// procFn is injected so tests can stub ps/lsof lookups.
func Match(projects []registry.Project, listeners []Listener, procFn func(pid int) ProcInfo) map[string]MatchResult {
	out := map[string]MatchResult{}
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
			out[p.Name] = MatchResult{PID: l.PID, Port: l.Port}
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
