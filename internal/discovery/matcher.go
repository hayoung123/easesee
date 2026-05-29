package discovery

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/hayoung123/easesee/internal/registry"
)

type MatchResult struct {
	PID       int
	Port      int
	StartedAt time.Time // zero if not tracked; used to distinguish "starting" from "no port"
}

// Match returns name -> MatchResult for projects whose cwd contains a LISTEN PID with matching cmd tokens.
// procFn is injected so tests can stub ps/lsof lookups.
func Match(projects []registry.Project, listeners []Listener, procFn func(pid int) ProcInfo) map[string]MatchResult {
	out := map[string]MatchResult{}
	for _, l := range listeners {
		info := procFn(l.PID)
		for _, p := range projects {
			if existing, ok := out[p.Name]; ok {
				// Same PID with a smaller port: prefer lower port (user-facing).
				if existing.PID == l.PID && l.Port < existing.Port {
					out[p.Name] = MatchResult{PID: l.PID, Port: l.Port}
				}
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

// cmdTokenOverlap returns true if any distinctive whole token of the registered
// cmd appears as a whole token in the live cmdline.
//
// "Whole token" matters: previously a substring match would let the registered
// token "order" match the live token "order-history" — making the dashboard
// light up unrelated rows that share a cwd in a monorepo.
//
// "Distinctive" tokens exclude very short or generic words (pnpm, dev, …).
// If the registered cmd has no distinctive tokens (e.g. just `pnpm dev`), the
// fallback compares all non-empty tokens — but still as whole-token matches.
func cmdTokenOverlap(registered, live string) bool {
	liveSet := tokenize(live)
	regTokens := strings.Fields(registered)

	var hasDistinctive bool
	for _, t := range regTokens {
		if isTrivialToken(t) {
			continue
		}
		hasDistinctive = true
		if liveSet[t] {
			return true
		}
		// Scoped filter tokens (`@app/direct-farm`) carry the distinctive name
		// as their last path segment, while the live vite path exposes only the
		// bare segment (`direct-farm`). Reconcile via the registered basename.
		if base := filepath.Base(t); base != t && liveSet[base] {
			return true
		}
	}
	if hasDistinctive {
		// Registered cmd had distinctive tokens but none matched: definitive no.
		// Don't fall through to trivial-token matches, or sibling projects in a
		// monorepo with the same cwd will all light up on any shared word.
		return false
	}
	// Registered cmd was only trivial tokens (e.g. `pnpm dev`).
	// Fall back to any whole-token match.
	for _, t := range regTokens {
		if liveSet[t] {
			return true
		}
	}
	return false
}

// tokenize splits a command line into whole-word tokens, also adding the
// basename of any path-shaped token (so `/usr/bin/pnpm` matches `pnpm`) and
// each path segment (so a vite child whose cmdline only carries the project
// name inside its path — `…/apps/order-history/…/vite.js` — still matches
// `order-history` as a whole token, without regressing to substring matching).
func tokenize(cmdline string) map[string]bool {
	out := map[string]bool{}
	for _, w := range strings.Fields(cmdline) {
		if w == "" {
			continue
		}
		out[w] = true
		if base := filepath.Base(w); base != w {
			out[base] = true
		}
		if strings.Contains(w, "/") {
			for _, seg := range strings.Split(w, "/") {
				if seg != "" {
					out[seg] = true
				}
			}
		}
	}
	return out
}

func isTrivialToken(t string) bool {
	if len(t) < 3 {
		return true
	}
	switch t {
	case "run", "dev", "start", "serve", "exec",
		"pnpm", "npm", "yarn", "node", "bun", "deno",
		"--filter":
		return true
	}
	return false
}
