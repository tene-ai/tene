# Staging QA Test Plan
## Tene Cloud Local Environment Full Integration Test

> **Version**: v1.0
> **Date**: 2026-04-08
> **Feature**: staging-qa-test
> **Status**: Plan Complete -- Design Phase Ready
> **Upstream**: `docs/00-pm/staging-qa-test.prd.md`
> **Method**: Plan Plus (Brainstorming-Enhanced)

---

## Executive Summary

| Perspective | Description |
|-------------|-------------|
| **Problem** | Tene Cloud 구현(API 29개 엔드포인트 + Dashboard 7개 페이지 + CLI 25+ 명령어)이 완료되었으나, 로컬 환경에서 전체 스택 통합 검증이 수행되지 않음 |
| **Solution** | Chrome 자동화로 Dashboard UX 검증 + tene CLI 빌드로 전체 명령어 테스트 + PostgreSQL/MinIO 데이터 정합성 확인을 5단계 UX Flow 순서로 실행 |
| **Functional UX Effect** | 실제 사용자 여정(Login → Vault 생성 → Push → Dashboard 확인)을 따라가며 모든 레이어(Frontend-API-DB-S3)의 유기적 동작을 검증 |
| **Core Value** | Production 배포 전 전체 기능의 정상 동작을 보장하고, 발견된 버그를 현장에서 즉시 수정하여 배포 리스크 제거 |

---

## Context Anchor

| Dimension | Content |
|-----------|---------|
| **WHY** | Cloud 기능 구현 후 Dashboard-CLI-API-DB-S3 간 통합 검증 부재. Production 배포 전 반드시 확인 필요 |
| **WHO** | tomo-kay GitHub 계정, 로컬 개발 환경 |
| **RISK** | OAuth 콜백 미동작, CORS 에러, DB 스키마 불일치, S3 업로드 실패, JWT 만료 처리, CLI 미구현 명령어 |
| **SUCCESS** | 전체 기능 정상 동작 확인, 발견 버그 즉시 수정, DB/S3 데이터 정합성 100% |
| **SCOPE** | 로컬 환경 (localhost:3001 + localhost:8080 + Docker PostgreSQL/MinIO) |

---

## User Intent Discovery (Plan Plus Phase 1)

| Item | Answer |
|------|--------|
| **Core Purpose** | 전체 E2E 통합 검증 — Dashboard-CLI-API-DB-S3 전체 데이터 흐름을 end-to-end로 검증 |
| **Target** | CLI와 Cloud에서 제공하기로 한 기능 모두. 데이터 흐름과 UX 흐름 고려하여 유기적 테스트 |
| **CLI Scope** | 전체 — 로컬(init/set/get/list/delete/run/import/export) + Cloud(push/pull/sync/team/billing/whoami) |
| **Extra** | DB 직접 쿼리, S3 객체 검증, API 로그 모니터링, Console 에러 모니터링, 디버그 로깅 시스템 보완 |

---

## Alternatives Explored (Plan Plus Phase 2)

| Approach | Description | Selected |
|----------|-------------|:--------:|
| **A. UX Flow 순서대로** | 사용자 여정 기준으로 Login → Dashboard → CLI Local → CLI Cloud → Cross-Validation | **Yes** |
| B. 기능별 병렬 테스트 | Auth, Vault, Team 등 기능 단위로 CLI+Dashboard+DB를 함께 검증 | No |
| C. 레이어별 테스트 | Frontend → API → DB/S3 순서로 레이어 단위 검증 | No |

**선택 근거**: 실제 사용자가 경험하는 순서로 테스트하여 자연스러운 흐름에서 발생하는 이슈를 발견할 수 있음

---

## YAGNI Review (Plan Plus Phase 3)

### Included (v1 필수)
- GitHub OAuth 로그인 + 토큰 관리
- Dashboard 7개 페이지 전체 탐색
- CLI 로컬 명령어 전체 (init/set/get/list/delete/run/import/export/env/passwd)
- CLI Cloud 명령어 전체 (login/push/pull/whoami/team/billing/devices)
- PostgreSQL 직접 쿼리 데이터 정합성 검증
- MinIO S3 객체 저장 검증
- API 로그 실시간 모니터링
- 브라우저 Console 에러 모니터링
- 디버그 로깅 시스템 보완 (필요시)

### Deferred (이번 테스트에서 제외)
- 성능 테스트 (P99 latency 측정)
- 보안 테스트 (penetration test)
- 부하 테스트 (concurrent users)
- CI/CD 파이프라인 검증
- Production 환경 테스트

---

## 1. Test Architecture

### 1.1 Environment

| Component | Address | Status |
|-----------|---------|--------|
| Dashboard | http://localhost:3001 | Running |
| API Server | http://localhost:8080 | Running (health OK) |
| PostgreSQL | localhost:5432 | Docker (healthy) |
| MinIO S3 | localhost:9000 | Docker (healthy) |
| MinIO Console | localhost:9001 | Docker |
| Test Account | tomo-kay | GitHub OAuth |

### 1.2 Tools

| Tool | Purpose |
|------|---------|
| Chrome (claude-in-chrome) | Dashboard UI 테스트 + Console 모니터링 |
| tene CLI (go build) | CLI 명령어 테스트 |
| docker exec psql | PostgreSQL 직접 쿼리 |
| mc (MinIO Client) / aws cli | S3 객체 확인 |
| /tmp/tene-api.log | API 서버 로그 |

---

## 2. Test Phases

### Phase 1: Auth Flow (Chrome + CLI)

**목표**: GitHub OAuth 로그인이 Dashboard와 CLI 모두에서 정상 동작하는지 확인

| Step | Action | Tool | Expected Result | Verify |
|------|--------|------|-----------------|--------|
| 1.1 | Dashboard http://localhost:3001 접속 | Chrome | /login 페이지 렌더링 | 스크린샷 |
| 1.2 | "Continue with GitHub" 클릭 | Chrome | GitHub OAuth 페이지로 리다이렉트 | URL |
| 1.3 | tomo-kay 계정으로 로그인 | Chrome | GitHub 인증 완료 | - |
| 1.4 | OAuth 콜백 처리 | Chrome | /auth/callback → / 리다이렉트 | URL + API 로그 |
| 1.5 | 인증 상태 확인 | Chrome | Overview 페이지 로드, 사용자 정보 표시 | 스크린샷 |
| 1.6 | localStorage 확인 | Chrome JS | tene-auth에 accessToken, user 존재 | JS 실행 |
| 1.7 | DB users 테이블 확인 | psql | tomo-kay 레코드 존재 | SQL |
| 1.8 | API 로그 확인 | tail log | OAuth + token 발급 로그 | 로그 |

### Phase 2: Dashboard Empty States (Chrome)

**목표**: 인증 후 모든 페이지가 정상 렌더링되는지 확인 (데이터 없는 상태)

| Step | Page | Route | Check Items |
|------|------|-------|-------------|
| 2.1 | Overview | / | StatCards(vault 0, secrets 0), Quick Start 안내, 레이아웃 |
| 2.2 | Vaults | /vaults | Empty State 메시지, `tene push` 안내 |
| 2.3 | Devices | /devices | Empty State 또는 현재 디바이스 표시 |
| 2.4 | Team | /team | Empty State, Pro plan 안내 |
| 2.5 | Audit Log | /audit | Empty State, 필터 버튼(All/push/pull/login/delete/create) |
| 2.6 | Billing | /billing | Free/Pro plan 카드, Usage bars (0 사용량) |
| 2.7 | Console 에러 | - | unhandled error 없음 확인 |
| 2.8 | Network 요청 | - | 모든 API 호출 2xx 응답 |

### Phase 3: CLI Local Commands (Terminal)

**목표**: tene CLI 로컬 기능이 모두 정상 동작하는지 확인

| Step | Command | Expected Result |
|------|---------|-----------------|
| 3.0 | `go build -o tene-dev ./cmd/tene` | 빌드 성공 |
| 3.1 | `./tene-dev init test-project` | vault 생성, 마스터 비밀번호 설정, recovery key 표시 |
| 3.2 | `./tene-dev set DB_HOST localhost` | 시크릿 저장 성공 |
| 3.3 | `./tene-dev set DB_PORT 5432` | 시크릿 저장 |
| 3.4 | `./tene-dev set DB_NAME tene` | 시크릿 저장 |
| 3.5 | `./tene-dev set API_KEY sk-test-12345` | 시크릿 저장 |
| 3.6 | `./tene-dev set JWT_SECRET test-jwt-secret-32bytes-long!!` | 시크릿 저장 |
| 3.7 | `./tene-dev get DB_HOST` | "localhost" 출력 |
| 3.8 | `./tene-dev list` | 5개 키 목록 (값 마스킹) |
| 3.9 | `./tene-dev run -- env \| grep DB_` | DB_HOST, DB_PORT, DB_NAME 환경변수 주입 확인 |
| 3.10 | `./tene-dev export` | .env 형식 출력 |
| 3.11 | `./tene-dev export --encrypted` | 암호화된 백업 출력 |
| 3.12 | `./tene-dev delete API_KEY` | 시크릿 삭제 |
| 3.13 | `./tene-dev list` | 4개 키 목록 (API_KEY 없음) |
| 3.14 | `./tene-dev env` | 현재 환경 표시 |

### Phase 4: CLI Cloud + Dashboard Sync

**목표**: CLI Cloud 명령어와 Dashboard 간 데이터 동기화 확인

| Step | Action | Tool | Expected Result | Dashboard Verify |
|------|--------|------|-----------------|-----------------|
| 4.1 | `./tene-dev login` | CLI | GitHub OAuth → 토큰 획득 | - |
| 4.2 | `./tene-dev whoami` | CLI | user_id, plan 표시 | - |
| 4.3 | `./tene-dev push` | CLI | vault 암호화 → S3 업로드 → 성공 | /vaults에 vault 표시 |
| 4.4 | Dashboard Vaults 새로고침 | Chrome | vault 목록에 test-project 표시 | 스크린샷 |
| 4.5 | Vault Detail 클릭 | Chrome | version, secret count, size 표시 | 스크린샷 |
| 4.6 | `./tene-dev pull` | CLI | S3 다운로드 → 복호화 → 성공 | - |
| 4.7 | `./tene-dev billing` | CLI | 구독 상태 표시 (Free) | /billing 페이지 일치 |
| 4.8 | `./tene-dev team create test-team` | CLI | 팀 생성 | /team에 팀 표시 |
| 4.9 | Dashboard Team 확인 | Chrome | test-team 팀 + owner 표시 | 스크린샷 |
| 4.10 | Dashboard Audit Log | Chrome | push, login 등 이벤트 기록 | 스크린샷 |
| 4.11 | DB audit_logs 확인 | psql | 모든 작업에 대한 audit 레코드 | SQL |
| 4.12 | S3 vault.enc 확인 | mc/aws | vault blob 객체 존재 | mc ls |

### Phase 5: Cross-Validation (DB + S3 + Logs)

**목표**: 전체 테스트 결과의 데이터 정합성 최종 확인

| Step | Action | Expected Result |
|------|--------|-----------------|
| 5.1 | `SELECT * FROM users` | tomo-kay 레코드, github_id, plan 확인 |
| 5.2 | `SELECT * FROM vaults` | test-project vault, version, hash, secret_count |
| 5.3 | `SELECT * FROM audit_logs ORDER BY created_at` | 시간순 작업 기록 전체 |
| 5.4 | `SELECT * FROM devices` | 등록된 디바이스 목록 |
| 5.5 | `SELECT * FROM teams` | test-team 레코드 |
| 5.6 | `SELECT * FROM team_members` | owner 멤버십 |
| 5.7 | MinIO 객체 목록 | vault.enc 파일 존재 + 크기 확인 |
| 5.8 | API 로그 요약 | 에러 로그 없음 확인 |
| 5.9 | Audit Log 정합성 | Dashboard audit == DB audit_logs |

---

## 3. Success Criteria

| # | Criteria | Threshold | Priority |
|---|----------|-----------|----------|
| SC-01 | Auth Flow 완료 | Dashboard + CLI 모두 로그인 성공 | Critical |
| SC-02 | Dashboard 전체 페이지 렌더링 | 7개 페이지 에러 없이 로드 | Critical |
| SC-03 | CLI 로컬 명령어 동작 | init/set/get/list/delete/run/export 모두 성공 | Critical |
| SC-04 | CLI Cloud 연동 | push/pull/whoami/billing/team 성공 | Critical |
| SC-05 | CLI-Dashboard 동기화 | CLI push → Dashboard 반영 확인 | Critical |
| SC-06 | DB 데이터 정합성 | 모든 CRUD 작업이 PostgreSQL에 정확히 반영 | High |
| SC-07 | S3 Blob 저장 | push 후 MinIO에 vault.enc 존재 | High |
| SC-08 | API 응답 코드 | 모든 정상 요청 2xx 응답 | High |
| SC-09 | Console 에러 없음 | 브라우저 console에 unhandled error 없음 | Medium |
| SC-10 | Audit Log 정합성 | 모든 작업이 audit_logs에 기록 | Medium |

---

## 4. Risk & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| OAuth 콜백 실패 | Auth 전체 불가 | CORS 설정 확인, CALLBACK_BASE=localhost:8080 확인 |
| CLI 미구현 명령어 | 테스트 불가 | 에러 메시지 기록, 구현 필요 항목 리스트업 |
| DB 마이그레이션 미적용 | 테이블 없음 | docker exec로 마이그레이션 수동 실행 |
| S3 버킷 미생성 | push 실패 | MinIO Console에서 버킷 수동 생성 |
| JWT 만료 | 401 에러 | 토큰 갱신 로직 확인, 필요시 재로그인 |

---

## 5. Brainstorming Log (Plan Plus)

| Phase | Decision | Rationale |
|-------|----------|-----------|
| Phase 1 | 전체 E2E 통합 검증 선택 | Production 배포 전 전체 스택 검증 필수 |
| Phase 1 | CLI 전체 범위 포함 | Cloud + 로컬 모든 기능이 유기적으로 동작해야 함 |
| Phase 2 | UX Flow 순서 선택 | 실제 사용자 여정을 따라가며 자연스러운 흐름에서 이슈 발견 |
| Phase 3 | DB/S3/로그 모니터링 포함 | 데이터 정합성까지 검증해야 진정한 통합 테스트 |
| Phase 3 | 디버그 로깅 보완 포함 | 테스트 과정에서 로깅 부족 발견 시 즉시 보완 |
