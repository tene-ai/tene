# Staging QA Test Design Document
## Tene Cloud Full Integration Test — Detailed Test Specification

> **Version**: v1.0
> **Date**: 2026-04-08
> **Feature**: staging-qa-test
> **Status**: Design Complete — Do Phase Ready
> **Architecture**: Option A — Pragmatic E2E (Chrome + CLI + DB/S3 Verification)
> **Upstream**: Plan (`staging-qa-test.plan.md`), PRD (`staging-qa-test.prd.md`)

---

## Context Anchor

| Dimension | Content |
|-----------|---------|
| **WHY** | Cloud 기능 구현 후 Dashboard-CLI-API-DB-S3 간 통합 검증 부재. Production 배포 전 필수 확인 |
| **WHO** | tomo-kay GitHub 계정, 로컬 환경 (localhost) |
| **RISK** | OAuth 콜백 실패, CORS, DB 스키마 미적용, S3 버킷 미생성, JWT 만료, CLI 스텁 명령어 |
| **SUCCESS** | 전체 기능 정상 동작, DB/S3 데이터 정합성 100%, 발견 버그 즉시 수정 |
| **SCOPE** | 로컬 환경 한정 (Dashboard:3001 + API:8080 + PostgreSQL:5432 + MinIO:9000) |

---

## 1. Overview

### 1.1 Test Approach

Pragmatic E2E: Chrome 자동화로 Dashboard UX 검증 + tene CLI 실행으로 기능 검증 + DB/S3 직접 쿼리로 데이터 정합성 확인. 버그 발견 시 즉시 수정하며 진행.

### 1.2 Implementation State Summary

| Component | State | Notes |
|-----------|-------|-------|
| **API Storage** | PostgreSQL (DATABASE_URL 설정 시) | Auth만 항상 in-memory |
| **CLI Local** | 전체 구현 완료 | init/set/get/list/delete/run/export/import/env/passwd/recover |
| **CLI Cloud** | 대부분 구현 | login/push/pull/sync/billing 완료. team invite/members 일부 스텁 |
| **Dashboard** | 7개 페이지 구현 | Overview/Vaults/Detail/Devices/Team/Audit/Billing |
| **Auth** | Refresh token in-memory only | PostgreSQL 테이블 존재하나 미연결 (known bug) |

### 1.3 Known Limitations (테스트 시 고려)

| Issue | Impact | Workaround |
|-------|--------|------------|
| Auth refresh token in-memory | 서버 재시작 시 재로그인 필요 | 테스트 중 서버 재시작 금지 |
| `team members` 스텁 | CLI에서 팀 멤버 조회 불가 | Dashboard에서 확인 |
| `team invite` fetchUserPublicKey 스텁 | 팀 초대 불가 | API 직접 호출 또는 스킵 |
| Pro plan 체크 (JWT client-side) | push/pull/team/billing은 Pro 필요 | Free plan에서 에러 메시지 확인 |

---

## 2. Pre-Test Setup

### 2.1 Environment Verification

```bash
# Step 0-1: 서비스 상태 확인
curl -s http://localhost:8080/health           # {"status":"ok"}
curl -s http://localhost:8080/health/ready      # DB 연결 확인
curl -s http://localhost:3001 -o /dev/null -w "%{http_code}"  # 200 or 302

# Step 0-2: Docker 컨테이너 확인
docker ps --format "table {{.Names}}\t{{.Status}}"
# tene-db (healthy), tene-s3 (healthy)

# Step 0-3: DB 마이그레이션 확인
docker exec tene-db psql -U tene_admin -d tene -c "\dt"
# users, vaults, teams, team_members, devices, audit_logs, waitlist, refresh_tokens
```

### 2.2 CLI Build

```bash
# Step 0-4: tene CLI 로컬 빌드
go build -o tene-dev ./cmd/tene

# Step 0-5: 빌드 확인
./tene-dev version
```

### 2.3 Test Directory Setup

```bash
# Step 0-6: 테스트 전용 디렉토리 생성
mkdir -p /tmp/tene-qa-test
cd /tmp/tene-qa-test
```

---

## 3. Test Phase 1: Authentication Flow

### 3.1 Dashboard OAuth Login (Chrome)

| Step | Action | Tool | Expected | Verify Method |
|------|--------|------|----------|---------------|
| **1.1** | Navigate to http://localhost:3001 | Chrome navigate | /login 페이지 리다이렉트 | URL 확인 |
| **1.2** | 스크린샷 캡처 | Chrome screenshot | 로그인 페이지 렌더링 확인 | - |
| **1.3** | Console 에러 확인 | Chrome console | 에러 없음 | read_console |
| **1.4** | "Continue with GitHub" 클릭 | Chrome click | GitHub OAuth 페이지 | URL 확인 |
| **1.5** | GitHub tomo-kay 로그인 | Chrome (사용자 개입) | OAuth 인증 완료 | - |
| **1.6** | Callback 처리 확인 | Chrome | /auth/callback → / 리다이렉트 | URL 확인 |
| **1.7** | Overview 페이지 로드 | Chrome | 인증된 대시보드 표시 | 스크린샷 |
| **1.8** | localStorage 확인 | Chrome JS | `tene-auth` key에 accessToken, user 존재 | JS 실행 |
| **1.9** | API 로그 확인 | Bash | OAuth + token 발급 로그 | tail log |

### 3.2 Dashboard Auth State Verification

```javascript
// Step 1.8: localStorage 검증 코드
const auth = JSON.parse(localStorage.getItem('tene-auth'));
console.log('[QA] Auth state:', JSON.stringify({
  hasAccessToken: !!auth?.state?.accessToken,
  hasRefreshToken: !!auth?.state?.refreshToken,
  isAuthenticated: auth?.state?.isAuthenticated,
  user: auth?.state?.user
}, null, 2));
```

### 3.3 DB User Verification

```sql
-- Step 1.10: 사용자 레코드 확인
SELECT id, email, name, auth_provider, github_id, plan, created_at
FROM users
WHERE auth_provider = 'github'
ORDER BY created_at DESC
LIMIT 5;
```

---

## 4. Test Phase 2: Dashboard Page Navigation

### 4.1 Page-by-Page Verification

| Step | Page | Route | Check Items | Console Check |
|------|------|-------|-------------|---------------|
| **2.1** | Overview | `/` | StatCards 렌더링, Quick Start 안내, 사이드바 네비게이션 | 에러 없음 |
| **2.2** | Vaults | `/vaults` | Empty State 메시지, `tene push` 안내 텍스트 | API 호출 확인 |
| **2.3** | Devices | `/devices` | Empty State 또는 디바이스 카드, 온/오프라인 표시 | 에러 없음 |
| **2.4** | Team | `/team` | Empty State, Pro plan 안내, Invite 버튼 | 에러 없음 |
| **2.5** | Audit Log | `/audit` | Empty State, 필터 버튼 5개 (All/push/pull/login/delete/create) | 에러 없음 |
| **2.6** | Billing | `/billing` | Free/Pro plan 카드, Usage bars (0/50 vaults 등) | 에러 없음 |

### 4.2 Each Page Test Protocol

```
For each page:
1. Navigate to route
2. Wait for page load (2s)
3. Read console messages (filter errors)
4. Take screenshot
5. Read page text (verify key elements)
6. Check network requests (API calls successful)
```

### 4.3 Cross-Page Navigation Test

| Step | From | To | Method | Expected |
|------|------|-----|--------|----------|
| **2.7** | Overview | Vaults | 사이드바 클릭 | /vaults 로드 |
| **2.8** | Vaults | Team | 사이드바 클릭 | /team 로드 |
| **2.9** | Team | Billing | 사이드바 클릭 | /billing 로드 |
| **2.10** | Any | Overview | 로고/홈 클릭 | / 로드 |

---

## 5. Test Phase 3: CLI Local Commands

### 5.1 Vault Initialization

| Step | Command | Expected Output | Exit Code |
|------|---------|-----------------|:---------:|
| **3.1** | `./tene-dev init test-qa-project` | vault 생성, master password 프롬프트, 12-word recovery key 표시 | 0 |
| **3.2** | `./tene-dev whoami` | project: test-qa-project, vault status, keychain info | 0 |

> **Note**: init은 interactive (password 입력 필요). 테스트 시 `echo "testpassword123!" | ./tene-dev init test-qa-project` 또는 직접 입력.

### 5.2 Secret CRUD

| Step | Command | Expected | Verify |
|------|---------|----------|--------|
| **3.3** | `./tene-dev set DB_HOST localhost` | 저장 성공 메시지 | - |
| **3.4** | `./tene-dev set DB_PORT 5432` | 저장 성공 | - |
| **3.5** | `./tene-dev set DB_NAME tene_test` | 저장 성공 | - |
| **3.6** | `./tene-dev set API_KEY sk-test-abc123` | 저장 성공 | - |
| **3.7** | `./tene-dev set JWT_SECRET super-secret-jwt-key-32bytes!!` | 저장 성공 | - |
| **3.8** | `./tene-dev get DB_HOST` | `localhost` 출력 | stdout |
| **3.9** | `./tene-dev get API_KEY` | `sk-test-abc123` 출력 | stdout |
| **3.10** | `./tene-dev list` | 5개 키 목록, 값 마스킹 (****) | stdout |
| **3.11** | `./tene-dev list --json` | JSON 배열 형태 출력 | JSON 파싱 |
| **3.12** | `./tene-dev delete API_KEY --force` | 삭제 성공 | - |
| **3.13** | `./tene-dev list` | 4개 키 (API_KEY 없음) | stdout |
| **3.14** | `./tene-dev get API_KEY` | 에러: not found | exit 1 |

### 5.3 Environment Variable Injection

| Step | Command | Expected |
|------|---------|----------|
| **3.15** | `./tene-dev run -- env \| grep DB_` | DB_HOST=localhost, DB_PORT=5432, DB_NAME=tene_test 출력 |
| **3.16** | `./tene-dev run -- echo $JWT_SECRET` | super-secret-jwt-key-32bytes!! 출력 |

### 5.4 Export / Import

| Step | Command | Expected |
|------|---------|----------|
| **3.17** | `./tene-dev export > /tmp/tene-qa-export.env` | .env 형식 파일 생성 |
| **3.18** | `cat /tmp/tene-qa-export.env` | DB_HOST=localhost 등 4개 키 |
| **3.19** | `./tene-dev export --encrypted > /tmp/tene-qa-backup.enc` | 암호화된 백업 |
| **3.20** | 새 vault에서 `./tene-dev import /tmp/tene-qa-export.env` | 4개 시크릿 임포트 성공 |

### 5.5 Environment Management

| Step | Command | Expected |
|------|---------|----------|
| **3.21** | `./tene-dev env` | 현재 환경 표시 (default) |
| **3.22** | `./tene-dev env create staging` | staging 환경 생성 |
| **3.23** | `./tene-dev env staging` | staging으로 전환 |
| **3.24** | `./tene-dev set STAGING_VAR test-value` | staging 환경에 저장 |
| **3.25** | `./tene-dev list` | STAGING_VAR만 표시 |
| **3.26** | `./tene-dev env default` | default로 복귀 |
| **3.27** | `./tene-dev list` | 4개 키 (staging 키 안 보임) |

### 5.6 Password Change

| Step | Command | Expected |
|------|---------|----------|
| **3.28** | `./tene-dev passwd` | 현재 비밀번호 확인 → 새 비밀번호 설정 → 전체 re-encrypt → 새 recovery key |
| **3.29** | `./tene-dev get DB_HOST` | 새 비밀번호로 복호화 성공, `localhost` 출력 |

---

## 6. Test Phase 4: CLI Cloud + Dashboard Sync

### 6.1 CLI Login

| Step | Command | Expected | Notes |
|------|---------|----------|-------|
| **4.1** | `./tene-dev login` | 브라우저 열림 → GitHub OAuth → callback → 토큰 저장 | 브라우저에서 tomo-kay 로그인 필요 |
| **4.2** | `./tene-dev whoami` | user_id, plan 표시 (또는 cloud 관련 정보) | - |

### 6.2 Vault Push + Dashboard Verification

| Step | Action | Tool | Expected | Dashboard Verify |
|------|--------|------|----------|-----------------|
| **4.3** | `./tene-dev push` | CLI | vault 암호화 → S3 업로드 성공 | - |
| **4.4** | Dashboard /vaults 새로고침 | Chrome | test-qa-project vault 표시 | 스크린샷 |
| **4.5** | Vault 클릭 → Detail | Chrome | version, secret_count, size 표시 | 스크린샷 |
| **4.6** | Dashboard /audit 확인 | Chrome | vault.push 이벤트 기록 | 스크린샷 |

### 6.3 Vault Pull

| Step | Command | Expected |
|------|---------|----------|
| **4.7** | `./tene-dev pull` | S3 다운로드 → 복호화 → 로컬 vault 업데이트 |
| **4.8** | `./tene-dev list` | push 전과 동일한 시크릿 목록 |

### 6.4 Push 후 Secret 변경 → 재Push

| Step | Action | Expected |
|------|--------|----------|
| **4.9** | `./tene-dev set NEW_KEY new_value` | 새 시크릿 추가 |
| **4.10** | `./tene-dev push` | version 증가 (v2) |
| **4.11** | Dashboard /vaults 확인 | version=2, secret_count 증가 | 

### 6.5 Team Commands (Free Plan 제한 테스트)

| Step | Command | Expected |
|------|---------|----------|
| **4.12** | `./tene-dev team create test-team` | Pro plan 필요 에러 OR 팀 생성 성공 |
| **4.13** | `./tene-dev team list` | 팀 목록 (있으면 표시, 없으면 빈 목록) |
| **4.14** | `./tene-dev billing` | Free plan 상태 표시 |

### 6.6 Billing Commands

| Step | Command | Expected |
|------|---------|----------|
| **4.15** | `./tene-dev billing` | 현재 구독 상태 (Free) 표시 |
| **4.16** | `./tene-dev billing upgrade` | LemonSqueezy checkout URL 열기 시도 |

---

## 7. Test Phase 5: Data Verification

### 7.1 PostgreSQL Queries

```sql
-- Step 5.1: Users
SELECT id, email, name, auth_provider, github_id, plan,
       created_at, updated_at
FROM users;

-- Step 5.2: Vaults
SELECT id, user_id, project_name, vault_version, secret_count,
       size, s3_key, created_at, last_pushed_at
FROM vaults;

-- Step 5.3: Audit Logs
SELECT id, user_id, vault_id, action, detail, ip_address, created_at
FROM audit_logs
ORDER BY created_at DESC;

-- Step 5.4: Devices
SELECT id, user_id, device_name, last_seen_at, created_at
FROM devices;

-- Step 5.5: Teams (if any)
SELECT id, name, slug, owner_id, created_at
FROM teams;

-- Step 5.6: Team Members (if any)
SELECT team_id, user_id, role, env_permissions, joined_at
FROM team_members;

-- Step 5.7: Waitlist (if any)
SELECT * FROM waitlist;

-- Step 5.8: Refresh Tokens (should be empty - known in-memory bug)
SELECT * FROM refresh_tokens;
```

### 7.2 MinIO S3 Verification

```bash
# Step 5.9: MinIO 객체 목록 확인
# Option A: aws cli
aws --endpoint-url http://localhost:9000 s3 ls s3://tene-vault-dev/ --recursive \
  --profile minio 2>/dev/null

# Option B: mc (MinIO Client)
mc alias set local http://localhost:9000 minioadmin minioadmin
mc ls local/tene-vault-dev/ --recursive

# Option C: curl (직접 API)
curl -s http://localhost:9000/minio/health/live
```

### 7.3 API Log Analysis

```bash
# Step 5.10: API 로그에서 에러 확인
grep -i "error\|fail\|panic" /tmp/tene-api.log | tail -20

# Step 5.11: OAuth 관련 로그
grep -i "oauth\|auth\|token\|callback" /tmp/tene-api.log | tail -20

# Step 5.12: Vault 관련 로그
grep -i "vault\|push\|pull\|s3" /tmp/tene-api.log | tail -20
```

### 7.4 Cross-Validation Checks

| Check | Source A | Source B | Match Criteria |
|-------|---------|---------|----------------|
| User 존재 | Dashboard /auth/me | DB users | user_id 일치 |
| Vault 목록 | Dashboard /vaults | DB vaults | id, project_name, version 일치 |
| Audit 기록 | Dashboard /audit | DB audit_logs | action, timestamp 일치 |
| Vault blob | DB vaults.s3_key | MinIO 객체 | 파일 존재 + 크기 일치 |
| Secret count | CLI `tene list` | DB vaults.secret_count | 개수 일치 |

---

## 8. Bug Discovery & Fix Protocol

### 8.1 Bug Severity Classification

| Severity | Definition | Action |
|----------|-----------|--------|
| **Critical** | Auth 실패, 데이터 손실, Crash | 즉시 수정 후 재테스트 |
| **High** | 기능 미동작, 잘못된 데이터 | 즉시 수정 |
| **Medium** | UI 깨짐, 부정확한 메시지 | 기록 후 나중에 수정 |
| **Low** | 미관, 미세 UX 개선 | 기록만 |

### 8.2 Fix-During-Test Protocol

```
1. 버그 발견
2. 심각도 판단
3. Critical/High → 즉시 수정
   - 코드 수정
   - 서버 재빌드/재시작 (필요시)
   - 해당 테스트 재실행
4. Medium/Low → 기록 후 계속 진행
5. 모든 버그는 Design 문서의 Test Results에 기록
```

---

## 9. Test Execution Checklist

### Phase 0: Pre-Test Setup
- [ ] 0-1. API health check
- [ ] 0-2. DB readiness check
- [ ] 0-3. Docker containers healthy
- [ ] 0-4. DB 마이그레이션 확인
- [ ] 0-5. tene CLI 빌드 (`go build -o tene-dev ./cmd/tene`)
- [ ] 0-6. MinIO 버킷 확인
- [ ] 0-7. API 로그 확인 시작

### Phase 1: Auth Flow
- [ ] 1.1. Dashboard /login 접속
- [ ] 1.2. 로그인 페이지 스크린샷
- [ ] 1.3. Console 에러 없음
- [ ] 1.4. GitHub OAuth 클릭
- [ ] 1.5. tomo-kay 로그인
- [ ] 1.6. Callback → Overview 리다이렉트
- [ ] 1.7. Overview 스크린샷
- [ ] 1.8. localStorage auth 확인
- [ ] 1.9. DB users 테이블 확인
- [ ] 1.10. API 로그 확인

### Phase 2: Dashboard Pages
- [ ] 2.1. Overview 페이지
- [ ] 2.2. Vaults 페이지 (Empty)
- [ ] 2.3. Devices 페이지
- [ ] 2.4. Team 페이지
- [ ] 2.5. Audit Log 페이지
- [ ] 2.6. Billing 페이지
- [ ] 2.7-2.10. 네비게이션 테스트

### Phase 3: CLI Local
- [ ] 3.1-3.2. init + whoami
- [ ] 3.3-3.7. set (5개 시크릿)
- [ ] 3.8-3.11. get + list
- [ ] 3.12-3.14. delete + verify
- [ ] 3.15-3.16. run (env injection)
- [ ] 3.17-3.20. export + import
- [ ] 3.21-3.27. env management
- [ ] 3.28-3.29. passwd change

### Phase 4: CLI Cloud + Dashboard Sync
- [ ] 4.1-4.2. login + whoami
- [ ] 4.3-4.6. push + Dashboard verify
- [ ] 4.7-4.8. pull
- [ ] 4.9-4.11. modify + re-push + Dashboard
- [ ] 4.12-4.14. team commands
- [ ] 4.15-4.16. billing commands

### Phase 5: Data Verification
- [ ] 5.1-5.8. PostgreSQL 전체 테이블 쿼리
- [ ] 5.9. MinIO S3 객체 확인
- [ ] 5.10-5.12. API 로그 분석
- [ ] 5.13. Cross-validation 체크

---

## 10. Test Results Template

### Summary

| Metric | Result |
|--------|--------|
| Total Test Steps | ~65 |
| Passed | - |
| Failed | - |
| Bugs Found | - |
| Bugs Fixed | - |
| Overall Status | - |

### Bugs Found

| # | Severity | Phase | Description | Status |
|---|----------|-------|-------------|--------|
| B-01 | - | - | - | - |

### Phase Results

| Phase | Steps | Passed | Failed | Notes |
|-------|-------|--------|--------|-------|
| 0. Setup | 7 | - | - | - |
| 1. Auth | 10 | - | - | - |
| 2. Dashboard | 10 | - | - | - |
| 3. CLI Local | 29 | - | - | - |
| 4. CLI Cloud | 16 | - | - | - |
| 5. Data | 13 | - | - | - |

---

## 11. Implementation Guide

### 11.1 Execution Order

1. **Pre-Test**: Environment 확인 + CLI 빌드
2. **Chrome Setup**: Tab 생성 + Dashboard 접속
3. **Phase 1**: Auth (Chrome OAuth → DB verify)
4. **Phase 2**: Dashboard 7개 페이지 순회 (Chrome)
5. **Phase 3**: CLI Local 명령어 (Terminal)
6. **Phase 4**: CLI Cloud + Dashboard 동기화 (Chrome + Terminal)
7. **Phase 5**: DB/S3/Log 크로스 검증 (Terminal)

### 11.2 Tool Usage

| Task | Tool |
|------|------|
| Dashboard 탐색 | `mcp__claude-in-chrome__navigate`, `read_page` |
| 페이지 텍스트 | `mcp__claude-in-chrome__get_page_text` |
| Console 에러 | `mcp__claude-in-chrome__read_console_messages` |
| JS 실행 | `mcp__claude-in-chrome__javascript_tool` |
| 네트워크 확인 | `mcp__claude-in-chrome__read_network_requests` |
| CLI 명령어 | `Bash` tool |
| DB 쿼리 | `docker exec tene-db psql -U tene_admin -d tene -c "SQL"` |
| S3 확인 | `aws --endpoint-url http://localhost:9000 s3 ls` or mc |
| API 로그 | `tail /tmp/tene-api.log` |

### 11.3 Session Guide

| Module | Scope | Estimated Steps |
|--------|-------|:---------------:|
| module-0 | Pre-Test Setup | 7 |
| module-1 | Auth Flow (Dashboard + DB) | 10 |
| module-2 | Dashboard Page Navigation | 10 |
| module-3 | CLI Local Commands | 29 |
| module-4 | CLI Cloud + Dashboard Sync | 16 |
| module-5 | Data Cross-Validation | 13 |

**Recommended**: 전체를 단일 세션에서 순차 실행 (데이터 의존성 있음)
