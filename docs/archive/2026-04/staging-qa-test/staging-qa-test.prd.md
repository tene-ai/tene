# Staging QA Test PRD (Product Requirements Document)
## Tene Cloud Local Integration Test Suite

> **Version**: v1.0
> **Date**: 2026-04-08
> **Feature**: staging-qa-test (Local Environment Full Integration Test)
> **Status**: PRD Complete -- Plan Phase Ready
> **Upstream**: `docs/00-pm/tene-cloud.prd.md`, `docs/02-design/features/tene-cloud.design.md`

---

## Executive Summary

| Perspective | Description |
|-------------|-------------|
| **Problem** | Tene Cloud 구현이 완료되었으나, Dashboard-CLI-API-DB-S3 간 전체 통합이 로컬 환경에서 검증되지 않음. 실제 데이터 흐름과 UX 흐름을 end-to-end로 확인해야 함 |
| **Solution** | Chrome 브라우저 자동화 + tene CLI 빌드로 6개 대시보드 페이지, 29개 API 엔드포인트, PostgreSQL/MinIO 데이터 저장을 유기적으로 검증 |
| **Functional UX Effect** | GitHub OAuth 로그인 -> Dashboard 전체 페이지 탐색 -> CLI push/pull -> DB/S3 데이터 확인을 단일 세션에서 완료 |
| **Core Value** | Production 배포 전 로컬에서 전체 기능 정상 동작을 보장하여 배포 리스크 제거 |

---

## Context Anchor

| Dimension | Content |
|-----------|---------|
| **WHY** | Cloud 기능 구현 후 Dashboard-CLI-API-DB-S3 간 통합 검증 부재. Production 배포 전 반드시 확인 필요 |
| **WHO** | Tene 개발자 (Steve), tomo-kay GitHub 계정으로 테스트 |
| **RISK** | OAuth 콜백 미동작, CORS 에러, DB 스키마 불일치, S3 업로드 실패, JWT 만료 처리 오류 |
| **SUCCESS** | 6개 Dashboard 페이지 정상 렌더링, Auth 플로우 완료, Vault CRUD + Push/Pull 동작, DB/S3 데이터 정합성 확인 |
| **SCOPE** | 로컬 환경 한정 (localhost:3001 Dashboard + localhost:8080 API + Docker PostgreSQL/MinIO) |

---

## 1. Test Scope: 검증 대상 기능

### 1.1 Authentication Flow (Critical Priority)

| # | Test Case | Expected Result | Verification |
|---|-----------|-----------------|--------------|
| A-01 | Dashboard 로그인 페이지 접근 | /login 페이지 렌더링, "Continue with GitHub" 버튼 표시 | Chrome 스크린샷 |
| A-02 | GitHub OAuth 인증 시작 | GitHub 로그인 페이지로 리다이렉트 | URL 확인 |
| A-03 | OAuth 콜백 처리 | /auth/callback → 토큰 교환 → / 리다이렉트 | API 로그 + Dashboard 상태 |
| A-04 | 인증 상태 유지 | Zustand store에 accessToken, user 정보 저장 | localStorage 확인 |
| A-05 | /auth/me 호출 | user_id, plan 반환 | API 응답 확인 |
| A-06 | JWT 토큰 갱신 | 401 시 자동 refresh → 새 토큰 발급 | 네트워크 요청 확인 |

### 1.2 Dashboard Pages (High Priority)

| # | Page | Route | Key Elements | Verification |
|---|------|-------|--------------|--------------|
| D-01 | Overview | / | StatCards, Quick Start, Recent Activity | 렌더링 + 데이터 바인딩 |
| D-02 | Vaults | /vaults | VaultTable, Empty State, tene push 안내 | API 호출 + 목록 표시 |
| D-03 | Vault Detail | /vaults/[id] | Version, Secret Count, Size, Audit Trail | 개별 vault 데이터 |
| D-04 | Devices | /devices | DeviceCard, Online/Offline, Revoke | 디바이스 목록 |
| D-05 | Team | /team | Members Table, Invite Modal, Roles | 팀 CRUD |
| D-06 | Audit Log | /audit | Action Filter, Table, Pagination | 로그 목록 |
| D-07 | Billing | /billing | Plan Cards, Usage Bars, Upgrade | 구독 상태 |

### 1.3 CLI-API Integration (Critical Priority)

| # | Test Case | Expected Result | Verification |
|---|-----------|-----------------|--------------|
| C-01 | `tene login` | GitHub OAuth → 토큰 획득 → 로컬 저장 | CLI 출력 + API 로그 |
| C-02 | `tene whoami` | user_id, plan, email 표시 | CLI 출력 |
| C-03 | Vault Create | API POST /vaults → DB 레코드 생성 | DB 쿼리 |
| C-04 | `tene push` | Vault 암호화 → S3 업로드 → DB 버전 갱신 | S3 객체 + DB 레코드 |
| C-05 | `tene pull` | Presigned URL → S3 다운로드 → 로컬 복호화 | 로컬 파일 + API 로그 |
| C-06 | Dashboard Vault 반영 | push 후 /vaults 페이지에 vault 표시 | Chrome 확인 |

### 1.4 Data Persistence (High Priority)

| # | Test Case | Expected Result | Verification |
|---|-----------|-----------------|--------------|
| P-01 | PostgreSQL users 테이블 | OAuth 로그인 후 사용자 레코드 생성 | SQL 쿼리 |
| P-02 | PostgreSQL vaults 테이블 | vault create 후 레코드 생성 | SQL 쿼리 |
| P-03 | PostgreSQL audit_logs 테이블 | push/pull 후 audit 레코드 생성 | SQL 쿼리 |
| P-04 | MinIO S3 vault blob | push 후 vault.enc 파일 저장 | mc/aws cli 확인 |
| P-05 | PostgreSQL devices 테이블 | 디바이스 등록 후 레코드 생성 | SQL 쿼리 |

### 1.5 Error Handling & Edge Cases (Medium Priority)

| # | Test Case | Expected Result |
|---|-----------|-----------------|
| E-01 | 미인증 상태 API 접근 | 401 Unauthorized + /login 리다이렉트 |
| E-02 | 존재하지 않는 vault 접근 | 404 Not Found |
| E-03 | CORS preflight | OPTIONS 요청 정상 응답 |
| E-04 | Rate limiting | 429 Too Many Requests (임계치 초과 시) |

---

## 2. Test Environment

| Component | Address | Technology |
|-----------|---------|------------|
| Dashboard | http://localhost:3001 | Next.js 15 |
| API Server | http://localhost:8080 | Go/Echo v4 |
| PostgreSQL | localhost:5432 | Docker (postgres:16-alpine) |
| MinIO S3 | localhost:9000 | Docker (minio/minio) |
| MinIO Console | localhost:9001 | Web UI (minioadmin/minioadmin) |

### Test Account
- GitHub: tomo-kay
- OAuth App: Tene Local (localhost callback)

---

## 3. Test Execution Strategy

### Phase 1: Auth & Dashboard UX (Chrome)
1. Dashboard 로그인 → GitHub OAuth → 인증 완료
2. 전체 6개 페이지 순회 → 렌더링 + 데이터 바인딩 확인
3. Empty State UX 확인

### Phase 2: CLI Integration (Terminal)
1. `tene` 로컬 빌드
2. `tene login` → OAuth → 토큰 획득
3. Vault 생성 → `tene push` → Dashboard 반영 확인

### Phase 3: Data Verification (DB + S3)
1. PostgreSQL 직접 쿼리 → 데이터 정합성
2. MinIO 객체 확인 → vault blob 존재
3. API 로그 확인 → 요청/응답 추적

### Phase 4: Cross-Functional Validation
1. CLI push → Dashboard 새로고침 → vault 표시
2. Dashboard 팀 생성 → CLI 팀 명령어 동기화
3. Audit 로그에 모든 작업 기록 확인

---

## 4. Success Criteria

| Criteria | Threshold |
|----------|-----------|
| Auth Flow 완료 | GitHub OAuth → Dashboard 로그인 100% |
| Dashboard 페이지 렌더링 | 7개 페이지 모두 에러 없이 로드 |
| API 응답 코드 | 모든 정상 요청 2xx 응답 |
| DB 데이터 생성 | users, vaults, audit_logs 레코드 존재 |
| S3 Blob 저장 | push 후 vault.enc 객체 확인 |
| CLI-Dashboard 연동 | CLI 작업 결과가 Dashboard에 실시간 반영 |
| Console 에러 없음 | 브라우저 console에 unhandled error 없음 |
