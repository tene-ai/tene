---
template: plan
version: 1.3
description: tene CLI v2.0 Sprint 1 — audit-reader (P0-A1 + P0-E1 + P0-G1 + P0-K1; tene audit subcmd + exit code drift + U-1 guard + keychain probe)
variables:
  - feature: audit-reader
  - displayName: "audit-reader (P0-A1 + P0-E1 + P0-G1 + P0-K1)"
  - sprint: tene-cli-v2-s1
  - date: 2026-05-13
  - author: cto-lead (PM perspective)
  - project: tene CLI
  - version: v2.0.0 (target)
  - masterPlan: docs/01-plan/features/tene-cli-v2-2026q3.master-plan.md
  - trustLevel: L2
---

# audit-reader Planning Document

> **Summary**: `tene audit` 명령 신설 (`--since/--actor/--limit/--json`) + `audit_log.actor` 컬럼 + exit code 표 정렬 (docs ↔ code drift 해소) + `get.go` U-1 guard JSON warning fix + keychain Set+Delete 프로빙 제거. 4 P0 (A1, E1, G1, K1) 를 Sprint 1 Trk-C 로 묶음.
>
> **Project**: tene CLI
> **Version**: target v2.0.0 (current v1.0.8)
> **Sprint**: `tene-cli-v2-s1` (W1-W2, 2026-05-13 → 2026-05-26)
> **Author**: cto-lead
> **Date**: 2026-05-13
> **Status**: Draft (L2)
> **Master Plan**: [tene-cli-v2-2026q3.master-plan.md](./tene-cli-v2-2026q3.master-plan.md) §2.2 P0-A1, P0-E1, P0-G1, P0-K1, §3.1 보정 6
> **Baseline 분석**: [cli-sprint-master-plan-2026-05-13.md](../cli-sprint-master-plan-2026-05-13.md) §2.6 (보정 6) + §6.1.3 Trk-C T1-C1~C7
> **감사 출처**: [cli-completeness-audit-2026-05-11.md](../../03-report/cli-completeness-audit-2026-05-11.md) A10 P0-5, A5 P0, DR-4, DR-5, A7 P0

---

## Executive Summary

| Perspective | Content |
|-------------|---------|
| **Problem** | (1) **`tene audit` reader 부재** — `vault.go:424` 가 9 종 이벤트 (`vault.init`, `vault.passwd_changed`, ...) 를 `audit_log` 테이블에 기록만 하고 read CLI 없음. 사용자가 `sqlite3 .tene/vault.db "SELECT * FROM audit_log"` 직접 실행해야 함 — 감사 표현: "audit story is theater". (2) **Exit code drift** — `docs/cli-reference.md:23-32` 가 exit 3/4/5/6/7 광고하지만 코드는 1/2 만 emit; `AUTH_REQUIRED` 코드 0개. (3) **`get.go` U-1 guard 미검증** — JSON 모드에서 stdout secret 차단 시 stderr warning emit 필요한데 `get_guard_test.go:96` 가 미검증 confess. (4) **keychain Set+Delete 프로빙** — `keychain.go:91-97` 가 `NewStore()` 매 호출마다 securityd IPC 5-15ms × 2 ≈ 20-30ms overhead. (5) `init.go:221` next-step footer 가 1줄 (`tene set`) 만 — `tene list` + `tene run` 빠짐. |
| **Solution** | (1) `internal/cli/audit.go` 신설 — `tene audit --since DURATION --actor [human\|ai\|any] --limit N --json` Cobra command. SQL placeholder ? 사용 (injection 방어). (2) `audit_log.actor` 컬럼 추가 (TEXT DEFAULT 'human'; `TENE_ACTOR_ID` env 로 override). (3) `pkg/errors/codes.go` 전면 갱신 — 6 신규 exit code (`3=VAULT_NOT_FOUND`, `4=AUTH_FAILED`, `5=SECRET_NOT_FOUND`, `6=DECRYPT_FAILED`, `7=INTERACTIVE_REQUIRED`, `8=STDOUT_SECRET_BLOCKED`). docs 동기화. (4) `get.go:93-99` JSON 모드에서도 stderr "Refusing..." 출력 + test 실제 stderr assert. (5) `keychain.go:91-97` Set+Delete 프로빙 제거 — fallback 결정은 `Load()` 실패 시까지 미룸. (6) `init.go:221` next-step 3-line. |
| **Function/UX Effect** | **AI forensics 핵심 UX**: `tene audit --since 24h --actor ai --json` 으로 "어떤 AI 세션이 어떤 secret 을 만졌나" 즉시 확인. Show HN 메시지 #3 ("which AI session touched which secret, when?"). Exit code 정확성으로 자동화 신뢰. keychain 8-30ms 시동 절감. |
| **Core Value** | "AI agent forensics" 차별화의 핵심 (감사: 1-day feature 인데 plan 누락). v2.0 Show HN 메시지 4개 중 #3. |

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | audit_log 가 기록되지만 read CLI 없음 (theater); exit code 광고 4건이 실제 1/2 만 emit (자동화 사용자 신뢰 0); JSON 모드 secret 차단 warning 미검증 (silent regression 가능); keychain 시동 20-30ms overhead. 본 plan 의 4 P0 는 모두 "기능은 있는데 surface 노출 안 됨 / 광고와 실제 다름" 카테고리. |
| **WHO** | **AI-vibe coder (50%)** — `tene audit` forensics 핵심 가치. **Indie OSS dev (25%)** — exit code 정확성 (`tene get FOO \|\| handle_secret_missing`) daily UX. **Sec-conscious team (15%)** — JSON 모드 stderr warning regression detection. |
| **RISK** | (a) `tene audit` SQL injection (mitigation: `?` placeholder; integration test `'; DROP TABLE`). (b) exit code 변경 BREAKING (`STDOUT_SECRET_BLOCKED` 2→8) — 사용자 스크립트 깨짐 (mitigation: CHANGELOG BREAKING + `docs/migration/exit-codes.md`). (c) `audit_log.actor` 컬럼 추가 — 기존 v1 vault 호환 (mitigation: `DEFAULT 'human'`; 신규 vault 만; s2 의 `002_audit_log_v2` 가 backfill). (d) keychain 프로빙 제거 후 fallback path 신뢰성 — `Load()` 시 명시적 검증 추가. |
| **SUCCESS** | `tene audit --since 24h --json` 이 `[{action, resource_name, details, actor, timestamp}]` 배열 출력; `docs/cli-reference.md` 의 5 exit code 모두 1:1 매핑 (drift 0); JSON+nonTTY 모드 stderr "Refusing..." emit + test assert; `time tene version` benchmark 43ms → ≤ 35ms; `tene init` next-step 3 lines. |
| **SCOPE** | Sprint 1 Trk-C 단독 (3 PR: #8 audit reader + #11 exit code drift + #12 keychain probe). 5 dev-day, 1 dev. +460 LOC (impl +250 + test +150 + docs +60). 변경 범위: internal/cli (audit.go 신규 + get.go + init.go) + internal/vault (schema.go + vault.go) + pkg/errors + internal/keychain + docs/cli-reference.md + docs/migration/exit-codes.md. |

---

## 1. Overview

### 1.1 Purpose

v2.0 Show HN 메시지 4개 중 **3번째** ("tene audit — which AI session touched which secret, when?") 의 기술적 prerequisite. 동시에 exit code drift / JSON guard test / keychain probe 의 3 P0 정리 — 모두 Trk-C "CLI UX 핵심" 카테고리.

### 1.2 Background

#### 1.2.1 P0-A1 (보정 §2.6 — audit reader 부재)

```sql
-- internal/vault/schema.go:29-35 (현재 v1.0.8)
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    action TEXT NOT NULL,
    resource_name TEXT,  -- 평문 secret 이름 (P0-V1 — s2 에서 encrypted_details)
    details TEXT
)
```

`vault.go:424-434` 의 `AddAuditLog` 가 9 종 이벤트 기록:
- `vault.init`, `vault.passwd_changed`, `vault.recovered`
- `secret.created`, `secret.updated`, `secret.deleted`, `secret.read`
- `vault.exported`, `vault.imported`

그러나 사용자가 읽으려면 `sqlite3 .tene/vault.db "SELECT * FROM audit_log ORDER BY timestamp DESC LIMIT 50"` 직접 실행. 감사 표현: "Without this, the audit story is theater" — 1-day feature 인데 1차 plan 에 누락.

#### 1.2.2 P0-E1 (Exit Code Drift)

```
$ cat docs/cli-reference.md | grep -A 1 'Exit Code'
| 1 | General error |
| 2 | Authentication required |
| 3 | Vault not found |       ← 광고만; 코드는 emit 안 함
| 4 | Authentication failed | ← 광고만
| 5 | Secret not found |      ← 광고만
| 6 | Decryption failed |     ← 광고만
| 7 | Interactive required |  ← 광고만
```

```go
// pkg/errors/codes.go (현재 v1.0.8)
ErrGeneral = teneerror{Code: 1, ...}
ErrInvalidPassword = teneerror{Code: 2, ...}
ErrVaultNotFound = teneerror{Code: 2, ...}  // ← exit 3 광고하는데 실제 2
// 5,6,7 광고는 0건 emit
ErrStdoutBlocked = teneerror{Code: 2, ...}  // ← auth 그룹과 충돌
```

**해결**: 옵션 A (코드 변경 + docs 동기화). 새 표:

| Code | Name | Reason |
|:----:|------|--------|
| 0 | OK | Success |
| 1 | GENERAL_ERROR | 분류 불가 |
| 2 | INVALID_INPUT | 사용자 입력 오류 |
| 3 | VAULT_NOT_FOUND | `~/.tene/vault.db` 부재 |
| 4 | AUTH_FAILED | Master password 불일치 / auth_hash mismatch |
| 5 | SECRET_NOT_FOUND | `tene get NAME` 시 NAME 없음 |
| 6 | DECRYPT_FAILED | AAD mismatch / corruption |
| 7 | INTERACTIVE_REQUIRED | non-TTY 에서 prompt 필요 |
| 8 | STDOUT_SECRET_BLOCKED | `tene get` TTY 차단 (auth 그룹 분리) — BREAKING |

#### 1.2.3 P0-G1 (U-1 Guard JSON Warning)

```go
// internal/cli/get.go:93-99 (현재)
if isTerminal(os.Stdout) && !flagJSON {
    fmt.Fprintln(os.Stderr, "Refusing to print secret to TTY (use --json or pipe). Exit 2.")
    return teneerr.ErrStdoutBlocked
}
// ← JSON 모드에서는 warning 없이 secret 출력. 단지 stdout 만 JSON 으로 wrap.
//   `cli-completeness-audit §A5 P0`: JSON 도 OS 환경 의존성 (Cursor 등) 따라 차단 필요
```

```go
// internal/cli/get_guard_test.go:96 (현재)
// FIXME: actual stderr assertion missing; test just checks return code
```

**해결**: JSON+nonTTY 시 stderr 에 "Refusing..." emit + test 실제 stderr assert.

#### 1.2.4 P0-K1 (Keychain Probe)

```go
// internal/keychain/keychain.go:91-97 (현재)
func NewStore() (*Store, error) {
    // Probe keychain availability via dummy Set+Delete
    err := keyring.Set(serviceName, "_probe", "_dummy")
    if err != nil {
        return nil, fmt.Errorf("keychain unavailable: %w", err)
    }
    _ = keyring.Delete(serviceName, "_probe")
    return &Store{...}, nil
}
```

매 `NewStore()` 호출마다 macOS `securityd` IPC × 2 (Set + Delete) = 5-15ms × 2. `tene version` 도 keychain 초기화 → 매 명령 8-30ms overhead.

**해결**: 프로빙 제거. fallback 결정은 실제 `Load()` 실패 시까지 미룸.

### 1.3 Related Documents

- Master Plan: [`tene-cli-v2-2026q3.master-plan.md`](./tene-cli-v2-2026q3.master-plan.md)
- Baseline: [`cli-sprint-master-plan-2026-05-13.md §2.6 §6.1.3 Trk-C`](../cli-sprint-master-plan-2026-05-13.md)
- 감사: [`cli-completeness-audit-2026-05-11.md`](../../03-report/cli-completeness-audit-2026-05-11.md) A10 P0-5, A5 P0, A7 P0, DR-4, DR-5
- Sibling: [`passwd-verify.plan.md`](./passwd-verify.plan.md) (audit_log 에 `vault.passwd_failed` 이벤트 기록), [`crypto-v2-keys.plan.md`](./crypto-v2-keys.plan.md), [`sync-cleanup.plan.md`](./sync-cleanup.plan.md)

---

## 2. Scope

### 2.1 In Scope

#### 2.1.1 `tene audit` reader 명령 (P0-A1)

- [ ] `internal/cli/audit.go` (신규) — Cobra subcommand 등록
- [ ] Flags:
  - `--since DURATION` (e.g., `24h`, `7d`, `1month`) — 기본 24h
  - `--actor [human|ai|any]` — 기본 any
  - `--limit N` — 기본 50, max 1000
  - `--json` — JSON 배열 출력
- [ ] SQL query: `SELECT timestamp, action, resource_name, details, actor FROM audit_log WHERE timestamp >= ? AND (actor = ? OR ? = 'any') ORDER BY timestamp DESC LIMIT ?` — `?` placeholder
- [ ] Human-readable output (default): `[2026-05-13T10:30:42Z] [ai] secret.read AWS_KEY ()`
- [ ] JSON output: `[{"timestamp":"...", "action":"...", "resource_name":"...", "details":"...", "actor":"..."}]`
- [ ] Audit log table 에 `actor TEXT DEFAULT 'human'` 컬럼 추가 (`internal/vault/schema.go`)
- [ ] `AddAuditLog` 시 `actor = os.Getenv("TENE_ACTOR_ID")` 또는 `'human'` default
- [ ] CHANGELOG.md — "feat(cli): tene audit reader command"

#### 2.1.2 Exit code drift fix (P0-E1)

- [ ] `pkg/errors/codes.go` — 새 8개 exit code 표 (위 §1.2.2)
- [ ] `STDOUT_SECRET_BLOCKED` exit 2 → 8 (auth group 분리) — **BREAKING**
- [ ] `docs/cli-reference.md` — exit code 표 동기화
- [ ] `docs/migration/exit-codes.md` (신규) — v1 → v2 mapping + 자동화 스크립트 예시
- [ ] CHANGELOG.md — "BREAKING: exit code 2 (STDOUT_BLOCKED) → 8 (AUTH 그룹 분리)"

#### 2.1.3 U-1 guard JSON warning fix (P0-G1)

- [ ] `internal/cli/get.go:93-99` — JSON 모드에서도 stderr "Refusing..." emit (terminal stdout 차단 시)
- [ ] `internal/cli/get_guard_test.go` — `os.Stderr` 캡처 → "Refusing" 문자열 assert (confess 제거)

#### 2.1.4 Keychain probe 제거 (P0-K1)

- [ ] `internal/keychain/keychain.go:91-97` — `keyring.Set("_probe", ...) + keyring.Delete` 제거
- [ ] `Load()` 함수에 첫 호출 시 명시적 fallback 결정 (lazy init)
- [ ] benchmark — `time tene version` 43ms → ≤ 35ms 확인

#### 2.1.5 `tene init` next-step 3-line (T1-C7, P1)

- [ ] `internal/cli/init.go:221` — footer 3-line:
  ```
  Next steps:
    tene set FOO bar    # store a secret
    tene list           # list all secrets
    tene run -- npm start  # run command with secrets injected
  ```

### 2.2 Out of Scope

- `audit_log.resource_name` + `details` 평문 → encrypted_details — **Sprint 2** `vault-v2-migration` (`002_audit_log_v2`)
- 기존 v1 vault 의 `audit_log.actor` 컬럼 backfill — **Sprint 2** schema migration
- audit log 자체 변조 방지 (chain hash, signed audit) — v2.1 또는 별도 plan
- `--format=csv` 또는 `--format=ndjson` flag — v2.1 (현재는 human + JSON 두 종)
- `tene audit follow` (tail mode) — v2.1
- HKDF info 도메인 (`tene/audit/v1`) 통합 — sibling `crypto-v2-keys.plan.md` 의 PurposeAudit constant
- `--limit > 1000` 의 page-based read — v2.1

---

## 3. Requirements

### 3.1 Functional Requirements

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-01 | `tene audit --since 24h --json` 이 JSON 배열 출력 (5 필드) | **P0** | Pending |
| FR-02 | `--since` flag 가 Go `time.ParseDuration` 또는 `1d`/`7d`/`30d` shorthand 지원 | **P0** | Pending |
| FR-03 | `--actor [human\|ai\|any]` flag — 기본 `any`; SQL filter | **P0** | Pending |
| FR-04 | `--limit N` flag — 기본 50, max 1000, 검증 | **P0** | Pending |
| FR-05 | SQL query 가 `?` placeholder 사용 — SQL injection 방어 | **P0** | Pending |
| FR-06 | 신규 vault 의 audit_log 에 `actor TEXT DEFAULT 'human'` 컬럼 | **P0** | Pending |
| FR-07 | `AddAuditLog` 가 `TENE_ACTOR_ID` env var 우선 사용 (없으면 `'human'`) | **P0** | Pending |
| FR-08 | exit code 1-8 새 표 — `pkg/errors/codes.go` + `docs/cli-reference.md` 1:1 매핑 | **P0** | Pending |
| FR-09 | `STDOUT_SECRET_BLOCKED` exit 2 → 8 (BREAKING) | **P0** | Pending |
| FR-10 | `get.go` JSON+nonTTY 시 stderr "Refusing..." emit + test assert | **P0** | Pending |
| FR-11 | `keychain.NewStore()` Set+Delete 프로빙 제거; lazy init in `Load()` | **P0** | Pending |
| FR-12 | `tene init` next-step footer 3-line | P1 | Pending |
| FR-13 | `tene audit --json` empty 결과 → `[]` (not null) | **P0** | Pending |
| FR-14 | `tene audit` 가 `'; DROP TABLE audit_log;` 같은 input 받아도 SQL 실행 안 함 (integration test) | **P0** | Pending |

### 3.2 Non-Functional Requirements

| Category | Criteria | Measurement Method |
|----------|----------|-------------------|
| Security | SQL injection 방어 (placeholder 강제) | integration test (malicious input) |
| Performance | `tene version` 8-30ms 절감 (keychain probe 제거) | `time` benchmark |
| Performance | `tene audit --limit 50` < 100ms (50 rows) | bench |
| Compatibility | 기존 v1 vault — `actor` 컬럼 부재 시 `'human'` 가정 (read tolerance) | integration test |
| Documentation | `docs/cli-reference.md` exit code 표 docs ↔ code drift 0 | manual diff |
| Auditability | audit_log read 자체도 audit_log 기록 (meta-audit) | integration test |

---

## 4. Success Criteria

### 4.1 Definition of Done

- [ ] FR-01 ~ FR-14 모두 충족 (grep + integration test + bench)
- [ ] `internal/cli/audit.go` + `audit_test.go` 신설
- [ ] `internal/vault/schema.go` — `audit_log.actor` 컬럼
- [ ] `pkg/errors/codes.go` — 8 exit code
- [ ] `docs/cli-reference.md` 동기화 (manual diff)
- [ ] `docs/migration/exit-codes.md` (신규) — v1 → v2 매핑
- [ ] `internal/cli/get.go` + `get_guard_test.go` — JSON warning fix
- [ ] `internal/keychain/keychain.go` — probe 제거
- [ ] `internal/cli/init.go:221` — 3-line footer
- [ ] CHANGELOG.md (BREAKING + feat + fix + perf 분리)
- [ ] `go test -race -count=10 ./...` 0 회귀
- [ ] gap-detector Match Rate ≥ 90% (M8)

### 4.2 Quality Criteria

- [ ] `internal/cli/audit.go` coverage ≥ 80%
- [ ] integration test — SQL injection 시도 (`'; DROP TABLE`) → 안전
- [ ] integration test — `tene get FOO --json` TTY stdout → exit 8 + stderr emit
- [ ] benchmark — `tene version` 시동 시간 측정 (before/after)

---

## 5. Risks and Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| `tene audit` SQL injection (--actor 등) | H | M | `?` placeholder 모든 user input; integration test `'; DROP TABLE` |
| Exit code 2→8 BREAKING 자동화 스크립트 깨짐 | M | M | CHANGELOG BREAKING 헤더 + `docs/migration/exit-codes.md` + Show HN 발사 24h 전 사전 공지 (Daily.dev squad) |
| 기존 v1 vault 의 audit_log 에 `actor` 컬럼 부재 → SELECT 실패 | M | H | SQL: `SELECT ..., COALESCE(actor, 'human') as actor FROM audit_log` (read tolerance); 신규 vault 만 컬럼 보유 |
| keychain probe 제거 후 fallback 시점 변경 → 일부 환경 회귀 | M | M | `Load()` 의 첫 호출 시 명시적 검증 + fallback 메시지 stderr emit |
| `--since 1y` 등 large window 시 vault.db 부담 | L | M | `--limit 1000` 강제 cap; SQL `LIMIT 1000` |
| `audit_log` 자체 변조 (root 권한) | M | L | scope out — v2.1 의 chain hash 또는 signed audit |
| audit log read 자체 audit (meta-audit) 누락 | L | M | `AddAuditLog("audit.read", ...)` 호출 추가 |
| docs/migration/exit-codes.md 작성 누락 | M | M | PR description checklist 에 포함 |

---

## 6. Impact Analysis

### 6.1 Changed Resources

| Resource | Type | Change Description |
|----------|------|--------------------|
| `internal/cli/audit.go` | File | NEW (Cobra subcommand) |
| `internal/cli/audit_test.go` | File | NEW |
| `internal/cli/root.go` | Function | `rootCmd.AddCommand(auditCmd)` 등록 |
| `internal/vault/schema.go:29` | Schema | `audit_log` 에 `actor TEXT DEFAULT 'human'` (신규 vault) |
| `internal/vault/vault.go:AddAuditLog` | Function | `actor` 파라미터 추가 (`TENE_ACTOR_ID` env or `'human'`) |
| `internal/vault/vault.go:GetAuditLog` | Function (신규) | `(since, actor, limit) → []AuditEntry` |
| `pkg/errors/codes.go` | Constants | 8 exit code 표 |
| `internal/cli/get.go:93-99` | Function | JSON+nonTTY 시 stderr emit |
| `internal/cli/get_guard_test.go:96` | Test | confess 제거 + 실제 stderr assert |
| `internal/keychain/keychain.go:91-97` | Function | probe 제거 |
| `internal/cli/init.go:221` | Function | next-step 3-line |
| `docs/cli-reference.md` | Documentation | exit code 표 동기화 |
| `docs/migration/exit-codes.md` | File (신규) | v1 → v2 mapping + 자동화 예시 |
| `CHANGELOG.md` | Documentation | BREAKING + feat + fix + perf |

### 6.2 Current Consumers

| Resource | Operation | Code Path | Impact |
|----------|-----------|-----------|--------|
| `audit_log` table | INSERT (9 events) | `vault.go:AddAuditLog` 9 caller (init.go, passwd.go, recover.go, get.go, set.go, delete.go, export.go, import_cmd.go) | Schema 호환 (DEFAULT 'human'); caller 시그니처 변경 — actor 파라미터 추가 |
| `audit_log` table | SELECT | (없음 — read CLI 0) | **NEW consumer** — audit.go |
| `pkg/errors.Err*` | Exit code | grep 결과 ~30 사이트 | Migration (대부분 unchanged; `ErrStdoutBlocked` 만 2→8) |
| `cli-reference.md` | Documentation | docs site | 동기화 |
| `keychain.NewStore` | Probe | `cli/root.go:loadApp()` + ~5 사이트 | Performance fix (semantic 동일) |
| `init.go:221` next-step | UX | terminal output | Cosmetic |

### 6.3 Verification

- [ ] grep `audit_log` table SELECT — `internal/cli/audit.go:?` 만 1 사이트 (신규)
- [ ] grep `TENE_ACTOR_ID` ≥ 2 사이트 (env read + audit.go)
- [ ] integration test — SQL injection 시도 → vault.db corruption 0
- [ ] benchmark — `time tene version` before vs after (8-30ms 절감 확인)
- [ ] manual diff — `docs/cli-reference.md` 의 exit code 표 ↔ `pkg/errors/codes.go` 의 Code 필드

---

## 7. Architecture Considerations

### 7.1 Project Level Selection

| Level | Selected |
|-------|:--------:|
| Enterprise (tene CLI) | **☑** |

### 7.2 Key Architectural Decisions

| Decision | Options | Selected | Rationale |
|----------|---------|----------|-----------|
| `tene audit` 명령 위치 | (a) `internal/cli/audit.go` 단일 subcommand, (b) `internal/cli/audit/` 패키지 (multi-subcommand), (c) `tene log` 별칭 | (a) 단일 subcommand | YAGNI — v2.0 은 read 만; backup/export 는 v2.1 |
| Exit code option | (a) BREAKING 변경 (옵션 A), (b) `--exit-code v2` flag (opt-in), (c) docs 만 수정 (코드 unchanged) | (a) BREAKING | 1차 plan 의 채택; v2.0 major bump 의 일부; 자동화 스크립트는 migration guide |
| `actor` 컬럼 표현 | (a) free-form TEXT (`'cursor-session-abc'`), (b) enum (`'human'\|'ai'\|'automation'`), (c) JSON `{kind, id, session}` | (a) free-form + `--actor [human\|ai\|any]` filter | flexibility; ai/human classification 은 `TENE_ACTOR_ID` 사용자 책임 (예: Cursor 가 `cursor-session-{uuid}` 설정) |
| SQL placeholder vs prepared stmt | (a) `?` literal, (b) `Prepare + Stmt`, (c) GORM | (a) `?` literal | `modernc.org/sqlite` 직접 사용; prepared stmt 는 v2.1 perf optimization (P2-Perf3) |
| keychain probe alternative | (a) Set+Delete 프로빙, (b) `Lookup()` (read-only) 1회, (c) probe 자체 제거 + lazy init | (c) probe 제거 + lazy | securityd IPC 자체가 비용; fallback 은 실제 사용 시점 |
| `--since` 형식 | (a) Go `time.ParseDuration` (`24h`만), (b) `1d`/`7d` shorthand 추가, (c) ISO 8601 datetime | (a) + (b) — `24h`, `7d`, `30d` shorthand | UX 친화; ISO 8601 은 v2.1 |

### 7.3 Clean Architecture Approach

```
Layer mapping:
┌──────────────────────────────────────────────────────────┐
│ Presentation (internal/cli/)                             │
│   - audit.go (NEW; Cobra command)                        │
│   - get.go (JSON guard fix)                              │
│   - init.go (next-step 3-line)                           │
├──────────────────────────────────────────────────────────┤
│ Application (internal/cli/ — same as above)              │
│   (각 명령 RunE 가 use case 역할; s2 의 architecture 에서  │
│    분리 예정)                                             │
├──────────────────────────────────────────────────────────┤
│ Domain (없음 — 이 plan 은 SQL CRUD + UX 만)               │
├──────────────────────────────────────────────────────────┤
│ Infrastructure (internal/vault/, internal/keychain/)     │
│   - vault.go (AddAuditLog actor 파라미터 추가)            │
│   - vault.go (GetAuditLog NEW)                           │
│   - schema.go (actor 컬럼)                                │
│   - keychain.go (probe 제거)                              │
├──────────────────────────────────────────────────────────┤
│ Cross-cutting (pkg/errors/)                              │
│   - codes.go (8 exit code 표)                            │
└──────────────────────────────────────────────────────────┘
```

---

## 8. Convention Prerequisites

### 8.1 Existing Project Conventions

- [x] Cobra subcommand 등록 패턴 — `root.go:rootCmd.AddCommand(...)`
- [x] `--json` flag 패턴 (`flagJSON` global) — 모든 명령 일관
- [x] SQLite KV/table read 패턴 — `internal/vault/vault.go` 의 `GetMeta`, `GetSecret`
- [x] Test fixture 패턴 — `testhelper_test.go` (s2 에서 리팩토)

### 8.2 Conventions to Define/Verify

| Category | Current State | To Define | Priority |
|----------|---------------|-----------|:--------:|
| Exit code 표 | 산발적 (drift) | 8 entry 표준 + docs/migration | **High** |
| `--since DURATION` 파싱 | (없음) | Go `time.ParseDuration` + `Nd` shorthand 헬퍼 | Medium |
| `actor` 컬럼 명명 | (없음) | `actor TEXT DEFAULT 'human'` (audit_log) | Medium |
| audit log meta-audit | (없음) | audit.read action 도 audit_log 기록 | Low |
| `TENE_ACTOR_ID` env var | (없음) | 표준 명명 + CLAUDE.md 안내 | Medium |

### 8.3 Environment Variables Needed

| Variable | Purpose | Scope | To Be Created |
|----------|---------|-------|:-------------:|
| `TENE_ACTOR_ID` | audit_log `actor` 컬럼 값 (사용자 직접 set; Cursor 등 wrapper 가 `cursor-session-{uuid}` 설정) | Client/Server | ☑ (신규) |

---

## 9. Testing Plan

### 9.1 Test Scope

| Type | Target | Tool | Phase |
|------|--------|------|-------|
| L1: Unit | `parseDuration` (shorthand + Go std), `GetAuditLog` SQL | `go test` | Do |
| L2: CLI | `tene audit --since 24h --json` 실행 + JSON 파싱 | testhelper_test.go | Do |
| L2: Security | SQL injection 시도 (`--actor "human'; DROP TABLE audit_log; --"`) | integration_test.go | Do |
| L3: Regression | `tene get FOO --json` TTY stdout — exit 8 + stderr emit | testhelper_test.go | Do |
| L4: Performance | `time tene version` benchmark — 43ms → ≤ 35ms | `go test -bench` + `time` | Do |
| L5: Compatibility | v1 vault fixture (actor 컬럼 없음) + `tene audit` | integration_test.go | Do |

### 9.2 Test Scenarios

| # | Scenario | Expected |
|---|----------|---------|
| T-1 | `tene audit --since 24h --json` (empty audit_log) | `[]` (not null) |
| T-2 | `tene audit --since 1h --actor ai --json` (5 events, 3 ai) | JSON array of 3 |
| T-3 | `tene audit --since 1y --limit 1000` | max 1000 rows |
| T-4 | `tene audit --since invalid` | exit 2 + error message |
| T-5 | `tene audit --actor "h'; DROP TABLE;--"` | SQL safe; vault.db intact |
| T-6 | `tene get FOO` TTY stdout (no --json) | exit 8 + stderr "Refusing..." |
| T-7 | `tene get FOO --json` TTY stdout | exit 8 + stderr "Refusing..." (NEW) |
| T-8 | `tene get FOO --json \| cat` (pipe, non-TTY stdout) | exit 0 + JSON output |
| T-9 | `time tene version` | ≤ 35ms (was 43ms) |
| T-10 | `tene init` | next-step footer 3 lines |
| T-11 | v1 vault fixture (actor 컬럼 없음) + `tene audit --json` | actor = 'human' (COALESCE) |
| T-12 | `tene audit` 자체 audit_log 기록 (meta-audit) | next `tene audit` 가 `audit.read` 이벤트 포함 |
| T-13 | `STDOUT_SECRET_BLOCKED` exit 8 | code emit 정확 |

---

## 10. Implementation Notes (Pre-Design Hints)

- `parseDuration` helper: 먼저 `time.ParseDuration` 시도; 실패 시 `Nd`/`Nw`/`Nmonth` regex 파싱 (`1d = 24h`, `1w = 168h`, `1month = 720h`)
- `GetAuditLog` SQL:
  ```sql
  SELECT timestamp, action, resource_name, details, COALESCE(actor, 'human') AS actor
  FROM audit_log
  WHERE timestamp >= ? AND (? = 'any' OR actor = ?)
  ORDER BY timestamp DESC
  LIMIT ?
  ```
- audit.read meta-audit: `defer app.Vault.AddAuditLog("audit.read", fmt.Sprintf("since=%s,actor=%s,limit=%d", since, actor, limit), "", actor)` (re-entry 위험 없음 — read 가 read 를 트리거 안 함)
- `keychain.NewStore()` 변경: 함수 body 를 `return &Store{}, nil` 로 단순화; `Load()` 에서 fallback 결정
- exit code constant: `pkg/errors/codes.go` 의 `Code` 필드 8개 — `Code: 3` (VAULT_NOT_FOUND), `Code: 8` (STDOUT_SECRET_BLOCKED) 등 — 컴파일 time constant

---

## 11. Blocking Relationships

| Blocks | Reason |
|--------|--------|
| **`vault-v2-migration`** (Sprint 2) | `002_audit_log_v2` migration 이 본 plan 의 `actor` 컬럼 + `resource_name_hmac` + `encrypted_details` 컬럼을 backfill — 본 plan 이 actor 컬럼 먼저 도입 |
| **Sprint 6 documentation-migration** | `docs/migration/exit-codes.md` 가 본 plan 에서 신설 → v1→v2 migration guide 의 일부 |

| Blocked by | Reason |
|-----------|--------|
| (none) | unblocked from start; sibling 4 features (passwd-verify, crypto-v2-keys, crypto-v2-aad, sync-cleanup) 와 병렬 가능 |

---

## 12. Next Steps

1. [x] Plan 문서 작성 완료
2. [ ] Design 문서 작성 — [`docs/02-design/features/audit-reader.design.md`](../../02-design/features/audit-reader.design.md) — frontend-architect 위임 (CLI UX 중심) + security-architect 위임 (SQL injection 방어)
3. [ ] L2 boundary: design 후 사용자 승인 → do phase
4. [ ] Sprint 1 PR #8 (audit reader) → #11 (exit code drift) → #12 (keychain probe)
5. [ ] Sprint 2 의 `002_audit_log_v2` migration 이 본 plan 의 actor 컬럼 활용 (encrypted_details + resource_name_hmac 추가)
6. [ ] Sprint 6 docs migration 에 `docs/migration/exit-codes.md` 포함

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-05-13 | Initial draft (Sprint 1 plan phase, L2 boundary) | cto-lead |
