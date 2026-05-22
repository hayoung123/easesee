# Changelog

All notable changes to this project will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/hayoung123/easesee/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/hayoung123/easesee/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/hayoung123/easesee/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/hayoung123/easesee/releases/tag/v0.1.0
