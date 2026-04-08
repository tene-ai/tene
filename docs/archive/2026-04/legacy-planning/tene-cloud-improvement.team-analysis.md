# Tene Cloud 개선 -- 팀 분석

> 날짜: 2026-04-08
> 범위: Tene Cloud 인증-결제 흐름, Dashboard-API 통합, CLI 보안, 플랜 게이팅, 기능 완성도 갭에 대한 풀스택 분석.
> 방법: 전체 코드베이스 리뷰를 기반으로 한 10가지 관점의 전문가 패널 분석.

---

## 핵심 요약

| 관점 | 문제 | 해결책 | 영향 |
|------------|---------|----------|--------|
| 인증-결제 흐름이 분리됨: 사용자가 가입(무료) 후 별도로 업그레이드해야 함 | 통합 결제 우선 흐름: 랜딩 Pro CTA -> LemonSqueezy 결제 후 가입 직접 연결 | 마찰 감소, Pro 전환율 증가 |
| Dashboard에 6개 기능 UI가 있지만 API 호출은 3개만 구현 | 나머지 TanStack Query 훅 연결: teams, devices, audit, vault 상세 | Dashboard를 데모에서 프로덕션 도구로 전환 |
| CLI가 인증 토큰을 ~/.tene/auth.json에 평문 JSON으로 저장 | go-keyring(이미 의존성에 포함)으로 마이그레이션 + 파일 폴백 | 가장 높은 심각도의 보안 갭 해소 |
| 플랜 게이팅이 서버에만 적용; CLI 사용자에게 불명확한 HTTP 402 에러 발생 | push/pull/sync/team 전에 CLI 측 플랜 확인 + 친절한 업그레이드 메시지 추가 | 무료 사용자 UX 4배 개선 |

---

## 1. 아키텍처 분석 -- 인증-결제 흐름 통합

**전문가: 아키텍처 리드**

### 현재 상태

인증과 결제 흐름이 완전히 분리되어 있음:

```
Landing Pro CTA
  -> app.tene.sh/login (GitHub OAuth)
    -> plan=free user created
      -> Dashboard /billing
        -> "Upgrade to Pro" button
          -> POST /api/v1/billing/checkout
            -> LemonSqueezy checkout URL
              -> Webhook -> plan=pro
```

이 6단계 흐름에는 3개의 주요 이탈 지점이 있음:
1. 랜딩 -> Dashboard 로그인 ("Pro를 원함"에서 "먼저 가입"으로의 맥락 전환)
2. Dashboard 개요 -> 결제 페이지 (사용자가 직접 탐색해야 함)
3. 결제 -> 외부 결제 (Tene를 완전히 벗어남)

### 핵심 발견사항

**A-01: 랜딩 페이지에서 직접 결제 경로 없음**
- 파일: `apps/web/src/data/pricing.ts` 47번째 줄
- Pro CTA 액션이 `"signup"`으로 `${dashboardUrl}/login`에 연결
- 사용자 필요 동작: 로그인 -> 결제로 이동 -> 업그레이드 클릭 -> 결제 정보 입력
- 기대 동작: 랜딩 Pro CTA에서 직접 결제 제공 (로그인 + 결제를 한 번의 흐름으로)

**A-02: 결제 서비스에서 UserStore가 nil**
- 파일: `internal/api/server.go` 76번째 줄
- `billingSvc := billing.NewService(..., nil)` -- UserStore가 nil
- Webhook의 `HandleWebhook`이 `s.store.UpdatePlan()`을 호출하면 nil에서 panic 발생
- 이는 Webhook을 통한 플랜 업그레이드가 완전히 작동하지 않음을 의미

**A-03: 인메모리 스토어가 재시작 시 모든 상태를 잃음**
- 모든 스토어(auth states, refresh tokens, vaults, teams, devices, audit)가 인메모리
- 마이그레이션이 정의되어 있음에도 PostgreSQL 리포지토리 구현이 없음
- 프로덕션 배포 시 ECS 태스크 재시작으로 모든 사용자 데이터 손실

**A-04: 인증 콜백이 URL에 토큰을 노출**
- 파일: `internal/api/handler/auth.go` 177-186번째 줄
- 토큰이 쿼리 파라미터로 전달: `?access_token=...&refresh_token=...`
- 브라우저 히스토리, 서버 로그, Referrer 헤더에 노출
- 표준 방식은 단기 인증 코드를 사용하여 토큰으로 교환하는 것

### 권장 아키텍처

```
Option 1: Checkout-First Flow (Recommended)
  Landing Pro CTA -> LemonSqueezy checkout (email only)
    -> Webhook: create user with plan=pro
      -> Post-checkout redirect -> Dashboard /login
        -> GitHub OAuth -> link to existing pro user by email

Option 2: Login-Then-Checkout Flow (Simpler)
  Landing Pro CTA -> Dashboard /login?intent=upgrade
    -> GitHub OAuth -> plan=free
      -> Auto-redirect to /billing/checkout
        -> LemonSqueezy -> Webhook -> plan=pro

Option 3: Embedded Checkout (Best UX)
  Landing Pro CTA -> LemonSqueezy.js overlay on landing page
    -> Webhook -> user created with plan=pro
      -> Email sent with setup link -> Dashboard
```

| 옵션 | 복잡도 | 전환율 | 권장 대상 |
|--------|-----------|------------|-----------------|
| 1 | 중간 | 최고 | 매출 중심 론칭 |
| 2 | 낮음 | 중간 | MVP/현재 단계 |
| 3 | 높음 | 최고 | PMF 달성 후 |

**권장사항**: 즉시 구현을 위해 옵션 2. 최소한의 변경만 필요:
- 로그인 페이지에 `?intent=upgrade` 쿼리 파라미터 지원 추가
- OAuth 콜백 후 intent를 확인하고 결제 체크아웃으로 자동 리다이렉트
- `apps/web/src/data/pricing.ts`의 CTA 액션을 intent 파라미터가 포함된 `"signup"`으로 수정

---

## 2. 프론트엔드 분석 -- Dashboard-API 통합

**전문가: 프론트엔드 아키텍트**

### API 클라이언트 커버리지

| API 엔드포인트 | 서버 라우트 | Dashboard 메서드 | 상태 |
|-------------|-------------|-----------------|--------|
| GET /auth/me | authH.Me | api.getMe() | 구현됨 |
| GET /vaults | vaultH.List | api.listVaults() | 구현됨 |
| GET /vaults/:id | (핸들러 없음) | -- | 서버 + 클라이언트 누락 |
| GET /billing/subscription | billingH.GetSubscription | api.getSubscription() | 구현됨 |
| POST /billing/checkout | billingH.CreateCheckout | api.createCheckout() | 구현됨 |
| POST /billing/portal | billingH.CreatePortal | api.getPortal() | 구현됨 |
| POST /waitlist | waitlistH.Register | api.joinWaitlist() | 구현됨 |
| GET /teams | teamH.List | -- | 클라이언트 누락 |
| POST /teams | teamH.Create | -- | 클라이언트 누락 |
| POST /teams/:id/invite | teamH.Invite | -- | 클라이언트 누락 |
| DELETE /teams/:id/members/:uid | teamH.RemoveMember | -- | 클라이언트 누락 |
| PATCH /teams/:id/members/:uid/role | teamH.UpdateRole | -- | 클라이언트 누락 |
| POST /devices | deviceH.Register | -- | 클라이언트 누락 |
| GET /devices | deviceH.List | -- | 클라이언트 누락 |
| DELETE /devices/:id | deviceH.Delete | -- | 클라이언트 누락 |
| GET /audit | auditH.List | -- | 클라이언트 누락 |

**커버리지: 6/17 엔드포인트 (35%)**

### 페이지별 분석

**F-01: 개요 페이지가 정적임**
- 파일: `apps/dashboard/src/app/(dashboard)/page.tsx`
- 모든 StatCard에 하드코딩된 `value={0}` 표시
- TanStack Query 훅 없음; API 호출 없음
- 수정: `/vaults` 개수, `/audit` 이벤트 수, `/devices` 수에 대한 `useQuery` 추가

**F-02: Vault 페이지에 데이터 페칭 없음**
- 파일: `apps/dashboard/src/app/(dashboard)/vaults/page.tsx`
- 빈 테이블이 있는 순수 정적 HTML
- 수정: VaultTable 컴포넌트와 함께 `useQuery('vaults', api.listVaults)` 추가

**F-03: Vault 상세 페이지가 플레이스홀더 데이터 사용**
- 파일: `apps/dashboard/src/app/(dashboard)/vaults/[id]/page.tsx`
- 하드코딩된 `vault = { id, project_name: "my-project", vault_version: 3 }`
- 서버에 GET /vaults/:id 엔드포인트 없음 (list + push/pull만 있음)
- 수정: 단일 vault 서버 엔드포인트 + 클라이언트 메서드 + TanStack Query 추가

**F-04: 팀 페이지에 API 통합 없음**
- 파일: `apps/dashboard/src/app/(dashboard)/team/page.tsx`
- `const team = null; const members: Member[] = [];` -- 순수 플레이스홀더
- InviteModal의 `onInvite`가 콘솔에 로그: `console.log("Invite:", email, role)`
- 수정: API 클라이언트에 5개 팀 API 메서드 추가 + TanStack Query 훅 + 뮤테이션

**F-05: 디바이스 페이지에 API 통합 없음**
- 파일: `apps/dashboard/src/app/(dashboard)/devices/page.tsx`
- API 호출 없이 "No devices registered" 표시
- 수정: API 클라이언트에 `listDevices()`와 `deleteDevice()` 추가

**F-06: 감사 페이지에 API 통합 없음 + 필터 작동 안 함**
- 파일: `apps/dashboard/src/app/(dashboard)/audit/page.tsx`
- 필터 버튼이 렌더링되지만 클릭 핸들러 없음
- 데이터 페칭 없음; 빈 테이블만 항상 표시
- 수정: 클라이언트에 `listAuditLogs(filter)` 추가 + 필터 상태와 쿼리

**F-07: 결제 "Upgrade to Pro" 버튼에 onClick 핸들러 없음**
- 파일: `apps/dashboard/src/app/(dashboard)/billing/page.tsx` 44번째 줄
- 이벤트 핸들러 없는 순수 `<button>`
- 수정: `api.createCheckout(email)`을 호출하고 리다이렉트하는 onClick 추가

**F-08: 인증 스토어가 토큰 갱신을 처리하지 않음**
- 파일: `apps/dashboard/src/lib/auth-store.ts`
- `login()`이 `refreshToken`을 받지만 `_refreshToken`(미사용)에 할당
- 갱신 로직 없음; 액세스 토큰 만료(15분) 시 사용자가 자동 로그아웃됨
- 수정: 영구 저장 상태에 `refreshToken` 추가 + API 클라이언트 인터셉터에서 자동 갱신

**F-09: 인증 스토어가 사용자 프로필을 가져오지 않음**
- 로그인 후 `accessToken`만 저장됨. `user`는 `setUser()`로 별도 설정
- 하지만 `setUser()`가 로그인 후 호출되지 않음
- 사용자의 플랜, 이메일, 이름을 Dashboard에서 알 수 없음
- 수정: 로그인 후 `api.getMe()`를 호출하고 `setUser(result)` 실행

### Dashboard API 통합 우선순위

```
Phase 1 (핵심 -- 인증 & 결제 엔드투엔드 동작):
  1. F-08: 토큰 갱신 메커니즘
  2. F-09: 로그인 시 사용자 프로필 가져오기
  3. F-07: 결제 업그레이드 버튼 연결
  4. F-01: 개요 페이지 실시간 데이터

Phase 2 (코어 -- 주요 기능):
  5. F-02: TanStack Query로 Vault 페이지
  6. F-03: 새 서버 엔드포인트로 Vault 상세
  7. F-06: 필터가 있는 감사 페이지

Phase 3 (팀/디바이스):
  8. F-04: 팀 페이지 전체 통합
  9. F-05: 디바이스 페이지 통합
```

### 필요한 API 클라이언트 메서드 추정

```typescript
// Missing methods to add to apps/dashboard/src/lib/api.ts
class ApiClient {
  // Existing: getMe, listVaults, getSubscription, createCheckout, getPortal, joinWaitlist

  // Vaults (1 new)
  getVault(id: string) { ... }

  // Teams (5 new)
  listTeams() { ... }
  createTeam(name: string, slug: string) { ... }
  inviteTeamMember(teamId: string, userId: string, role: string) { ... }
  removeTeamMember(teamId: string, userId: string) { ... }
  updateMemberRole(teamId: string, userId: string, role: string) { ... }

  // Devices (3 new)
  listDevices() { ... }
  registerDevice(name: string, publicKey?: Uint8Array) { ... }
  deleteDevice(id: string) { ... }

  // Audit (1 new)
  listAuditLogs(filter?: string) { ... }

  // Auth (1 new)
  refreshTokens(refreshToken: string) { ... }
}
// Total: 11 new methods
```

---

## 3. 백엔드 분석 -- API 갭과 플랜 게이팅

**전문가: 백엔드 엔지니어**

### 누락된 API 엔드포인트

**B-01: GET /vaults/:id 엔드포인트 없음**
- Dashboard vault 상세 페이지에 개별 vault 데이터 필요
- 현재 VaultStore 인터페이스에 `GetVault(id, userID)`가 있지만 라우트 없음
- 수정: 새 핸들러와 함께 `authed.GET("/vaults/:id", vaultH.Get)` 추가

**B-02: GET /teams/:id/members 엔드포인트 없음**
- CLI `tene team members`가 "Use dashboard"라고 표시 (team.go 356번째 줄)
- Dashboard 팀 페이지에 멤버 목록 필요
- TeamStore.ListMembers가 존재하지만 HTTP 라우트 없음
- 수정: `authed.GET("/teams/:id/members", teamH.ListMembers)` 추가

**B-03: /auth/me 스텁 외에 사용자 프로필 엔드포인트 없음**
- `/auth/me`가 `{ user_id, plan }`만 반환 (auth.go 265번째 줄)
- 이메일, 이름, 아바타 없음 -- Dashboard 프로필 표시에 필요
- 수정: DB 연결 시 전체 사용자 객체 반환

**B-04: 감사 로그에 페이지네이션과 필터링 없음**
- `AuditHandler.List`가 쿼리 파라미터 없이 `limit=100`을 하드코딩
- 액션 유형, 날짜 범위, vault별 필터 없음
- 수정: 쿼리 파라미터 추가: `?action=push&limit=50&offset=0&after=2026-04-01`

### 플랜 게이팅 불일치

| 동작 | 서버 게이팅 | CLI 게이팅 | Dashboard 게이팅 |
|-----------|:------------:|:----------:|:----------------:|
| Vault 생성 | 402 (Pro) | 사전 확인 없음 | 게이팅 없음 |
| Vault push | 402 (Pro) | 사전 확인 없음 | N/A (CLI 전용) |
| Vault pull | 402 (Pro) | 사전 확인 없음 | N/A (CLI 전용) |
| 팀 생성 | 402 (Pro) | 사전 확인 없음 | 게이팅 없음 |
| 팀 초대 | 402 (Pro) | 사전 확인 없음 | 게이팅 없음 |
| 결제 체크아웃 | Pro이면 차단 | 사전 확인 없음 | 버튼 있지만 핸들러 없음 |
| 디바이스 등록 | 게이팅 없음 | N/A | 게이팅 없음 |
| 감사 로그 | 게이팅 없음 | N/A | 게이팅 없음 |

**B-05: CLI가 클라우드 작업 전에 플랜을 사전 확인하지 않음**
- 무료 사용자가 `tene push`를 실행하면: `"push API: PRO_PLAN_REQUIRED - pro plan required"` 표시
- 더 나은 UX: JWT 클레임을 로컬에서 디코딩하고 표시: "Cloud sync requires Pro plan ($5/month). Run `tene billing upgrade` to subscribe."
- JWT에 이미 `plan` 클레임이 포함 -- API 호출 전에 로컬에서 디코딩

**B-06: 리프레시 토큰 패밀리 추적이 메모리에만 존재**
- PostgreSQL의 `refresh_tokens` 테이블에 `family` 컬럼 없음
- Go 코드가 인메모리 `family` 필드를 사용 (auth.go 33번째 줄)
- PostgreSQL 마이그레이션에 `ALTER TABLE refresh_tokens ADD COLUMN family UUID` 필요

**B-07: 데이터베이스 리포지토리 레이어 구현 없음**
- 모든 핸들러가 인메모리 스토어 사용: `MemVaultStore`, `MemTeamStore` 등
- PostgreSQL 마이그레이션 존재 (8개 파일, 잘 구성됨) 하지만 이를 사용하는 Go 코드 없음
- internal/에 `database/sql`, `pgx`, `sqlx` import 없음 (grep으로 확인)
- 이것이 프로덕션 준비의 가장 큰 단일 블로커

### 서버 측 갭 테이블

| 갭 ID | 설명 | 심각도 | 공수 |
|--------|-------------|----------|--------|
| B-01 | GET /vaults/:id 누락 | 중간 | 0.5일 |
| B-02 | GET /teams/:id/members 누락 | 중간 | 0.5일 |
| B-03 | /auth/me가 최소 데이터만 반환 | 낮음 | 0.5일 |
| B-04 | 감사에 페이지네이션/필터 없음 | 중간 | 1일 |
| B-05 | CLI 플랜 사전 확인 | 중간 | 0.5일 |
| B-06 | DB 스키마에 리프레시 토큰 패밀리 없음 | 낮음 | 0.5일 |
| B-07 | PostgreSQL 리포지토리 레이어 없음 | 심각 | 5-7일 |

---

## 4. 보안 분석 -- 인증 토큰과 Zero-Knowledge

**전문가: 보안 아키텍트**

### 핵심 보안 이슈

**S-01: 인증 토큰이 평문 JSON으로 저장됨 (심각)**
- 파일: `internal/cli/login.go` 151-215번째 줄
- 경로: `~/.tene/auth.json` (`0600` 퍼미션)
- 포함 내용: `access_token` (JWT, 15분) 및 `refresh_token` (30일)
- 위험: 동일 사용자로 실행되는 모든 프로세스가 토큰을 읽을 수 있음. 악성코드, npm/pip 패키지의 공급망 공격, 공유 개발 머신이 현실적인 위협 벡터.
- 수정: 토큰 저장에 `go-keyring` 사용 (`internal/keychain/`에 이미 import됨)
  - macOS: Keychain Access
  - Linux: libsecret (GNOME Keyring)
  - Windows: Credential Manager
  - 폴백: 머신 바인딩 키로 암호화된 파일

**S-02: OAuth 콜백 URL에 토큰 노출 (높음)**
- 파일: `internal/api/handler/auth.go` 177-186번째 줄
- 액세스 토큰과 리프레시 토큰 모두 URL 쿼리 파라미터로 전달
- 위험:
  - 브라우저 히스토리에 토큰이 포함된 전체 URL 저장
  - 서버 액세스 로그에 쿼리 스트링 캡처
  - Referrer 헤더가 다음 내비게이션에 토큰 유출
  - localhost의 CLI 콜백이 더 안전하지만 여전히 로깅됨
- 수정: 인가 코드 흐름 사용:
  1. API가 단기 코드(30초 TTL) 생성
  2. 콜백 URL에 `?code=...`만 포함
  3. Dashboard/CLI가 POST를 통해 코드를 토큰으로 교환

**S-03: Dashboard가 토큰을 localStorage에 저장 (중간)**
- 파일: `apps/dashboard/src/lib/auth-store.ts` -- Zustand이 localStorage에 persist
- 인증 콜백 페이지에서 `tene_access_token` 쿠키도 설정
- XSS 취약점으로 토큰 노출 가능
- 수정: 리프레시 토큰을 httpOnly 쿠키에 저장 (서버가 설정); 액세스 토큰은 메모리에만 유지 (토큰에 대해 persist 없는 Zustand)

**S-04: Dashboard에 리프레시 토큰이 저장되지 않음 (높음)**
- `login()`이 `_refreshToken`을 받음 (언더스코어 접두사 = 미사용)
- 15분 후 액세스 토큰이 만료되면 사용자가 자동 로그아웃됨
- 자동 갱신 메커니즘 없음
- 수정: 리프레시 토큰 저장, 인터셉터 기반 자동 갱신 구현

**S-05: JWT 서명에 ES256 대신 HS256 사용 (낮음)**
- CLAUDE.md에 "ES256 JWT"라고 명시하지만 구현은 `jwt.SigningMethodHS256` 사용 (jwt.go 66번째 줄)
- HS256은 대칭키 (공유 비밀); ES256은 비대칭키 (개인/공개 키 쌍)
- ES256이 선호되는 이유: API가 서명 키를 알지 못해도 검증 가능
- 현재 구현은 기능적이지만 문서가 오해를 일으킴
- 수정: 문서를 HS256으로 업데이트하거나 ES256으로 마이그레이션 (더 많은 작업이 필요하지만 마이크로서비스 아키텍처에 더 적합)

**S-06: 상태 변경 작업에 CSRF 보호 없음 (중간)**
- Dashboard가 Bearer 토큰을 전송하여 일부 CSRF 보호 제공
- 하지만 middleware.ts의 쿠키 기반 인증 (`tene_access_token` 쿠키)은 취약함
- 수정: 쿠키에 `SameSite=Strict` 사용 + 쿠키 기반 인증에 CSRF 토큰 추가

### Zero-Knowledge 무결성 평가

| 주장 | 상태 | 근거 |
|-------|--------|----------|
| 서버가 평문 시크릿을 절대 보지 않음 | 검증됨 | Sync Envelope가 클라이언트 측에서 vault.db를 암호화 (sync/engine.go) |
| 4계층 암호화 | 검증됨 | L1 (XChaCha20), L2 (Sync Envelope), L3 (TLS), L4 (S3 SSE) |
| X25519 ECDH를 통한 팀 키 | 부분 구현됨 | crypto/teamkey.go 동작함; CLI 팀 초대가 미완성 (437, 442번째 줄에 TODO) |
| BIP-39를 통한 복구 | 검증됨 | recovery/mnemonic.go + recover.go |
| 멤버 제거 시 키 로테이션 | 미구현 | 서버가 `key_rotation: true`를 반환하지만 실제로 키를 로테이션하지 않음 |

**S-07: 팀 키 로테이션이 공지되지만 수행되지 않음**
- 파일: `internal/api/handler/team.go` 326-330번째 줄
- 멤버 제거 시 응답이 `"key_rotation": true, "rotation_pending": true`로 표시
- 하지만 실제 키 로테이션 발생하지 않음 (새 프로젝트 키 생성 없음, 재래핑 없음)
- 이는 제거된 멤버가 캐시된 키로 팀 시크릿을 여전히 복호화할 수 있음을 의미
- 수정: 실제 키 로테이션 구현: 새 PK -> 남은 멤버에 대해 재래핑 -> 이전 래핑된 키를 폐기로 표시

---

## 5. QA 분석 -- 엔드투엔드 테스트 전략

**전문가: QA 전략가**

### 테스트 커버리지 평가

**발견된 현재 테스트:**
- `internal/crypto/*_test.go` -- 암호화, KDF, 로테이션, X25519, 제로 와이프 (좋은 커버리지)
- `internal/sync/*_test.go` -- 충돌 해결, 머지, 엔벨로프 (좋은 커버리지)
- `internal/cli/*_test.go` -- CLI 명령어, 색상, 에러
- `internal/auth/jwt_test.go` -- 토큰 생성/검증
- `internal/billing/billing_test.go` -- 결제 서비스
- `internal/errors/errors_test.go` -- 에러 코드

**테스트가 없는 영역:**
- API 핸들러 (handler 패키지에 테스트 파일 없음)
- 미들웨어 (auth, rate limit, RBAC, security)
- API 응답 매핑
- Dashboard (apps/dashboard/에 테스트 파일 없음)
- 랜딩 페이지 (apps/web/에 테스트 파일 없음)
- 엔드투엔드 흐름 (e2e 디렉토리 없음)

### 권장 E2E 테스트 시나리오

**T-01: 인증 흐름 (핵심 경로)**
```
1. GET /api/v1/auth/github/authorize?redirect=dashboard
   -> Verify redirect URL contains state + code_challenge
2. GET /api/v1/auth/github/callback?code=...&state=...
   -> Verify redirect to dashboard with tokens
3. GET /api/v1/auth/me (with Bearer token)
   -> Verify 200 with user_id and plan
4. POST /api/v1/auth/refresh { refresh_token: "..." }
   -> Verify new token pair returned
   -> Verify old refresh token is invalidated (rotation)
5. POST /api/v1/auth/signout { refresh_token: "..." }
   -> Verify token family revoked
```

**T-02: Vault 동기화 흐름 (핵심 경로)**
```
1. POST /api/v1/vaults { project_name: "test" } (Pro user)
   -> Verify 201 with vault ID
2. POST /api/v1/vaults/:id/push (encrypted blob + If-Match)
   -> Verify 200 with version + hash
3. GET /api/v1/vaults/:id/pull
   -> Verify presigned URL returned
4. POST /api/v1/vaults/:id/push with wrong If-Match
   -> Verify 409 VERSION_CONFLICT
5. POST /api/v1/vaults (Free user)
   -> Verify 402 PRO_PLAN_REQUIRED
```

**T-03: 결제 흐름 (매출 경로)**
```
1. POST /api/v1/billing/checkout { email: "..." }
   -> Verify checkout_url returned
2. POST /api/v1/billing/webhook (subscription_created, valid HMAC)
   -> Verify user plan updated to "pro"
3. POST /api/v1/billing/webhook (subscription_cancelled, valid HMAC)
   -> Verify user plan reverted to "free"
4. POST /api/v1/billing/webhook (invalid HMAC)
   -> Verify 400 rejected
```

**T-04: 팀 관리 흐름**
```
1. POST /api/v1/teams (Pro user)
   -> Verify team created with owner as admin
2. POST /api/v1/teams/:id/invite (admin)
   -> Verify member added
3. POST /api/v1/teams/:id/invite (non-admin)
   -> Verify 403 FORBIDDEN
4. DELETE /api/v1/teams/:id/members/:uid
   -> Verify removal + key_rotation flag
5. POST /api/v1/teams (Free user)
   -> Verify 402 PRO_PLAN_REQUIRED
```

**T-05: Dashboard E2E (Playwright)**
```
1. Login flow: /login -> GitHub OAuth mock -> /auth/callback -> /
2. Overview: Verify StatCards show actual counts
3. Vaults: List -> click row -> detail page
4. Billing: Free user -> Upgrade button -> checkout redirect
5. Team: Create -> invite -> member list -> remove
6. Audit: Verify events appear after push/pull operations
```

### 테스트 인프라 권장사항

| 컴포넌트 | 도구 | 우선순위 |
|-----------|------|----------|
| API 통합 테스트 | Go httptest + testify | P0 |
| 핸들러 단위 테스트 | Go httptest + 모의 스토어 | P0 |
| 미들웨어 테스트 | Go httptest | P1 |
| Dashboard 컴포넌트 테스트 | Vitest + @testing-library/react | P1 |
| E2E 테스트 | Playwright | P2 |
| 부하 테스트 | k6 | P3 |

---

## 6. 제품 분석 -- UX 우선순위 매트릭스

**전문가: 프로덕트 매니저**

### 사용자 여정 갭

```
현재 무료 사용자 여정:
  Install CLI -> tene init -> set secrets -> list secrets
  -> login -> push -> "PRO_PLAN_REQUIRED" (막다른 길)
  -> billing upgrade -> opens browser -> checkout -> wait for webhook
  -> push again -> works (maybe, if webhook processed)

이상적 무료 사용자 여정:
  Install CLI -> tene init -> set secrets -> list secrets
  -> push -> "Cloud sync requires Pro. Upgrade? (y/N)"
  -> y -> opens checkout -> completes -> returns to CLI
  -> push -> success
```

### 우선순위 매트릭스

| 이슈 | 사용자 영향 | 매출 영향 | 구현 공수 | 우선순위 |
|-------|:----------:|:-------------:|:--------------------:|:--------:|
| A-02: Webhook 핸들러 panic (nil 스토어) | 표면적 없음 | 모든 업그레이드 차단 | 1일 | P0 |
| S-01: 평문 인증 토큰 | 높음 (보안) | 신뢰 손상 | 2일 | P0 |
| F-08: 토큰 갱신 없음 | 높음 (15분 로그아웃) | Dashboard 이탈 | 1일 | P0 |
| B-07: PostgreSQL 리포지토리 없음 | 프로덕션 차단 | 론칭 차단 | 5-7일 | P0 |
| B-05: CLI 플랜 사전 확인 | 중간 (혼란스러운 에러) | 업그레이드 기회 상실 | 0.5일 | P1 |
| F-07: 결제 버튼 미연결 | 높음 | Dashboard 업그레이드 차단 | 0.5일 | P1 |
| F-01: 정적 개요 페이지 | 중간 | 참여도 감소 | 1일 | P1 |
| A-04: URL에 토큰 | 중간 (보안) | -- | 2일 | P1 |
| F-02: Vault 페이지 데이터 없음 | 중간 | -- | 1일 | P2 |
| F-04: 팀 페이지 API 없음 | 중간 | -- | 2일 | P2 |
| F-05: 디바이스 페이지 API 없음 | 낮음 | -- | 1일 | P2 |
| F-06: 감사 페이지 API 없음 | 낮음 | -- | 1일 | P2 |
| S-07: 팀 키 로테이션 누락 | 중간 (보안) | -- | 3일 | P2 |
| B-01: GET /vaults/:id 누락 | 낮음 | -- | 0.5일 | P3 |
| B-02: GET /teams/:id/members 누락 | 낮음 | -- | 0.5일 | P3 |
| S-05: HS256 vs ES256 불일치 | 낮음 | -- | 2일 | P3 |

---

## 7. 데이터 레이어 분석 -- PostgreSQL 마이그레이션 갭

**전문가: 데이터베이스 아키텍트**

### 현재 상태

PostgreSQL 마이그레이션은 완성도가 높고 잘 설계되어 있음:
- 7개 테이블 + 제약조건을 커버하는 8개 마이그레이션 파일
- 적절한 인덱스 (unique, composite, partial)
- 트리거 기반 `updated_at` 관리
- 감사 로그 파티셔닝 준비 설계
- 외래 키 캐스케이딩 설정됨

### Go 리포지토리 레이어: 완전히 누락

| 인터페이스 | 정의 위치 | Mem 구현 | PG 구현 |
|-----------|-----------|:------------------:|:-----------------:|
| VaultStore | handler/vault.go | MemVaultStore (vault_store_mem.go) | 없음 |
| TeamStore | handler/team.go | MemTeamStore (team.go) | 없음 |
| DeviceStore | handler/device.go | MemDeviceStore (device.go) | 없음 |
| AuditStore | handler/audit.go | MemAuditStore (audit.go) | 없음 |
| WaitlistStore | handler/waitlist.go | MemWaitlistStore | 없음 |
| UserStore (billing) | billing/billing.go | 없음 (nil 전달) | 없음 |

**D-01: go.mod에 데이터베이스 드라이버 없음**
- `pgx`, `database/sql`, `sqlx` 의존성 없음
- 추가 필요: `github.com/jackc/pgx/v5` (Go + PostgreSQL 권장)

**D-02: UserStore 인터페이스가 billing 패키지에 정의됨**
- `UpdatePlan()`과 `GetLemonCustomerID()`만 있음
- 공유 리포지토리 패키지로 확장 및 이동 필요
- 필요처: auth (사용자 upsert), billing (플랜 업데이트), team (멤버 조회)

### 권장 리포지토리 구조

```
internal/repository/          # New package
  repository.go               # Combined interface (all stores)
  postgres/
    postgres.go               # Connection pool, migrations
    user_repo.go              # UserRepository impl
    vault_repo.go             # VaultRepository impl
    team_repo.go              # TeamRepository impl
    device_repo.go            # DeviceRepository impl
    audit_repo.go             # AuditRepository impl
    waitlist_repo.go          # WaitlistRepository impl
    refresh_token_repo.go     # RefreshTokenRepository impl
```

### 스키마 갭

**D-03: refresh_tokens에 family 컬럼 누락**
- Go 코드가 재사용 탐지를 위해 family를 추적 (H-04)
- PostgreSQL 스키마에 `family` 컬럼 없음
- 추가: `ALTER TABLE refresh_tokens ADD COLUMN family UUID;`
- 추가: `CREATE INDEX idx_refresh_tokens_family ON refresh_tokens (family);`

**D-04: 커넥션 풀링 설정 없음**
- ECS Fargate (0.25 vCPU, 512 MiB)는 보수적인 풀 사용 필요: 최대 5-10 커넥션
- RDS db.t4g.micro는 ~100 커넥션 지원
- 설정: `pgx.Pool`에 `MaxConns=10, MinConns=2`

---

## 8. 인프라 분석 -- 프로덕션 준비도

**전문가: 인프라/DevOps 리드**

### 배포 파이프라인 평가

| 컴포넌트 | CI/CD | 상태 |
|-----------|-------|--------|
| Go API -> ECR -> ECS | GitHub Actions | 워크플로우 존재 (.github/workflows/ci.yml) |
| 랜딩 -> Vercel | Git Integration | 자동화됨 |
| Dashboard -> Vercel | Git Integration | 자동화됨 |
| DB 마이그레이션 | 수동 | 자동화된 마이그레이션 러너 없음 |
| 시크릿 동기화 | 수동 | 스크립트 존재하지만 수동 트리거 필요 |

**I-01: CI/CD에 자동화된 데이터베이스 마이그레이션 없음**
- `migrations/` 디렉토리에 마이그레이션 존재
- 마이그레이션 도구 설정 안 됨 (golang-migrate, goose 등)
- ECS 태스크가 앱 시작 전에 마이그레이션을 실행해야 함
- 수정: Dockerfile에 마이그레이션 단계 추가 또는 init 컨테이너 사용

**I-02: Health check 준비도 프로브가 불완전**
- 파일: `internal/api/handler/health.go`
- `Readiness` 핸들러에 DB ping을 위한 TODO
- DB가 다운되어도 ECS가 컨테이너로 트래픽 라우팅
- 수정: 준비도 확인에 PostgreSQL ping 추가

**I-03: Graceful shutdown 처리 없음**
- `cmd/server/main.go`에서 SIGTERM/SIGINT 처리 필요
- ECS가 30초 유예 기간으로 SIGTERM 전송
- 처리 없이는 진행 중인 요청과 DB 커넥션이 끊어질 수 있음

### 누락 컴포넌트의 비용 영향

| 추가 항목 | 월간 비용 영향 |
|----------|:------------------:|
| PostgreSQL 리포지토리 레이어 | $0 (코드만) |
| 마이그레이션 러너 (goose) | $0 (코드만) |
| 세션/레이트리밋용 Redis | +$15 (ElastiCache t4g.micro) |
| CloudWatch 알람 | +$2 |
| WAF 기본 규칙 | +$5 |
| **추가 합계** | **+$22/월** |

---

## 9. CLI 분석 -- 명령어 완성도

**전문가: CLI/DX 엔지니어**

### 명령어 완성도 매트릭스

| 명령어 | 로컬 로직 | API 통합 | 플랜 확인 | 에러 메시지 |
|---------|:-----------:|:---------------:|:----------:|:--------------:|
| init | 완성 | N/A | N/A | 양호 |
| set/get/list/delete | 완성 | N/A | N/A | 양호 (코드 포함) |
| run | 완성 | N/A | N/A | 양호 |
| import/export | 완성 | N/A | N/A | 양호 |
| env | 완성 | N/A | N/A | 양호 |
| passwd | 완성 | N/A | N/A | 양호 |
| recover | 완성 | N/A | N/A | 양호 |
| login | 완성 | 완성 | N/A | 양호 |
| logout | 완성 | N/A | N/A | 양호 |
| push | 완성 | 완성 | 서버 전용 | 미흡 (원시 API 에러) |
| pull | 완성 | 완성 | 서버 전용 | 미흡 (원시 API 에러) |
| sync | 완성 | push/pull에 위임 | 서버 전용 | 보통 (안내 텍스트 있음) |
| billing | 완성 | 완성 | N/A | 양호 |
| billing upgrade | 완성 | 완성 | N/A | 양호 |
| billing portal | 완성 | 완성 | N/A | 양호 |
| team create | 부분 | 완성 | 서버 전용 | 보통 |
| team list | 완성 | 완성 | N/A | 양호 |
| team invite | 미완성 | 완성 | 서버 전용 | 보통 |
| team remove | 완성 | 완성 | N/A | 양호 |
| team members | 미완성 | 미완성 | N/A | 미흡 (하드코딩된 메시지) |
| whoami | 완성 | N/A | N/A | 양호 |
| version | 완성 | N/A | N/A | 양호 |
| update | 완성 | N/A | N/A | 양호 |

### CLI 특정 이슈

**C-01: team invite X25519 키 교환이 미완성**
- `loadOrGenerateX25519()`가 매번 임시 키를 생성 (437번째 줄)
- `fetchUserPublicKey()`가 `"not implemented"` 반환 (446번째 줄)
- `resolveTeamSlug()`가 teamID를 8자로 잘라 slug로 사용 (453번째 줄)
- 결과: 프로젝트 키 래핑이 자동으로 실패하며, 초대는 작동하지만 암호화 없이 진행

**C-02: team members 명령어가 작동하지 않음**
- 멤버 목록 대신 팀 목록을 가져옴 (전용 엔드포인트 없음)
- "Use the dashboard" 메시지로 넘어감
- 수정: GET /teams/:id/members 라우트 추가 (B-02) + CLI 페치

**C-03: push/pull 에러에 `--json` 출력 없음**
- push/pull 실패 시 에러가 일반 텍스트로 반환
- `--json` 플래그가 성공 출력에만 적용
- JSON 에러를 출력하는 다른 명령어들과 불일치

**C-04: sync 명령어가 로컬에서 플랜을 확인하지 않음**
- 로그인하지 않았을 때 마케팅 텍스트 표시 (양호)
- 하지만 무료 플랜으로 로그인 시 push/pull을 자동으로 시도하고 실패
- 수정: JWT 디코딩, 플랜 클레임 확인, 업그레이드 프롬프트 표시

---

## 10. 통합 분석 -- 크로스 컴포넌트 일관성

**전문가: 통합 아키텍트**

### 데이터 흐름 갭

```
Landing Page -----(signup CTA)-----> Dashboard Login
     |                                    |
     | Pro CTA goes to /login,            | OAuth callback stores
     | not to checkout                     | tokens but no refresh
     |                                    |
     v                                    v
LemonSqueezy Checkout <------(no link)------ Dashboard Billing
     |                                    |
     | Webhook hits server                | Upgrade button has
     | but UserStore is nil               | no onClick handler
     |                                    |
     v                                    v
API Server (in-memory) <-------(works)-------- CLI
     |                                    |
     | All data lost on restart           | Auth stored plaintext
     | No PostgreSQL repos                | No plan pre-check
```

### 크로스 컴포넌트 이슈 테이블

| 이슈 | 영향받는 컴포넌트 | 근본 원인 |
|-------|-------------------|------------|
| 결제 Webhook 크래시 | API + LemonSqueezy | UserStore nil (A-02) |
| 토큰 갱신 누락 | Dashboard + API | _refreshToken 미사용 (F-08) |
| 재시작 시 데이터 손실 | API (모든 라우트) | 인메모리 스토어 (A-03) |
| 팀 초대 고장 | CLI + API | fetchUserPublicKey 미구현 (C-01) |
| 플랜 게이팅 UX | CLI + API | 서버 전용 확인 (B-05) |
| Dashboard 비활성 페이지 | Dashboard + API | 클라이언트 메서드 없음 (F-02..F-06) |

### 통합 테스트 갭

전체 흐름을 검증하는 통합 테스트가 없음:
1. CLI 로그인 -> API OAuth -> 토큰 저장
2. CLI push -> API 핸들러 -> S3 업로드 -> 동기화 상태
3. Dashboard 로그인 -> API 콜백 -> 토큰 유지 -> API 호출
4. LemonSqueezy Webhook -> API 핸들러 -> 플랜 업데이트 -> JWT 갱신

---

## 구현 로드맵

### Sprint 1: 기반 (P0 -- 모든 것을 차단)
**예상: 7-8일**

1. **PostgreSQL 리포지토리 레이어** (B-07, D-01)
   - pgx/v5 의존성 추가
   - `internal/repository/postgres/` 패키지 생성
   - 7개 스토어 인터페이스 전체 구현
   - 커넥션 풀과 함께 `api/server.go`에 연결
   
2. **결제 Webhook 크래시 수정** (A-02)
   - billing.NewService()에 실제 UserStore 전달
   - LemonSqueezy 테스트 모드로 Webhook 엔드투엔드 테스트

3. **인증 토큰 Keychain 마이그레이션** (S-01)
   - `~/.tene/auth.json`에서 OS Keychain으로 토큰 이동
   - `go-keyring` 사용 (이미 의존성에 포함)
   - 폴백을 위한 `--no-keychain` 플래그 추가

4. **Dashboard 토큰 갱신** (F-08, F-09, S-04)
   - auth 스토어에 리프레시 토큰 저장
   - 자동 갱신을 위한 API 클라이언트 인터셉터 추가
   - 로그인 시 사용자 프로필 가져오기

### Sprint 2: 핵심 경험 (P1)
**예상: 4-5일**

5. **CLI 플랜 사전 확인** (B-05)
   - API 호출 전에 JWT를 로컬에서 디코딩
   - 무료 사용자에게 친절한 업그레이드 프롬프트 표시
   
6. **결제 업그레이드 버튼 연결** (F-07)
   - onClick을 api.createCheckout()에 연결
   - 성공 리다이렉트와 에러 상태 처리

7. **Dashboard 개요 실시간 데이터** (F-01)
   - vault 수, 이벤트, 디바이스에 대한 TanStack Query 훅

8. **인증 코드 교환 패턴** (A-04, S-02)
   - URL 내 토큰을 단기 인증 코드로 교체
   - Dashboard와 CLI가 POST를 통해 코드 교환

### Sprint 3: 기능 완성 (P2)
**예상: 5-6일**

9. **Vault 페이지 + vault 상세** (F-02, F-03, B-01)
   - GET /vaults/:id 엔드포인트 추가
   - TanStack Query로 Dashboard 페이지 연결

10. **팀 페이지 전체 통합** (F-04, B-02)
    - GET /teams/:id/members 엔드포인트 추가
    - Dashboard에 5개 팀 API 메서드 연결
    - CLI team members 명령어 수정

11. **디바이스 + 감사 페이지** (F-05, F-06, B-04)
    - 디바이스 API 호출 연결
    - 감사 페이지네이션 + 필터링 추가

12. **팀 키 로테이션** (S-07)
    - 멤버 제거 시 실제 로테이션 구현
    - 새 PK 생성, 남은 멤버에 대해 재래핑

### Sprint 4: 마무리 (P3)
**예상: 3-4일**

13. **API 핸들러 테스트** (T-01 ~ T-04)
    - 모의 스토어를 사용한 httptest 기반 핸들러 테스트
    - 최소: 인증 흐름, vault CRUD, 결제 Webhook

14. **데이터베이스 마이그레이션 자동화** (I-01)
    - goose 또는 golang-migrate 추가
    - ECS 태스크 시작 시 실행

15. **문서 정렬** (S-05)
    - HS256 vs ES256 불일치 해결
    - HS256을 유지하는 경우 CLAUDE.md 업데이트

---

## 리스크 평가

| 리스크 | 발생 가능성 | 영향 | 완화 방안 |
|------|:----------:|:------:|------------|
| 프로덕션 데이터 손실 (인메모리 스토어) | 확실 (재시작 시) | 심각 | Sprint 1: PostgreSQL 리포지토리 |
| 결제 Webhook panic으로 모든 업그레이드 차단 | 확실 (Webhook 시) | 심각 | Sprint 1: UserStore 연결 |
| auth.json 파일에서 토큰 유출 | 중간 | 높음 | Sprint 1: Keychain 마이그레이션 |
| Dashboard 세션 15분마다 만료 | 확실 | 높음 | Sprint 1: 토큰 갱신 |
| 팀 멤버 제거 후에도 접근 유지 | 중간 | 높음 | Sprint 3: 키 로테이션 |
| 무료 사용자 push 실패 혼란 | 확실 | 중간 | Sprint 2: 플랜 사전 확인 |

---

## 부록 A: 파일 참조

| 영역 | 주요 파일 |
|------|-----------|
| API 서버 | `internal/api/server.go`, `cmd/server/main.go` |
| 인증 핸들러 | `internal/api/handler/auth.go` |
| 결제 핸들러 | `internal/api/handler/billing.go` |
| 결제 서비스 | `internal/billing/billing.go` |
| Vault 핸들러 | `internal/api/handler/vault.go` |
| 팀 핸들러 | `internal/api/handler/team.go` |
| 디바이스 핸들러 | `internal/api/handler/device.go` |
| 감사 핸들러 | `internal/api/handler/audit.go` |
| JWT 인증 | `internal/auth/jwt.go` |
| 동기화 엔진 | `internal/sync/engine.go` |
| CLI 로그인 | `internal/cli/login.go` |
| CLI Push/Pull | `internal/cli/push.go`, `internal/cli/pull.go` |
| CLI 팀 | `internal/cli/team.go` |
| CLI 결제 | `internal/cli/billing.go` |
| Dashboard API 클라이언트 | `apps/dashboard/src/lib/api.ts` |
| Dashboard 인증 스토어 | `apps/dashboard/src/lib/auth-store.ts` |
| Dashboard 인증 콜백 | `apps/dashboard/src/app/auth/callback/page.tsx` |
| Dashboard 미들웨어 | `apps/dashboard/src/middleware.ts` |
| 랜딩 가격 | `apps/web/src/components/pricing.tsx`, `apps/web/src/data/pricing.ts` |
| 도메인 모델 | `internal/domain/user.go`, `vault.go`, `team.go`, `errors.go` |
| 응답 매퍼 | `internal/api/response/response.go` |
| DB 마이그레이션 | `migrations/000001-000008` |
| 암호화 | `internal/crypto/*.go` |

## 부록 B: 이슈 ID 인덱스

| ID | 제목 | 심각도 | 섹션 |
|----|-------|----------|---------|
| A-01 | 랜딩에서 직접 결제 경로 없음 | 중간 | 아키텍처 |
| A-02 | 결제 UserStore가 nil (panic) | 심각 | 아키텍처 |
| A-03 | 인메모리 스토어 데이터 손실 | 심각 | 아키텍처 |
| A-04 | 콜백 URL에 토큰 노출 | 높음 | 아키텍처 |
| F-01 | 개요 페이지가 정적임 | 중간 | 프론트엔드 |
| F-02 | Vault 페이지 데이터 페칭 없음 | 중간 | 프론트엔드 |
| F-03 | Vault 상세 플레이스홀더 사용 | 중간 | 프론트엔드 |
| F-04 | 팀 페이지 API 통합 없음 | 중간 | 프론트엔드 |
| F-05 | 디바이스 페이지 API 통합 없음 | 낮음 | 프론트엔드 |
| F-06 | 감사 페이지 API 없음 + 필터 고장 | 중간 | 프론트엔드 |
| F-07 | 결제 버튼 onClick 없음 | 높음 | 프론트엔드 |
| F-08 | 토큰 갱신 메커니즘 없음 | 높음 | 프론트엔드 |
| F-09 | 사용자 프로필 미가져옴 | 중간 | 프론트엔드 |
| B-01 | GET /vaults/:id 누락 | 중간 | 백엔드 |
| B-02 | GET /teams/:id/members 누락 | 중간 | 백엔드 |
| B-03 | /auth/me 최소 데이터만 반환 | 낮음 | 백엔드 |
| B-04 | 감사에 페이지네이션/필터 없음 | 중간 | 백엔드 |
| B-05 | CLI 플랜 사전 확인 누락 | 중간 | 백엔드 |
| B-06 | DB에 리프레시 토큰 패밀리 없음 | 낮음 | 백엔드 |
| B-07 | PostgreSQL 리포지토리 레이어 없음 | 심각 | 백엔드 |
| S-01 | 파일에 인증 토큰 평문 저장 | 심각 | 보안 |
| S-02 | OAuth 콜백 URL에 토큰 | 높음 | 보안 |
| S-03 | localStorage에 토큰 (XSS 위험) | 중간 | 보안 |
| S-04 | Dashboard에서 리프레시 토큰 미사용 | 높음 | 보안 |
| S-05 | HS256 vs ES256 불일치 | 낮음 | 보안 |
| S-06 | 쿠키에 CSRF 보호 없음 | 중간 | 보안 |
| S-07 | 팀 키 로테이션 미구현 | 중간 | 보안 |
| D-01 | go.mod에 데이터베이스 드라이버 없음 | 심각 | 데이터 레이어 |
| D-02 | UserStore가 잘못된 패키지에 위치 | 낮음 | 데이터 레이어 |
| D-03 | DB에 family 컬럼 누락 | 낮음 | 데이터 레이어 |
| D-04 | 커넥션 풀 설정 없음 | 낮음 | 데이터 레이어 |
| I-01 | 자동화된 DB 마이그레이션 없음 | 중간 | 인프라 |
| I-02 | 준비도 프로브 불완전 | 중간 | 인프라 |
| I-03 | Graceful shutdown 없음 | 중간 | 인프라 |
| C-01 | 팀 초대 X25519 미완성 | 중간 | CLI |
| C-02 | 팀 멤버 명령어 작동 안 함 | 중간 | CLI |
| C-03 | push 에러에 JSON 출력 없음 | 낮음 | CLI |
| C-04 | sync에 로컬 플랜 확인 없음 | 중간 | CLI |
| T-01-05 | 테스트 스위트 누락 | 중간 | QA |
