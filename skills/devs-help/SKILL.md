---
name: devs-help
description: Reference for using the `devs` local-server dashboard. Trigger when the user asks how to use devs, how to start/stop/view a registered dev server via devs, or anything about the devs dashboard.
---

# devs-help

Quick reference for the `devs` TUI dashboard.

## Launching

```bash
devs
```

Launches the TUI. If no projects are registered yet, the table is empty — use the `/devs-register` skill or `devs register --help` to add one.

## Inside the TUI

| Key | Action |
|-----|--------|
| ↑/↓ or j/k | Move cursor between rows |
| `enter` | Toggle the highlighted server (start if OFF, stop if ON) |
| `s` | Start |
| `x` | Stop (SIGTERM → SIGKILL after 3s) |
| `r` | Restart |
| `l` | Toggle log pane (last 200 lines of stdout) |
| `a` | Open inline Add form (name / cwd / cmd) |
| `e` | Open registry YAML in `$EDITOR` (reloads on save) |
| `R` (shift+r) | Manual refresh |
| `q` | Quit dashboard (servers keep running) |
| `Q` (shift+q) | Quit AND kill all dashboard-spawned servers |

## Files to know

| Path | Purpose |
|------|---------|
| `~/.config/devs/registry.yaml` | Registered projects. Edit manually or via TUI/skill. |
| `~/.local/state/devs/state.json` | Tracks PIDs the dashboard started. Don't edit. |
| `~/.local/state/devs/logs/<name>.log` | Captured stdout for dashboard-spawned servers. |

## Common scenarios

### "Why is my row stuck on OFF when the server is actually running?"

The matcher matches by **cwd** (the registered cwd must be a prefix of the live process's cwd) and a **distinctive token in the command** (e.g. `order-history` in `pnpm order-history dev`).

- Verify cwd: `lsof -a -p <pid> -d cwd`
- Verify command: `ps -p <pid> -o command=`

Fix: edit registry so `cwd` and `cmd` match the live invocation.

### "Server won't stop"

The dashboard sends SIGTERM to the process **group** (because we spawned with `setsid`). If a server traps SIGTERM or forks daemons we don't know about:

```bash
# Find rogue children
pgrep -P <pid>
# Hard kill the group
kill -9 -<pgid>
```

### "Dashboard says lock held"

```bash
cat ~/.local/state/devs/lock         # see which PID is supposedly running
ps -p <pid>                          # check if alive
rm ~/.local/state/devs/lock          # remove if stale
```

## Don't

- Don't edit `state.json` while `devs` is running.
- Don't kill the dashboard process with SIGKILL — `q` cleanly releases the lock; SIGKILL leaves it.
- Don't register a project with cwd inside a node_modules folder — match will be fragile.
