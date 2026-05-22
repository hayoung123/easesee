# easesee

A terminal dashboard for managing locally registered dev servers — see what's running where, start/stop with one keystroke, and let agents do the registration for you.

## Why

If you juggle multiple dev servers across cmux/tmux panes you lose track of which is running where. `easesee` shows them all in one place, recognizes servers started in other panes, and survives its own restart so your work isn't tied to its uptime.

## Features

- **One panel, all your dev servers** — see ports, branches, run state
- **Survives restart** — dashboard exit doesn't kill your servers
- **Detects external** — a server started in another pane shows up automatically (matched by cwd)
- **Agent-friendly** — `/easesee-register` skill lets your agent register projects from project context with one sentence
- **Single Go binary** — no daemon

## Quick start

See [INSTALL.md](./INSTALL.md) for step-by-step installation (agent-friendly).

```bash
make install        # → ~/.local/bin/easesee
make install-skills # → ~/.claude/skills/easesee-{register,help}

# Register a project (manually or via the skill)
devs register --name my-app --cwd ~/path/to/repo --cmd "pnpm dev"

# Launch the dashboard
devs
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
| e | Edit `registry.yaml` in $EDITOR |
| R | Manual refresh |
| q | Quit (servers keep running) |
| Q | Quit + kill all dashboard-spawned |

## Files

- Registry: `~/.config/easesee/registry.yaml`
- Runtime state: `~/.local/state/easesee/state.json`
- Logs: `~/.local/state/easesee/logs/<name>.log`

## Skills

- `~/.claude/skills/easesee-register` — agent skill: "서버 등록해줘" → scans project, calls `easesee register`
- `~/.claude/skills/easesee-help` — usage reference for agents

## Design

See [docs/superpowers/specs/2026-05-22-devs-design.md](./docs/superpowers/specs/2026-05-22-devs-design.md).

## License

MIT (or your choice).
