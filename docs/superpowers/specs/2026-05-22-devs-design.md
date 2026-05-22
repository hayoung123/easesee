# devs — Local Dev Server Dashboard

> Date: 2026-05-22
> Owner: proshy
> Status: Design (v1)

## Goal

한 터미널 안에서 **등록된** 로컬 dev server를 보고 컨트롤할 수 있는 TUI. 등록된 서버는 다른 패널에서 띄웠어도 자동으로 인식되고, 대시보드에서 시작한 서버는 대시보드 종료에도 살아남는다. 등록은 에이전트가 대신 처리해서 사용자 부담을 0에 가깝게 한다.

## Non-goals (v1)

- 등록되지 않은 LISTEN 포트 추적 (별도 `dev-status` CLI가 담당)
- 로그 grep/filter — 단순 tail까지만
- 다중 인스턴스 동시 실행
- 원격 호스트 dev server 관리
- 자동 헬스체크/재시작
- cmux 깊은 통합(패널 점프 등) — v2 후보

## User stories

1. **첫 등록 (에이전트)**: 새 프로젝트 폴더에서 "서버 등록해줘" → 에이전트가 `package.json` 등을 읽고 dev script 식별 → `devs register` 호출 → 대시보드에 OFF 상태로 등장.
2. **시작/정지**: 대시보드 표에서 row 선택 → `enter`로 ON/OFF 토글. 다른 패널에서 같은 cwd로 띄워도 ON으로 자동 인식.
3. **재시작/로그**: `r`로 재시작, `l`로 로그 페인 토글.
4. **생존**: 대시보드 `q` → 띄워둔 서버는 그대로 작동. 다시 켜면 자동 재인식.
5. **수동 등록 fallback**: TUI에서 `a` 누르면 인라인 폼으로 직접 추가.

## Architecture

```
┌────────────────────────────────────────────┐
│  TUI (Bubbletea)                           │
│  표 / 로그 페인 / 키 입력 / 등록 폼        │
└────────┬───────────────────────────────────┘
         │
   ┌─────┴───────┐
   │ Aggregator  │  ← 2초 폴링
   └─────┬───────┘
         │
   ┌─────┼──────────────────┬──────────────┐
   │     │                  │              │
[Registry]   [State store]      [Process matcher]
 registry.yaml   state.json       lsof + ps
 (사용자 등록)   (대시보드 PID)    (cwd 기반 매칭)
                                  
[devs register CLI] ──> [Registry] (사람·에이전트 공통 진입점)
```

### 컴포넌트

| 이름 | 역할 |
|------|------|
| **Registry** | `~/.config/devs/registry.yaml`. 등록 프로젝트 (name, cwd, cmd, env). 대시보드 표의 원천. |
| **State store** | `~/.local/state/devs/state.json`. 대시보드가 spawn한 PID·log path 추적. 재시작 후 복구용. |
| **Process matcher** | 등록된 각 항목에 대해 lsof(LISTEN) + ps(cwd)로 살아있는지 검사. 매칭되면 ON으로 표시. |
| **Process runner** | `devs register` 항목을 `setsid` detached로 spawn. stdout/stderr → 로그 파일. |
| **Git enricher** | row의 cwd가 git 워킹트리면 `git -C cwd rev-parse --abbrev-ref HEAD`로 브랜치 표시. dirty이면 `★`. |
| **TUI** | Bubbletea + Bubbles(table, textinput, viewport) + Lipgloss. |
| **`devs register` CLI** | 사람·에이전트 공통 등록 진입점. |

### Process matching

등록 항목 `{name, cwd, cmd}`에 대해:

1. lsof로 LISTEN 포트 + PID 목록 수집
2. 각 PID의 cwd(`lsof -a -p PID -d cwd`)와 명령(`ps -p PID -o command=`) 추출
3. 등록 cwd로 시작하는 PID이고 명령이 등록 cmd의 핵심 토큰을 포함하면 매칭
4. 매칭되면 row의 STATE를 `ON`, port/PID/uptime 표시

장점: 외부 패널에서 띄워도 자동 인식. 우리 도구로만 띄울 필요 없음.

### Survival semantics

- 대시보드가 spawn하는 자식은 `syscall.SysProcAttr{Setsid: true}` 새 세션 시작
- stdout/stderr → `~/.local/state/devs/logs/<name>.log`
- 대시보드 종료해도 자식 살아남음
- 다음 실행 시 state.json + lsof 매칭으로 재구성
- 죽은 PID는 폴링 시 cleanup

### Single instance guard

- 시작 시 `~/.local/state/devs/lock` 의 PID 확인. 살아있는 다른 인스턴스 있으면 종료 안내.

## TUI

### Main view

```
┌─ devs ──────────────────────────────────────── 2026-05-22 18:53 ─┐
│ NAME            STATE   PORT   BRANCH         CMD                 │
│ ────────────── ─────── ────── ────────────── ──────────────────── │
│▶order-history  ● ON    5173   feat/abandon★  pnpm order-history…  │
│ food           ● ON    5174   main           pnpm dev             │
│ membership     ○ OFF   —      —              pnpm membership dev  │
│ appsim         ● ON    4991   —              appsim-cli --profile…│
│                                                                    │
│ enter:toggle  s:start  x:stop  r:restart  l:log  a:add  e:edit  q │
└────────────────────────────────────────────────────────────────────┘
```

`★` = git dirty. CMD 컬럼은 자르고 hover로 전체 보여주는 식.

### Add form (`a` 키)

```
┌─ Add project ───────────────────────────┐
│ name: order-history__                   │
│ cwd:  ~/Desktop/order-platform-client__ │
│ cmd:  pnpm order-history dev__          │
│                                          │
│ tab:next  enter:save  esc:cancel        │
└─────────────────────────────────────────┘
```

저장 시 내부적으로 `devs register` 호출(통일된 진입점).

### Log pane

- `l` 로 토글. 화면 하단 split, 선택 row의 로그 파일 tail.
- 외부 인식된 row(우리가 spawn하지 않은)는 "log not captured" 표시. `enter`로 "재시작해서 dashboard 관리로 가져오기" 옵션.

## 등록 방식 (Registration paths)

모두 결국 `devs register --name X --cwd Y --cmd "Z"` 호출로 수렴.

### 1. 에이전트 스킬 ⭐ (메인)

`/devs-register` 스킬 (또는 자연어 "서버 등록해줘").

스킬이 에이전트한테 가르치는 절차:

1. 현재 cwd에서 프로젝트 root 탐색 (`package.json`, `pyproject.toml`, `build.gradle*`, `go.mod`, `Cargo.toml`, `Makefile`, `Procfile`)
2. 빌드/실행 도구에 맞춰 dev 후보 추출:
   - **Node**: `package.json` scripts → `dev`, `start`, `serve`, `start:dev`
   - **monorepo (pnpm-workspace 등)**: 각 워크스페이스의 scripts 확인
   - **Python**: `pyproject.toml [project.scripts]`, `uvicorn`, `flask run`, `manage.py runserver`
   - **JVM**: `gradle bootRun`, `gradle run`, `mvn spring-boot:run`
   - **Go**: `make dev`, `go run ./cmd/<name>`
   - **Procfile**: 라인 그대로 옵션
3. 후보가 둘 이상이면 사용자한테 한 줄 질문
4. 이름 제안 (cwd 마지막 폴더 또는 monorepo면 워크스페이스명)
5. `devs register --name <X> --cwd <Y> --cmd "<Z>"` 실행
6. "등록 완료. `devs` 실행해서 `s` 키로 시작" 안내

스킬 파일: `skills/devs-register/SKILL.md` (이 레포에 포함, 설치 시 `~/.claude/skills/`에 심볼릭 링크).

### 2. CLI 직접

```bash
devs register --name order-history \
  --cwd ~/Desktop/order-platform-client \
  --cmd "pnpm order-history dev"
```

검증:
- 같은 이름 있으면 에러 + `--force`로 덮어쓰기
- cwd가 존재해야 함
- cmd가 비어있으면 에러

### 3. TUI 인라인 폼

`a` 키 → 폼 → 저장 시 `devs register` 호출.

### 4. (옵션) 첫 실행 import 마법사

registry가 비었고 사용자가 동의하면 현재 LISTEN 중인 dev server들을 한 번에 import. v1.5로 미룸 — 스킬이 더 빠를 가능성 큼.

## 파일 레이아웃

```
~/.config/devs/
  registry.yaml             ← 사용자 등록
~/.local/state/devs/
  state.json                ← runtime 추적
  lock                      ← single-instance lock
  logs/
    order-history.log
    food.log
~/.local/bin/devs           ← binary
~/.claude/skills/devs-register/   ← 심볼릭 링크 → 레포 skills/devs-register/
~/.claude/skills/devs-help/       ← 심볼릭 링크 → 레포 skills/devs-help/
```

### Registry 포맷

```yaml
version: 1
projects:
  - name: order-history
    cwd: ~/Desktop/order-platform-client
    cmd: pnpm order-history dev
    env:
      NODE_ENV: development
  - name: food
    cwd: ~/Desktop/food-order-client/apps/food
    cmd: pnpm dev
  - name: appsim
    cwd: ~
    cmd: appsim-cli --profile ~/.appsim/proshy.json --address ~/.appsim/jamsil.json
```

- `version`: 향후 마이그레이션 대비
- `env`: 옵션
- 경로의 `~` 는 expand

## 키 바인딩

| 키 | 동작 |
|----|------|
| `↑` `↓` `j` `k` | row navigate |
| `enter` | ON/OFF 토글 (= start/stop) |
| `s` | start |
| `x` | stop (SIGTERM → 5s → SIGKILL) |
| `r` | restart |
| `l` | log pane toggle |
| `a` | add (인라인 폼) |
| `e` | edit registry (`$EDITOR` 호출, 종료 시 reload) |
| `R` | manual refresh |
| `q` | quit (자식 그대로 둠) |
| `Q` | quit + kill all dashboard-spawned |

## Error handling

| 케이스 | 대응 |
|--------|------|
| lsof/ps 실패 | 상태바 "discovery error" 표시, 다음 폴링 재시도 |
| spawn 실패 | 로그·상태바에 메시지 |
| state.json 파싱 실패 | `.bak` 백업 후 새로 시작 |
| 죽은 PID가 state에 남음 | 폴링 시 자동 정리 |
| registry.yaml syntax error | 상태바 표시, 마지막 valid 버전 유지 |
| worker spawn (vite 등) | process group 단위 SIGTERM (`kill -- -<pgid>`) |
| `devs register` 충돌 | 같은 이름이면 에러, `--force` 옵션 제공 |

## 테스팅

- **단위**: registry 파싱·검증, state 직렬화, lsof/ps 출력 파서, cwd 매칭 휴리스틱
- **통합**: 등록된 `sleep 1000` 같은 더미 명령 spawn → kill → state 정리 시나리오
- **TUI**: Bubbletea의 `teatest`로 핵심 시나리오 1-2개 (start/stop, add 폼)
- **스킬**: SKILL.md 자체 시나리오 검증 (몇 가지 sample repo에서 의도한 명령이 추출되는지)

## Stack

- Go 1.22+
- [Bubbletea](https://github.com/charmbracelet/bubbletea)
- [Bubbles](https://github.com/charmbracelet/bubbles)
- [Lipgloss](https://github.com/charmbracelet/lipgloss)
- [gopkg.in/yaml.v3](https://pkg.go.dev/gopkg.in/yaml.v3)
- 외부 의존: `lsof`, `git`, `ps` (macOS/Linux 표준)

## Distribution & Installation

### 레포 레이아웃

```
devs/
  cmd/devs/main.go
  internal/
    registry/       # YAML 로드·검증
    state/          # runtime state
    process/        # spawn, signal, matching
    discovery/      # lsof/ps wrapper
    tui/            # bubbletea models
    git/            # branch/dirty
  skills/
    devs-register/SKILL.md
    devs-help/SKILL.md
  docs/
    superpowers/specs/...
  Makefile
  README.md
  INSTALL.md            # 에이전트가 따라가는 설치 가이드 (appsim-cli SETUP.md 패턴)
  go.mod
```

### 설치 경로 (에이전트 친화)

대상 사용자는 "에이전트한테 설치 시키는 사람". appsim-cli가 `SETUP.md`로 에이전트한테 1단계씩 가르친 패턴을 그대로.

**`INSTALL.md` 구조 (에이전트 읽기용)**:

1. **Go 1.22+ 확인** — 없으면 `brew install go`
2. **레포 클론** — `git clone https://github.com/proshy/devs.git ~/.local/share/devs && cd ~/.local/share/devs`
3. **빌드 + 설치** — `make install` (→ `go build` → `~/.local/bin/devs`)
4. **스킬 링크** — `make install-skills` (→ `~/.claude/skills/devs-*` 심볼릭 링크)
5. **확인** — `devs --version`

각 단계에 검증 명령 포함. 실패 시 무엇이 잘못됐는지 명확한 메시지.

### Makefile 타깃

```makefile
build:        # go build → bin/devs
install:      # build + cp to ~/.local/bin/devs
install-skills: # symlink skills/* to ~/.claude/skills/
uninstall:    # remove binary + symlinks (state·registry는 보존)
test:
clean:
```

### 릴리스 (v2 후보, v1엔 git clone만)

- GitHub Releases + 미리 빌드된 바이너리 (macOS arm64/amd64, Linux amd64)
- `curl -sL …/install.sh | sh` 원라이너
- 가능하면 Homebrew tap

## Skills

레포에 두 개 포함:

### `skills/devs-register/SKILL.md`

(앞서 설명한 등록 절차) — 에이전트에게 어떻게 프로젝트 정보 추출하고 `devs register` 호출할지 가르침. 트리거: "서버 등록", "devs에 등록", "프로젝트 등록".

### `skills/devs-help/SKILL.md`

devs 자체 사용법을 에이전트한테 알려주는 reference 스킬. 트리거: "devs 어떻게 써", "devs 실행", "로컬 서버 켜". 내용:

- 대시보드 띄우는 법 (`devs` 한 줄)
- 키바인딩 요약
- registry 위치 / 직접 편집 방법
- 자주 묻는 시나리오 (특정 프로젝트만 시작, 로그 확인, kill 안 될 때 등)

## Risks / unknowns

- macOS lsof는 SIP-protected 프로세스 일부 가시성 제한 (사용자 프로세스 위주라 영향 적음)
- vite/next dev처럼 worker 띄우는 도구 — process group 단위 종료 필요
- 같은 cwd에서 두 명령 동시 실행 시 매칭 모호 — 등록 cmd의 핵심 토큰 일치까지 검사
- monorepo에서 워크스페이스별 등록 시 cwd가 sub-path여야 매칭 정확

## v1 deliverables checklist

- [ ] `devs` binary (TUI + `register` 서브커맨드)
- [ ] `skills/devs-register/SKILL.md`
- [ ] `skills/devs-help/SKILL.md`
- [ ] `README.md` (소개 + 빠른 시작)
- [ ] `INSTALL.md` (에이전트가 따라가는 설치 절차)
- [ ] `Makefile` (`build`, `install`, `install-skills`, `test`, `clean`)
- [ ] 단위·통합·스킬 시나리오 테스트
