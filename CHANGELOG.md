# Changelog

All notable changes to this project will be documented here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

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

[Unreleased]: https://github.com/hayoung123/easesee/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/hayoung123/easesee/releases/tag/v0.1.0
