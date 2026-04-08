# Tene Cloud 개선 — Plan Plus

> **버전**: v1.0
> **일자**: 2026-04-08
> **기능**: tene-cloud-improvement
> **방법론**: Plan Plus (브레인스토밍 강화 PDCA 계획)
> **상위 문서**: CTO Lead 10-관점 팀 분석 + PM PRD (tene-cloud.prd.md)
> **브랜치**: fix/staging-qa (staging에서 분기)

---

## 요약

| 관점 | 설명 |
|------|------|
| **문제** | Tene Cloud API가 프로덕션 수준에 미달: in-memory 전용(ECS 재시작 시 데이터 유실), billing webhook panic, auth token plaintext 저장, Dashboard 15분 로그아웃. 37개 이슈 발견 |
| **해결** | Bottom-Up 4-Sprint 접근: PG Repository → 핵심 UX → Dashboard 완전 연동 → 마무리. pgx v5 + auth code exchange + go-keyring + token refresh |
| **기능적 UX 효과** | 가입→결제→push/pull→팀 초대→Dashboard 확인까지 E2E 끊김 없이 동작. Free 유저에게 친절한 업그레이드 안내 |
| **핵심 가치** | Zero-Knowledge 보안 유지하면서 프로덕션 안정성 확보. $5/mo Solo + $10/user/mo Team 유료 전환 가능 상태 달성 |

---

## 1. 사용자 의도 탐색 (Phase 1)

### 핵심 문제
Tene Cloud가 배포되었지만 프로덕션 사용 불가:
- P0 Critical 4건이 유료 전환/데이터 영속성/보안을 차단
- Dashboard가 35% 연동 (6/17 endpoints)
- 인증-결제 플로우가 분리되어 유료 전환 마찰 높음

### 대상 사용자
- **솔로 개발자**: 멀티 디바이스 동기화 + 클라우드 백업 (Pro $5/mo 주 타겟)
- **소규모 팀**: 시크릿 공유 + RBAC + 환경별 권한 (Team $10/user/mo)
- **AI-first 개발자**: 에이전트 자동 인식/주입이 핵심 차별점

### 성공 기준
1. **E2E 작동**: 가입 → 결제 → push/pull → 팀 초대 → Dashboard 확인 전체 플로우 무결
2. **보안 무결성**: Zero-Knowledge 유지, auth token keychain 저장, URL token 노출 제거
3. **유료 전환 가능**: LemonSqueezy webhook → plan 업데이트 → JWT 반영 정상 동작

---

## 2. 대안 탐색 (Phase 2)

### 접근법 A: Bottom-Up (선택됨)
DB 레이어 → API 수정 → CLI 개선 → Dashboard 연동

- **장점**: 기반부터 탄탄, 각 Sprint이 독립적으로 테스트 가능
- **단점**: 사용자 경험 개선이 Sprint 2-3에 집중
- **적합 상황**: 현재 상태 (기반이 없는 상태에서 UX를 먼저 하면 허공에 구축)

### 접근법 B: 사용자 여정 우선 (미선택)
가입→Sync 먼저, 나머지 나중

- **장점**: 핵심 유저 여정이 빨리 완성
- **단점**: PG 일부만 구현 후 나중에 다시 건드려야 함. 중복 작업
- **적합 상황**: 빠른 데모가 필요한 경우

### 접근법 C: 리스크 우선 (미선택)
PG 전체 + 보안 전체 먼저, UX 나중

- **장점**: 안정성 최우선
- **단점**: 사용자 경험 개선이 2주 이상 뒤로 밀림
- **적합 상황**: 보안 감사가 임박한 경우

**결정 근거**: 현재 in-memory 전용 상태에서 Dashboard/CLI UX를 먼저 개선해도 ECS 재시작 시 모든 데이터가 사라지므로 무의미. Bottom-Up이 가장 합리적.

---

## 3. YAGNI 검토 (Phase 3)

### 포함 항목 (Sprint 1-4)

| ID | 항목 | Sprint | 근거 |
|----|------|:------:|------|
| B-07 | PostgreSQL repository layer (7 stores) | 1 | 없으면 프로덕션 불가 |
| A-02 | Billing webhook crash fix (nil UserStore) | 1 | 유료 전환 차단 |
| S-01 | Auth token keychain 이전 | 1 | 보안 P0 |
| F-08 | Dashboard token refresh | 1 | 15분 로그아웃 해결 |
| B-05 | CLI plan pre-check | 2 | Free 유저 UX |
| A-04 | Auth code exchange | 2 | 보안 P1 |
| F-07 | Billing upgrade button wiring | 2 | 유료 전환 경로 |
| F-01 | Dashboard overview live data | 2 | 핵심 페이지 |
| F-02~06 | Vault/Team/Device/Audit 페이지 | 3 | Dashboard 완성 |
| B-01~02 | Missing API endpoints | 3 | Dashboard 연동 필요 |
| S-07 | Team key rotation | 3 | 보안 완성 |
| T-01~05 | Handler tests + E2E | 4 | 품질 보증 |
| I-01 | Migration 자동화 | 4 | 운영 안정화 |

### 보류 항목 (범위 외)

| 항목 | 이유 |
|------|------|
| S-05: HS256→ES256 전환 | 현재 HS256으로 기능 동작. 마이크로서비스 전환 시 필요 |
| S-03: localStorage→httpOnly cookie | Token refresh 구현 후 평가 |
| S-06: CSRF 토큰 | Bearer 인증 사용 중이므로 우선순위 낮음 |
| Redis 세션/레이트리밋 | In-memory rate limiter가 단일 ECS에서 충분 |
| WAF rules | 트래픽 증가 후 도입 |
| Playwright E2E | Sprint 4에서 Go 테스트 먼저, 프론트엔드 E2E는 후순위 |
| Load test (k6) | 사용자 수 증가 후 |

---

## 4. 아키텍처 설계

### 4.1 PostgreSQL 레포지토리 레이어

```
현재                                  변경 후
─────                                ──────
internal/api/handler/                internal/api/handler/
  vault_store_mem.go (MemVaultStore)   (유지, 테스트용)
  team.go (MemTeamStore)               (유지, 테스트용)
  device.go (MemDeviceStore)           (유지, 테스트용)
  ...                                  ...

                                     internal/repository/         ← NEW
                                       postgres/
                                         postgres.go             # pgxpool.Pool
                                         user_repo.go            # users table
                                         vault_repo.go           # vaults table
                                         team_repo.go            # teams + team_members
                                         device_repo.go          # devices table
                                         audit_repo.go           # audit_logs table
                                         refresh_token_repo.go   # refresh_tokens table
                                         waitlist_repo.go        # waitlist table

internal/api/server.go               internal/api/server.go
  vaultStore = MemVaultStore{}          pool := pgxpool.New(dbURL)
  teamStore = MemTeamStore{}            vaultStore = postgres.NewVaultRepo(pool)
  billing.NewService(..., nil)          teamStore = postgres.NewTeamRepo(pool)
                                        billing.NewService(..., postgres.NewUserRepo(pool))
```

**DB Driver**: `github.com/jackc/pgx/v5` (pure Go, PostgreSQL 전용, connection pool 내장)

**Connection Pool**: MaxConns=10, MinConns=2 (ECS 0.25 vCPU에 적합)

### 4.2 인증 보안 강화

```
현재 (Token in URL)                   변경 후 (Auth Code Exchange)
────────────────                      ────────────────────────
OAuth callback                        OAuth callback
  → redirect?access_token=...          → redirect?code=TEMP_CODE
  → 브라우저 히스토리에 저장됨            → 30초 TTL, 1회 사용
                                        → POST /api/v1/auth/exchange
                                          { code: "..." }
                                          → { access_token, refresh_token }

CLI auth storage                      CLI auth storage
  ~/.tene/auth.json (plaintext)        go-keyring (OS keychain)
  0600 permissions                     macOS: Keychain
                                       Linux: libsecret
                                       Fallback: --no-keychain → file
```

### 4.3 Dashboard 토큰 갱신

```typescript
// apps/dashboard/src/lib/api.ts — interceptor 추가
private async request<T>(path, options): Promise<T> {
  const res = await fetch(...)
  if (res.status === 401 && this.refreshToken) {
    const newTokens = await this.refreshTokens()
    // retry original request with new token
  }
}

// apps/dashboard/src/lib/auth-store.ts
login(accessToken, refreshToken) {
  set({ accessToken, refreshToken, isAuthenticated: true })
  api.setToken(accessToken)
  api.setRefreshToken(refreshToken)  // 실제 저장
  // Fetch user profile
  const user = await api.getMe()
  set({ user })
}
```

### 4.4 데이터 흐름 (변경 후)

```
Signup:   Landing → OAuth → plan=free → Dashboard
          (if intent=upgrade → auto-redirect to billing)
Upgrade:  Dashboard /billing → checkout → LemonSqueezy
          → webhook → PG: user.plan='pro' → next JWT has plan=pro
Sync:     CLI push → JWT plan check (local) → API plan check (server) → S3
Team:     CLI/Dashboard → API → PG → X25519 key wrap per member
Refresh:  Dashboard 401 → POST /auth/refresh → new tokens → retry
```

---

## 5. Sprint 계획

### Sprint 1: 기반 구축 (7-8일) — P0

| 작업 | 설명 | 파일 | 예상 |
|------|-------------|-------|:---:|
| 1-1 | pgx/v5 + repository package 생성 | `go.mod`, `internal/repository/postgres/*.go` | 5d |
| 1-2 | server.go MemStore → PgStore 교체 | `internal/api/server.go` | 1d |
| 1-3 | CLI auth → go-keyring 이전 | `internal/cli/login.go`, `internal/cli/auth.go` | 1d |
| 1-4 | Dashboard token refresh + user fetch | `apps/dashboard/src/lib/api.ts`, `auth-store.ts` | 1d |

**Sprint 1 완료 조건**:
- [x] API 서버가 PostgreSQL에 연결하여 데이터 영속
- [x] billing webhook 정상 동작 (plan 업데이트)
- [x] CLI auth token이 OS keychain에 저장
- [x] Dashboard 15분 로그아웃 해결

### Sprint 2: 핵심 경험 (4-5일) — P1

| 작업 | 설명 | 파일 | 예상 |
|------|-------------|-------|:---:|
| 2-1 | Auth code exchange endpoint | `handler/auth.go`, `server.go` | 1.5d |
| 2-2 | CLI pre-flight plan check | `internal/cli/push.go`, `pull.go`, `sync_cmd.go`, `team.go` | 0.5d |
| 2-3 | Landing Pro CTA intent=upgrade | `apps/web/src/data/pricing.ts`, Dashboard login page | 0.5d |
| 2-4 | Dashboard billing button + overview | `apps/dashboard/src/app/(dashboard)/billing/page.tsx`, overview | 1.5d |

**Sprint 2 완료 조건**:
- [x] Token이 URL에 노출되지 않음
- [x] Free 유저 push 시 "Upgrade to Pro" 안내
- [x] Landing Pro CTA → login → 자동 checkout redirect
- [x] Dashboard billing 버튼 → LemonSqueezy checkout

### Sprint 3: 기능 완성 (5-6일) — P2

| 작업 | 설명 | 파일 | 예상 |
|------|-------------|-------|:---:|
| 3-1 | Dashboard API client 11개 메서드 추가 | `apps/dashboard/src/lib/api.ts` | 1d |
| 3-2 | GET /vaults/:id + GET /teams/:id/members endpoints | `handler/vault.go`, `handler/team.go`, `server.go` | 1d |
| 3-3 | Vault list + detail 페이지 라이브 연동 | `apps/dashboard/src/app/(dashboard)/vaults/` | 1d |
| 3-4 | Team 페이지 CRUD + invite 연동 | `apps/dashboard/src/app/(dashboard)/team/` | 1d |
| 3-5 | Device + Audit 페이지 연동 | `apps/dashboard/src/app/(dashboard)/devices/`, `audit/` | 1d |
| 3-6 | Team key rotation 구현 | `handler/team.go`, `internal/crypto/teamkey.go` | 1d |

**Sprint 3 완료 조건**:
- [x] Dashboard 모든 페이지가 실제 API 데이터 표시
- [x] API client 17/17 메서드 구현
- [x] Team 멤버 제거 시 실제 key rotation 수행

### Sprint 4: 마무리 (3-4일) — P3

| 작업 | 설명 | 파일 | 예상 |
|------|-------------|-------|:---:|
| 4-1 | API handler 단위 테스트 (table-driven) | `handler/*_test.go` | 2d |
| 4-2 | DB migration 자동 실행 (서버 시작 시) | `cmd/server/main.go`, goose 도입 | 0.5d |
| 4-3 | Health check readiness + graceful shutdown | `handler/health.go`, `cmd/server/main.go` | 0.5d |
| 4-4 | 문서 정렬 (CLAUDE.md, HS256 명시) | `CLAUDE.md`, `.claude/rules/` | 0.5d |

**Sprint 4 완료 조건**:
- [x] 핵심 handler 테스트 커버리지 확보
- [x] ECS 배포 시 migration 자동 실행
- [x] 문서와 코드 일치

---

## 6. 리스크 평가

| 리스크 | 가능성 | 영향 | 완화 방안 |
|------|:----------:|:------:|------------|
| PG repo 구현이 예상보다 오래 걸림 | Medium | High | 기존 interface 그대로 사용, 1:1 매핑 |
| LemonSqueezy webhook 테스트 어려움 | Medium | Medium | test mode + ngrok으로 로컬 테스트 |
| go-keyring이 Linux headless에서 실패 | Medium | Low | --no-keychain file fallback |
| Auth code exchange가 CLI callback 복잡화 | Low | Medium | 기존 localhost callback 패턴 유지 |

---

## 7. 브레인스토밍 기록

### 핵심 결정 사항

1. **Bottom-Up 선택**: In-memory 상태에서 UX 개선은 무의미 → DB부터
2. **pgx v5 선택**: database/sql 대신 — PostgreSQL 전용 기능, 내장 pool, 더 나은 성능
3. **Auth code exchange**: URL token 노출은 보안 이슈 — industry standard 패턴 적용
4. **Sprint 4에 테스트**: Sprint 1-3에서 기능 구현 후 테스트 집중 (TDD 대신 pragmatic)
5. **HS256 유지**: ES256 전환은 YAGNI — 단일 서버에서 불필요한 복잡성
6. **Redis 제외**: 단일 ECS에서 in-memory rate limiter 충분

### YAGNI 적용 사례 (불필요 기능 제거)
- ~~S3 presigned URL 캐싱~~ → pull 빈도가 낮아 불필요
- ~~Team billing (팀 단위 구독)~~ → 개인 Pro 구독으로 충분한 MVP
- ~~Audit log partitioning~~ → 데이터 소량, 추후 도입
- ~~OAuth provider 추가 (Google)~~ → GitHub만으로 타겟 유저 커버

---

## 8. 이슈 참조

전체 37개 이슈 상세는 `docs/01-plan/tene-cloud-improvement.team-analysis.md` 참조.

| 우선순위 | 수량 | ID |
|----------|:-----:|-----|
| P0 Critical | 4 | A-02, A-03/B-07/D-01, S-01, F-08/S-04 |
| P1 High | 4 | B-05, A-04/S-02, A-01, F-01/F-07/F-09 |
| P2 Medium | 12 | F-02~F-06, B-01~B-04, S-07, C-01~C-04 |
| P3 Low | 6 | S-05, S-03, S-06, D-02~D-04 |
| 보류 | 5 | Redis, WAF, k6, Playwright, ES256 |

---

## 9. 다음 단계

```
현재: Plan Plus ✓ (이 문서)
다음: /pdca design tene-cloud-improvement
      → Sprint별 상세 설계 (DB 스키마 매핑, API 시그니처, 컴포넌트 구조)
```

---

## 부록: 상위 문서

| 문서 | 경로 | 내용 |
|------|------|------|
| PM PRD | `docs/00-pm/tene-cloud.prd.md` | 49KB, 전체 제품 요구사항 |
| 팀 분석 | `docs/01-plan/tene-cloud-improvement.team-analysis.md` | 10-관점 분석, 37개 이슈 |
| 탐색 | `docs/00-pm/tene-discovery.md` | 기회 솔루션 트리 |
| 전략 | `docs/00-pm/tene-strategy.md` | 가치 제안 + 린 캔버스 |
| 시장 조사 | `docs/00-pm/tene-research.md` | 페르소나 + 경쟁사 + 시장 |
