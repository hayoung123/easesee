---
name: easesee-register
description: Register a project in the devs local-server dashboard. Trigger when the user mentions "easesee 등록", "서버 등록", "dev server 등록", or similar requests to add a project to the local dev dashboard.
---

# devs-register

Helps the user add a local project to the `easesee` dashboard by detecting how its dev server is run and calling `easesee register`.

## Preconditions

Before running, verify:

```bash
which easesee
```

If missing, direct the user to [INSTALL.md](https://github.com/hayoung123/easesee/blob/main/INSTALL.md).

## Flow

1. **Find project root** — starting from `$PWD` (or a path the user gave you), walk up until you find one of:
   - `package.json`
   - `pyproject.toml`
   - `build.gradle`, `build.gradle.kts`, `pom.xml`
   - `go.mod`
   - `Cargo.toml`
   - `Makefile`
   - `Procfile`

2. **Identify dev candidates** — based on what you found:

   **Node** (`package.json`):
   - Look at `scripts` field. Common dev names: `dev`, `start`, `serve`, `start:dev`, `dev:server`.
   - For monorepos (`pnpm-workspace.yaml`, `lerna.json`, `nx.json`, workspaces in package.json): check each package's `scripts` and prefer the one matching the project root path semantics. Ask the user if multiple candidates.

   **Python** (`pyproject.toml`):
   - `[project.scripts]` entries
   - Common: `uvicorn app:app --reload`, `flask run`, `python manage.py runserver`

   **JVM** (`build.gradle*`):
   - `./gradlew bootRun`, `./gradlew run`
   - For Maven: `mvn spring-boot:run`

   **Go** (`go.mod`):
   - `go run ./cmd/<name>` if `cmd/*` exists
   - `make dev` if Makefile has dev target

   **Procfile**:
   - First line, or specific named process

3. **If multiple candidates**, ask the user once with a numbered list. If just one obvious candidate, proceed directly.

4. **Suggest a name** — basename of the project root, or for monorepos the workspace name. Ask the user once if unsure.

5. **Call `easesee register`**:
   ```bash
   easesee register --name <NAME> --cwd "<ABSOLUTE_PATH>" --cmd "<DEV_COMMAND>"
   ```
   If duplicate name error, ask if user wants `--force` to replace.

6. **Confirm with `easesee ls`**:
   ```bash
   easesee ls
   ```
   Verify the new entry appears.

7. **Tell the user**: "등록 완료. `easesee` 실행해서 `s` 또는 `enter`로 시작."

## Examples

### Example 1: pnpm monorepo

User says: "서버 등록해줘" while in `~/Desktop/order-platform-client`.

You:
1. Find `package.json` at root with `"order-history": "pnpm --filter order-history"` etc.
2. Notice `apps/order-history/` is a workspace with its own `dev` script.
3. Ask: "어떤 걸 등록할까요? (1) order-history, (2) ordersheet, (3) membership"
4. User picks 1.
5. Run: `easesee register --name order-history --cwd ~/Desktop/order-platform-client --cmd "pnpm order-history dev"`

### Example 2: Single Python service

User in `~/projects/api`. `pyproject.toml` has no scripts, but `main.py` runs `uvicorn`.

You:
1. Read `main.py`, see `uvicorn.run(...)`.
2. Suggest cmd: `uvicorn main:app --reload --port 8000`
3. Run: `easesee register --name api --cwd ~/projects/api --cmd "uvicorn main:app --reload --port 8000"`

## Don't

- Don't invent dev commands not present in the project. If you can't find one, ask the user.
- Don't register with absolute paths to executables that may not exist in the user's `PATH`. Prefer the form the user would type (`pnpm dev`, not `/Users/foo/.nvm/.../bin/pnpm dev`).
- Don't bulk-register without asking. One project at a time.
