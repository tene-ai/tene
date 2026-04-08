# Dashboard Redesign Plan
## Tene Cloud Dashboard 전면 개편 — 프로젝트 중심 Zero-Knowledge 보안 관제탑

> **Version**: v1.0
> **Date**: 2026-04-08
> **Feature**: dashboard-redesign
> **Status**: Plan Complete — Design Phase Ready
> **Upstream**: `docs/00-pm/dashboard-redesign.prd.md`
> **Method**: Plan Plus (Brainstorming-Enhanced)

---

## Executive Summary

| Perspective | Description |
|-------------|-------------|
| **Problem** | 현재 대시보드는 CLI의 결과 뷰어에 불과. 사용자 프로필/로그아웃 없음, CLI↔Dashboard 단절, Free 사용자 무차별 접근으로 Pro 전환 동력 부재, 6개 동등 메뉴가 사용 빈도와 불일치 |
| **Solution** | 프로젝트 중심 IA(Projects/Team/Activity/Settings) + Pro 전용 Paywall + 사용자 프로필 메뉴 + 온보딩 체크리스트 + Cmd+K 커맨드 팔레트로 전면 재설계. CLI에서 Pro 전환 유도 |
| **Functional UX Effect** | GitHub OAuth → Plan 확인 → Free면 결제 유도 / Pro면 Dashboard 진입 → 프로젝트별 환경/시크릿/팀/감사를 한 곳에서 관리. Cmd+K로 빠른 탐색 |
| **Core Value** | Doppler($10+) 대비 $5로 ZK 암호화 감사 + AI 에이전트 인식 + 프로젝트 중심 관제를 제공하는 개발자 대시보드. 대시보드 자체가 Pro 결제의 이유가 됨 |

---

## Context Anchor

| Dimension | Content |
|-----------|---------|
| **WHY** | 대시보드가 CLI 보조 도구에 머물러 독자적 가치 부재. Pro 전환 동력이 없고, 사용자 기본 기능(프로필/로그아웃) 미구현 |
| **WHO** | Pro 결제 사용자: 솔로 개발자(멀티 디바이스 sync), 팀 리드(권한 관리), DevOps(감사/환경 관리) |
| **RISK** | IA 전면 재설계로 기존 페이지 70% 재작성, 과도한 기능으로 미니멀 정체성 상실, API 변경 필요 |
| **SUCCESS** | 온보딩 완료율 80%+, Free→Pro 전환율 15%+, 대시보드 주간 재방문 40%+ |
| **SCOPE** | Dashboard Next.js 전면 재설계 + CLI 유도 메시지 + 랜딩 페이지 CTA + API 확장(시크릿 키 목록) |

---

## User Intent Discovery (Plan Plus Phase 1)

| Item | Answer |
|------|--------|
| **Core Purpose** | CLI 보조 → 독립적 가치 제공. 대시보드 = Zero-Knowledge 보안 관제탑 |
| **Scope** | 전면 리디자인 — IA + UI + 기능. 기존 코드 70% 이상 재작성 |
| **비즈니스 모델** | 대시보드 = Pro 전용. Free는 CLI만. OAuth → 결제 여부 확인 → 결제 유도 → Pro만 진입 |
| **CLI 유도** | `tene init` 응답에 클라우드 안내, `tene login`에 Pro 유도 + 대시보드 링크 |

---

## Alternatives Explored (Plan Plus Phase 2)

| Approach | Description | Selected |
|----------|-------------|:--------:|
| **A. 프로젝트 중심** | Doppler/Vercel 방식. 프로젝트가 중심, 환경/팀/감사가 프로젝트 내부에 통합 | **Yes** |
| B. 단일 페이지 대시보드 | 모든 정보를 한 페이지에 카드로 배치 | No |
| C. 현재 IA 개선 | 6개 메뉴 유지 + 사용자 프로필 추가 | No |

---

## YAGNI Review (Plan Plus Phase 3)

### Included (전체 포함 — 사용자 요청)
- 사용자 프로필 메뉴 (아바타 + 로그아웃 + Plan + Billing)
- Projects 페이지 (환경별 시크릿 키 목록)
- Activity 페이지 (감사 로그 + 필터링 + Pagination)
- 온보딩 체크리스트 (첫 방문 가이드)
- Cmd+K 커맨드 팔레트
- Team 페이지 (대시보드에서 직접 초대/권한 관리)
- Settings 페이지 (Devices + API Keys)
- 랜딩 페이지 Dashboard 링크 + Free CTA
- CLI Pro 유도 메시지 (`tene init`, `tene login`)
- Pro Paywall (Free 사용자 → 결제 유도)

### Deferred (없음)
사용자 요청: 모든 기능을 완벽하게 기획

---

## 1. 인증 & 결제 플로우 (Paywall)

### 1.1 Dashboard 접근 제어

```
사용자 → app.tene.sh → GitHub OAuth
  │
  ├─ 인증 성공 → /auth/me → plan 확인
  │   ├─ plan: "pro" → Dashboard 메인 진입 ✅
  │   └─ plan: "free" → /upgrade 페이지 (결제 CTA)
  │       ├─ "Upgrade to Pro" → LemonSqueezy checkout
  │       └─ 결제 완료 → webhook → plan 업데이트 → Dashboard 진입
  │
  └─ 인증 실패 → /login 페이지
```

### 1.2 /upgrade 페이지 (Free 사용자 전용)

Free 사용자가 OAuth 로그인 후 보는 전용 페이지:
- Pro 기능 하이라이트 (Vault sync, Team, Audit, Dashboard)
- 가격: $5/month
- "Upgrade to Pro" CTA → LemonSqueezy checkout
- "CLI만 사용하기" 링크 → tene.sh로 이동

### 1.3 CLI Pro 유도

```bash
# tene init 응답 마지막에 추가
  Tip: Sync across devices with Tene Cloud ($5/mo)
       Run: tene login

# tene login 응답 (Free 사용자)
  ✓ Signed in as agent-kay
  Plan: Free

  Upgrade to Pro for cloud sync, team sharing, and dashboard.
  → https://app.tene.sh/upgrade

# tene login 응답 (Pro 사용자)
  ✓ Signed in as agent-kay
  Plan: Pro

  Dashboard: https://app.tene.sh
  Run 'tene push' to sync your vault.
```

---

## 2. 메뉴 & IA 구조

### 2.1 사이드바 네비게이션

```
[logo] tene cloud
─────────────────
◈  Projects        ← 메인. 프로젝트 목록 + 상세
👥 Team            ← 팀 관리 (생성, 초대, 역할)
📋 Activity        ← 감사 로그 (전체)
⚙️  Settings        ← Devices, API Keys, 알림

─────────────────
[avatar] agent-kay ▾
         ├─ k99402802@gmail.com
         ├─ Plan: Pro ✓
         ├─ ─────────
         ├─ Billing
         └─ Sign out
```

### 2.2 모바일 네비게이션

```
하단 바: [Projects] [Team] [Activity] [Settings]
상단 우측: [avatar ▾]
```

---

## 3. Projects 페이지 (핵심)

### 3.1 프로젝트 목록

```
Projects (3)
──────────────────────────────────────────────
┌─────────────────────────────────────────────┐
│ my-backend                                  │
│ 3 environments · 12 secrets · synced 2m ago │
│ [dev] [staging] [prod]                      │
├─────────────────────────────────────────────┤
│ my-frontend                                 │
│ 1 environment · 3 secrets · synced 1h ago   │
│ [default]                                   │
├─────────────────────────────────────────────┤
│ infra-configs                               │
│ 2 environments · 8 secrets · synced 1d ago  │
│ [staging] [prod]                            │
└─────────────────────────────────────────────┘
```

### 3.2 프로젝트 상세

프로젝트 클릭 → 4개 탭:

```
my-backend    [Secrets] [Members] [Activity] [Settings]
──────────────────────────────────────────────────────
환경: [development ▾] [staging] [production]

KEY                     UPDATED          VERSION
DATABASE_URL            2m ago           v3
JWT_SECRET              1d ago           v1
STRIPE_KEY              3d ago           v2
AWS_ACCESS_KEY_ID       1w ago           v1

Secret values are never visible here.
Use 'tene get KEY' in your CLI to access values.

Zero-Knowledge: encrypted with XChaCha20-Poly1305
```

### 3.3 API 확장 필요

| API | 현재 | 추가 필요 |
|-----|------|----------|
| `GET /vaults` | vault 메타데이터만 | ✅ 충분 |
| `GET /vaults/:id` | vault 상세 | ✅ 충분 |
| `GET /vaults/:id/keys` | **없음** | 시크릿 키 이름 + 메타데이터 (값 제외) |
| `GET /vaults/:id/keys?env=` | **없음** | 환경별 키 필터링 |

**Zero-Knowledge 유지**: API는 키 이름, 버전, 업데이트 시간만 반환. 암호화된 값은 push/pull 시에만 전체 vault 단위로 전송.

---

## 4. Team 페이지

### 4.1 팀 생성 (대시보드에서 직접)

```
Team
──────────────────────────────────
No team yet

Create a team to share secrets securely
with X25519 ECDH key wrapping.

Team name: [_______________]
Team slug: [_______________]

[Create Team]
```

### 4.2 팀 관리

```
Team: my-startup (3 members)
──────────────────────────────────────────
USER            ROLE      ENVS          JOINED      
agent-kay       admin     all           2026-04-01  [···]
dev-member      member    dev, staging  2026-04-05  [···]
ops-member      member    prod          2026-04-06  [···]

[+ Invite Member]
```

### 4.3 초대 모달

```
Invite Team Member
──────────────────────
Email:  [_____________________]
Role:   [Member ▾]  (Member / Admin)
Environments: [☑ dev] [☑ staging] [☐ prod]

[Send Invite]
```

---

## 5. Activity 페이지

### 5.1 감사 로그 + 필터링

```
Activity
──────────────────────────────────────────────────
Filters: [All ▾] [All Projects ▾] [All Users ▾] [Date Range ▾]

TIME              ACTION         PROJECT        USER         IP
2m ago            vault.push     my-backend     agent-kay    ::1
15m ago           vault.pull     my-frontend    dev-member   1.2.3.4
1h ago            team.invite    —              agent-kay    ::1
2h ago            vault.create   infra-configs  agent-kay    ::1

[Load more...]
```

### 5.2 Pagination

- 기본 50건 표시
- "Load more" 버튼으로 추가 로딩
- offset/limit 기반 API 쿼리

---

## 6. Settings 페이지

### 6.1 Devices 탭

```
Settings > Devices
──────────────────────────────────────────────
DEVICE             LAST SEEN      ADDED        
Seoul-MBP          2m ago (●)     2026-04-01   [Revoke]
Home-iMac          3h ago (○)     2026-04-03   [Revoke]
CI-Bot             1d ago (○)     2026-04-05   [Revoke]
```

### 6.2 API Keys 탭 (향후)

```
Settings > API Keys
──────────────────────────────────────────────
NAME           CREATED        EXPIRES        LAST USED
ci-deploy      2026-04-01     2026-10-01     2h ago     [Revoke]

[+ Create API Key]
```

---

## 7. 온보딩 체크리스트

### 7.1 첫 방문 시 표시

OAuth 로그인 + Pro 결제 완료 후 첫 대시보드 진입 시:

```
Welcome to Tene Cloud! 🔐

Complete these steps to get started:

✅ 1. GitHub account connected
⬜ 2. Install CLI
   curl -fsSL tene.sh/install.sh | sh
⬜ 3. Push your first vault
   cd your-project && tene init && tene push
⬜ 4. Connect a second device (optional)
   On another device: tene login && tene pull

[Skip onboarding]
```

### 7.2 완료 조건 자동 감지

| Step | 감지 방법 |
|------|----------|
| GitHub 연결 | OAuth 로그인 완료 → 항상 ✅ |
| CLI 설치 | 첫 `tene login` API 호출 시 |
| 첫 push | `vaults` 테이블에 1개 이상 레코드 |
| 두 번째 디바이스 | `devices` 테이블에 2개 이상 레코드 |

---

## 8. Cmd+K 커맨드 팔레트

### 8.1 기능

```
[Cmd+K] 입력창
──────────────────────────────────
🔍 Search projects, commands...

Projects
  → my-backend
  → my-frontend

Commands
  → Go to Team
  → Go to Activity
  → Go to Settings
  → Copy 'tene push' command
  → Copy 'tene pull' command

Actions
  → Invite team member
  → Sign out
```

### 8.2 구현

- `cmdk` 라이브러리 사용 (또는 자체 구현)
- 키보드 네이티브: 화살표 탐색, Enter 실행
- 프로젝트 검색: fuzzy match
- CLI 명령어 복사 기능

---

## 9. 랜딩 페이지 개선

### 9.1 Nav 추가

```
현재: [tene] Features  Pricing  Docs  [GitHub]
개편: [tene] Features  Pricing  Docs  [Dashboard →]  [GitHub]
```

### 9.2 Hero CTA 변경

```
현재: [Get Started]  (어디로 가는지 불명확)
개편: [Get Started — Free CLI]  [Dashboard →]
```

### 9.3 Pricing 카드 CTA

```
Free ($0/forever)      Pro ($5/month)
[Install CLI]          [Start Pro Trial →]
```

---

## 10. 기술 설계 요약

### 10.1 프론트엔드

| 항목 | 현재 | 개편 |
|------|------|------|
| 프레임워크 | Next.js 15 App Router | 유지 |
| 상태 관리 | Zustand + TanStack Query | 유지 |
| 아이콘 | Lucide React | 유지 |
| 새 의존성 | — | `cmdk` (커맨드 팔레트) |
| 페이지 수 | 7개 | 8개 (Projects, Project Detail, Team, Activity, Settings, Upgrade, Login, Callback) |
| 컴포넌트 | 11개 | ~20개 (신규: CommandPalette, UserMenu, OnboardingChecklist, EnvTabs, SecretKeyTable, ProjectCard, ...) |

### 10.2 백엔드 API 추가

| API | Method | 설명 |
|-----|--------|------|
| `/api/v1/vaults/:id/keys` | GET | 시크릿 키 이름 목록 (환경별, 값 제외) |
| `/api/v1/vaults/:id/keys?env=` | GET | 환경 필터링 |
| `/api/v1/onboarding/status` | GET | 온보딩 체크리스트 상태 |

### 10.3 미들웨어 변경

```typescript
// 현재: 인증만 체크
if (!token) → /login

// 개편: 인증 + Plan 체크
if (!token) → /login
if (plan !== "pro") → /upgrade
```

---

## 11. 구현 로드맵

### Phase 1: 기반 (1주)
- [ ] 미들웨어: Plan 체크 + /upgrade 라우트
- [ ] /upgrade 페이지 (결제 CTA)
- [ ] 사용자 프로필 메뉴 (아바타 + 로그아웃 + Plan)
- [ ] 메뉴 IA 재편 (Projects/Team/Activity/Settings)
- [ ] CLI 유도 메시지 수정 (`tene init`, `tene login`)

### Phase 2: 핵심 페이지 (2주)
- [ ] Projects 목록 페이지
- [ ] Project 상세 페이지 (환경 탭 + 시크릿 키 테이블)
- [ ] API: `GET /vaults/:id/keys`
- [ ] Activity 페이지 (필터링 + Pagination)
- [ ] Team 페이지 (생성 + 초대 + 역할 관리)

### Phase 3: 고급 기능 (1주)
- [ ] 온보딩 체크리스트
- [ ] Cmd+K 커맨드 팔레트
- [ ] Settings 페이지 (Devices + API Keys)
- [ ] 랜딩 페이지 개선 (Dashboard 링크 + CTA)

### Phase 4: 검증 (3일)
- [ ] E2E 테스트 (OAuth → 결제 → Dashboard)
- [ ] 모바일 반응형 테스트
- [ ] 성능 최적화

---

## 12. Success Criteria

| # | Criteria | Metric |
|---|----------|--------|
| SC-01 | Pro Paywall 동작 | Free 사용자 → /upgrade 리다이렉트 100% |
| SC-02 | 사용자 프로필 접근 | 1클릭 로그아웃, Plan 상태 표시 |
| SC-03 | Projects 페이지 | 프로젝트별 환경/시크릿 키 표시 |
| SC-04 | Team 대시보드 관리 | 생성, 초대, 역할 변경 CLI 없이 가능 |
| SC-05 | Activity 필터링 | 액션/프로젝트/사용자별 필터 + Pagination |
| SC-06 | 온보딩 완료율 | 80%+ (3단계 중 2단계 이상) |
| SC-07 | Cmd+K 동작 | 프로젝트 검색 + 페이지 이동 + CLI 복사 |
| SC-08 | CLI 유도 | `tene init`, `tene login`에 Pro 안내 |
| SC-09 | 랜딩 개선 | Dashboard 링크 + Free CLI CTA |
| SC-10 | 모바일 반응형 | 4개 메뉴 하단 바 정상 동작 |

---

## 13. Brainstorming Log (Plan Plus)

| Phase | Decision | Rationale |
|-------|----------|-----------|
| Phase 1 | CLI 보조 → 독립적 가치 제공 | 대시보드가 Pro 결제의 이유가 되어야 함 |
| Phase 1 | 전면 리디자인 (70% 재작성) | 점진적 개선으로는 근본적 UX 문제 해결 불가 |
| Phase 2 | 프로젝트 중심 IA 선택 | Doppler/Vercel에서 검증된 패턴, 직관적 계층 |
| Phase 3 | 모든 기능 포함 (YAGNI 없음) | 상용화 전 완벽한 기획 필요 |
| Phase 3 | 대시보드 = Pro 전용 | Free는 CLI, 대시보드 접근이 Pro 전환 동기 |
| Phase 3 | CLI에서 Pro 유도 | `tene init`/`tene login` 응답에 Pro 안내 |
| Phase 4 | 전체 아키텍처 승인 | Paywall + 4메뉴 + 프로젝트 상세 + Cmd+K |
