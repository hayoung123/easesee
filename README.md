# easesee

[![npm version](https://img.shields.io/npm/v/easesee.svg)](https://www.npmjs.com/package/easesee)
[![release](https://img.shields.io/github/v/release/hayoung123/easesee)](https://github.com/hayoung123/easesee/releases)
[![license](https://img.shields.io/github/license/hayoung123/easesee)](./LICENSE)

A terminal dashboard for managing locally registered dev servers — see what's running where, start/stop with one keystroke, and let agents do the registration for you.

## Why

If you juggle multiple dev servers across cmux/tmux panes you lose track of which is running where. `easesee` shows them all in one place, recognizes servers started in other panes, and survives its own restart so your work isn't tied to its uptime.

## Features

- **One panel, all your dev servers** — see ports, branches, run state
- **Survives restart** — dashboard exit doesn't kill your servers
- **Detects external** — a server started in another pane shows up automatically (matched by cwd)
- **Agent-friendly** — the `easesee-register` skill lets an agent register projects from project context with one sentence
- **Single Go binary** — no daemon, no runtime dependency at install time

## Install

Pick one. All paths land at the same place: a single `easesee` binary on your `PATH`.

### npm (recommended)

```bash
npm install -g easesee
# or
pnpm add -g easesee
```

A small wrapper downloads the right native binary for your platform from the matching [GitHub release](https://github.com/hayoung123/easesee/releases) on first run.

### Go

```bash
go install github.com/hayoung123/easesee/cmd/easesee@latest
```

Requires Go 1.22+. Binary lands at `$(go env GOBIN)/easesee` (usually `~/go/bin/easesee`).

### From source

```bash
git clone https://github.com/hayoung123/easesee.git ~/.local/share/easesee
cd ~/.local/share/easesee
make install         # → ~/.local/bin/easesee
make install-skills  # → ~/.claude/skills/easesee-{register,help}
```

This is also the path used by the [agent-friendly INSTALL guide](./INSTALL.md), which an AI agent can follow step by step.

### Verify

```bash
easesee --version
easesee --help
```

## Quick start

```bash
# Register a project — manually
easesee register --name order-history \
  --cwd ~/Desktop/order-platform-client \
  --cmd "pnpm order-history dev"

# Or, with the skill installed, just ask your agent:
#   "서버 등록해줘"

# Launch the dashboard
easesee
```

## Keys

| Key | Action |
|-----|--------|
| ↑/↓ or j/k | Move cursor |
| enter | Toggle running/stopped |
| s | Start |
| x | Stop |
| r | Restart |
| l | Toggle log pane |
| a | Add project (inline form) |
| e | Edit `registry.yaml` in `$EDITOR` |
| R | Manual refresh |
| q | Quit (servers keep running) |
| Q | Quit + kill all dashboard-spawned |

## Files

| Path | Purpose |
|------|---------|
| `~/.config/easesee/registry.yaml` | Registered projects |
| `~/.local/state/easesee/state.json` | Dashboard-managed PIDs |
| `~/.local/state/easesee/logs/<name>.log` | Captured stdout/stderr |
| `~/.local/state/easesee/lock` | Single-instance lock |

## Skills (Claude Code)

- `easesee-register` — "서버 등록해줘" / "register this dev server" → scans the project (`package.json`, `pyproject.toml`, `build.gradle`, `go.mod`, `Procfile`…) and calls `easesee register` with the right cwd + cmd.
- `easesee-help` — reference for the TUI itself.

Install via `make install-skills` after cloning the repo. They live as symlinks under `~/.claude/skills/`.

## Supported platforms

- macOS (Apple Silicon + Intel)
- Linux (x86_64 + arm64)

Windows is not supported — uses Unix process groups (`setsid`) for survival semantics.

## Documentation

- [Design spec](./docs/superpowers/specs/2026-05-22-devs-design.md)
- [Implementation plan (26 tasks)](./docs/superpowers/plans/2026-05-22-devs-implementation.md)
- [CHANGELOG](./CHANGELOG.md)

## Releasing

See [RELEASING.md](./RELEASING.md) for the version bump → build → tag → npm publish flow.

## License

MIT — see [LICENSE](./LICENSE).
