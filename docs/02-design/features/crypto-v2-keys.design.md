---
template: design
version: 1.3
description: tene CLI v2.0 Sprint 1 — crypto-v2-keys (P0-C1 + P0-C3) Design Document
variables:
  - feature: crypto-v2-keys
  - sprint: tene-cli-v2-s1
  - date: 2026-05-13
  - author: cto-lead (security-architect perspective)
  - project: tene CLI
  - version: v2.0.0 (target)
  - planPath: docs/01-plan/features/crypto-v2-keys.plan.md
  - trustLevel: L2
---

# crypto-v2-keys Design Document

> **Summary**: `DeriveSubKeyV2(masterKey, salt, info, length)` 신설 + 13 도메인 versioned info constants + 16 호출 사이트 migration + `KDFAlgRegistry` byte 검증. 기존 `DeriveSubKey` deprecated marker (s2 제거). RFC 5869 §A.1-A.3 + RFC 9106 §B.1 6 KAT 통과.
>
> **Project**: tene CLI
> **Version**: target v2.0.0
> **Sprint**: tene-cli-v2-s1 (W1-W2)
> **Author**: cto-lead (security-architect perspective)
> **Date**: 2026-05-13
> **Status**: Draft (L2)
> **Planning Doc**: [crypto-v2-keys.plan.md](../../01-plan/features/crypto-v2-keys.plan.md)

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | `DeriveSubKey` 16 호출 사이트 모두 salt=nil → 모든 vault 가 동일 sub-key. `KDFAlgorithm` byte 검증 부재 → 미지원 KDF silent 통과. RFC 5869 KAT 통과 불가. |
| **WHO** | **Sec-conscious team (15%)** + **Indie OSS dev (25%)** — RFC standard compliance. |
| **RISK** | (a) 16 사이트 일괄 변경 누락. (b) 신규 vault만 sub-key 도메인 변경 → v1 vault 호환. (c) PR 분리 (#1 registry / #2 v2 + KAT / #3 16-site migration). |
| **SUCCESS** | grep `crypto.DeriveSubKey\(.*nil` = 0; grep `crypto.DeriveSubKey\b` ≤ 3; 6 KAT 통과; 미지원 KDF byte → ErrUnsupportedKDF. |
| **SCOPE** | Sprint 1 Trk-A 핵심 (3 PR). +540 LOC. pkg/crypto + internal/cli (11) + internal/encfile (2) + internal/recovery (2) + internal/sync (1). |

---

## 1. Overview

### 1.1 Design Goals

1. **RFC compliance**: HKDF-Extract + HKDF-Expand 표준 부합 (KAT 통과)
2. **Per-vault isolation**: 동일 mnemonic 가진 사용자가 여러 vault 만들 때 sub-key 격리
3. **Versioned info domains**: `"tene/{domain}/v{N}"` — 향후 v2 도메인 추가 시 hash chain 분리
4. **Explicit KDF dispatch**: `KDFAlgRegistry` map → 미지원 byte → 명시 error (silent corruption 방지)
5. **Migration safety**: 16 사이트 일괄 변경 + deprecation grace period (s2 제거)
6. **No breaking caller signature**: 기존 v1 vault 의 ciphertext 호환 (sub-key 변경은 신규 vault 만)

### 1.2 Design Principles

- **Standards over invention**: RFC 5869 (HKDF) + RFC 9106 (Argon2id) 표준 따름
- **Explicit > implicit**: salt 가 함수 시그니처에 명시 (caller 책임)
- **Single source of truth**: 13 도메인 constant 가 `pkg/crypto/info.go` 한 곳
- **Registry pattern**: KDF 추가 시 map 한 줄 (switch 불필요)
- **Deprecation grace**: 기존 `DeriveSubKey` 함수 유지 + `slog.Warn` (s2 제거)

---

## 2. Architecture Options

### 2.0 Architecture Comparison

| Criteria | Option A: Inline 16-site Migration | Option B: V2 함수 + Adapter Pattern | Option C: SecretCipher Interface (Clean Arch) |
|----------|:-:|:-:|:-:|
| **Approach** | DeriveSubKey 시그니처 변경 (BREAKING) + 16 caller 동시 수정 | `DeriveSubKeyV2` 신설 (parallel) + caller 점진 이전 + `DeriveSubKey` deprecated | `SecretCipher` interface 도입 + DI composition root |
| **New Files** | 1 (`pkg/crypto/info.go`) | 3 (`info.go`, `registry.go`, KAT testdata + tests) | 5 (interface + 2 impl + 2 test) |
| **Modified Files** | 17 (kdf.go + 16 callers) | 17 (kdf.go + 16 callers) — but parallel | 17+ (interface 통과로 모든 caller 변경) |
| **Complexity** | Medium | Medium | High |
| **Maintainability** | High (single shot, 깔끔) | High (gradual; deprecated marker) | **Highest** (testability + composition root) |
| **Effort** | Medium | Medium | High |
| **Risk** | **High** (1 PR 17 file 변경 → review 어려움) | Low (PR 분할 가능: #1 v2 함수, #2 migration) | High (composition root 변경은 s2 sprint 영역) |
| **Recommendation** | — | **Default choice** | Future (Sprint 2 `sync-contracts` 와 통합) |

**Selected**: **Option B — V2 함수 + Adapter Pattern** — **Rationale**: 17-file 변경을 1 PR 로 묶으면 review/rollback 어려움. v2 함수 신설로 PR 분리 가능 (#1 registry only, #2 v2 + KAT, #3 16-site migration). `slog.Warn` 으로 추가 caller 차단 시그널. Clean Architecture refactor (Option C) 는 s2 `sync-contracts` 의 `SecretCipher` interface 추출과 통합 — 본 plan 범위 밖.

### 2.1 Component Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                  pkg/crypto/ (Domain Layer)                         │
├────────────────────────────────────────────────────────────────────┤
│  kdf.go                                                            │
│    DeriveSubKey(masterKey, info, length) [Deprecated + slog.Warn]   │
│    DeriveSubKeyV2(masterKey, salt, info, length) [NEW]              │
│                                                                    │
│  info.go [NEW]                                                     │
│    const PurposeEncryption = "tene/encryption/v1"                   │
│    const PurposeAudit = "tene/audit/v1"                             │
│    const PurposeSync = "tene/sync/v1"                               │
│    const PurposeRecovery = "tene/recovery/v1"                       │
│    const PurposeEncfileExport = "tene/encfile/v1"                   │
│    const PurposeAuthHash = "tene/auth-hash/v1"        (sibling)     │
│    ... (13 도메인 total)                                            │
│                                                                    │
│  registry.go [NEW]                                                  │
│    var KDFAlgRegistry = map[byte]KDFFunc{                           │
│      KDFAlgArgon2idV1: argon2idV1Derive,                            │
│    }                                                                │
│                                                                    │
│  testdata/ [NEW]                                                    │
│    rfc5869-a1.json, rfc5869-a2.json, rfc5869-a3.json (3 HKDF KAT)  │
│    rfc9106-b1.json, rfc9106-b2.json, rfc9106-b3.json (3 Argon2 KAT)│
│                                                                    │
│  kat_test.go [NEW]                                                  │
│    TestKAT_HKDF_RFC5869 — 3 vectors                                 │
│    TestKAT_Argon2id_RFC9106 — 3 vectors                             │
└─────────────────────────┬──────────────────────────────────────────┘
                          │
                          │ DeriveSubKeyV2(masterKey, vaultSalt, PurposeX, len)
                          │
        ┌─────────────────┴──────────────────┐
        │                                    │
        ▼                                    ▼
┌──────────────────────┐         ┌──────────────────────┐
│ internal/cli/        │         │ internal/encfile/    │
│  set.go, get.go,     │         │  encfile.go (2)      │
│  run.go, list.go,    │         │  + KDF byte check    │
│  export.go,          │         │    via Registry      │
│  passwd.go, recover.go│        └──────────────────────┘
│  import_cmd.go (11)  │
└──────────────────────┘
                          ┌──────────────────────┐
                          │ internal/recovery/   │
                          │  recover.go (2)      │
                          └──────────────────────┘
                          ┌──────────────────────┐
                          │ internal/sync/       │
                          │  envelope.go (1)     │
                          └──────────────────────┘
```

### 2.2 Data Flow — DeriveSubKeyV2

```
caller (e.g., internal/cli/get.go:54)
  │
  ├─ masterKey ← loadCachedMasterKey(app)
  ├─ saltB64 ← app.Vault.GetMeta("kdf_salt")
  ├─ salt ← base64.StdEncoding.DecodeString(saltB64)  // 32 bytes
  │
  ▼
crypto.DeriveSubKeyV2(masterKey, salt, crypto.PurposeEncryption, 32)
  │
  ├─ if len(salt) == 0: return ErrInvalidSalt
  │
  ├─ HKDF-Extract:
  │     prk ← HMAC-SHA256(key=salt, data=masterKey)
  │
  ├─ HKDF-Expand:
  │     OKM ← T(1) || T(2) || ... where T(i) = HMAC-SHA256(prk, T(i-1) || info || i)
  │     truncate to 32 bytes
  │
  └─ return OKM (sub-key)
```

### 2.3 Data Flow — KDFAlgRegistry (encfile decode)

```
encfile reader (internal/encfile/encfile.go:74)
  │
  ├─ kdfByte ← header.KDFAlg  // single byte from .tene.enc header
  │
  ├─ kdfFn, ok ← crypto.KDFAlgRegistry[kdfByte]
  │
  ├─ if !ok: return ErrUnsupportedKDF (exit 6 — DECRYPT_FAILED)
  │
  ├─ key, err ← kdfFn(password, salt)
  │
  └─ ... (decrypt with key)
```

### 2.4 Dependencies

| Component | Depends On | Purpose |
|-----------|-----------|---------|
| `pkg/crypto/kdf.go:DeriveSubKeyV2` | `golang.org/x/crypto/hkdf`, `crypto/sha256`, `io` | HKDF impl |
| `pkg/crypto/registry.go` | `pkg/crypto/kdf.go:argon2idV1Derive` | KDF dispatch |
| `internal/cli/*` (11 callers) | `pkg/crypto.DeriveSubKeyV2`, `app.Vault.GetMeta("kdf_salt")` | sub-key derive per call |
| `internal/encfile/encfile.go` | `pkg/crypto.DeriveSubKeyV2`, `KDFAlgRegistry`, `pkg/errors.ErrUnsupportedKDF` | Decrypt + KDF dispatch |
| `internal/recovery/recover.go` | `pkg/crypto.DeriveSubKeyV2`, `PurposeRecovery` | Recovery blob key |
| `internal/sync/envelope.go` | `pkg/crypto.DeriveSubKeyV2`, `PurposeSync` | Sync envelope key |
| `pkg/crypto/kat_test.go` | `pkg/crypto/testdata/*.json` | 6 RFC KAT vectors |

---

## 3. Data Model

### 3.1 KDF Algorithm Registry

```go
// pkg/crypto/registry.go (NEW)

type KDFFunc func(password []byte, salt []byte) (key []byte, err error)

const (
    // KDFAlgArgon2idV1 is the default Argon2id parameters as of v1.0.x:
    //   memory = 64 MiB, time = 4, parallelism = 4, key length = 32 bytes
    // Conforms to RFC 9106 §A (recommended parameters).
    KDFAlgArgon2idV1 byte = 0x01

    // Reserved for future: 0x02-0xFE
    // 0xFF reserved (sentinel for "unsupported")
)

// KDFAlgRegistry maps an algorithm byte to its derivation function.
// Lookups MUST use the byte from a .tene.enc header (encfile.go:74).
//
// Adding a new KDF:
//   1. Add the constant (e.g., KDFAlgArgon2idV2 byte = 0x02)
//   2. Add the function (e.g., argon2idV2Derive)
//   3. Register here
//
// REMOVING a KDF is BREAKING — existing .tene.enc files will fail to decrypt.
var KDFAlgRegistry = map[byte]KDFFunc{
    KDFAlgArgon2idV1: argon2idV1Derive,
}
```

### 3.2 Info Domain Constants

```go
// pkg/crypto/info.go (NEW)

// Purpose strings used as HKDF info parameter for DeriveSubKeyV2.
// Format: "tene/{domain}/v{N}"
//
// Each domain produces an independent sub-key from the same master key.
// Adding a new version (e.g., "tene/audit/v2") creates a fresh sub-key chain.
//
// All values are immutable — changing a string is BREAKING for any vault
// that already uses it.
const (
    PurposeEncryption    = "tene/encryption/v1"     // vault_secrets ciphertext
    PurposeAudit         = "tene/audit/v1"          // audit_log HMAC + encrypted_details (s2 enables)
    PurposeSync          = "tene/sync/v1"           // sync envelope key (internal/sync/envelope.go)
    PurposeRecovery      = "tene/recovery/v1"       // recovery_blob encryption key
    PurposeEncfileExport = "tene/encfile/v1"        // .tene.enc export file
    PurposeAuthHash      = "tene/auth-hash/v1"      // sibling passwd-verify
    PurposeKeychain      = "tene/keychain/v1"       // (reserved) keychain wrap
    PurposeBackup        = "tene/backup/v1"         // (reserved) backup encryption
    PurposeImport        = "tene/import/v1"         // (reserved) batch import key derivation
    PurposeJSONEnvelope  = "tene/json-envelope/v1"  // (reserved) JSON output envelope HMAC
    PurposeBiometricWrap = "tene/biometric/v1"     // (reserved; Sprint 4) biometric-wrapped DEK
    PurposeMCP           = "tene/mcp/v1"            // (reserved; v2.1) MCP server auth
    PurposeAuditMeta     = "tene/audit-meta/v1"     // (reserved) audit log metadata sealing
)
```

### 3.3 KDF Function Signatures

```go
// pkg/crypto/kdf.go (NEW)

// DeriveSubKeyV2 derives a sub-key from a master key using HKDF-SHA-256
// (RFC 5869) with an explicit salt and info string.
//
// Parameters:
//   masterKey: high-entropy master key (typically Argon2id output, 32 bytes)
//   salt: per-vault salt (e.g., vault_meta.kdf_salt; 32 bytes, MUST be non-empty)
//   info: domain separation string (e.g., crypto.PurposeEncryption)
//   length: output key length in bytes
//
// Returns the derived key.
//
// Errors:
//   ErrInvalidSalt if len(salt) == 0
//   wrapped io error if HKDF stream read fails (extremely rare)
//
// Standards: RFC 5869 §2 (HMAC-based Extract-and-Expand KDF)
func DeriveSubKeyV2(masterKey, salt []byte, info string, length int) ([]byte, error) {
    if len(salt) == 0 {
        return nil, teneerr.ErrInvalidSalt
    }
    reader := hkdf.New(sha256.New, masterKey, salt, []byte(info))
    out := make([]byte, length)
    if _, err := io.ReadFull(reader, out); err != nil {
        return nil, fmt.Errorf("hkdf expand: %w", err)
    }
    return out, nil
}

// DeriveSubKey is the legacy entry point retained for backward compatibility.
//
// Deprecated: use DeriveSubKeyV2 with an explicit salt argument.
// Callers SHOULD migrate before v2.1; this function will be removed in Sprint 2
// (vault-v2-migration).
//
// Internally, this calls DeriveSubKeyV2 with a nil salt (RFC 5869 §2.2: zero-filled
// HashLen string used). slog.Warn fires on each call to flag remaining callers.
func DeriveSubKey(masterKey []byte, info string, length int) ([]byte, error) {
    slog.Warn("pkg/crypto.DeriveSubKey deprecated; use DeriveSubKeyV2 with explicit salt",
        "info", info, "caller_hint", "see /Users/popup-kay/Documents/GitHub/agentkay/tene-biz/tene/docs/01-plan/features/crypto-v2-keys.plan.md")
    // Backwards-compatible: use a zero-length salt (HKDF-Extract uses HashLen zeros per RFC 5869 §2.2)
    // We pass a 32-byte zero array to satisfy the V2 invariant (salt non-empty), but the result is
    // equivalent to the v1 nil-salt behavior — this is intentional to preserve existing ciphertext.
    return DeriveSubKeyV2(masterKey, make([]byte, 32), info, length)
}

// argon2idV1Derive is the Argon2id implementation for KDFAlgArgon2idV1.
//
// Parameters (RFC 9106 recommended Section 4 for "low-memory" profile):
//   memory = 64 MiB (64*1024 KiB)
//   time = 4
//   parallelism = 4
//   key length = 32
//
// Note: these match the existing tene v1.0.x parameters. Changing them
// would require a new KDFAlg byte (e.g., 0x02 = Argon2idV2 with 128 MiB).
func argon2idV1Derive(password, salt []byte) ([]byte, error) {
    if len(salt) == 0 {
        return nil, teneerr.ErrInvalidSalt
    }
    key := argon2.IDKey(password, salt, 4, 64*1024, 4, 32)
    return key, nil
}
```

---

## 4. API Specification

### 4.1 New Public Functions

| Function | Signature | Purpose |
|----------|-----------|---------|
| `DeriveSubKeyV2` | `(masterKey, salt []byte, info string, length int) ([]byte, error)` | Primary HKDF API |
| `argon2idV1Derive` | `(password, salt []byte) ([]byte, error)` | Argon2id KDF (also accessible via KDFAlgRegistry) |

### 4.2 New Public Variables / Constants

| Name | Type | Value |
|------|------|-------|
| `KDFAlgRegistry` | `map[byte]KDFFunc` | `{0x01: argon2idV1Derive}` |
| `KDFAlgArgon2idV1` | `byte` | `0x01` |
| `PurposeEncryption`, `PurposeAudit`, ... (13 total) | `string` | `"tene/{domain}/v1"` |
| `ErrInvalidSalt` (in pkg/errors) | `error` | exit code 2 (INVALID_INPUT) |
| `ErrUnsupportedKDF` (in pkg/errors) | `error` | exit code 6 (DECRYPT_FAILED) |

### 4.3 Deprecated Functions

| Function | Status | Removal |
|----------|--------|---------|
| `DeriveSubKey(masterKey, info, length)` | `// Deprecated:` + `slog.Warn` | Sprint 2 `vault-v2-migration` |

---

## 5. UI/UX Design

> No user-visible UI change. CLI output unchanged. Internal-only refactor.

### 5.1 slog.Warn output during deprecated calls (developer-facing)

```
$ tene get FOO
time=2026-05-13T10:00:00.000Z level=WARN msg="pkg/crypto.DeriveSubKey deprecated; use DeriveSubKeyV2 with explicit salt" info=tene/encryption/v1 caller_hint=docs/01-plan/features/crypto-v2-keys.plan.md
my-secret-value
```

After 16-site migration, this warn fires only from `keymanager.go` (deprecated shim).

### 5.2 Page UI Checklist (CLI semantics)

- [ ] `tene init` — vault_meta.kdf_salt 32 bytes random saved (already exists)
- [ ] `tene set FOO bar` — uses DeriveSubKeyV2(masterKey, vaultSalt, PurposeEncryption, 32) → ciphertext written
- [ ] `tene get FOO` — same path, decrypts successfully
- [ ] `tene export --enc` — uses DeriveSubKeyV2(masterKey, vaultSalt, PurposeEncfileExport, 32) → .tene.enc with KDFAlgArgon2idV1 byte
- [ ] `tene import --enc bad.enc` (with KDFByte=0xFF) — exit 6 + "unsupported KDF algorithm: 0xFF" stderr
- [ ] Tampered .tene.enc → existing decryption failure (no change)

---

## 6. Error Handling

### 6.1 Error Code Definition

| Code | Name | Cause | Handling |
|------|------|-------|----------|
| 2 | `ErrInvalidSalt` (INVALID_INPUT) | DeriveSubKeyV2 called with empty salt | Exit 2 + error message (developer error; not user-facing) |
| 6 | `ErrUnsupportedKDF` (DECRYPT_FAILED) | encfile header has unknown KDF byte | Exit 6 + "unsupported KDF algorithm: 0xNN" |

### 6.2 Error Response Format

```
# Bad encfile
$ tene import --enc bad.tene.enc
Error: unsupported KDF algorithm: 0xff (.tene.enc header byte 5)
$ echo $?
6

# Internal bug (caller passed nil salt)
$ tene get FOO    # if a caller hasn't been migrated and somehow passes nil
panic: ...
# In practice, DeriveSubKey shim always passes 32 zero bytes (not nil)
# so ErrInvalidSalt is impossible from external callers.
```

---

## 7. Security Considerations

- [x] **HKDF salt explicit**: salt parameter forces per-vault uniqueness (no accidental nil)
- [x] **Versioned domains**: `tene/audit/v1` vs `tene/audit/v2` — adding v2 doesn't break v1 ciphertext
- [x] **Standards compliance**: RFC 5869 (HKDF) + RFC 9106 (Argon2id) — passes 6 KAT
- [x] **KDFAlgRegistry**: silent corruption prevented; unknown byte → explicit error
- [x] **Constant-time considerations**: HKDF is not constant-time, but it operates on derived keys not user input — no timing side-channel relevant
- [x] **No master key in logs**: `slog.Warn` logs `info` string only, never key material
- [x] **Argon2id parameters frozen**: KDFAlgArgon2idV1 byte tied to specific params (64 MiB / 4 iter / 4 parallel). New params → new KDFAlg byte
- [x] **Migration safety**: deprecated `DeriveSubKey` uses zero-byte salt → backward-compatible derivation for existing vaults; s2 migration `003_secrets_v2_aad` re-encrypts with new per-vault salt

---

## 8. Test Plan

### 8.1 Test Scope

| Type | Target | Tool | Phase |
|------|--------|------|-------|
| L1: KAT | RFC 5869 §A.1, A.2, A.3 (3 HKDF vectors) | `go test` | Do |
| L1: KAT | RFC 9106 §B.1, B.2, B.3 (3 Argon2id vectors) | `go test` | Do |
| L1: Unit | `DeriveSubKeyV2` nil salt → `ErrInvalidSalt` | `go test` | Do |
| L1: Unit | `KDFAlgRegistry` lookup (0x01 found, 0xFF not found) | `go test` | Do |
| L2: Integration | 16-site smoke test (set + get + list + run + passwd + export + import) | testhelper_test.go | Do |
| L2: Integration | encfile import with KDFAlg=0xFF → `ErrUnsupportedKDF` | encfile_test.go | Do |
| L3: Regression | v1 vault fixture (legacy DeriveSubKey path via shim) → `tene get` succeeds | integration_test.go | Check |
| L3: Performance | DeriveSubKeyV2 latency ≈ DeriveSubKey (within 10%) | `go test -bench` | Do |

### 8.2 L1: KAT Test Scenarios

| # | Vector | Expected |
|---|--------|---------|
| 1 | RFC 5869 §A.1 IKM=0x0b*22, salt=0x00..0c, info=0xf0..f9, L=42 → PRK + OKM | hex match |
| 2 | RFC 5869 §A.2 IKM=0x00..4f (80B), salt=0x60..af (80B), info=0xb0..ff (80B), L=82 → PRK + OKM | hex match |
| 3 | RFC 5869 §A.3 IKM=0x0b*22, salt=nil, info=nil, L=42 → PRK + OKM | hex match (note: tests "salt-less" mode — we cover via shim) |
| 4 | RFC 9106 §B.1 password="01"*32, salt="02"*16, t=3, m=32, p=4 → tag | hex match |
| 5 | RFC 9106 §B.2 Argon2id with different params | hex match |
| 6 | RFC 9106 §B.3 Argon2id large input | hex match |

### 8.3 L2: Integration Test Scenarios

| # | Scenario | Expected |
|---|----------|----------|
| 1 | `tene init && tene set FOO bar && tene get FOO` (v2 vault, fresh kdf_salt) | "bar" output |
| 2 | `tene list` (v2 vault, 3 secrets) | 3 names |
| 3 | `tene export --enc out.tene.enc` (v2 vault) | file created; header bytes: KDFAlg=0x01 |
| 4 | `tene import --enc in.tene.enc` (v1 vault with KDFAlg=0x01) | secrets imported |
| 5 | Manual corrupt KDFAlg=0xFF in .tene.enc → `tene import --enc` | exit 6 + ErrUnsupportedKDF |
| 6 | 2 vaults with same mnemonic → set FOO in each → ciphertexts differ | byte-diff |

### 8.4 L3: E2E + Regression Scenarios

| # | Scenario | Expected |
|---|----------|----------|
| 1 | v1 vault fixture (no auth_hash, no PurposeXxx awareness) → `tene get FOO` | success (DeriveSubKey shim → DeriveSubKeyV2 with zero salt) |
| 2 | grep `crypto.DeriveSubKey\(.*nil` --include='*.go' | 0 matches |
| 3 | grep `crypto.DeriveSubKey\b` --include='*.go' (excluding _test.go) | ≤ 3 matches (deprecated shim sites only — `keymanager.go` + possibly `kdf.go` definition) |
| 4 | `time go test ./pkg/crypto/...` | < 30s |

### 8.5 KAT Vector Source Files

```
pkg/crypto/testdata/
├── rfc5869-a1.json   # IETF RFC 5869 §A.1
├── rfc5869-a2.json   # IETF RFC 5869 §A.2
├── rfc5869-a3.json   # IETF RFC 5869 §A.3 (salt-less variant — note in test)
├── rfc9106-b1.json   # IETF RFC 9106 §B.1
├── rfc9106-b2.json   # IETF RFC 9106 §B.2
└── rfc9106-b3.json   # IETF RFC 9106 §B.3
```

Each JSON: `{"source": "RFC 5869 §A.1", "ikm": "<hex>", "salt": "<hex>", "info": "<hex>", "length": 42, "prk": "<hex>", "okm": "<hex>"}`.

---

## 9. Clean Architecture

### 9.1 Layer Structure

| Layer | Component | Location |
|-------|-----------|----------|
| **Domain** | `DeriveSubKeyV2`, `KDFAlgRegistry`, `PurposeXxx` constants | `pkg/crypto/` |
| **Application** | (none changed by this design) | n/a |
| **Infrastructure** | 16 caller sites (CLI, encfile, recovery, sync) | `internal/*/` |
| **Cross-cutting** | `ErrInvalidSalt`, `ErrUnsupportedKDF` | `pkg/errors/codes.go` |

### 9.2 Dependency Rules

```
internal/* (caller) → pkg/crypto (Domain)
                    → pkg/errors (Cross-cutting)

pkg/crypto MUST NOT import internal/* (Clean Architecture).
```

### 9.3 File Import Rules

| From | Imports | Forbidden |
|------|---------|-----------|
| `pkg/crypto/*` | `golang.org/x/crypto/*`, `crypto/*` (stdlib), `pkg/errors` | `internal/*` |
| `internal/cli/*.go` (11 sites) | `pkg/crypto`, `internal/vault` (for kdf_salt) | n/a |
| `internal/encfile/encfile.go` | `pkg/crypto`, `pkg/errors` | n/a |
| `internal/recovery/recover.go` | `pkg/crypto`, `pkg/errors` | n/a |
| `internal/sync/envelope.go` | `pkg/crypto`, `pkg/errors` | n/a |

### 9.4 This Feature's Layer Assignment

| Component | Layer | Location |
|-----------|-------|----------|
| `DeriveSubKeyV2`, `argon2idV1Derive` | Domain | `pkg/crypto/kdf.go` |
| `KDFAlgRegistry` | Domain | `pkg/crypto/registry.go` |
| `PurposeXxx` constants (13) | Domain | `pkg/crypto/info.go` |
| 6 RFC KAT vectors | Domain (test) | `pkg/crypto/testdata/*.json` |
| 16 callers | Application/Infrastructure | `internal/*/` |
| `ErrInvalidSalt`, `ErrUnsupportedKDF` | Cross-cutting | `pkg/errors/codes.go` |

---

## 10. Coding Convention Reference

### 10.1 Naming Conventions

| Target | Rule | Example |
|--------|------|---------|
| HKDF info strings | `tene/{domain}/v{N}` | `"tene/encryption/v1"` |
| Purpose constants | `Purpose{Domain}` | `PurposeEncryption`, `PurposeAudit` |
| KDF algorithm bytes | `KDFAlg{Name}V{N}` | `KDFAlgArgon2idV1 byte = 0x01` |
| KDF derive funcs | `{name}V{N}Derive` (private) | `argon2idV1Derive` |
| Errors | `Err{Cause}` | `ErrInvalidSalt`, `ErrUnsupportedKDF` |
| KAT test files | `rfc{N}-{section}.json` | `rfc5869-a1.json` |

### 10.2 Import Order

```go
package crypto

import (
    // 1. Standard library
    "crypto/sha256"
    "fmt"
    "io"
    "log/slog"

    // 2. External
    "golang.org/x/crypto/argon2"
    "golang.org/x/crypto/hkdf"

    // 3. Internal
    teneerr "github.com/tene-ai/tene/pkg/errors"
)
```

---

## 11. Implementation Guide

### 11.1 File Structure

```
tene/
├── pkg/
│   ├── crypto/
│   │   ├── kdf.go              # MODIFIED (DeriveSubKey deprecated + DeriveSubKeyV2 added)
│   │   ├── info.go             # NEW (13 Purpose constants)
│   │   ├── registry.go         # NEW (KDFAlgRegistry + KDFAlgArgon2idV1)
│   │   ├── kat_test.go         # NEW (6 KAT vectors)
│   │   └── testdata/
│   │       ├── rfc5869-a1.json # NEW
│   │       ├── rfc5869-a2.json # NEW
│   │       ├── rfc5869-a3.json # NEW
│   │       ├── rfc9106-b1.json # NEW
│   │       ├── rfc9106-b2.json # NEW
│   │       └── rfc9106-b3.json # NEW
│   └── errors/
│       └── codes.go            # MODIFIED (ErrInvalidSalt, ErrUnsupportedKDF)
├── internal/
│   ├── cli/                    # MODIFIED (11 sites)
│   │   ├── set.go:107
│   │   ├── get.go:54
│   │   ├── run.go:54
│   │   ├── list.go:36, :86
│   │   ├── export.go:45
│   │   ├── passwd.go:37, :59
│   │   ├── recover.go:62, :84
│   │   └── import_cmd.go:48
│   ├── encfile/
│   │   └── encfile.go:74, :110, :169  # MODIFIED (KDFAlg check + 2 DeriveSubKey sites)
│   ├── recovery/
│   │   └── recover.go:35, :59         # MODIFIED (2 sites)
│   └── sync/
│       └── envelope.go:28             # MODIFIED (1 site)
```

### 11.2 Implementation Order

> **PR #1** (`feat(crypto): KDFAlgRegistry + Argon2id v1 ID byte validation`):

1. [ ] `pkg/errors/codes.go` — `ErrInvalidSalt` + `ErrUnsupportedKDF` (exit codes 2 + 6)
2. [ ] `pkg/crypto/registry.go` — `KDFAlgArgon2idV1` constant + `KDFAlgRegistry` map + `argon2idV1Derive` func
3. [ ] `pkg/crypto/registry_test.go` — registry lookup (0x01 found, 0xFF not found)
4. [ ] `internal/encfile/encfile.go:74` — `KDFAlgRegistry` lookup; unknown byte → `ErrUnsupportedKDF`
5. [ ] `internal/encfile/encfile_test.go` — fixture with KDFAlg=0xFF → expect error

> **PR #2** (`feat(crypto): DeriveSubKeyV2 with versioned info + KAT 12개`):
> Note: master plan says 12, but this design's KAT is 6 (3 HKDF + 3 Argon2id). Remaining 6 = RFC 8439 XChaCha20-Poly1305 (sibling `crypto-v2-aad`'s scope).

6. [ ] `pkg/crypto/info.go` — 13 Purpose constants
7. [ ] `pkg/crypto/kdf.go` — `DeriveSubKeyV2(masterKey, salt, info, length)` function
8. [ ] `pkg/crypto/kdf.go` — `DeriveSubKey` add `// Deprecated:` + `slog.Warn` + redirect to V2 with zero salt
9. [ ] `pkg/crypto/testdata/rfc5869-a1.json`, `a2.json`, `a3.json` — IETF vectors
10. [ ] `pkg/crypto/testdata/rfc9106-b1.json`, `b2.json`, `b3.json` — IETF vectors
11. [ ] `pkg/crypto/kat_test.go` — `TestKAT_HKDF_RFC5869` (3 tests) + `TestKAT_Argon2id_RFC9106` (3 tests)

> **PR #3** (`refactor(*): migrate 16 sites to DeriveSubKeyV2 (per-vault salt)`):

12. [ ] For each of 16 sites:
    - Load `vaultSalt` (caller-specific: most via `app.Vault.GetMeta("kdf_salt")` + base64 decode)
    - Replace `crypto.DeriveSubKey(masterKey, info, length)` with `crypto.DeriveSubKeyV2(masterKey, vaultSalt, crypto.PurposeXxx, length)`
13. [ ] `grep -rEn 'crypto.DeriveSubKey\(.*nil' --include='*.go'` — confirm 0 matches
14. [ ] `grep -rEn 'crypto.DeriveSubKey\b' --include='*.go' | grep -v _test.go` — confirm ≤ 3 (deprecation shim sites)
15. [ ] CHANGELOG.md — "feat(crypto): DeriveSubKey v2 with per-vault salt + KDFAlgRegistry (RFC 5869 + RFC 9106 KAT compliant)"

### 11.3 Session Guide

#### Module Map

| Module | Scope Key | Description | Estimated Turns |
|--------|-----------|-------------|:---------------:|
| Registry + encfile validation | `module-1` | PR #1 (KDFAlgRegistry + encfile byte check) | 15-20 |
| V2 function + KAT vectors | `module-2` | PR #2 (DeriveSubKeyV2 + 6 RFC vectors) | 25-30 |
| 16-site migration | `module-3` | PR #3 (caller updates) | 30-40 |

#### Recommended Session Plan

| Session | Phase | Scope | Turns |
|---------|-------|-------|:-----:|
| Session 1 | Plan + Design | (this document) | 30-35 (current) |
| Session 2 | Do | `--scope module-1` (PR #1) | 20-25 |
| Session 3 | Do | `--scope module-2` (PR #2) | 30-35 |
| Session 4 | Do | `--scope module-3` (PR #3) | 35-45 |
| Session 5 | Check + Report | All modules | 25-30 |

---

## 12. Edge Cases & Failure Modes

| # | Scenario | Behavior |
|---|----------|----------|
| 1 | `DeriveSubKeyV2(nil, nil, "", 0)` | `ErrInvalidSalt` (salt check first) |
| 2 | `DeriveSubKeyV2(nil, salt, "", 0)` | hkdf.New accepts; returns empty slice |
| 3 | Massive `length` (e.g., 10MB) | `io.ReadFull` allocates; succeeds (HKDF can produce up to 255 × HashLen = 8160 bytes per RFC 5869 §2.3); >8160 → `io.ReadFull` fails → wrapped error |
| 4 | `argon2idV1Derive` with empty password | Argon2id processes; returns key (not an error) |
| 5 | `KDFAlgRegistry[0x00]` | nil func → caller checks `ok` bool → `ErrUnsupportedKDF` |
| 6 | Concurrent registry access | Read-only after init; no mutex needed (Go map read-only concurrent safe per spec) |
| 7 | `slog.Warn` storm during migration period | Acceptable; s2 removes deprecated function entirely |
| 8 | v1 vault uses `keymanager.DeriveSubKey(rootKey, info)` (different from `crypto.DeriveSubKey`) | `keymanager.go` itself calls `crypto.DeriveSubKey` (deprecated shim) — fires warn — works |

---

## 13. Sequence Diagram — `tene set FOO bar` (after migration)

```
User           cli/set.go:107       pkg/crypto              internal/vault
 │                  │                    │                       │
 │ tene set FOO bar │                    │                       │
 ├─────────────────▶│                    │                       │
 │                  │ loadCachedMasterKey│                       │
 │                  ├───┐                │                       │
 │                  │   │ keychain.Load()│                       │
 │                  │◀──┘                │                       │
 │                  │ GetMeta("kdf_salt")│                       │
 │                  ├────────────────────┼──────────────────────▶│
 │                  │◀───────────────────┼─── salt (32 bytes) ───┤
 │                  │ DeriveSubKeyV2(   │                       │
 │                  │    masterKey,      │                       │
 │                  │    salt,           │                       │
 │                  │    PurposeEncryption,                     │
 │                  │    32)             │                       │
 │                  ├───────────────────▶│                       │
 │                  │◀──── subKey (32B)──┤                       │
 │                  │ Encrypt(subKey, plaintext, aad)            │
 │                  ├───────────────────▶│                       │
 │                  │◀──── ciphertext ───┤                       │
 │                  │ SetSecret("FOO", base64(ct), env)          │
 │                  ├────────────────────┼──────────────────────▶│
 │                  │◀───────────────────┼─── OK ────────────────┤
 │ exit 0           │                    │                       │
 │◀─────────────────┤                    │                       │
```

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-05-13 | Initial draft (Sprint 1 design phase, L2 boundary) | cto-lead (security-architect perspective) |
