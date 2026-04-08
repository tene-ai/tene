# MVP 범위 브레인스토밍 v4
## Go CLI + Claude Code 전용, Cloud 제외, Fake Door Test

> v4 (2026-04-06) — Go 언어 전환 + Claude Code 전용 MVP 반영
> 목적: Go CLI + Cloud 제외 + Claude Code 전용 MVP 최소 범위, Fake Door Test 정의

---

## 1. MVP 철학: Go 바이너리, 서버 비용 $0, Claude Code가 자동 인식

### 1.1 핵심 질문 (v4)

> "솔로 바이브코더가 서버 가입 없이, Node.js 없이, brew install 하나로,
> 1분 안에 '와, Claude Code가 알아서 인식하네!'라고 느낄 수 있는 최소 기능은 무엇인가?"

### 1.2 MVP 원칙 (v4)

| 원칙 | v3 | v4 |
|------|----|----|
| **One Job** | 서버 없이 시크릿 저장+주입 + AI 자동 인식 | 서버 없이 시크릿 저장+주입 + **Claude Code 자동 인식** |
| **1 Minute Magic** | 가입 없이 설치→AI 인식→사용 1분 | 가입 없이 **brew 한 줄→AI 인식→사용 1분** |
| **No Server** | 서버 없음 (MVP), Cloud = Phase 2 | 유지 |
| **Zero Dependencies** | Node.js 런타임 필요 | **Go 단일 바이너리, 런타임 불필요** |
| **AI Native** | CLAUDE.md/.cursorrules 자동 생성 | **CLAUDE.md만 (MVP), .cursorrules는 Phase 2** |
| **Fake Door** | `tene sync` Fake Door (waitlist) | 유지 |
| **Honest** | 못 하는 것 정직하게 명시 | 유지 |

### 1.3 MVP = Go CLI + modernc.org/sqlite + golang.org/x/crypto

```
brew install tomo-kay/tap/tene → tene init → tene set → tene run
```

- Go 단일 바이너리: ~10-15MB, ~5ms 시작, Node.js 불필요
- 서버 없음, 회원가입 없음, 비용 없음
- CLAUDE.md 자동 생성 (Claude Code 자동 인식)
- `tene export --encrypted`로 수동 백업
- `tene sync` = Fake Door (waitlist 안내만)

---

## 2. MVP 기능 범위 (Cloud 완전 제외, Claude Code 전용)

### Phase 1: Go CLI MVP (서버 없음, 무료, $0)

**"서버 없이, 가입 없이, Claude Code가 자동 인식하는 Go CLI"**

| # | 기능 | 명령어 | 우선순위 | 근거 |
|---|------|--------|:--------:|------|
| L1 | **프로젝트 초기화 + CLAUDE.md** | `tene init` | P0 | 볼트 생성 + **CLAUDE.md 자동 생성** |
| L2 | **시크릿 저장** | `tene set KEY VALUE` | P0 | 핵심 가치의 시작점 |
| L3 | **시크릿 조회** | `tene get KEY` | P0 | AI 에이전트 Bash 호출 핵심 |
| L4 | **시크릿 주입** | `tene run -- CMD` | P0 | "한 줄이면 끝" 핵심 가치 |
| L5 | **시크릿 목록** | `tene list` | P0 | 시크릿 현황 파악 |
| L6 | **시크릿 삭제** | `tene delete KEY` | P0 | 기본 CRUD |
| L7 | **.env 가져오기** | `tene import .env` | P1 | 전환 비용 제로 |
| L8 | **.env 내보내기** | `tene export` | P1 | 기존 도구 호환성 |
| L9 | **암호화 백업** | `tene export --encrypted` | P1 | Cloud 없이 수동 백업 |
| L10 | **환경 전환** | `tene env [dev/prod]` | P1 | 개발/프로덕션 분리 |
| L11 | **Fake Door** | `tene sync` | P1 | Cloud 수요 확인 (waitlist) |
| L12 | **Master Password 암호화** | (내부) | P0 | 보안의 전제 조건 |
| L13 | **Recovery Key 생성** | (내부) | P0 | 패스워드 분실 대비 |
| L14 | **.gitignore 자동 추가** | (내부) | P0 | .tene/ 노출 방지 |

### Phase 2: 멀티 에디터 + Cloud (수요 검증 후)

**"Claude Code 외 AI 에디터 지원 + Cloud 동기화"**

| # | 기능 | 설명 | 전제 조건 |
|---|------|------|----------|
| M1 | **Cursor 통합** | `tene init --cursor` (.cursorrules) | Claude Code MVP 성공 |
| M2 | **Windsurf 통합** | `tene init --windsurf` (.windsurfrules) | Claude Code MVP 성공 |
| C1 | Cloud 인증 | `tene login` | waitlist 100명+ |
| C2 | 암호화 클라우드 백업 | 볼트 blob → S3 | waitlist 확인 |
| C3 | 멀티 디바이스 동기화 | 자동 동기화 | waitlist 확인 |
| C4 | 웹 대시보드 | 시크릿 현황, 감사 로그 | Cloud 출시 후 |

**Cloud 인프라 (서버리스 사용 안 함):**
- ECS Fargate + NLB + RDS PostgreSQL + S3

### Phase 2+: 팀 기능 (가설)

| # | 기능 | 상태 |
|---|------|:----:|
| T1 | 팀 볼트 | **가설** — Fake Door 후 결정 |
| T2 | RBAC | **가설** |
| T3 | 에이전트 스코핑 | Phase 2+ |
| T4 | MCP 서버 | Phase 2 |
| T5 | 자동 로테이션 | **못 함** (Phase 2+ 가설) |
| T6 | API key 만료 확인 | **못 함** (Phase 2+ 가설) |

---

## 3. YAGNI 분석: MVP에서 제외하는 것

### 3.1 "지금 만들면 안 되는" 기능

**1. Cursor/Windsurf 지원 (Phase 2)**
- 이유: **Claude Code 전용으로 MVP 완성도를 높이는 것이 우선**
- `--cursor`, `--windsurf` 플래그는 Phase 2에서 추가
- MVP에서 CLAUDE.md 자동 생성의 가치가 검증되면 다른 에디터로 확장

**2. 클라우드 서버/API (Phase 2)**
- 이유: **수요 미확인 상태에서 서버 구축은 낭비**. Fake Door로 수요 먼저 확인
- 대신: `tene sync` Fake Door + `tene export --encrypted` 수동 백업

**3. 회원가입/로그인 (Phase 2)**
- 이유: **가입이 필요 없는 것이 핵심 차별점**
- 대신: Phase 2 Cloud 가입 시에만

**4. 웹 대시보드 (Phase 2)**
- 이유: CLI로 핵심 가치 전달 가능
- 대신: `tene list`, `tene export`로 CLI에서 관리

**5. MCP 서버 (Phase 2)**
- 이유: Claude Code는 **CLAUDE.md 자동 인식 + CLI Bash 호출**이면 충분
- MCP 설정에 시크릿 저장하지 않아도 됨 → 오히려 보안적 이점

**6. 자동 로테이션 (못 함)**
- 이유: 각 서비스 API 개별 통합 필요, 구현 비용 매우 높음
- 정직하게 "못 하는 것"으로 명시. Phase 2+ 가설

**7. API key 만료 확인 (못 함)**
- 이유: 서비스마다 만료 확인 API가 다르거나 없음
- 정직하게 "못 하는 것"으로 명시

**8. Node.js/npm 배포 (폐기)**
- 이유: **Go 단일 바이너리가 모든 면에서 우수** (설치 속도, 시작 속도, 의존성, 보안)
- 대신: Homebrew tap + curl 설치 + go install

**9. 서버리스 인프라 (사용 안 함)**
- 이유: 트래픽 늘면 비용 역전. ECS Fargate가 예측 가능
- Phase 2 Cloud 구축 시: ECS Fargate + NLB + RDS PostgreSQL + S3

### 3.2 반드시 있어야 하는 "숨겨진" 기능

| # | 기능 | 이유 |
|---|------|------|
| H1 | **OS Keychain 세션 캐시** | 매 명령마다 패스워드 입력하면 사용 불가 (go-keyring) |
| H2 | **Recovery Key 생성** | Master Password 분실 시 유일한 복구 수단 (go-bip39) |
| H3 | **CLAUDE.md 자동 생성** | Claude Code 자동 인식의 핵심 메커니즘 |
| H4 | **에러 메시지 품질** | "tene init으로 먼저 초기화하세요" |
| H5 | **.gitignore 자동 추가** | .tene/ 노출 방지 |
| H6 | **시크릿 값 마스킹** | `tene list`에서 값 마스킹 |
| H7 | **빠른 응답** | ~5ms (Go 단일 바이너리의 자연스러운 장점) |
| H8 | **Master Password 변경** | `tene passwd` 명령어 |

---

## 4. Claude Code 자동 인식 기능 상세

### 4.1 `tene init` → CLAUDE.md 자동 생성

```
$ tene init
? Master Password: ********
? Confirm Password: ********

> 로컬 볼트 생성: .tene/vault.db (XChaCha20-Poly1305)
> .gitignore에 .tene/ 추가 완료
> CLAUDE.md 생성 완료 (Claude Code 자동 인식)
> Recovery Key: apple banana cherry ... (안전한 곳에 보관하세요!)
```

**생성되는 CLAUDE.md 내용:**
```markdown
# Secrets Management

This project uses [tene](https://github.com/agentkay/tene) for secret management.

## Usage
- Get a secret: `tene get <KEY>`
- List secrets: `tene list`
- Run with secrets injected: `tene run -- <command>`
- Set a secret: `tene set <KEY> <VALUE>`

## Rules
- Never hardcode secret values in source code
- Access secrets via `process.env.KEY_NAME`
- Do not create .env files — use `tene run` instead
- Use `tene list` to see available secrets
```

### 4.2 Cursor/Windsurf 지원 (Phase 2)

MVP에서는 Claude Code 전용에 집중한다:
- `tene init --cursor` → Phase 2 (.cursorrules 자동 추가)
- `tene init --windsurf` → Phase 2 (.windsurfrules 자동 추가)

### 4.3 기존 CLAUDE.md가 있는 경우

- 기존 CLAUDE.md가 있으면 하단에 tene 섹션을 **추가** (덮어쓰지 않음)
- 이미 tene 섹션이 있으면 스킵

---

## 5. Fake Door Test: `tene sync`

### 5.1 구현

```
$ tene sync

  +----------------------------------------------------+
  | 클라우드 동기화 기능을 준비하고 있습니다!              |
  |                                                    |
  | 관심이 있으시면 waitlist에 등록하세요:               |
  | https://tene.sh/cloud                             |
  |                                                    |
  | 현재는 tene export --encrypted로                   |
  | 수동 암호화 백업을 권장합니다.                       |
  +----------------------------------------------------+
```

### 5.2 수요 검증 프로세스

```
[Step 1] CLI 출시 → 기본 지표 추적
   Homebrew/curl/go install 설치 수, GitHub Stars
   성공 기준: 주간 설치 500+, Stars 1,000+

[Step 2] tene sync Fake Door → Cloud 수요 확인
   tene sync 실행 → waitlist 페이지
   성공 기준: CLI 사용자의 10%+ waitlist 등록

[Step 3] waitlist 반응 → Cloud 구축 결정
   100명+ → Cloud 구축 시작
   < 50명 → Cloud 보류, CLI 기능 강화
```

---

## 6. Go CLI MVP 핵심 플로우 상세

### 6.1 첫 사용 플로우 (1분 목표)

```
[0:00] $ brew install tomo-kay/tap/tene
       > tene installed (~5초, Node.js 불필요)

[0:05] $ cd my-project
       $ tene init
       ? Master Password: ********
       ? Confirm Password: ********
       > 로컬 볼트 생성: .tene/vault.db
       > .gitignore에 .tene/ 추가
       > CLAUDE.md 생성 완료 (Claude Code 자동 인식)
       > Recovery Key: apple banana cherry ... (보관하세요!)
       (15초)

[0:20] $ tene import .env
       > 5개 시크릿 가져오기 완료 (암호화)
       (5초)

[0:25] $ tene run -- claude
       > 5 secrets injected as environment variables
       > Starting: claude
       (5초)

[0:30] 완료! 30초 만에.
       Claude Code가 CLAUDE.md를 읽고 tene을 자동 인식.
       서버 없음. 가입 없음. Node.js 없음. 비용 $0.
```

### 6.2 Claude Code 사용 플로우 (CLAUDE.md 자동 인식)

```
[Claude Code 세션 - CLAUDE.md가 있는 프로젝트]

사용자: "Stripe 결제 기능 만들어줘"

Claude Code:
  (CLAUDE.md 읽음: "This project uses tene for secret management")
  
  "이 프로젝트는 tene으로 시크릿을 관리하고 있습니다.
   STRIPE_KEY가 필요합니다. 환경변수로 참조하겠습니다."
  
  코드 생성:
  import Stripe from 'stripe';
  const stripe = new Stripe(process.env.STRIPE_KEY!);
  
  "tene run -- npm run dev 로 실행하면 STRIPE_KEY가 자동 주입됩니다."
```

### 6.3 수동 백업 플로우

```
$ tene export --encrypted > ~/backup/my-project-$(date +%Y%m%d).enc
> 암호화된 백업 파일 생성 완료

# 복원
$ tene import --encrypted ~/backup/my-project-20260406.enc
? Master Password: ********
> 5개 시크릿 복원 완료
```

---

## 7. 기술적 MVP 범위

### 7.1 Go CLI 아키텍처

```
cmd/tene/                  ← Go CLI 엔트리포인트
  main.go
internal/
  commands/                ← cobra CLI 명령어
    init.go                # Master Password + 볼트 + CLAUDE.md 생성
    set.go                 # 시크릿 암호화 저장
    get.go                 # 시크릿 복호화 조회
    run.go                 # 환경변수 주입 실행
    list.go                # 시크릿 목록 (마스킹)
    delete.go              # 시크릿 삭제
    env.go                 # 환경 전환
    import.go              # .env / --encrypted 가져오기
    export.go              # .env / --encrypted 내보내기
    recover.go             # Recovery Key로 복구
    passwd.go              # Master Password 변경
    sync.go                # Fake Door Test (waitlist 안내)
  crypto/                  ← 암호화 모듈
    encryption.go          # XChaCha20-Poly1305 (golang.org/x/crypto/nacl/secretbox)
    kdf.go                 # Argon2id (golang.org/x/crypto/argon2)
    keys.go                # Master Key, Recovery Key 관리
  vault/                   ← 로컬 저장소
    sqlite.go              # SQLite 볼트 관리 (modernc.org/sqlite)
    session.go             # OS Keychain 세션 캐시 (go-keyring)
  claudemd/                ← Claude Code 통합
    generate.go            # CLAUDE.md 생성
  recovery/                ← 복구
    bip39.go               # BIP-39 니모닉 (tyler-smith/go-bip39)
  keychain/                ← OS Keychain 연동
    keyring.go             # zalando/go-keyring
apps/web/                  ← Next.js 랜딩페이지 (유지)
go.mod
go.sum
.goreleaser.yml            ← goreleaser 멀티 플랫폼 빌드 설정
```

**v3 대비 핵심 변경**:
- TypeScript → Go 전환
- Commander.js → cobra
- better-sqlite3 → modernc.org/sqlite (순수 Go, CGo 없음)
- libsodium-wrappers → golang.org/x/crypto (네이티브 Go)
- keytar → go-keyring
- BIP-39 npm → go-bip39
- npm publish → goreleaser (Homebrew tap + 멀티 플랫폼 바이너리)
- `agent/cursor.ts` 삭제 → Phase 2
- pnpm workspace → Go modules + apps/web (Next.js)

### 7.2 서버 API (Phase 2에서만, 서버리스 X)

> MVP에는 서버 엔드포인트가 **0개**.
> Phase 2에서 ECS Fargate + NLB + RDS PostgreSQL + S3로 구축.

---

## 8. MVP 성공 기준 (v4)

### 8.1 출시 후 30일 목표

| 지표 | 목표 | 측정 |
|------|------|------|
| 설치 수 (brew+curl+go) | 2,000+ | GitHub Releases + Homebrew analytics |
| GitHub Stars | 500+ | GitHub |
| `tene sync` Fake Door 실행 | 100+ | (간접 추정) |

### 8.2 출시 후 90일 목표

| 지표 | 목표 | 측정 |
|------|------|------|
| 총 설치 수 | 10,000+ | GitHub Releases |
| GitHub Stars | 2,000+ | GitHub |
| waitlist 등록 | 100+ | tene.sh/cloud |
| Cloud 구축 결정 | Yes/No | waitlist 분석 |

---

*Tene v4 Brainstorming Document 03/05*
