# Tene CLI 구현 계획 (Go Implementation Plan)

> **Summary**: Tene Go CLI의 패키지별 구현 순서, 의존성, 빌드/배포, 테스트 전략을 정의하는 실행 계획
>
> **Project**: Tene
> **Version**: 1.0.0
> **Author**: CTO Lead (Steve)
> **Date**: 2026-04-06
> **Status**: Draft
> **Related**: [tene-mvp.plan.md](./tene-mvp.plan.md) (전체 프로젝트 방향), [tene-architecture-briefing.md](../tene-architecture-briefing.md) (아키텍처 브리핑)

---

## Executive Summary

| 관점 | 내용 |
|------|------|
| **Problem** | Tene MVP Plan v4에서 Go 전환이 결정되었으나, Go 패키지별 구현 순서/인터페이스/테스트 전략이 정의되지 않아 즉시 코딩 착수가 불가능 |
| **Solution** | 7개 internal 패키지 + cmd/tene 엔트리포인트를 의존성 순서대로 구현. crypto -> recovery -> vault -> keychain -> claudemd -> cli -> cmd/tene 순서. goreleaser + Homebrew tap + install.sh 배포 |
| **Function/UX Effect** | 개발자가 이 문서만 보고 Go 코드를 즉시 작성 가능. 패키지별 독립 테스트 후 통합. Week 1에 핵심 암호화+저장소, Week 2에 CLI 명령어+배포 완료 |
| **Core Value** | "코딩 시작 전 모든 인터페이스가 확정된 상태". 패키지 간 의존성이 명확하여 병렬 개발 가능. 테스트 커버리지 90%+ 목표 |

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | Go CLI 구현을 위한 상세 실행 계획 부재. MVP Plan이 전체 방향만 다루고 코드 레벨 구현 가이드가 없음 |
| **WHO** | Tene CLI를 구현하는 Go 개발자 (Steve + Claude Code AI Agent) |
| **RISK** | 암호화 구현 결함 시 제품 신뢰도 치명적 타격. modernc.org/sqlite 순수 Go 성능 검증 필요. OS Keychain 크로스 플랫폼 호환성 |
| **SUCCESS** | 13개 CLI 명령어 로컬 동작, 오프라인 100%, XChaCha20-Poly1305 + Argon2id 암호화, CLAUDE.md 자동 생성, 테스트 커버리지 90%+, goreleaser 빌드 성공 |
| **SCOPE** | Week 1: crypto + recovery + vault + keychain (핵심 인프라). Week 2: claudemd + cli + cmd/tene + goreleaser + Homebrew tap |

---

## 1. 구현 범위

### 1.1 In Scope

- Go 프로젝트 초기화 (go.mod, 디렉토리 구조)
- 7개 internal 패키지 구현
- cmd/tene/main.go 엔트리포인트
- 13개 CLI 명령어 (init, set, get, run, list, delete, import, export, env, passwd, recover, sync, whoami)
- 패키지별 단위 테스트 + CLI 통합 테스트
- goreleaser 설정 + Homebrew tap + install.sh
- GitHub Actions CI/CD (lint, test, build, release)
- golangci-lint 설정

### 1.2 Out of Scope

- Cloud API 서버 (Phase 2)
- --cursor, --windsurf 플래그 (Phase 2)
- 웹 대시보드 (Phase 2)
- 팀 볼트 / RBAC (Phase 3)
- apps/web Next.js 랜딩페이지 수정 (별도 작업)

---

## 2. Go 프로젝트 구조

### 2.1 go.mod

```go
module github.com/tomo-kay/tene

go 1.22

require (
    github.com/spf13/cobra         v1.8.1
    modernc.org/sqlite              v1.34.5
    golang.org/x/crypto             v0.31.0
    golang.org/x/term               v0.27.0
    github.com/zalando/go-keyring   v0.2.6
    github.com/tyler-smith/go-bip39 v1.1.0
    github.com/stretchr/testify     v1.10.0
)
```

### 2.2 디렉토리 구조 (상세)

```
tene/
├── cmd/tene/
│   └── main.go                        # cobra 루트 명령어 + Execute()
│
├── internal/
│   ├── crypto/
│   │   ├── kdf.go                     # Argon2id 키 유도 함수
│   │   ├── encrypt.go                 # XChaCha20-Poly1305 암호화
│   │   ├── decrypt.go                 # XChaCha20-Poly1305 복호화
│   │   ├── keymanager.go              # MasterKey 파생, HKDF 기반 서브키 유도
│   │   ├── errors.go                  # 패키지 에러 타입 정의
│   │   └── crypto_test.go            # 테스트 (목표: 95%+ 커버리지)
│   │
│   ├── recovery/
│   │   ├── mnemonic.go                # BIP-39 니모닉 생성/검증
│   │   ├── recover.go                 # Recovery Key로 Master Key 복구
│   │   ├── errors.go                  # 패키지 에러 타입
│   │   └── mnemonic_test.go
│   │
│   ├── vault/
│   │   ├── vault.go                   # Vault struct + CRUD 메서드
│   │   ├── schema.go                  # CREATE TABLE DDL
│   │   ├── migration.go               # 스키마 버전 관리 + 마이그레이션
│   │   ├── models.go                  # Secret, Environment, AuditLog struct
│   │   ├── errors.go                  # 패키지 에러 타입
│   │   └── vault_test.go
│   │
│   ├── keychain/
│   │   ├── keychain.go                # KeyStore 인터페이스 + go-keyring 구현
│   │   ├── fallback.go                # 파일 기반 폴백 (환경변수로 제어)
│   │   ├── errors.go                  # 패키지 에러 타입
│   │   └── keychain_test.go
│   │
│   ├── claudemd/
│   │   ├── generator.go               # CLAUDE.md 생성/병합 로직
│   │   ├── template.go                # CLAUDE.md 템플릿 상수
│   │   └── generator_test.go
│   │
│   └── cli/
│       ├── root.go                    # rootCmd + 글로벌 플래그 (--json, --env)
│       ├── init.go                    # tene init
│       ├── set.go                     # tene set KEY VALUE
│       ├── get.go                     # tene get KEY [--json]
│       ├── run.go                     # tene run -- CMD
│       ├── list.go                    # tene list [--json]
│       ├── delete.go                  # tene delete KEY
│       ├── import_cmd.go              # tene import .env / --encrypted
│       ├── export.go                  # tene export / --encrypted
│       ├── env.go                     # tene env [name]
│       ├── passwd.go                  # tene passwd
│       ├── recover.go                 # tene recover
│       ├── sync_cmd.go               # tene sync (Fake Door)
│       ├── whoami.go                  # tene whoami
│       ├── helpers.go                 # 공통 헬퍼 (패스워드 입력, 에러 출력)
│       └── cli_test.go               # CLI 통합 테스트
│
├── .github/workflows/
│   ├── ci.yml                         # golangci-lint + go test + go build
│   └── release.yml                    # goreleaser 자동 릴리즈 (tag push)
│
├── go.mod
├── go.sum
├── .goreleaser.yml                    # 멀티 플랫폼 빌드 + Homebrew tap
├── .golangci.yml                      # 린터 설정
├── Makefile                           # 개발 편의 (make build, make test, make lint)
├── install.sh                         # curl -fsSL https://tene.sh/install.sh | sh
├── CLAUDE.md                          # 프로젝트 컨텍스트
└── .gitignore
```

**참고**: `import.go`와 `sync.go`는 Go 예약어/표준 패키지명과 충돌 가능하므로 `import_cmd.go`, `sync_cmd.go`로 명명.

---

## 3. 패키지별 구현 계획

### 3.1 구현 순서 (의존성 순서)

```
Phase 1 (Week 1): 핵심 인프라
──────────────────────────────
1. internal/crypto     (의존성 없음 - 독립)
2. internal/recovery   (의존: crypto)
3. internal/vault      (의존성 없음 - 독립, crypto와 병렬 가능)
4. internal/keychain   (의존성 없음 - 독립)

Phase 2 (Week 2): CLI + 배포
──────────────────────────────
5. internal/claudemd   (의존성 없음 - 독립)
6. internal/cli        (의존: crypto, recovery, vault, keychain, claudemd)
7. cmd/tene            (의존: cli)
8. goreleaser + CI/CD  (의존: cmd/tene)
```

### 3.2 internal/crypto -- 암호화 코어

**역할**: Argon2id KDF, XChaCha20-Poly1305 암호화/복호화, HKDF 기반 서브키 유도

**구현 파일**:

| 파일 | 함수 | 설명 |
|------|------|------|
| `kdf.go` | `DeriveKey(password, salt) -> masterKey` | Argon2id KDF (64MB, 3 iter, 256-bit) |
| `kdf.go` | `GenerateSalt() -> salt` | 128-bit 랜덤 salt 생성 |
| `keymanager.go` | `DeriveSubKey(masterKey, purpose, length) -> subKey` | HKDF-SHA256 기반 서브키 유도 |
| `encrypt.go` | `Encrypt(key, plaintext, aad) -> ciphertext` | XChaCha20-Poly1305 + 192-bit nonce + AAD |
| `decrypt.go` | `Decrypt(key, ciphertext, aad) -> plaintext` | XChaCha20-Poly1305 복호화 |
| `errors.go` | `ErrDecryptionFailed`, `ErrInvalidKey` 등 | sentinel 에러 정의 |

**외부 의존성**: `golang.org/x/crypto/argon2`, `golang.org/x/crypto/nacl/secretbox`, `golang.org/x/crypto/hkdf`

**테스트 시나리오**:
- 암호화 -> 복호화 라운드트립 정상
- 잘못된 키로 복호화 시 ErrDecryptionFailed
- 다른 AAD로 복호화 시 실패 (변조 감지)
- salt 고유성 검증 (중복 없음)
- 대용량 데이터 (1MB) 암복호화

### 3.3 internal/recovery -- BIP-39 Recovery Key

**역할**: 12단어 니모닉 생성, 니모닉에서 Recovery Key 유도, Master Key 암호화/복구

**구현 파일**:

| 파일 | 함수 | 설명 |
|------|------|------|
| `mnemonic.go` | `GenerateMnemonic() -> mnemonic` | 128-bit entropy -> 12단어 BIP-39 |
| `mnemonic.go` | `ValidateMnemonic(mnemonic) -> bool` | 니모닉 유효성 검증 |
| `recover.go` | `EncryptMasterKey(masterKey, mnemonic) -> blob` | Recovery Key로 Master Key 암호화 |
| `recover.go` | `RecoverMasterKey(blob, mnemonic) -> masterKey` | Recovery Key로 Master Key 복구 |

**의존성**: internal/crypto, github.com/tyler-smith/go-bip39

**테스트 시나리오**:
- 니모닉 생성 -> 12단어 확인
- 니모닉 검증 (유효/무효)
- Master Key 암호화 -> 복구 라운드트립
- 잘못된 니모닉으로 복구 시 실패

### 3.4 internal/vault -- SQLite 볼트

**역할**: SQLite DB 초기화, 시크릿 CRUD, 환경 관리, 감사 로그, 스키마 마이그레이션

**구현 파일**:

| 파일 | 함수/구조체 | 설명 |
|------|------------|------|
| `models.go` | `Secret`, `Environment`, `AuditLog` struct | 데이터 모델 |
| `schema.go` | `initSchema(db)` | CREATE TABLE DDL 실행 |
| `vault.go` | `New(dbPath) -> *Vault` | Vault 생성자 |
| `vault.go` | `SetSecret(name, encryptedValue, env)` | 시크릿 저장 (UPSERT) |
| `vault.go` | `GetSecret(name, env) -> Secret` | 시크릿 조회 |
| `vault.go` | `ListSecrets(env) -> []Secret` | 시크릿 목록 |
| `vault.go` | `DeleteSecret(name, env)` | 시크릿 삭제 |
| `vault.go` | `SetMeta(key, value)` | 볼트 메타데이터 저장 |
| `vault.go` | `GetMeta(key) -> value` | 볼트 메타데이터 조회 |
| `vault.go` | `AddAuditLog(action, resource)` | 감사 로그 기록 |
| `vault.go` | `ListEnvironments() -> []Environment` | 환경 목록 |
| `migration.go` | `migrate(db)` | 스키마 버전 확인 + 마이그레이션 |

**외부 의존성**: `modernc.org/sqlite`

**SQLite 스키마 (v1)**:
```sql
-- 볼트 메타데이터
CREATE TABLE vault_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
-- 초기값: schema_version=1, created_at, vault_name, kdf_salt, recovery_blob

-- 시크릿 저장
CREATE TABLE secrets (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    name            TEXT NOT NULL,
    encrypted_value TEXT NOT NULL,     -- base64(nonce + ciphertext)
    environment     TEXT NOT NULL DEFAULT 'default',
    version         INTEGER NOT NULL DEFAULT 1,
    created_at      TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at      TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(name, environment)
);

-- 환경 관리
CREATE TABLE environments (
    name       TEXT PRIMARY KEY,
    is_active  INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
-- 초기값: default (is_active=1)

-- 감사 로그
CREATE TABLE audit_log (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    action        TEXT NOT NULL,     -- secret.read, secret.write, secret.delete, vault.init, ...
    resource_name TEXT,
    timestamp     TEXT NOT NULL DEFAULT (datetime('now'))
);
```

**테스트 시나리오**:
- DB 초기화 및 스키마 생성
- CRUD 라운드트립 (Set -> Get -> List -> Delete)
- UPSERT 동작 (같은 key+env에 덮어쓰기)
- 환경별 시크릿 격리
- 감사 로그 기록 확인
- 존재하지 않는 시크릿 조회 시 ErrSecretNotFound

### 3.5 internal/keychain -- OS Keychain 연동

**역할**: Master Key를 OS 네이티브 키체인에 저장/로드/삭제. 키체인 사용 불가 시 파일 폴백

**구현 파일**:

| 파일 | 함수 | 설명 |
|------|------|------|
| `keychain.go` | `KeyStore` 인터페이스 | `Store(key)`, `Load()`, `Delete()` |
| `keychain.go` | `NewKeyringStore(service, user)` | go-keyring 기반 구현 |
| `fallback.go` | `NewFileStore(path)` | 파일 기반 폴백 (0600 퍼미션) |

**외부 의존성**: `github.com/zalando/go-keyring`

**폴백 전략**:
1. go-keyring으로 OS 키체인 시도
2. 실패 시 (CI, Docker, 미지원 OS) -> `~/.tene/keyfile` 파일 폴백 (0600)
3. `TENE_KEYCHAIN_FALLBACK=file` 환경변수로 강제 파일 모드

**테스트 시나리오**:
- Store -> Load 라운드트립
- Delete 후 Load 시 ErrKeyNotFound
- 파일 폴백 모드 동작
- 퍼미션 0600 검증 (파일 폴백)

### 3.6 internal/claudemd -- CLAUDE.md 생성

**역할**: `tene init` 시 CLAUDE.md를 생성하거나 기존 파일에 Secrets Management 섹션 추가

**구현 파일**:

| 파일 | 함수 | 설명 |
|------|------|------|
| `template.go` | `SecretsMdTemplate` 상수 | CLAUDE.md에 삽입할 영어 텍스트 |
| `generator.go` | `Generate(projectDir) -> error` | CLAUDE.md 생성/병합 |
| `generator.go` | `HasTeneSection(path) -> bool` | 기존 tene 섹션 존재 확인 |

**병합 정책**:

| 상황 | 동작 |
|------|------|
| CLAUDE.md 없음 | 새로 생성 |
| CLAUDE.md 있음 + tene 섹션 없음 | 파일 끝에 섹션 추가 |
| CLAUDE.md 있음 + tene 섹션 있음 | 스킵 (중복 방지) |

**테스트 시나리오**:
- 새 파일 생성 확인
- 기존 파일에 섹션 추가
- 이미 존재할 때 스킵
- 기존 내용 보존 확인

### 3.7 internal/cli -- Cobra 명령어 정의

**역할**: 13개 CLI 명령어 정의. 각 명령어의 RunE 함수에서 crypto, vault, keychain, recovery, claudemd 패키지를 조합

**구현 파일**:

| 파일 | 명령어 | 핵심 로직 |
|------|--------|----------|
| `root.go` | `tene` | rootCmd, 글로벌 플래그 (--json, --env), version |
| `init.go` | `tene init` | 패스워드 입력 -> KDF -> Master Key -> Keychain -> SQLite -> Recovery Key -> CLAUDE.md |
| `set.go` | `tene set KEY VALUE` | Keychain에서 Master Key -> Encrypt -> Vault.SetSecret |
| `get.go` | `tene get KEY` | Vault.GetSecret -> Decrypt -> stdout (또는 --json) |
| `run.go` | `tene run -- CMD` | 모든 시크릿 복호화 -> 환경변수 주입 -> exec.Command |
| `list.go` | `tene list` | Vault.ListSecrets -> 테이블 출력 (값 마스킹) |
| `delete.go` | `tene delete KEY` | 확인 프롬프트 -> Vault.DeleteSecret |
| `import_cmd.go` | `tene import` | .env 파싱 -> 각 키 암호화 -> Vault.SetSecret (배치) |
| `export.go` | `tene export` | 모든 시크릿 복호화 -> .env 형식 출력 / --encrypted: 볼트 파일 암호화 내보내기 |
| `env.go` | `tene env` | 환경 목록/전환/생성 |
| `passwd.go` | `tene passwd` | 현재 패스워드 확인 -> 새 패스워드 -> 볼트 재암호화 -> 새 Recovery Key |
| `recover.go` | `tene recover` | 12단어 입력 -> Master Key 복구 -> 새 패스워드 설정 -> 볼트 재암호화 |
| `sync_cmd.go` | `tene sync` | Fake Door: Cloud waitlist 안내 출력 |
| `whoami.go` | `tene whoami` | 현재 프로젝트 정보, 활성 환경, 시크릿 수 출력 |
| `helpers.go` | (공통) | 패스워드 입력 (golang.org/x/term), JSON 출력, 에러 포맷 |

**글로벌 플래그**:

| 플래그 | 타입 | 설명 |
|--------|------|------|
| `--json` | bool | JSON 형식 출력 (AI 에이전트 파싱용) |
| `--env` | string | 환경 지정 (default: 활성 환경) |
| `--dir` | string | 프로젝트 디렉토리 (default: 현재 디렉토리) |
| `--version` | bool | 버전 출력 |

### 3.8 cmd/tene/main.go -- 엔트리포인트

```go
package main

import (
    "os"
    "github.com/tomo-kay/tene/internal/cli"
)

// version, commit, date are set by goreleaser ldflags
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    cli.SetVersion(version, commit, date)
    if err := cli.Execute(); err != nil {
        os.Exit(1)
    }
}
```

---

## 4. 의존성 목록

### 4.1 런타임 의존성

| 패키지 | 버전 | 용도 |
|--------|------|------|
| `github.com/spf13/cobra` | v1.8+ | CLI 프레임워크 |
| `modernc.org/sqlite` | v1.34+ | 순수 Go SQLite (CGo 불필요) |
| `golang.org/x/crypto` | v0.31+ | Argon2id, XChaCha20-Poly1305, HKDF |
| `golang.org/x/term` | v0.27+ | 터미널 패스워드 입력 (에코 숨김) |
| `github.com/zalando/go-keyring` | v0.2+ | OS 키체인 (macOS/Linux/Windows) |
| `github.com/tyler-smith/go-bip39` | v1.1+ | BIP-39 니모닉 생성/검증 |

### 4.2 개발/테스트 의존성

| 패키지 | 용도 |
|--------|------|
| `github.com/stretchr/testify` | 테스트 어서션 (assert, require) |

### 4.3 빌드/도구 의존성

| 도구 | 용도 |
|------|------|
| `golangci-lint` | Go 린터 집합 |
| `goreleaser` | 멀티 플랫폼 빌드 + GitHub Releases + Homebrew tap |

---

## 5. 빌드/배포 계획

### 5.1 goreleaser 설정 (.goreleaser.yml)

```yaml
version: 2
project_name: tene

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
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - format: tar.gz
    name_template: "tene_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

brews:
  - repository:
      owner: tomo-kay
      name: homebrew-tap
    homepage: "https://tene.sh"
    description: "Agentic Secret Runtime - Local-first encrypted secret management"
    install: |
      bin.install "tene"

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"
```

### 5.2 Homebrew tap

```bash
# 사용자 설치
brew install tomo-kay/tap/tene

# tap 저장소: github.com/tomo-kay/homebrew-tap
# goreleaser가 자동으로 Formula 생성/업데이트
```

### 5.3 curl 설치 스크립트 (install.sh)

```bash
#!/bin/sh
# tene installer - https://tene.sh/install.sh
set -e

REPO="tomo-kay/tene"
INSTALL_DIR="/usr/local/bin"

# OS/아키텍처 감지
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
    x86_64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
esac

# 최신 릴리즈 다운로드
LATEST=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep tag_name | cut -d'"' -f4)
URL="https://github.com/${REPO}/releases/download/${LATEST}/tene_${OS}_${ARCH}.tar.gz"

echo "Installing tene ${LATEST}..."
curl -sL "$URL" | tar xz -C /tmp
sudo mv /tmp/tene "$INSTALL_DIR/tene"
echo "tene installed to ${INSTALL_DIR}/tene"
echo "Run 'tene init' to get started."
```

### 5.4 go install

```bash
go install github.com/tomo-kay/tene@latest
```

### 5.5 GitHub Actions CI (.github/workflows/ci.yml)

```yaml
name: CI
on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - uses: golangci/golangci-lint-action@v6

  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest]
        go: ["1.22"]
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}
      - run: go test -race -coverprofile=coverage.out ./...
      - run: go tool cover -func coverage.out

  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - run: CGO_ENABLED=0 go build -o tene ./cmd/tene
```

### 5.6 GitHub Actions Release (.github/workflows/release.yml)

```yaml
name: Release
on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - uses: goreleaser/goreleaser-action@v6
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## 6. 테스트 전략

### 6.1 테스트 계층

| 계층 | 범위 | 도구 | 커버리지 목표 |
|------|------|------|:------------:|
| **단위 테스트** | 패키지별 함수 | Go testing + testify | 90%+ |
| **통합 테스트** | CLI 명령어 E2E | Go testing (exec) | 주요 플로우 |
| **보안 테스트** | 암호화 라운드트립, 키 유출 방지 | 커스텀 테스트 | 100% (crypto) |

### 6.2 패키지별 테스트 파일

| 패키지 | 테스트 파일 | 핵심 시나리오 |
|--------|------------|--------------|
| `crypto` | `crypto_test.go` | 암복호화 라운드트립, 잘못된 키, AAD 변조, 대용량 |
| `recovery` | `mnemonic_test.go` | 니모닉 생성/검증, Master Key 복구 라운드트립 |
| `vault` | `vault_test.go` | CRUD, UPSERT, 환경 격리, 감사 로그, 마이그레이션 |
| `keychain` | `keychain_test.go` | Store/Load/Delete, 파일 폴백, 퍼미션 |
| `claudemd` | `generator_test.go` | 새 생성, 섹션 추가, 중복 스킵, 내용 보존 |
| `cli` | `cli_test.go` | init->set->get->list->delete E2E, import/export |

### 6.3 테스트 헬퍼

```go
// internal/vault/vault_test.go
func setupTestVault(t *testing.T) (*Vault, string) {
    t.Helper()
    dir := t.TempDir()
    dbPath := filepath.Join(dir, ".tene", "vault.db")
    os.MkdirAll(filepath.Dir(dbPath), 0700)
    v, err := New(dbPath)
    require.NoError(t, err)
    t.Cleanup(func() { v.Close() })
    return v, dir
}
```

### 6.4 테스트 실행

```bash
# 전체 테스트
make test
# = go test -race -v ./...

# 특정 패키지
go test -v ./internal/crypto/...

# 커버리지
make cover
# = go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

---

## 7. 구현 일정

### Week 1: 핵심 인프라 (Day 1-5)

| Day | 작업 | 산출물 |
|-----|------|--------|
| 1 | Go 프로젝트 초기화 (go.mod, 디렉토리, .gitignore, Makefile) | 빌드 가능한 빈 프로젝트 |
| 1-2 | internal/crypto 구현 + 테스트 | kdf.go, encrypt.go, decrypt.go, keymanager.go, 95%+ 커버리지 |
| 2-3 | internal/recovery 구현 + 테스트 | mnemonic.go, recover.go, 니모닉 라운드트립 |
| 3-4 | internal/vault 구현 + 테스트 | schema, CRUD, migration, 환경 격리 |
| 4-5 | internal/keychain 구현 + 테스트 | go-keyring 래퍼, 파일 폴백 |

### Week 2: CLI + 배포 (Day 6-10)

| Day | 작업 | 산출물 |
|-----|------|--------|
| 6 | internal/claudemd 구현 + 테스트 | CLAUDE.md 생성/병합 |
| 6-8 | internal/cli 핵심 명령어 (init, set, get, run, list, delete) | 6개 핵심 명령어 동작 |
| 8-9 | internal/cli 나머지 (import, export, env, passwd, recover, sync, whoami) | 7개 추가 명령어 |
| 9 | CLI 통합 테스트 | E2E 테스트 통과 |
| 10 | goreleaser + CI/CD + Homebrew tap + install.sh | 릴리즈 파이프라인 |

### 마일스톤

| 마일스톤 | 기준 | 예상 시점 |
|---------|------|----------|
| M1: crypto 완료 | 암복호화 테스트 100% 통과 | Day 2 |
| M2: vault 완료 | CRUD + 환경 격리 테스트 통과 | Day 4 |
| M3: init->set->get 동작 | 첫 E2E 플로우 성공 | Day 7 |
| M4: 13개 명령어 완료 | 모든 명령어 동작 | Day 9 |
| M5: v0.1.0 릴리즈 | goreleaser 빌드 + brew 설치 성공 | Day 10 |

---

## 8. 코딩 컨벤션

### 8.1 Go 네이밍

| 대상 | 규칙 | 예시 |
|------|------|------|
| 패키지명 | 소문자, 단수 | `crypto`, `vault`, `keychain` |
| 공개 함수/타입 | PascalCase | `DeriveKey()`, `Vault`, `Secret` |
| 비공개 함수/타입 | camelCase | `deriveSubKey()`, `initSchema()` |
| 상수 | PascalCase 또는 ALL_CAPS | `DefaultSaltLen`, `SchemaVersion` |
| 인터페이스 | -er 접미사 (단일 메서드) | `KeyStore` (다중 메서드이므로 예외) |
| 테스트 함수 | `Test_FunctionName_Scenario` | `Test_Encrypt_RoundTrip` |

### 8.2 에러 처리

```go
// 1. 패키지별 sentinel 에러 정의
var (
    ErrDecryptionFailed = errors.New("crypto: decryption failed")
    ErrInvalidKey       = errors.New("crypto: invalid key length")
)

// 2. 에러 래핑
func Decrypt(key, ciphertext, aad []byte) ([]byte, error) {
    result, ok := secretbox.Open(nil, ciphertext, &nonce, &key32)
    if !ok {
        return nil, fmt.Errorf("%w: authentication failed", ErrDecryptionFailed)
    }
    return result, nil
}

// 3. CLI에서 에러 출력
func handleError(err error) {
    if errors.Is(err, crypto.ErrDecryptionFailed) {
        fmt.Fprintln(os.Stderr, "Error: Wrong master password or corrupted data")
        os.Exit(1)
    }
}
```

### 8.3 파일 구조 규칙

- 파일당 하나의 주요 타입 또는 기능 그룹
- 파일 상단: package 선언 -> import -> 상수 -> 타입 -> 함수
- 공개 함수 먼저, 비공개 함수 나중에
- godoc 주석은 모든 공개 타입/함수에 필수

### 8.4 종료 코드

| 코드 | 의미 |
|:----:|------|
| 0 | 성공 |
| 1 | 일반 에러 |
| 2 | 인증 실패 (잘못된 패스워드) |
| 3 | 시크릿 미발견 |
| 4 | 볼트 미초기화 (`tene init` 필요) |

### 8.5 --json 출력 형식

```json
// 성공
{"status":"ok","data":{"key":"STRIPE_KEY","value":"sk_test_xxx"}}

// 에러
{"status":"error","code":3,"message":"secret not found: STRIPE_KEY"}

// 목록
{"status":"ok","data":{"secrets":[{"name":"STRIPE_KEY","environment":"default","updated_at":"2026-04-06T12:00:00Z"}]}}
```

---

## 9. Makefile

```makefile
.PHONY: build test lint cover clean

VERSION ?= dev
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o tene ./cmd/tene

test:
	go test -race -v ./...

lint:
	golangci-lint run

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func coverage.out
	go tool cover -html=coverage.out -o coverage.html

clean:
	rm -f tene coverage.out coverage.html

install: build
	sudo cp tene /usr/local/bin/tene
```

---

## 10. 리스크 및 완화

| 리스크 | 영향 | 완화 방안 |
|--------|------|----------|
| modernc.org/sqlite 성능 | CLI 응답 느림 | 벤치마크 테스트, 인덱스 최적화 |
| go-keyring 크로스 플랫폼 | 특정 OS에서 실패 | 파일 폴백 구현, CI에서 macOS + Linux 테스트 |
| XChaCha20-Poly1305 구현 오류 | 보안 치명적 | 벡터 테스트, 외부 도구로 교차 검증 |
| goreleaser 설정 오류 | 배포 실패 | 로컬에서 `goreleaser check` + `--snapshot` 테스트 |
| BIP-39 니모닉 엔트로피 | Recovery 불안정 | crypto/rand 사용 확인, 기존 테스트 벡터로 검증 |

---

## 11. 성공 기준

| 기준 | 측정 방법 |
|------|----------|
| 13개 CLI 명령어 정상 동작 | CLI 통합 테스트 통과 |
| 테스트 커버리지 90%+ | `go test -cover` |
| 오프라인 100% 동작 | 네트워크 차단 상태에서 모든 명령어 실행 |
| 바이너리 크기 < 20MB | `ls -lh tene` |
| CLI 시작 시간 < 50ms | `time tene --version` |
| macOS + Linux 빌드 성공 | goreleaser CI |
| `brew install tomo-kay/tap/tene` 성공 | 실제 설치 테스트 |
| CLAUDE.md 자동 생성 | `tene init` 후 CLAUDE.md 존재 확인 |
| 암복호화 라운드트립 100% | crypto 테스트 전수 통과 |
