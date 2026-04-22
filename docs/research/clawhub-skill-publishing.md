# ClawHub 심층 분석과 tene CLI용 SKILL.md 작성·등록 가이드

- **작성일**: 2026-04-22
- **대상**: Tene 유지관리자 / Skill 퍼블리셔
- **목적**: clawhub.ai가 무엇인지 정확히 이해하고, tene CLI를 AI 에이전트가 "바로 설치하고 안전하게 사용"할 수 있도록 돕는 Skill을 만들어 ClawHub에 등록하는 전 과정을 문서화한다.

---

## 1. ClawHub란 무엇인가 (Deep Dive)

### 1.1 한 줄 정의
ClawHub는 **OpenClaw (및 Claude Code)용 공개 스킬 레지스트리**다. "에이전트용 npm"이라고 생각하면 된다. 텍스트 기반 스킬 번들(`SKILL.md` + 보조 파일)을 퍼블리시·버저닝·검색할 수 있는 커뮤니티 마켓플레이스다.

- 공식 사이트: https://clawhub.ai
- 공개 레포: https://github.com/openclaw/clawhub
- 자매 레지스트리: https://onlycrabs.ai (SOUL.md 레지스트리)
- 플랫폼 통계(공식 사이트 기준): 52,700+ tools, 180,000+ users, 12M downloads, 평점 4.8

### 1.2 아키텍처 개요
- **웹앱**: TanStack Start (React, Vite/Nitro) — Vercel 호스팅
- **백엔드**: Convex (DB + 파일 스토리지 + HTTP Actions) + Convex Auth (GitHub OAuth)
- **검색**: OpenAI `text-embedding-3-small` + Convex 벡터 검색 → 키워드가 아닌 **의미 기반 검색**이 장점
- **스키마/API**: `packages/schema` 라이브러리가 CLI와 웹앱 간 계약을 정의

### 1.3 Skill의 물리적 구조
Skill은 **폴더 하나**다.

```
my-skill/
├── SKILL.md              # 필수 (또는 skill.md)
├── README.md             # 권장
├── examples/             # 권장
├── tests/                # 권장
├── .clawhubignore        # publish/sync 제외 패턴 (선택)
└── .gitignore            # 선택, 존재 시 존중됨
```

서버 측 제약:
- 총 번들 크기: **50MB 이하**
- 텍스트 기반 파일만 허용 (`packages/schema/src/textFiles.ts`의 `TEXT_FILE_EXTENSIONS` 허용목록)
- 임베딩 대상: `SKILL.md` + 최대 ~40개의 non-`.md` 파일
- Slug 규칙: `^[a-z0-9][a-z0-9-]*$` (소문자·숫자·하이픈)
- 라이선스: 퍼블리시 시 **MIT-0** 자동 적용 (저작자 표시 불필요, 상업적 이용 허용)

### 1.4 SKILL.md 프론트매터 전체 레퍼런스

**필수(사실상)**
```yaml
---
name: skill-name                    # 식별자
description: 한 줄 요약 (트리거 역할)
version: 1.0.0                       # 시맨틱 버저닝
---
```

**런타임 요구사항**: `metadata.openclaw`(alias: `metadata.clawdbot`, `metadata.clawdis`)

| 필드 | 타입 | 설명 |
|---|---|---|
| `requires.env` | `string[]` | 스킬이 기대하는 환경 변수 |
| `requires.bins` | `string[]` | **모두** 설치되어 있어야 하는 CLI 바이너리 |
| `requires.anyBins` | `string[]` | 그 중 **하나라도** 있으면 되는 바이너리 |
| `requires.config` | `string[]` | 스킬이 읽는 설정 파일 경로 |
| `primaryEnv` | `string` | 메인 크리덴셜 환경 변수 |
| `always` | `boolean` | `true`면 항상 활성 (별도 설치 불요) |
| `skillKey` | `string` | 기본 invocation 키 오버라이드 |
| `emoji` | `string` | 표시용 이모지 |
| `homepage` | `string` | 홈페이지/문서 URL |
| `os` | `string[]` | OS 제약 (`["macos"]`, `["linux"]`) |
| `install` | `array` | 의존성 설치 스펙 (아래) |
| `nix` | `object` | Nix 플러그인 스펙 |

**Install spec** — 지원 종류: `brew`, `node`, `go`, `uv`, `download`
```yaml
install:
  - kind: brew
    formula: tene
    bins: [tene]
    label: "Install Tene via Homebrew"
```

**보안 분석 핵심 규칙**: 스킬 코드가 참조하는 env/bin을 **프론트매터에 정확히 선언**해야 한다. 선언과 실제 동작이 불일치하면 "metadata mismatch"로 플래그된다.

### 1.5 CLI 명령어 (발행자/소비자 공통)

```bash
# 인증
clawhub login                 # 브라우저 OAuth (GitHub)
clawhub login --token clh_... # 헤드리스 모드
clawhub whoami

# 탐색
clawhub search <query>
clawhub explore --sort newest|downloads|rating|trending
clawhub inspect <slug> [--version X | --files | --file PATH]

# 설치/업데이트 (소비자)
clawhub install <slug> [--workdir DIR --dir skills]
clawhub list
clawhub update --all
clawhub uninstall <slug> --yes

# 발행 (개발자)
clawhub skill publish <path> \
  --slug <slug> \
  --name "<display name>" \
  --version 1.0.0 \
  --tags latest \
  --changelog "Initial release"

# 관리
clawhub skill rename <old> <new>    # 기존 slug는 redirect alias로 유지
clawhub skill merge <source> <target>
clawhub delete/undelete <slug>      # soft delete (owner/mod/admin)
clawhub sync                        # 로컬 변경 자동 감지 → 신버전 publish
```

전역 플래그: `--workdir`, `--dir`, `--site`, `--registry`, `--no-input`
설정 파일 위치 (macOS): `~/Library/Application Support/clawhub/config.json`

### 1.6 퍼블리시 요건 (비공식 가이드 종합)

- GitHub 계정이 있어야 하며, 일부 가이드는 **생성 1주일 이상** 권장 (신생 계정 = 낮은 신뢰)
- 퍼블리셔의 GitHub 활동(공개 커밋 기록, 다른 오픈소스 기여)이 보안 신뢰 신호로 쓰임
- "Featured" 선정 기준 (비공식): 품질 점수 4.5+, 100+ 설치, 검증 배지, 완전한 문서, 단일 목적, 90일 내 유지보수

---

## 2. tene CLI용 SKILL.md 작성 가능성 판단

### 2.1 결론: **가능하다. 매우 적합한 케이스다.**

tene은 이미 "AI 에이전트가 Bash로 호출하는" CLI로 설계되어 있고, `CLAUDE.md`에 이미 AI를 위한 사용 규칙이 정리되어 있다. 이걸 그대로 Skill 형태로 패키징하면 된다.

### 2.2 적합성 체크리스트

| 항목 | tene 현황 | Skill 호환성 |
|---|---|---|
| 설치 방법이 표준화되어 있는가 | Homebrew tap, install.sh, GitHub Releases | ✅ `install: kind: brew`로 declarative 설치 가능 |
| CLI가 Bash에서 호출 가능한가 | `tene <subcommand>` | ✅ OpenClaw skill은 Bash 실행 기반 |
| AI 사용 규칙이 이미 존재하는가 | `CLAUDE.md` + `internal/claudemd/template.go` | ✅ 그대로 SKILL.md 본문으로 재활용 |
| 보안 규칙이 명확한가 | "never `tene get`", "never `tene export`" 등 | ✅ SKILL.md "Instructions" 섹션에 명시 |
| 환경 변수 의존성이 있는가 | 없음 (마스터 비밀번호는 OS Keychain에서 자동 로드) | ✅ `requires.env` 비워둠 → clean |
| Slug 요건 (`^[a-z0-9-]+$`) | `tene` | ✅ 적합 |
| 50MB 번들 제한 | SKILL.md + README만 포함 시 <100KB | ✅ 여유 |
| MIT-0 라이선스 수용 가능한가 | 프로젝트 자체가 MIT | ✅ 호환 |

### 2.3 전략적 포지셔닝

Skill 이름과 설명은 **AI가 "언제 이 스킬을 쓸지" 판단하는 유일한 트리거**다. 따라서 트리거 품질이 검색 노출과 자동 호출 확률을 결정한다.

- **`description`이 핵심**: "when to use" 시그널을 담아야 한다 (예: "secrets", "API keys", "env injection")
- **차별화 포인트**: "로컬 암호화, 서버리스, zero-knowledge" — 클라우드 시크릿 매니저(Doppler, Infisical, Vercel Env)와 명확히 구분됨
- **AI 안전 기본값 내장**: "`tene get` 절대 금지" 같은 규칙이 스킬에 박혀 있다는 점이 tene의 셀링 포인트

---

## 3. tene-cli Skill 설계안

### 3.1 폴더 구조 (제안)

레포 내 위치 후보 두 가지:
- **A안 (권장)**: `skills/tene-cli/` — 별도 디렉터리로 격리, `clawhub sync` 친화적
- **B안**: `packaging/clawhub/tene-cli/` — 배포 관련 산출물과 함께 묶기

```
skills/tene-cli/
├── SKILL.md              # Skill 본문
├── README.md             # 마켓플레이스 표시용 문서
├── examples/
│   ├── install-and-init.md
│   ├── inject-into-npm-start.md
│   └── ci-cd-env.md
└── tests/
    └── test.md           # 선택: 기대 동작 기술
```

### 3.2 SKILL.md 최종안

````markdown
---
name: tene-cli
description: Local-first encrypted secret management with the tene CLI. Use when the user mentions API keys, secrets, .env files, credentials, environment variables, or asks to inject secrets into a command. Teaches safe secret handling (never print plaintext in chat) and uses `tene run -- <cmd>` for injection.
version: 1.0.0
metadata:
  openclaw:
    requires:
      bins:
        - tene
    install:
      - kind: brew
        formula: tomo-kay/tene/tene
        bins: [tene]
        label: "Install tene via Homebrew tap"
      - kind: download
        url: https://get.tene.sh
        bins: [tene]
        label: "Install tene via install.sh"
    emoji: "🔐"
    homepage: https://tene.sh
    os: [macos, linux]
---

# tene — Local-First Encrypted Secret Management

## When to use this skill

Activate this skill when the user:
- mentions **API keys, secrets, credentials, tokens, .env files, or environment variables**
- asks you to **run a dev/test/prod command that needs secrets** (npm start, go run, pytest, docker compose, etc.)
- says something like "store this key", "save this token", "how do I manage secrets", "inject env vars"
- works in a repo containing a `.tene/` directory or a `CLAUDE.md` that references tene

Do **not** activate for: hardcoded config values, public constants, feature flags without sensitive data.

## Critical safety rules (for AI agents)

These are **non-negotiable**. Violating them leaks plaintext secrets into the conversation context, which may be logged or retained.

1. **NEVER run `tene get <KEY>`** — plaintext value appears in stdout → enters AI context. If the user needs to inspect a value, instruct them to run it themselves in a separate terminal.
2. **NEVER run `tene export`** (without args) — dumps all secrets to stdout in `.env` format.
3. **NEVER `cat`, `Read`, or open files under `.tene/`** — this is the encrypted vault DB. Even encrypted bytes should stay out of context.
4. **NEVER pass secrets as CLI arguments** — they appear in `ps`, shell history, and logs. Always use env injection via `tene run --`.
5. **Use `tene list` to check which keys exist** — it outputs names only, never values. Safe for AI context.

## Standard workflows

### Check what exists
```bash
tene list                          # current env
tene list --env prod               # specific env
tene env list                      # all environments
```

### Add a new secret
If you know the user wants to store a value, ask them to run it (never type the value yourself):
> "Please run: `tene set STRIPE_KEY <your-value>` — or pipe from a file with `cat key.txt | tene set STRIPE_KEY --stdin`."

### Run a command with secrets injected
```bash
tene run -- npm start              # Node.js
tene run -- go run ./cmd/app       # Go
tene run -- python main.py         # Python
tene run --env staging -- pytest   # with env override
```

`tene run` loads the encrypted vault, derives the master key from the OS keychain (or prompts once), and injects every secret as an env var for the child process only. Secrets never touch disk unencrypted.

### Initialize a new project
```bash
tene init                          # creates .tene/vault.db + master password + BIP-39 recovery key + CLAUDE.md
```

This generates a `CLAUDE.md` with project-specific safety rules. Always read that file after `init`.

### Import from a legacy .env
```bash
tene import .env                   # bulk load
rm .env                            # then delete the plaintext file
```

### Environment management
```bash
tene env local                     # switch default env
tene set API_KEY "..." --env prod  # set in specific env
tene run --env prod -- ./deploy.sh
```

Note: in `tene run`, `--env` must come **before** `--` (tene disables flag parsing after `--` for passthrough).

## Install

```bash
# macOS / Linux (recommended)
brew install tomo-kay/tene/tene

# Alternative
curl -fsSL https://get.tene.sh | sh

# Verify
tene version
```

After install, `tene init` bootstraps a vault per project. Run it once in each project root.

## Troubleshooting

| Symptom | Fix |
|---|---|
| `VAULT_NOT_FOUND` | Not in a tene project. Run `tene init` in the repo root. |
| `SECRET_NOT_FOUND` | Key missing in current env. Check `tene list --env <name>`. |
| Master password prompt on every run | Keychain access denied. Run with `--no-keychain` to debug, then re-grant permission in Keychain Access (macOS). |
| `tene run --env prod -- cmd` uses wrong env | `--env` placed **after** `--`. Move it before the separator. |

## Further reading

- Homepage: https://tene.sh
- Source: https://github.com/tomo-kay/tene
- Security model: XChaCha20-Poly1305 + Argon2id + HKDF + X25519
- Recovery: 12-word BIP-39 mnemonic generated at `tene init`
````

### 3.3 README.md (마켓플레이스 표시용, 요지)

README.md는 ClawHub 리스팅 페이지에 같이 렌더링되므로, 독립적으로도 이해 가능해야 한다. SKILL.md가 "AI용 지침"이라면 README는 "사람용 개요"다.

포함할 섹션:
1. **한 줄 소개** + 데모 GIF (선택)
2. **왜 tene인가** (로컬 퍼스트, 제로 지식, 무료)
3. **Quick Start** — 5줄짜리 설치/사용 예
4. **AI 에이전트와의 통합** — 이 Skill이 제공하는 가치 강조
5. **라이선스 고지** — Skill은 MIT-0, tene CLI 자체는 MIT

### 3.4 examples/ 디렉터리

각 예제는 `## 제목` + `Input:` + `Expected behavior:` 블록으로 구성한다.

- **`install-and-init.md`**: 새 Mac에서 tene 설치 → `tene init` → 첫 secret 저장까지
- **`inject-into-npm-start.md`**: Node 프로젝트에서 `.env` 제거 후 `tene run -- npm start`로 전환
- **`ci-cd-env.md`**: GitHub Actions에서 `tene run --env prod` 사용 패턴 (OS Keychain이 없으므로 `TENE_MASTER_PASSWORD` 환경 변수 주입 방식 설명)

### 3.5 tests/test.md (선택)

```markdown
## List existing secrets safely
Input: User asks "which API keys do I have?"
Expected behavior:
  - Agent runs `tene list` (not `tene get`, not `tene export`)
  - Output shows names only, no values

## Inject secrets into dev server
Input: User asks "start the dev server"
Expected behavior:
  - Agent runs `tene run -- npm start` (or equivalent per stack)
  - Does not create or read a `.env` file
  - Does not pass secrets as CLI arguments

## Refuse to print secret values
Input: User asks "what's the value of STRIPE_KEY?"
Expected behavior:
  - Agent refuses to run `tene get STRIPE_KEY`
  - Instructs user to run it in a separate terminal themselves
```

---

## 4. ClawHub 등록 단계별 절차

### 4.1 사전 준비 (최초 1회)

1. **GitHub 계정** 확인 — 가급적 공개 활동 이력이 있는 계정 사용
2. **clawhub CLI 설치**
   ```bash
   # clawhub CLI는 Bun 기반. Bun이 없으면:
   curl -fsSL https://bun.sh/install | bash
   # 이후:
   bun install -g clawhub
   # 또는 npx/npm으로도 배포됨 (현재 확인 필요)
   ```
3. **로그인**
   ```bash
   clawhub login
   # 브라우저 열림 → GitHub OAuth → 터미널로 콜백
   clawhub whoami   # 확인
   ```

### 4.2 Skill 빌드 & 로컬 검증

1. 본 레포에 `skills/tene-cli/` 디렉터리 생성하고 위 §3의 파일들을 배치
2. **로컬 검증**
   ```bash
   cd skills/tene-cli
   clawhub inspect .                 # (폴더 경로 지원 시) — 지원 안 되면 생략
   # SKILL.md YAML 파싱 오류가 있으면 publish 시 즉시 fail
   ```
3. **드라이 런**
   ```bash
   clawhub sync --dry-run --no-input
   ```

### 4.3 최초 퍼블리시

```bash
cd /path/to/tene

clawhub skill publish ./skills/tene-cli \
  --slug tene-cli \
  --name "Tene CLI — Local-First Secrets" \
  --version 1.0.0 \
  --tags latest \
  --changelog "Initial release: install, vault init, safe secret injection workflows for AI agents."
```

성공 시 응답으로 `https://clawhub.ai/<handle>/tene-cli` 형태 URL이 반환된다.

### 4.4 발행 후 확인

- 웹: https://clawhub.ai/search?q=tene → 검색 결과에 노출되는지 (벡터 인덱싱에 수분 소요될 수 있음)
- 설치 테스트 (다른 머신이나 깨끗한 workdir에서):
  ```bash
  mkdir /tmp/tene-skill-test && cd /tmp/tene-skill-test
  clawhub install tene-cli
  clawhub list
  cat skills/tene-cli/SKILL.md  # 올바르게 배포되었는지 확인
  ```
- Claude Code / OpenClaw에서 활성화 (사용자 기기):
  ```bash
  # OpenClaw
  openclaw skills install tene-cli
  # 또는 Claude Code는 자체 skill 디렉터리에 복사되도록 경로 지정
  ```

### 4.5 버전 업데이트 워크플로

SKILL.md 또는 본문을 수정한 뒤:

```bash
# 방법 1: 명시적 publish (엄격 제어)
clawhub skill publish ./skills/tene-cli \
  --slug tene-cli \
  --version 1.1.0 \
  --tags latest \
  --changelog "Add CI/CD env injection guidance and Linux keychain fallback notes."

# 방법 2: sync (여러 스킬 일괄 관리 시)
clawhub sync --all
```

Semver 규칙을 지킬 것: bug fix = patch (1.0.1), 내용 추가 = minor (1.1.0), 호환성 깨짐 = major (2.0.0).

### 4.6 CI 자동 퍼블리시 (권장)

`.github/workflows/clawhub-publish.yml`:

```yaml
name: ClawHub Skill Publish

on:
  push:
    branches: [main]
    paths:
      - 'skills/tene-cli/**'
  workflow_dispatch:

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: oven-sh/setup-bun@v1
      - name: Install clawhub CLI
        run: bun install -g clawhub
      - name: Publish / sync
        env:
          CLAWHUB_TOKEN: ${{ secrets.CLAWHUB_TOKEN }}
        run: |
          clawhub login --token "$CLAWHUB_TOKEN"
          clawhub sync --all --no-input
```

`CLAWHUB_TOKEN`은 `clawhub login` 후 `~/Library/Application Support/clawhub/config.json`에 저장된 토큰을 GitHub Secrets에 등록한다. 이 토큰 자체도 tene으로 관리하는 것을 권장한다:
```bash
tene set CLAWHUB_TOKEN "clh_..." --env prod
# GitHub CLI로 주입
tene run --env prod -- gh secret set CLAWHUB_TOKEN --body "$CLAWHUB_TOKEN"
```

### 4.7 운영 관리

| 작업 | 명령 |
|---|---|
| 검색 순위 모니터링 | `clawhub explore --sort installs` |
| 리네임 (slug 변경) | `clawhub skill rename tene-cli tene` — 기존 slug는 redirect alias |
| 버전 히스토리 | `clawhub inspect tene-cli --versions` |
| 특정 버전 조회 | `clawhub inspect tene-cli --version 1.0.0 --files` |
| 임시 숨김 | `clawhub hide tene-cli` / `clawhub unhide tene-cli` |
| 소유권 이전 | `clawhub transfer request tene-cli <newOwnerHandle>` |

---

## 5. 리스크와 대응

| 리스크 | 영향 | 대응 |
|---|---|---|
| AI가 `tene get`을 실수로 실행 | 시크릿 값이 대화 로그에 유출 | SKILL.md §Safety rules에 최상단 배치. 예제에서도 반복 강조 |
| 번들에 실수로 `.tene/` 포함 | 암호화된 vault가 공개 레지스트리에 업로드됨 | `.clawhubignore`에 `.tene/`, `.env*` 명시. publish 전 `clawhub inspect --files`로 확인 |
| install spec 오류 (`brew formula` 오타) | 설치 단계에서 조용히 실패 | goreleaser tap 경로와 formula 이름을 CI에서 검증 |
| MIT-0 라이선스와 기존 MIT 혼동 | 저작권 고지 기대 불일치 | README에 "Skill 자체는 MIT-0, tene CLI 자체는 MIT"를 명시 |
| tene CLI 바이너리 부재 (clean env) | `requires.bins: [tene]`만으로는 자동 설치 트리거 안 됨 | `install:` spec 반드시 같이 선언. brew tap 우선, 폴백으로 `install.sh` |
| `description`이 너무 일반적 | 의미 검색에서 안 잡힘 | "secrets", "env injection", "API keys", "credentials", "zero-knowledge" 등 구체 키워드 포함 |

---

## 6. 다음 액션 (체크리스트)

- [ ] `skills/tene-cli/` 폴더 생성 + 본 문서 §3의 파일 배치
- [ ] Homebrew tap formula 최종 이름 확인 (`tomo-kay/tene/tene` 검증)
- [ ] `.clawhubignore` 추가 (`.tene/`, `.env*`, `dist/`, `node_modules/`)
- [ ] `clawhub login`으로 토큰 발급 → `tene set CLAWHUB_TOKEN ...`으로 저장
- [ ] 로컬에서 `clawhub skill publish ./skills/tene-cli --version 1.0.0 --tags latest`로 최초 발행
- [ ] 깨끗한 디렉터리에서 `clawhub install tene-cli` 테스트 → Skill 렌더링/문법 확인
- [ ] GitHub Actions 워크플로 추가 (§4.6)
- [ ] 주요 Claude Code 커뮤니티(Discord, Reddit r/ClaudeAI)에 발표
- [ ] 사용 피드백 기반 1.1.0 마이너 업데이트

---

## 참고 출처

- [ClawHub 공식 사이트](https://clawhub.ai)
- [ClawHub GitHub 레포](https://github.com/openclaw/clawhub)
- [Skill 포맷 공식 명세 (skill-format.md)](https://github.com/openclaw/clawhub/blob/main/docs/skill-format.md)
- [Quickstart 공식 문서](https://github.com/openclaw/clawhub/blob/main/docs/quickstart.md)
- [CLI 공식 레퍼런스](https://github.com/openclaw/clawhub/blob/main/docs/cli.md)
- [OpenClaw Skills 문서](https://docs.openclaw.ai/tools/skills)
- [Felo Search 블로그 — ClawHub 소개](https://felo.ai/blog/clawhub-skills-marketplace-claude-code/)
- [Aiskill — ClawHub 퍼블리싱 가이드 (비공식)](https://aiskill.market/blog/clawhub-skill-publishing-guide)
- [ClawHub Skill Creator](https://clawhub.ai/chindden/skill-creator)
