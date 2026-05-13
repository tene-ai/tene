---
template: design
version: 1.3
description: tene CLI v2.0 Sprint 1 — sync-cleanup (P0-S1 + P0-S2) Design Document
variables:
  - feature: sync-cleanup
  - sprint: tene-cli-v2-s1
  - date: 2026-05-13
  - author: cto-lead (code-analyzer perspective)
  - project: tene CLI
  - version: v2.0.0 (target)
  - planPath: docs/01-plan/features/sync-cleanup.plan.md
  - trustLevel: L2
---

# sync-cleanup Design Document

> **Summary**: `internal/sync/merge.go` (160 LOC) + `queue.go` (84 LOC) + 관련 test 완전 삭제 + `internal/sync/state.go` 신설 (5-field schema, single source of truth). `cli/push.go:163` 의 1-field saveSyncState 삭제 (engine.go state.go 사용). `engine.go:211-219` vault.db.base 생성 dead path 삭제.
>
> **Project**: tene CLI
> **Version**: target v2.0.0
> **Sprint**: tene-cli-v2-s1 (W1-W2)
> **Author**: cto-lead (code-analyzer perspective)
> **Date**: 2026-05-13
> **Status**: Draft (L2)
> **Planning Doc**: [sync-cleanup.plan.md](../../01-plan/features/sync-cleanup.plan.md)

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | 244 LOC dead code + saveSyncState 데이터 손실 결함 동시 존재. unused linter 비활성 강제 (s3 차단). 매 push 마다 4 필드 silent 손실. |
| **WHO** | **Solo founder (10%)** + **Sec-conscious team (15%)** — state 정확성 + 코드베이스 신뢰. |
| **RISK** | (a) merge.go 삭제가 v1.2 merge UX 복원 시 비용 ↑ (s4 design spike). (b) 기존 sync_state.json 호환 (omitempty). (c) backup 정책 변경 (vault.db.base 제거; timestamp 백업은 유지). |
| **SUCCESS** | grep `ThreeWayMerge\|SyncQueue\|EnqueueOperation` = 0; grep `func saveSyncState` = 1; integration test 5 field 모두 유지. |
| **SCOPE** | Sprint 1 Trk-B 단독 (2 PR). 1 dev × 3 dev-day. -309 LOC net. internal/sync + internal/cli/push.go. |

---

## 1. Overview

### 1.1 Design Goals

1. **Dead code complete removal**: merge.go + queue.go + 관련 test 완전 삭제 (rename/deprecate 없음)
2. **Single saveSyncState**: 1 정의, 1 schema (5 fields), 1 caller pattern
3. **Backward compatibility**: 기존 v1.0.x sync_state.json (1-field) 도 load 가능 (omitempty)
4. **Atomic state update**: existing record + partial update → merge (full-replace 아님)
5. **Document forward path**: merge UX 의 v1.2 부활은 새 architecture (CRDT/OT) — 본 plan 의 dead code 복원 아님
6. **Pre-build for s3 lint**: unused linter 활성화 가능 상태 (cloud commands 는 별도 build tag — s3 처리)

### 1.2 Design Principles

- **YAGNI**: ThreeWayMerge 가 한 번도 production 사용 안 됐다면 제거; 미래 부활은 새 design
- **Single source of truth**: SyncState struct + saveSyncState 함수 각 1곳
- **Forward-compatibility via omitempty**: 미래 필드 추가 시 backward read 깨지지 않게
- **Separation of concerns**: state.go 가 file I/O + JSON marshal 만; engine.go 는 sync logic 만
- **Explicit deletes**: git rm 으로 추적 가능 (file moves 안 함)

---

## 2. Architecture Options

### 2.0 Architecture Comparison

| Criteria | Option A: Inline Cleanup | Option B: state.go 분리 | Option C: pkg/sync 외부화 |
|----------|:-:|:-:|:-:|
| **Approach** | merge.go 삭제 + saveSyncState 통합을 engine.go 안에 fold | merge.go 삭제 + 신규 `state.go` 에 saveSyncState 분리 | `internal/sync` → `pkg/sync` 이동 (외부 import 가능) |
| **New Files** | 0 | 2 (`state.go`, `state_test.go`) | 다수 (전체 패키지 이동) |
| **Modified Files** | 2 (engine.go, cli/push.go) — delete 4 | 3 (engine.go, cli/push.go, conflict.go) — delete 4 | 다수 (모든 caller import path 변경) |
| **Complexity** | Low | Medium | High |
| **Maintainability** | Medium (engine.go 가 비대) | **High** (single responsibility) | High but overkill (scope 밖) |
| **Effort** | Low | Medium | High |
| **Test isolation** | Medium (engine_test.go 가 state 테스트도 포함) | **High** (`state_test.go` 분리) | High |
| **Risk** | Low | Low | High (import path 변경 BREAKING) |
| **Recommendation** | — | **Default choice** | Out of scope |

**Selected**: **Option B — state.go 분리** — **Rationale**: engine.go 가 이미 472 LOC. state 관리를 분리하면 SRP + 테스트 격리. `state.go` 는 ~80 LOC 소규모이므로 분리 overhead 최소. pkg/sync 이동 (Option C) 는 s2 의 `sync-contracts` (DR-2 blueprint) 와 통합 — 본 plan 의 범위 밖.

### 2.1 Component Diagram

```
BEFORE (현재 v1.0.8)
┌────────────────────────────────────────────────────────────────────┐
│                       internal/sync/                                │
├────────────────────────────────────────────────────────────────────┤
│  engine.go (472 LOC)                                                │
│    type syncStateFile struct { 5 fields }                          │
│    func saveSyncState(path, *syncStateFile) error  ← 5-field write │
│    func (e *Engine) Push / Pull                                    │
│    line 211-219: vault.db.base 생성 (ThreeWayMerge 대비)            │
│  merge.go (160 LOC)                                                │
│    func ThreeWayMerge(...) — production caller 0                   │
│  queue.go (84 LOC)                                                  │
│    type SyncQueue + EnqueueOperation — production caller 0         │
│  merge_test.go (139 LOC)                                            │
│  conflict_test.go (27 LOC)                                          │
│  envelope.go, types.go (other files)                                │
└────────────────────────────────────────────────────────────────────┘
            ↑
            │ caller
┌────────────────────────────────────────────────────────────────────┐
│ internal/cli/push.go:163                                            │
│    func saveSyncState(path, vaultID) error  ← 1-field write        │
│       data, _ := json.Marshal(map[string]string{"vault_id": ...})   │
│       os.WriteFile(path, data, 0600)         ← TRUNCATE-WRITE       │
│                                              ← engine 의 4 필드 nuke│
└────────────────────────────────────────────────────────────────────┘


AFTER (v2.0)
┌────────────────────────────────────────────────────────────────────┐
│                       internal/sync/                                │
├────────────────────────────────────────────────────────────────────┤
│  engine.go (~412 LOC) — merge.base path 삭제                        │
│    func (e *Engine) Push / Pull                                    │
│    (saveSyncState 호출은 state.go 의 함수 사용)                      │
│                                                                    │
│  state.go [NEW, ~80 LOC]                                            │
│    type SyncState struct {                                          │
│      VaultID, Version, Hash, LastPushedAt, LastPulledAt             │
│    } (omitempty JSON tags)                                          │
│    func LoadSyncState(path string) (*SyncState, error)              │
│    func SaveSyncState(path string, update func(*SyncState)) error   │
│      — load → mutate → write (atomic-ish)                          │
│                                                                    │
│  state_test.go [NEW]                                                │
│                                                                    │
│  envelope.go, types.go (unchanged)                                  │
│                                                                    │
│  [DELETED] merge.go, queue.go, merge_test.go, conflict_test.go      │
└────────────────────────────────────────────────────────────────────┘
            ↑
            │ caller (via state.SaveSyncState)
┌────────────────────────────────────────────────────────────────────┐
│ internal/cli/push.go (cleaned up)                                   │
│    state.SaveSyncState(path, func(s *SyncState) {                   │
│        s.VaultID = vaultID                                          │
│    })                                                               │
│ (1-field saveSyncState 함수는 삭제됨)                                │
└────────────────────────────────────────────────────────────────────┘
```

### 2.2 Data Flow — SaveSyncState (callback pattern)

```
caller: cli/push.go OR engine.go (Push / Pull)
  │
  ├─ state.SaveSyncState(path, func(s *SyncState) {
  │      s.VaultID = vaultID         // push fills these
  │      s.Version = version
  │      s.Hash = hash
  │      s.LastPushedAt = pushedAt
  │  })
  ▼
SaveSyncState(path, update)
  │
  ├─ existing, _ := LoadSyncState(path)  // ignore error if not exists
  │     │
  │     └─ if file exists: read JSON → &SyncState{...} (partial OK)
  │     └─ if file missing: existing = &SyncState{}
  │
  ├─ update(existing)  // caller mutates only the fields they own
  │
  ├─ data, err := json.MarshalIndent(existing, "", "  ")
  ├─ os.WriteFile(path, data, 0600)  // still truncate-write, but all 5 fields preserved
  │
  └─ return err
```

### 2.3 Dependencies

| Component | Depends On | Purpose |
|-----------|-----------|---------|
| `internal/sync/state.go` | `encoding/json`, `os`, `fmt` (stdlib only) | File I/O + marshal |
| `internal/sync/state_test.go` | `t.TempDir`, `os.ReadFile` | Unit tests |
| `internal/sync/engine.go` (modified) | `internal/sync/state.go` | Push/Pull use SaveSyncState |
| `internal/cli/push.go` (modified) | `internal/sync/state.go` | Same callback pattern |

---

## 3. Data Model

### 3.1 Entity Definition

```go
// internal/sync/state.go (NEW)

// SyncState is the persisted state of a tene vault's sync history.
// Stored as JSON in `~/.tene/sync_state.json` (or test-specific tmp).
//
// All fields are optional (`omitempty`) so partial updates remain backward
// compatible with older 1-field state files.
type SyncState struct {
    // VaultID is the server-side vault identifier (e.g., UUID).
    VaultID      string `json:"vault_id,omitempty"`

    // Version is the monotonically increasing version of the remote vault.
    Version      int64  `json:"version,omitempty"`

    // Hash is the SHA-256 hex digest of the encrypted envelope last seen.
    Hash         string `json:"hash,omitempty"`

    // LastPushedAt is the RFC 3339 timestamp of the last successful push.
    LastPushedAt string `json:"last_pushed_at,omitempty"`

    // LastPulledAt is the RFC 3339 timestamp of the last successful pull.
    LastPulledAt string `json:"last_pulled_at,omitempty"`
}
```

### 3.2 Entity Relationships

```
[Vault DB] 1 ──── 1 [sync_state.json]  (sibling files in ~/.tene/)
```

### 3.3 File Schema

```json
// ~/.tene/sync_state.json (after v2.0; both push and pull produce this)
{
  "vault_id": "8b1c3d4e-5f67-...",
  "version": 42,
  "hash": "0a1b2c3d...",
  "last_pushed_at": "2026-05-13T10:30:42Z",
  "last_pulled_at": "2026-05-12T18:00:00Z"
}
```

Backward compatibility (legacy v1.0.x file with only `vault_id`):

```json
// ~/.tene/sync_state.json (legacy 1-field; still loadable post-v2.0)
{
  "vault_id": "8b1c3d4e-5f67-..."
}
// LoadSyncState() returns &SyncState{VaultID: "8b1c3d..."}; other fields zero
// First push after upgrade fills in remaining fields atomically
```

---

## 4. API Specification

### 4.1 New Public Functions

#### `LoadSyncState(path string) (*SyncState, error)`

```go
// internal/sync/state.go (NEW)
//
// LoadSyncState reads the JSON file at `path` and returns the deserialized
// state. Missing file returns (&SyncState{}, nil) — caller treats as fresh.
//
// JSON decode errors are wrapped and returned.
func LoadSyncState(path string) (*SyncState, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if errors.Is(err, os.ErrNotExist) {
            return &SyncState{}, nil
        }
        return nil, fmt.Errorf("read sync state: %w", err)
    }
    var s SyncState
    if err := json.Unmarshal(data, &s); err != nil {
        return nil, fmt.Errorf("decode sync state: %w", err)
    }
    return &s, nil
}
```

#### `SaveSyncState(path string, update func(*SyncState)) error`

```go
// internal/sync/state.go (NEW)
//
// SaveSyncState reads the existing state, applies `update`, and writes it back.
// This preserves all fields the caller doesn't touch — fixing the v1.0.8 bug
// where push.go's truncate-write nuked 4 fields per call.
//
// Missing file → fresh SyncState passed to update.
//
// Atomicity note: this uses os.WriteFile (truncate + write). For atomic write
// with rename, see v2.1 (out of scope for Sprint 1).
func SaveSyncState(path string, update func(*SyncState)) error {
    s, err := LoadSyncState(path)
    if err != nil {
        return err
    }
    update(s)
    data, err := json.MarshalIndent(s, "", "  ")
    if err != nil {
        return fmt.Errorf("encode sync state: %w", err)
    }
    if err := os.WriteFile(path, data, 0600); err != nil {
        return fmt.Errorf("write sync state: %w", err)
    }
    return nil
}
```

### 4.2 Caller Migration

#### `internal/sync/engine.go:Push` (after migration)

```go
// internal/sync/engine.go (Push, line ~160 region — replaces existing saveSyncState call)

// 6. Save sync state (5-field merge)
err = state.SaveSyncState(opts.SyncStatePath, func(s *SyncState) {
    s.VaultID = result.VaultID
    s.Version = result.Version
    s.Hash = result.Hash
    s.LastPushedAt = result.PushedAt
    // LastPulledAt: untouched (preserved from last pull)
})
if err != nil {
    slog.Warn("sync.push.state_save_failed", "error", err)
}
```

#### `internal/sync/engine.go:Pull` (after migration)

```go
// internal/sync/engine.go (Pull, line ~226-234 region)

now := time.Now().UTC().Format(time.RFC3339)
err = state.SaveSyncState(opts.SyncStatePath, func(s *SyncState) {
    s.VaultID = manifest.VaultID
    s.Version = manifest.Version
    s.Hash = manifest.Hash
    s.LastPulledAt = now
    // LastPushedAt: untouched (preserved from last push)
})
if err != nil {
    slog.Warn("sync.pull.state_save_failed", "error", err)
}
```

#### `internal/cli/push.go:163` (after migration)

```go
// internal/cli/push.go (replaces old saveSyncState function)

// Old (DELETED):
// func saveSyncState(path, vaultID string) error { ... }

// New: inline use of state.SaveSyncState (only updates vault_id when needed)
err := sync.SaveSyncState(syncStatePath, func(s *sync.SyncState) {
    s.VaultID = vaultID
    // Don't touch other fields — engine.Push handles those after API success
})
if err != nil {
    return fmt.Errorf("save sync state: %w", err)
}
```

### 4.3 Deleted APIs

| Function | File | LOC |
|----------|------|----:|
| `ThreeWayMerge(local, remote, base) (*MergeResult, error)` | `internal/sync/merge.go` | 160 |
| `SyncQueue` struct + methods | `internal/sync/queue.go` | 84 |
| `(*SyncQueue).EnqueueOperation(...)` | `internal/sync/queue.go` | (part of 84) |
| `saveSyncState(path, vaultID string)` (1-field version) | `internal/cli/push.go:163` | 4 |
| `saveSyncState(path, *syncStateFile)` (5-field version, replaced by state.SaveSyncState callback) | `internal/sync/engine.go` | ~10 |

### 4.4 Deleted Test Files

- `internal/sync/merge_test.go` (139 LOC)
- `internal/sync/conflict_test.go` (27 LOC — only tests helpers from merge.go)

`engine_test.go` is **kept** (tests Push/Pull mechanics, not merge).

---

## 5. UI/UX Design

> No user-visible UI change. CLI commands behave identically.

### 5.1 Behavioral Differences (subtle)

| Behavior | Before (v1.0.8) | After (v2.0) |
|----------|----------------|--------------|
| `tene push` → `cat sync_state.json` | 1 field (`vault_id`) only | 4 fields (`vault_id`, `version`, `hash`, `last_pushed_at`) |
| `tene pull` → `cat sync_state.json` | 5 fields (full state) | 5 fields (full state) — but next push doesn't nuke 4 |
| `tene push` then `tene pull` then `tene push` | Final state: 1 field only (push truncated 4 fields back to 1) | Final state: 5 fields preserved |
| `~/.tene/vault.db.base` file presence | exists (dead path; never read) | NOT created |
| `~/.tene/vault.db.backup.20060102-150405` (timestamp backups) | exists | exists (unchanged) |

### 5.2 Page UI Checklist (CLI semantics)

- [ ] `tene push` (first time) → sync_state.json with 4 fields (no last_pulled_at yet)
- [ ] `tene push` (after pull) → sync_state.json with 5 fields (last_pulled_at preserved)
- [ ] `tene push` (running multiple times) → all 4 push fields update; last_pulled_at unchanged
- [ ] `tene pull` (running multiple times) → last_pulled_at updates; last_pushed_at preserved
- [ ] `tene push` on a v1.0.x leftover sync_state.json (1 field only) → upgrades to 4 fields atomically
- [ ] `~/.tene/vault.db.base` NOT created on first pull
- [ ] `~/.tene/vault.db.backup.YYYYMMDD-HHMMSS` still created on pull (safety backup)

---

## 6. Error Handling

### 6.1 Error Code Definition

| Code | Cause | Handling |
|------|-------|----------|
| 1 | `LoadSyncState`: JSON decode failure (corrupted file) | Exit 1 + "sync state corrupted; delete sync_state.json" |
| 1 | `SaveSyncState`: write error (disk full / permission) | Exit 1 + "save sync state: <reason>" |
| (sync engine errors unchanged) | network / API failures | Existing behavior |

### 6.2 Error Response Format

```
# Corrupted sync_state.json
$ tene push
Error: decode sync state: invalid character '{' looking for beginning of value
Exit code: 1

# Permission denied
$ tene push
Error: write sync state: open /home/user/.tene/sync_state.json: permission denied
Exit code: 1
```

---

## 7. Security Considerations

- [x] **File mode 0600**: `os.WriteFile(path, data, 0600)` — owner read/write only (existing pattern)
- [x] **No sensitive data in state**: vault_id, version, hash, timestamps only — no master key / secrets
- [x] **Hash is SHA-256 of ciphertext envelope**: ciphertext is already authenticated (AAD); hash is integrity check
- [x] **JSON injection**: `json.Marshal` escapes properly; no manual string concatenation
- [x] **gosec G404** (math/rand jitter at engine.go:407): annotated with `// #nosec G404 -- jitter only`. Plan/Design defers true crypto/rand jitter to v2.1 perf optimization (P2-Perf1 follow-up)
- [x] **Race condition**: SQLite WAL handles vault.db concurrency; sync_state.json is single-writer per process (no goroutine race in CLI flows)

---

## 8. Test Plan

### 8.1 Test Scope

| Type | Target | Tool | Phase |
|------|--------|------|-------|
| L1: Unit | `LoadSyncState` (missing file, valid file, corrupted file) | `go test` | Do |
| L1: Unit | `SaveSyncState` (fresh write, partial update preserves others) | `go test` | Do |
| L1: Unit | Backward compat — legacy 1-field JSON → SyncState{VaultID only} | `go test` | Do |
| L2: Integration | `tene push` after legacy 1-field state file → 4 fields persist | testhelper_test.go | Do |
| L2: Integration | `tene push` then `tene pull` then `tene push` → all 5 fields preserved | testhelper_test.go | Do |
| L3: Deletion verification | grep `ThreeWayMerge\|SyncQueue` = 0 | shell test | Check |
| L3: Deletion verification | grep `vault.db.base` = 0 | shell test | Check |
| L3: Deletion verification | grep `func saveSyncState` = 1 (state.go only) | shell test | Check |

### 8.2 L1: Unit Test Scenarios

| # | Test | Setup | Expected |
|---|------|-------|---------|
| 1 | `LoadSyncState(missing)` | path doesn't exist | `&SyncState{}, nil` |
| 2 | `LoadSyncState(legacy)` | file: `{"vault_id":"abc"}` | `&SyncState{VaultID:"abc"}, nil`; other fields zero |
| 3 | `LoadSyncState(full)` | file: 5 fields | full struct; nil error |
| 4 | `LoadSyncState(corrupted)` | file: `{invalid json` | error wrapped "decode sync state" |
| 5 | `SaveSyncState(fresh)` | path missing, update sets VaultID | new file with 1 field |
| 6 | `SaveSyncState(merge)` | file has 5 fields, update sets VaultID only | output has 5 fields — VaultID overwritten, 4 preserved |
| 7 | `SaveSyncState(missing dir)` | parent dir doesn't exist | error wrapped "write sync state" |
| 8 | `SaveSyncState(no update)` | update is no-op | file rewritten unchanged (still 5 fields) |

### 8.3 L2: Integration Test Scenarios

| # | Scenario | Expected |
|---|----------|----------|
| 1 | New vault → `tene push` (mock API) → cat sync_state.json | 4 fields (vault_id, version, hash, last_pushed_at); no last_pulled_at |
| 2 | After push → `tene pull` → cat sync_state.json | 5 fields (last_pulled_at added) |
| 3 | After pull → `tene push` again → cat sync_state.json | 5 fields preserved (last_pulled_at NOT nuked) |
| 4 | Pre-existing 1-field state file + `tene push` | 4 fields populated post-push |
| 5 | `tene pull` on fresh setup | no `~/.tene/vault.db.base` created (verify with stat) |
| 6 | `tene pull` on fresh setup | `~/.tene/vault.db.backup.20060102-150405` created (timestamp backup preserved) |

### 8.4 L3: Deletion Verification Scenarios

| # | Verification | Expected |
|---|--------------|----------|
| 1 | `find . -name 'merge.go' -o -name 'queue.go' -o -name 'merge_test.go' -o -name 'conflict_test.go' -path './internal/sync/*'` | 0 results |
| 2 | `grep -rn 'ThreeWayMerge\|SyncQueue\|EnqueueOperation' --include='*.go'` | 0 results |
| 3 | `grep -rn 'func saveSyncState' --include='*.go'` | 0 results (replaced by state.SaveSyncState) |
| 4 | `grep -rn 'vault.db.base' --include='*.go'` | 0 results |
| 5 | `grep -c 'defer.*resp.Body.Close()' internal/sync/engine.go` | ≥ 4 (existing http.Do sites) |
| 6 | `go build ./...` | 0 errors |
| 7 | `go test -race -count=10 ./internal/sync/...` | 0 failures |
| 8 | `go vet ./...` | 0 issues |

### 8.5 Seed Data Requirements

| Entity | Min Count | Key Fields |
|--------|:---------:|------------|
| Legacy `sync_state.json` (1 field) | 1 fixture (for backcompat test) | `{"vault_id":"test-vault-id"}` |
| Full `sync_state.json` (5 fields) | 1 fixture | All 5 populated |

> Fixtures at `internal/sync/testdata/sync_state_legacy.json` + `sync_state_full.json`.

---

## 9. Clean Architecture

### 9.1 Layer Structure

| Layer | Component | Location |
|-------|-----------|----------|
| **Presentation** | `cli/push.go` (uses state.SaveSyncState directly) | `internal/cli/push.go` |
| **Application** | `engine.go` Push/Pull (uses state.SaveSyncState) | `internal/sync/engine.go` |
| **Domain** | (none — state is data persistence only) | n/a |
| **Infrastructure** | `state.go` (file I/O + JSON marshal) | `internal/sync/state.go` |

### 9.2 Dependency Rules

```
cli/push.go ─→ internal/sync/state.go
                       │
                       │ (uses stdlib only: os, json, errors, fmt)
                       ▼
                  (no further deps)

internal/sync/engine.go ─→ internal/sync/state.go
```

### 9.3 File Import Rules

| From | Imports | Forbidden |
|------|---------|-----------|
| `internal/sync/state.go` | `encoding/json`, `errors`, `fmt`, `os` (stdlib only) | `internal/cli/*`, `pkg/crypto` |
| `internal/sync/engine.go` | `internal/sync/state`, `pkg/crypto`, `net/http`, `pkg/errors` | `internal/cli/*` |
| `internal/cli/push.go` | `internal/sync`, `internal/vault`, `pkg/errors` | n/a |

### 9.4 This Feature's Layer Assignment

| Component | Layer | Location |
|-----------|-------|----------|
| `SyncState` struct + JSON tags | Infrastructure (data) | `internal/sync/state.go` |
| `LoadSyncState`, `SaveSyncState` | Infrastructure (I/O) | `internal/sync/state.go` |
| `Engine.Push / Pull` (modified callers) | Application | `internal/sync/engine.go` |
| `cmd/push.go RunE` (modified caller) | Presentation | `internal/cli/push.go` |

---

## 10. Coding Convention Reference

### 10.1 Naming Conventions

| Target | Rule | Example |
|--------|------|---------|
| Struct exported | `PascalCase` | `SyncState` |
| Function exported | `PascalCase` | `LoadSyncState`, `SaveSyncState` |
| JSON tags | `snake_case` + `omitempty` | `"vault_id,omitempty"` |
| Files | `lowercase.go` | `state.go`, `state_test.go` |

### 10.2 Import Order

```go
package sync

import (
    // 1. Standard library
    "encoding/json"
    "errors"
    "fmt"
    "os"

    // (no external, no internal for state.go — Domain isolation)
)
```

### 10.3 Environment Variables

None for this design.

---

## 11. Implementation Guide

### 11.1 File Structure

```
tene/internal/sync/
├── engine.go              # MODIFIED (lines 211-219 deleted; saveSyncState refs swapped)
├── envelope.go            # UNCHANGED
├── types.go               # UNCHANGED
├── merge.go               # DELETED (160 LOC)
├── queue.go               # DELETED (84 LOC)
├── merge_test.go          # DELETED (139 LOC)
├── conflict_test.go       # DELETED (27 LOC)
├── state.go               # NEW (~80 LOC)
├── state_test.go          # NEW (~140 LOC)
└── testdata/              # NEW
    ├── sync_state_legacy.json
    └── sync_state_full.json
```

### 11.2 Implementation Order

> **PR #9** (`refactor(sync): remove merge.go + queue.go + base snapshot path`):

1. [ ] Pre-flight grep: confirm 0 production callers of `ThreeWayMerge`, `SyncQueue`, `EnqueueOperation`
2. [ ] `git rm internal/sync/merge.go`
3. [ ] `git rm internal/sync/merge_test.go`
4. [ ] `git rm internal/sync/queue.go`
5. [ ] `git rm internal/sync/conflict_test.go`
6. [ ] `internal/sync/engine.go:211-219` — delete `vault.db.base` write path (keep timestamp backup at line 217)
7. [ ] `internal/sync/engine.go:407` — add `// #nosec G404 -- jitter only, not security-critical` annotation
8. [ ] Verify: `go build ./...` 0 errors; `go test -race -count=10 ./internal/sync/...` 0 failures
9. [ ] CHANGELOG.md — "refactor(sync): remove 244 LOC dead merge code (ThreeWayMerge, SyncQueue, base snapshot path). Sync remains overwrite-only until v1.2 merge UX (new architecture)."

> **PR #10** (`refactor(sync): unify saveSyncState into sync/state.go (5-field)`):

10. [ ] `internal/sync/state.go` (NEW) — `SyncState` struct + `LoadSyncState` + `SaveSyncState`
11. [ ] `internal/sync/testdata/sync_state_legacy.json` (1 field)
12. [ ] `internal/sync/testdata/sync_state_full.json` (5 fields)
13. [ ] `internal/sync/state_test.go` (NEW) — 8 unit tests (see §8.2)
14. [ ] `internal/sync/engine.go` (Push and Pull regions) — replace direct `saveSyncState(...)` calls with `state.SaveSyncState(path, func(s *SyncState) {...})` callback
15. [ ] `internal/sync/engine.go` — remove the `func saveSyncState(...)` definition (no longer needed; replaced by state.SaveSyncState)
16. [ ] `internal/sync/engine.go` — remove the `type syncStateFile struct` (replaced by `state.SyncState`)
17. [ ] `internal/cli/push.go:163` — delete `func saveSyncState(path, vaultID string) error`
18. [ ] `internal/cli/push.go` (caller site) — use `sync.SaveSyncState(path, func(s *sync.SyncState) { s.VaultID = vaultID })`
19. [ ] Verify: grep `func saveSyncState` = 1 site (state.go); grep `type syncStateFile` = 0
20. [ ] Integration test: `tene push` (mock) → verify 4-field output; `tene push` after legacy 1-field → 4 fields after
21. [ ] CHANGELOG.md — "refactor(sync): unify saveSyncState into internal/sync/state.go (single source of truth; 5-field schema preserved across push/pull)"

### 11.3 Session Guide

#### Module Map

| Module | Scope Key | Description | Estimated Turns |
|--------|-----------|-------------|:---------------:|
| Dead code removal | `module-1` | PR #9 — delete merge.go + queue.go + base path | 10-15 |
| State consolidation | `module-2` | PR #10 — state.go new + 2 caller migrations + tests | 25-30 |

#### Recommended Session Plan

| Session | Phase | Scope | Turns |
|---------|-------|-------|:-----:|
| Session 1 | Plan + Design | (this document) | 20-25 (current) |
| Session 2 | Do | `--scope module-1` (PR #9) | 15-20 |
| Session 3 | Do | `--scope module-2` (PR #10) | 25-30 |
| Session 4 | Check + Report | Both PRs | 15-20 |

---

## 12. Edge Cases & Failure Modes

| # | Scenario | Behavior |
|---|----------|----------|
| 1 | `SaveSyncState` called with `update` callback that panics | Panic propagates; caller responsibility |
| 2 | Concurrent `SaveSyncState` calls (race) | Last-write-wins; not protected (single-writer CLI flow assumed) |
| 3 | Disk full mid-write | `os.WriteFile` returns error; file may be partially written (truncate-write semantic). Caller can retry. v2.1 atomic rename addresses |
| 4 | Network failure during sync but state file already written | Engine errors fire; state file may have stale data (caller handles via re-push) |
| 5 | Legacy `sync_state.json` with extra fields (e.g., v0.5 experimental) | `json.Unmarshal` ignores unknown fields; safe |
| 6 | `update` callback that doesn't set any fields | File rewritten unchanged (no-op); valid |
| 7 | `LoadSyncState` race with `SaveSyncState` (read while write) | Possible truncated read → JSON parse error → caller exits 1. Acceptable for CLI; not a server |
| 8 | `vault.db.base` left over from v1.0.x | File remains untouched on disk (v2.0 doesn't read or write it); user can delete manually |

---

## 13. Sequence Diagram — `tene push` (after migration)

```
User           cli/push.go        sync/state.go        sync/engine.go      filesystem
 │                  │                    │                    │                │
 │ tene push        │                    │                    │                │
 ├─────────────────▶│                    │                    │                │
 │                  │ (existing flow:    │                    │                │
 │                  │  loadApp, etc.)    │                    │                │
 │                  │                    │                    │                │
 │                  │ Engine.Push()      │                    │                │
 │                  ├────────────────────┼───────────────────▶│                │
 │                  │                    │                    │ (HTTP POST API)│
 │                  │                    │                    │ (encrypt env)  │
 │                  │                    │                    │ ...            │
 │                  │                    │                    │ result {VaultID, Version, Hash, PushedAt}
 │                  │                    │ SaveSyncState(path, update)         │
 │                  │                    │◀───────────────────┤                │
 │                  │                    │ LoadSyncState(path)                  │
 │                  │                    ├──────────────────────────────────────▶│
 │                  │                    │◀─── existing (4 fields possibly) ────│
 │                  │                    │ update(existing) — sets 4 fields    │
 │                  │                    │ json.Marshal(existing)               │
 │                  │                    │ os.WriteFile(path, ...) — 5 fields preserved
 │                  │                    ├──────────────────────────────────────▶│
 │                  │                    │◀───────────── OK ────────────────────│
 │                  │                    ├───────────────────▶│                │
 │                  │◀───────────────────┤                    │                │
 │ "Pushed: ok"     │                    │                    │                │
 │◀─────────────────┤                    │                    │                │
```

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-05-13 | Initial draft (Sprint 1 design phase, L2 boundary) | cto-lead (code-analyzer perspective) |
