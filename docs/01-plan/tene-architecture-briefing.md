# Tene 아키텍처 브리핑

> v4.0 (2026-04-06) -- Go 전환, Claude Code 전용 MVP, brew/curl 배포
>
> 목적: Tene의 각 구성요소가 무엇을 하고, 어떻게 연결되는지 개발자 관점에서 설명

---

## 한 줄 요약

**Go 단일 바이너리로 로컬에서 무료(서버 비용 $0)로 시크릿을 암호화 관리하는 CLI. Claude Code AI Agent가 자동으로 tene 사용법을 인식하는 것이 핵심 차별점.**

---

## 1. 전체 구조

```
[MVP -- Phase 1] 사용자 디바이스 (서버 없음, 오프라인 동작, 비용 $0)
+-------------------------------------------------------+
|                                                         |
|  tene CLI (Go 단일 바이너리, CGo 불필요)                 |
|       |                                                 |
|       +--- internal/crypto --- OS Keychain (go-keyring) |
|       |    Argon2id + XChaCha20-Poly1305                 |
|       |    (golang.org/x/crypto)                         |
|       |                                                 |
|       +--- SQLite (.tene/vault.db)                      |
|       |    암호화된 시크릿 저장                            |
|       |    (modernc.org/sqlite, 순수 Go)                 |
|       |                                                 |
|       +--- Claude Code 자동 인식                         |
|            CLAUDE.md 자동 생성                            |
|                                                         |
|  tene sync -> Fake Door (waitlist 안내)                  |
|  tene export --encrypted -> 수동 암호화 백업              |
|                                                         |
+-------------------------------------------------------+

[Phase 2 -- 수요 검증 후] AWS Cloud ($1/월)
+-------------------------------------------------------+
|                                                         |
|  NLB --- ECS Fargate (Hono API)                         |
|               |                                         |
|          +----+----+                                    |
|          |         |                                    |
|     RDS PostgreSQL  S3 Bucket                            |
|     (users, audit)  (암호화된 vault blob)                 |
|                                                         |
|  CloudFront --- S3 (대시보드 정적 파일)                  |
|  Route 53 (tene.sh, api.tene.sh)                      |
|                                                         |
+-------------------------------------------------------+
```

---

## 2. Tene가 해결하는 것 / 못 하는 것

### 해결하는 것 (MVP)

- **안전 저장**: XChaCha20-Poly1305 암호화 + SQLite. .env 평문 대비 완전한 암호화
- **환경변수 주입**: `tene run -- <command>`로 시크릿을 환경변수로 자동 주입
- **Claude Code 자동 인식**: `tene init` 시 CLAUDE.md 자동 생성. Claude Code가 즉시 tene을 인식
- **암호화 백업**: `tene export --encrypted`로 수동 백업
- **즉시 설치**: Go 단일 바이너리. Node.js 런타임 불필요. brew/curl로 수초 내 설치

### 못 하는 것 (MVP 한계)

- **만료 확인**: 시크릿 만료일 설정/알림 불가
- **자동 갱신**: 시크릿 자동 로테이션 불가
- **프로젝트 간 글로벌 키 공유**: 프로젝트별 독립 볼트 (MVP)
- **Cloud 동기화 / 대시보드**: Phase 2에서 제공
- **Cursor/Windsurf 지원**: Phase 2에서 --cursor, --windsurf 플래그 추가

---

## 3. 구성요소별 역할

### 3.1 Tene CLI (핵심 -- 무료, 오픈소스)

**뭘 하는가:** 시크릿의 암호화/복호화/저장/주입 + Claude Code 컨텍스트 자동 생성.

**기술:** Go + cobra + modernc.org/sqlite + golang.org/x/crypto

| 명령어 | 하는 일 | 서버 필요 |
|--------|---------|:---------:|
| `tene init` | Master Password 설정, SQLite 볼트 생성, Recovery Key 발급, **CLAUDE.md 자동 생성** | X |
| `tene set KEY VALUE` | XChaCha20-Poly1305 암호화 -> SQLite 저장 | X |
| `tene get KEY` | SQLite에서 읽기 -> 복호화 -> stdout 출력 | X |
| `tene run -- claude` | 모든 시크릿을 환경변수로 주입하고 명령 실행 | X |
| `tene list` | 시크릿 목록 (값은 마스킹) | X |
| `tene delete KEY` | 시크릿 삭제 | X |
| `tene import .env` | .env 파일에서 일괄 가져오기 | X |
| `tene export` | .env 형식으로 내보내기 | X |
| `tene export --encrypted` | **암호화된 볼트 파일로 수동 백업** | X |
| `tene env [name]` | 환경 전환 (dev/staging/prod) | X |
| `tene passwd` | Master Password 변경 + 볼트 재암호화 + 새 Recovery Key 발급 | X |
| `tene recover` | Recovery Key(12단어)로 Master Password 재설정 | X |
| `tene sync` | **Fake Door: Cloud waitlist 안내** | X |

**Claude Code 사용법:**
```bash
# Claude Code가 Bash에서 호출
STRIPE_KEY=$(tene get STRIPE_KEY)

# 또는 JSON 출력
tene get STRIPE_KEY --json
```

**중요한 점:**
- 서버 없이 100% 로컬에서 동작 (비용 $0)
- `tene init`만 하면 Claude Code가 tene을 자동으로 인식 (CLAUDE.md)
- Master Password 없이는 누구도 시크릿을 볼 수 없음
- Go 단일 바이너리. Node.js 런타임 불필요

**설치:**
```bash
# macOS
brew install tomo-kay/tap/tene

# Linux / Windows WSL
curl -fsSL https://tene.sh/install.sh | sh

# Go 사용자
go install github.com/tomo-kay/tene@latest
```

---

### 3.2 Claude Code 자동 인식 (핵심 차별점)

**뭘 하는가:** `tene init` 시 CLAUDE.md를 자동 생성하여, Claude Code가 별도 설정 없이 tene 명령어를 인식.

**자동 생성되는 CLAUDE.md (영어):**
```markdown
# Secrets Management

This project uses [tene](https://github.com/tomo-kay/tene) for secret management.

## Usage
- Get a secret: `tene get <KEY>`
- List secrets: `tene list`
- Run with secrets injected: `tene run -- <command>`
- Set a secret: `tene set <KEY> <VALUE>`

## Rules
- Never hardcode secret values in source code
- Access secrets via environment variables
- Do not create .env files -- use `tene run` instead
- Use `tene list` to see available secrets
```

**왜 핵심 차별점인가:**
- 기존 시크릿 매니저 중 AI Agent 컨텍스트를 자동 생성하는 도구가 없음
- 개발자가 Claude Code에게 "시크릿은 tene으로 관리해"라고 매번 말할 필요 없음
- `tene init` 한 번이면 Claude Code가 영구적으로 tene을 인식

**Phase 2 확장:**
- `--cursor`: .cursorrules에 추가
- `--windsurf`: .windsurfrules에 추가

---

### 3.3 SQLite 로컬 볼트 (.tene/vault.db)

**뭘 하는가:** 암호화된 시크릿과 메타데이터를 프로젝트별로 로컬 저장.

**기술:** modernc.org/sqlite (순수 Go, CGo 불필요)

```
프로젝트별 구조:
my-project/.tene/
+-- vault.db        <- 암호화된 시크릿 + 메타데이터 + 감사 로그
+-- vault.json      <- 프로젝트 이름, 생성일, 활성화된 Agent 목록
+-- .gitignore      <- Git에서 자동 제외

my-project/CLAUDE.md     <- Claude Code 컨텍스트 (tene init이 자동 생성)

글로벌 설정:
~/.tene/
+-- config.json     <- CLI 설정, Fake Door 측정 데이터
```

**왜 SQLite인가:**
- .env 파일 대비: 환경별 분리, 인덱스 검색, 부분 업데이트 가능
- 외부 DB 대비: 설치 불필요, 서버 불필요, 단일 파일
- modernc.org/sqlite: 순수 Go, CGo 불필요, 크로스 컴파일 호환

---

### 3.4 암호화 코어 (internal/crypto)

**뭘 하는가:** 모든 암호화/복호화 로직.

**기술:** golang.org/x/crypto (nacl/secretbox + argon2)

**암호화 플로우:**
```
Master Password (사용자 입력)
    |
    v
Argon2id KDF (memory: 64MB, iterations: 3)
    |
    v
Master Key (256-bit) --> OS Keychain에 저장
    |
    +-- Encryption Key (시크릿 암호화용)
    |       |
    |       v
    |   XChaCha20-Poly1305
    |   - 192-bit random nonce (충돌 안전)
    |   - AAD = 키 이름 (변조 방지)
    |   - 결과: nonce + ciphertext -> SQLite 저장
    |
    +-- Auth Hash (Phase 2: Cloud 인증용)
```

**Recovery Key:**
- `tene init` 시 12단어 니모닉으로 발급 (BIP-39 기반, 예: `apple banana cherry ...`)
- Master Password 분실 시 유일한 복구 수단
- Recovery Key로 Master Key를 암호화해서 볼트에 저장
- `tene recover` 명령어로 Master Password 재설정 가능

---

### 3.5 OS Keychain

**뭘 하는가:** Master Key를 OS 네이티브 보안 저장소에 저장.

**기술:** zalando/go-keyring

| OS | Keychain | 보안 수준 |
|----|----------|----------|
| macOS | Keychain Services | T2/Apple Silicon 하드웨어 암호화 |
| Linux | libsecret (GNOME Keyring) | 로그인 세션 암호화 |
| Windows | Credential Vault | DPAPI 암호화 |

---

### 3.6 Fake Door Test (tene sync)

**뭘 하는가:** `tene sync` 실행 시 Cloud waitlist 안내를 표시. 실제 동기화는 미구현.

**수요 검증 기준:**
- `tene sync` 실행률 15%+ 또는 waitlist 100명+ -> Phase 2 시작
- 5% 미만 -> Cloud 보류, 로컬 기능 강화

**대안:** `tene export --encrypted`로 암호화된 볼트 파일을 USB, 클라우드 드라이브 등에 수동 백업

---

## 4. Phase 2 Cloud 구성요소 (수요 검증 후)

> 아래 구성요소는 Phase 2에서 구현한다. ECS Fargate + RDS PostgreSQL 사용.

### 4.1 ECS Fargate + NLB (Cloud API)

**뭘 하는가:** Cloud 유료 사용자를 위한 인증, 볼트 동기화, 감사 로그, 결제 처리.

**기술:** Hono (Node.js) + ECS Fargate + Network Load Balancer

**중요: 이 서버는 시크릿 평문을 절대 모른다.** 암호화된 blob을 중계할 뿐.

### 4.2 RDS PostgreSQL (Cloud DB)

**뭘 하는가:** Cloud 사용자 계정, 디바이스, 동기화 상태, 감사 로그 저장.

**시크릿은 여기에 저장되지 않는다.** 시크릿 blob은 S3에 저장.

### 4.3 S3 Bucket (볼트 백업 저장소)

**뭘 하는가:** 암호화된 볼트 파일(blob)을 저장. 멀티 디바이스 동기화의 중간 저장소.

**이중 암호화:** 클라이언트 XChaCha20-Poly1305 + AWS S3 AES-256

### 4.4 CloudFront + S3 (웹 대시보드)

**뭘 하는가:** 유료 사용자를 위한 웹 대시보드. 시크릿 현황, 감사 로그 조회.

**기술:** Next.js 15 static export -> S3 + CloudFront

**대시보드에서 할 수 없는 것 (보안):** 시크릿 값 조회/복호화 (Master Key가 브라우저에 없으므로)

---

## 5. 보안 요약: 누가 뭘 알고 있는가

### MVP (Phase 1 -- 로컬 전용)

| 구성요소 | 시크릿 평문 | Master Key | 암호화된 데이터 |
|---------|:----------:|:----------:|:------------:|
| **CLI** | O | O (Keychain) | O |
| **SQLite** | X (암호화됨) | X | O |
| **CLAUDE.md** | X (명령어만) | X | X |
| **서버** | 없음 | 없음 | 없음 |

### Phase 2 (Cloud 포함)

| 구성요소 | 시크릿 평문 | Master Key | 암호화된 blob | 메타데이터 |
|---------|:----------:|:----------:|:------------:|:----------:|
| **CLI** | O | O (Keychain) | O | O |
| **ECS API** | X | X | 중계만 | O |
| **RDS** | X | X | X | O |
| **S3** | X | X | O (저장) | X |

**결론: 시크릿 평문은 CLI(사용자 디바이스)에서만 존재한다.**

---

## 6. 비용 구조

### MVP (Phase 1): $0

서버 없음. AWS 비용 없음. Go 바이너리 배포 비용만 (GitHub Releases 무료).

| 항목 | 비용 |
|------|------|
| 서버 인프라 | **$0** |
| GitHub Releases | $0 |
| Homebrew tap | $0 |
| 도메인 | $15/년 (~$1.25/월) |
| **월간 총** | **~$1.25/월** |

### Phase 2 Cloud (수요 검증 후): ECS + RDS 비용

| 항목 | 100 유료 사용자 | 1,000 유료 사용자 |
|------|:--------------:|:----------------:|
| **ECS Fargate** | ~$10/월 | ~$20/월 |
| **NLB** | ~$16/월 | ~$16/월 |
| **RDS PostgreSQL** | ~$15/월 | ~$30/월 |
| **S3 + CloudFront** | ~$2/월 | ~$11/월 |
| **월간 총 비용** | **~$43/월** | **~$77/월** |
| **월간 총 수익** | **$100** | **$1,000** |
| **이익률** | **57%** | **92.3%** |

---

## 7. 구현 순서

### Phase 1: MVP Go CLI (2주) -- 이것만으로 제품

| 주차 | 구현 대상 |
|------|----------|
| 1주 | internal/crypto (Argon2id + XChaCha20 + Recovery Key) + SQLite 볼트 + init/set/get + **CLAUDE.md 자동 생성** |
| 2주 | run/list/delete/import/export/env + export --encrypted + **Fake Door(tene sync)** + Keychain + goreleaser + Homebrew tap |

**이 시점에서 제품 출시 가능.** 서버 없이, 비용 $0로, Claude Code가 자동 인식하는 시크릿 관리가 된다.

### Phase 2: $1 Cloud + Agent 확장 (수요 검증 후, 3-4주)

| 주차 | 구현 대상 |
|------|----------|
| 3주 | AWS 인프라 (ECS Fargate + NLB + RDS PostgreSQL + S3) + GitHub OAuth |
| 4주 | 동기화 엔진 (push/pull) + Stripe 결제 |
| 5주 | 웹 대시보드 (Next.js static) + --cursor/--windsurf 플래그 추가 + 배포 |

### Phase 3: 팀 기능 (가설 -- 시장 검증 후)

- 팀 볼트 공유, RBAC
- 에이전트 스코핑
- MCP 서버
- 가격 미정 (Fake Door Test로 수요 확인)

---

## 8. 모노레포 구조

```
tene/
├── cmd/tene/              # Go CLI main.go (cobra)
├── internal/
│   ├── crypto/            # XChaCha20-Poly1305 + Argon2id (golang.org/x/crypto)
│   ├── vault/             # SQLite 볼트 CRUD (modernc.org/sqlite)
│   ├── keychain/          # OS Keychain (go-keyring)
│   ├── claudemd/          # CLAUDE.md 생성
│   ├── recovery/          # BIP-39 니모닉 (go-bip39)
│   └── cli/               # cobra 명령어 정의
├── apps/web/              # Next.js 랜딩페이지 (유지)
├── docs/                  # 기획 문서
├── go.mod
├── go.sum
├── goreleaser.yml         # 멀티 플랫폼 빌드 + Homebrew tap
├── Makefile
└── install.sh             # curl 설치 스크립트
```

---

## 9. 한 장 요약

```
"시크릿의 Git" + Claude Code 자동 인식

MVP (Phase 1):                          Phase 2 (수요 검증 후):
- Go 단일 바이너리 (런타임 불필요)       - $1/월로 클라우드 백업
- 로컬에서 무료로 동작 ($0)              - 멀티 디바이스 동기화
- 서버 불필요                            - 웹 대시보드
- 오프라인 OK                            - 감사 로그
- Claude Code 자동 인식 (핵심!)          - Cursor/Windsurf 지원
- Fake Door로 수요 검증

설치:
brew install tomo-kay/tap/tene           # macOS
curl -fsSL https://tene.sh/install.sh | sh  # Linux/WSL

만드는 것 (MVP):
[Go CLI]      로컬 시크릿 암호화/관리          <- 핵심. 이것만으로 제품.
[SQLite]      암호화된 시크릿 로컬 저장         <- CLI의 일부.
[CLAUDE.md]   Claude Code 자동 인식            <- 핵심 차별점.
[Fake Door]   tene sync -> waitlist            <- Cloud 수요 검증.
[Encrypted]   tene export --encrypted          <- 수동 백업.
[goreleaser]  멀티 플랫폼 빌드 + Homebrew tap  <- 배포.

Phase 2에서 추가:
[ECS+NLB]     Cloud API (인증/동기화/결제)     <- $1 유료 전용.
[RDS]         사용자/디바이스/감사 로그 저장     <- $1 유료 전용.
[S3]          암호화된 볼트 백업               <- $1 유료 전용.
[Dashboard]   시크릿 현황 + 감사 로그 조회     <- $1 유료 전용.
[--cursor]    Cursor AI Agent 지원            <- Phase 2.
[--windsurf]  Windsurf AI Agent 지원          <- Phase 2.
```
