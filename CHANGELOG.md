# Changelog

All notable changes to this project will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.7] — 2026-05-29

### Fixed

- Dev servers started outside the dashboard (or whose tracked PID went stale) now show as ON with their real port. The matcher attributes child listeners — e.g. vite spawned by pnpm, whose command line carries the project name only inside its path — by matching path segments, including scoped pnpm filters (`@app/foo`).

### Changed

- Restyled the dashboard with a neutral palette: muted header/help/status colors, an underlined table header, and tidied key bar and state labels.

## [0.1.6] — 2026-05-26

### Added

- Kill-port command (`K`): enter a port number to see which process is using it, then confirm to kill it.

### Fixed

- Stopping a server now kills the entire process group so child processes (e.g. vite spawned by pnpm) release their ports before the TUI updates.
- Port discovery now uses `syscall.Getpgid` instead of `ps` for faster and more reliable process group lookup.
- When a process listens on multiple ports (e.g. appsim), the lowest port is shown instead of whichever lsof returns first.
- Processes running longer than 8 seconds without a port (e.g. CLI tools) now show `—` instead of `…` in the PORT column.

## [0.1.5] — 2026-05-24

### Fixed

- `easesee-register` skill now uses `easesee` instead of `devs` in precondition check and register/ls command examples.

## [0.1.4] — 2026-05-23

### Fixed

- PORT column now resolves for dashboard-spawned servers running under `pnpm`/`yarn`. Those tools launch the actual listener (e.g. `vite`) as a child whose cmdline is just `node …/vite` — no project name — so the cwd+cmd matcher couldn't connect it back to the registered project. The TUI now attributes listeners to the project by **process group**: every descendant of a `setsid`-spawned process shares the parent's PGID, so any listener whose PGID equals the recorded `state.Managed.PID` is recognised as that project's port.

## [0.1.3] — 2026-05-23

### Fixed

- Header now reads `easesee` (was `devs`, a leftover from the rename).
- STATE/BRANCH columns no longer render as `�`. The lipgloss-wrapped `● ON` / `○ OFF` / `★` strings were being truncated by `bubbles/table` without ANSI-stripping first, mangling the UTF-8 multibyte sequences. Cell content is now plain unicode; the rest of the TUI keeps its styling.

### Added

- Rows for dashboard-spawned servers flip to `● ON` immediately and show `…` for PORT until the dev server actually binds. Previously the row stayed `OFF` for the few seconds between spawn and lsof catching the new listener. Dead PIDs left in the state store are now garbage-collected on each refresh.

## [0.1.2] — 2026-05-23

### Fixed

- Sibling projects in a monorepo no longer light up together. Previously the matcher used `strings.Contains` for command-token matching, so the token `order` from a registered `ordersheet` project would falsely match the live cmdline `pnpm order-history dev` and flip the `ordersheet` row to `ON` while sharing the same port. The matcher now requires **whole-token** matches and gates the fallback so projects whose distinctive token didn't match don't fall through to generic words like `pnpm`.

## [0.1.1] — 2026-05-22

### Fixed

- `Stop` (and the TUI `x`/`Q` shortcuts) now successfully kills externally-started servers. Previously the function only signalled the process group via `-pid`, which silently missed any server whose PGID differed from its PID (i.e. anything not spawned by `easesee` itself — the common case for servers running in another cmux/tmux pane). `signal()` now sends to both the resolved PGID and the PID directly.
- Removed an over-eager post-SIGKILL liveness check that could spuriously fail when the OS hadn't yet reaped the zombie.

## [0.1.0] — 2026-05-22

### Added

- TUI dashboard for registered dev servers (Bubbletea + Bubbles + Lipgloss)
- `easesee register` CLI for adding projects
- `easesee ls` CLI for listing the registry
- Registry stored as YAML at `~/.config/easesee/registry.yaml`
- Runtime state with detached process spawn (`setsid`) — dashboard exit doesn't kill servers
- 2-second discovery refresh (lsof + ps) matches registered projects by cwd + command tokens
- Inline add form (`a`), `$EDITOR` integration (`e`), log pane (`l`), start/stop/restart/quit-with-cleanup
- Git branch + dirty indicator per row
- Single-instance lock via stale-PID-aware file lock
- Agent skills: `easesee-register` (auto-detect dev command from project context), `easesee-help` (reference)
- Cross-platform binaries: darwin-arm64, darwin-amd64, linux-amd64, linux-arm64
- Distribution: GitHub Releases + npm (`easesee` package with download wrapper)
- GitHub Actions CI on macOS + Ubuntu

[Unreleased]: https://github.com/hayoung123/easesee/compare/v0.1.4...HEAD
[0.1.4]: https://github.com/hayoung123/easesee/compare/v0.1.3...v0.1.4
[0.1.3]: https://github.com/hayoung123/easesee/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/hayoung123/easesee/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/hayoung123/easesee/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/hayoung123/easesee/releases/tag/v0.1.0
