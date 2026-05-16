---
template: plan
version: 1.3
description: tene CLI v2.0 Sprint 1 — crypto-v2-keys (P0-C1 + P0-C3; DeriveSubKeyV2 + per-vault salt + 16 사이트 migration + KDFAlgRegistry)
variables:
  - feature: crypto-v2-keys
  - displayName: "crypto-v2-keys (P0-C1 + P0-C3)"
  - sprint: tene-cli-v2-s1
  - date: 2026-05-13
  - author: cto-lead (PM perspective)
  - project: tene CLI
  - version: v2.0.0 (target)
  - masterPlan: docs/01-plan/features/tene-cli-v2-2026q3.master-plan.md
  - trustLevel: L2
---

# crypto-v2-keys Planning Document

> **Summary**: `DeriveSubKey(rootKey, info, length)` 가 현재 nil salt 로 호출되어 모든 vault 가 동일 sub-key 도메인 derivation 사용. 16 호출 사이트를 `DeriveSubKeyV2(masterKey, vaultSalt, "tene/{domain}/v1")` 으로 일괄 이전 + KDFAlgRegistry 도입.
>
> **Project**: tene CLI
> **Version**: target v2.0.0 (current v1.0.8)
> **Sprint**: `tene-cli-v2-s1` (W1-W2, 2026-05-13 → 2026-05-26)
> **Author**: cto-lead
> **Date**: 2026-05-13
> **Status**: Draft (L2)
> **Master Plan**: [tene-cli-v2-2026q3.master-plan.md](./tene-cli-v2-2026q3.master-plan.md) §2.2 P0-C1, P0-C3, §3.1 보정 1
> **Baseline 분석**: [cli-sprint-master-plan-2026-05-13.md](../cli-sprint-master-plan-2026-05-13.md) §2.1 (보정 1) + §6.1.3 Trk-A T1-A3/A4/A6
> **감사 출처**: [cli-completeness-audit-2026-05-11.md](../../03-report/cli-completeness-audit-2026-05-11.md) A1 P2 (RecoverySalt) + DR-3

---

## Executive Summary

| Perspective | Content |
|-------------|---------|
| **Problem** | `pkg/crypto/keymanager.go:DeriveSubKey(rootKey, "tene/audit")` 같은 호출이 16 사이트에서 발견 — **모두 salt=nil 전달**. RFC 5869 (HKDF) 는 salt 옵션이지만, 미설정 시 동일 masterKey 가 입력될 때 동일 sub-key 가 derive 됨. 이는 **per-vault 고유 sub-key 가 없다**는 의미. 추가로 `KDFAlgorithm` byte 검증 부재 (`encfile.go:74`) → 알 수 없는 KDF byte 가 silent 통과 → 향후 KDF 알고리즘 추가 시 미지원 byte 의 silent corruption. |
| **Solution** | (1) `DeriveSubKeyV2(masterKey, salt []byte, info string, length int)` 신설 — salt 필수 인자화. salt = `base64decode(vault_meta.kdf_salt)` (이미 init 시 32-byte 랜덤 저장). (2) info 문자열 13개 도메인 모두 버전화: `"tene/audit/v1"`, `"tene/sync/v1"`, `"tene/recovery/v1"`, `"tene/encfile/v1"`, `"tene/auth-hash/v1"` 등. (3) 16 호출 사이트 일괄 이전. (4) 기존 `DeriveSubKey` 는 deprecated marker + `slog.Warn` (s2 에서 제거). (5) `KDFAlgRegistry` 도입 — id 0x01 = Argon2idV1 (64MB), 미지원 byte → `teneerr.ErrUnsupportedKDF`. |
| **Function/UX Effect** | 사용자 가시 변화 없음 (CLI UX 0 회귀). 내부적으로 모든 vault 가 고유 sub-key 도메인 derive → 동일 mnemonic 사용자가 여러 vault 만들어도 sub-key 분리. encfile (.tene.enc export) 의 향후 KDF 변경 시 explicit error. |
| **Core Value** | 보안 narrative 의 기술적 토대. Show HN 메시지 "36 RFC test vectors" 의 prerequisite (RFC 5869 KAT 3건 통과 가능해짐). |

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | `DeriveSubKey` 16 호출 사이트가 모두 salt=nil → 모든 vault 가 동일 sub-key 도메인 분리. `KDFAlgorithm` byte 검증 부재 → 미지원 KDF silent 통과. RFC 5869 KAT 통과 불가 (salt 없이는 표준 벡터 일치 안 함). |
| **WHO** | **Sec-conscious team (15%)** — KAT 통과 + per-vault 보안 격리 가치. **Indie OSS dev (25%)** — 36 RFC vector pass 가 GitHub README badge 가치. |
| **RISK** | (a) 16 사이트 일괄 변경 시 누락 → 일부 keyset nil salt 잔존 (mitigation: grep + integration test). (b) `vault_meta.kdf_salt` 가 없는 환경 (테스트 fixture 등) → fallback 필요. (c) DeriveSubKey 시그니처 변경 BREAKING — internal 함수이므로 pkg/crypto 외부 영향 0 (사용자 영향 없음). (d) `encfile` 형식 byte 추가 → 기존 export 파일 호환성 유지 필수. |
| **SUCCESS** | grep `crypto.DeriveSubKey\(.*nil` = 0 (전체 codebase); grep `crypto.DeriveSubKey\b` ≤ 3 (deprecated wrapper 만); `pkg/crypto/testdata/rfc5869-*.json` 3 KAT 통과; `encfile` 미지원 KDF byte → `ErrUnsupportedKDF` (테스트로 검증); 16 도메인 모두 `tene/{domain}/v1` 패턴. |
| **SCOPE** | Sprint 1 Trk-A 핵심 (4 PR: #1 KDFAlgRegistry, #2 DeriveSubKeyV2 + KAT, #3 16-site migration, #6 encfile validation). +540 LOC (impl +180 + KAT test +340 + 16 site migration -40/+60). 변경 범위: pkg/crypto + internal/cli (11) + internal/encfile (2) + internal/recovery (2) + internal/sync (1). |

---

## 1. Overview

### 1.1 Purpose

암호화 핵심 primitive 의 표준 부합. RFC 5869 (HKDF) 와 RFC 9106 (Argon2id) 의 KAT (Known Answer Test) 벡터 통과를 위한 **prerequisite**. 또한 per-vault 보안 격리 (동일 master key 로 여러 vault 만들 때 sub-key 분리) 확보.

### 1.2 Background

#### 1.2.1 발견 경로 (보정 1)

1차 plan (2026-05-12) 은 "vault.go 16곳" 으로 표기. 실제 grep:

```
internal/cli/set.go:107         internal/cli/passwd.go:37
internal/cli/get.go:54          internal/cli/passwd.go:59
internal/cli/run.go:54          internal/cli/recover.go:62
internal/cli/list.go:36         internal/cli/recover.go:84
internal/cli/list.go:86         internal/cli/import_cmd.go:48
internal/cli/export.go:45       internal/encfile/encfile.go:110
internal/encfile/encfile.go:169 internal/recovery/recover.go:35
internal/recovery/recover.go:59 internal/sync/envelope.go:28
```

= 16 사이트, **vault.go 에는 0 곳**. 정확한 분포: cli 11 + encfile 2 + recovery 2 + sync 1. 보정 1 의 의미: 단순 grep-replace 아닌 **A2 (Clean Architecture) P0 의 증거** — CLI 가 crypto 를 직접 호출하는 11 사이트 → 향후 `SecretCipher` interface 추출 필요 (s2 의 `sync-contracts` 와 통합).

#### 1.2.2 nil salt 의 의미 (RFC 5869)

RFC 5869 §2.2: `HKDF-Extract(salt, IKM) = HMAC-Hash(salt, IKM)`. salt 가 nil 또는 zero-length 면 RFC 는 "use a string of HashLen zeros" 명시. 즉 모든 vault 의 동일 masterKey + 동일 info 는 **항상 동일 sub-key** 출력. per-vault 격리 0.

`vault_meta.kdf_salt` 는 이미 init 시 32-byte 랜덤 저장 (`internal/vault/vault.go:?`). 그러나 sub-key derivation 에는 사용 안 됨 → 이 plan 의 핵심 fix.

#### 1.2.3 KDFAlgRegistry (P0-C3)

`encfile.go:74` 의 KDF byte 검증:

```go
// 현재 (v1.0.8)
header.KDFAlg = data[5]  // ← byte 추출만, 검증 없음
// 미지원 byte (예: 0xFF) 는 silent 통과 → Argon2id hardcoded 사용 → mismatch
```

수정: registry pattern 으로 명시적 dispatch.

```go
// v2 (목표)
var KDFAlgRegistry = map[byte]func(password, salt []byte) ([]byte, error){
    0x01: argon2idV1Derive,  // 64MB, 4 iter, 4 parallel
}
fn, ok := KDFAlgRegistry[header.KDFAlg]
if !ok {
    return teneerr.ErrUnsupportedKDF
}
key, err := fn(password, salt)
```

### 1.3 Related Documents

- Master Plan: [`tene-cli-v2-2026q3.master-plan.md`](./tene-cli-v2-2026q3.master-plan.md)
- Baseline: [`cli-sprint-master-plan-2026-05-13.md §2.1 §6.1.3`](../cli-sprint-master-plan-2026-05-13.md)
- 감사: [`cli-completeness-audit-2026-05-11.md DR-3`](../../03-report/cli-completeness-audit-2026-05-11.md)
- Sibling: [`passwd-verify.plan.md`](./passwd-verify.plan.md) (PurposeAuthHash 도메인 공유), [`crypto-v2-aad.plan.md`](./crypto-v2-aad.plan.md) (이 plan 의 follow-up; AAD enrichment 는 별도 PR)

---

## 2. Scope

### 2.1 In Scope

- [ ] `pkg/crypto/kdf.go` — `DeriveSubKeyV2(masterKey, salt []byte, info string, length int) ([]byte, error)` 신설; salt 빈 슬라이스 거부 (`teneerr.ErrInvalidSalt`)
- [ ] `pkg/crypto/info.go` (신규) — 13 도메인 constant: `PurposeEncryption`, `PurposeAudit`, `PurposeSync`, `PurposeRecovery`, `PurposeEncfileExport`, `PurposeAuthHash` 등 (모두 `"tene/{domain}/v1"`)
- [ ] `pkg/crypto/kdf.go:DeriveSubKey` (기존) — `// Deprecated: use DeriveSubKeyV2` marker + `slog.Warn` 호출 (s2 에서 제거)
- [ ] `pkg/crypto/registry.go` (신규) — `KDFAlgRegistry map[byte]KDFFunc` + `KDFAlgArgon2idV1 byte = 0x01` constant
- [ ] `pkg/errors/codes.go` — `ErrUnsupportedKDF` 신규 (exit code 는 `audit-reader` plan 표 따름; 신규 6 = DECRYPT_FAILED 같은 카테고리)
- [ ] 16 호출 사이트 일괄 이전:
  - `internal/cli/`: set.go:107, get.go:54, run.go:54, list.go:36, list.go:86, export.go:45, passwd.go:37, passwd.go:59, recover.go:62, recover.go:84, import_cmd.go:48 (11)
  - `internal/encfile/encfile.go`: 110, 169 (2)
  - `internal/recovery/recover.go`: 35, 59 (2)
  - `internal/sync/envelope.go`: 28 (1)
- [ ] `internal/encfile/encfile.go:74` — `KDFAlgRegistry` lookup 분기 + `ErrUnsupportedKDF`
- [ ] `pkg/crypto/testdata/rfc5869-*.json` (3 KAT — RFC 5869 §A.1, A.2, A.3 표준 벡터)
- [ ] `pkg/crypto/testdata/rfc9106-*.json` (3 KAT — RFC 9106 §B.1 Argon2id)
- [ ] `pkg/crypto/kat_test.go` (신규) — 6 벡터 (RFC 5869 + 9106) load + assertion
- [ ] CHANGELOG.md — "feat(crypto): DeriveSubKey v2 with per-vault salt + KDFAlgRegistry"

### 2.2 Out of Scope

- AAD enrichment (`EncryptV2`/`DecryptV2`) — sibling [`crypto-v2-aad.plan.md`](./crypto-v2-aad.plan.md) 가 별도 plan/PR (#4)
- BIP39 KAT 24 벡터 — Sprint 2 `test-infra` 의 T2-B4
- RFC 8439 (XChaCha20-Poly1305) KAT 6 벡터 — sibling `crypto-v2-aad.plan.md` 가 처리
- DeriveSubKey 의 완전 제거 (`// Deprecated` 만 추가; 제거는 s2 의 `vault-v2-migration` 후)
- Zero-out 보강 (`runtime.KeepAlive`) — Sprint 1 T1-A7 (별도 PR; 본 plan 범위 밖)
- per-vault recovery_salt 도입 — 보정 §2.3 으로 P2 강등; Sprint 후반 또는 v2.1
- `SecretCipher` interface 추출 (A2 P0 의 본격 해소) — Sprint 2 `sync-contracts`

---

## 3. Requirements

### 3.1 Functional Requirements

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-01 | `DeriveSubKeyV2(masterKey, salt, info, length)` 가 RFC 5869 HKDF-Extract → HKDF-Expand 표준 따름 (salt 필수) | **P0** | Pending |
| FR-02 | salt 가 nil 또는 0 length → `teneerr.ErrInvalidSalt` (불가) | **P0** | Pending |
| FR-03 | 13 도메인 constant 모두 `"tene/{domain}/v1"` 패턴 | **P0** | Pending |
| FR-04 | 16 호출 사이트 모두 `DeriveSubKeyV2(masterKey, vaultSalt, PurposeX, length)` 로 이전 | **P0** | Pending |
| FR-05 | `vaultSalt = base64decode(vault_meta.kdf_salt)` — caller 가 vault 로부터 로드 | **P0** | Pending |
| FR-06 | 기존 `DeriveSubKey(rootKey, info, length)` 는 `// Deprecated` marker + `slog.Warn("crypto.DeriveSubKey deprecated...")` | **P0** | Pending |
| FR-07 | `KDFAlgRegistry` 도입 — `KDFAlgArgon2idV1 = 0x01` reserved; 0x02+ 향후 | **P0** | Pending |
| FR-08 | `encfile.go:74` 미지원 KDF byte → `ErrUnsupportedKDF` (exit 1 + 명시 메시지) | **P0** | Pending |
| FR-09 | RFC 5869 §A.1, A.2, A.3 KAT 3 통과 (`go test -run KAT_HKDF`) | **P0** | Pending |
| FR-10 | RFC 9106 §B.1 Argon2id KAT 3 통과 (`go test -run KAT_Argon2id`) | **P0** | Pending |
| FR-11 | grep `crypto.DeriveSubKey\(.*nil` = 0 (production code) | **P0** | Pending |
| FR-12 | grep `crypto.DeriveSubKey\b` ≤ 3 (deprecated wrapper 사이트만) | **P0** | Pending |

### 3.2 Non-Functional Requirements

| Category | Criteria | Measurement Method |
|----------|----------|-------------------|
| Compatibility | 기존 v1 vault.db 의 ciphertext 는 그대로 복호화 가능 (sub-key 도메인 변경은 신규 vault 만) | integration test (v1 vault fixture + tene get) |
| Standards Compliance | RFC 5869 (HKDF) + RFC 9106 (Argon2id) KAT 6 통과 | crypto_test.go (M9 게이트) |
| Performance | DeriveSubKeyV2 latency ≈ DeriveSubKey (HKDF 1단계 차이) | `go test -bench` |
| Migration Safety | 16 사이트 일괄 변경 후 모든 명령 동작 (passwd / get / set / list / run / export / import / push / pull / recover / audit) | integration test |
| Audit Trail | `crypto.DeriveSubKey` 호출 잔존 grep 결과 PR description 첨부 | manual verification |

---

## 4. Success Criteria

### 4.1 Definition of Done

- [ ] FR-01 ~ FR-12 모두 구현 (grep + KAT + integration test 통과)
- [ ] `pkg/crypto/kat_test.go` — RFC 5869 §A.1, A.2, A.3 + RFC 9106 §B.1 6 벡터 hex 일치
- [ ] `pkg/crypto/info.go` — 13 도메인 constant 정의 + godoc
- [ ] `pkg/crypto/registry.go` — KDFAlgRegistry + KDFAlgArgon2idV1 byte
- [ ] `internal/encfile/encfile_test.go` — 미지원 KDF byte (0xFF) → ErrUnsupportedKDF
- [ ] CHANGELOG.md 업데이트
- [ ] `go test -race -count=10 ./pkg/crypto/...` 0 flaky
- [ ] gap-detector Match Rate ≥ 90% (M8)

### 4.2 Quality Criteria

- [ ] pkg/crypto coverage ≥ 80% (Trk-A 핵심)
- [ ] 16 호출 사이트 grep 결과 PR description 첨부 (검증 evidence)
- [ ] KAT 벡터는 RFC 표준 문서 직접 인용 (testdata/*.json 의 source 주석)
- [ ] linter clean — gosec G401 (weak cryptographic primitive) 0 issue

---

## 5. Risks and Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| 16 사이트 일괄 변경 시 누락 → nil salt 잔존 | H | M | T1-A4 후 grep 명시: `grep -rEn "DeriveSubKey\\(.*nil" --include='*.go'`; integration test 가 모든 호출 path 한 번씩 실행 |
| AAD enrichment (sibling crypto-v2-aad) 와 동일 PR 에서 변경 시 충돌 | M | M | **PR 분리**: #2 = DeriveSubKeyV2 + KAT, #3 = 16-site migration, #4 = EncryptV2/DecryptV2 (sibling). 시간 분리 명확 |
| 기존 v1 vault.db 호환 깨짐 (sub-key 다르게 derive) | H | H | **신규 vault 만** v2 sub-key 도메인 사용; 기존 vault 는 deprecated DeriveSubKey 그대로 (s2 의 003 migration 에서 정식 re-encrypt) |
| KDFAlgRegistry 추가로 `encfile` 기존 export 호환 깨짐 | H | L | byte 0x01 = Argon2idV1 (현재 사용 중) 이미 reserved; 기존 .tene.enc 파일은 그대로 동작 |
| RFC 5869 KAT 벡터 출처 모호 (RFC 본문 자체 KAT 부재) | M | M | golang.org/x/crypto 표준 라이브러리 자체 test vectors 활용 + IETF draft KAT 참조 |
| slog.Warn 폭주 (deprecated wrapper 호출 시) | L | L | s2 종료 시점에 `DeriveSubKey` 완전 제거; warn 폭주는 일시적 |

---

## 6. Impact Analysis

### 6.1 Changed Resources

| Resource | Type | Change Description |
|----------|------|--------------------|
| `pkg/crypto/kdf.go:DeriveSubKey` | Function | Deprecated marker + slog.Warn 추가 (지금은 유지) |
| `pkg/crypto/kdf.go:DeriveSubKeyV2` | Function (신규) | 시그니처 `(masterKey, salt, info, length) → (key, error)`; salt 빈 거부 |
| `pkg/crypto/info.go` | File (신규) | 13 도메인 constant |
| `pkg/crypto/registry.go` | File (신규) | KDFAlgRegistry + KDFAlgArgon2idV1 |
| `pkg/crypto/testdata/rfc5869-*.json` | Files (3 신규) | KAT 벡터 |
| `pkg/crypto/testdata/rfc9106-*.json` | Files (3 신규) | KAT 벡터 |
| `pkg/crypto/kat_test.go` | File (신규) | KAT 6 벡터 로드 + assertion |
| `internal/encfile/encfile.go:74` | Function | KDFAlg 검증 분기 |
| 16 호출 사이트 (위 리스트) | Function call | DeriveSubKey → DeriveSubKeyV2 (vaultSalt + Purpose constant) |
| `pkg/errors/codes.go` | Constant | `ErrUnsupportedKDF` 신규 |

### 6.2 Current Consumers

| Resource | Operation | Code Path | Impact |
|----------|-----------|-----------|--------|
| `crypto.DeriveSubKey` | sub-key derive (16 사이트) | 위 16 grep 결과 | **Migration** — 모두 DeriveSubKeyV2 로 이전 |
| `encfile.KDFAlg` byte | encfile header read | encfile.go:74 + decoder | **Validation 추가** — silent pass → explicit error |
| `pkg/crypto/encrypt.go:Encrypt` | AAD 사용 | 모든 caller | None (이 plan 은 sub-key derivation 만; AAD 는 sibling crypto-v2-aad) |
| `vault_meta.kdf_salt` (KV) | salt 저장 | `internal/vault/vault.go:?` | None (이미 32-byte 저장; 이 plan 은 사용처 추가) |

### 6.3 Verification

- [ ] grep `crypto.DeriveSubKey\(.*nil` = 0 — PR description 에 grep 결과 첨부
- [ ] grep `crypto.DeriveSubKey\b` ≤ 3 — deprecated wrapper 사이트만 (예: keymanager.go 의 backward-compat shim)
- [ ] grep `tene/.*?/v1` ≥ 13 — 모든 도메인 constant 사용
- [ ] integration test (smoke): `tene init && tene set FOO bar && tene get FOO` (v2 vault)
- [ ] integration test (compatibility): v1 fixture vault → `tene get FOO` (현재 패턴 그대로)

---

## 7. Architecture Considerations

### 7.1 Project Level Selection

| Level | Selected |
|-------|:--------:|
| Enterprise (tene CLI) | **☑** |

### 7.2 Key Architectural Decisions

| Decision | Options | Selected | Rationale |
|----------|---------|----------|-----------|
| `DeriveSubKey` 의 backward-compat 처리 | (a) 즉시 제거, (b) Deprecated marker + slog.Warn, (c) build tag 로 격리 | (b) Deprecated + Warn | s2 의 `vault-v2-migration` 까지 grace period; 즉시 제거 시 PR 매트릭스 폭주 |
| salt 파라미터 전달 방식 | (a) explicit `salt []byte` 인자, (b) `vault *Vault` 인자, (c) context.Value | (a) explicit salt | pkg/crypto 가 vault 의존 안 함 (Clean Architecture); caller 가 vaultSalt 로드 |
| Info 도메인 명명 | (a) `"tene-audit"`, (b) `"tene/audit"`, (c) `"tene/audit/v1"` | (c) versioned namespace | 향후 v2 도메인 추가 시 hash chain 분리 (e.g., `"tene/audit/v2"` 별도 sub-key) |
| KDF registry pattern | (a) switch/case 분기, (b) map registry, (c) interface | (b) map registry | 향후 KDF 추가 시 map 한 줄 등록; switch 보다 확장성 우수 |
| RFC 5869 KAT 출처 | (a) RFC 본문 (불충분), (b) golang.org/x/crypto test data, (c) NIST CAVS | (b) + (c) 혼합 | RFC §A.1-A.3 + golang.org/x/crypto/hkdf/hkdf_test.go 의 표준 벡터 |

### 7.3 Clean Architecture Approach

```
Layer mapping:
┌──────────────────────────────────────────────────────────┐
│ Presentation (internal/cli/)                             │
│   - 11 사이트 — vaultSalt 로드 + DeriveSubKeyV2 호출       │
├──────────────────────────────────────────────────────────┤
│ Application (internal/encfile/, internal/recovery/,      │
│              internal/sync/)                             │
│   - 5 사이트 — DeriveSubKeyV2 호출                        │
├──────────────────────────────────────────────────────────┤
│ Domain (pkg/crypto/)                                     │
│   - DeriveSubKeyV2 (RFC 5869 HKDF)                       │
│   - KDFAlgRegistry                                       │
│   - PurposeXxx constants                                 │
│   - kat_test.go (6 벡터)                                  │
├──────────────────────────────────────────────────────────┤
│ Infrastructure (internal/vault/)                         │
│   - vault_meta.kdf_salt (already exists)                 │
└──────────────────────────────────────────────────────────┘
```

---

## 8. Convention Prerequisites

### 8.1 Existing Project Conventions

- [x] Go module: `github.com/agent-kay-it/tene`
- [x] `pkg/crypto/` 패키지 (10 파일 + 5 test)
- [x] `golang.org/x/crypto/hkdf` 의존성 (이미 use)
- [x] `golang.org/x/crypto/argon2` 의존성 (이미 use)

### 8.2 Conventions to Define/Verify

| Category | Current State | To Define | Priority |
|----------|---------------|-----------|:--------:|
| HKDF info 도메인 명명 | 산발적 | `tene/{domain}/v{N}` 표준 + `PurposeXxx` constant | High |
| KDF byte ID | byte 1개 (검증 없음) | KDFAlgRegistry map[byte]Func | High |
| Crypto error 패키지 | `pkg/errors` (이미 존재) | `ErrInvalidSalt`, `ErrUnsupportedKDF` 신규 | Medium |
| KAT testdata 명명 | (없음) | `rfc{N}-{name}.json` 패턴 | Medium |

### 8.3 Environment Variables Needed

None (this plan).

---

## 9. Testing Plan

### 9.1 Test Scope

| Type | Target | Tool | Phase |
|------|--------|------|-------|
| L1: Unit | `DeriveSubKeyV2`, KDFAlgRegistry, info constants | `go test` | Do |
| L1: KAT | RFC 5869 §A.1-A.3 + RFC 9106 §B.1 (6 벡터) | crypto_test.go | Do |
| L2: Integration | 16 호출 사이트 — vault smoke test | testhelper_test.go | Do |
| L3: Regression | v1 vault fixture 호환성 | integration_test.go | Check |

### 9.2 Test Scenarios

| # | Scenario | Expected |
|---|----------|---------|
| T-1 | `DeriveSubKeyV2(masterKey, nil, info, 32)` | `ErrInvalidSalt` |
| T-2 | RFC 5869 §A.1 IKM/salt → PRK (hex 일치) | match |
| T-3 | RFC 5869 §A.1 PRK + info → OKM (hex 일치) | match |
| T-4 | RFC 5869 §A.2, A.3 동일 패턴 | match |
| T-5 | RFC 9106 §B.1 password/salt → tag (hex 일치) | match |
| T-6 | KDFAlgRegistry lookup 0x01 → argon2idV1Derive | non-nil func |
| T-7 | KDFAlgRegistry lookup 0xFF | nil func → caller `ErrUnsupportedKDF` |
| T-8 | encfile.go:74 미지원 byte → `ErrUnsupportedKDF` | exit 1 + 메시지 |
| T-9 | grep `crypto.DeriveSubKey\(.*nil` = 0 (production code) | match |
| T-10 | v1 vault fixture + `tene get FOO` | success (compatibility) |
| T-11 | 신규 vault + `tene set FOO bar && tene get FOO` | success (smoke) |
| T-12 | 동일 mnemonic 2개 vault — sub-key derive 결과 다름 (per-vault salt) | different |

---

## 10. Implementation Notes (Pre-Design Hints)

- HKDF-Expand: `golang.org/x/crypto/hkdf.Expand(hash, prk, info, length)` 사용
- HKDF-Extract: `golang.org/x/crypto/hkdf.Extract(hash, ikm, salt)` 사용
- 13 도메인 constant 예시:
  - `PurposeEncryption = "tene/encryption/v1"` — vault_secrets ciphertext
  - `PurposeAudit = "tene/audit/v1"` — audit_log HMAC (s2 의 encrypted_details 와 연결)
  - `PurposeSync = "tene/sync/v1"` — sync envelope
  - `PurposeRecovery = "tene/recovery/v1"` — recovery_blob
  - `PurposeEncfileExport = "tene/encfile/v1"` — .tene.enc export
  - `PurposeAuthHash = "tene/auth-hash/v1"` — passwd-verify auth_hash (sibling plan)
- `slog.Warn` 메시지 예시: `slog.Warn("crypto.DeriveSubKey deprecated; use DeriveSubKeyV2 with explicit salt", "caller", caller)`
- vaultSalt 로드 패턴: `saltB64, _ := app.Vault.GetMeta("kdf_salt"); salt, _ := base64.StdEncoding.DecodeString(saltB64)` (caller 책임)

---

## 11. Blocking Relationships

| Blocks | Reason |
|--------|--------|
| **`crypto-v2-aad`** (sibling plan; AAD enrichment) | EncryptV2 가 sub-key derivation 결과를 input — DeriveSubKeyV2 정착 후 EncryptV2 의미 |
| **`audit-reader`** (sibling plan; `audit_log.actor`) | 아니다 (audit-reader 는 SQL 만 추가; crypto 무관). 동시 진행 가능 |
| **Sprint 2 `test-infra`** (KAT 36 완성) | RFC 5869 + 9106 12 벡터가 본 plan 의 6 KAT 기반 — 본 plan 완료 후 BIP39 24 + RFC 8439 6 추가 가능 |
| **Sprint 2 `vault-v2-migration`** (`001_v2_envelope` + `002_audit_log_v2`) | migration 대상 ciphertext 가 DeriveSubKeyV2 결과 사용 |

| Blocked by | Reason |
|-----------|--------|
| (none) | unblocked from start |

---

## 12. Next Steps

1. [x] Plan 문서 작성 완료
2. [ ] Design 문서 작성 — [`docs/02-design/features/crypto-v2-keys.design.md`](../../02-design/features/crypto-v2-keys.design.md) — security-architect 위임
3. [ ] L2 boundary: design 후 사용자 승인 → do phase
4. [ ] Sprint 1 PR #1 (KDFAlgRegistry) → #2 (DeriveSubKeyV2 + KAT 6) → #3 (16-site migration)
5. [ ] Sibling `crypto-v2-aad` 의 EncryptV2/DecryptV2 가 본 plan 의 sub-key derivation 결과를 사용

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-05-13 | Initial draft (Sprint 1 plan phase, L2 boundary) | cto-lead |
