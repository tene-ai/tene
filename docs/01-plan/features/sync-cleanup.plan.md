---
template: plan
version: 1.3
description: tene CLI v2.0 Sprint 1 — sync-cleanup (P0-S1 dead code 244 LOC 제거 + P0-S2 saveSyncState 통합)
variables:
  - feature: sync-cleanup
  - displayName: "sync-cleanup (P0-S1 + P0-S2)"
  - sprint: tene-cli-v2-s1
  - date: 2026-05-13
  - author: cto-lead (PM perspective)
  - project: tene CLI
  - version: v2.0.0 (target)
  - masterPlan: docs/01-plan/features/tene-cli-v2-2026q3.master-plan.md
  - trustLevel: L2
---

# sync-cleanup Planning Document

> **Summary**: `internal/sync/merge.go` (160 LOC) + `queue.go` (84 LOC) + 관련 test 가 0 production caller. 동시에 `saveSyncState` 가 2곳에 다른 schema 로 정의되어 매 push 마다 4 필드 데이터 손실. dead code 제거 + state file 통합.
>
> **Project**: tene CLI
> **Version**: target v2.0.0 (current v1.0.8)
> **Sprint**: `tene-cli-v2-s1` (W1-W2, 2026-05-13 → 2026-05-26)
> **Author**: cto-lead
> **Date**: 2026-05-13
> **Status**: Draft (L2)
> **Master Plan**: [tene-cli-v2-2026q3.master-plan.md](./tene-cli-v2-2026q3.master-plan.md) §2.2 P0-S1, P0-S2, §3.1 보정 4
> **Baseline 분석**: [cli-sprint-master-plan-2026-05-13.md](../cli-sprint-master-plan-2026-05-13.md) §2.4 (보정 4) + §6.1.3 Trk-B T1-B1~B5
> **감사 출처**: [cli-completeness-audit-2026-05-11.md](../../03-report/cli-completeness-audit-2026-05-11.md) DR-2

---

## Executive Summary

| Perspective | Content |
|-------------|---------|
| **Problem** | (1) **244 LOC dead code**: `internal/sync/merge.go` 의 `ThreeWayMerge` (160 LOC) + `internal/sync/queue.go` 의 `SyncQueue.EnqueueOperation` (84 LOC) 가 production 0 호출. `Pull` 은 unconditional overwrite (engine.go:222) 로 base snapshot/merge 안 함. (2) **데이터 손실 결함**: `saveSyncState` 가 `engine.go` (5 필드 schema) 와 `cli/push.go:163` (1 필드 schema) 두 곳에 정의되어 **같은 파일 `sync_state.json` 에 다른 schema 로 truncate-write**. 매 push 마다 `version`, `hash`, `last_pushed_at`, `last_pulled_at` 4 필드 silent 손실. 보정 §2.4 — 단순 race 가 아니라 **데이터 손실 defect**. |
| **Solution** | (1) `merge.go` + `merge_test.go` + `queue.go` + `conflict_test.go` 완전 삭제 (caller 0 확인 후). PR 메시지: "post-mortem: 244 LOC removed; sync currently overwrite-only until v1.2 merge UX." (2) `saveSyncState` 통합: `internal/sync/state.go` 신설 → 5-field schema 만 유지 (`vault_id`, `version`, `hash`, `last_pushed_at`, `last_pulled_at`). `cli/push.go:163` 의 truncate-write 버전 삭제; 기존 record 있으면 vault_id 만 update. (3) Pull 의 `vault.db.base` 생성 코드 삭제 (ThreeWayMerge 의존 dead path). (4) `engine.go:407` math/rand jitter 에 `// #nosec G404` 주석 추가 (gosec 활성화 사전). |
| **Function/UX Effect** | 사용자 가시: 회귀 0 (Pull 은 이미 overwrite-only; merge 는 한 번도 동작 안 했음). 내부: `sync_state.json` 5 필드 모두 유지 → 멀티-환경 push/pull 시 정확한 hash drift detection 가능. |
| **Core Value** | 코드베이스 신뢰. 244 LOC dead removal + 데이터 손실 fix 는 모든 후속 sprint (vault-v2-migration, biometric, supply chain) 의 prerequisite. |

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | 244 LOC dead code + saveSyncState 데이터 손실 결함 동시 존재. dead code 가 unused linter 비활성을 강제 (P0-CI3) → s3 의 lint 강화 차단. 데이터 손실은 user-facing silent corruption (sync_state.json hash 사라짐 → 향후 conflict detection 불가). |
| **WHO** | **Solo founder (10%)** — multi-environment push/pull 의존 → state 정확성이 daily UX. **Sec-conscious team (15%)** — 244 LOC dead code 가 audit attack surface. |
| **RISK** | (a) `merge.go` 삭제가 향후 v1.2 merge UX (CRDT or OT) 복원 시 비용 ↑ (mitigation: PR 메시지에 명시 + s4 design spike 예약). (b) `saveSyncState` 통합 시 기존 sync_state.json 호환 (mitigation: 5-field 모두 optional; vault_id 만 필수). (c) cloud 명령 (push/pull/login/logout/sync_cmd) 는 s3 의 build tag 격리 대상 — 본 plan 의 변경이 build tag 후에도 동작 검증 필요. |
| **SUCCESS** | grep `ThreeWayMerge\|SyncQueue\|EnqueueOperation` = 0 (전체 codebase); grep `func saveSyncState` = 1 (state.go 만); 매 push 후 sync_state.json 의 5 필드 모두 유지 (integration test); `internal/sync/` 비-테스트 LOC -303 / 테스트 LOC -86 / 신규 +80 = 순감 -309 LOC. |
| **SCOPE** | Sprint 1 Trk-B 단독 (2 PR: #9 dead code 제거 + #10 saveSyncState 통합). 1 dev × 3 dev-day. 변경 범위: internal/sync/* (4 파일 삭제 + 1 신규) + internal/cli/push.go (-50 LOC). |

---

## 1. Overview

### 1.1 Purpose

코드베이스 신뢰 회복 + s3 의 unused linter 활성화 prerequisite + 사용자-facing 데이터 손실 fix.

### 1.2 Background

#### 1.2.1 Dead Code 의 발견 (P0-S1)

```
$ grep -rn "ThreeWayMerge\|SyncQueue\|EnqueueOperation" --include='*.go'
internal/sync/merge.go:41:    func ThreeWayMerge(...) (...)
internal/sync/merge_test.go:32:    result, err := ThreeWayMerge(...)
internal/sync/queue.go:12:    type SyncQueue struct {
internal/sync/queue.go:84:    func (q *SyncQueue) EnqueueOperation(...)

# Production caller 검색 (test 제외)
$ grep -rn "ThreeWayMerge" --include='*.go' | grep -v _test.go
internal/sync/merge.go:41:    func ThreeWayMerge(...)  # 정의 자체만
# = 0 production caller
```

`internal/sync/engine.go:174-242` 의 `Pull` 함수:
- line 174-210: download + checksum + decrypt
- line 211-220: `vault.db.base` 생성 (3-way merge 용; **dead** — merge 호출 0)
- line 222: `os.WriteFile(opts.VaultDBPath, plaintext, 0600)` — **unconditional overwrite**
- line 226+: `saveSyncState` 호출

= **Pull 은 overwrite-only**. 사용자가 local 변경한 vault 와 remote 가 conflict 시 silent overwrite. UX 불완전하지만 v1.0.x 의 contract.

#### 1.2.2 saveSyncState 데이터 손실 (P0-S2, 보정 §2.4)

**1차 plan** 은 "중복 정의로 race condition" 이라 표기. 코드 정독 결과 더 심각:

```go
// internal/sync/engine.go (5-field schema)
type syncStateFile struct {
    VaultID      string `json:"vault_id"`
    Version      int64  `json:"version"`
    Hash         string `json:"hash"`
    LastPushedAt string `json:"last_pushed_at"`
    LastPulledAt string `json:"last_pulled_at"`
}
func saveSyncState(path string, s *syncStateFile) error { /* 5 필드 marshal */ }

// internal/cli/push.go:163 (1-field schema)
func saveSyncState(path, vaultID string) error {
    data, _ := json.Marshal(map[string]string{"vault_id": vaultID})
    return os.WriteFile(path, data, 0600)  // ← truncate-write
}
```

같은 파일 `sync_state.json` 에 다른 schema 로 truncate-write — **매 push 마다 engine 이 저장한 version/hash/last_pushed_at/last_pulled_at silent 손실**. 매 push 가 4 필드를 nuke.

#### 1.2.3 unused linter 비활성 (P0-CI3 / 보정 §2.7)

`.golangci.yml:16` 의 `disable: [unused]` 가 명시 → cloud 명령 6개 (push/pull/login/logout/sync_cmd/billing/team) 의 **1,481 LOC dead code** 가 매 빌드 시 컴파일됨 (unused 만 무력화). 본 plan 은 sync 만 처리; cloud 명령 격리는 s3 의 `lint-hardening` 의 `//go:build cloud` 빌드 태그.

### 1.3 Related Documents

- Master Plan: [`tene-cli-v2-2026q3.master-plan.md`](./tene-cli-v2-2026q3.master-plan.md)
- Baseline: [`cli-sprint-master-plan-2026-05-13.md §2.4 §6.1.3 Trk-B`](../cli-sprint-master-plan-2026-05-13.md)
- 감사: [`cli-completeness-audit-2026-05-11.md DR-2`](../../03-report/cli-completeness-audit-2026-05-11.md)
- Sibling: [`passwd-verify.plan.md`](./passwd-verify.plan.md), [`crypto-v2-keys.plan.md`](./crypto-v2-keys.plan.md), [`audit-reader.plan.md`](./audit-reader.plan.md)

---

## 2. Scope

### 2.1 In Scope

- [ ] `internal/sync/merge.go` 완전 삭제 (160 LOC)
- [ ] `internal/sync/merge_test.go` 완전 삭제 (139 LOC)
- [ ] `internal/sync/queue.go` 완전 삭제 (84 LOC)
- [ ] `internal/sync/conflict_test.go` 완전 삭제 (27 LOC, conflict 헬퍼만 사용 → dead)
- [ ] `internal/sync/state.go` (신규) — `saveSyncState(path, state)` 단일 정의; 5-field schema; existing record 보존 (vault_id 만 update)
- [ ] `internal/cli/push.go:163` — `saveSyncState` 1-field 버전 삭제 (engine.go 의 함수 호출하도록 변경)
- [ ] `internal/sync/engine.go:211-219` — `vault.db.base` 생성 코드 삭제 (ThreeWayMerge dead path); `vault.db.backup.*` timestamp 백업은 유지 (안전망)
- [ ] `internal/sync/engine.go:407` — math/rand jitter 에 `// #nosec G404 -- jitter only, not security-critical` 주석 (s3 의 gosec 활성화 사전)
- [ ] `internal/sync/state_test.go` (신규) — saveSyncState 통합 후 5 필드 모두 유지 검증
- [ ] (verify only) `engine.go` 의 4 `http.Do` 사이트 `defer resp.Body.Close()` 보유 — bodyclose lint 통과 사전 검증
- [ ] CHANGELOG.md — "refactor(sync): remove 244 LOC dead merge code; unify saveSyncState into state.go (5-field schema)"

### 2.2 Out of Scope

- merge UX 의 새 architecture (CRDT or OT) — **v1.2 stretch** + s4 design spike 예약 (T3-D 가 biometric design 인데 별도 spike 추가)
- cloud 명령의 `//go:build cloud` 빌드 태그 — **Sprint 3** `lint-hardening` (T3-B2)
- `engine.go:82` 의 vault.New 중복 open 해소 — **Sprint 2** `sync-contracts` (T2-A6)
- sync engine 단위 테스트 (mock Transport) — **Sprint 2** `test-infra` (T2-B5)
- `engine.go` 의 `Pull` 함수의 backup retention policy (현재 timestamp 백업 무한 누적) — v2.1 (backup pruning command)
- `sync_state.json` 의 schema_migrations (forward compatibility) — Sprint 2 `vault-v2-migration` 의 schema_migrations 메커니즘 일반화

---

## 3. Requirements

### 3.1 Functional Requirements

| ID | Requirement | Priority | Status |
|----|-------------|----------|--------|
| FR-01 | `merge.go`, `merge_test.go`, `queue.go`, `conflict_test.go` 완전 삭제 | **P0** | Pending |
| FR-02 | grep `ThreeWayMerge\|SyncQueue\|EnqueueOperation` = 0 (전체 codebase) | **P0** | Pending |
| FR-03 | `internal/sync/state.go` 신설 — `saveSyncState(path string, state *SyncState) error` 단일 정의 | **P0** | Pending |
| FR-04 | `SyncState` struct — `VaultID`, `Version`, `Hash`, `LastPushedAt`, `LastPulledAt` 5 필드 (existing schema 유지) | **P0** | Pending |
| FR-05 | `saveSyncState` 가 existing sync_state.json 을 load → merge → write (full replace 아님) — vault_id 만 update 시 다른 4 필드 보존 | **P0** | Pending |
| FR-06 | `cli/push.go:163` 의 saveSyncState 1-field 버전 삭제; `state.saveSyncState(path, &SyncState{VaultID: vaultID})` 호출 | **P0** | Pending |
| FR-07 | `engine.go:211-216` `vault.db.base` 생성 코드 삭제 (3-way merge dead path) | **P0** | Pending |
| FR-08 | `engine.go` `Pull` 의 timestamp 백업 (`vault.db.backup.20060102-150405`) 유지 (안전망) | **P0** | Pending |
| FR-09 | `engine.go:407` math/rand jitter 주석 — `// #nosec G404 -- jitter only` | P1 | Pending |
| FR-10 | grep `func saveSyncState` = 1 (state.go 단 1곳) | **P0** | Pending |
| FR-11 | integration test — push 후 sync_state.json 5 필드 모두 유지 | **P0** | Pending |
| FR-12 | grep `vault.db.base` = 0 | **P0** | Pending |

### 3.2 Non-Functional Requirements

| Category | Criteria | Measurement Method |
|----------|----------|-------------------|
| Code Health | LOC -309 (244 dead + saveSyncState consolidation) | `wc -l` before/after |
| Linter Compliance | unused linter 활성화 시 0 issue (s3 에서 enable) | `golangci-lint run --enable=unused` |
| Backward Compatibility | 기존 sync_state.json 5-field 모두 호환 (read + write) | integration test |
| Build Stability | `go build ./...` 0 회귀 | CI matrix (s3) |
| Test Stability | 삭제된 test 파일 제거 후 `go test ./internal/sync/...` 0 회귀 | `go test -race -count=10` |

---

## 4. Success Criteria

### 4.1 Definition of Done

- [ ] FR-01 ~ FR-12 모두 충족 (grep + integration test 통과)
- [ ] `internal/sync/state.go` + `state_test.go` 신설
- [ ] `internal/cli/push.go:163` saveSyncState 삭제
- [ ] `engine.go:211-219` vault.db.base 삭제
- [ ] LOC count: 비-테스트 -303 / 테스트 -86 + 신규 +80 (총 -309)
- [ ] `go test -race -count=10 ./internal/sync/... ./internal/cli/...` 0 회귀
- [ ] CHANGELOG.md "refactor(sync)" 엔트리
- [ ] PR description 에 dead code 삭제 evidence (grep 결과) 첨부
- [ ] gap-detector Match Rate ≥ 90% (M8)

### 4.2 Quality Criteria

- [ ] `internal/sync/` total coverage 유지 또는 향상 (dead test 제거로 coverage % 변화 측정)
- [ ] linter clean (현재 6 linter; s3 의 unused 활성화 후에도 clean)
- [ ] 변경 후 모든 cloud 명령 (push/pull/sync/login/logout) 컴파일 통과 (build tag 없이; s3 의 build tag 와 호환)

---

## 5. Risks and Mitigation

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| `merge.go` 삭제가 향후 v1.2 merge UX 복원 시 비용 ↑ | L | M | PR 메시지에 "merge UX 는 v1.2 에서 새 architecture (CRDT or OT)" 명시 + s4 design spike 예약 (현재 master plan §3.3 의 trk-D 추가 사항으로 fold) |
| 기존 sync_state.json (1-field 버전) 호환 깨짐 | M | M | `state.go` 가 file load 시 partial field tolerance (`omitempty` JSON tags); vault_id 만 있는 file 도 load 성공 |
| `vault.db.base` 생성 코드 삭제가 backup chain 망가뜨림 | M | L | timestamp 백업 (`vault.db.backup.20060102-150405`) 유지 — `.base` 는 dead path |
| s3 의 cloud 빌드 태그 도입 시 본 plan 변경과 충돌 | M | M | 본 plan 은 cloud 명령 자체는 건드리지 않음 (`cli/push.go:163` 만); s3 의 build tag 가 cloud commands 격리 시 본 plan 의 변경은 build tag 안에 fall |
| `engine.go:407` math/rand jitter 가 gosec 외 다른 linter 도 트리거 (예: gocritic) | L | L | `// #nosec G404 -- jitter only, not security-critical` 주석으로 충분; 다른 linter 는 s3 의 lint-hardening 에서 처리 |
| 4 파일 삭제가 git history 추적 어렵게 만듦 | L | L | `git rm` 명시 + PR description 에 삭제된 파일 명시 |

---

## 6. Impact Analysis

### 6.1 Changed Resources

| Resource | Type | Change Description |
|----------|------|--------------------|
| `internal/sync/merge.go` | File | DELETE (160 LOC) |
| `internal/sync/merge_test.go` | File | DELETE (139 LOC) |
| `internal/sync/queue.go` | File | DELETE (84 LOC) |
| `internal/sync/conflict_test.go` | File | DELETE (27 LOC) |
| `internal/sync/state.go` | File | NEW (~80 LOC) |
| `internal/sync/state_test.go` | File | NEW (~140 LOC) |
| `internal/sync/engine.go:174-242` | Function | Pull — line 211-219 base snapshot 삭제 |
| `internal/sync/engine.go:407` | Statement | `// #nosec G404` 주석 추가 |
| `internal/cli/push.go:163` | Function | `saveSyncState` 1-field 버전 삭제 |
| `sync_state.json` (runtime) | File schema | 5-field 유지 (기존과 동일) |

### 6.2 Current Consumers

| Resource | Operation | Code Path | Impact |
|----------|-----------|-----------|--------|
| `ThreeWayMerge` | Merge | `merge.go` 내부 + `merge_test.go` | **DELETE** (production caller 0) |
| `SyncQueue.EnqueueOperation` | Queue | `queue.go` 내부 | **DELETE** (production caller 0) |
| `saveSyncState` (engine.go) | Sync state write | `engine.go:?` (Push line 161, Pull line 227) | **KEEP** but moved to state.go |
| `saveSyncState` (push.go:163) | Sync state write | `internal/cli/push.go:?` | **DELETE** (state.go 호출로 교체) |
| `vault.db.base` | 3-way merge base | `engine.go:213-215` (write) — 어디서도 read 안 함 | **DELETE** (dead write) |
| `vault.db.backup.{timestamp}` | Safety backup | `engine.go:217` | **KEEP** |
| math/rand `Intn` (engine.go:407) | Jitter | retry backoff | **KEEP** + nosec 주석 |
| `sync_state.json` (KV file) | State persist | engine + cli/push | **CONSOLIDATE** (5-field schema) |

### 6.3 Verification

- [ ] grep `ThreeWayMerge\|SyncQueue\|EnqueueOperation` = 0
- [ ] grep `vault.db.base` = 0
- [ ] grep `func saveSyncState` = 1
- [ ] integration test: `tene push` → `cat sync_state.json | jq '.vault_id, .version, .hash, .last_pushed_at, .last_pulled_at'` 모두 non-null
- [ ] integration test: `tene push && tene pull` 후 5 필드 다 유지
- [ ] `go test -race -count=10 ./internal/sync/...` 0 회귀

---

## 7. Architecture Considerations

### 7.1 Project Level Selection

| Level | Selected |
|-------|:--------:|
| Enterprise (tene CLI) | **☑** |

### 7.2 Key Architectural Decisions

| Decision | Options | Selected | Rationale |
|----------|---------|----------|-----------|
| dead code 처리 | (a) Deprecated marker 후 v3 제거, (b) 즉시 삭제, (c) build tag 격리 | (b) 즉시 삭제 | 244 LOC 가 unused linter 비활성 강제 → s3 의 lint-hardening 차단; deprecated 단계 의미 없음 (caller 0) |
| saveSyncState 통합 위치 | (a) `internal/sync/state.go` 신규, (b) `engine.go` 안에 통합, (c) `pkg/sync/` 외부화 | (a) state.go 분리 | Single Responsibility; test 분리; engine.go (472 LOC) 추가 비대 방지 |
| existing record merge 전략 | (a) full replace, (b) field-level merge (omitempty), (c) version-aware migration | (b) field-level merge | push 가 vault_id 만 가지고 호출해도 기존 4 필드 보존 |
| vault.db.base 처리 | (a) keep + 향후 merge UX 부활 대비, (b) delete | (b) delete | merge UX 가 v1.2 stretch — 그 때 새 architecture 로 다시 도입; 현재는 dead write |
| backup retention | (a) timestamp 백업 무한 보존, (b) latest N 만 보존, (c) backup 자체 제거 | (a) 유지 + v2.1 backup pruning command | 현재 사용자 데이터 안전 우선; pruning 은 별도 feature |

### 7.3 Clean Architecture Approach

```
Layer mapping (Sprint 1 scope):
┌──────────────────────────────────────────────────────────┐
│ Presentation (internal/cli/)                             │
│   - push.go (saveSyncState 호출 → state.go 위임)          │
├──────────────────────────────────────────────────────────┤
│ Application (internal/sync/)                             │
│   - engine.go (Push/Pull; saveSyncState 호출)             │
│   - state.go (NEW; SyncState struct + saveSyncState)     │
├──────────────────────────────────────────────────────────┤
│ Domain (없음 — sync 는 application 레벨에서 완결됨)        │
├──────────────────────────────────────────────────────────┤
│ Infrastructure (none in this plan; net/http은 standard)   │
└──────────────────────────────────────────────────────────┘
```

---

## 8. Convention Prerequisites

### 8.1 Existing Project Conventions

- [x] `internal/sync/` 패키지 (6 src + 4 test)
- [x] JSON state file pattern (`sync_state.json`)
- [x] Error handling: `pkg/errors` (s4 에서 `teneerr` rename)

### 8.2 Conventions to Define/Verify

| Category | Current State | To Define | Priority |
|----------|---------------|-----------|:--------:|
| Sync state file schema | engine.go 의 `syncStateFile` (private) | `internal/sync/state.go` 의 `SyncState` (public) | High |
| nosec annotation 패턴 | (없음) | `// #nosec G404 -- 이유` 표준 | Medium |
| dead code 처리 정책 | (없음) | git rm + PR description evidence | Low |

### 8.3 Environment Variables Needed

None.

---

## 9. Testing Plan

### 9.1 Test Scope

| Type | Target | Tool | Phase |
|------|--------|------|-------|
| L1: Unit | `state.go:saveSyncState`, `SyncState` JSON marshal/unmarshal | `go test` | Do |
| L2: Integration | `tene push` after fixture sync_state.json (1-field) → 5-field merge | testhelper_test.go | Do |
| L3: Regression | `tene push` + `tene pull` 후 state 5 필드 모두 유지 | integration_test.go | Check |

### 9.2 Test Scenarios

| # | Scenario | Expected |
|---|----------|---------|
| T-1 | `saveSyncState(path, &SyncState{VaultID: "abc"})` — empty path | new file, 1 field non-empty |
| T-2 | existing 5-field file → saveSyncState with VaultID only — 4 fields preserved | match |
| T-3 | grep `ThreeWayMerge\|SyncQueue\|EnqueueOperation` = 0 (production code) | match |
| T-4 | grep `func saveSyncState` = 1 | match |
| T-5 | grep `vault.db.base` = 0 | match |
| T-6 | `tene push` (mock API) → sync_state.json 5 필드 모두 채워짐 | match |
| T-7 | `tene push` 후 `tene pull` (mock API) → last_pulled_at 갱신 + 나머지 4 필드 유지 | match |
| T-8 | `go test ./internal/sync/...` 0 회귀 (merge_test.go 삭제 후) | match |
| T-9 | bodyclose linter — engine.go 4 http.Do 사이트 (s3 에서 활성화) | clean |

---

## 10. Implementation Notes (Pre-Design Hints)

- `os.ReadFile + json.Unmarshal → modify → json.Marshal → os.WriteFile` 패턴 (atomic write 는 v2.1; 현재는 truncate-write 유지하되 5-field 보존)
- `SyncState` JSON tag: `omitempty` 모든 필드 (partial update 지원)
- `state.go` 의 함수 시그니처 변경 가능성 — `saveSyncState(path string, update func(*SyncState))` callback 패턴도 검토 (Design 단계 결정)
- math/rand 주석 위치: `engine.go:407` 의 `time.Duration(rand.Intn(jitter)) * time.Millisecond` 위 한 줄
- 백업 정책 변경 (`vault.db.base` 삭제) 는 user-visible 가시 변화 — release note 에 명시

---

## 11. Blocking Relationships

| Blocks | Reason |
|--------|--------|
| **Sprint 2 `test-infra`** (T2-B5 sync engine test) | 실제 사용 path 만 테스트해야 함 — dead code 제거 후 mock Transport / mock StateStore 작성 의미 |
| **Sprint 3 `lint-hardening`** (T3-B2 `//go:build cloud`) | dead code 244 LOC 제거가 unused linter 활성화 prerequisite (그 다음 cloud 명령 격리) |
| **Sprint 2 `sync-contracts`** (T2-A5 DR-2 blueprint) | `Engine` 의 DI 설계 시 — dead merge 코드 없는 깨끗한 surface 위에서 인터페이스 추출 |

| Blocked by | Reason |
|-----------|--------|
| (none) | unblocked from start |

---

## 12. Next Steps

1. [x] Plan 문서 작성 완료
2. [ ] Design 문서 작성 — [`docs/02-design/features/sync-cleanup.design.md`](../../02-design/features/sync-cleanup.design.md) — code-analyzer 위임 (dead code 검증 + state schema 설계)
3. [ ] L2 boundary: design 후 사용자 승인 → do phase
4. [ ] Sprint 1 PR #9 (dead code 제거) → PR #10 (saveSyncState 통합)
5. [ ] s2 의 sync-contracts 가 DR-2 blueprint (Engine DI) 도입 — 본 plan 의 깨끗한 surface 위에서

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-05-13 | Initial draft (Sprint 1 plan phase, L2 boundary) | cto-lead |
