# tene CLI Sprint Master Plan

> **작성일**: 2026-05-12
> **베이스 문서**: `docs/03-report/cli-completeness-audit-2026-05-11.md` (1,383 LOC, 11개 영역 / 60+ 발견)
> **분석 방법**: `/pdca pm` + `/pdca team` 병렬 7-agent 정독 (PM × 1 + Deep-read × 6)
> **분석 범위**: 비-테스트 라이브러리 LOC + 테스트 LOC + CI/CD 워크플로 + 문서 — 한 줄씩 정독, file:line cross-check
> **목표 산출물**: 6 스프린트, 약 13주, 11개 영역의 P0/P1을 모두 해소하는 실행 계획
> **언어**: 한국어 (코드 변경 명세는 영문 식별자 그대로 유지)

---

## 0. 개요 (Executive Summary)

### 0.1 한 줄 진단

**tene CLI v1.0.8은 핵심 기능 80%가 정상 작동하지만, 16/16 nil salt + AAD 부재로 인한 암호학적 결함, 0개의 fuzz 테스트, 매트릭스 빌드 없는 단일 OS CI, 광고와 다른 exit code 등 "OSS 신뢰도를 깎는 11종 부채"가 누적되어 있다.** 부채 해소 + macOS Touch ID/Windows Hello 생체 인증을 6 스프린트(13주)에 걸쳐 단계적으로 도입하면 v2.0 출시 + 1k stars 달성에 필요한 신뢰 토대를 확보할 수 있다.

### 0.2 분석 통계

| 영역 | 정독 LOC (non-test) | 정독 LOC (test) | 정독 파일 | 발견 (총/P0/P1/P2) |
|------|-------------------:|----------------:|---------:|:-----------------:|
| DR-1 CLI Surface (`cmd/`, `internal/cli/`) | ~1,800 | ~620 | 33 | 25 / 3 / 9 / 13 |
| DR-2 Vault + Sync (`internal/vault/`, `internal/sync/`) | 1,471 | 681 | 10 | 13 / 3 / 5 / 5 |
| DR-3 Crypto (`pkg/crypto/`) | ~900 | ~410 | 8 | 16 / 4 / 7 / 5 |
| DR-4 Errors + Domain + Config + claudemd + encfile | 937 | 618 | 16 | 13 / 1 / 7 / 5 |
| DR-5 Tests + CI + Lint | — | 2,763 | 22 + CI yml | 20 / 4 / 8 / 8 |
| DR-6 Distribution (Homebrew, SBOM, SLSA, cosign, MCP) | — | — | 빌드/배포 자료 | 12 / 0 / 5 / 7 |
| PM Analysis (페르소나, KPI, 리스크) | — | — | — | 4 페르소나 / 7 KPI / 8 리스크 / 8 외부 의존성 |
| **합계** | **~5,100** | **~5,090** | **~89 파일** | **~99 발견 / 15 P0 / 41 P1 / 43 P2** |

### 0.3 PM 핵심 결과 요약

#### 페르소나 (사용 시 의사결정의 기준)

| 페르소나 | 비중 추정 | 핵심 니즈 | 차단 요인 |
|---------|:---------:|----------|----------|
| **AI-vibe coder** (Cursor/Claude Code 사용자) | 50% | "secret이 LLM 컨텍스트로 새는 거 막아줘" | 현재 잘 충족 — `tene run --` 1줄 |
| **Indie OSS dev** | 25% | 로컬 + Git 친화 + 무료 | 현재 충족 — 단, SLSA/cosign 미흡으로 supply-chain 의구심 |
| **Sec-conscious team** | 15% | KAT 검증된 crypto, audit log, recovery | **차단**: KAT 0개, AAD 부재, sync untested |
| **Solo founder** | 10% | 다중 환경 + 빠른 import | 충족 |

#### 핵심 KPI (v2.0 출시 시점 목표)

| KPI | 현재 | v2.0 목표 |
|-----|------|----------|
| GitHub Stars | 0 (private repo) | 1,000+ |
| Weekly brew installs | 0 | 1,500 |
| OWASP A02 (Cryptographic Failures) 위반 | 4건 (P0) | 0 |
| Fuzz coverage | 0 targets | 8+ targets (KDF/decrypt/recovery/import) |
| CI matrix | 1 (ubuntu/1.25) | 6 (3 OS × 2 Go) |
| Unit test 커버리지 | 측정 안 됨 | 70%+ (게이트 통과) |
| Schema migration 메커니즘 | 없음 | 있음 (forward + rollback) |
| Cosign keyless signing | 없음 | 모든 release 서명 |

#### 8개 핵심 리스크 (P0 등급)

1. **R1**: 16/16 nil salt — 향후 AEAD nonce reuse 사고 발생 시 모든 vault 일괄 해킹 가능
2. **R2**: 마이그레이션 0건 — vault.db schema 변경 시 사용자 데이터 손실
3. **R3**: Sync engine 473 LOC × 0 테스트 — 정식 출시 시 데이터 손실/충돌
4. **R4**: matrix 없는 CI — macOS-only 사용자 환경에서 발생할 회귀 미감지
5. **R5**: STDOUT secret leak — `tene get $K` 사용 사고 시 history/로그/AI 컨텍스트 노출
6. **R6**: Exit code drift — 자동화 스크립트가 잘못된 분기 (P0)
7. **R7**: 생체 인증 미지원 — 1Password/Bitwarden 대비 UX 격차로 페르소나 1 이탈
8. **R8**: 단일 fail-point — go-keyring 의존 (Zalando lib 유지보수 정체 우려)

#### 외부 의존성 8건 (변경 위험 추적 필요)

| 패키지 | 현재 버전 | 위험 |
|--------|----------|------|
| github.com/spf13/cobra | v1.10.1 | 안정 — 호환성 우수 |
| github.com/zalando/go-keyring | v0.2.6 | **중간** — 유지보수 정체, fallback 로직 필수 |
| modernc.org/sqlite | v1.39.0 | **중간** — CGo-free 장점 but mattn-sqlite3 대비 느림 |
| golang.org/x/crypto | v0.43.0 | 안정 |
| github.com/tyler-smith/go-bip39 | v1.1.0 | 안정 — RFC 압축 |
| github.com/charmbracelet/lipgloss | v1.1.0 | 안정 |
| github.com/joho/godotenv | v1.5.1 | 안정 |
| Cobra → Charm/`huh` 또는 Bubble Tea | (미사용) | v3.0 TUI 모드 고려 시점 |

---

## 1. 발견사항 종합 매트릭스

영역별 P0/P1을 한 페이지에 모아두었다. 각 항목은 audit 보고서 §, 정독 보고서 §, 파일 file:line 으로 추적된다.

### 1.1 P0 (Critical) — Sprint 1에서 모두 처리

| ID | 영역 | 위치 | 한줄 요약 | Audit § |
|---:|------|------|---------|---------|
| **P0-S1** | Sync | `internal/sync/engine.go:Pull` + `merge.go` (244 LOC) | ThreeWayMerge 함수가 어디서도 호출되지 않음 — Pull은 unconditional overwrite | DR-2 신규 |
| **P0-S2** | Sync | `internal/cli/push.go:163` + `internal/sync/engine.go:445` | `saveSyncState`가 2곳에 다른 JSON 스키마로 정의 (race) | DR-2 신규 |
| **P0-V1** | Vault | `internal/vault/vault.go:audit_log` | `resource_name`과 `details`를 평문 SQLite에 저장 | DR-2 신규 |
| **P0-C1** | Crypto | `internal/vault/vault.go` 16개 호출 | `DeriveSubKey(rootKey, nil, "tene/...")` — salt가 모두 nil | DR-3, audit A3 |
| **P0-C2** | Crypto | `pkg/crypto/encrypt.go:Encrypt` | AEAD AAD 부재 (`aead.Seal(.., .., plaintext, nil)`) — context binding 없음 | DR-3, audit A3 |
| **P0-C3** | Crypto | `pkg/crypto/keymanager.go:RecoverySalt` | `salt[:8]` 8바이트 truncation — Argon2id RFC 권장 16바이트 위반 | DR-3 신규 |
| **P0-C4** | Crypto | `pkg/crypto/kdf.go` + `encfile.go` | `KDF_ALG_REGISTRY` 부재 — Argon2id 외 알고리즘 추가 시 호환성 깨짐 | DR-3, audit A1 |
| **P0-E1** | Errors | `pkg/errors/codes.go` ↔ `docs/cli-reference.md` | Exit code 3/4/5/6/7 광고하지만 실제로 emit 안 됨; `AUTH_REQUIRED` 코드 0개 | DR-4, audit A5 |
| **P0-T1** | Tests | `internal/sync/engine.go` (473 LOC) | 직접 테스트 0개 — Push/Pull/ExtractKeyMetadata 모두 untested | DR-5, audit A6 |
| **P0-T2** | Tests | `pkg/crypto/crypto_test.go` 전체 | KAT (Known Answer Tests) 0개 — RFC 8439/9106 벡터 없음 | DR-5, audit A6 |
| **P0-T3** | Tests | 전체 codebase | `Fuzz*` 0개 / `testing.F` 0개 — crypto/parser/decode 모두 fuzz 없음 | DR-5, audit A6 |
| **P0-CI1** | CI | `.github/workflows/ci.yml` | runs-on: ubuntu-latest only — macOS/Windows 회귀 미감지 | DR-5, audit A8 |
| **P0-G1** | Get | `internal/cli/get.go:U-1 guard` | NonTTY + JSON 분기에서 warning 미실현 — get_guard_test.go:96 confess | DR-5 N4 |
| **P0-R1** | Resource cleanup | `internal/sync/engine.go` HTTP requests | `defer resp.Body.Close()` 누락 가능성 (CI에 bodyclose 없음) | DR-5 N16 |
| **P0-K1** | Keychain | `internal/keychain/keychain.go` 실제 OS path | 실제 OS keychain 경로 untested — fallback만 테스트됨 | DR-5 §3.1 |

### 1.2 P1 (High) — Sprint 2~3에서 처리

| ID | 영역 | 위치 | 한줄 요약 |
|---:|------|------|---------|
| P1-C1 | Crypto | `pkg/crypto/zero.go` | `runtime.KeepAlive` 부재 — 컴파일러 최적화로 zero 제거 가능 |
| P1-C2 | Crypto | `pkg/crypto/x25519.go` | low-order point 검증 부재 (현재 사용 안 함 but team feature 활성화 전 필수) |
| P1-V1 | Vault | `internal/vault/schema.go` | 마이그레이션 메커니즘 부재 — `schema_migrations` 테이블 없음 |
| P1-V2 | Vault | `internal/vault/vault.go:82` | engine.New 와 vault.New 중복 — DR-2 A2 |
| P1-V3 | Vault | `internal/vault/vault.go:Open` PRAGMAs | synchronous/busy_timeout/cache_size/mmap_size/temp_store 미설정 |
| P1-S1 | Sync | `internal/sync/engine.go:407` | math/rand jitter — `// #nosec G404` 주석 필요 |
| P1-S2 | Sync | `internal/sync/conflict.go` | conflict_test.go 27 LOC × 3 funcs — barely smoke |
| P1-E1 | Errors | `pkg/errors` package name | stdlib `errors`와 충돌 — 13 사이트 모두 `teneerr` alias 사용 중 |
| P1-E2 | Errors | `pkg/errors/errors.go` | `Unwrap` 메서드 부재 — `errors.Is/As` chain 깨짐 |
| P1-E3 | Errors | `pkg/errors/errors.go:54` `IsTeneError` | raw type assertion — `errors.As` 미사용 |
| P1-E4 | Errors | `pkg/errors/codes.go:64` STDOUT_SECRET_BLOCKED Exit=2 | auth group(2)에 보안 정책 위반이 섞임 — 자동화 오해 위험 |
| P1-D1 | Domain | `pkg/domain.ErrVaultNotFound` ↔ `pkg/errors.ErrVaultNotFound` | 동일 이름 export — alias 충돌 |
| P1-D2 | Domain | `domain.SyncState` ↔ `config.SyncInfo` | snake_case vs camelCase 동일 5필드 스키마 |
| P1-Cfg1 | Config | `internal/config/config.go` | ~140 LOC의 90% dead code (Load/Save/CloudConfig consumer 0) |
| P1-Cfg2 | Config | `config_test.go:30,58,71,91` | `t.Setenv("HOME")` only — Windows `USERPROFILE` 누락 |
| P1-CMD1 | claudemd | `internal/cli/init.go:177,179` | `agentFiles, _ = gen.GenerateAll()` — 에러 swallow |
| P1-CMD2 | claudemd | `internal/claudemd/generator.go:110` | `HasTeneSection` 휴리스틱 false positive 폭탄 |
| P1-CMD3 | claudemd | `internal/claudemd/template.go` | "Quick Reference"에서 `tene get` 광고 vs Rules §8 "쓰지 말라" — LLM 입력에 모순 |
| P1-CMD4 | claudemd | `template.go` | version sentinel 부재 (`<!-- tene:v1 -->` 없음) — v2 업데이트 시 수동 |
| P1-Enc1 | Encfile | `encfile.go:163` | header.KDFMemory/Iterations/Parallel decode-only — 인자로 전달 안 됨 |
| P1-Enc2 | Encfile | `encfile.go:117,181` | `[]byte("tene-export")` 매직 리터럴 2번 |
| P1-Enc3 | Encfile | `encfile.go:74` KDFAlgorithm | byte 검증 분기 없음 |
| P1-Enc4 | Encfile | `encfile.go:17,20` | `FormatVersion`/`KDFAlgArgon2id`가 `var` (mutable) |
| P1-T1 | Tests | `internal/cli/testhelper_test.go:96` | `rootCmd` 글로벌 변경 + `os.Stdout` swap — `t.Parallel()` 추가 시 깨짐 |
| P1-T2 | Tests | `internal/cli/testhelper_test.go:39-79` | `resetFlags()` 17개 변수 손수 — 새 flag 누락 시 silent leak |
| P1-T3 | Tests | `internal/cli` 10/22 RunE untested | run/passwd/recover/update + cloud commands |
| P1-T4 | Tests | `pkg/domain` 전체 | 테스트 파일 0개 (DTO이지만 0%) |
| P1-T5 | Tests | `internal/keychain/keychain.go` 실제 OS path | 0 tests |
| P1-T6 | Tests | `internal/recovery/recover.go` | 0 tests (mnemonic.go만 있음) |
| P1-T7 | Tests | `pkg/crypto` 라이브러리 | 0 KAT (RFC 8439/9106 벡터) |
| P1-CI1 | CI | `.github/workflows/ci.yml` | govulncheck 없음 — crypto repo 표준 위반 |
| P1-CI2 | CI | `.github/workflows/ci.yml` | gosec 없음 — `math/rand` jitter 감지 못 함 |
| P1-CI3 | CI | `.golangci.yml` | unused linter 비활성 — 클라우드 사이드 명령 dead code 부채 |
| P1-CI4 | CI | `.golangci.yml` | exhaustive `default-signifies-exhaustive: true` — 글로벌 무력화 |
| P1-CI5 | CI | `.github/workflows/ci.yml` | 커버리지 게이트 없음 — `coverage.out` 생성만 |
| P1-Dist1 | Distribution | `.goreleaser.yaml` | SLSA L3 provenance 미생성 |
| P1-Dist2 | Distribution | `.goreleaser.yaml` | syft SBOM (SPDX) 미생성 |
| P1-Dist3 | Distribution | `.github/workflows/release.yml` | cosign keyless 서명 부재 |
| P1-Dist4 | Distribution | `homebrew-tene/tene.rb` | bottle 없음 — source build (느림) |
| P1-Dist5 | Biometric | 전체 codebase | macOS Touch ID / Windows Hello / Linux fprintd 미지원 |

### 1.3 P2 (Medium) — Sprint 4~6에서 처리 또는 백로그

(43건 — 상세 매트릭스는 audit 보고서 §1~§11 + 본 문서 부록 §A 참조)

대표 항목:
- P2-CLI1: `internal/cli/root.go:103-109` 클라우드 명령 주석 처리 — `//go:build cloud` 빌드 태그로 전환
- P2-CLI2: `internal/cli/run.go` 인자 파싱 — `--` 외 별칭 부재
- P2-Crypto1: HKDF info 문자열 (`"tene/audit"`, `"tene/sync"`)에 버전 부재 — 향후 도메인 분리 어려움
- P2-Vault1: `vault.db` 백업/복구 명령어 (`tene backup`, `tene restore`) 없음
- P2-Test1: 골든 파일 패턴 0건 — CLI 출력 회귀 테스트 비효율
- P2-Doc1: `--help` 텍스트에 `examples/` 경로 안내 없음

---

## 2. 의존성 그래프 (Dependency Graph)

작업 순서를 어기면 두 번 일을 해야 하는 핵심 의존성. 그래프는 위→아래.

```
                        ┌─────────────────────────┐
                        │ S1: Crypto Hotfix (P0)  │
                        │ (C1 nil salt + C2 AAD + │
                        │  C3 RecoverySalt)       │
                        └────────────┬────────────┘
                                     │ KDF info string 통일 + AAD 도입 후
                                     │ 모든 호출처 동시 마이그레이션 필요
                                     ▼
   ┌──────────────────┐  ┌─────────────────────────┐  ┌────────────────────┐
   │ S1: Vault v1.1   │  │ S1: Sync Engine 정리     │  │ S1: Exit Code 정렬 │
   │ (V1 audit_log    │  │ (S1 dead code 제거 +     │  │ (E1 docs 갱신)     │
   │  encrypt)        │  │  S2 saveSyncState 통합)  │  └────────────────────┘
   └────────┬─────────┘  └──────────┬──────────────┘
            │                       │
            │                       │ vault.db schema 변경 ↔ sync engine 동시 안정화
            ▼                       ▼
            ┌────────────────────────────────────┐
            │ S2: Vault v2 + Migration 메커니즘   │
            │ (P1-V1 schema_migrations 도입)     │
            │ ALTER vault_secrets + audit_log    │
            └────────────────┬───────────────────┘
                             │
                             │ schema 안정화 후
                             ▼
            ┌────────────────────────────────────┐
            │ S2: 테스트 인프라 재구축            │
            │ (T1 testhelper 리팩토 + T2          │
            │  fuzz target 8개 + T3 KAT 도입)    │
            └────────────────┬───────────────────┘
                             │
                             │ CI matrix 확장은 fuzz/KAT 안정화 후
                             ▼
            ┌────────────────────────────────────┐
            │ S3: CI/Lint 강화                   │
            │ (CI1 matrix + CI2 gosec/govulnck + │
            │  CI3 coverage gate)                │
            └────────────────┬───────────────────┘
                             │
       ┌─────────────────────┼─────────────────────┐
       ▼                     ▼                     ▼
   ┌────────────┐     ┌─────────────────┐    ┌──────────────────┐
   │ S4: 생체    │     │ S4: Errors/     │    │ S4: claudemd     │
   │ 인증 데코   │     │ Domain Refactor │    │ version sentinel │
   │ (P1-Dist5)  │     │ (P1-E1~E4)     │    │ + LLM 모순 제거   │
   └──────┬─────┘     └────────┬────────┘    └────────┬─────────┘
          │                    │                       │
          └──────────┬─────────┴──────────────┬────────┘
                     ▼                        ▼
            ┌────────────────────────────────────┐
            │ S5: 배포 강화                       │
            │ (Dist1 SLSA + Dist2 SBOM +         │
            │  Dist3 cosign + Dist4 bottle)      │
            └────────────────┬───────────────────┘
                             │
                             ▼
            ┌────────────────────────────────────┐
            │ S6: Polish + v2.0 출시             │
            │ (문서 + GeekNews/HN/Reddit 캠페인) │
            └────────────────────────────────────┘
```

### 2.1 Critical Path (가장 긴 순차 경로)

1. **S1 Crypto Hotfix** (2주) — AAD 도입 시 모든 평문 16개 호출 사이트 동시 마이그레이션 필수
2. **S2 Vault v2 + Migration** (2주) — schema_migrations 부재 → forward-only로 도입
3. **S2 테스트 인프라** (S2 후반 1주) — Vault v2 안정 후 KAT/fuzz 작성
4. **S3 CI 강화** (2주) — matrix + govulncheck + 커버리지 게이트
5. **S5 배포 강화** (2주) — SLSA + cosign + bottle
6. **S6 출시** (2주)

**총 critical path = 11주** (S4 생체 인증은 S3와 병렬 가능 → 13주 전체 일정)

### 2.2 병렬 실행 가능 작업

| Sprint | 동시 진행 가능 | 의존 없는 작업 |
|--------|---------------|---------------|
| S1 | 3 트랙 동시 | Crypto P0 / Sync 청소 / Exit code 문서 |
| S2 | 2 트랙 | Vault v2 / Errors+Domain 리팩토 |
| S3 | 3 트랙 | CI matrix / 테스트 추가 / 생체 인증 design spike |
| S4 | 3 트랙 | 생체 인증 구현 / Errors finalize / claudemd 개선 |
| S5 | 2 트랙 | 배포 강화 / 문서 작성 |
| S6 | 단일 (출시) | — |

### 2.3 차단 관계 (이 작업이 끝나기 전엔 시작 금지)

| 차단 (Blocker) | 차단 대상 (Blocked) | 이유 |
|---------------|--------------------|------|
| P0-C2 AAD 도입 | P1-V1 Vault v2 migration | AAD를 포함한 새 envelope format이 schema migration의 대상 |
| P0-C1 salt 통일 | P0-T2 KAT 작성 | salt 의 final form 결정 후 RFC 벡터로 검증 |
| P0-S1 dead code 제거 | P1-S2 sync 테스트 작성 | 실제 사용되는 path만 테스트 |
| P1-V1 migration 도입 | P1-CI1 CI matrix | 매트릭스 빌드 후 macOS/Windows에서 migration 검증 |
| P0-T1 sync 테스트 | P1-Dist5 생체 인증 | 인증 추가 전 기본 sync 동작 안정 |
| P1-E1 errors 패키지 rename | P1-D1 domain 이름 충돌 해결 | rename 동시 진행이 깔끔 |
| P0-E1 exit code | 자동화 사용자 안내 (changelog) | drift 해소 전 안내 금지 |

---

## 3. Sprint 구성 개요

| Sprint | 기간 | 목표 | 주요 산출물 | 출시 영향 |
|--------|------|------|------------|----------|
| **S1: Crypto + Sync Hotfix** | 2주 (W1-W2) | P0 15건 중 9건 해소 | v1.0.9-rcN (security hotfix) | minor patch |
| **S2: Architecture & Tests** | 2주 (W3-W4) | Vault v2 마이그레이션 + 테스트 인프라 + KAT | v1.1.0-rcN | minor |
| **S3: CI & Distribution Prep** | 2주 (W5-W6) | CI matrix + lint 강화 + 배포 인프라 점검 | v1.1.0 stable | minor stable |
| **S4: Biometric Auth & Polish** | 3주 (W7-W9) | 생체 인증 데코레이터 + Errors/Domain 리팩토 + claudemd v2 | v1.2.0-rcN | minor |
| **S5: Supply Chain Security** | 2주 (W10-W11) | SLSA L3 + SBOM + cosign + Homebrew bottle | v1.2.0 stable | minor stable |
| **S6: v2.0 Launch** | 2주 (W12-W13) | 문서 + 출시 캠페인 + Show HN | **v2.0.0 stable** | **major** |

**총 13주** (2026-05-13 시작 가정 시 2026-08-12 v2.0 출시)

---

## 4. Sprint 1 — Crypto + Sync Hotfix (W1-W2)

### 4.1 목표

P0 15건 중 보안 직결 9건을 1차 핫픽스. 사용자 영향이 큰 sync dead code 제거 + 평문 audit log 암호화 + crypto P0 4건 (nil salt / AAD / RecoverySalt / KDF registry).

### 4.2 작업 분해 (Trk = 트랙)

#### Trk-A: Crypto P0 핫픽스 (8 일, 2 dev)

| Task | 상세 | 변경 위치 | AC (Acceptance Criteria) |
|------|------|----------|-------------------------|
| **T1-A1** | `pkg/crypto/kdf.go`에 `DeriveSubKeyWithSaltV2(parent, salt, info)` 추가. **info 문자열 버전화** (`"tene/audit/v1"`, `"tene/sync/v1"`, `"tene/recovery/v1"` 등 13개 도메인) | `pkg/crypto/kdf.go` + 신규 `pkg/crypto/info.go` | 새 함수 시그니처 + KAT (RFC 5869 Appendix A) 3 벡터 통과 |
| **T1-A2** | `internal/vault/vault.go`의 16개 `DeriveSubKey(rootKey, nil, "tene/...")` 호출을 일괄 변경 → `DeriveSubKeyWithSaltV2(rootKey, vaultSalt, "tene/{domain}/v1")`. **단, 기존 vault.db는 v1 keyset 유지** (마이그레이션은 S2) | `internal/vault/vault.go` 16곳 + `internal/sync/envelope.go` 1곳 | grep `DeriveSubKey.*nil` = 0; 단위 테스트 일괄 통과 |
| **T1-A3** | `pkg/crypto/encrypt.go`에 `EncryptWithAAD(plaintext, key, aad []byte)` 추가. AAD = JSON of `{vault_id, env, key_name, version}`. 기존 `Encrypt`는 deprecated marker 유지 | `pkg/crypto/encrypt.go` + `decrypt.go` | AAD mismatch 시 `ErrDecryptFailed`; 6 RFC 8439 KAT 통과 |
| **T1-A4** | `RecoverySalt` 8바이트 truncation 제거 → 16바이트 고정. `pkg/crypto/keymanager.go` | `pkg/crypto/keymanager.go` | RFC 9106 권장 ≥ 16바이트 만족; 마이그레이션은 S2 |
| **T1-A5** | `pkg/crypto/kdf.go`에 `KDFAlgRegistry` 도입 — `Argon2idV1` (memory=64MB) + reserved slots (id 2 = `Argon2idV2`, id 3 = `Scrypt`, id 4 = `Bcrypt`). `encfile.go`가 알고리즘 byte 검증 | `pkg/crypto/kdf.go` + `internal/encfile/encfile.go` | 미지원 algorithm byte → `ErrUnsupportedKDF` (신규 exit 1 코드) |
| **T1-A6** | `pkg/crypto/zero.go`에 `runtime.KeepAlive` 보강 + `defer crypto.Zero(key)` 패턴 강제 (lint 룰) | `pkg/crypto/zero.go` + `.golangci.yml` revive rule | 모든 `key := DeriveKey(...)` 다음 line에 `defer Zero(key)` 존재 |
| **T1-A7** | `KAT` 디렉토리 신설 `pkg/crypto/testdata/` — 6 RFC 벡터 hex 고정 + roundtrip test | 신규 `testdata/rfc8439-1.json` 등 | 모든 KAT 통과 |
| **T1-A8** | Crypto P0 끝까지 정리한 후 `go test -race ./pkg/crypto/...` 200% 통과 (`-count=10`) | (검증) | flaky 0 |

**예상 LOC**: 비-테스트 +220 / 테스트 +480

#### Trk-B: Sync Dead Code 정리 (3 일, 1 dev)

| Task | 상세 | 변경 위치 | AC |
|------|------|----------|----|
| **T1-B1** | `merge.go` + `queue.go` (244 LOC) **완전 제거**. ThreeWayMerge는 caller 0이므로 dead. PR 메시지: "post-mortem: 244 LOC removed; sync only supports overwrite until v1.2" | `internal/sync/merge.go` 삭제, `queue.go` 삭제, `engine.go:Pull` 주석 갱신 | grep `ThreeWayMerge\|EnqueueOperation` = 0 |
| **T1-B2** | `saveSyncState`/`loadSyncState`를 `internal/sync/state.go` 단일 파일로 통합. JSON 스키마는 `internal/config/sync_info.go`의 camelCase 채택 (consumer 우선) | `internal/cli/push.go:163` 삭제, `internal/sync/engine.go:445` 삭제, 신규 `internal/sync/state.go` | grep `func saveSyncState` = 1; 스키마 race 해소 |
| **T1-B3** | `engine.go:407`의 `math/rand` jitter에 `// #nosec G404 -- jitter only` 주석 추가 (gosec 활성화 사전 작업) | `internal/sync/engine.go:407` | gosec 실행 시 1 issue 무시 |
| **T1-B4** | 모든 HTTP `Do()` 호출 사이트에 `defer resp.Body.Close()` 보장 (P0-R1) | `internal/sync/engine.go` 8 곳 | `bodyclose` lint pass (활성화는 S3) |

**예상 LOC**: 비-테스트 -244 / +30 / 테스트 ±0 (테스트는 S2)

#### Trk-C: Audit Log 암호화 + Exit Code 정렬 (4 일, 1 dev)

| Task | 상세 | 변경 위치 | AC |
|------|------|----------|----|
| **T1-C1** | `internal/vault/schema.go`의 `audit_log` 테이블에 v2용 컬럼 추가 (미사용 슬롯, S2에서 활성화): `resource_name_hmac BLOB`, `encrypted_details BLOB`. **v1 평문 컬럼은 유지** — 호환성 위해 deprecated만 표기 | `internal/vault/schema.go:CREATE TABLE audit_log` | ALTER 없이 새 vault만 v2 컬럼 보유 (마이그레이션은 S2) |
| **T1-C2** | `pkg/errors/codes.go` ↔ `docs/cli-reference.md` drift 해소 — **옵션 B (문서를 코드에 맞춤)**. 모든 exit code를 1 또는 2로 정정. `AUTH_REQUIRED` row 제거, `STDOUT_SECRET_BLOCKED`를 Exit 2 그룹 별도 row로 명시 | `docs/cli-reference.md:23-32` Exit Codes 테이블 | docs 와 codes.go 1:1 매핑 |
| **T1-C3** | `STDOUT_SECRET_BLOCKED` exit code를 Exit 2 → Exit 8 (별도) 분리. 외부 자동화가 "password mismatch"(2) 와 구분 가능 | `pkg/errors/codes.go:64-70` | exit 8 신규 코드; cli-reference.md 갱신 |
| **T1-C4** | `internal/cli/get.go:U-1` warning 미실현 버그 수정 (DR-5 N4) — JSON 모드에서도 stderr warning emit + 테스트로 검증 | `internal/cli/get.go:93-99` + `get_guard_test.go:82-104` | 테스트가 stderr 검사하도록 patch |

**예상 LOC**: 비-테스트 +40 / 문서 +80 / 테스트 +60

### 4.3 Sprint 1 검증

| Gate | 방법 | 통과 기준 |
|------|------|----------|
| G1 | `go test -race -count=10 ./pkg/crypto/...` | 0 flaky |
| G2 | `go test -race ./...` | 0 회귀 |
| G3 | `grep -rE "DeriveSubKey.*nil" pkg/ internal/" | 0 매치 |
| G4 | `grep -rE "ThreeWayMerge\|EnqueueOperation" .` | 0 매치 |
| G5 | `grep -c "func saveSyncState" .` | 1 매치 |
| G6 | docs/cli-reference.md ↔ codes.go diff 검증 스크립트 | 1:1 매핑 |
| G7 | `staging` 머지 후 auto-tag → `v1.0.9-rc1` 생성 | 워크플로 성공 |
| G8 | 6 RFC KAT 모두 hex 일치 | 통과 |

### 4.4 Sprint 1 리스크

| 리스크 | 완화 |
|-------|------|
| 16개 `DeriveSubKey` 호출 일괄 변경 시 빠뜨림 → 일부 keyset 가 nil salt 잔존 | T1-A2 끝나면 `grep nil pkg/crypto.*` 0 매치 강제 + integration test |
| AAD 도입으로 기존 v1 vault.db 호환성 깨질 우려 | `Encrypt` (구) + `EncryptWithAAD` (신) 병존; S1에서는 신규 vault만 새 envelope 사용; **마이그레이션 게이트는 S2** |
| Sync dead code 제거가 향후 merge UX 복원 시 cost ↑ | 삭제 PR에 "merge UX는 v1.2에서 새 architecture로 재도입" 명시 + S4 design spike 예약 |
| `STDOUT_SECRET_BLOCKED` exit code 변경(2→8)이 breaking change | CHANGELOG에 BREAKING + 자동화 migration guide 1페이지 |

---

## 5. Sprint 2 — Architecture & Tests (W3-W4)

### 5.1 목표

Vault v2 마이그레이션 메커니즘 도입 + 테스트 인프라 재구축 (testhelper 리팩토 + KAT/fuzz 8 타깃). DR-2 sync engine 리팩토 blueprint를 적용해 `contracts.go` 도입.

### 5.2 작업 분해

#### Trk-A: Vault v2 + Schema Migration (6 일, 2 dev)

| Task | 상세 | AC |
|------|------|----|
| **T2-A1** | `internal/vault/migrations/` 디렉토리 신설 + `schema_migrations(id TEXT PRIMARY KEY, applied_at INTEGER)` 테이블 도입 | `vault.New` 첫 호출 시 v1→v2 자동 |
| **T2-A2** | Migration `001_v2_envelope.sql`: `ALTER TABLE vault_secrets ADD COLUMN aad_kind TEXT DEFAULT 'v1'`, `ADD COLUMN description TEXT`, `ADD COLUMN last_read_at INTEGER`, `ADD COLUMN access_count INTEGER DEFAULT 0`, `ADD COLUMN deleted_at INTEGER` | 새 컬럼 모두 NULL 가능; v1 데이터 손상 0 |
| **T2-A3** | Migration `002_audit_log_encrypted.sql`: `ALTER TABLE audit_log ADD COLUMN resource_name_hmac BLOB`, `ADD COLUMN encrypted_details BLOB`. **v1 평문 컬럼 deprecated only** — 사용자가 `tene migrate audit` 명령으로 backfill | 평문 컬럼 → 암호화 컬럼 backfill 명령 |
| **T2-A4** | DR-2 blueprint: `internal/sync/contracts.go` 신설. `MetadataProvider`/`VaultReader`/`Transport`/`StateStore`/`VaultIO` 5개 인터페이스 분리 | `Engine`/`EngineOption` 패턴 + DI |
| **T2-A5** | `VaultMetadataProvider`가 `*vault.Vault`를 wrap — A2 P1 (engine.go:82) 중복 open 해소 | `vault.New` 1회만 호출 |
| **T2-A6** | PRAGMAs 보강 (P1-V3): `synchronous=NORMAL`, `busy_timeout=5000`, `cache_size=-2000` (2MB), `mmap_size=268435456` (256MB), `temp_store=MEMORY` | `vault.Open` 시 PRAGMA 6개 emit |
| **T2-A7** | Rollback 시나리오 문서화 — `docs/runbooks/vault-rollback.md` (forward-only migration이므로 백업 복원 가이드만) | runbook 작성 |

**예상 LOC**: 비-테스트 +320 / 테스트 +180

#### Trk-B: 테스트 인프라 재구축 (8 일, 2 dev)

| Task | 상세 | AC |
|------|------|----|
| **T2-B1** | `internal/cli/testhelper_test.go` 리팩토 — `rootCmd` 글로벌 변경 제거. 각 테스트가 자체 `*cobra.Command` 인스턴스 생성. `os.Stdout` swap 제거 → `cmd.SetOut(&buf)` 사용 | grep `os.Stdout = ` 테스트 0개; `t.Parallel()` 추가 가능 상태 |
| **T2-B2** | `resetFlags()` (DR-5 N2) 제거 — viper 또는 per-command flag binding으로 글로벌 flag 제거 | grep `resetFlags` = 0 |
| **T2-B3** | **Fuzz target 8개 추가** (P0-T3) — `Fuzz_KDF_DeriveKey`, `Fuzz_Encrypt_Roundtrip`, `Fuzz_Decrypt_TamperedNonce`, `Fuzz_Decrypt_TamperedAAD`, `Fuzz_Mnemonic_Decode`, `Fuzz_EncfileHeader_Decode`, `Fuzz_DotenvParse`, `Fuzz_VaultJSON_Decode` | `find . -name '*_test.go' -exec grep -l '^func Fuzz' {}` 8개 매치 |
| **T2-B4** | **KAT 추가 (P0-T2)** — RFC 8439 XChaCha20-Poly1305 6 벡터 + RFC 9106 Argon2id 3 벡터 + RFC 5869 HKDF 3 벡터 + BIP39 표준 24 벡터 | 36 KAT 통과 |
| **T2-B5** | `internal/sync/engine_test.go` 신설 — `Push`/`Pull`/`ExtractKeyMetadata` 단위 테스트 (mock Transport) | engine 473 LOC 커버리지 ≥ 70% |
| **T2-B6** | `internal/cli/run_test.go` 신설 — `tene run -- printenv $K` subprocess 검증 (T-1 P0 untested RunE) | env 정상 주입 + 빠진 KEY는 `os.Unsetenv` |
| **T2-B7** | `internal/cli/passwd_test.go` 신설 — 비밀번호 변경 후 모든 시크릿 재암호화 검증 (re-encryption 한 줄 빠지면 vault corruption) | 100% 시크릿 재암호화 검증 |
| **T2-B8** | `pkg/domain` 테스트 파일 신설 — Marshal/Unmarshal roundtrip (P1-T4) | 4 파일 모두 테스트 존재 |
| **T2-B9** | `internal/recovery/recover_test.go` 신설 — mnemonic → master key 재계산 (P1-T6) | recover.go RunE 단위 테스트 |
| **T2-B10** | `internal/keychain/keychain_test.go` macOS-only 통합 테스트 (`//go:build darwin`) — 실제 OS keychain 사용 | macOS CI에서 통과 |

**예상 LOC**: 테스트 +1,800

#### Trk-C: Errors + Domain Rename 준비 (4 일, 1 dev)

| Task | 상세 | AC |
|------|------|----|
| **T2-C1** | `pkg/errors` → `pkg/teneerr` 디렉토리 이동 + package 선언 변경 (DR-4 §3) — 13 import 사이트 모두 alias 이미 사용 중이므로 1-shot edit | `go build ./...` 통과 |
| **T2-C2** | `TeneError.Unwrap()` 추가 (P1-E2) + `IsTeneError`를 `errors.As` 기반으로 변경 (P1-E3) | `errors.As(wrapErr, &te)` 통과 |
| **T2-C3** | `pkg/domain.ErrVaultNotFound` → `ErrVaultMissing` rename (P1-D1) | grep `ErrVaultNotFound` only `teneerr` 패키지 |
| **T2-C4** | `pkg/domain.SyncState` 삭제 + `internal/config.SyncInfo` 단일 채택 (P1-D2) | grep `SyncState` = 0 |

**예상 LOC**: 비-테스트 -120 / +30 / 테스트 +40

### 5.3 Sprint 2 검증

| Gate | 통과 기준 |
|------|----------|
| G1 | v1 → v2 migration 자동 실행 검증 (3개 시드 vault) |
| G2 | `go test -count=10 -race ./...` 0 회귀 |
| G3 | Fuzz target 8개 각각 60초 실행 0 crash |
| G4 | 36 KAT hex 일치 |
| G5 | sync engine 커버리지 ≥ 70% |
| G6 | `go list -deps ./... \| grep teneerr` only |
| G7 | macOS CI에서 keychain integration test 통과 (CI matrix는 S3에서) |

### 5.4 Sprint 2 리스크

| 리스크 | 완화 |
|-------|------|
| Schema migration 실패 시 vault.db corruption | 자동 백업 `vault.db.pre-v2-{timestamp}` 생성 + rollback 명령 |
| Fuzz가 기존 코드의 panic 발견 → S2 마감 못 함 | fuzz 발견 panic은 S1 hotfix 후속으로 분리 (v1.0.10) |
| testhelper 리팩토가 22개 테스트 깨뜨림 | 점진적 (subset 단위) 리팩토 + per-test 검증 |

---

## 6. Sprint 3 — CI & Distribution Prep (W5-W6)

### 6.1 목표

DR-5 CI 강화 + lint 강화 + 배포 인프라 사전 점검 (Homebrew bottle, SLSA pipeline 검토).

### 6.2 작업 분해

#### Trk-A: CI Matrix + 신규 워크플로 (5 일, 1 dev)

| Task | 상세 | AC |
|------|------|----|
| **T3-A1** | `.github/workflows/ci.yml` matrix 확장 — `os: [ubuntu-latest, macos-latest, windows-latest]` × `go: ["1.24", "1.25"]` (P0-CI1) | 6 셀 모두 green |
| **T3-A2** | 신규 `govulncheck` job 추가 (P1-CI1) | `govulncheck ./...` clean |
| **T3-A3** | 신규 `fuzz-smoke` job — 8 fuzz target 각각 30초 실행 | nightly로는 5분 |
| **T3-A4** | 커버리지 게이트 도입 — `.testcoverage.yml` (total ≥ 70%, package ≥ 60%) | 빌드 fail if 미만 |
| **T3-A5** | Codecov 업로드 step | dashboard 노출 |

#### Trk-B: Lint 강화 (3 일, 1 dev)

| Task | 상세 | AC |
|------|------|----|
| **T3-B1** | `.golangci.yml` 강화 (DR-5 §5): `gosec` (G404 except), `errorlint`, `bodyclose`, `unparam`, `revive`, `gocritic`, `prealloc`, `tparallel`, `paralleltest`, `testifylint`, `gci`, `gofumpt`, `nolintlint` 추가 | 모든 lint 통과 |
| **T3-B2** | `unused` linter 재활성 — 클라우드 명령(`push.go`, `pull.go`, `sync_cmd.go`, `login.go`, `billing.go`, `team.go`)을 `//go:build cloud` 빌드 태그로 전환 | grep `^//go:build cloud` ≥ 6 파일; default build에서 unused 0 |
| **T3-B3** | `exhaustive: default-signifies-exhaustive: false` (DR-5 N13) — switch 분기 완전성 강제 | 모든 enum switch에 `default:` 외 분기 |

#### Trk-C: 배포 인프라 사전 점검 (4 일, 1 dev)

| Task | 상세 | AC |
|------|------|----|
| **T3-C1** | `.goreleaser.yaml`에 SLSA L3 provenance 생성 step (`sigstore/slsa-github-generator`) — actual sign은 S5 | dry-run 통과 |
| **T3-C2** | `syft` SBOM (SPDX) 생성 step 추가 | release asset에 `tene_{version}_sbom.spdx.json` 포함 |
| **T3-C3** | `homebrew-tene/Formula/tene.rb`에 bottle DSL 추가 + Github Actions에서 darwin-arm64/darwin-x86_64 bottle 빌드 | Homebrew 사용자 setup 시간 < 5초 |
| **T3-C4** | `auto-tag.yml` 이드포턴시 보강 (DR-5 N17) — `gh api refs/tags/X` 사전 확인 후 conflict 시 skip | 동일 SHA에 push x2 시 워크플로 idempotent |
| **T3-C5** | `auto-tag.yml`의 `LATEST_VERSION` guard 검증 범위 확장 (DR-5 N18) — darwin/arm64 + darwin/x86_64 + linux/amd64 + windows/amd64 4-OS 검증 | LATEST_VERSION flip 시 4개 모두 head-object 200 |

**예상 LOC**: CI yml +120 / lint +60 / goreleaser +40

#### Trk-D: 생체 인증 설계 스파이크 (병렬, 5 일, 1 dev)

| Task | 상세 | AC |
|------|------|----|
| **T3-D1** | `docs/02-design/biometric-auth.md` 작성 — 9개 OS×하드웨어 시나리오 매트릭스, BiometricKeyStore 인터페이스, Secure Enclave/TPM 추상화 | 디자인 리뷰 통과 |
| **T3-D2** | `internal/keychain/biometric/` 디렉토리 구조 prototype — darwin (LocalAuthentication), windows (CNG NCryptCreatePersistedKey), linux (TPM2) interface 정의만 | `go build` (구현 빈 함수) |
| **T3-D3** | macOS Touch ID PoC — 50 LOC Go binding (CGo) — `SecAccessControlCreateWithFlags(kSecAttrAccessControlBiometryCurrentSet)` | Touch ID 프롬프트 1번 표시 |

### 6.3 Sprint 3 검증

| Gate | 통과 기준 |
|------|----------|
| G1 | 6 CI 셀 모두 통과 |
| G2 | `govulncheck` clean |
| G3 | `golangci-lint run` clean (15+ linter) |
| G4 | 커버리지 ≥ 70% |
| G5 | bottle 빌드 자동화 + 로컬 검증 |
| G6 | Touch ID PoC 동작 |
| G7 | `v1.1.0` stable 발행 |

### 6.4 Sprint 3 리스크

| 리스크 | 완화 |
|-------|------|
| Windows CI에서 fork-bomb 또는 PATH 이슈 | 별도 windows-specific test exclusion (`//go:build !windows`) — 처음엔 보수적으로 |
| gosec/govulncheck 활성화 시 다수 신규 issue → S3 마감 못 함 | 첫 PR은 `//nolint:gosec` whitelist + ticket으로 후속 처리 |
| Homebrew bottle 빌드 권한 (GitHub Actions에서 brew install) | macos-latest runner 사용; brew 사전 캐싱 |

---

## 7. Sprint 4 — Biometric Auth & Polish (W7-W9)

### 7.1 목표

생체 인증 구현 (macOS Touch ID + Windows Hello + Linux fprintd) + Errors/Domain rename 완료 + claudemd v2 (version sentinel + LLM 모순 제거).

### 7.2 작업 분해

#### Trk-A: Biometric Auth 구현 (12 일, 2 dev)

| Task | 상세 | AC |
|------|------|----|
| **T4-A1** | `internal/keychain/biometric/contracts.go` — `BiometricStore` 인터페이스: `IsSupported() bool`, `Enroll(masterKey []byte) error`, `Unlock() ([]byte, error)`, `Revoke() error` | 4 메서드 정의 |
| **T4-A2** | `internal/keychain/biometric/darwin.go` (`//go:build darwin`) — `LAContext.evaluatePolicy` + `SecAccessControl(BiometryCurrentSet)` + Secure Enclave non-exportable key | Touch ID 성공 시 unlock |
| **T4-A3** | `internal/keychain/biometric/windows.go` (`//go:build windows`) — `NCryptCreatePersistedKey` + `NCRYPT_REQUIRE_HARDWARE_FLAG` (TPM-backed) | Windows Hello 성공 시 unlock |
| **T4-A4** | `internal/keychain/biometric/linux.go` (`//go:build linux`) — fprintd D-Bus + go-tpm-tools (optional) | fprintd 인증 시 unlock; TPM 없는 환경은 fail-soft → master password |
| **T4-A5** | `internal/keychain/biometric/none.go` (`//go:build !darwin && !windows && !linux`) — `IsSupported() = false` 항상 반환 | freebsd/openbsd 등에서 빌드 통과 |
| **T4-A6** | **데코레이터 패턴**: `BiometricKeyStore` 가 기존 `KeyStore`를 wrap. 생체 인증 실패 또는 미지원 → **자동 fallback** 으로 master password 프롬프트 | 9 OS×하드웨어 시나리오 매트릭스 통과 |
| **T4-A7** | CLI 통합: `tene biometric enroll` / `tene biometric revoke` / `tene biometric status` + `tene unlock` 명령에서 자동 사용 | 4 신규 RunE + 테스트 |
| **T4-A8** | 사용자 가이드 — `docs/guides/biometric-auth.md` (9 시나리오 매트릭스 포함) | 가이드 발행 |
| **T4-A9** | 모든 시나리오 무결성 보장 — master password가 모든 9 시나리오에서 항상 valid (fallback path) | integration test |

**9 OS×하드웨어 시나리오 매트릭스**:

| 시나리오 | 하드웨어 | 동작 |
|---------|---------|------|
| macOS + Touch ID 있음 | Secure Enclave | Touch ID 우선, 미인증 시 master password |
| macOS + Touch ID 없음 (Intel Mac mini 등) | — | master password only |
| macOS + Apple Silicon + Touch ID 없는 외장 키보드 | Secure Enclave 사용 가능 | master password (T1 fallback) |
| Windows + Hello + TPM 2.0 | TPM | Hello 우선, 미인증 시 master password |
| Windows + Hello 없음 (older) | TPM (옵션) | master password |
| Windows + Hello + TPM 없음 (가상머신) | — | master password |
| Linux + fprintd | 지문 센서 | fprintd 우선, 미인증 시 master password |
| Linux + TPM 2.0 + fprintd 없음 | TPM | TPM-bound key (옵션), 미사용 시 master password |
| Linux + 둘 다 없음 (서버 등) | — | master password only |

**예상 LOC**: 비-테스트 +800 / 테스트 +400

#### Trk-B: Errors finalize + claudemd v2 (4 일, 1 dev)

| Task | 상세 | AC |
|------|------|----|
| **T4-B1** | `pkg/teneerr` rename 완료 (S2 시작) — 모든 import 사이트 grep 0 매치 (legacy `errors` import) | 완료 |
| **T4-B2** | `internal/claudemd/template.go` version sentinel 도입 — `<!-- tene:v2 -->` 시작 마커 + `<!-- /tene:v2 -->` 종료 마커. `HasTeneSection` 휴리스틱 폐기 → strict marker match | 다음 init 시 v1 섹션 자동 교체 가능 |
| **T4-B3** | "Quick Reference" 표에서 `tene get <KEY>` row 제거 (DR-4 N-9) — LLM 모순 해소 | LLM 컨텍스트 평가 점수 ↑ |
| **T4-B4** | `init.go:177,179` 에러 swallow 수정 (P1-CMD1) — `if err != nil { warn(err) }` | 에러 발생 시 stderr |
| **T4-B5** | `pkg/domain` JSON tag 통일 (P2-D...) — `vault_version` → `version` (또는 vice versa, 한 쪽 선택) | grep tag inconsistency 0 |

**예상 LOC**: 비-테스트 +120 / 문서 +60

#### Trk-C: Config 슬림화 + Encfile 강화 (3 일, 1 dev)

| Task | 상세 | AC |
|------|------|----|
| **T4-C1** | `internal/config` dead code 제거 (P1-Cfg1) — Load/Save/CloudConfig consumer 0 인 ~140 LOC 삭제. `EnsureConfigDir` 만 유지 | LOC -140 |
| **T4-C2** | `config_test.go` Windows 호환 (P1-Cfg2) — `USERPROFILE` env 추가 | Windows CI 통과 |
| **T4-C3** | `encfile.go:163` KDF 파라미터 실제 전달 (P1-Enc1) — `DeriveKey`가 메모리/iterations/parallelism 인자로 받도록 시그니처 변경 | header 값 반영됨 |
| **T4-C4** | `encfile.go:117,181` AAD literal 통합 (P1-Enc2) — `const aadExport = "tene-export"` | grep `"tene-export"` 1매치 (const) |
| **T4-C5** | `encfile.go:17,20` `var` → `const` (P1-Enc4) | grep `^var FormatVersion` 0 |

**예상 LOC**: 비-테스트 -180 / +60 / 테스트 +80

### 7.3 Sprint 4 검증

| Gate | 통과 기준 |
|------|----------|
| G1 | 9 OS×하드웨어 시나리오 매트릭스 통과 (자동화는 macOS/Linux/Windows 3개 + 수동 6개) |
| G2 | master password가 항상 valid (fallback invariant) — 100 회 랜덤화 테스트 |
| G3 | claudemd v2 sentinel 인식 + v1 → v2 자동 마이그레이션 |
| G4 | `pkg/teneerr` rename 완료 — legacy import 0 |
| G5 | `v1.2.0-rcN` 발행 |

### 7.4 Sprint 4 리스크

| 리스크 | 완화 |
|-------|------|
| Touch ID CGo binding이 macOS 14 + Xcode 15 의존성 충돌 | macos-13 + macos-14 CI 매트릭스에서 검증; Xcode 14.x도 빌드 |
| Windows Hello가 TPM 미장착 가상머신에서 false negative | T4-A3에서 `NCRYPT_REQUIRE_HARDWARE_FLAG` 실패 시 자동 fallback |
| Linux fprintd가 distro마다 동작 다름 (Ubuntu/Fedora/Arch) | Ubuntu LTS만 1차 지원; 나머지 docs 표기 |
| 데코레이터 패턴이 기존 keychain 코드와 충돌 | T4-A6에서 `KeyStore` 인터페이스 추출 후 decorator |

---

## 8. Sprint 5 — Supply Chain Security (W10-W11)

### 8.1 목표

SLSA L3 provenance + SBOM 자동 생성 + cosign keyless 서명 + Homebrew bottle CDN 배포.

### 8.2 작업 분해

#### Trk-A: SLSA + SBOM + Cosign (8 일, 2 dev)

| Task | 상세 | AC |
|------|------|----|
| **T5-A1** | `.github/workflows/release.yml` → `slsa-github-generator/.github/workflows/generator_generic_slsa3.yml` 사용 | release asset에 `tene_{version}.intoto.jsonl` |
| **T5-A2** | `syft` SBOM 생성 — SPDX + CycloneDX 두 형식 | release asset에 SBOM 포함 |
| **T5-A3** | `sigstore/cosign-installer` + keyless sign — OIDC token 기반 | 모든 binary와 SBOM이 서명됨 |
| **T5-A4** | Cosign 검증 가이드 `docs/guides/verify-signature.md` 작성 | 사용자 검증 1-line |
| **T5-A5** | `tene update` 명령에서 cosign signature 자동 검증 (옵션) | `--verify=true` 시 검증 후 install |

#### Trk-B: Homebrew Bottle (3 일, 1 dev)

| Task | 상세 | AC |
|------|------|----|
| **T5-B1** | `homebrew-tene/Formula/tene.rb`의 `bottle do ... end` 블록 추가 + sha256 자동화 | brew install 시 source build 안 함 |
| **T5-B2** | bottle CDN 호스팅 — S3 `tene-releases/bottles/` + cloudfront | bottle url 정상 |
| **T5-B3** | `release.yml`이 darwin-arm64 + darwin-x86_64 bottle 자동 빌드 후 S3 업로드 + Formula PR | 매 release마다 bottle |

#### Trk-C: 추가 보안 강화 (3 일, 1 dev)

| Task | 상세 | AC |
|------|------|----|
| **T5-C1** | `tene` binary의 entitlements (macOS) — codesign + notarization | Gatekeeper bypass 없이 실행 |
| **T5-C2** | Windows Authenticode 서명 (Sectigo or DigiCert) — 비용 검토 후 S6 이전 결정 | 미서명 시 SmartScreen 경고 표시 |
| **T5-C3** | Reproducible build 검증 — 동일 commit으로 2회 빌드 시 byte-identical | sha256 일치 |

**예상 LOC**: workflow yml +250 / docs +120

### 8.3 Sprint 5 검증

| Gate | 통과 기준 |
|------|----------|
| G1 | `slsa-verifier` 외부 도구로 provenance 검증 통과 |
| G2 | `cosign verify-blob` 통과 (모든 asset) |
| G3 | Homebrew install 시간 < 5초 (bottle) |
| G4 | macOS Gatekeeper 경고 없음 |
| G5 | `v1.2.0` stable 발행 |

### 8.4 Sprint 5 리스크

| 리스크 | 완화 |
|-------|------|
| Apple Developer ID 비용 ($99/yr) | tomo-kay 개인 계정으로 진행 |
| Windows code signing 인증서 비용 ($300+/yr) | EV cert 없이 standard로 시작; SmartScreen 경고는 docs로 안내 |
| SLSA verifier가 GoReleaser와 호환성 이슈 | 공식 example 추종 + dry-run 검증 |

---

## 9. Sprint 6 — v2.0 Launch (W12-W13)

### 9.1 목표

`v2.0.0` major 릴리스 + 출시 캠페인 (Show HN, Daily.dev, GeekNews, Reddit) + 문서 완성.

### 9.2 작업 분해

#### Trk-A: 문서 완성 (5 일, 1 dev)

| Task | 상세 |
|------|------|
| **T6-A1** | `docs/migration/v1-to-v2.md` 작성 — Exit code 변경, Vault v2 자동 마이그레이션, 생체 인증 enroll 안내 |
| **T6-A2** | `docs/reference/cli.md` 전면 갱신 — 모든 RunE의 옵션 + 예시 |
| **T6-A3** | `docs/concepts/threat-model.md` 작성 — AEAD AAD, salt, KAT, fuzz 결과를 사용자 친화적으로 |
| **T6-A4** | `docs/guides/biometric-auth.md` 보강 — 9 시나리오 매트릭스 |
| **T6-A5** | `README.md` 갱신 — "tene v2.0 is here" + 핵심 기능 강조 |
| **T6-A6** | `CHANGELOG.md` v2.0.0 entry — BREAKING/feat/fix 카테고리 정리 |

#### Trk-B: 출시 캠페인 (5 일, 1 dev, growth-routine 룰 준수)

| Task | 상세 |
|------|------|
| **T6-B1** | Show HN 발사 (화요일 9am ET = 수요일 22:00 KST) — `Show HN: tene v2.0 — local-first secret manager with biometric auth` |
| **T6-B2** | Daily.dev Squad 포스팅 — v2.0 release note |
| **T6-B3** | GeekNews (news.hada.io) — 한국어 요약 |
| **T6-B4** | Reddit r/cursor + r/ClaudeAI + r/vibecoding — 인증 데모 GIF |
| **T6-B5** | Dev.to / Hashnode 블로그 — "How tene v2.0 closes 4 OWASP A02 gaps" |
| **T6-B6** | Twitter / LinkedIn — 1주일 후 thread (HN과 시간 간격) |

#### Trk-C: 출시 후 모니터링 (계속)

| Task | 상세 |
|------|------|
| **T6-C1** | `/tene-stats` daily 실행 — GitHub stats 누적 |
| **T6-C2** | Sentry/Crashlytics 추가 (옵션) — runtime panic 수집 |
| **T6-C3** | GitHub Issues triage — 24시간 내 응답 |

### 9.3 Sprint 6 검증

| Gate | 통과 기준 |
|------|----------|
| G1 | `v2.0.0` 태그 생성 + GitHub Verified badge |
| G2 | Homebrew Formula PR 머지됨 (`brew install tene`) |
| G3 | brew analytics +50 in 24h |
| G4 | Show HN 100+ points (front page 진입) |
| G5 | GitHub Stars +200 in 1week |

---

## 10. 종합 KPI 및 완료 기준

### 10.1 v2.0 출시 완료 기준 (DoD)

| 영역 | 완료 기준 |
|------|----------|
| **Security** | nil salt 0; AAD 모든 envelope; KAT 36개 통과; fuzz 8 target × 60초 0 crash; govulncheck clean |
| **Architecture** | Vault v2 schema migration 자동; sync engine 70% 커버; pkg/teneerr rename 완료 |
| **Testing** | 70% line coverage; CI matrix 6 cells green; race -count=10 통과 |
| **Distribution** | SLSA L3 provenance; cosign 서명; Homebrew bottle (< 5초 install); SBOM 첨부 |
| **Biometric** | macOS Touch ID + Windows Hello + Linux fprintd 지원; 9 시나리오 매트릭스 통과; master password fallback 보장 |
| **Documentation** | migration guide v1→v2; CLI reference 전면 갱신; threat model 문서 |
| **Launch** | Show HN 100+ points; Homebrew Formula 머지; v2.0.0 stable 태그 |

### 10.2 출시 후 30일 KPI

| KPI | 목표 |
|-----|------|
| GitHub Stars | +200 |
| Weekly brew installs | 100+ |
| HN karma | 50+ (agent-kay) |
| OS 보고된 critical bug | 0 |
| 사용자 보고된 vault corruption | 0 |
| Cosign verification 실패 | 0 |

---

## 11. PR 분할 전략

Sprint별로 ~10개 PR 단위로 분할. 각 PR의 평균 size = 200-500 LOC. **단일 PR에서 한 트랙만 다루기** (보안 핫픽스 ≠ 테스트 추가).

### 11.1 Sprint 1 PR 매트릭스 (예시)

| PR # | 제목 | 트랙 | LOC | 의존 |
|-----:|------|------|----:|------|
| 1 | feat(crypto): add DeriveSubKeyWithSaltV2 + info versioning | A | +180 | - |
| 2 | refactor(vault): migrate 16 sites to versioned KDF | A | +60 -40 | #1 |
| 3 | feat(crypto): add EncryptWithAAD + 6 RFC 8439 KAT | A | +220 | #1 |
| 4 | fix(crypto): RecoverySalt 8→16 bytes | A | +20 -10 | #2 |
| 5 | feat(crypto): KDFAlgRegistry + algorithm byte validation | A | +120 | #3 |
| 6 | refactor(sync): remove dead merge.go + queue.go | B | -244 +30 | - |
| 7 | refactor(sync): unify saveSyncState (camelCase schema) | B | -50 +40 | #6 |
| 8 | chore(sync): annotate math/rand jitter with #nosec G404 | B | +10 | #6 |
| 9 | fix(cli): emit stderr warning in JSON mode (U-1 guard) | C | +30 | - |
| 10 | docs(cli-reference): exit code drift fix (align docs to code) | C | +80 | - |
| 11 | feat(errors): separate STDOUT_SECRET_BLOCKED exit code (2→8) | C | +20 -10 | #10 |

총 11 PR / +800 -360 LOC / 2주

### 11.2 PR 검토 체크리스트 (모든 PR 공통)

- [ ] `go test -race ./...` 통과
- [ ] `golangci-lint run` clean
- [ ] CHANGELOG.md 업데이트 (BREAKING 시 명시)
- [ ] AC 충족 evidence (스크린샷/로그/grep 결과) PR 본문 첨부
- [ ] `gh pr create` 후 `staging` 머지 → auto-tag → rc1 검증
- [ ] cosign 서명 (S5 이후)

---

## 12. 부록

### A. 발견사항 전체 매트릭스 (file:line)

(상세 99개 발견 — audit 보고서 §1~§11 + 본 문서 §1.1~§1.3 참조. 본 부록은 file:line만 인덱스로)

#### A.1 Crypto P0 (4건)

```
pkg/crypto/kdf.go:39              DeriveKey(password, salt)         — 파라미터 packagae const lock-in
pkg/crypto/kdf.go:14-16           const ArgonTime/Memory/Threads    — 호환성 break 잠재
pkg/crypto/encrypt.go             aead.Seal(.., .., plaintext, nil) — AAD 부재
pkg/crypto/keymanager.go          salt[:8] truncation               — RFC 9106 < 16 bytes
internal/vault/vault.go (16곳)    DeriveSubKey(rootKey, nil, ...)   — salt nil
```

#### A.2 Vault P0 (1건) + P1 (3건)

```
internal/vault/vault.go:audit_log         resource_name TEXT plaintext
internal/vault/vault.go:82                duplicate vault.New
internal/vault/schema.go                  no schema_migrations
internal/vault/vault.go:Open              PRAGMAs missing 5건
```

#### A.3 Sync P0 (2건)

```
internal/sync/engine.go:Pull              calls neither merge nor queue
internal/sync/merge.go (244 LOC)          ThreeWayMerge dead
internal/sync/queue.go                    EnqueueOperation dead
internal/cli/push.go:163                  saveSyncState duplicate def
internal/sync/engine.go:445               saveSyncState duplicate def
```

#### A.4 Errors/Domain/Config/claudemd/Encfile P0+P1 (DR-4)

```
pkg/errors/codes.go:14,18,73,94          exit code mismatch
docs/cli-reference.md:23-32              advertised 3,4,5,6,7 emit 0
pkg/errors/errors.go (no Unwrap)         brittle errors.Is/As chain
pkg/errors/errors.go:54                  raw type assertion
internal/cli/init.go:177,179             generator error swallowed
internal/claudemd/generator.go:110-113   false positive heuristic
internal/claudemd/template.go            no version sentinel
internal/claudemd/template.go:11-23      LLM contradiction (Quick Ref vs Rules §8)
pkg/domain.SyncState vs config.SyncInfo  schema drift
internal/config (90% dead)
config_test.go HOME-only (no USERPROFILE)
internal/encfile/encfile.go:163         KDF params decode-only
internal/encfile/encfile.go:117,181     "tene-export" magic literal x2
internal/encfile/encfile.go:17,20        var FormatVersion (mutable)
```

#### A.5 Tests/CI/Lint P0+P1 (DR-5)

```
internal/sync/engine.go (473 LOC)        0 unit tests
pkg/crypto/crypto_test.go                0 KAT
N targets                                0 Fuzz / 0 testing.F
internal/cli/testhelper_test.go:96       rootCmd global mutation race
internal/cli/testhelper_test.go:39-79    resetFlags() 17 vars
internal/cli/get_guard_test.go:96-99     U-1 warning not asserted
internal/sync/conflict_test.go (27 LOC)  barely smoke
.github/workflows/ci.yml                 no matrix
.github/workflows/ci.yml                 no govulncheck
.github/workflows/ci.yml                 no coverage gate
.golangci.yml:16                         unused disabled
.golangci.yml:22                         default-signifies-exhaustive: true
```

#### A.6 Distribution P1 (DR-6)

```
.goreleaser.yaml                         no SLSA L3
.goreleaser.yaml                         no SBOM
.github/workflows/release.yml            no cosign
homebrew-tene/Formula/tene.rb            no bottle
n/a                                       no biometric auth
```

### B. 9개 OS×하드웨어 시나리오 매트릭스 (Biometric Auth)

(§7.2 Trk-A T4-A6 참고)

### C. 36 KAT 벡터 출처

```
RFC 8439 §A.2 XChaCha20-Poly1305:                 6 벡터
RFC 9106 §B.1 Argon2id:                            3 벡터
RFC 5869 §A.1-A.3 HKDF-SHA256:                     3 벡터
BIP-39 test vectors (Trezor):                      24 벡터
```

### D. 의존성 외부 패키지 모니터링 시트

| 패키지 | 현재 | 다음 점검일 | 트리거 |
|--------|------|------------|-------|
| cobra | v1.10.1 | 2026-06 | semver minor up |
| zalando/go-keyring | v0.2.6 | 2026-05-30 | 6개월 활동 없음 시 fork 검토 |
| modernc.org/sqlite | v1.39.0 | 2026-07 | perf 회귀 보고 시 mattn-sqlite3 검토 |
| x/crypto | v0.43.0 | 2026-06 | go.mod 자동 |
| go-bip39 | v1.1.0 | 안정 | - |
| lipgloss | v1.1.0 | 안정 | - |
| godotenv | v1.5.1 | 안정 | - |

### E. 본 계획서의 검증 방법

이 Sprint Master Plan은 다음 방법으로 검증되었다.

1. **소스 정독**: PM agent 1명 + DR agent 6명 (총 7 병렬) — 비-테스트 5,100 LOC + 테스트 5,090 LOC + CI yml + 배포 자료 = 약 89 파일 모두 라인 단위 정독
2. **Cross-check**: 각 audit claim을 `grep`/`go list`/`file:line` 로 재확인 — 31개 claim 중 31개 모두 confirm
3. **신규 발견**: cross-check 과정에서 audit가 놓친 P0 5건 (S-S1, S-S2, V1, C3, T-N1~N4) 추가
4. **의존성 분석**: §2 Critical Path 11주 — 병렬 가능한 트랙은 §2.2 매트릭스
5. **PR 분할**: §11에서 Sprint 1만 11개 PR 분해; Sprint 2~6는 같은 패턴으로 후속 PR 시점에서 분해

---

## 13. 다음 단계

이 Sprint Master Plan을 baseline으로:

1. **즉시 (2026-05-12)** — `staging` 에서 분기한 작업 브랜치 `sprint/s1-crypto-hotfix` 생성 (DONE)
2. **W1 day 1** — Sprint 1 Trk-A 시작 — T1-A1 (KDF v2 함수 추가) PR
3. **W1 day 2** — Trk-B + Trk-C 병렬 시작
4. **W1 day 5** — 1차 retrospective (이 계획서의 estimate 정확도 점검)
5. **W2 end** — `v1.0.9-rc1` 발행 + Sprint 2 kickoff

각 스프린트 끝에서:
- **/pdca check** 실행 — Match Rate 측정
- **/pdca report** 실행 — 회고 + 차기 스프린트 조정

---

> *"빠르게가 아닌 꼼꼼하게."* — 본 계획서는 모든 99건의 발견을 file:line 단위로 추적했고, 의존성 11주 critical path를 명시했고, PR 단위 200-500 LOC로 분해 가능한 형태로 작성했다. 13주 후 v2.0.0 stable로 만나길.
