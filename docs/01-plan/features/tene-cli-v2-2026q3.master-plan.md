---
template: sprint-master-plan
version: 1.0
description: tene CLI v2.0 13-Week Sprint Master Plan (6 sprints × ~7,897 LOC)
variables:
  - feature: tene-cli-v2-2026q3
  - displayName: "tene CLI v2.0 13-Week Plan"
  - date: 2026-05-13
  - author: sprint-master-planner
  - trustLevel: L2
  - duration: 13 weeks (2026-05-13 → 2026-08-12)
---

# tene CLI v2.0 13-Week Plan — Sprint Master Plan

> **Sprint ID**: `tene-cli-v2-2026q3`
> **Date**: 2026-05-13
> **Author**: sprint-master-planner (bkit Sprint 4 표준)
> **Trust Level (시작)**: L2 Semi-Auto (현 dashboard 기준)
> **예상 기간**: 13 weeks (W1=2026-05-13 → W13=2026-08-12)
> **Master Plan template**: bkit v2.1.13 (Sprint 4 Presentation 산출)
> **Baseline 문서**: `docs/01-plan/cli-sprint-master-plan-2026-05-13.md` (1,194 LOC v2 분석)
> **출처 감사**: `docs/03-report/cli-completeness-audit-2026-05-11.md` (1,383 LOC, 10-agent 병렬)

---

## 0. Executive Summary

| 항목 | 내용 |
|------|------|
| **Mission** | tene CLI v1.0.8 의 16 P0 부채 + 6 보정사항 해소 + 생체인증·공급망 보안 도입으로 **v2.0.0 stable** 출시 (13주, 6 sprint, ~7,897 신규 LOC, 58 PR) |
| **Anti-Mission** | DAU/유료 KPI (free OSS), marketing site (`apps/web/` 별도 PDCA), tene-cloud 서버 인프라, MCP server (`tene serve --mcp` 는 Q3 후반 v2.1 stretch) |
| **Core Primitives** | 19 sub-features × 6 sprints × ~7,897 LOC × bkit 8-phase per sprint × 4 auto-pause triggers × M1-M10 + S1-S4 quality gates |
| **Trust Level** | **L2 Semi-Auto** — plan→design 자동, do 부터 사용자 승인. `stopAfter=archived` (sprint master plan 은 L2 standard) |
| **Auto-pause 조건** | 4 triggers 활성 (QUALITY_GATE_FAIL / ITERATION_EXHAUSTED / BUDGET_EXCEEDED / PHASE_TIMEOUT) — §6 참조 |
| **Success Criteria** | matchRate 100% × criticalIssue 0 × dataFlowIntegrity 100% × featureCompletion 100% × Show HN front page 진입 (§5 참조) |

### 0.1 핵심 차별화 — 1차 plan(2026-05-12) 대비 6 보정

본 master plan 은 1차 plan 의 6 claim 을 코드 정독 + grep 으로 재검증한 결과를 반영:

1. **DeriveSubKey 16 호출 위치** — vault.go(가정) → 실제는 CLI 11 + encfile 2 + recovery 2 + sync 1 분산 (§2 보정 1)
2. **AAD "부재" 잘못** — parameter 자체는 존재 (`Encrypt(key, plaintext, aad)`); 진짜 결함은 context 빈약 (`[]byte(name)` 만 → cross-env replay 가능) (§2 보정 2)
3. **RecoverySalt 8-byte truncation 잘못** — 실제는 16 byte fixed-string ("tene-recovery-salt"); 결함은 per-vault 고유성 부재 → P0 → P2 강등 (§2 보정 3)
4. **saveSyncState race ≠ 단순 중복** — same file 에 different schema 로 truncate-write → **데이터 손실 결함** (§2 보정 4)
5. **신규 P0-P1 — passwd 검증 부재** — `tene passwd` 가 old password 검증 없이 master 회전 가능 (감사 A1 P0-1 plan 누락) (§2 보정 5)
6. **신규 P0-A1 — `tene audit` reader 부재** — audit_log 기록만 / read CLI 없음 ("audit story is theater") (§2 보정 6)

총 **16 P0** (1차 plan 15 → +2 신규 P0, -1 강등) 모두 §2.2 Features 표 + §3 Phase Roadmap + §8 Risk Register 에 반영.

---

## 1. Context Anchor (Plan → Design → Do 전파)

| Key | Value |
|-----|-------|
| **WHY** | tene v1.0.8 은 "암호 기반 견고 + 구조적 부채 + 공급망 신뢰 격차" 의 세 측면을 동시에 갖고 있다. AEAD 배선은 교과서적이지만 `tene passwd` 가 old password 를 검증 안 함 (P0-P1), sync engine 의 244 LOC dead code + `saveSyncState` truncate-write (P0-S1/S2), CI ubuntu-only (P0-CI1), 코드 서명/SBOM/SLSA 전무, biometric 없음 — 11종 부채로 **현재 상태로는 1k stars / `brew install tene` 가능 상태로 출시할 수 없음**. |
| **WHO** | **AI-vibe coder (50%)** — Cursor/Claude Code 사용자, `tene audit` + biometric "Touch ID once per session" 가치<br>**Indie OSS dev (25%)** — Homebrew bottle < 5s + cosign signed + MIT 가치<br>**Sec-conscious team (15%)** — KAT 36 + Fuzz 8 + AAD 4-tuple + SE/TPM bound + SLSA L3 가치<br>**Solo founder (10%)** — multi-env + recovery key + 빠른 import 가치 |
| **WHAT (도메인)** | Crypto (XChaCha20/Argon2id/HKDF/X25519/BIP39), Vault SQLite schema_migrations, Sync engine cleanup + contracts, Test infrastructure (Fuzz/KAT/testhelper), CI matrix (3 OS × 2 Go), Distribution (brew tap + SLSA + cosign + SBOM + bottle), Biometric (macOS Touch ID + Windows Hello + Linux fprintd/TPM2), Documentation v1→v2 migration |
| **WHAT NOT** | tene-cloud 서버 변경, marketing site 재작성, MCP server 정식 구현 (v2.1 stretch), Apple universal binary fusion, dashboard.tene.sh 신기능, 한/일/중 i18n |
| **RISK** | (a) AAD v2 enrichment 가 기존 v1 vault.db 호환 깨뜨림 → 자동 migration 003 으로 해소<br>(b) 16 DeriveSubKey 사이트 누락 → nil salt 잔존<br>(c) Schema migration 실패 시 vault.db corruption<br>(d) Exit code 2→8 BREAKING 사용자 스크립트 깨짐<br>(e) Biometric 9 OS×하드웨어 매트릭스 회귀<br>(f) Apple Developer ID / Windows EV cert 비용 |
| **SUCCESS** | v2.0.0 stable tag 발행 + Show HN front page 진입 (100+ points) + GitHub Stars +200/1w / +1000/90d + Homebrew weekly installs 1,500 + 0 사용자 보고 vault corruption / cosign verification 실패 |
| **SCOPE (정량)** | 19 sub-features × 6 sprints × **13 weeks** × **~7,897 신규 LOC** (그중 ~30% test, ~20% yaml/docs) × **58 PR** × 평균 4 dev peak / sprint × ~1.24M token budget (PDCA 통합 추정) |
| **OUT-OF-SCOPE** | `tene serve --mcp` (v2.1 W14+), `bkit:tene-audit` PDCA hook (v2.1+), per-secret ACL (`--only KEY1,KEY2`), Apple universal binary fusion, headless machine-bound encryption (P2-Sec1), property-based merge testing (sync merge UX 가 v1.2 부활 후) |

---

## 2. Features (Sprint 구성 작업 묶음)

> **§2 Features 표는 v2 plan §3.1 의 16 P0 + 보정 §2 의 신규 P0 2건 모두 포함.** Sprint별 분배는 §3 Sprint Phase Roadmap.

### 2.1 19 Sub-Features 전체 표

| # | Feature ID | Sprint | 우선순위 | 1줄 요약 | 상태 |
|--:|------------|:------:|:--------:|---------|------|
| 1 | `passwd-verify` | s1 | **P0** | `tene passwd` auth_hash verification (감사 A1 P0-1, 본 plan 보정 §2.5) | pending |
| 2 | `crypto-v2-keys` | s1 | **P0** | `DeriveSubKeyV2(masterKey, salt, info)` + per-vault salt + 16 호출 사이트 migration (보정 §2.1) | pending |
| 3 | `crypto-v2-aad` | s1 | **P0** | `EncryptV2`/`DecryptV2` + AAD 4-tuple `{vault_id, env, key_name, version}` (보정 §2.2) | pending |
| 4 | `sync-cleanup` | s1 | **P0** | merge.go + queue.go 244 LOC 삭제 + saveSyncState 통합 (5-field schema, truncate-write 제거) | pending |
| 5 | `audit-reader` | s1 | **P0** | `tene audit --since/--actor/--limit/--json` 신설 + `audit_log.actor` 컬럼 (보정 §2.6) | pending |
| 6 | `vault-v2-migration` | s2 | **P0** | `schema_migrations` 테이블 + 001_v2_envelope + 002_audit_log_v2 + 003_secrets_v2_aad + PRAGMAs 5건 | pending |
| 7 | `sync-contracts` | s2 | P1 | DR-2 blueprint: `contracts.go` (MetadataProvider/VaultReader/Transport/StateStore/VaultIO 5 인터페이스) + DI | pending |
| 8 | `test-infra` | s2 | **P0** | testhelper 리팩토 + Fuzz 8 target + KAT 36건 (RFC 8439/9106/5869/BIP39) + sync engine test | pending |
| 9 | `ci-matrix` | s3 | **P0** | 6-cell matrix (3 OS × 2 Go) + govulncheck + coverage gate (≥70%) + 11 action SHA pin + dependabot | pending |
| 10 | `lint-hardening` | s3 | P1 | 15 linter 활성화 + `//go:build cloud` 빌드 태그 + `unused` 재활성 + 22 `fmt.Errorf` → `%w` | pending |
| 11 | `brew-tap-reactivation` | s3 | P1 | Homebrew tap 활성 + Windows `.zip` fix + CRLF 처리 + ACL + auto-tag.yml idempotency | pending |
| 12 | `biometric-auth` | s4 | **P0** | macOS Touch ID + Windows Hello + Linux fprintd/TPM2 + 9 OS×하드웨어 매트릭스 + decorator pattern | pending |
| 13 | `teneerr-rename-claudemd-v2` | s4 | P1 | `pkg/errors` → `pkg/teneerr` finalize + Unwrap + claudemd v2 sentinel + LLM 모순 제거 (Quick Ref) | pending |
| 14 | `config-slim-encfile-hardening` | s4 | P1 | config dead code -140 + encfile KDF params honor + AAD literal const + var → const + Windows USERPROFILE | pending |
| 15 | `slsa-cosign-sbom` | s5 | **P0** | SLSA L3 provenance (slsa-github-generator@v2.0.0) + cosign keyless + SPDX/CycloneDX SBOM | pending |
| 16 | `brew-bottle` | s5 | P1 | darwin-arm64/x86_64/linux-amd64 bottle + S3+CloudFront CDN + Formula auto-PR | pending |
| 17 | `codesign-notarization` | s5 | P1 | macOS notarization (Apple Developer ID + `gon`) + (선택) Windows Authenticode 비용 결정 | pending |
| 18 | `documentation-migration` | s6 | **P0** | v1→v2 migration guide + threat model + CLI reference (Cobra `doc.GenMarkdownTree`) + CHANGELOG | pending |
| 19 | `launch-campaign` | s6 | **P0** | Show HN (화요일 9am ET) + Daily.dev squad + GeekNews + Reddit (4 sub) + Awesome PRs + brew Formula 머지 | pending |

**합계**: 10 P0 + 9 P1, 13 weeks, ~7,897 LOC, 58 PR

### 2.2 16 P0 ↔ Feature 매핑 (감사 출처 cross-reference)

| P0 ID | Feature | file:line | 감사 출처 |
|-------|---------|-----------|----------|
| P0-S1 | `sync-cleanup` | `internal/sync/engine.go:174-242` + merge.go(160) + queue.go(84) | DR-2 |
| P0-S2 | `sync-cleanup` | `internal/cli/push.go:163` ↔ `internal/sync/engine.go:445` | DR-2 + 보정 §2.4 |
| P0-V1 | `audit-reader` + `vault-v2-migration` | `internal/vault/schema.go:29-35` + `vault.go:424-434` | DR-2 |
| P0-C1 | `crypto-v2-keys` | (보정) cli 11 + encfile 2 + recovery 2 + sync 1 = 16 사이트 | DR-3 + 보정 §2.1 |
| P0-C2 | `crypto-v2-aad` | (보정) `pkg/crypto/encrypt.go:32` + 모든 caller (AAD = `[]byte(name)` 빈약) | 보정 §2.2 |
| P0-C3 | `crypto-v2-keys` | `pkg/crypto/kdf.go:14-16` + `encfile.go:74` (KDFAlgRegistry 부재) | DR-3 |
| **P0-P1** (신규) | `passwd-verify` | `internal/cli/passwd.go:30-34` + `root.go:173-198` | A1 P0-1 + 보정 §2.5 |
| P0-E1 | `audit-reader` (subcmd group) | `pkg/errors/codes.go:14,18,73,94` ↔ `docs/cli-reference.md:23-32` | DR-4 + A5 P0 |
| P0-Enc1 | `config-slim-encfile-hardening` | `internal/encfile/encfile.go:163` (header KDF params decode-only) | A1 P0-2 |
| P0-G1 | `audit-reader` | `internal/cli/get.go:93-99` + `get_guard_test.go:96` (U-1 JSON warning) | DR-5 |
| P0-T1 | `test-infra` | `internal/sync/engine.go` 472 LOC × 0 직접 test | DR-5 + A6 P0 |
| P0-T2 | `test-infra` | `pkg/crypto/*_test.go` 5 파일 (KAT 0건) | DR-5 + A6 P1 |
| P0-T3 | `test-infra` | 전체 codebase (Fuzz 0건, testing.F 0건) | DR-5 + A6 P0 |
| P0-CI1 | `ci-matrix` | `.github/workflows/ci.yml:11,22` (ubuntu only) | DR-5 + A8 P0 |
| P0-K1 | `audit-reader` (sub-task T1-C6) | `internal/keychain/keychain.go:91-97` (Set+Delete probe 5-15ms) | A7 P0 |
| **P0-A1** (신규) | `audit-reader` | `internal/vault/vault.go:424` 기록만 / read CLI 없음 | A10 P0-5 + 보정 §2.6 |

**보강**: 보정 §2.7 (`.golangci.yml unused 비활성` + 1,481 LOC 잠재 dead code) 는 `lint-hardening` (s3 P1) 에서 `//go:build cloud` 빌드 태그로 해소.

---

## 3. Sprint Phase Roadmap

### 3.1 6 Sprint 구성 + 기간 + 의존성

| Sprint ID | Name | 기간 (W) | 일자 | Features (count) | 의존 Sprint |
|-----------|------|:--------:|------|:----------------:|:-----------:|
| `tene-cli-v2-s1` | Crypto + Sync + passwd Hotfix | W1-W2 | 2026-05-13 → 2026-05-26 | 5 (passwd-verify · crypto-v2-keys · crypto-v2-aad · sync-cleanup · audit-reader) | (root) |
| `tene-cli-v2-s2` | Architecture & Tests | W3-W4 | 2026-05-27 → 2026-06-09 | 3 (vault-v2-migration · sync-contracts · test-infra) | s1 |
| `tene-cli-v2-s3` | CI & Distribution Prep | W5-W6 | 2026-06-10 → 2026-06-23 | 3 (ci-matrix · lint-hardening · brew-tap-reactivation) | s2 |
| `tene-cli-v2-s4` | Biometric Auth & Polish | W7-W9 (3주) | 2026-06-24 → 2026-07-14 | 3 (biometric-auth · teneerr-rename-claudemd-v2 · config-slim-encfile-hardening) | s3 |
| `tene-cli-v2-s5` | Supply Chain Security | W10-W11 | 2026-07-15 → 2026-07-28 | 3 (slsa-cosign-sbom · brew-bottle · codesign-notarization) | s3 + s4 |
| `tene-cli-v2-s6` | v2.0 Launch | W12-W13 | 2026-07-29 → 2026-08-12 | 2 (documentation-migration · launch-campaign) | s5 |

### 3.2 Sprint × Phase 매트릭스 (bkit 8 phase)

각 sprint 는 bkit Sprint v2.1.13 표준 8 phase 를 순차 진행: `prd → plan → design → do → iterate → qa → report → archived`.

| Phase | 활성 시점 | 산출물 | Quality Gates 적용 |
|-------|---------|------|-------------------|
| **prd** | sprint 시작 | PRD 문서 (`docs/01-plan/features/{sprintId}.prd.md`) | M8 (Match Rate ≥ 90%) |
| **plan** | PRD 승인 후 | Plan 문서 (`{sprintId}.plan.md`) | M8 |
| **design** | Plan 승인 후 | Design 문서 (`docs/02-design/features/{sprintId}.design.md`) + 코드베이스 분석 | M4 (lint), M8 |
| **do** | Design 승인 후 (L2 사용자 승인) | 구현 코드 + PR 매트릭스 | M2 Unit, M3 Coverage, M5 Security, M7 Fuzz (s2+) |
| **iterate** | matchRate < 100% 시 | matchRate 100% 달성 (max 5 iter — `ITERATION_EXHAUSTED` 트리거) | M1 Build (= 100%) |
| **qa** | iterate 후 | 7-Layer S1 검증 + L1-L5 테스트 | M3 (= 0 critical), S1 dataFlowIntegrity (= 100) |
| **report** | qa 후 | sprint-report-writer aggregate (phaseHistory/iterateHistory/kpi) | M10 SupplyChain (s5+), S2, S4 |
| **archived** | report 후 (L4 자동) 또는 사용자 명시 (L0-L3) | terminal state, readonly | — |

### 3.3 Sprint 별 Trust Level 진행 범위

| Sprint | 자동 진행 phase | 사용자 승인 phase |
|--------|----------------|------------------|
| s1-s6 (L2 동일) | prd → plan → design (자동) | do → iterate → qa → report → archived (각각 사용자 승인) |

> **Notes**: dashboard 가 L3 로 승격되면 (조건: 누적 Match Rate ≥ 95% × 0 critical issue × 사용자 만족도 응답) `do` 도 자동. L4 (Full Auto) 는 v2.0 launch 후 별도 검토.

---

## 4. Quality Gates 활성화 매트릭스

### 4.1 Gate 정의 (M1-M10 + S1-S4)

| Gate | 검증 항목 | 자동화 | 통과 기준 |
|:----:|----------|:------:|----------|
| **M1** Build | `go build ./...` 6 cell (3 OS × 2 Go) | CI matrix (s3+) | 6/6 green |
| **M2** Unit Test | `go test -race -count=10 ./...` | CI | 0 회귀, 0 flaky |
| **M3** Coverage | `go test -coverprofile + .testcoverage.yml` | CI postrun | total ≥ 70%, package ≥ 60% (s3+); s1=60% |
| **M4** Lint | `golangci-lint run` (15+ linter) | CI lint job (s3+) | 0 issue (`//nolint:` ≤ 5건) |
| **M5** Security | `govulncheck` + `gosec` + dependabot | CI (s3+) | govulncheck clean, gosec ≤ 5 (whitelisted) |
| **M6** Race | `go test -race -count=10` | CI | 0 data race |
| **M7** Fuzz | 8 target × 30s on PR; nightly cron 5min | CI fuzz-smoke (s2+) | 0 crash, 0 oom |
| **M8** Match Rate | gap-detector (plan ↔ impl) | bkit | ≥ 90% (do phase), 100% (iterate exit) |
| **M9** KAT | RFC 8439 + 9106 + 5869 + BIP39 = 36 벡터 | CI crypto test | 36/36 hex 일치 |
| **M10** Supply Chain | SLSA verifier + cosign verify-blob | release workflow (s5+) | 모든 asset 서명 검증 통과 |
| **S1** dataFlowIntegrity (7-Layer) | sprint-qa-flow agent | qa phase | = 100 |
| **S2** featureMap completion | report phase aggregate | — | = 100% |
| **S3** criticalIssueCount | code-analyzer | iterate exit | = 0 |
| **S4** cycle time | M10 | report phase | sprint duration ≤ 14d (s4 = 21d 예외) |

### 4.2 Sprint × Gate 적용 매트릭스

| Sprint | M1 | M2 | M3 | M4 | M5 | M6 | M7 | M8 | M9 | M10 | S1 | S2 | S3 | S4 |
|:------:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:--:|:---:|:--:|:--:|:--:|:--:|
| s1 | ubuntu | ✓ | ≥60% | ✓ 6 lint | ✓ | ✓ | — | ✓ | ✓ (12 KAT) | — | ✓ | ✓ | ✓ | ✓ |
| s2 | ubuntu | ✓ | ≥70% | ✓ | ✓ | ✓ | ✓ 8 target | ✓ | ✓ (36 KAT) | — | ✓ | ✓ | ✓ | ✓ |
| s3 | **6 cell** | ✓ | ≥70% | ✓ 15 lint | ✓ govulncheck | ✓ | ✓ | ✓ | ✓ | — | ✓ | ✓ | ✓ | ✓ |
| s4 | 6 cell | ✓ | ≥70% | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | — | ✓ + 9-매트릭스 | ✓ | ✓ | ✓ (21d) |
| s5 | 6 cell | ✓ | ≥70% | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | **✓ SLSA L3** | ✓ | ✓ | ✓ | ✓ |
| s6 | 6 cell | ✓ | ≥70% | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ | ✓ |

---

## 5. Success Metrics (5건 + DoD)

### 5.1 bkit 표준 5 metrics (sprint 마감 시 측정)

| # | Metric | Target | 측정 방법 |
|---|--------|--------|----------|
| 1 | matchRate (Design ↔ Code) | **100%** | gap-detector (`/pdca check`) |
| 2 | criticalIssueCount | **0** | code-analyzer (`/pdca check`) |
| 3 | dataFlowIntegrity (7-Layer S1) | **100%** | sprint-qa-flow agent (qa phase) |
| 4 | featureCompletion | **100%** | featureMap aggregate (report phase) |
| 5 | sprint cycle time | s1-s3,s5,s6 ≤ 14d / s4 ≤ 21d | M10 |

### 5.2 v2.0 Definition of Done (출시 시점 누적)

| 영역 | DoD (출시 시점) |
|------|-----------------|
| **Security** | passwd verify 동작 + nil salt 0 (grep) + AAD 4-tuple 모든 envelope + KAT 36 통과 + Fuzz 8 × 60s 0 crash + govulncheck clean + cosign signed |
| **Architecture** | schema_migrations 자동 + v1→v2 자동 re-encrypt + teneerr rename 완료 + sync engine 70% coverage + `unused` linter 재활성 |
| **Testing** | 70% total coverage + 60% per-package + CI matrix 6 cell green + race -count=10 통과 + `t.Parallel()` ≥ 50개 |
| **Distribution** | SLSA L3 provenance + cosign 서명 모든 asset + Homebrew bottle (<5초 install) + SBOM (SPDX+CycloneDX) + reproducible build |
| **Biometric** | macOS Touch ID + Windows Hello + Linux fprintd 모두 동작 + 9 OS×하드웨어 매트릭스 통과 + master password fallback invariant (100회 fuzz 통과) |
| **AI Integration** | `tene audit` reader 동작 + `schemaVersion` 모든 JSON + `--strict` mode + claudemd v2 sentinel + Quick Reference 모순 해소 |
| **CLI UX** | exit code drift 0 + `tene init` next-step 3-line + `Example:` 모든 명령 |
| **Documentation** | v1→v2 migration guide + CLI reference 자동 생성 + threat model + biometric guide + CHANGELOG v2.0 |
| **Launch** | Show HN 100+ points + Homebrew Formula 머지 + v2.0.0 stable 태그 + GitHub Verified badge |

### 5.3 출시 후 KPI (감사 §0.3 페르소나별)

| KPI | 30d | 90d |
|-----|----:|----:|
| GitHub Stars | +200 | 1,000+ |
| Weekly brew installs | 100+ | 1,500 |
| HN agent-kay karma | 50+ | (Show HN 1회 front page) |
| Daily.dev reputation | (성장 중) | 500+ (RSS Source 등록 threshold) |
| Dev.to 누적 조회수 | — | 50K+ |
| User-reported vault corruption | 0 | 0 |
| Cosign verification 실패 | 0 | 0 |
| Issues 첫 응답 시간 | < 24h | < 24h |

---

## 6. Auto-Pause Triggers (4 활성)

| Trigger | 조건 | 사용자 결정 옵션 | sprint master plan 적용 |
|---------|------|----------------|------------------------|
| **QUALITY_GATE_FAIL** | M1-M10 중 1개 게이트 통과 못 함 OR S1 < 100 OR S3 > 0 | (a) fix & resume / (b) forward fix (다음 sprint 으로 carry) / (c) abort | s1 의 KAT 12 통과 후에만 s2 진입; s5 의 cosign verify 실패 시 release 차단 |
| **ITERATION_EXHAUSTED** | iterate phase 가 ≥ 5회 반복 후에도 matchRate < 90% | (a) forward fix / (b) carry (Backlog 으로 이동) / (c) abort & 회고 | s4 biometric 9 OS 매트릭스 회귀 시 가능성 가장 큼 → cto-lead escalate |
| **BUDGET_EXCEEDED** | cumulativeTokens > sprint budget × 80% | (a) budget 증액 & resume / (b) abort / (c) archive partial | 부록 D 의 sprint 별 토큰 예산 (s4 ~320k 최대) 초과 시 |
| **PHASE_TIMEOUT** | 단일 phase 진행 시간 > sprint 길이 × 50% | (a) timeout 연장 / (b) force-advance / (c) abort | s4 do phase 21d × 50% = 10.5d 한도; sprint-orchestrator 가 sub-task 분할 |

---

## 7. Cross-Sprint Dependency

### 7.1 Critical Path (11주, s1→s2→s3→s5→s6)

```
W1-W2  [s1: Crypto + Sync + passwd Hotfix]
         │ AAD enrichment 후 모든 ciphertext 새 format
         ▼
W3-W4  [s2: Vault v2 + Schema Migration + Test Infra]
         │ schema 안정 후 matrix build 의미 있음
         ▼
W5-W6  [s3: CI Matrix + Distribution Prep]
         │
   ┌─────┴─────┐
   ▼           ▼
W7-W9  [s4: Biometric & Polish (병렬 3 트랙)]
W10-W11 ← s4 와 부분 병렬 가능 (s5 시작 가능 시점은 s3 brew tap 완료 후)
         │
W10-W11 [s5: Supply Chain Security]
         │
W12-W13 [s6: v2.0 Launch]
```

### 7.2 차단 관계 매트릭스 (v2 plan §4.2 채택)

| Blocker (먼저 끝) | Blocked (시작 불가) | 이유 |
|------------------|---------------------|------|
| `crypto-v2-aad` (P0-C2 AAD enrichment) | `vault-v2-migration` (P1-V1 schema migration) | AAD 변경된 ciphertext 가 schema migration 의 vault_secrets 컬럼 변환 대상 |
| `crypto-v2-keys` (P0-C1 nil salt 통일) | `test-infra` (P0-T2 KAT 작성) | salt = vault_meta.kdf_salt 정착 후 RFC 8439 벡터 확정 |
| `sync-cleanup` (P0-S1 dead code 제거) | `test-infra` (sync engine test) | 실제 사용 path 만 테스트해야 함 |
| `passwd-verify` (P0-P1) | `vault-v2-migration` (P1-V1) | `auth_hash` 컬럼 추가가 schema v2 의 일부 |
| `vault-v2-migration` (P1-V1) | `ci-matrix` (P0-CI1) | matrix 후 3 OS 에서 migration 검증 |
| `test-infra` (P0-T1 sync test) | `biometric-auth` (P1-Bio1) | 인증 추가 전 sync 기본 동작 안정 |
| `teneerr-rename-claudemd-v2` (P1-E1) | `pkg/domain` 이름 충돌 해소 | 동시 rename 이 깔끔 (S4 Trk-B) |
| `audit-reader` (P0-E1 exit code) | 자동화 changelog 안내 | drift 해소 전 안내 시 사용자 혼란 |
| `brew-tap-reactivation` (P1-Dist1) | `brew-bottle` (P1-Dist4 bottle) | tap 활성 후 bottle CDN 의미 있음 |
| 모든 s3 lint 변경 | s4 시작 | gosec/errorlint 통과 후 새 코드 작성 |

### 7.3 병렬 실행 가능 트랙

| Sprint | 트랙 수 | 트랙 구성 | 최대 병렬 dev |
|--------|:-------:|----------|:-------------:|
| s1 | 3 | Trk-A Crypto + passwd / Trk-B Sync / Trk-C UX+Audit | 4 |
| s2 | 2 | Trk-A Migration + Contracts / Trk-B Test infra + teneerr prep | 4 |
| s3 | 4 | Trk-A CI / Trk-B Lint / Trk-C Dist prep / Trk-D Biometric design spike | 4 |
| s4 | 3 | Trk-A Bio impl (15 dev-day) / Trk-B teneerr+claudemd / Trk-C Config+encfile | 4 |
| s5 | 3 | Trk-A SLSA+cosign / Trk-B brew bottle / Trk-C notarization | 3 |
| s6 | 3 | Trk-A docs / Trk-B 캠페인 / Trk-C 모니터링 | 2 |

### 7.4 외부 의존성 모니터링 (13주)

13주 sprint 기간 중 추적할 11 패키지:

| 패키지 | 현재 | 점검일 | 트리거 | 영향 |
|--------|------|--------|--------|:----:|
| github.com/spf13/cobra | v1.10.1 | 2026-06-15 | semver minor up | L |
| github.com/zalando/go-keyring | v0.2.6 | 2026-05-30 | **6개월 활동 없으면 fork 검토** (s4 전) | M |
| github.com/keybase/go-keychain | (s3 신규) | 2026-06-30 | macOS Touch ID 의존 | H |
| modernc.org/sqlite | v1.39.0 | 2026-07-15 | perf 회귀 시 대안 검토 | M |
| golang.org/x/crypto | v0.43.0 | 2026-06-30 | go.mod 자동 + KAT 재검증 | M |
| github.com/tyler-smith/go-bip39 | v1.1.0 | 안정 | — | L |
| google/go-tpm-tools | (s4 신규) | Linux TPM | Linux fprintd 대안 | M |
| saltosystems/winrt-go | (s4 신규) | Windows Hello CGo 대안 | H |
| sigstore/cosign | (s5 신규) | 매 release | OIDC token 갱신 | H |
| slsa-framework/slsa-github-generator | v2.0.0 (s5) | 매 release | SLSA L3 generator | H |

---

## 8. Risks & Mitigation

### 8.1 Sprint 별 Risk Register

#### Sprint 1 (Crypto + Sync + passwd Hotfix)

| 리스크 | 가능성 | 영향 | 완화 |
|--------|:------:|:----:|------|
| 16 DeriveSubKey 사이트 일괄 변경 시 누락 → 일부 keyset nil salt 잔존 (보정 §2.1) | M | H | T1-A4 후 `grep -rEn "DeriveSubKey\\((rootKey\|masterKey)," --include='*.go'` 분석 + integration test 로 모든 path 한 번씩 실행 |
| AAD v2 enrichment 가 기존 v1 vault.db 호환 깨뜨림 (보정 §2.2) | H | H | EncryptV2/DecryptV2 신규 함수 분리; v1 vault 는 기존 Encrypt 유지; **마이그레이션 게이트는 s2** (003_secrets_v2_aad 자동 re-encrypt) |
| Sync dead code 제거가 향후 merge UX 복원 시 비용 ↑ | L | M | 삭제 PR 에 "merge UX 는 v1.2 에서 새 architecture (CRDT or OT) 로 재도입" 명시 + s4 design spike 예약 |
| Exit code 변경 (2→8) BREAKING 사용자 스크립트 깨짐 | M | M | CHANGELOG BREAKING 헤드라인 + `docs/migration/exit-codes.md` + Show HN 발사 24h 전 사전 공지 (Daily.dev squad) |
| passwd verify 도입 시 keychain user UX 변경 — 매번 password 입력 요구? | L | M | passwd/recover 명령만 keychain bypass; `tene get/run/list` 등은 기존 keychain 캐시 유지 |
| `tene audit` 추가 후 SQL injection 위험 (보정 §2.6) | M | H | `--actor` 등 user input 은 `?` placeholder; integration test 에 `'; DROP TABLE` 시도 |

#### Sprint 2 (Vault v2 + Test Infrastructure)

| 리스크 | 완화 |
|-------|------|
| `003_secrets_v2_aad.sql` re-encrypt 가 interactive (master password 필요) — CI 자동 검증 불가 | `TENE_MASTER_PASSWORD` env var 로 비대화 path; 별도 integration test |
| Schema migration 실패 시 vault.db corruption | `vault.db.pre-v2-{timestamp}` 자동 백업 + rollback 명령 (T2-A8) |
| Fuzz 가 기존 코드의 panic 발견 → s2 마감 지연 | fuzz panic 은 s1 hotfix 후속으로 분리 (v1.0.10) — sprint phase 격리 |
| testhelper 리팩토가 22개 테스트 깨뜨림 | 점진적 (subset 단위) 리팩토 — 각 sub-PR 마다 테스트 통과 |

#### Sprint 3 (CI & Distribution Prep)

| 리스크 | 완화 |
|-------|------|
| Windows CI 에서 fork-bomb 또는 PATH 이슈 | windows-specific test exclusion (`//go:build !windows`) — 보수적 시작 |
| gosec/govulncheck 활성화 시 다수 신규 issue | 첫 PR 은 whitelist + 후속 ticket — sprint 마감 보호 |
| Homebrew bottle 빌드 권한 (GitHub Actions 에서 brew install) | macos-latest runner 사용; brew 사전 캐싱 (`actions/cache`) |
| `homebrew-tene` repo 생성 권한 (사용자가 직접 실행 필요) | 사용자 가이드: `gh repo create tomo-kay/homebrew-tene --public --license MIT --add-readme` |
| Touch ID PoC CGo 가 Xcode 15 의존성 충돌 | macos-13 + macos-14 CI 매트릭스에서 검증; Xcode 14.x 도 빌드 |
| Lint 강화 후 기존 PR 22개 bare fmt.Errorf 일괄 변경 시 race | 단일 PR 로 묶지 말고 5-10 site 단위 분할 |

#### Sprint 4 (Biometric Auth & Polish) — 3주 sprint, 최고 위험

| 리스크 | 완화 |
|-------|------|
| 9 OS×하드웨어 매트릭스 회귀 — `IsNonInteractive()` 감지 누락 시 CI에서 영원히 행 | T4-A9: SSH_CONNECTION/CI/GITHUB_ACTIONS/TENE_BIOMETRIC=skip 모두 자동 감지; CI matrix 에 비-interactive 테스트 우선 |
| macOS Touch ID PoC CGo 가 Xcode 15.x 의존성 폭발 | s3 T3-D3 spike 에서 미리 검증; macos-13 + macos-14 매트릭스 |
| Windows Hello WinRT binding 선택 (winrt-go vs cgo shim) 미결정 | s3 T3-D4 spike 에서 결정 문서화 (W6 end) |
| Linux TPM2 없는 환경에서 sealed object 실패 | fail-soft → master password (T4-A4); fprintd D-Bus 로 기본 |
| master password fallback invariant 100회 fuzz 회귀 | T4-A6 decorator chain (env → biometric → keychain → password → recovery) — invariant test |
| teneerr rename 충돌 (s2 시작 + s4 finalize) | s2 T2-C1 에서 1-shot edit; s4 T4-B1 은 grep 검증만 |

#### Sprint 5 (Supply Chain Security)

| 리스크 | 완화 |
|-------|------|
| Apple Developer ID 비용 ($99/yr) | tomo-kay 개인 계정 결제 (회사 entity 없음) |
| Windows code signing 비용 ($300+/yr EV cert) | EV cert 없이 standard 로 시작; SmartScreen 경고는 `docs/guides/windows-install.md` 안내 |
| SLSA verifier 가 GoReleaser 호환성 이슈 | 공식 example 추종 + dry-run 검증 (s3 T3-C1 사전) |
| Bottle 빌드 시간 (macOS arm64 cross-compile) | matrix 별 native runner 사용; build time < 5min |

#### Sprint 6 (v2.0 Launch)

| 리스크 | 완화 |
|-------|------|
| Show HN flag 위험 (agent-kay <7일 계정) | s1-s5 동안 HN karma 50+ 달성; Show HN 발사 2주 전 dang 에게 사전 안내 메일 |
| Homebrew Formula 머지 지연 (외부 maintainer 리뷰) | 자체 tap (tomo-kay/homebrew-tene) 우선; upstream Formula PR 은 v2.0 launch 후 |
| GitHub Stars 회귀 (Show HN 실패 시) | 4 채널 (HN+Daily.dev+GeekNews+Reddit) 동시 cross-share 로 risk 분산 |

### 8.2 Cross-Sprint Risk (전체 13주)

| 리스크 | 완화 |
|-------|------|
| s1 의 v2 envelope 변경이 s2 의 schema migration 과 race — 동일 PR 에서 양쪽 수정 시 충돌 | s1 의 EncryptV2 는 신규 함수 only; s2 의 003 migration 만 기존 v1 ciphertext 를 EncryptV2 로 변환. 시간 분리 명확 |
| s4 의 21일 sprint 가 s5 시작 지연 → s6 launch window 놓침 | s4 Trk-A 가 W7-W9 전체, Trk-B/C 는 W7-W8 종료. s5 시작은 s3 완료 후 가능하므로 s4 와 부분 병렬 |
| Trust Level L2 가 13주 내 L3 승격 안 됨 → 매 do phase 사용자 승인 누적 80회 | s2 끝나면 cto-lead 가 dashboard L3 승격 검토 (조건: 누적 matchRate ≥ 95% + 0 critical) |
| Token budget 1.24M 초과 (특히 s4 320k) | BUDGET_EXCEEDED 트리거; sprint-orchestrator 가 feature 우선순위 재조정 (P1 항목 backlog 이동) |

---

## 9. Resume / Abort 흐름

### 9.1 Phase 별 Resume / Abort 매트릭스

| 상황 | 절차 | 명령 |
|------|------|------|
| Auto-pause (QUALITY_GATE_FAIL) | 사유 확인 → fix → 게이트 재실행 → resume | `/sprint resume tene-cli-v2-s{N}` |
| Auto-pause (ITERATION_EXHAUSTED) | cto-lead escalate → scope 재협상 → forward fix or carry | `/sprint resume --force-advance` 또는 `/sprint feature carry {featureId}` |
| Auto-pause (BUDGET_EXCEEDED) | budget 증액 결정 → resume / abort | `/sprint resume --budget +50k` |
| Auto-pause (PHASE_TIMEOUT) | timeout 연장 or sub-task 분할 → resume | `/sprint resume --phase-timeout 168h` |
| 사용자 abort | terminal state archived; rollback 가능한 모든 PR 은 v1.0.8 hotfix branch 로 cherry-pick | `/sprint archive tene-cli-v2-s{N} --reason "..."` |
| Sprint 진행 중 v1.0.x emergency fix | 별도 hotfix branch (`hotfix/v1.0.x`); sprint 와 독립 | 표준 git flow |

### 9.2 Sprint Master Plan 의 forward-only 보장

bkit Sprint v2.1.13 표준: sprint 의 terminal state 는 `archived` (rolled-back 아님). 개별 feature 단위 rollback 은 sprint 내부 phase 에서 가능하나, sprint 자체는 forward-only — 즉 sprint master plan 의 6 sprint 중 어느 것도 "되돌릴" 수 없음. 대신 `/sprint fork` 로 시간 분기 (master plan 자체는 immutable, 분기는 새 sprint id 생성).

---

## 10. Sprint 추적 (Living Document)

본 master plan 은 sprint 진행 중 다음 필드가 cumulative 로 갱신됨:

| Section | 갱신 시점 | 갱신 주체 |
|---------|----------|----------|
| §2.1 Features 표의 `상태` 컬럼 (`pending`→`prd`→`plan`→...→`archived`) | 각 sprint phase 전이 시 | sprint-orchestrator |
| §5.3 출시 후 KPI (30d/90d) | s6 launch 24h 후부터 weekly | growth-routine (`/tene-stats`) |
| §7.4 외부 의존성 점검일 | 점검일 도래 시 | external-dependency-watch |
| §8 Risk Register (실현된 리스크는 ✅ 또는 ❌ 마크) | sprint report phase 시 | sprint-report-writer |
| Phase History (별도 sidecar `tene-cli-v2-2026q3.history.jsonl`) | 매 phase 전이 시 | bkit audit-logger |

### 10.1 매 Sprint 종료 시 자동 실행

```
/pdca check        — Match Rate 측정
/pdca qa           — Zero Script QA + L1-L5 테스트
/pdca report       — sprint-report-writer aggregate
/sprint phase next — 다음 sprint phase 자동 진입 (L2 는 사용자 승인)
```

### 10.2 Archived 시 readonly 전환

s6 launch + 1주 (W14) 시점에 `/sprint archive tene-cli-v2-2026q3` 실행 → master plan 파일 readonly chmod + `archivedAt` 필드 state JSON 에 기록 + audit log `sprint_archived` action 기록.

---

## 11. 부록

### A. 정독 통계 (1차 plan 대비 갱신)

| 항목 | 1차 (2026-05-12) | 본 plan (2026-05-13) |
|------|-----------------|----------------------|
| 정독 LOC | "~10,190" 추정 | **10,267 정확 측정** |
| 비-테스트 LOC | "~5,100" | **7,002** (audit 추정 7,504 조정) |
| 테스트 LOC | "~5,090" | **2,493** (audit 추정 2,763 조정) |
| 정독 파일 | "~89" | **86 (62 src + 24 test)** |
| P0 발견 | 15 | **16** (P0-P1 + P0-A1 신규, RecoverySalt P0→P2 강등) |
| Sprint 수 | 6 | 6 (동일) |
| 기간 | 13주 | 13주 (동일) |
| 보정 건 | — | **6** (§2.1-§2.6) |

### B. KAT 36 벡터 출처

| 벡터 출처 | 알고리즘 | 벡터 수 |
|----------|---------|--------:|
| RFC 8439 §A.2 + libsodium test vectors | XChaCha20-Poly1305 (Encrypt/Decrypt) | 6 |
| RFC 9106 §B.1 | Argon2id (KDF) | 3 |
| RFC 5869 §A.1-A.3 | HKDF-SHA256 (DeriveSubKey) | 3 |
| BIP-39 (Trezor reference) | mnemonic ↔ seed | 24 |
| **합계** | | **36** |

각 벡터는 `pkg/crypto/testdata/{rfc}-{name}.json` — `{input, salt, key, expected_hex}` 4-tuple 고정.

### C. 9 OS×하드웨어 시나리오 (s4 biometric)

| # | 시나리오 | 1차 unlock | Fallback | 자동화 |
|:-:|---------|-----------|----------|:------:|
| 1 | macOS Touch ID 노트북 | SE-wrapped + Touch ID | master password | macOS CI |
| 2 | macOS T2 (Touch ID 없음) | Apple Watch / password | master password | 부분 manual |
| 3 | macOS Intel 구형 (T2 없음) | (미지원) | master password | macOS CI |
| 4 | Windows 11 + Hello + 카메라/지문 | CNG TPM + 얼굴 | master password | Hello manual |
| 5 | Windows 11 PIN-only | CNG TPM + PIN | master password | Windows CI |
| 6 | Windows TPM 없거나 비활성 | (미지원) | master password | Windows CI |
| 7 | Linux + TPM2 + fprintd | TPM2 sealed + 지문 | master password | fprintd manual |
| 8 | Linux 데스크탑 TPM 없음 | libsecret only | master password | Linux CI |
| 9 | CI / SSH / Docker / WSL headless | (자동 skip) | TENE_MASTER_PASSWORD | 모든 CI |

### D. Sprint 별 Token Budget (PDCA 통합)

| Sprint | Plan | Design | Do | Check+QA | Report | 합계 |
|--------|-----:|-------:|---:|---------:|-------:|-----:|
| s1 | ~25k | ~30k | ~120k | ~40k | ~15k | **~230k** |
| s2 | ~20k | ~25k | ~150k | ~50k | ~15k | **~260k** |
| s3 | ~15k | ~20k | ~60k | ~30k | ~10k | **~135k** |
| s4 | ~25k | ~35k | ~180k | ~60k | ~20k | **~320k** |
| s5 | ~15k | ~20k | ~80k | ~40k | ~15k | **~170k** |
| s6 | ~20k | ~10k | ~40k | ~20k | ~30k | **~120k** |
| **총합** | **~120k** | **~140k** | **~630k** | **~240k** | **~105k** | **~1.24M** |

Opus 4.7 1M context 사용 가정. Trust Level L2 기준 사용자 승인 약 80회 (13주).

### E. PR 분할 견적

| Sprint | PR 수 | 평균 LOC | 총 LOC |
|--------|:-----:|---------:|-------:|
| s1 | 12 | +146 | +1,757 |
| s2 | 10 | +220 | +2,200 |
| s3 | 8 | +85 | +680 (yaml-heavy) |
| s4 | 14 | +120 | +1,680 |
| s5 | 6 | +130 | +780 (yaml + docs) |
| s6 | 8 | +100 | +800 (docs-heavy) |
| **합계** | **58** | **~140** | **~7,897 LOC** |

비교: 현재 codebase 10,267 LOC → 13주에 약 77% 증가 (그중 ~30% test, ~20% yaml/docs).

### F. Show HN 핵심 메시지 (s6 launch 24h 전 finalize)

```
Title: Show HN: tene v2.0 — local-first secret manager with biometric auth and SLSA L3 signing

Body:
I built tene to stop pasting API keys into Cursor/Claude Code/Cursor chat windows.

v2 adds:
1. macOS Touch ID + Windows Hello + Linux fprintd unlock (SE/TPM-bound, fallback-safe)
2. SLSA L3 provenance + cosign keyless signing on every release
3. tene audit — "which AI session touched which secret, when?" forensics
4. 36 RFC test vectors (8439, 9106, 5869, BIP39) + 8 fuzz targets (no crash in 60s)
5. CI matrix on macOS / Windows / Linux × Go 1.24/1.25

Local-only by default. No server, no telemetry. MIT. brew install tene.

Threat model + crypto self-audit: tene.sh/threat-model
Repo: github.com/tomo-kay/tene
```

(HN 발사 시간: 화요일 9am ET = 수요일 22:00 KST. growth-routine `agent-kay` 계정 사용; karma 50+ 도달 후)

---

> **Status**: Draft v1.0 — pending review.
>
> 본 master plan 은 v2 plan (`cli-sprint-master-plan-2026-05-13.md`, 1,194 LOC) 의 모든 핵심 발견 (16 P0 + 6 보정 + 11 외부 의존성 + 9 OS×하드웨어 매트릭스 + 58 PR 견적) 을 bkit Sprint v2.1.13 master-plan template 의 10절 구조로 재구성한 결과다. 6 sprint × 13 weeks × ~7,897 LOC × 58 PR 로 v2.0.0 stable 까지 가는 critical path 가 명시되어 있고, 4 auto-pause trigger + M1-M10 + S1-S4 quality gates 가 모든 sprint phase 에 자동 적용되며, Trust Level L2 (`stopAfter=archived`) 기준 매 do phase 사용자 승인 필요.
>
> — sprint-master-planner @ 2026-05-13
