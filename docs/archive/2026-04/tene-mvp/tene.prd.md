# Tene PRD (Product Requirements Document) v4
## Go CLI + Claude Code 전용 MVP: 시크릿 관리

> v4 (2026-04-06) — Go 언어 전환 + Claude Code 전용 MVP 반영, 단일 바이너리 배포
> Version: 4.0
> Status: Draft
> Frameworks: Geoffrey Moore Beachhead, Lean PRD, Pre-mortem, INVEST User Stories
> Architecture: Go CLI Local-Only MVP ($0) → Cloud Phase 2 (수요 검증 후)

---

## Executive Summary

| 관점 | 내용 |
|------|------|
| **Problem** | 바이브코더/AI 에이전트 사용자의 75%가 시크릿을 부적절하게 관리하며, AI 에이전트는 시크릿 관리 도구의 존재를 자동으로 인식하지 못한다. 2025년 GitHub에 2,865만 시크릿 노출(34% 증가), AI 관련 81% 급증. |
| **Solution** | Tene는 **Claude Code가 자동으로 인식하는** 시크릿 관리 Go CLI. `tene init` → CLAUDE.md 자동 생성으로 Claude Code가 즉시 인식. 로컬 XChaCha20-Poly1305 암호화. 서버 없음, 가입 없음, 비용 $0. Go 단일 바이너리로 Node.js 불필요, ~5ms 시작. |
| **Target User** | 솔로 바이브코더 (Claude Code 사용자) |
| **Core Value** | "Claude Code 자동 인식 + 서버 없음 = 해킹 대상 없음 + Go 바이너리 = 의존성 제로" |
| **MVP 범위** | Go CLI + modernc.org/sqlite + golang.org/x/crypto. Cloud 없음, 서버 없음 |

---

## 1. 제품 비전 & 미션 (v4)

### 1.1 비전

> **모든 AI 에이전트가 시크릿을 자동으로 인식하고 안전하게 사용하는 세상**

### 1.2 미션

> **AI 에이전트가 시크릿 관리 도구를 자동으로 인식하는 최초의 CLI를 만든다**

### 1.3 핵심 원칙 (v4)

| 원칙 | 설명 |
|------|------|
| **Claude Code 자동 인식** | `tene init` → CLAUDE.md 생성 → Claude Code가 즉시 인식 |
| **Local-Only** | 서버 없이, 가입 없이, 로컬에서 완전히 동작 |
| **Zero Friction** | brew 한 줄 설치, 첫 시크릿 주입까지 1분 이내 |
| **Zero Dependencies** | Go 단일 바이너리, Node.js 불필요, ~5ms 시작 |
| **Server-Free Security** | 서버가 없으면 해킹 대상도 없다 |
| **정직한 범위** | 못 하는 것(만료/로테이션)을 명확히 밝힌다 |

---

## 2. ICP (Ideal Customer Profile) — v4

### 2.1 Primary ICP: 솔로 바이브코더 (무료 사용자)

| 속성 | 상세 |
|------|------|
| **프로필** | Claude Code로 사이드 프로젝트 1-5개 진행 |
| **관리 시크릿** | 5-15개 |
| **현재 방법** | .env 파일 + .gitignore |
| **핵심 Pain** | "AI가 시크릿 도구를 모른다" + "서버에 시크릿 맡기기 불안" |
| **Trigger** | "tene init 하면 AI가 자동으로 인식한다고? brew 한 줄이면 설치 끝?" |
| **WTP** | **$0** (무료 로컬이면 충분) |
| **핵심 가치** | Claude Code 자동 인식, 서버 없음, 가입 없음, 무료, 의존성 제로 |

### 2.2 Beachhead Segment

> **Beachhead: Claude Code 사용 솔로 바이브코더**
> 이들이 무료로 Tene를 채택 → Claude Code 자동 인식 경험 → 습관화 → Cloud 수요 발생 시 Phase 2
> Cursor/Windsurf 사용자 지원은 Phase 2

---

## 3. MVP 범위 정의 (v4)

### 3.1 MVP 원칙: Go CLI 로컬 전용, 서버 비용 $0

> **MVP = Go CLI + modernc.org/sqlite + golang.org/x/crypto**
> **서버 없음, 회원가입 없음, 비용 없음**
> **Cloud = Phase 2 (수요 검증 후)**

### 3.2 MVP 기능 범위

| 우선순위 | 기능 | 설명 | Phase |
|:--------:|------|------|:-----:|
| **P0** | `tene init` | 프로젝트 초기화 + **CLAUDE.md 자동 생성** | **MVP** |
| **P0** | `tene set KEY VALUE` | 시크릿 저장 (XChaCha20-Poly1305 암호화) | **MVP** |
| **P0** | `tene get KEY` | 시크릿 조회 (stdout, AI 에이전트 Bash 호출) | **MVP** |
| **P0** | `tene run -- COMMAND` | 시크릿이 환경변수로 주입된 상태에서 명령 실행 | **MVP** |
| **P0** | `tene list` | 프로젝트 시크릿 목록 (값 마스킹) | **MVP** |
| **P0** | `tene delete KEY` | 시크릿 삭제 | **MVP** |
| **P0** | Master Password 설정 | 첫 `tene init` 시 마스터 패스워드 설정 | **MVP** |
| **P1** | `tene import .env` | 기존 .env 파일에서 시크릿 일괄 가져오기 | **MVP** |
| **P1** | `tene export` | 시크릿을 .env 형식으로 내보내기 | **MVP** |
| **P1** | `tene export --encrypted` | 암호화된 백업 파일 생성 (수동 백업) | **MVP** |
| **P1** | `tene env [dev/prod]` | 환경별 시크릿 세트 전환 | **MVP** |
| **P1** | `tene sync` | **Fake Door Test** — waitlist 안내만 표시 | **MVP** |
| --- | `tene init --cursor` | .cursorrules에 tene 가이드 추가 | **Phase 2** |
| --- | `tene init --windsurf` | .windsurfrules에 tene 가이드 추가 | **Phase 2** |
| --- | Cloud 백업/동기화 | 암호화 클라우드 백업 + 멀티 디바이스 | **Phase 2** |
| --- | 웹 대시보드 | 시크릿 현황, 감사 로그 | **Phase 2** |
| --- | 팀 볼트 / RBAC | 팀 시크릿 공유 + 역할 기반 접근 | **Phase 2+ (가설)** |
| --- | MCP 서버 | AI 에이전트 네이티브 통합 | **Phase 2** |

### 3.3 Tene가 해결하는 것 / 못 하는 것 (정직한 범위)

| 구분 | 내용 |
|------|------|
| **해결** | 키값 안전 저장, XChaCha20-Poly1305 암호화, 프로젝트별 관리, 환경변수 주입, Claude Code 자동 인식 (CLAUDE.md) |
| **부분적** | 프로젝트 간 시크릿 재사용 (글로벌 키 공유는 향후 검토) |
| **못 함** | API key 만료 기간 확인, 자동 갱신/로테이션 (Phase 2+ 가설), 클라우드 동기화 (Phase 2), Cursor/Windsurf 지원 (Phase 2) |

### 3.4 MVP 핵심 플로우 (서버 없음, Claude Code 자동 인식)

```
[사용자의 시크릿 관리 - 로컬 전용 + Claude Code 자동 인식]

1. 설치 (가입 없음! Node.js 불필요!)
   # macOS
   $ brew install tomo-kay/tap/tene
   
   # Linux / Windows WSL
   $ curl -fsSL https://tene.sh/install.sh | sh
   
   # Go 사용자
   $ go install github.com/tomo-kay/tene@latest

2. 프로젝트 초기화 + Claude Code 자동 인식 설정
   $ cd my-project
   $ tene init
   ? Master Password 설정: ********
   ? Master Password 확인: ********
   > 로컬 볼트 생성 완료 (.tene/vault.db)
   > .gitignore에 .tene/ 추가 완료
   > CLAUDE.md 생성 완료 (Claude Code 자동 인식)
   > Recovery Key: apple banana cherry dolphin eagle frost grape harbor island jungle kite lemon (안전한 곳에 보관하세요!)

3. 시크릿 저장
   $ tene set STRIPE_KEY sk_test_xxxxx
   > STRIPE_KEY 저장 완료 (XChaCha20-Poly1305 암호화)

4. 시크릿 주입 (핵심!)
   $ tene run -- claude
   > 3개 환경변수 주입 완료
   > claude 실행 중...

   [Claude Code가 자동 인식]
   Claude Code: "이 프로젝트는 tene으로 시크릿을 관리합니다.
   STRIPE_KEY는 process.env.STRIPE_KEY로 참조합니다."

5. 수동 백업
   $ tene export --encrypted > ~/backup/my-project-secrets.enc
```

### 3.5 `tene sync` Fake Door Test

```
$ tene sync
> 클라우드 동기화 기능을 준비하고 있습니다!
> 관심이 있으시면 waitlist에 등록하세요.
>
> https://tene.sh/cloud
>
> 현재는 tene export --encrypted로 수동 백업을 권장합니다.
```

- 실제 Cloud 기능 없음 — waitlist 안내만 표시
- waitlist 가입 수로 Cloud 수요 확인
- 수요 확인 후 Phase 2에서 Cloud 구축

---

## 4. Claude Code 자동 인식 기능 상세 (v4 핵심)

### 4.1 CLAUDE.md 자동 생성

`tene init` 실행 시 프로젝트 루트에 CLAUDE.md 파일이 자동 생성된다:

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

**왜 이것이 핵심 차별점인가:**
- Claude Code는 프로젝트 열 때 CLAUDE.md를 자동으로 읽는다
- AI가 "이 프로젝트는 tene을 사용한다"를 즉시 인식
- 수동 설정, MCP 설정, SDK 통합 없이 자동 연동
- **Vault, Doppler, Infisical, 1Password 중 어느 것도 이 기능이 없다**

### 4.2 Cursor/Windsurf 지원 (Phase 2)

MVP에서는 Claude Code 전용에 집중한다. Cursor (.cursorrules) 및 Windsurf (.windsurfrules) 지원은 Phase 2에서 `--cursor`, `--windsurf` 플래그로 추가한다.

### 4.3 GitHub Secrets과의 관계

| 영역 | 도구 |
|------|------|
| CI/CD 파이프라인 (GitHub Actions) | GitHub Secrets |
| 로컬 개발 + AI 에이전트 | **Tene** |

Tene와 GitHub Secrets는 **완전히 다른 제품**이며 겹치지 않는다.

---

## 5. 보안 신뢰 전략 (v4)

### 5.1 MVP 보안 모델: "서버가 없다"

```
[사용자 디바이스 — 오직 여기에만 존재]
+------------------------------------------+
|                                          |
|  Master Password (사용자만 알고 있음)     |
|       |                                  |
|       v                                  |
|  Argon2id KDF (memory: 64MB, iter: 3)    |
|  (golang.org/x/crypto/argon2)            |
|       |                                  |
|       v                                  |
|  XChaCha20-Poly1305 Encrypt              |
|  (golang.org/x/crypto/nacl/secretbox)    |
|       |                                  |
|       v                                  |
|  SQLite DB (.tene/vault.db)              |
|  (modernc.org/sqlite — 순수 Go, CGo 없음)|
|                                          |
|  ** 네트워크 통신: 없음 **                |
|  ** 서버: 없음 **                        |
|  ** 해킹 대상: 없음 **                   |
+------------------------------------------+
```

### 5.2 보안 차별화 메시지

> **"Tene는 서버가 없습니다. 해킹 대상이 없습니다."**
> **"코드는 오픈소스입니다. 직접 확인하세요."**

---

## 6. User Stories (v4 — MVP 전용)

### 6.1 Epic 1: 로컬 시크릿 관리 + Claude Code 자동 인식 (MVP Core)

| ID | User Story | 우선순위 |
|----|-----------|:--------:|
| US-01 | 바이브코더로서, `tene init`으로 Master Password를 설정하고 로컬 볼트를 생성하며 **CLAUDE.md를 자동 생성**할 수 있다, Claude Code가 즉시 인식하므로. | **P0** |
| US-02 | 바이브코더로서, `tene set KEY VALUE`로 시크릿을 로컬에 암호화 저장할 수 있다. | **P0** |
| US-03 | 바이브코더로서, `tene get KEY`로 시크릿을 조회할 수 있다, AI 에이전트가 Bash에서 사용하므로. | **P0** |
| US-04 | 바이브코더로서, `tene run -- claude`로 시크릿이 주입된 환경에서 Claude Code를 실행할 수 있다. | **P0** |
| US-05 | 바이브코더로서, `tene list`로 시크릿 목록을 볼 수 있다. | **P0** |
| US-06 | 바이브코더로서, `tene delete KEY`로 시크릿을 삭제할 수 있다. | **P0** |
| US-07 | 바이브코더로서, `tene import .env`로 기존 .env 파일을 마이그레이션할 수 있다. | **P1** |
| US-08 | 바이브코더로서, `tene export`로 .env 형식으로 내보낼 수 있다. | **P1** |
| US-09 | 바이브코더로서, `tene export --encrypted`로 암호화된 백업을 생성할 수 있다. | **P1** |
| US-10 | 바이브코더로서, `tene env dev/prod`로 환경을 전환할 수 있다. | **P1** |
| US-11 | 바이브코더로서, 오프라인에서도 모든 CLI 명령이 동작한다. | **P0** |
| US-12 | 바이브코더로서, `brew install tomo-kay/tap/tene` 한 줄로 Node.js 없이 설치할 수 있다. | **P0** |

### 6.2 Epic 2: 수요 검증 (MVP 내장)

| ID | User Story | 우선순위 |
|----|-----------|:--------:|
| US-13 | 바이브코더로서, `tene sync`를 실행하면 Cloud waitlist 안내를 볼 수 있다 (Fake Door). | **P1** |

### 6.3 Epic 3: 멀티 에디터 지원 (Phase 2)

| ID | User Story | 상태 |
|----|-----------|:----:|
| US-14 | Cursor 사용자로서, `tene init --cursor`로 .cursorrules에 가이드를 추가할 수 있다. | **Phase 2** |
| US-15 | Windsurf 사용자로서, `tene init --windsurf`로 .windsurfrules에 가이드를 추가할 수 있다. | **Phase 2** |

### 6.4 Epic 4: Cloud (Phase 2 — 수요 확인 후)

| ID | User Story | 상태 |
|----|-----------|:----:|
| US-16 | 사용자로서, 클라우드에 암호화 백업 + 멀티 디바이스 동기화할 수 있다. | **Phase 2** |
| US-17 | 사용자로서, 웹 대시보드에서 시크릿 현황을 볼 수 있다. | **Phase 2** |

### 6.5 Epic 5: 팀 기능 (Phase 2+ — 가설)

| ID | User Story | 상태 |
|----|-----------|:----:|
| US-18 | 팀 리더로서, 팀 볼트를 생성하고 팀원을 초대할 수 있다. | **가설** |

---

## 7. 기술 요구사항 (v4)

### 7.1 CLI 기술 스택 (MVP)

| 구성요소 | 기술 | 근거 |
|----------|------|------|
| **언어** | **Go** | 단일 바이너리, ~5ms 시작, 런타임 불필요 |
| **CLI 프레임워크** | **cobra** | Go CLI 표준, 풍부한 생태계 |
| **로컬 DB** | **modernc.org/sqlite** | 순수 Go SQLite, CGo 없음 → 크로스 컴파일 용이 |
| **암호화** | **golang.org/x/crypto/nacl/secretbox** (XChaCha20-Poly1305) | Go 표준 암호화 라이브러리 |
| **KDF** | **golang.org/x/crypto/argon2** (Argon2id) | 현대적 키 유도 |
| **Keychain** | **zalando/go-keyring** | OS Keychain 연동 (macOS Keychain, Linux Secret Service) |
| **니모닉** | **tyler-smith/go-bip39** | BIP-39 Recovery Key 생성 |
| **빌드/배포** | **goreleaser** | 멀티 플랫폼 바이너리 + Homebrew tap |
| **테스트** | **Go testing** | 표준 테스트 프레임워크 |
| **린트** | **golangci-lint** | Go 표준 린트 도구 |

### 7.2 프로젝트 구조 (Go + Next.js 혼합)

```
cmd/tene/              ← Go CLI 엔트리포인트
  main.go
internal/
  crypto/              ← 암호화 모듈 (XChaCha20-Poly1305 + Argon2id)
  vault/               ← SQLite 볼트 관리 (modernc.org/sqlite)
  keychain/            ← OS Keychain 연동 (go-keyring)
  claudemd/            ← CLAUDE.md 생성
  recovery/            ← BIP-39 Recovery Key
  commands/            ← cobra CLI 명령어
apps/web/              ← Next.js 랜딩페이지 (유지)
go.mod
go.sum
```

### 7.3 Phase 2 Cloud 기술 스택 (수요 확인 후, 서버리스 사용 안 함)

| 구성요소 | 기술 | 근거 |
|----------|------|------|
| **컴퓨팅** | **ECS Fargate** | 트래픽 증가에 선형적 비용 |
| **로드밸런서** | **NLB** | 낮은 지연시간 |
| **데이터베이스** | **RDS PostgreSQL** | 안정적, 확장 가능 |
| **저장소** | **S3** | 암호화 blob 저장 |

**왜 서버리스를 사용하지 않는가:**
- 트래픽 늘면 비용 역전 (Lambda/API Gateway 예측 어려움)
- ECS Fargate + NLB + RDS가 트래픽 증가에 선형적 비용
- Steve가 서버리스를 선호하지 않음

### 7.4 비기능 요구사항

| 항목 | 요구사항 | 측정 기준 |
|------|----------|----------|
| **성능** | CLI 명령 응답 < 10ms (cold start ~5ms) | P95 latency |
| **오프라인** | 100% 동작 | 네트워크 차단 테스트 |
| **보안** | Master Password + XChaCha20-Poly1305 + Argon2id | 코드 리뷰 + 보안 감사 |
| **AI 인식** | CLAUDE.md 자동 생성 | 통합 테스트 |
| **용량** | 시크릿 1,000개 이상 지원 | SQLite 성능 테스트 |
| **호환성** | macOS, Linux, Windows (WSL) | CI 테스트 (goreleaser) |
| **바이너리 크기** | < 15MB | 빌드 검증 |
| **설치 시간** | brew: < 10초, curl: < 10초 | 실측 |

---

## 8. GTM (Go-To-Market) 전략 (v4)

### 8.1 핵심 GTM: 오픈소스 Go CLI + Claude Code 자동 인식 → 커뮤니티 성장 → 수요 검증

```
[Stage 1: 씨앗 뿌리기 (Month 1-3)]
brew install tomo-kay/tap/tene
   |
   v
"Claude Code가 자동 인식하는 시크릿 관리. Go 바이너리. 5ms." 메시지
   |
   v
GitHub 오픈소스 → Stars → 개발자 발견
   |
   v
목표: 설치 10,000 / GitHub Stars 1,000

[Stage 2: 수요 검증 (Month 3-6)]
tene sync Fake Door → waitlist 가입 수 확인
   |
   v
설치 수 추세, GitHub Stars 추세 분석
   |
   v
Cloud 구축 여부 결정 + Cursor/Windsurf 지원 시기 결정

[Stage 3: Phase 2 (Month 6-12, 수요 확인 시)]
Cloud 구축: ECS Fargate + NLB + RDS PostgreSQL + S3
   |
   v
--cursor, --windsurf 플래그 추가
   |
   v
팀 기능 수요 확인 (Fake Door Test)
```

### 8.2 설치 방법 (v4)

| 플랫폼 | 설치 명령 | 비고 |
|--------|----------|------|
| **macOS** | `brew install tomo-kay/tap/tene` | Homebrew tap |
| **Linux** | `curl -fsSL https://tene.sh/install.sh \| sh` | 단일 바이너리 다운로드 |
| **Windows** | WSL에서 curl 설치 | WSL 권장 |
| **Go 사용자** | `go install github.com/tomo-kay/tene@latest` | 소스 빌드 |

### 8.3 수요 검증 방법

| Step | 방법 | 성공 기준 |
|------|------|----------|
| **Step 1** | CLI 출시 → 설치 수, GitHub Stars 추적 | 주간 설치 500+, Stars 1,000+ |
| **Step 2** | `tene sync` Fake Door → waitlist 가입 수 | CLI 사용자의 10%+ waitlist 등록 |
| **Step 3** | waitlist 반응 보고 Cloud 구축 여부 결정 | waitlist 100명+ → Cloud 구축 시작 |

### 8.4 Battlecards (v4)

#### vs .env 파일 (가장 큰 경쟁)
| 고객 반론 | Tene 대응 |
|-----------|----------|
| ".env로 충분한데?" | "Tene도 로컬이고 무료입니다. 하지만 암호화되어 있고, **Claude Code가 자동으로 인식**합니다. `tene import .env` 한 줄이면 전환 끝." |

#### vs GitHub Secrets
| 고객 반론 | Tene 대응 |
|-----------|----------|
| "GitHub Secrets 쓰면 되잖아?" | "GitHub Secrets는 CI/CD 파이프라인 전용입니다. 로컬 개발 중에는 사용할 수 없습니다. Tene는 로컬 개발 + Claude Code 전용입니다. 완전히 다른 제품이며 보완적입니다." |

#### vs Doppler
| 고객 반론 | Tene 대응 |
|-----------|----------|
| "Doppler 잘 쓰고 있는데?" | "Doppler는 서버에 시크릿을 보내야 하고 $21/유저/월입니다. Tene는 로컬 전용이고 무료입니다. Claude Code가 자동 인식합니다." |

#### vs Infisical
| 고객 반론 | Tene 대응 |
|-----------|----------|
| "Infisical도 오픈소스인데?" | "Infisical은 셀프호스팅 서버가 필요합니다. Tene는 서버 자체가 불필요하고, Claude Code가 자동 인식합니다." |

#### vs Dotenvx
| 고객 반론 | Tene 대응 |
|-----------|----------|
| "Dotenvx도 로컬이고 암호화하던데?" | "Dotenvx는 .env 파일을 암호화합니다. Tene는 SQLite 볼트 + Claude Code 자동 인식(CLAUDE.md)을 제공합니다. AI가 Tene 사용법을 알아서 인식합니다." |

---

## 9. Pre-mortem 분석 (v4)

### 9.1 "Tene가 실패한 이유는..." (v4)

| # | 실패 시나리오 | 확률 | 심각도 | 대응 |
|---|-------------|:----:|:------:|------|
| **R1** | ".env로 충분": 전환 동기 부족 | 높음 | 높음 | Claude Code 자동 인식이 .env에 없는 핵심 가치. `tene import .env` 전환 비용 제로 |
| **R2** | Master Password 분실: 복구 불가 | 중간 | 치명적 | Recovery Key + `tene export --encrypted` 수동 백업 |
| **R3** | CLAUDE.md가 효과 없음: AI가 인식해도 큰 차이 없음 | 중간 | 높음 | 사용자 테스트로 빠르게 검증, 가치 없으면 다른 차별점 강화 |
| **R4** | Dotenvx AS2와 직접 경쟁 | 중간 | 높음 | Claude Code 자동 인식 + Go 바이너리로 차별화 |
| **R5** | Claude Code 전용이라 Cursor 사용자 이탈 | 중간 | 중간 | Phase 2에서 Cursor/Windsurf 지원 빠르게 추가, MVP 완성도 우선 |
| **R6** | Cloud 수요 없음: waitlist 반응 미미 | 중간 | 중간 | CLI 무료 도구로서의 가치에 집중, 수익 모델 재검토 |
| **R7** | 만료/로테이션 불가에 대한 불만 | 낮음 | 중간 | 정직하게 "못 하는 것" 명시, Phase 2+ 가설로 관리 |

---

## 10. 성공 지표 (v4)

### 10.1 North Star Metric

> **주간 설치 수** (Weekly Installs — Homebrew + curl + go install 합산)

### 10.2 단계별 KPI (v4)

| 지표 | MVP (3개월) | 수요 검증 (6개월) | Phase 2 (12개월) |
|------|:----------:|:----------------:|:---------------:|
| 주간 설치 수 | 500 | 2,000 | 5,000 |
| GitHub Stars | 1,000 | 3,000 | 10,000 |
| 활성 사용자 (WAU) | 500 | 2,000 | 5,000 |
| `tene sync` Fake Door 실행 | -- | 200+ | -- |
| waitlist 등록 | -- | 100+ | -- |
| Cloud 유료 사용자 | -- | -- | 수요에 따라 |

### 10.3 MVP 출시 기준 (Definition of Done) — v4

- [ ] `tene init` Master Password + SQLite 볼트 + **CLAUDE.md 자동 생성**
- [ ] `tene set/get/list/delete` 시크릿 CRUD (로컬 암호화)
- [ ] `tene run -- COMMAND` 환경변수 주입
- [ ] `tene env dev/prod` 환경 전환
- [ ] `tene import .env` 마이그레이션
- [ ] `tene export` .env 내보내기
- [ ] `tene export --encrypted` 암호화 수동 백업
- [ ] `tene sync` Fake Door Test (waitlist 안내)
- [ ] XChaCha20-Poly1305 + Argon2id 암호화 구현 (golang.org/x/crypto)
- [ ] Recovery Key 생성 + 복구 플로우 (go-bip39)
- [ ] 오프라인 100% 동작 확인
- [ ] macOS (brew) + Linux (curl) + WSL 설치 가능
- [ ] CLI cold start < 10ms (P95)
- [ ] 바이너리 크기 < 15MB
- [ ] Go testing 커버리지 > 80%
- [ ] golangci-lint 통과
- [ ] goreleaser 빌드 + Homebrew tap 설정
- [ ] GitHub MIT 오픈소스 공개

---

## 11. 로드맵 (v4)

```
2026 Q2 (Apr-Jun)          2026 Q3 (Jul-Sep)          2026 Q4 (Oct-Dec)
-------------------        -------------------        -------------------
Phase 1: Go CLI MVP        수요 검증                   Phase 2: 수요 확인 시

[Go CLI 코어 - 무료, $0]   [검증]                     [Cloud 구축 (수요 시)]
- set/get/run/list         - 설치 수 추적              - ECS Fargate + NLB
- delete/import/export     - GitHub Stars 추적         - RDS PostgreSQL + S3
- Master Password          - tene sync Fake Door       - 멀티 디바이스 동기화
- SQLite 로컬 볼트          - waitlist 분석             - 웹 대시보드
- env 환경 전환             - Cloud 구축 결정
- Recovery Key             
                           [기능 강화]                [멀티 에디터]
[Claude Code 자동 인식]     - CLI 안정화               - --cursor 플래그
- CLAUDE.md 자동 생성      - 사용자 피드백 반영        - --windsurf 플래그
                           - export --encrypted
[배포]                       개선                     [시장 검증]
- goreleaser 빌드                                     - 팀 기능 Fake Door
- Homebrew tap             [수요 검증]                - 에이전트 스코핑 검토
- curl 설치 스크립트        - tene sync Fake Door
- Hacker News              - waitlist 가입 수 확인    [보안]
- GeekNews                                            - 보안 감사 (외부)
                                                      - 버그 바운티 검토
```

---

## Attribution

이 PRD는 다음 PM Agent Team 분석 결과를 통합하여 작성되었습니다:

| 문서 | 경로 | 내용 |
|------|------|------|
| Discovery v4 | `docs/00-pm/tene-discovery.md` | Claude Code 자동 인식 기회, Go CLI, Fake Door Test |
| Strategy v4 | `docs/00-pm/tene-strategy.md` | Claude Code 자동 인식 포지셔닝, Go 바이너리 $0 MVP 전략 |
| Market Research v4 | `docs/00-pm/tene-research.md` | AI Agent 통합 비교, GitHub Secrets 차별점 |

---

*Generated by PM Agent Team | 2026-04-06 (v4)*
*Architecture: Go CLI Local-Only MVP ($0) + Claude Code 자동 인식*
*Tech Stack: Go + cobra + modernc.org/sqlite + golang.org/x/crypto + goreleaser*
*Cloud: Phase 2 (수요 검증 후) — ECS Fargate + NLB + RDS PostgreSQL (서버리스 X)*
