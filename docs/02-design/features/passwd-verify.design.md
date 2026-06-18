---
template: design
version: 1.3
description: tene CLI v2.0 Sprint 1 — passwd-verify (P0-P1) Design Document
variables:
  - feature: passwd-verify
  - sprint: tene-cli-v2-s1
  - date: 2026-05-13
  - author: cto-lead (security-architect perspective)
  - project: tene CLI
  - version: v2.0.0 (target)
  - planPath: docs/01-plan/features/passwd-verify.plan.md
  - trustLevel: L2
---

# passwd-verify Design Document

> **Summary**: `vault_meta` KV 에 `auth_hash` (32-byte HKDF-Expand) 도입 + `loadOrPromptMasterKey` 를 `loadCachedMasterKey()` + `promptAndVerifyMasterKey()` 2 분리. passwd/recover 는 항상 후자 호출 → keychain bypass + ConstantTimeCompare gate.
>
> **Project**: tene CLI
> **Version**: target v2.0.0
> **Sprint**: tene-cli-v2-s1 (W1-W2)
> **Author**: cto-lead (security-architect perspective)
> **Date**: 2026-05-13
> **Status**: Draft (L2)
> **Planning Doc**: [passwd-verify.plan.md](../../01-plan/features/passwd-verify.plan.md)

### Pipeline References

| Phase | Document | Status |
|-------|----------|--------|
| Phase 1 (Schema) | Inline (§3) — `vault_meta.auth_hash` KV row | N/A separate doc |
| Phase 2 (Convention) | `pkg/crypto` 에 `PurposeAuthHash` constant | N/A separate doc |

---

## Context Anchor

> Copied from Plan document. Ensures strategic context survives Design→Do handoff.

| Key | Value |
|-----|-------|
| **WHY** | `tene passwd` / `tene recover` 가 old password 검증을 안 함 → 노트북 unlock 상태에서 1줄 명령으로 master 회전 가능 (P0 보안 결함). |
| **WHO** | 모든 tene 사용자. 특히 **AI-vibe coder (50%)** 와 **Sec-conscious team (15%)**. |
| **RISK** | (a) auth_hash 컬럼 부재 v1 vault → grace path 필요 (s2 backfill). (b) UX 회귀 우려 → passwd/recover 한정. (c) ConstantTimeCompare 잘못 사용 시 timing leak → 32 bytes fixed. |
| **SUCCESS** | grep `auth_hash` ≥ 2; passwd_test.go 3 케이스 통과; `tene passwd` 가 keychain 캐시 있어도 prompt 표시. |
| **SCOPE** | Sprint 1 Trk-A 핫픽스 2 PR (#5 + #6). +240 LOC. internal/cli + internal/vault + pkg/crypto. |

---

## 1. Overview

### 1.1 Design Goals

1. **Verify gate 강제**: passwd/recover RunE 가 keychain 캐시 우회 + `term.ReadPassword` + `subtle.ConstantTimeCompare`
2. **UX 회귀 0**: get/run/list/set/export 등 일상 명령은 keychain 캐시 그대로
3. **Backward compat**: 기존 v1 vault (auth_hash 부재) 에서 grace path (1회 backfill prompt)
4. **Testability**: `promptAndVerifyMasterKey()` helper 분리로 unit test 용이
5. **Timing-safe**: `crypto/subtle.ConstantTimeCompare` 명시 사용

### 1.2 Design Principles

- **Defense in depth**: keychain bypass + ReadPassword + ConstantTimeCompare 3 단계
- **Least privilege**: verify gate 는 passwd/recover 만; 일상 명령은 회귀 없음
- **Explicit over implicit**: helper 이름 (`promptAndVerifyMasterKey`) 이 의도 명시
- **Auditability**: 잘못된 password 시도는 audit_log 에 `vault.passwd_failed` 기록

---

## 2. Architecture Options

### 2.0 Architecture Comparison

| Criteria | Option A: Minimal Wrapper | Option B: Helper Split | Option C: Middleware (Cobra PreRunE) |
|----------|:-:|:-:|:-:|
| **Approach** | passwd 함수 안에 verify 인라인 | `loadOrPromptMasterKey` 를 2 함수로 분리 | `cobra.Command.PreRunE` 에 verify decorator 등록 |
| **New Files** | 0 (passwd.go 만 수정) | 0 (root.go 함수 추가) | 1 (`internal/cli/middleware.go`) |
| **Modified Files** | 2 (passwd.go, recover.go) | 3 (root.go, passwd.go, recover.go) | 4 (root.go, passwd.go, recover.go, middleware.go) |
| **Complexity** | Low | Medium | High |
| **Maintainability** | Low (코드 중복 — passwd 와 recover) | High (DRY; single source of truth) | Medium (Cobra middleware abstraction overhead) |
| **Effort** | Low | Medium | High |
| **Testability** | Medium (passwd 단위 테스트만) | **High** (helper 단위 테스트 + integration) | Medium (middleware mock 복잡) |
| **Risk** | Medium (passwd/recover 중복 코드 drift) | Low | Medium (Cobra PreRunE 가 모든 subcommand 영향 — 검증 범위 넓음) |
| **Recommendation** | — | **Default choice** | Future (v2.1+ 의 통일된 auth 미들웨어) |

**Selected**: **Option B — Helper Split** — **Rationale**: passwd 와 recover 가 동일 verify gate 필요 → DRY 원칙. helper 분리가 unit test 가장 쉽고, root.go 의 기존 함수 (`loadOrPromptMasterKey`) 와 명명 일관성 유지. Option A 의 코드 중복은 향후 drift 위험; Option C 는 v2.1 의 다중 명령 통합 시점까지 보류.

### 2.1 Component Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                  internal/cli/passwd.go (runPasswd)                 │
│                          (verify gate enforced)                     │
└───────────┬────────────────────────────────────────┬───────────────┘
            │                                        │
            ▼                                        ▼
┌────────────────────────────┐         ┌───────────────────────────┐
│ promptAndVerifyMasterKey() │         │ loadCachedMasterKey()     │
│   (passwd/recover only)    │         │   (get/set/list/run/...)  │
└────────────┬───────────────┘         └───────────┬───────────────┘
             │                                     │
             │  1. term.ReadPassword               │  1. keychain.Load()
             │  2. deriveMasterKeyFromPassword     │  2. env TENE_MASTER_PASSWORD
             │  3. HKDF-Expand → authHashCandidate │  3. (env path) verify? NO
             │  4. subtle.ConstantTimeCompare      │     — env 우회 가능
             │     (vault_meta.auth_hash)          │
             │  5. mismatch → ErrInvalidPassword   │
             │     + AddAuditLog("vault.passwd_failed")
             ▼
        ┌────────────────┐
        │   *App.Vault   │ (internal/vault)
        │   GetMeta('auth_hash') / SetMeta
        └────────────────┘
                 │
                 ▼
        ┌────────────────┐
        │  pkg/crypto    │
        │  HKDF-Expand   │ (golang.org/x/crypto/hkdf)
        │  PurposeAuthHash = "tene/auth-hash/v1"
        └────────────────┘
```

### 2.2 Data Flow — passwd 명령

```
User runs: tene passwd
    │
    ▼
runPasswd(cmd, args)
    │
    ├─ isTerminal() check → ErrInteractiveRequired if non-TTY (TENE_MASTER_PASSWORD env path 도 verify)
    │
    ├─ loadApp() → *App (vault open, keychain ready, env resolved)
    │
    ├─ promptAndVerifyMasterKey(app, "current") → oldMasterKey
    │     │
    │     ├─ fmt.Fprint(os.Stderr, "Enter current Master Password: ")
    │     ├─ password ← term.ReadPassword(os.Stdin.Fd())
    │     ├─ saltB64 ← app.Vault.GetMeta("kdf_salt")
    │     ├─ salt ← base64.StdEncoding.DecodeString(saltB64)
    │     ├─ candidateMasterKey ← crypto.DeriveKey(password, salt)  // Argon2id
    │     ├─ candidateAuthHash ← hkdf.Expand(sha256, candidateMasterKey, "tene/auth-hash/v1", 32)
    │     ├─ storedAuthHashB64 ← app.Vault.GetMeta("auth_hash")
    │     │     │
    │     │     ├─ found → continue
    │     │     └─ NotFound → grace path (FR-07; backfill prompt + warn)
    │     │
    │     ├─ storedAuthHash ← base64.StdEncoding.DecodeString(storedAuthHashB64)
    │     ├─ if subtle.ConstantTimeCompare(candidateAuthHash, storedAuthHash) != 1:
    │     │     ├─ app.Vault.AddAuditLog("vault.passwd_failed", "", "", actor)
    │     │     └─ return ErrInvalidPassword (exit 4)
    │     └─ return candidateMasterKey
    │
    ├─ (existing flow unchanged from passwd.go:43+) — prompt new password, re-encrypt, etc.
    │
    └─ (NEW) on success:
           newMasterKey → newAuthHash ← hkdf.Expand(sha256, newMasterKey, "tene/auth-hash/v1", 32)
           app.Vault.SetMeta("auth_hash", base64(newAuthHash))
           app.Vault.AddAuditLog("vault.passwd_changed", "", "", actor)
```

### 2.3 Dependencies

| Component | Depends On | Purpose |
|-----------|-----------|---------|
| `internal/cli/passwd.go:runPasswd` | `promptAndVerifyMasterKey()` | Verify gate |
| `internal/cli/recover.go:runRecover` | `promptAndVerifyMasterKey()` | Verify gate |
| `internal/cli/root.go:promptAndVerifyMasterKey` | `internal/vault.Vault.GetMeta`, `pkg/crypto.DeriveKey`, `hkdf.Expand` | KDF + auth_hash compare |
| `internal/cli/get.go/set.go/list.go/...` | `loadCachedMasterKey()` (renamed `loadOrPromptMasterKey`) | Existing keychain cache |
| `pkg/crypto/info.go` (or kdf.go) | (none — constant only) | `PurposeAuthHash = "tene/auth-hash/v1"` |

---

## 3. Data Model

### 3.1 Entity Definition

`audit_log` rows (existing schema; this design just adds new actions + new `actor` column managed by sibling `audit-reader`):

```go
// internal/vault/types.go (existing)
type AuditEntry struct {
    ID           int64
    Timestamp    int64  // Unix seconds
    Action       string // "vault.passwd_failed" (NEW), "vault.passwd_changed", ...
    ResourceName string // "" for vault-level events
    Details      string // "" for vault-level events
    Actor        string // 'human' (default) or TENE_ACTOR_ID
}
```

### 3.2 Entity Relationships

```
[Vault] 1 ──── 1 [vault_meta KV: auth_hash]
   │
   └── 1 ──── N [audit_log entries]
```

### 3.3 Database Schema (vault_meta KV — already exists)

```sql
-- internal/vault/schema.go (already exists, no migration needed)
CREATE TABLE IF NOT EXISTS vault_meta (
    key   TEXT PRIMARY KEY,
    value TEXT NOT NULL
)

-- New rows (logical, not schema change):
--   key='auth_hash', value=<base64-encoded 32 bytes>
--   key='kdf_salt' (already exists), value=<base64-encoded 32 bytes>
--   key='recovery_blob' (already exists)
--   key='sync_status' (already exists)
```

**No ALTER TABLE required** — `vault_meta` is a KV table; new keys add rows.

---

## 4. API Specification

### 4.1 New Public Functions

#### `loadCachedMasterKey(app *App) ([]byte, error)` (renamed from `loadOrPromptMasterKey`)

```go
// internal/cli/root.go (replaces old loadOrPromptMasterKey)
//
// Returns the master key from cache (keychain) or via env var.
// Falls back to interactive prompt IF no cache and no env.
// Used by: get, set, list, run, export, import, push, pull, etc.
//
// SECURITY: this function does NOT verify the password against auth_hash.
// For passwd/recover commands, use promptAndVerifyMasterKey instead.
func loadCachedMasterKey(app *App) ([]byte, error) {
    key, err := app.Keychain.Load()
    if err == nil {
        return key, nil
    }
    if pw := os.Getenv("TENE_MASTER_PASSWORD"); pw != "" {
        return deriveMasterKeyFromPassword(app, pw)
    }
    if !isTerminal() {
        return nil, teneerr.ErrInteractiveRequired
    }
    fmt.Fprint(os.Stderr, "Enter Master Password: ")
    password, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Fprintln(os.Stderr)
    if err != nil {
        return nil, fmt.Errorf("failed to read password: %w", err)
    }
    return deriveMasterKeyFromPassword(app, string(password))
}
```

#### `promptAndVerifyMasterKey(app *App, label string) ([]byte, error)` (NEW)

```go
// internal/cli/root.go (NEW)
//
// Prompts the user for the master password (or uses TENE_MASTER_PASSWORD env)
// and VERIFIES it against vault_meta.auth_hash before returning.
// Used by: passwd, recover.
//
// SECURITY:
//   - Always bypasses keychain cache (forces user to know the password)
//   - Uses subtle.ConstantTimeCompare (no timing leak)
//   - Logs `vault.passwd_failed` audit entry on mismatch
//   - Supports v1 vault grace path (auth_hash missing → 1-time backfill)
//
// label: "current" or "old" — used in prompt string ("Enter {label} Master Password: ")
func promptAndVerifyMasterKey(app *App, label string) ([]byte, error) {
    // 1. Read password (env or interactive)
    password, err := readMasterPassword(app, label)
    if err != nil {
        return nil, err
    }

    // 2. Derive candidate master key (Argon2id)
    saltB64, err := app.Vault.GetMeta("kdf_salt")
    if err != nil {
        return nil, fmt.Errorf("read kdf_salt: %w", err)
    }
    salt, err := decodeBase64(saltB64)
    if err != nil {
        return nil, fmt.Errorf("decode kdf_salt: %w", err)
    }
    candidateMasterKey, err := crypto.DeriveKey(password, salt)
    if err != nil {
        return nil, fmt.Errorf("derive master key: %w", err)
    }

    // 3. Compute candidate auth_hash
    candidateAuthHash, err := crypto.DeriveAuthHash(candidateMasterKey)
    if err != nil {
        crypto.ZeroBytes(candidateMasterKey)
        return nil, fmt.Errorf("compute auth_hash: %w", err)
    }
    defer crypto.ZeroBytes(candidateAuthHash)

    // 4. Load stored auth_hash
    storedAuthHashB64, err := app.Vault.GetMeta("auth_hash")
    if err != nil {
        if errors.Is(err, vault.ErrMetaNotFound) {
            // FR-07: Grace path — v1 vault, no auth_hash yet
            return graceBackfillAuthHash(app, candidateMasterKey, candidateAuthHash)
        }
        crypto.ZeroBytes(candidateMasterKey)
        return nil, fmt.Errorf("read auth_hash: %w", err)
    }
    storedAuthHash, err := decodeBase64(storedAuthHashB64)
    if err != nil {
        crypto.ZeroBytes(candidateMasterKey)
        return nil, fmt.Errorf("decode auth_hash: %w", err)
    }

    // 5. Constant-time compare
    if subtle.ConstantTimeCompare(candidateAuthHash, storedAuthHash) != 1 {
        // Audit log failure
        actor := os.Getenv("TENE_ACTOR_ID")
        if actor == "" {
            actor = "human"
        }
        _ = app.Vault.AddAuditLog("vault.passwd_failed", "", "", actor)
        crypto.ZeroBytes(candidateMasterKey)
        return nil, teneerr.ErrInvalidPassword
    }

    return candidateMasterKey, nil
}
```

#### `readMasterPassword(app *App, label string) (string, error)` (helper)

```go
// internal/cli/root.go (NEW; private)
func readMasterPassword(app *App, label string) (string, error) {
    // Env var path (also goes through verify gate)
    if pw := os.Getenv("TENE_MASTER_PASSWORD"); pw != "" {
        return pw, nil
    }
    // Interactive prompt
    if !isTerminal() {
        return "", teneerr.ErrInteractiveRequired
    }
    fmt.Fprintf(os.Stderr, "Enter %s Master Password: ", label)
    pwBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
    fmt.Fprintln(os.Stderr)
    if err != nil {
        return "", fmt.Errorf("failed to read password: %w", err)
    }
    return string(pwBytes), nil
}
```

#### `graceBackfillAuthHash(...)` (helper for FR-07)

```go
// internal/cli/root.go (NEW; private)
//
// FR-07: For v1 vaults (no auth_hash in vault_meta), warn user once and
// backfill the auth_hash atomically. This trusts the candidate password
// (no prior auth_hash to compare against) — same risk as upgrading any
// legacy auth system to a new hash scheme.
func graceBackfillAuthHash(app *App, masterKey, authHash []byte) ([]byte, error) {
    fmt.Fprintln(os.Stderr,
        "warning: auth_hash backfill required for legacy vault (v1.0.x).")
    fmt.Fprintln(os.Stderr,
        "         storing computed auth_hash now; future passwd/recover will verify against it.")
    if err := app.Vault.SetMeta("auth_hash", encodeBase64(authHash)); err != nil {
        crypto.ZeroBytes(masterKey)
        return nil, fmt.Errorf("backfill auth_hash: %w", err)
    }
    return masterKey, nil
}
```

#### `crypto.DeriveAuthHash(masterKey []byte) ([]byte, error)` (NEW in pkg/crypto)

```go
// pkg/crypto/auth.go (NEW)
//
// DeriveAuthHash computes a 32-byte HKDF-Expand of the master key using
// the constant info string "tene/auth-hash/v1". Used by vault_meta.auth_hash
// for passwd/recover verification (see internal/cli/root.go).
//
// This is a one-way derivation — given a master key, the auth_hash is
// deterministic. Compared with subtle.ConstantTimeCompare.
//
// Note: sibling feature `crypto-v2-keys` introduces DeriveSubKeyV2 with
// explicit salt parameter. This function uses an empty salt (PRK = HMAC(nil, IKM))
// since the master key is already high-entropy (Argon2id output). This is
// acceptable per RFC 5869 §3.1 ("salt is optional").
func DeriveAuthHash(masterKey []byte) ([]byte, error) {
    const info = "tene/auth-hash/v1"
    reader := hkdf.New(sha256.New, masterKey, nil, []byte(info))
    out := make([]byte, 32)
    if _, err := io.ReadFull(reader, out); err != nil {
        return nil, fmt.Errorf("hkdf expand: %w", err)
    }
    return out, nil
}
```

**Note**: If sibling `crypto-v2-keys` introduces `DeriveSubKeyV2(masterKey, salt, info, length)`, this function could be `DeriveSubKeyV2(masterKey, vaultSalt, "tene/auth-hash/v1", 32)` using the vault's kdf_salt as the explicit HKDF salt. Final decision is at Do phase — Plan/Design propose the standalone version first to avoid blocking on the sibling.

### 4.2 Detailed Specification — Vault.SetMeta / GetMeta (existing)

```go
// internal/vault/vault.go (existing; no change needed)
func (v *Vault) GetMeta(key string) (string, error)
func (v *Vault) SetMeta(key, value string) error
```

`vault_meta` rows used by this design:
- `kdf_salt` (existing) — Argon2id input salt (32 bytes random per vault)
- `auth_hash` (NEW) — HKDF-Expand of master key (32 bytes)

---

## 5. UI/UX Design

### 5.1 CLI Output — `tene passwd` (success case)

```
$ tene passwd
Enter current Master Password: ********
Enter new Master Password: ********
Confirm new Master Password: ********

  Re-encrypting vault...
  17 secrets re-encrypted.
  Master Key updated in OS Keychain.

  New Recovery Key (write this down and keep it safe!):
  +--------------------------------------------------+
  |   word1 word2 word3 word4 word5 word6           |
  |   word7 word8 word9 word10 word11 word12        |
  |                                                  |
  |   Your previous Recovery Key is now invalid.     |
  +--------------------------------------------------+

  Master Password changed successfully.
$
```

### 5.2 CLI Output — `tene passwd` (failure — wrong old password)

```
$ tene passwd
Enter current Master Password: ********
Error: invalid password
Exit code: 4 (AUTH_FAILED)
$
```

### 5.3 CLI Output — `tene passwd` (grace path — v1 vault)

```
$ tene passwd
Enter current Master Password: ********
warning: auth_hash backfill required for legacy vault (v1.0.x).
         storing computed auth_hash now; future passwd/recover will verify against it.
Enter new Master Password: ********
... (rest of flow)
$
```

### 5.4 Page UI Checklist (CLI semantics)

#### `tene passwd` command output

- [ ] Prompt: "Enter current Master Password: " on stderr (always — keychain bypass)
- [ ] Wrong password → "Error: invalid password" on stderr + exit 4
- [ ] Wrong password → audit_log row with action="vault.passwd_failed"
- [ ] Correct password → existing flow (re-encrypt + new recovery key)
- [ ] On success → vault_meta.auth_hash updated to new master key's HKDF
- [ ] On success → audit_log row with action="vault.passwd_changed"
- [ ] Grace path (v1 vault) → stderr warning + auth_hash backfill + continue
- [ ] Non-TTY + no env TENE_MASTER_PASSWORD → exit 7 (INTERACTIVE_REQUIRED)
- [ ] Non-TTY + TENE_MASTER_PASSWORD env → verify gate still applies (env path not bypass)

#### `tene recover` command output

- [ ] Prompt: "Enter recovery mnemonic (12-24 words): "
- [ ] Mnemonic verification (existing flow)
- [ ] After mnemonic verify → "Enter new Master Password: " (FR-05)
- [ ] On success → vault_meta.auth_hash updated to new master key's HKDF
- [ ] On success → audit_log "vault.recovered" entry

---

## 6. Error Handling

### 6.1 Error Code Definition

| Code | Name | Cause | Handling |
|------|------|-------|----------|
| 4 | `ErrInvalidPassword` (AUTH_FAILED) | auth_hash mismatch | Exit 4 + audit_log "vault.passwd_failed" |
| 7 | `ErrInteractiveRequired` | Non-TTY + no env var | Exit 7 + stderr message |
| 1 | `ErrGeneral` | Internal error (HKDF/Argon2id failure) | Exit 1 + log error |

> Exit code numbering follows sibling `audit-reader` plan's new exit code table.

### 6.2 Error Response Format

```
# Wrong password
$ tene passwd
Enter current Master Password: ********
Error: invalid password
$ echo $?
4

# Non-interactive without env
$ tene passwd < /dev/null
Error: interactive prompt required (set TENE_MASTER_PASSWORD or run in a terminal)
$ echo $?
7
```

---

## 7. Security Considerations

- [x] **Constant-time comparison**: `crypto/subtle.ConstantTimeCompare` — no timing leak
- [x] **Fixed length**: auth_hash always 32 bytes; mismatch length → exit 4 (defense in depth)
- [x] **Memory zeroing**: `crypto.ZeroBytes` on candidateMasterKey when verify fails
- [x] **No password in logs**: audit_log records "vault.passwd_failed" without password content
- [x] **Audit trail**: every failed attempt logged for forensics (sibling `audit-reader` enables read)
- [x] **HKDF salt choice**: empty salt is RFC 5869 §3.1 compliant when IKM is high-entropy (Argon2id output)
- [x] **Keychain bypass invariant**: `promptAndVerifyMasterKey` MUST NOT call `keychain.Load()` — enforced by code review checklist
- [x] **Env var path verify**: `TENE_MASTER_PASSWORD` also goes through ConstantTimeCompare (FR-06)
- [x] **Grace path acceptance**: v1 vault backfill trusts the candidate password (same as any auth scheme upgrade); user is warned on stderr

---

## 8. Test Plan

### 8.1 Test Scope

| Type | Target | Tool | Phase |
|------|--------|------|-------|
| L1: Unit | `promptAndVerifyMasterKey`, `DeriveAuthHash`, `graceBackfillAuthHash` | `go test` | Do |
| L2: CLI | `tene passwd` end-to-end (stdin mock) | testhelper_test.go | Do |
| L3: Regression | `tene get/set/list/run` keychain cache unaffected | testhelper_test.go | Do |

### 8.2 L1: Unit Test Scenarios

| # | Target | Test | Expected |
|---|--------|------|---------|
| 1 | `DeriveAuthHash` | RFC 5869 §A.1 IKM → expected OKM | hex match |
| 2 | `DeriveAuthHash` | Same masterKey 2 calls | same output (deterministic) |
| 3 | `DeriveAuthHash` | Different masterKey | different output |
| 4 | `promptAndVerifyMasterKey` | correct password (mock stdin + vault) | masterKey returned |
| 5 | `promptAndVerifyMasterKey` | wrong password | `ErrInvalidPassword` returned |
| 6 | `promptAndVerifyMasterKey` | wrong password — audit_log written | `vault.passwd_failed` row |
| 7 | `promptAndVerifyMasterKey` | v1 vault (no auth_hash) | grace path — warning + backfill |
| 8 | `promptAndVerifyMasterKey` | non-TTY + no env | `ErrInteractiveRequired` |
| 9 | `promptAndVerifyMasterKey` | TENE_MASTER_PASSWORD env (correct) | masterKey returned |
| 10 | `promptAndVerifyMasterKey` | TENE_MASTER_PASSWORD env (wrong) | `ErrInvalidPassword` |

### 8.3 L2: CLI Test Scenarios

| # | Command | Setup | Expected |
|---|---------|-------|----------|
| 1 | `tene passwd` (correct old) | keychain cached, vault has auth_hash | Re-encrypt success, new auth_hash, new recovery key |
| 2 | `tene passwd` (wrong old) | keychain cached, vault has auth_hash | Exit 4, no re-encrypt, audit_log "vault.passwd_failed" |
| 3 | `tene passwd` (v1 vault) | no auth_hash | Warning + backfill + continue |
| 4 | `tene recover` (correct mnemonic + new password) | vault has auth_hash | Re-encrypt, new auth_hash matches new password |

### 8.4 L3: E2E Regression Scenarios

| # | Scenario | Expected |
|---|----------|----------|
| 1 | keychain cached → `tene get FOO` | No prompt (cache hit) — UX unchanged |
| 2 | keychain cached → `tene list` | No prompt — UX unchanged |
| 3 | keychain cached → `tene run -- env` | No prompt — UX unchanged |
| 4 | keychain cached → `tene passwd` | Prompt shown (keychain bypass — security fix) |

### 8.5 Seed Data Requirements

| Entity | Minimum Count | Key Fields |
|--------|:------------:|------------|
| `vault_meta` row 'kdf_salt' | 1 | value = 32-byte base64 |
| `vault_meta` row 'auth_hash' | 1 (post-init) | value = 32-byte base64 |
| `audit_log` | 0 initial; assertions add rows | (auto-incrementing id) |

> Test fixture: `internal/cli/testdata/v2-vault.db` (with auth_hash) + `v1-vault.db` (without auth_hash, for grace path test).

---

## 9. Clean Architecture

### 9.1 Layer Structure

| Layer | Component | Location |
|-------|-----------|----------|
| **Presentation** | `runPasswd`, `runRecover` Cobra RunE | `internal/cli/passwd.go`, `recover.go` |
| **Application** | `promptAndVerifyMasterKey`, `loadCachedMasterKey`, `readMasterPassword`, `graceBackfillAuthHash` | `internal/cli/root.go` |
| **Domain** | `DeriveAuthHash`, `PurposeAuthHash` constant | `pkg/crypto/auth.go`, `info.go` |
| **Infrastructure** | `vault_meta.auth_hash` KV row, `audit_log` table | `internal/vault/vault.go`, `schema.go` |
| **Cross-cutting** | `ErrInvalidPassword`, `ErrInteractiveRequired` | `pkg/errors/codes.go` |

### 9.2 Dependency Rules

```
Presentation (passwd.go) → Application (root.go) → Domain (pkg/crypto)
                                  │
                                  └─→ Infrastructure (internal/vault)
```

### 9.3 File Import Rules

| From | Imports | Forbidden |
|------|---------|-----------|
| `passwd.go` | `root.go` helpers, `pkg/crypto`, `pkg/errors` | direct `internal/vault` SQL |
| `root.go` | `pkg/crypto`, `internal/vault`, `pkg/errors` | none |
| `pkg/crypto/auth.go` | `golang.org/x/crypto/hkdf`, `crypto/sha256`, `io` | `internal/*` (Domain isolation) |
| `internal/vault/vault.go` | `modernc.org/sqlite`, `pkg/crypto` (for HMAC in audit_log v2) | `internal/cli/*` |

### 9.4 This Feature's Layer Assignment

| Component | Layer | Location |
|-----------|-------|----------|
| `runPasswd`, `runRecover` | Presentation | `internal/cli/passwd.go`, `recover.go` |
| `promptAndVerifyMasterKey`, `loadCachedMasterKey` | Application | `internal/cli/root.go` |
| `DeriveAuthHash`, `PurposeAuthHash` | Domain | `pkg/crypto/auth.go` |
| `Vault.GetMeta/SetMeta`, `AddAuditLog` | Infrastructure | `internal/vault/vault.go` |
| `ErrInvalidPassword` | Cross-cutting | `pkg/errors/codes.go` |

---

## 10. Coding Convention Reference

### 10.1 Naming Conventions

| Target | Rule | Example |
|--------|------|---------|
| Helper functions (private) | `camelCase` | `promptAndVerifyMasterKey`, `readMasterPassword`, `graceBackfillAuthHash` |
| Public functions | `PascalCase` | `DeriveAuthHash`, `SetMeta`, `GetMeta` |
| Constants | `PascalCase` (Go convention) | `PurposeAuthHash` |
| HKDF info strings | `tene/{domain}/v{N}` | `"tene/auth-hash/v1"` |
| Audit log actions | `{resource}.{action}` snake-style | `"vault.passwd_failed"`, `"vault.passwd_changed"` |
| Error variables | `Err{Cause}` | `ErrInvalidPassword`, `ErrInteractiveRequired` |
| vault_meta keys | snake_case | `auth_hash`, `kdf_salt`, `recovery_blob` |

### 10.2 Import Order

```go
package cli

import (
    // 1. Standard library
    "crypto/subtle"
    "errors"
    "fmt"
    "os"

    // 2. External
    "github.com/spf13/cobra"
    "golang.org/x/term"

    // 3. Internal
    "github.com/tene-ai/tene/internal/vault"
    "github.com/tene-ai/tene/pkg/crypto"
    teneerr "github.com/tene-ai/tene/pkg/errors"
)
```

### 10.3 Environment Variables

| Variable | Purpose | Scope |
|----------|---------|-------|
| `TENE_MASTER_PASSWORD` | Non-interactive automation (CI) — still goes through verify gate | Optional |
| `TENE_ACTOR_ID` | audit_log actor column (default 'human') | Optional |

---

## 11. Implementation Guide

### 11.1 File Structure

```
tene/
├── internal/
│   ├── cli/
│   │   ├── passwd.go         # MODIFIED (line 30-31: helper call)
│   │   ├── recover.go        # MODIFIED (same pattern)
│   │   ├── root.go           # MODIFIED (new helpers)
│   │   ├── init.go           # MODIFIED (save auth_hash on vault create)
│   │   ├── passwd_test.go    # NEW (3 scenarios)
│   │   ├── root_test.go      # NEW or MODIFIED (helper tests)
│   │   └── testdata/
│   │       ├── v1-vault.db   # NEW fixture (no auth_hash)
│   │       └── v2-vault.db   # NEW fixture (with auth_hash)
│   └── vault/
│       └── vault.go          # MODIFIED (AddAuditLog actor param — interop with audit-reader)
└── pkg/
    ├── crypto/
    │   ├── auth.go           # NEW (DeriveAuthHash)
    │   ├── info.go           # NEW or MODIFIED (PurposeAuthHash constant)
    │   └── auth_test.go      # NEW (KAT-style test for DeriveAuthHash)
    └── errors/
        └── codes.go          # MODIFIED (ErrInvalidPassword exit 4)
```

### 11.2 Implementation Order

> **PR #5** (`feat(vault): auth_hash column + actor column for audit log v2`):

1. [ ] `pkg/crypto/info.go` — `PurposeAuthHash = "tene/auth-hash/v1"`
2. [ ] `pkg/crypto/auth.go` — `DeriveAuthHash(masterKey)` function
3. [ ] `pkg/crypto/auth_test.go` — 3 unit tests (deterministic, different inputs, length=32)
4. [ ] `internal/cli/init.go` — after vault creation, compute auth_hash, call `app.Vault.SetMeta("auth_hash", base64)`
5. [ ] `internal/cli/init_test.go` — verify `SELECT value FROM vault_meta WHERE key='auth_hash'` returns 32-byte base64

> **PR #6** (`feat(cli): passwd/recover prompt + auth_hash verify (P0-P1)`):

6. [ ] `internal/cli/root.go` — split `loadOrPromptMasterKey` → `loadCachedMasterKey` (rename)
7. [ ] `internal/cli/root.go` — add `promptAndVerifyMasterKey`, `readMasterPassword`, `graceBackfillAuthHash`
8. [ ] `internal/cli/passwd.go:30-31` — replace `loadOrPromptMasterKey(app)` with `promptAndVerifyMasterKey(app, "current")`
9. [ ] `internal/cli/passwd.go` (end) — after master rotation, compute new auth_hash + `SetMeta`
10. [ ] `internal/cli/recover.go` — apply same pattern (after mnemonic verify, prompt new password, set new auth_hash)
11. [ ] `internal/cli/passwd_test.go` — 3 scenarios (correct, wrong, v1 grace)
12. [ ] `internal/cli/recover_test.go` — 2 scenarios (correct + new password, wrong mnemonic)
13. [ ] `pkg/errors/codes.go` — `ErrInvalidPassword.Code = 4` (was 2)
14. [ ] CHANGELOG.md — "BREAKING: tene passwd now always prompts for current Master Password (security hotfix). Exit code changed from 2 to 4 for AUTH_FAILED."
15. [ ] `docs/migration/exit-codes.md` (cross-references sibling `audit-reader`)

### 11.3 Session Guide

#### Module Map

| Module | Scope Key | Description | Estimated Turns |
|--------|-----------|-------------|:---------------:|
| Crypto + Schema | `module-1` | `pkg/crypto/auth.go` + `info.go` + `init.go` auth_hash save (PR #5) | 15-20 |
| Verify gate + tests | `module-2` | `root.go` helpers + `passwd.go` + `recover.go` + tests (PR #6) | 25-30 |

#### Recommended Session Plan

| Session | Phase | Scope | Turns |
|---------|-------|-------|:-----:|
| Session 1 | Plan + Design | (this document) | 30-35 (current) |
| Session 2 | Do | `--scope module-1` (PR #5) | 25-30 |
| Session 3 | Do | `--scope module-2` (PR #6) | 35-40 |
| Session 4 | Check + Report | All modules | 25-30 |

---

## 12. Edge Cases & Failure Modes

| # | Scenario | Behavior |
|---|----------|----------|
| 1 | User presses Ctrl-C during password prompt | `term.ReadPassword` returns error → exit 1 |
| 2 | `vault_meta.kdf_salt` corrupted (base64 decode fails) | exit 1 + error message |
| 3 | `vault_meta.auth_hash` exists but length ≠ 32 | exit 1 + "auth_hash corrupted; restore from backup" |
| 4 | `subtle.ConstantTimeCompare` called with mismatched length | Returns 0 (Go stdlib behavior); we treat as mismatch → exit 4 |
| 5 | New password derivation fails (Argon2id panic) | exit 1; no master rotation; vault intact |
| 6 | Race: 2 `tene passwd` simultaneously | SQLite WAL serializes; second call sees new auth_hash; verify fails → exit 4 |
| 7 | Recovery mnemonic correct but auth_hash already exists | Skip grace path; recovery sets new auth_hash for new master |
| 8 | `TENE_ACTOR_ID="../injection"` | audit_log stores literal string (not interpreted; SQL placeholder safe) |

---

## 13. Sequence Diagram — `tene passwd` (mismatch path)

```
User              passwd.go         root.go              pkg/crypto      internal/vault     audit_log
 │                   │                  │                    │                │                │
 │ tene passwd       │                  │                    │                │                │
 ├──────────────────▶│                  │                    │                │                │
 │                   │ promptAndVerify  │                    │                │                │
 │                   ├─────────────────▶│                    │                │                │
 │                   │                  │ readMasterPassword │                │                │
 │                   │                  ├───┐                │                │                │
 │ "wrong-pass"      │                  │   │ term.ReadPassword               │                │
 │◀──────────────────┼──────────────────┤   │                │                │                │
 │                   │                  │◀──┘                │                │                │
 │                   │                  │ GetMeta("kdf_salt")│                │                │
 │                   │                  ├────────────────────┼───────────────▶│                │
 │                   │                  │◀───────────────────┼─── salt ───────┤                │
 │                   │                  │ DeriveKey(pw, salt)│                │                │
 │                   │                  ├───────────────────▶│                │                │
 │                   │                  │◀──── candidateMaster                │                │
 │                   │                  │ DeriveAuthHash     │                │                │
 │                   │                  ├───────────────────▶│                │                │
 │                   │                  │◀──── candidateHash │                │                │
 │                   │                  │ GetMeta("auth_hash")│               │                │
 │                   │                  ├────────────────────┼───────────────▶│                │
 │                   │                  │◀───────────────────┼─── storedHash ─┤                │
 │                   │                  │ ConstantTimeCompare│                │                │
 │                   │                  ├───┐                │                │                │
 │                   │                  │   │ != 1 (mismatch)│                │                │
 │                   │                  │◀──┘                │                │                │
 │                   │                  │ AddAuditLog        │                │                │
 │                   │                  ├────────────────────┼────────────────┼───────────────▶│
 │                   │                  │                    │                │  "vault.passwd_failed"
 │                   │  ErrInvalidPwd   │                    │                │                │
 │                   │◀─────────────────┤                    │                │                │
 │  exit 4           │                  │                    │                │                │
 │◀──────────────────┤                  │                    │                │                │
```

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-05-13 | Initial draft (Sprint 1 design phase, L2 boundary) | cto-lead (security-architect perspective) |
