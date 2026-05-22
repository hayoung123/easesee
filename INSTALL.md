# Installing easesee

This guide is written so an AI agent (Claude Code, etc.) can follow it step-by-step. Each step has a verification command — if it fails, fix that before continuing.

## 1. Prerequisites

### Go 1.22+

```bash
go version
```
Expected: `go version go1.22.x` or higher.

Install if missing:
```bash
brew install go
```

### git, lsof, ps

These ship with macOS by default:
```bash
which git lsof ps
```
Expected: three paths printed.

## 2. Clone

```bash
git clone https://github.com/hayoung123/easesee.git ~/.local/share/easesee
cd ~/.local/share/easesee
```

## 3. Build + Install

```bash
make install
```
Expected: `installed → /Users/<you>/.local/bin/easesee`

Verify:
```bash
which devs
devs --help
```
Expected: cobra help text.

If `which devs` fails, ensure `~/.local/bin` is in `PATH`:
```bash
echo $PATH | tr ':' '\n' | grep local/bin
```

## 4. Install Skills (optional but recommended)

```bash
make install-skills
```
Expected: two symlinks created in `~/.claude/skills/`.

Verify:
```bash
ls -la ~/.claude/skills/ | grep devs
```

## 5. First Registration

In a project root that has a dev server:
```bash
cd ~/path/to/your-project
devs register --name your-project --cwd "$PWD" --cmd "pnpm dev"
```

Or, with the skill installed, just tell your agent:
> 서버 등록해줘

The agent will read `package.json` (or pyproject/build.gradle/etc), pick the right dev script, and call `easesee register` for you.

## 6. Launch

```bash
devs
```
Expected: TUI shows your registered project as `○ OFF`. Press `s` to start it. After 2 seconds the row flips to `● ON`.

## Uninstall

```bash
cd ~/.local/share/easesee
make uninstall
```
Registry and state files are preserved; remove them manually if desired:
```bash
rm -rf ~/.config/easesee ~/.local/state/easesee
```
