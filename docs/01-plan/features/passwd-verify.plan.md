---
template: plan
version: 1.3
description: tene CLI v2.0 Sprint 1 — passwd-verify (P0-P1 hotfix; old password verification gate via auth_hash)
variables:
  - feature: passwd-verify
  - displayName: "passwd-verify (P0-P1)"
  - sprint: tene-cli-v2-s1
  - date: 2026-05-13
  - author: cto-lead (PM perspective)
  - project: tene CLI
  - version: v2.0.0 (target)
  - masterPlan: docs/01-plan/features/tene-cli-v2-2026q3.master-plan.md
  - trustLevel: L2
---

# passwd-verify Planning Document

> **Summary**: `tene passwd` 와 `tene recover` 가 현재 old password 를 검증하지 않음. keychain 캐시가 silent 반환되어 사용자가 old password 를 모르고도 master 회전 가능. `vault_meta.auth_hash` 컬럼 + 명시적 prompt + `subtle.ConstantTimeCompare` 로 강제 검증한다.
>
> **Project**: tene CLI
> **Version**: target v2.0.0 (current v1.0.8)
> **Sprint**: `tene-cli-v2-s1` (W1-W2, 2026-05-13 → 2026-05-26)
> **Author**: cto-lead (orchestrating product-manager perspective)
> **Date**: 2026-05-13
> **Status**: Draft (L2 — design 진행 전 사용자 검토 가능)
> **Master Plan**: [tene-cli-v2-2026q3.master-plan.md](./tene-cli-v2-2026q3.master-plan.md) §2.2 P0-P1
> **Baseline 분석**: [cli-sprint-master-plan-2026-05-13.md](../cli-sprint-master-plan-2026-05-13.md) §2.5 (보정 5) + §6.1.3 Trk-A T1-A1/A2
> **감사 출처**: [cli-completeness-audit-2026-05-11.md](../../03-report/cli-completeness-audit-2026-05-11.md) A1 P0-1

---

## Executive Summary

| Perspective | Content |
|-------------|---------|
| **Problem** | `tene passwd` 가 old password 를 검증하지 않는다. `loadOrPromptMasterKey` 가 keychain 캐시 hit 시 즉시 return 하므로 UI 는 "Enter current Master Password" 메시지를 표시하지만 실제로는 password 입력을 요구하지 않는다. 침입자가 노트북 unlock 상태에서 `tene passwd` 1줄로 master 회전 가능. 동일 결함이 `tene recover` 에도 존재. |
| **Solution** | (1) `vault_meta` 테이블에 `auth_hash BLOB` (32 bytes) 컬럼 추가 (init 시 `HKDF-Expand(masterKey, "tene-auth-hash-v1", 32)` 저장). (2) `loadOrPromptMasterKey` 를 `loadCachedMasterKey()` + `promptAndVerifyMasterKey()` 두 함수로 분리. (3) passwd/recover RunE 는 **항상** keychain 우회 후 `term.ReadPassword` 강제 + `subtle.ConstantTimeCompare` 로 `auth_hash` 비교. 불일치 시 `ErrInvalidPassword` (exit 4 — `audit-reader` 의 새 exit code 표 참조). |
| **Function/UX Effect** | passwd / recover 명령만 keychain bypass — `tene get/run/list/set` 등 일상 명령은 keychain 캐시 그대로 사용 (UX 회귀 없음). 한 번의 추가 password 입력으로 master 회전 보안 보장. |
| **Core Value** | "Local-first + AI-vibe coder 안전" 의 신뢰 토대. v2.0 Show HN 메시지 "biometric + supply chain + audit forensics" 의 전제. |

---

## Context Anchor

> Auto-generated from Executive Summary. Propagated to Design/Do documents for context continuity.

| Key | Value |
|-----|-------|
| **WHY** | `tene passwd` / `tene recover` 가 old password 검증을 안 함 → 노트북 unlock 상태에서 1줄 명령으로 master 회전 가능 (P0 보안 결함). |
| **WHO** | 모든 tene 사용자. 특히 **AI-vibe coder (50%)** 가 노트북을 자주 unlock 상태로 둠 → 동거인/사무실 동료 잠재 위협 모델. **Sec-conscious team (15%)** 는 이 결함 발견 시 즉시 도입 거부. |
| **RISK** | (a) auth_hash 컬럼이 schema_migrations 없는 상태에서 추가 → 신규 vault 만 동작; 기존 v1.0.x vault 는 grace path 필요 (s2 의 `001_v2_envelope` migration 에서 backfill). (b) passwd 도입 시 사용자가 매번 password 입력 → 불만 가능 → **passwd/recover 에 한정**. (c) `subtle.ConstantTimeCompare` 잘못 사용 시 timing leak (mitigation: bytes 길이 32 고정 강제). |
| **SUCCESS** | grep `auth_hash` 매치 ≥ 2 (`vault_meta.auth_hash` 저장 + ConstantTimeCompare 호출); passwd_test.go 에 "잘못된 old password → ErrInvalidPassword" 테스트 통과; `tene passwd` 가 keychain 캐시 있어도 prompt 표시; T2-B7 passwd_test.go (s2) 도 통과 가능 구조. |
| **SCOPE** | Sprint 1 Trk-A 핫픽스 2 PR (PR #5 auth_hash 컬럼 + PR #6 verify gate). +240 LOC (impl +160 + test +80). 모든 변경은 internal/cli + internal/vault + pkg/crypto (3 패키지). |

---

## 1. Overview

### 1.1 Purpose

tene CLI v1.0.8 의 가장 심각한 사용자-facing 보안 결함을 즉시 해소한다. v2.0 의 다른 모든 보안 강화 (생체 인증, SLSA L3, KAT 36) 가 이 결함 위에 쌓이면 의미 없음 — **passwd 검증 결함은 모든 보안 narrative 의 prerequisite**.

### 1.2 Background

#### 1.2.1 결함의 발견 경로

감사 A1 P0-1 (`cli-completeness-audit-2026-05-11.md` §A1) 가 발견. baseline 보정 §2.5 가 코드 정독으로 재검증. 1차 plan (2026-05-12) 에 누락 → 본 plan 에서 정식 도입.

#### 1.2.2 결함의 정확한 위치

```go
// internal/cli/passwd.go:29-34 (현재 v1.0.8)
_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Enter current Master Password:")
oldMasterKey, err := loadOrPromptMasterKey(app)
//                  ↑ UI 는 "password 입력" 보여주지만 ↓ 함수는 keychain 에서 silent load
if err != nil {
    return teneerr.ErrInvalidPassword  // ← 도달 불가능 (keychain hit 시 err == nil)
}

// internal/cli/root.go:173-198 (현재)
func loadOrPromptMasterKey(app *App) ([]byte, error) {
    key, err := app.Keychain.Load()
    if err == nil {
        return key, nil  // ← 즉시 return, prompt 없음
    }
    if pw := os.Getenv("TENE_MASTER_PASSWORD"); pw != "" {
        return deriveMasterKeyFromPassword(app, pw)
    }
    if !isTerminal() {
        return nil, teneerr.ErrInteractiveRequired
    }
    fmt.Fprint(os.Stderr, "Enter Master Password: ")
    password, err := term.ReadPassword(int(os.Stdin.Fd()))
    // ...
}
```

#### 1.2.3 공격 시나리오

1. 사용자 A 가 tene 설치 + master password 입력 → keychain 캐시 활성.
2. 사용자 A 가 노트북을 unlock 상태로 사무실에 두고 외출.
3. 동료 B 가 터미널 열고 `tene passwd` 실행.
4. UI: "Enter current Master Password:" — B 는 아무 키도 누르지 않음.
5. `loadOrPromptMasterKey` 가 keychain hit → 즉시 reply.
6. UI: "Enter new Master Password: " — B 가 임의 password 입력.
7. master 회전 완료. 사용자 A 는 다음 unlock 시점에 본인 password 가 안 먹는다.
8. **암호화된 데이터 자체는 안전** (B 는 vault 의 secret 평문에 접근 불가; new master 도 모름 — 회전 후 vault 는 B 의 new master 로 재암호화됨) — 그러나 **사용자 A 의 vault 가 사실상 영구적으로 잠김** (recovery key 가 없으면 복구 불가).

이는 통상적인 "snapshot encryption" 의 약속을 깨는 결함. v2.0 Show HN 발사 전에 반드시 해소.

### 1.3 Related Documents

- Master Plan: [`tene-cli-v2-2026q3.master-plan.md`](./tene-cli-v2-2026q3.master-plan.md)
- Baseline 분석: [`cli-sprint-master-plan-2026-05-13.md §2.5 §6.1.3`](../cli-sprint-master-plan-2026-05-13.md)
- 감사 출처: [`cli-completeness-audit-2026-05-11.md §A1`](../../03-report/cli-completeness-audit-2026-05-11.md)
- Sibling features (Sprint 1): [`crypto-v2-keys.plan.md`](./crypto-v2-keys.plan.md) (P0-C1; HKDF info 버전화), [`crypto-v2-aad.plan.md`](./crypto-v2-aad.plan.md), [`sync-cleanup.plan.md`](./sync-cleanup.plan.md), [`audit-reader.plan.md`](./audit-reader.plan.md)

---

## 2. Scope

### 2.1 In Scope

- [ ] `internal/vault/schema.go` — `vault_meta` 에 `auth_hash` row (key='auth_hash', value=base64(HKDF expanded 32 bytes)) 추가 (별도 컬럼 ALTER 불필요 — 현재 `vault_meta` 는 (key, value) KV table)
- [ ] `internal/cli/init.go` — vault 생성 시 `app.Vault.SetMeta("auth_hash", encodeBase64(authHash))` 호출
- [ ] `internal/cli/root.go` — `loadOrPromptMasterKey` 분리 → `loadCachedMasterKey()` + `promptAndVerifyMasterKey()`
- [ ] `internal/cli/passwd.go:30-31` — `promptAndVerifyMasterKey()` 호출로 교체 (keychain bypass)
- [ ] `internal/cli/recover.go` — 동일 패턴 적용 (recover 도 사실상 master 회전; 현재 mnemonic-only verify; 추가로 reset 후 새 password 강제는 별개)
- [ ] `pkg/crypto/kdf.go` — `PurposeAuthHash` constant (`"tene-auth-hash-v1"`) 추가 (또는 sibling `crypto-v2-keys` 의 `pkg/crypto/info.go` 와 통합)
- [ ] `internal/cli/passwd_test.go` — 신규 unit test (3 케이스: 정확한 old password, 잘못된 old password, keychain cache 무시)
- [ ] `internal/cli/root_test.go` — `promptAndVerifyMasterKey()` 단위 테스트 (mock keychain)
- [ ] CHANGELOG.md — "BREAKING: tene passwd now always prompts for current Master Password (security hotfix)"

### 2.2 Out of Scope

- 기존 v1.0.x vault 의 `auth_hash` 자동 backfill — Sprint 2 의 `vault-v2-migration` (`001_v2_envelope`) 에서 처리. 현재 plan 은 **신규 vault 만** auth_hash 보유; 기존 vault 는 `auth_hash` 부재 시 grace path (warn + 1회 backfill prompt)
- biometric prompt 통합 — Sprint 4 `biometric-auth` 가 `promptAndVerifyMasterKey` decorator 추가
- recovery key 의 자동 회전 — passwd 변경 시 새 recovery key 발급 (이미 현재 코드에 있음; 검증만)
- `TENE_MASTER_PASSWORD` env var 도 동일 verify gate 통과 강제 — env path 도 verify (auto-pause 없음, 동일 ConstantTimeCompare)
- exit code 변경 (2→8) — `audit-reader` plan 의 P0-E1 에서 처리

---

## 3. Requirements

### 3.1 Functional Requirements

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-01 | 신규 vault 생성 시 `vault_meta.auth_hash` 가 32-byte HKDF-Expand(masterKey, "tene-auth-hash-v1", 32) 결과로 저장 | **P0** | Pending |
| FR-02 | `tene passwd` 가 keychain 캐시 hit 여부 상관없이 `term.ReadPassword` 로 old password 입력을 받음 | **P0** | Pending |
| FR-03 | 입력한 old password 로 derive 한 master key 의 auth_hash 가 vault_meta.auth_hash 와 `subtle.ConstantTimeCompare == 1` 일치할 때만 master 회전 진행 | **P0** | Pending |
| FR-04 | 불일치 시 `teneerr.ErrInvalidPassword` 반환 (exit code 는 `audit-reader` 의 새 표 따름 — 신규 4 = AUTH_FAILED) | **P0** | Pending |
| FR-05 | `tene recover` 도 동일 verify gate 적용 (mnemonic 검증 + new password 강제 + 새 auth_hash 저장) | **P0** | Pending |
| FR-06 | `TENE_MASTER_PASSWORD` env var path 도 verify gate 통과 (env로 자동화 가능하지만 보안 우회 불가) | **P0** | Pending |
| FR-07 | 기존 v1.0.x vault (auth_hash 부재) 에서 `tene passwd` 실행 시: stderr 에 "auth_hash backfill required" warning + 1회 prompt 후 새 auth_hash 저장 후 진행 (grace path) | P1 | Pending |
| FR-08 | passwd 성공 후 새 master key 의 auth_hash 도 vault_meta 에 갱신 | **P0** | Pending |

### 3.2 Non-Functional Requirements

| Category | Criteria | Measurement Method |
|----------|----------|-------------------|
| Security | OWASP A02 (Crypto Failures) — passwd verify gate 충족 | grep + integration test |
| Security | Constant-time comparison (no timing leak) | `subtle.ConstantTimeCompare` 명시 사용; 32 bytes fixed length |
| Performance | passwd RunE 추가 latency < 200ms (Argon2id 1회 + HKDF 1회) | `go test -bench` |
| Compatibility | 기존 v1.0.x vault 호환 (grace path) | integration test (v1 vault fixture + passwd 명령) |
| Auditability | `audit_log` 에 `vault.passwd_failed` 이벤트 기록 (잘못된 old password 시도) | `tene audit --json` 결과 검증 |

---

## 4. Success Criteria

### 4.1 Definition of Done

- [ ] FR-01 ~ FR-08 모두 구현 (grep + 테스트 통과)
- [ ] `internal/cli/passwd_test.go` — 3 케이스 통과 (정확/잘못/keychain bypass)
- [ ] `internal/cli/recover_test.go` — 동일 패턴 통과 (s2 의 T2-B8 과 합쳐도 무방)
- [ ] CHANGELOG.md 에 BREAKING 헤더 + migration note
- [ ] gap-detector Match Rate ≥ 90% (M8 gate; design ↔ impl 비교)
- [ ] code-analyzer critical issue = 0 (M3 / S3 gate)
- [ ] `go test -race -count=10 ./internal/cli/... ./pkg/crypto/...` 0 flaky

### 4.2 Quality Criteria

- [ ] passwd 명령 단위 테스트 coverage ≥ 90% (Trk-A 핵심)
- [ ] subtle 패키지 사용 — `crypto/subtle` import 명시
- [ ] `TENE_MASTER_PASSWORD` env var path 도 verify 통과 (E2E test 1건)
- [ ] Audit log 에 `vault.passwd_failed` 기록 (잘못된 password 시도)

---

## 5. Risks and Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| 기존 v1.0.x vault 가 auth_hash 부재 → passwd 명령 break | H | H | FR-07 grace path; 1회 backfill prompt + warning |
| subtle.ConstantTimeCompare 길이 mismatch panic | M | L | 32 bytes fixed length 강제; `len()` 검증 후 비교 |
| keychain bypass 가 daily UX 회귀로 오인됨 | M | M | passwd/recover **한정** — get/run/list/set 은 unchanged |
| HKDF info 문자열 도메인 분리 (sibling crypto-v2-keys 와 충돌) | L | M | `pkg/crypto/info.go` 에 `PurposeAuthHash = "tene/auth-hash/v1"` 정의 — crypto-v2-keys 와 통합 |
| audit_log 에 `vault.passwd_failed` 기록 시 평문 컬럼 사용 (P0-V1 와 충돌) | M | M | s2 의 `002_audit_log_v2` 가 encrypted 컬럼 도입 — s1 에선 평문 그대로 (resource_name 만 'master_password' 같은 generic) |

---

## 6. Impact Analysis

### 6.1 Changed Resources

| Resource | Type | Change Description |
|----------|------|--------------------|
| `vault_meta` table (SQLite, key 'auth_hash') | DB schema (KV add) | 신규 키 'auth_hash' 도입 (별도 컬럼 추가 X) |
| `internal/cli/root.go:loadOrPromptMasterKey` | Function refactor | `loadCachedMasterKey()` + `promptAndVerifyMasterKey()` 2 분리 |
| `internal/cli/passwd.go:runPasswd` | Function | line 30-31 교체 (keychain bypass + verify gate) |
| `internal/cli/recover.go:runRecover` | Function | 동일 패턴 적용 |
| `internal/cli/init.go:initCmd RunE` | Function | line ~150 (vault 생성 후) `auth_hash` 저장 추가 |
| `pkg/crypto/kdf.go` | Constant | `PurposeAuthHash = "tene/auth-hash/v1"` (or in `pkg/crypto/info.go`) |
| `teneerr.ErrInvalidPassword` | Exit code | 현재 exit 2; `audit-reader` plan 의 새 표에서 exit 4 (AUTH_FAILED) |

### 6.2 Current Consumers

| Resource | Operation | Code Path | Impact |
|----------|-----------|-----------|--------|
| `loadOrPromptMasterKey` | READ keychain → return key | `internal/cli/passwd.go:31`, `recover.go:?`, `get.go`, `set.go`, `list.go`, `run.go`, `export.go`, `import_cmd.go` | **Breaking** (passwd/recover 만) — get/run/list 등은 `loadCachedMasterKey()` 로 직접 호출하도록 변경 (semantic 동일) |
| `vault_meta` KV | READ kdf_salt, recovery_blob, sync_status | `internal/cli/root.go:202` + 다수 | None (auth_hash 는 신규 키이므로 기존 read 영향 없음) |
| `ErrInvalidPassword` | Error code | grep 결과 7 사이트 | None (의미 동일; exit code 표는 audit-reader plan 에서 조정) |

### 6.3 Verification

- [ ] 모든 8 명령 (`passwd`, `recover`, `get`, `set`, `list`, `run`, `export`, `import_cmd`) 이 새 함수 (`loadCachedMasterKey` or `promptAndVerifyMasterKey`) 중 하나 호출 — grep 검증
- [ ] passwd/recover 만 verify gate 통과 — 나머지는 caching 유지 (회귀 0)
- [ ] integration test: keychain 활성 상태에서 `tene get` 호출 → prompt 없음 (UX 회귀 0)
- [ ] integration test: keychain 활성 상태에서 `tene passwd` 호출 → prompt 표시 (보안 fix 동작)

---

## 7. Architecture Considerations

### 7.1 Project Level Selection

| Level | Characteristics | Recommended For | Selected |
|-------|-----------------|-----------------|:--------:|
| **Starter** | Simple structure | Static sites | ☐ |
| **Dynamic** | Feature-based modules | Web apps with backend | ☐ |
| **Enterprise** | Strict layer separation, DI, microservices | High-traffic, complex architectures | **☑** |

tene CLI 는 **Enterprise**: `cmd/tene` (entry) + `internal/cli` (presentation) + `internal/vault` (infrastructure) + `pkg/crypto` (domain) + `pkg/errors` (cross-cutting). Clean Architecture 의 4-layer.

### 7.2 Key Architectural Decisions

| Decision | Options | Selected | Rationale |
|----------|---------|----------|-----------|
| auth_hash 저장 위치 | (a) `vault_meta` KV table, (b) 신규 `vault_auth` table, (c) `vault_secrets` 특수 row | (a) `vault_meta` KV | 기존 패턴 일관; ALTER 불필요; schema_migrations 없는 현재도 동작 |
| auth_hash 계산 방식 | (a) `HKDF-Expand(masterKey, "tene-auth-hash-v1", 32)`, (b) `Argon2id(password, salt)` (별도 KDF), (c) `SHA-256(masterKey)` | (a) HKDF-Expand | masterKey 가 이미 Argon2id 결과 → HKDF 1단계로 충분; RFC 5869 표준; 새 Argon 비용 없음 |
| Verify gate 위치 | (a) `runPasswd` 함수 내부, (b) Cobra `PreRunE` middleware, (c) 별도 `verifyOldPassword()` helper | (c) helper 분리 | 테스트 용이; recover 도 동일 helper 재사용; Single Responsibility |
| Constant-time compare | (a) `bytes.Equal`, (b) `subtle.ConstantTimeCompare` | (b) subtle | timing leak 방지 표준 라이브러리 |
| Recovery 명령의 verify 방식 | (a) mnemonic 만 검증 (현재), (b) mnemonic + new password verify | (b) mnemonic + new auth_hash | recovery 도 master 회전이므로 동일 보안 모델 |

### 7.3 Clean Architecture Approach

```
Selected Level: Enterprise

Layer mapping:
┌──────────────────────────────────────────────────────────┐
│ Presentation (internal/cli/)                             │
│   - passwd.go (runPasswd RunE)                           │
│   - recover.go (runRecover RunE)                         │
│   - root.go (loadCachedMasterKey, promptAndVerifyMasterKey) │
│   - init.go (auth_hash initial save)                     │
├──────────────────────────────────────────────────────────┤
│ Application (없음 — 현재 cli 가 use case 도 겸함;          │
│              s2 의 architecture sprint 에서 분리 예정)     │
├──────────────────────────────────────────────────────────┤
│ Domain (pkg/crypto/)                                     │
│   - HKDF-Expand 함수 (이미 존재)                          │
│   - PurposeAuthHash constant                             │
│   - subtle.ConstantTimeCompare 사용                       │
├──────────────────────────────────────────────────────────┤
│ Infrastructure (internal/vault/)                         │
│   - vault_meta KV table (key='auth_hash')                │
│   - SetMeta/GetMeta (이미 존재)                           │
├──────────────────────────────────────────────────────────┤
│ Cross-cutting (pkg/errors/)                              │
│   - ErrInvalidPassword (이미 존재)                        │
└──────────────────────────────────────────────────────────┘
```

---

## 8. Convention Prerequisites

### 8.1 Existing Project Conventions

- [x] `CLAUDE.md` (project root) — Go 1.25, Clean Architecture, internal/* 패키지 layout
- [x] `tene/.golangci.yml` — 6 linter (s3 에서 15 로 확장)
- [x] Go module: `github.com/agent-kay-it/tene`
- [x] Error handling: `pkg/errors` (s4 에서 `teneerr` rename)
- [x] Testing: `go test -race` + `testhelper_test.go` (s2 에서 리팩토)

### 8.2 Conventions to Define/Verify

| Category | Current State | To Define | Priority |
|----------|---------------|-----------|:--------:|
| HKDF info 도메인 명명 | 산발적 (`"tene-audit"`, `"tene/sync"`) | `tene/{domain}/v{N}` 표준 (sibling crypto-v2-keys plan 와 통합) | High |
| Vault meta key 명명 | `kdf_salt`, `recovery_blob`, `sync_status` | `auth_hash` (snake_case) — 기존 패턴 따름 | Medium |
| Exit code (auth 그룹) | exit 2 (현재) | exit 4 (AUTH_FAILED — `audit-reader` plan 의 표) | High |

### 8.3 Environment Variables Needed

| Variable | Purpose | Scope | To Be Created |
|----------|---------|-------|:-------------:|
| `TENE_MASTER_PASSWORD` | Non-interactive automation (CI 등) | Server/CI | (이미 존재) — verify gate 통과 강제 |
| `TENE_ACTOR_ID` | audit_log actor 식별 (human/ai/automation) | Optional | (`audit-reader` plan 에서 도입) |

---

## 9. Testing Plan

### 9.1 Test Scope (s1 의 M2/M3/M6 게이트)

| Type | Target | Tool | Phase |
|------|--------|------|-------|
| L1: Unit | `loadCachedMasterKey`, `promptAndVerifyMasterKey`, `runPasswd`, `runRecover` | `go test` | Do |
| L2: CLI 통합 | `tene passwd` end-to-end (interactive prompt mock) | testhelper_test.go | Do |
| L3: Regression | 기존 8 명령 (`get/set/list/run` 등) keychain 캐시 동작 변화 없음 | testhelper_test.go | Do |

### 9.2 Test Scenarios

| # | Scenario | Expected |
|---|----------|---------|
| T-1 | 정확한 old password 입력 → master 회전 성공 | exit 0, vault_meta.auth_hash 갱신 |
| T-2 | 잘못된 old password 입력 → `ErrInvalidPassword` | exit 4 (AUTH_FAILED), audit_log 에 `vault.passwd_failed` 기록 |
| T-3 | keychain 캐시 활성 + 빈 password 입력 → fail | exit 4, prompt 표시됨 (keychain bypass 증거) |
| T-4 | `TENE_MASTER_PASSWORD=wrong` env → fail | exit 4 (env path 도 verify) |
| T-5 | 기존 v1.0.x vault (auth_hash 부재) → grace path | warn 출력 + 1회 prompt + auth_hash 저장 + 진행 |
| T-6 | 정상 명령 (`tene get FOO`) — keychain 캐시 활성 | prompt 없음 (UX 회귀 0) |
| T-7 | recover 명령 동일 패턴 | recover_test.go (s2 T2-B8 과 합쳐도 무방) |

---

## 10. Implementation Notes (Pre-Design Hints)

> 본 섹션은 Plan 단계의 hint. 정식 구현 가이드는 Design 문서 §11 참조.

- `subtle.ConstantTimeCompare(a, b []byte) int` 반환 1=match, 0=mismatch (Go 표준)
- HKDF: `pkg/crypto/kdf.go` 의 `DeriveSubKey` 가 이미 HKDF-Expand 기반 (sibling crypto-v2-keys plan 이 시그니처 변경; 본 plan 은 그 후속)
- `vault_meta` 테이블 schema: `CREATE TABLE vault_meta (key TEXT PRIMARY KEY, value TEXT)` (현재); SetMeta/GetMeta 함수 이미 존재
- passwd 명령의 `defer crypto.ZeroBytes` 패턴 유지 (auth_hash 비교 후 oldEncKey 즉시 zeroize)

---

## 11. Next Steps

1. [x] Plan 문서 작성 완료 (본 문서)
2. [ ] Design 문서 작성 — [`docs/02-design/features/passwd-verify.design.md`](../../02-design/features/passwd-verify.design.md) — security-architect 위임
3. [ ] L2 boundary 도달: design 후 사용자 승인 필요 (do phase 진입)
4. [ ] Sprint 1 PR #5 + PR #6 (참조: master plan §9.1 PR matrix)
5. [ ] Sprint 2 의 `vault-v2-migration` 에서 `001_v2_envelope` migration 으로 기존 v1 vault 자동 backfill

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-05-13 | Initial draft (Sprint 1 plan phase, L2 boundary) | cto-lead |
