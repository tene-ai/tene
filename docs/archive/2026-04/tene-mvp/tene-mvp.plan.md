# Tene MVP 기술 Plan v4 -- Go 전환 + Claude Code 전용

> v4.0 (2026-04-06) -- TypeScript/Node.js에서 Go로 전환, Claude Code 전용 MVP, brew/curl 배포, goreleaser
>
> **Summary**: Tene Agentic Secret Runtime Platform의 Local-Only MVP 기술 Plan. Go 단일 바이너리, 서버 비용 $0, Claude Code AI Agent 자동 인식이 핵심 차별점
>
> **Project**: Tene
> **Version**: 4.0.0
> **Author**: CTO Lead (Steve)
> **Date**: 2026-04-06
> **Status**: Draft (v4 -- Go 전환, Claude Code 전용 MVP, --cursor/--windsurf Phase 2)

---

## Executive Summary

| 관점 | 내용 |
|------|------|
| **Problem** | AI 에이전트와 바이브코더의 75%가 시크릿을 .env/하드코딩으로 관리. 기존 도구(Vault, Doppler, Infisical)는 서버 가입을 강제하고, $6-21/유저/월로 비싸며, AI 에이전트를 1등 시민으로 지원하지 않음. 2025년 GitHub 시크릿 노출 2,865만 건(+34%), AI 서비스 시크릿 누출 81% 급증 |
| **Solution** | **Go 단일 바이너리** + **서버 없는** CLI 시크릿 관리 + **Claude Code AI Agent 자동 인식**. Master Password + Argon2id KDF + XChaCha20-Poly1305 + SQLite 로컬 볼트. `brew install tomo-kay/tap/tene` -> `tene init` 하면 CLAUDE.md 자동 생성으로 Claude Code가 즉시 시크릿 관리법을 인식. 서버 비용 $0, CGo 불필요 |
| **Function/UX Effect** | CLI 명령어로 로컬 오프라인 시크릿 관리. `tene init`으로 CLAUDE.md 자동 생성. `tene sync` 실행 시 Cloud waitlist 안내 (Fake Door Test). `tene export --encrypted`로 암호화 백업. Go 단일 바이너리로 설치 즉시 실행 (Node.js 런타임 불필요) |
| **Core Value** | "서버가 없다 = 해킹 대상이 없다." 로컬 암호화 + Claude Code 자동 인식 + 오픈소스 + 서버 비용 $0 + Go 단일 바이너리 = 시크릿 관리의 Git. Claude Code가 tene 사용법을 자동으로 학습하는 유일한 시크릿 매니저 |

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | 바이브코더/AI 에이전트의 시크릿 하드코딩 및 .env 노출 문제 해결. 기존 도구의 서버 강제 + 고가격 + AI 에이전트 미지원 해소. "서버가 없으면 해킹 대상도 없다" |
| **WHO** | 솔로 바이브코더 (Claude Code 사용자). 5-15개 시크릿을 관리하는 개인 개발자. 서버 가입/결제에 거부감이 있는 개발자 |
| **RISK** | 암호화 구현 결함 시 제품 신뢰도 치명적 타격. Master Password 분실 시 복구 불가 (Recovery Key로 완화). ".env로 충분" 인식 극복 필요 |
| **SUCCESS** | CLI 명령어 로컬 동작, 오프라인 100%, XChaCha20-Poly1305 + Argon2id 암호화, 설치->첫 시크릿 1분 이내, Claude Code 자동 인식(CLAUDE.md), Fake Door Test로 Cloud 수요 검증 |
| **SCOPE** | Phase 1 (MVP, 2주): Go CLI + Claude Code 통합 + Fake Door. Phase 2 (수요 검증 후): $1 Cloud + --cursor/--windsurf 확장. Phase 3: 팀 기능 (가설) |

---

## 1. 개요 및 배경

### 1.1 목적

Tene MVP의 Local-Only 기술 아키텍처를 정의한다. **Go 단일 바이너리**로 구현하여, "서버 없이, 가입 없이, 무료로, 서버 비용 $0로, Node.js 런타임 없이" 시크릿을 관리하는 CLI를 코어로 하고, Claude Code AI Agent가 자동으로 시크릿 관리법을 인식하도록 하는 것이 핵심 차별점이다. Cloud 기능은 MVP에서 완전 제외하고, Fake Door Test(tene sync -> waitlist)로 수요를 검증한 후 Phase 2에서 구축한다.

### 1.2 배경

- **바이브코딩 폭발적 성장**: AI 생성 코드의 45%가 보안 취약점 포함, 시크릿 하드코딩이 주요 문제
- **NHI(비인간 ID) 폭증**: 인간 ID 대비 100:1 비율에 도달한 기업 존재
- **기존 도구의 빈틈**: Vault(과잉 + $1,152+/월), Doppler($21/유저/월), Infisical(셀프호스팅 서버 필요)
- **Local-First 트렌드**: 개발자 클라우드 피로감 증가, "내 데이터는 내 디바이스에" 선호 급증
- **경쟁 공백**: 로컬 전용 + AI 에이전트 네이티브 + $0 가격대의 시크릿 관리 도구 부재
- **AI Agent 시대**: Claude Code가 코드 작성의 주류가 됨. 시크릿 관리법을 자동으로 알려주는 도구가 없음

### 1.3 v3 -> v4 핵심 변화

| 항목 | v3 Plan | v4 Plan |
|------|---------|---------|
| **언어/런타임** | TypeScript / Node.js | **Go** (단일 바이너리, 런타임 불필요) |
| **패키지 매니저** | npm (`npm install -g @tene/cli`) | **Homebrew / curl / go install** |
| **CLI 프레임워크** | Commander.js | **cobra** |
| **로컬 DB** | better-sqlite3 (Node.js 바인딩) | **modernc.org/sqlite** (순수 Go, CGo 불필요) |
| **암호화** | libsodium-wrappers | **golang.org/x/crypto** (nacl/secretbox + argon2) |
| **키체인** | keytar | **zalando/go-keyring** |
| **니모닉** | BIP-39 JS 라이브러리 | **tyler-smith/go-bip39** |
| **빌드/배포** | npm publish | **goreleaser** (멀티 플랫폼 바이너리 + Homebrew tap) |
| **테스트** | Vitest | **Go testing + testify** |
| **린트** | Biome | **golangci-lint** |
| **AI Agent 타겟** | Claude Code, Cursor, Windsurf | **Claude Code 전용** (--cursor, --windsurf는 Phase 2) |
| **모노레포** | pnpm workspace | **Go 모듈 단일 모듈** (apps/web은 별도 유지) |

### 1.4 왜 Go인가

| 기준 | TypeScript/Node.js | Go |
|------|-------------------|-----|
| **배포** | npm + Node.js 런타임 필수 | 단일 바이너리, 런타임 불필요 |
| **설치 속도** | npm install -g (수초~수십초) | brew install 또는 curl (수초) |
| **CLI 시작 시간** | ~300ms (Node.js cold start) | ~10ms |
| **바이너리 크기** | 수십 MB (node_modules) | ~15MB (단일 바이너리) |
| **크로스 컴파일** | 복잡 (native addons) | goreleaser로 자동 (macOS/Linux/Windows) |
| **CGo 의존** | better-sqlite3 native binding | modernc.org/sqlite (순수 Go, CGo 불필요) |
| **타겟 사용자** | npm 생태계 바이브코더 | brew/curl로 설치하는 모든 개발자 |

### 1.5 관련 문서

- PRD v2: `docs/00-pm/tene.prd.md`
- Strategy v2: `docs/00-pm/tene-strategy.md`
- Discovery: `docs/00-pm/tene-discovery.md`
- Research: `docs/00-pm/tene-research.md`

---

## 2. 범위

### 2.1 In Scope -- Phase 1: MVP (Go CLI, 2주)

#### 2.1.1 CLI 코어

- [ ] CLI 코어: 10개 명령어 (init, set, get, run, list, delete, import, export, passwd, recover)
- [ ] 환경 전환: `tene env` (dev/staging/prod)
- [ ] Master Password 설정 + Recovery Key 생성 (12단어 니모닉, BIP-39)
- [ ] Master Password 변경: `tene passwd` (볼트 재암호화 + 새 Recovery Key)
- [ ] Master Password 복구: `tene recover` (Recovery Key로 재설정)
- [ ] 로컬 암호화: Argon2id KDF + XChaCha20-Poly1305 + SQLite
- [ ] OS Keychain 연동 (Master Key 저장)
- [ ] 오프라인 100% 동작 (네트워크 통신 없음)
- [ ] .env 마이그레이션 (import/export)
- [ ] 플랫폼: macOS (arm64/amd64), Linux (arm64/amd64), Windows (WSL)
- [ ] 배포: Homebrew tap (`brew install tomo-kay/tap/tene`) + curl 설치 스크립트 + `go install`
- [ ] --json 플래그 (AI 에이전트 파싱용)

#### 2.1.2 AI Agent 자동 인식 (핵심 차별점 -- Claude Code 전용)

- [ ] `tene init` 시 CLAUDE.md 자동 생성 (기본 동작)
- [ ] Claude Code가 tene 명령어를 즉시 인식하도록 컨텍스트 파일 자동 관리

#### 2.1.3 Fake Door Test

- [ ] `tene sync` 명령어 포함 (실제 동기화 미구현)
- [ ] 실행 시 Cloud waitlist 안내 화면 표시
- [ ] waitlist 등록 수 수집 (Analytics)

#### 2.1.4 암호화 백업

- [ ] `tene export --encrypted` 암호화된 볼트 파일 수동 백업

### 2.2 Out of Scope -- Phase 2: $1 Cloud + Agent 확장 (수요 검증 후)

Cloud 기능 및 추가 AI Agent 지원은 MVP에서 완전 제외한다.

- `tene init --cursor`: Cursor AI Agent 지원 (.cursorrules 생성)
- `tene init --windsurf`: Windsurf AI Agent 지원 (.windsurfrules 생성)
- GitHub OAuth 인증 (Cloud 계정)
- 암호화된 볼트 백업: 로컬 SQLite -> 암호화 blob -> AWS S3
- 멀티 디바이스 동기화 프로토콜
- API 서버: ECS Fargate + NLB (Hono 프레임워크)
- Cloud DB: RDS PostgreSQL (사용자 + 메타데이터)
- 웹 대시보드: Next.js static export (S3 + CloudFront)
- 감사 로그 (Cloud 측)
- Stripe 결제 ($1/월 구독)

### 2.3 Out of Scope -- Phase 3: 팀 기능 (가설)

- 팀 볼트 / RBAC
- 에이전트 스코핑 (에이전트별 시크릿 접근 제어)
- 프로젝트 간 글로벌 키 공유
- 자동 시크릿 로테이션
- MCP 서버 (AI 에이전트 네이티브 통합)
- SSO / SCIM
- Docker / 셀프호스팅
- 모바일 앱

---

## 3. Tene가 해결하는 것 / 못 하는 것

### 3.1 해결하는 것 (MVP)

| 문제 | Tene의 해결 방식 |
|------|-----------------|
| **시크릿 안전 저장** | XChaCha20-Poly1305 + SQLite 로컬 볼트. .env 평문 저장 대비 암호화 |
| **시크릿 암호화** | Argon2id KDF -> Master Key -> XChaCha20-Poly1305. 디스크에 평문 없음 |
| **환경변수 주입** | `tene run -- <command>`로 암호화된 시크릿을 환경변수로 자동 주입 |
| **Claude Code 자동 인식** | `tene init`시 CLAUDE.md 자동 생성. Claude Code가 시크릿 관리법을 즉시 학습 |
| **.env 마이그레이션** | `tene import .env` 한 줄로 기존 .env에서 전환 |
| **환경별 시크릿 분리** | `tene env dev/staging/prod`로 환경별 시크릿 관리 |
| **암호화 백업** | `tene export --encrypted`로 수동 백업 가능 |
| **즉시 설치** | Go 단일 바이너리. Node.js 런타임 불필요. brew/curl로 즉시 설치 |

### 3.2 못 하는 것 (MVP 한계)

| 한계 | 설명 | 해결 시기 |
|------|------|----------|
| **시크릿 만료 확인** | 시크릿에 만료일을 설정하거나 만료 알림을 받을 수 없음 | 미정 |
| **자동 갱신** | 시크릿 자동 로테이션/갱신 기능 없음 | Phase 3+ |
| **프로젝트 간 글로벌 키 공유** | 같은 키를 여러 프로젝트에서 공유할 수 없음 (프로젝트별 독립) | Phase 3 (팀 기능) |
| **Cloud 동기화** | 멀티 디바이스 동기화 미지원 (로컬 전용) | Phase 2 (수요 검증 후) |
| **웹 대시보드** | 브라우저에서 시크릿 현황 조회 불가 | Phase 2 |
| **팀 협업** | 팀원 간 시크릿 공유/RBAC 불가 | Phase 3 |
| **Cursor/Windsurf 지원** | --cursor, --windsurf 플래그 미지원 | Phase 2 |

---

## 4. AI Agent 자동 인식 -- 핵심 차별점 (Claude Code 전용)

### 4.1 개요

Tene의 핵심 차별점은 Claude Code AI Agent가 시크릿 관리법을 **자동으로 인식**하는 것이다. `tene init` 실행 시 CLAUDE.md를 자동 생성하여, Claude Code가 별도 학습 없이 즉시 tene을 통해 시크릿을 관리한다.

MVP에서는 Claude Code만 지원한다. Cursor(`.cursorrules`), Windsurf(`.windsurfrules`) 지원은 Phase 2에서 `--cursor`, `--windsurf` 플래그로 추가한다.

### 4.2 자동 생성되는 CLAUDE.md 내용 (영어)

```markdown
# Secrets Management

This project uses [tene](https://github.com/tene-ai/tene) for secret management.

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

### 4.3 기존 파일 병합 정책

| 상황 | 동작 |
|------|------|
| CLAUDE.md가 없을 때 | 새로 생성 |
| CLAUDE.md가 이미 있을 때 | "# Secrets Management" 섹션만 추가/업데이트 |
| tene 섹션이 이미 있을 때 | 스킵 (중복 방지) |

### 4.4 구현 설계

```go
// internal/claudemd/generator.go

type Generator struct {
    projectDir string
}

func (g *Generator) Generate() error {
    path := filepath.Join(g.projectDir, "CLAUDE.md")

    // 기존 CLAUDE.md 확인
    if exists(path) {
        return g.appendSection(path)
    }
    return g.createNew(path)
}

func (g *Generator) appendSection(path string) error {
    content, _ := os.ReadFile(path)
    if strings.Contains(string(content), "# Secrets Management") {
        return nil // 이미 존재, 스킵
    }
    // 섹션 추가
    return appendToFile(path, claudeMdTemplate)
}
```

### 4.5 AI Agent 사용 시나리오

```bash
# 1. 프로젝트 초기화 (CLAUDE.md 자동 생성)
$ tene init
  Master Password: ********
  Vault created (.tene/vault.db)
  CLAUDE.md generated (Claude Code will auto-detect tene)

# 2. 시크릿 저장
$ tene set STRIPE_KEY sk_test_xxxxx

# 3. Claude Code가 코드 작성 시 자동으로 tene 사용:
#    (CLAUDE.md를 읽고 tene 명령어를 인식)
#    "이 프로젝트는 tene으로 시크릿을 관리하므로,
#     STRIPE_KEY=$(tene get STRIPE_KEY) 를 사용합니다."
```

---

## 5. Fake Door Test -- Cloud 수요 검증

### 5.1 개요

Cloud 동기화 기능(`tene sync`)을 MVP에 명령어로 포함하되, 실제 동기화는 구현하지 않는다. 사용자가 `tene sync`를 실행하면 Cloud waitlist 안내 화면을 표시하고, waitlist 등록 수를 수집하여 Cloud 수요를 검증한다.

### 5.2 tene sync 실행 화면

```
$ tene sync

  Tene Cloud Sync -- Coming Soon!

  Cloud sync will enable:
  - Multi-device secret synchronization
  - Encrypted cloud backup (zero-knowledge)
  - Web dashboard for secret overview
  - All for just $1/month

  Join the waitlist to get early access:
  -> https://tene.sh/waitlist

  In the meantime, use `tene export --encrypted` for local backup.

  [Open waitlist page? (Y/n)]
```

### 5.3 수요 검증 기준

| 지표 | 목표 | 행동 |
|------|------|------|
| `tene sync` 실행 횟수 / DAU | 15%+ | Cloud 구축 시작 (Phase 2) |
| waitlist 등록 수 | 100명+ | Cloud 구축 시작 (Phase 2) |
| `tene sync` 실행 횟수 / DAU | < 5% | Cloud 보류, 로컬 기능 강화 |

### 5.4 수동 백업 대안: tene export --encrypted

Cloud 동기화가 없는 MVP 기간에는 `tene export --encrypted`로 암호화된 볼트 파일을 수동 백업할 수 있다.

```
$ tene export --encrypted
  Encrypted vault exported to: ./my-project.tene.enc

  This file is encrypted with your Master Password.
  To restore: tene import --encrypted my-project.tene.enc

  Store this file in a safe place (USB, cloud drive, etc.)
```

---

## 6. 시스템 아키텍처

### 6.1 전체 아키텍처 -- Phase 1 MVP (로컬 전용)

```
[MVP -- Phase 1] 사용자 디바이스 (서버 없음, 오프라인 동작, 비용 $0)
+-------------------------------------------------------+
|                                                         |
|  tene CLI (Go 단일 바이너리)                             |
|       |                                                 |
|       +---- internal/crypto                             |
|       |     Argon2id KDF + XChaCha20-Poly1305            |
|       |     (golang.org/x/crypto)                        |
|       |                                                 |
|       +---- internal/vault                              |
|       |     SQLite 로컬 볼트 (.tene/vault.db)            |
|       |     (modernc.org/sqlite, CGo 불필요)              |
|       |                                                 |
|       +---- internal/keychain                           |
|       |     OS Keychain (go-keyring)                     |
|       |                                                 |
|       +---- internal/claudemd                           |
|       |     CLAUDE.md 자동 생성 (Claude Code 전용)        |
|       |                                                 |
|       +---- internal/recovery                           |
|             BIP-39 니모닉 (go-bip39)                     |
|                                                         |
|  tene sync -> Fake Door (waitlist 안내)                  |
|  tene export --encrypted -> 수동 암호화 백업              |
|                                                         |
|  ** 네트워크 통신: 없음 **                               |
|  ** 서버: 없음 **                                       |
|  ** 해킹 대상: 없음 **                                   |
|  ** 비용: $0 **                                         |
+-------------------------------------------------------+
```

### 6.2 핵심 아키텍처 원칙

| 원칙 | 설명 |
|------|------|
| **Local-Only (MVP)** | MVP는 100% 로컬. Cloud는 Phase 2 (수요 검증 후) |
| **Server Cost $0** | MVP 기간 서버 인프라 비용 완전 제로 |
| **Claude Code First** | Claude Code 자동 인식이 핵심 차별점. CLAUDE.md 자동 생성 |
| **Zero-Knowledge (설계)** | Phase 2 Cloud 구축 시에도 서버는 시크릿 평문을 절대 알 수 없는 구조 |
| **Offline 100%** | 모든 기능이 인터넷 없이 완전히 동작 |
| **Encrypted at Rest** | 모든 시크릿은 XChaCha20-Poly1305로 암호화 저장 |
| **Go 단일 바이너리** | 런타임 의존 없음. CGo 불필요. 크로스 컴파일 자동화 |

---

## 7. 모노레포 디렉토리 구조

```
tene/
├── cmd/tene/                              # Go CLI 엔트리포인트
│   └── main.go                            # cobra 루트 명령어
│
├── internal/                              # Go 내부 패키지 (비공개)
│   ├── crypto/                            # 암호화 코어
│   │   ├── kdf.go                         # Argon2id 키 유도
│   │   ├── encrypt.go                     # XChaCha20-Poly1305 암호화
│   │   ├── decrypt.go                     # XChaCha20-Poly1305 복호화
│   │   ├── keymanager.go                  # 마스터키 / 파생키 관리
│   │   └── crypto_test.go                 # 암호화 테스트 (95%+ 커버리지)
│   │
│   ├── vault/                             # SQLite 볼트 매니저
│   │   ├── vault.go                       # 볼트 CRUD
│   │   ├── schema.go                      # 테이블 생성 DDL
│   │   ├── migration.go                   # 스키마 마이그레이션
│   │   └── vault_test.go
│   │
│   ├── keychain/                          # OS Keychain 연동
│   │   ├── keychain.go                    # go-keyring 래퍼
│   │   ├── fallback.go                    # 파일 폴백 (Keychain 불가 시)
│   │   └── keychain_test.go
│   │
│   ├── claudemd/                          # CLAUDE.md 생성
│   │   ├── generator.go                   # CLAUDE.md 템플릿 + 생성/병합
│   │   └── generator_test.go
│   │
│   ├── recovery/                          # BIP-39 니모닉 Recovery Key
│   │   ├── mnemonic.go                    # 니모닉 생성/검증/키 유도
│   │   └── mnemonic_test.go
│   │
│   └── cli/                               # cobra 명령어 정의
│       ├── root.go                        # 루트 + 글로벌 플래그
│       ├── init.go                        # tene init
│       ├── set.go                         # tene set KEY VALUE
│       ├── get.go                         # tene get KEY
│       ├── run.go                         # tene run -- CMD
│       ├── list.go                        # tene list
│       ├── delete.go                      # tene delete KEY
│       ├── import.go                      # tene import .env / --encrypted
│       ├── export.go                      # tene export / --encrypted
│       ├── env.go                         # tene env [name]
│       ├── passwd.go                      # tene passwd
│       ├── recover.go                     # tene recover
│       ├── sync.go                        # tene sync (Fake Door)
│       └── whoami.go                      # tene whoami
│
├── apps/web/                              # Next.js 랜딩페이지 (유지)
│   ├── src/
│   └── package.json
│
├── docs/                                  # PDCA 문서
│   ├── 00-pm/
│   ├── 01-plan/
│   │   └── features/
│   ├── 02-design/
│   ├── 03-analysis/
│   └── 04-report/
│
├── .github/
│   └── workflows/
│       ├── ci.yml                         # CI: golangci-lint, go test, go build
│       └── release.yml                    # goreleaser 배포
│
├── go.mod                                 # Go 모듈
├── go.sum
├── goreleaser.yml                         # 멀티 플랫폼 빌드 + Homebrew tap
├── Makefile                               # 개발 편의 명령어
├── install.sh                             # curl 설치 스크립트
└── CLAUDE.md
```

### 7.1 Go 모듈 의존성

```
tene (go.mod)
  ├── github.com/spf13/cobra          # CLI 프레임워크
  ├── modernc.org/sqlite               # 순수 Go SQLite (CGo 불필요)
  ├── golang.org/x/crypto/nacl/secretbox  # XChaCha20-Poly1305
  ├── golang.org/x/crypto/argon2       # Argon2id KDF
  ├── github.com/zalando/go-keyring    # OS Keychain
  ├── github.com/tyler-smith/go-bip39  # BIP-39 니모닉
  └── github.com/stretchr/testify      # 테스트 어서션

핵심 원칙:
- internal/ 패키지: Go 접근 제한으로 외부 직접 사용 불가
- crypto: CLI에서 사용하는 암호화 코어 (향후 Cloud에서도 동일 로직 재사용 가능)
- vault: SQLite CRUD (modernc.org/sqlite로 CGo 없이 크로스 컴파일)
- apps/web: Next.js 랜딩페이지 (Go 모듈과 독립, 별도 npm 프로젝트)
```

---

## 8. 기술 스택 선정 및 근거

### 8.1 기술 스택 총괄 -- Phase 1 MVP

| 영역 | 기술 | 선정 근거 |
|------|------|----------|
| **언어** | Go 1.22+ | 단일 바이너리, 크로스 컴파일, 빠른 CLI 시작 (~10ms), CGo 불필요 |
| **CLI 프레임워크** | cobra (spf13/cobra) | Go CLI 표준, 자동 완성, 서브커맨드, 풍부한 생태계 |
| **로컬 DB** | modernc.org/sqlite | 순수 Go SQLite, CGo 불필요, 크로스 컴파일 호환 |
| **대칭 암호화** | golang.org/x/crypto/nacl/secretbox | XChaCha20-Poly1305, Go 표준 확장 라이브러리, 검증됨 |
| **KDF** | golang.org/x/crypto/argon2 | Argon2id, Go 표준 확장, OWASP 권장 |
| **키체인** | zalando/go-keyring | OS 네이티브 키체인 (macOS Keychain, Linux Secret Service, Win Credential Vault) |
| **니모닉** | tyler-smith/go-bip39 | BIP-39 표준, Go 네이티브 |
| **빌드/배포** | goreleaser | 멀티 플랫폼 바이너리 + Homebrew tap + GitHub Releases |
| **테스트** | Go testing + testify | Go 내장 테스트 + 어서션 라이브러리 |
| **린트** | golangci-lint | Go 표준 린터 집합 |
| **CI/CD** | GitHub Actions | goreleaser 자동 릴리즈 |

### 8.2 핵심 기술 선정 상세 근거

#### 로컬 DB: modernc.org/sqlite (vs. mattn/go-sqlite3)

| 기준 | modernc.org/sqlite | mattn/go-sqlite3 |
|------|-------------------|------------------|
| CGo 의존 | 없음 (순수 Go) | 필요 (C 컴파일러) |
| 크로스 컴파일 | goreleaser로 자동 | CGo cross-compile 복잡 |
| 빌드 속도 | 빠름 | 느림 (C 컴파일) |
| 성능 | 약간 느림 (~10-20%) | 최적 (네이티브 C) |
| 결론 | **선택** -- CGo 불필요가 결정적 | CLI용으로 과잉 |

#### 암호화: golang.org/x/crypto (vs. 외부 libsodium 바인딩)

| 기준 | golang.org/x/crypto | go-libsodium 바인딩 |
|------|--------------------|--------------------|
| CGo 의존 | 없음 | 필요 |
| 유지보수 | Go 팀 공식 | 서드파티 |
| XChaCha20-Poly1305 | nacl/secretbox | 지원 |
| Argon2id | argon2 패키지 | 지원 |
| 결론 | **선택** -- Go 공식 + CGo 불필요 | CGo 의존 |

---

## 9. 보안 아키텍처

### 9.1 보안 모델 (로컬 전용)

| 층위 | 전략 | 적용 |
|------|------|------|
| **Layer 1: Server-Free** | 서버가 없다 (MVP) | 해킹 대상 자체가 없음. 공격 표면 = 0 |
| **Layer 2: Encrypted at Rest** | 모든 시크릿이 암호화 저장 | 디바이스 분실/도난 시에도 시크릿 안전 |
| **Layer 3: OS Keychain** | Master Key를 OS 보안 저장소에 보관 | 프로세스 간 격리, 하드웨어 암호화 |

### 9.2 보안 플로우

```
[사용자 디바이스 -- 오직 여기에만 존재]
+----------------------------------------------+
|                                                |
|  Master Password (사용자만 알고 있음)           |
|       |                                        |
|       v                                        |
|  Argon2id KDF                                  |
|  (memory: 64MB, iterations: 3,                 |
|   parallelism: 1, outputLen: 32)               |
|       |                                        |
|       v                                        |
|  Master Key (256-bit)                          |
|       |                                        |
|       +---> OS Keychain 저장                    |
|       |     (macOS Keychain /                  |
|       |      Linux Secret Service)             |
|       |                                        |
|       v                                        |
|  XChaCha20-Poly1305 Encrypt                    |
|  (192-bit random nonce, AAD=key name)          |
|       |                                        |
|       v                                        |
|  SQLite DB (.tene/vault.db)                    |
|  encrypted secrets stored locally              |
|                                                |
|  ** 네트워크 통신: 없음 **                      |
|  ** 서버: 없음 **                              |
|  ** 해킹 대상: 없음 **                         |
+----------------------------------------------+
```

### 9.3 보안 원칙 요약

| 원칙 | MVP (Phase 1) |
|------|:-------------:|
| 시크릿 평문은 로컬에만 존재 | O |
| 모든 시크릿은 암호화 저장 | O (SQLite) |
| Master Key는 OS Keychain에 저장 | O |
| 오프라인 100% 동작 | O |
| 네트워크 통신 없음 | O |
| 오픈소스 검증 가능 | O |

---

## 10. 키 유도 및 암호화/복호화 플로우

### 10.1 키 유도 (Key Derivation)

```go
// internal/crypto/kdf.go

// 1단계: Master Password에서 Master Key 유도
salt := make([]byte, 16) // 128-bit random salt
io.ReadFull(rand.Reader, salt)

masterKey := argon2.IDKey(
    []byte(password),
    salt,
    3,              // iterations
    64*1024,        // 64MB memory
    1,              // parallelism
    32,             // 256-bit output
)

// 2단계: Master Key에서 용도별 키 파생 (HKDF)
encKey := hkdf.Expand(sha256.New, masterKey, []byte("tene-encryption-key"), 32)
authHash := hkdf.Expand(sha256.New, masterKey, []byte("tene-auth-hash"), 32)
```

### 10.2 Recovery Key 생성 (12단어 니모닉, BIP-39)

```go
// internal/recovery/mnemonic.go

// tene init 시 Recovery Key 생성
// 128-bit entropy -> BIP-39 워드리스트 기반 12단어 니모닉
entropy, _ := bip39.NewEntropy(128)
mnemonic, _ := bip39.NewMnemonic(entropy)
// 예: "apple banana cherry dolphin eagle frost grape harbor island jungle kite lemon"

// 니모닉에서 Recovery Key 유도
recoveryKey := argon2.IDKey([]byte(mnemonic), []byte("tene-recovery"), 3, 64*1024, 1, 32)

// Recovery Key로 Master Key 암호화하여 볼트에 저장
var nonce [24]byte
io.ReadFull(rand.Reader, nonce[:])

var recoveryEncKey [32]byte
copy(recoveryEncKey[:], deriveKey(recoveryKey, "tene-recovery", 32))

encryptedMasterKey := secretbox.Seal(nonce[:], masterKey, &nonce, &recoveryEncKey)
// vault_meta 테이블에 recovery_blob (base64)으로 저장
```

### 10.3 시크릿 암호화 플로우 (tene set KEY VALUE)

```
[CLI] tene set STRIPE_KEY sk_test_xxxxx

1. OS Keychain에서 Master Key 로드
   (없으면 Master Password 입력 요청 -> Argon2id -> Master Key)

2. Master Key에서 Encryption Key 파생
   encKey = deriveKey(masterKey, "tene-encryption-key", 32)

3. 시크릿 값 암호화:
   nonce = random_192bit()
   encrypted = XChaCha20-Poly1305.Seal(
     key: encKey,
     plaintext: "sk_test_xxxxx",
     nonce: nonce,
     additionalData: "STRIPE_KEY"      // 키 이름을 AAD로 (변조 방지)
   )

4. SQLite 볼트에 저장:
   INSERT INTO secrets (name, encrypted_value, environment, version)
   VALUES ('STRIPE_KEY', base64(nonce + encrypted), 'default', 1)

5. 감사 로그 기록 (로컬):
   INSERT INTO audit_log (action, resource_name, timestamp)
   VALUES ('secret.write', 'STRIPE_KEY', NOW())
```

### 10.4 시크릿 복호화 플로우 (tene get KEY)

```
[CLI] tene get STRIPE_KEY

1. SQLite 볼트에서 암호화된 blob 조회:
   SELECT encrypted_value FROM secrets
   WHERE name='STRIPE_KEY' AND environment='default'

2. OS Keychain에서 Master Key 로드

3. Encryption Key 파생:
   encKey = deriveKey(masterKey, "tene-encryption-key", 32)

4. 복호화:
   blob = base64Decode(encrypted_value)
   nonce = blob[0:24]                    // 앞 24바이트
   ciphertext = blob[24:]
   plaintext = XChaCha20-Poly1305.Open(
     key: encKey,
     ciphertext: ciphertext,
     nonce: nonce,
     additionalData: "STRIPE_KEY"        // AAD 검증
   )

5. stdout 출력: "sk_test_xxxxx"
   (Claude Code가 Bash에서 파싱 가능)
```

### 10.5 시크릿 주입 플로우 (tene run -- COMMAND)

```
[CLI] tene run -- claude

1. 현재 환경의 모든 시크릿을 SQLite에서 조회
2. 각 시크릿을 로컬에서 복호화
3. 환경변수로 설정한 자식 프로세스 생성:
   cmd := exec.Command("claude")
   cmd.Env = append(os.Environ(),
     "STRIPE_KEY=sk_test_xxxxx",
     "DATABASE_URL=postgresql://...",
   )
   cmd.Stdin = os.Stdin
   cmd.Stdout = os.Stdout
   cmd.Stderr = os.Stderr
   cmd.Run()
4. 자식 프로세스 종료 시 환경변수 자동 정리
   (디스크에 시크릿이 평문으로 저장되지 않음)
```

### 10.6 암호화 백업 플로우 (tene export --encrypted)

```
[CLI] tene export --encrypted

1. SQLite 볼트 전체를 읽기 (이미 암호화된 시크릿 포함)
2. 볼트 전체를 XChaCha20-Poly1305로 2차 암호화:
   - Master Key로 전체 볼트 데이터를 암호화
   - 결과물에 KDF 파라미터, salt 포함
3. 단일 .tene.enc 파일로 저장:
   - 파일 헤더: magic bytes + version + KDF params
   - 페이로드: 암호화된 볼트 데이터

[CLI] tene import --encrypted my-project.tene.enc

1. 파일 헤더에서 KDF 파라미터 추출
2. Master Password 입력 요청
3. Argon2id로 Master Key 유도
4. 2차 암호화 복호화 -> 볼트 데이터 복원
5. 로컬 SQLite 볼트에 머지
```

---

## 11. 데이터 모델

### 11.1 Local SQLite 스키마 (.tene/vault.db)

```sql
-- vault_meta: 볼트 메타데이터
CREATE TABLE vault_meta (
  key             TEXT PRIMARY KEY,
  value           TEXT NOT NULL
);
-- 저장 항목: vault_version, created_at, kdf_salt (base64),
--           kdf_params (JSON), recovery_blob (base64)

-- environments: 환경 관리
CREATE TABLE environments (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  name            TEXT NOT NULL UNIQUE,              -- default, dev, staging, prod
  is_default      INTEGER NOT NULL DEFAULT 0,
  created_at      TEXT NOT NULL DEFAULT (datetime('now'))
);

-- secrets: 암호화된 시크릿
CREATE TABLE secrets (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  environment_id  INTEGER NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
  name            TEXT NOT NULL,                     -- 시크릿 키 이름 (평문)
  encrypted_value TEXT NOT NULL,                     -- nonce + 암호문 (base64)
  version         INTEGER NOT NULL DEFAULT 1,
  created_at      TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at      TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(environment_id, name)
);

-- audit_log: 로컬 감사 로그
CREATE TABLE audit_log (
  id              INTEGER PRIMARY KEY AUTOINCREMENT,
  action          TEXT NOT NULL,                     -- secret.read, secret.write, etc.
  resource_name   TEXT,                              -- 시크릿 이름
  environment     TEXT,                              -- 환경 이름
  source          TEXT NOT NULL DEFAULT 'cli',       -- cli | agent | export
  timestamp       TEXT NOT NULL DEFAULT (datetime('now'))
);

-- 인덱스
CREATE INDEX idx_secrets_env_name ON secrets(environment_id, name);
CREATE INDEX idx_audit_timestamp ON audit_log(timestamp DESC);
```

### 11.2 로컬 저장소 구조

```
~/.tene/                              # 글로벌 CLI 설정 디렉토리
└── config.json                       # CLI 전역 설정 (파일 퍼미션 0600)
    {
      "defaultEnvironment": "default",
      "analytics": {
        "syncAttempts": 0,            // Fake Door 측정용
        "lastSyncAttempt": null
      }
    }

project/.tene/                        # 프로젝트 로컬 볼트 (tene init으로 생성)
├── vault.db                          # SQLite 암호화 볼트
├── vault.json                        # 볼트 메타데이터
│   {
│     "projectName": "my-project",
│     "createdAt": "2026-04-06T12:00:00Z",
│     "vaultVersion": 1,
│     "agents": ["claude"]
│   }
└── .gitignore                        # .tene/ 전체를 Git에서 제외

project/CLAUDE.md                     # Claude Code 컨텍스트 (tene init이 생성)

OS Keychain:
  - Service: "tene"
  - Account: "{project-path}" 또는 "global"
  - Password: Master Key (32 bytes, base64 encoded)
```

---

## 12. CLI 명령어 설계

### 12.1 Phase 1 MVP 명령어

| 명령어 | 우선순위 | 설명 | 인자 |
|--------|:--------:|------|------|
| `tene init` | P0 | 프로젝트 초기화 (볼트 + Master PW + CLAUDE.md) | `[name]` |
| `tene set <key> <value>` | P0 | 시크릿 저장 (로컬 암호화) | `--env <name>` `--stdin` |
| `tene get <key>` | P0 | 시크릿 조회 (stdout) | `--env <name>` |
| `tene run -- <command>` | P0 | 시크릿 주입 후 명령 실행 | `--env <name>` |
| `tene list` | P0 | 시크릿 목록 (값 마스킹) | `--env <name>` |
| `tene delete <key>` | P0 | 시크릿 삭제 | `--env <name>` `--force` |
| `tene import <file>` | P1 | .env에서 일괄 가져오기 | `--env <name>` `--overwrite` `--encrypted` |
| `tene export` | P1 | .env 형식/암호화 백업 내보내기 | `--env <name>` `--file <path>` `--encrypted` |
| `tene env [name]` | P1 | 환경 전환/목록/생성 | `list`, `create <name>` |
| `tene passwd` | P0 | Master Password 변경 + 볼트 재암호화 + 새 Recovery Key 발급 | -- (대화형) |
| `tene recover` | P0 | Recovery Key로 Master Password 재설정 | -- (대화형) |
| `tene sync` | P1 | Fake Door: waitlist 안내 표시 | -- |
| `tene whoami` | P2 | 현재 상태 표시 | -- |

### 12.2 명령어 상세 동작

#### tene init (핵심 -- 서버 없음, CLAUDE.md 자동 생성)

```
$ cd my-project
$ tene init

  Welcome to Tene! Let's set up your local secret vault.

  Project name (my-project):

  Set your Master Password (used to encrypt all secrets):
  Master Password: ********
  Confirm: ********

  Generating encryption keys...

  Recovery Key (write this down and keep it safe!):
  +--------------------------------------------------+
  |   apple banana cherry dolphin eagle frost          |
  |   grape harbor island jungle kite lemon            |
  |                                                    |
  |   If you forget your Master Password,              |
  |   this is the ONLY way to recover.                 |
  +--------------------------------------------------+

  Created .tene/vault.db (local encrypted vault)
  Added .tene/ to .gitignore
  Master Key saved to OS Keychain
  Generated CLAUDE.md (Claude Code will auto-detect tene)

  Project "my-project" initialized.
  Default environment "default" created.

  Next: tene set KEY VALUE to add your first secret.

  Tip: No server needed. Your secrets stay on this device.
       Claude Code will automatically use tene.
```

#### tene set (로컬 암호화 저장)

```
$ tene set STRIPE_KEY sk_test_xxxxx
  STRIPE_KEY saved (encrypted, default)

# stdin 지원 (shell history 방지)
$ echo "sk_test_xxxxx" | tene set STRIPE_KEY --stdin
  STRIPE_KEY saved (encrypted, default)

# 환경 지정
$ tene set DATABASE_URL postgresql://user:pass@host/db --env prod
  DATABASE_URL saved (encrypted, prod)
```

#### tene get (stdout 출력 -- Claude Code 호출용)

```
$ tene get STRIPE_KEY
sk_test_xxxxx

# Claude Code 활용:
$ STRIPE_KEY=$(tene get STRIPE_KEY)

# JSON 출력:
$ tene get STRIPE_KEY --json
{"name":"STRIPE_KEY","value":"sk_test_xxxxx","environment":"default"}
```

#### tene run (시크릿 주입 실행)

```
$ tene run -- claude
  Injecting 5 secrets into environment...
  Starting: claude

$ tene run --env prod -- node server.js
  Injecting 8 secrets (prod) into environment...
  Starting: node server.js
```

#### tene list (목록 + 마스킹)

```
$ tene list
  Project: my-project (default)

  NAME              VALUE           UPDATED
  STRIPE_KEY        sk_te*****      2 minutes ago
  DATABASE_URL      postg*****      5 minutes ago
  API_SECRET        eyJhb*****      1 hour ago

  3 secrets in "default" environment

# JSON 출력:
$ tene list --json
[{"name":"STRIPE_KEY","preview":"sk_te*****","updatedAt":"..."}]
```

#### tene import / export

```
$ tene import .env
  Found 5 secrets in .env:
    STRIPE_KEY, DATABASE_URL, API_SECRET, SENDGRID_KEY, JWT_SECRET

  Import 5 secrets to "my-project" (default)? (y/N) y
  5 secrets imported (encrypted).

  Tip: You can now delete .env and use tene run instead.

$ tene export
  STRIPE_KEY=sk_test_xxxxx
  DATABASE_URL=postgresql://user:pass@host/db

$ tene export --file .env.local
  5 secrets exported to .env.local
  Warning: This file contains plain-text secrets. Do not commit it.

$ tene export --encrypted
  Encrypted vault exported to: ./my-project.tene.enc

  This file is encrypted with your Master Password.
  To restore: tene import --encrypted my-project.tene.enc

$ tene import --encrypted my-project.tene.enc
  Enter Master Password: ********
  5 secrets restored to "my-project" vault.
```

#### tene sync (Fake Door Test)

```
$ tene sync

  Tene Cloud Sync -- Coming Soon!

  Cloud sync will enable:
  - Multi-device secret synchronization
  - Encrypted cloud backup (zero-knowledge)
  - Web dashboard for secret overview
  - All for just $1/month

  Join the waitlist to get early access:
  -> https://tene.sh/waitlist

  In the meantime, use `tene export --encrypted` for local backup.

  [Open waitlist page? (Y/n)]
```

### 12.3 CLI 에러 처리

| 상황 | 에러 메시지 | 종료 코드 |
|------|-----------|:---------:|
| 볼트 미초기화 | `Not in a Tene project. Run "tene init" first.` | 1 |
| 시크릿 없음 | `Secret "KEY" not found in "env-name" environment.` | 1 |
| Master PW 오류 | `Invalid Master Password. Try again or use Recovery Key.` | 2 |
| Recovery Key 오류 | `Invalid Recovery Key.` | 2 |
| 이미 존재 | `Secret "KEY" already exists. Use --overwrite to replace.` | 1 |
| Keychain 실패 | `Cannot access OS Keychain. Enter Master Password manually.` | 0 (폴백) |

### 12.4 글로벌 플래그

| 플래그 | 설명 |
|--------|------|
| `--version, -v` | 버전 출력 |
| `--help, -h` | 도움말 |
| `--json` | JSON 형식 출력 (AI 에이전트 파싱용) |
| `--quiet, -q` | 최소 출력 (에러만) |
| `--env <name>` | 환경 지정 (기본: 현재 환경) |
| `--no-color` | 색상 출력 비활성화 |

---

## 13. 빌드 및 배포

### 13.1 goreleaser 설정

```yaml
# goreleaser.yml
builds:
  - id: tene
    main: ./cmd/tene
    binary: tene
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

brews:
  - name: tene
    repository:
      owner: tomo-kay
      name: homebrew-tap
    homepage: "https://tene.sh"
    description: "Secret management that AI agents understand"
    install: |
      bin.install "tene"

archives:
  - format: tar.gz
    name_template: "tene_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

release:
  github:
    owner: tomo-kay
    name: tene
```

### 13.2 설치 방법

```bash
# macOS (Homebrew)
brew install tomo-kay/tap/tene

# Linux / Windows WSL (curl 스크립트)
curl -fsSL https://tene.sh/install.sh | sh

# Go 사용자
go install github.com/tene-ai/tene@latest

# GitHub Releases에서 직접 다운로드
# https://github.com/tene-ai/tene/releases
```

### 13.3 CI/CD 파이프라인

```
Push to main
    |
    v
GitHub Actions (CI)
    - golangci-lint
    - go test ./...
    - go build ./cmd/tene
    |
    v
Tag push (v*)
    |
    v
goreleaser release
    - 멀티 플랫폼 바이너리 빌드 (macOS/Linux/Windows x amd64/arm64)
    - GitHub Releases 업로드
    - Homebrew tap 업데이트 (tomo-kay/homebrew-tap)
    - install.sh 업데이트
```

---

## 14. Phase 정리

### 14.1 Phase 1: MVP (2주) -- Go CLI + Claude Code 통합 + Fake Door

| 구성 | 내용 |
|------|------|
| **핵심** | Go CLI + SQLite + crypto + Claude Code 자동 인식 + Fake Door |
| **비용** | $0 (서버 없음) |
| **기간** | 2주 |
| **목표** | 시크릿 관리 + Claude Code 자동 인식. Cloud 수요 검증 |
| **배포** | Homebrew tap + curl 스크립트 + go install |

### 14.2 Phase 2: $1 Cloud + Agent 확장 (수요 검증 후)

**진입 조건**: Fake Door Test에서 `tene sync` 실행률 15%+ 또는 waitlist 100명+

| 구성 | 내용 |
|------|------|
| **인프라** | ECS Fargate + NLB + RDS PostgreSQL + S3 |
| **기능** | Cloud 동기화, 웹 대시보드, 감사 로그, Stripe 결제 |
| **Agent 확장** | `--cursor` (Cursor), `--windsurf` (Windsurf) 플래그 추가 |
| **비용** | 사용자에게 $1/월. 인프라 비용은 ECS+RDS 기준 |
| **기간** | 3-4주 (수요 확인 후) |

**Phase 2 비용 추정 (ECS + RDS)**

| 항목 | 100 유료 사용자 | 1,000 유료 사용자 |
|------|:--------------:|:----------------:|
| **ECS Fargate** (0.25 vCPU + 0.5GB) | ~$10/월 | ~$20/월 |
| **NLB** | ~$16/월 | ~$16/월 |
| **RDS PostgreSQL** (db.t4g.micro) | ~$15/월 | ~$30/월 |
| **S3 (볼트 저장)** | ~$0.50 | ~$5 |
| **CloudFront** | ~$1 | ~$5 |
| **Route 53** | ~$0.50 | ~$0.50 |
| **월간 총 비용** | **~$43** | **~$77** |
| **월간 총 수익** | **$100** | **$1,000** |
| **이익률** | **57%** | **92.3%** |

### 14.3 Phase 3: 팀 기능 (가설)

| 구성 | 내용 |
|------|------|
| **기능** | 팀 볼트 공유, RBAC, 에이전트 스코핑, MCP 서버 |
| **진입 조건** | Phase 2에서 유료 사용자 확보 + 팀 기능 수요 확인 |
| **가격** | 미정 (Fake Door Test로 수요 확인) |

---

## 15. 요구사항

### 15.1 기능 요구사항

| ID | 요구사항 | 우선순위 | Phase | 상태 |
|----|---------|:--------:|:-----:|:----:|
| FR-01 | `tene init`: Master Password + 볼트 + Recovery Key + CLAUDE.md 생성 | P0 | 1 | Pending |
| FR-02 | `tene set KEY VALUE`: XChaCha20-Poly1305 로컬 암호화 저장 | P0 | 1 | Pending |
| FR-03 | `tene get KEY`: 복호화 후 stdout 출력 (Claude Code Bash 호출) | P0 | 1 | Pending |
| FR-04 | `tene run -- CMD`: 모든 시크릿을 환경변수로 주입 후 명령 실행 | P0 | 1 | Pending |
| FR-05 | `tene list`: 시크릿 목록 표시 (값 마스킹) | P0 | 1 | Pending |
| FR-06 | `tene delete KEY`: 시크릿 삭제 (확인 프롬프트) | P0 | 1 | Pending |
| FR-07 | `tene import .env`: .env 파일에서 시크릿 일괄 가져오기 | P1 | 1 | Pending |
| FR-08 | `tene export`: .env 형식으로 내보내기 | P1 | 1 | Pending |
| FR-09 | `tene export --encrypted`: 암호화 볼트 백업 | P1 | 1 | Pending |
| FR-10 | `tene import --encrypted`: 암호화 볼트 복원 | P1 | 1 | Pending |
| FR-11 | `tene env [name]`: 환경 전환/목록/생성 | P1 | 1 | Pending |
| FR-12 | Master Password + Argon2id KDF + XChaCha20-Poly1305 암호화 | P0 | 1 | Pending |
| FR-13 | OS Keychain 연동: Master Key를 OS 키체인에 저장 + 파일 폴백 | P0 | 1 | Pending |
| FR-14 | Recovery Key: 생성, 표시, Master Password 복구 | P0 | 1 | Pending |
| FR-15 | --json 플래그: JSON 형식 출력 (AI 에이전트 파싱용) | P1 | 1 | Pending |
| FR-16 | --stdin 플래그: tene set에서 stdin 입력 (shell history 방지) | P1 | 1 | Pending |
| FR-17 | 로컬 감사 로그: 모든 시크릿 접근/수정 기록 (SQLite) | P2 | 1 | Pending |
| FR-18 | Claude Code 자동 인식: `tene init`시 CLAUDE.md 자동 생성 | P0 | 1 | Pending |
| FR-19 | Fake Door: `tene sync` -> waitlist 안내 화면 | P1 | 1 | Pending |
| FR-20 | Homebrew tap + curl 설치 + go install 배포 | P0 | 1 | Pending |
| FR-21 | goreleaser 멀티 플랫폼 빌드 (macOS/Linux/Windows x amd64/arm64) | P0 | 1 | Pending |
| FR-22 | `tene passwd`: Master Password 변경 -> 볼트 재암호화 -> 새 Recovery Key 발급 | P0 | 1 | Pending |
| FR-23 | `tene recover`: Recovery Key 입력 -> Master Password 재설정 | P0 | 1 | Pending |
| FR-24 | `tene init --cursor`: Cursor .cursorrules 생성 | P1 | 2 | Pending |
| FR-25 | `tene init --windsurf`: Windsurf .windsurfrules 생성 | P1 | 2 | Pending |
| FR-26 | Cloud 동기화: GitHub OAuth + ECS API + S3 백업 | P0 | 2 | Pending |
| FR-27 | 웹 대시보드: 시크릿 목록 조회 + 접근 로그 | P1 | 2 | Pending |
| FR-28 | Stripe 결제: $1/월 구독 | P0 | 2 | Pending |

### 15.2 비기능 요구사항

| 범주 | 기준 | Phase | 측정 방법 |
|------|------|:-----:|----------|
| **성능** | CLI 명령 응답 < 100ms (로컬, P95) | 1 | go test -bench |
| **성능** | CLI 시작 시간 < 20ms (cold start) | 1 | time 명령 |
| **오프라인** | 모든 기능 100% 오프라인 동작 | 1 | 네트워크 차단 테스트 |
| **보안** | 암호화: XChaCha20-Poly1305 (256-bit) | 1 | 코드 리뷰 |
| **보안** | KDF: Argon2id (64MB, 3 iterations) | 1 | 코드 리뷰 |
| **확장성** | 사용자당 시크릿 1,000개 지원 | 1 | SQLite 벤치마크 |
| **호환성** | macOS 12+, Ubuntu 20.04+, Windows 10+ (WSL) | 1 | CI 크로스 빌드 |
| **크기** | CLI 바이너리 < 20MB | 1 | goreleaser 빌드 |
| **코드** | 테스트 커버리지 > 80% (crypto > 95%) | 1 | go test -cover |
| **AI Agent** | CLAUDE.md 생성 시간 < 10ms | 1 | 벤치마크 |
| **보안** | Zero-Knowledge: Cloud에서 시크릿 복호화 불가 | 2 | 보안 감사 |
| **성능** | API 응답 < 500ms (P95) | 2 | CloudWatch 메트릭 |
| **동기화** | Cloud 동기화 < 3초 | 2 | E2E 테스트 |

---

## 16. 구현 우선순위

### 16.1 Phase 1: MVP Go CLI (2주)

```
Week 1: 코어 인프라 + 암호화 + Claude Code 통합
--------------------------------------------------
Day 1-2:
  1. Go 모듈 초기화 (go mod init)
  2. internal/crypto 패키지 구현:
     - Argon2id KDF (golang.org/x/crypto/argon2)
     - XChaCha20-Poly1305 encrypt/decrypt (nacl/secretbox)
     - Key derivation (Master Key -> Enc Key)
     - Recovery Key 생성/검증 (go-bip39)
  3. internal/crypto 테스트 (95%+ 커버리지)

Day 3-4:
  4. cobra CLI 기본 구조 (cmd/tene/main.go + internal/cli/)
  5. internal/vault 구현 (modernc.org/sqlite)
  6. internal/keychain 구현 (go-keyring) + 파일 폴백
  7. tene init (볼트 생성 + Master PW + Recovery Key)
  8. internal/claudemd 구현:
     - CLAUDE.md 자동 생성
     - 기존 파일 병합 로직

Day 5:
  9. tene set / get (암호화/복호화 + SQLite)
  10. tene run (환경변수 주입)

Week 2: CLI 확장 + Fake Door + 배포
--------------------------------------------------
Day 6-7:
  11. tene list / delete
  12. tene import / export (.env 형식)
  13. tene export --encrypted / import --encrypted
  14. tene env (환경 전환)
  15. --json / --stdin / --quiet 플래그

Day 8-9:
  16. tene passwd / tene recover
  17. tene sync (Fake Door -> waitlist 안내)
  18. tene whoami
  19. 로컬 감사 로그
  20. 에러 처리 + UX 개선
  21. 통합 테스트

Day 10:
  22. goreleaser 설정 + 멀티 플랫폼 빌드
  23. Homebrew tap 설정 (tomo-kay/homebrew-tap)
  24. install.sh curl 스크립트
  25. CI 파이프라인 (GitHub Actions)
  26. README + 퀵스타트 문서
```

### 16.2 Critical Path

```
Phase 1:
  internal/crypto -> internal/vault -> tene init (+ CLAUDE.md) -> tene set/get -> tene run
       |
       +---> 이 경로의 지연이 Phase 1 전체에 영향

공통:
  internal/crypto 품질 = 전체 제품 보안 신뢰도
  CLAUDE.md 자동 생성 = 핵심 차별점 (경쟁사 대비)
  goreleaser 빌드 = 배포 필수
```

---

## 17. 리스크 및 완화

| 리스크 | 영향 | 가능성 | 완화 방안 |
|--------|:----:|:------:|----------|
| **암호화 구현 결함**: 암호화 로직 버그로 시크릿 노출 | 치명적 | 중간 | golang.org/x/crypto 사용(Go 공식 라이브러리), internal/crypto 95%+ 테스트 커버리지, 오픈소스 커뮤니티 보안 리뷰 |
| **Master Password 분실**: 로컬 전용이라 서버에서 복구 불가 | 치명적 | 높음 | Recovery Key 생성 + 안전 보관 가이드 + `tene export --encrypted` 백업 유도 |
| **".env로 충분"**: 사용자가 전환 동기 부족 | 높음 | 높음 | `tene import .env` 한 줄 마이그레이션 + Claude Code 자동 인식 차별점 + "Git 커밋해도 안전" |
| **modernc.org/sqlite 성능**: 순수 Go SQLite가 네이티브 대비 느림 | 낮음 | 낮음 | CLI 사용 패턴에서 성능 차이 무시 가능 (~10-20% 차이). CGo 불필요 이점이 압도적 |
| **go-keyring 호환성**: OS Keychain이 특정 환경에서 동작하지 않음 | 중간 | 중간 | 폴백: 암호화된 파일 저장 (~/.tene/keyfile, 퍼미션 0600). Master Password 재입력 |
| **Shell History 노출**: `tene set KEY VALUE`가 shell history에 남음 | 높음 | 높음 | `--stdin` 플래그 지원. 문서에서 `echo VALUE \| tene set KEY --stdin` 권장 |
| **Go 바이너리 크기**: modernc.org/sqlite 포함 시 15-20MB | 낮음 | 중간 | UPX 압축 옵션 제공, goreleaser ldflags 최적화 |
| **Dotenvx 직접 경쟁**: 유사 포지셔닝 | 높음 | 중간 | Claude Code 자동 인식(CLAUDE.md) + SQLite 볼트(구조화) + 환경 전환으로 차별화 |
| **Cloud 수요 부재**: Fake Door Test 결과 수요 없음 | 중간 | 중간 | 로컬 기능 강화에 집중. Cloud 비용 $0 유지 |

---

## 18. 성공 기준

### 18.1 Phase 1 MVP Definition of Done

- [ ] Go CLI 코어 명령어 전체 작동 (init, set, get, run, list, delete, import, export, env, passwd, recover, whoami, sync)
- [ ] Master Password + Argon2id KDF + XChaCha20-Poly1305 암호화 작동
- [ ] Recovery Key 생성 및 복구 작동
- [ ] OS Keychain 연동 + 파일 폴백
- [ ] 오프라인 100% 동작 확인
- [ ] **Claude Code 자동 인식**: `tene init` 시 CLAUDE.md 자동 생성
- [ ] **Fake Door**: `tene sync` 실행 시 waitlist 안내 표시
- [ ] **암호화 백업**: `tene export --encrypted` / `tene import --encrypted` 동작
- [ ] macOS + Linux 크로스 빌드 성공
- [ ] **Homebrew tap** 배포 (`brew install tomo-kay/tap/tene`)
- [ ] **curl 설치 스크립트** 동작
- [ ] **go install** 동작
- [ ] 설치 -> 첫 시크릿 저장 1분 이내 달성
- [ ] --json 플래그 동작 (Claude Code 파싱)
- [ ] --stdin 플래그 동작 (shell history 방지)
- [ ] 전체 테스트 커버리지 > 80%
- [ ] internal/crypto 테스트 커버리지 > 95%
- [ ] 바이너리 크기 < 20MB
- [ ] CLI 응답 시간 < 100ms (로컬, P95)
- [ ] CLI 시작 시간 < 20ms
- [ ] GitHub 오픈소스 공개 (MIT License)
- [ ] 퀵스타트 README 작성

### 18.2 품질 기준

- [ ] golangci-lint 에러 0개
- [ ] go build 성공 (모든 타겟 플랫폼)
- [ ] 시크릿이 에러 메시지에 노출되지 않음

---

## 19. 컨벤션

### 19.1 코딩 컨벤션

| 범주 | 규칙 |
|------|------|
| **네이밍** | Go 표준: camelCase (비공개), PascalCase (공개), SCREAMING_SNAKE_CASE (상수) |
| **파일 이름** | snake_case.go (Go 표준) |
| **패키지 구조** | internal/ 하위에 기능별 패키지 분리 |
| **에러 처리** | fmt.Errorf + %w 래핑. 커스텀 에러 타입 (ErrNotInitialized, ErrDecryptFailed) |
| **주석** | godoc 스타일 (공개 API), 인라인 주석은 WHY에만 사용 |

### 19.2 Git 컨벤션

```
feat(cli): add tene init with CLAUDE.md generation
feat(crypto): implement XChaCha20-Poly1305 encryption
feat(cli): add tene sync fake door test
fix(keychain): fix fallback on Linux
chore(deps): update golang.org/x/crypto
test(crypto): add KDF edge case tests
docs: add quickstart guide
```

---

## 20. Phase 2 Cloud 상세 (수요 검증 후)

> 이 섹션은 Phase 2에서 구현할 Cloud 기능의 참고 자료이다. MVP에는 포함되지 않는다.

### 20.1 Cloud 아키텍처 (ECS Fargate + RDS PostgreSQL)

```
+-----------------------------------------------------------------------+
|                     $1/월 CLOUD TIER (Phase 2, 수요 검증 후)             |
|                                                                         |
|   CLI (sync cmd)          Web Dashboard (Next.js static)                |
|        |                       |                                        |
|        v                       v                                        |
|   NLB + ECS Fargate (Hono API)                                          |
|     Auth | Sync | Audit | Billing                                       |
|        |                       |                                        |
|        v                       v                                        |
|   RDS PostgreSQL          S3 Bucket                                     |
|   (users, devices,        (Encrypted Vault Blobs)                       |
|    subscriptions)                                                       |
|                                                                         |
|   CloudFront + Route 53 (tene.sh, api.tene.sh)                        |
+-----------------------------------------------------------------------+
```

### 20.2 Cloud API 엔드포인트 (Phase 2 참고)

| Method | Path | 설명 | 인증 |
|--------|------|------|:----:|
| POST | `/api/auth/github` | GitHub OAuth 코드 교환 | No |
| POST | `/api/auth/refresh` | JWT 토큰 갱신 | Refresh |
| POST | `/api/sync/upload` | 암호화 볼트 blob 업로드 | Yes |
| GET | `/api/sync/download` | 암호화 볼트 blob 다운로드 | Yes |
| GET | `/api/audit/logs` | 감사 로그 조회 | Yes |
| POST | `/api/billing/subscribe` | $1/월 구독 시작 | Yes |
