# Dashboard Redesign 상세 설계서
## Tene Cloud Dashboard 전면 개편 — 프로젝트 중심 Zero-Knowledge 보안 관제탑

> **Version**: v1.0
> **Date**: 2026-04-08
> **Feature**: dashboard-redesign
> **Status**: Design Complete — Do Phase Ready
> **Architecture**: Option A — 프로젝트 중심 IA + Pro Paywall
> **Upstream**: Plan (`dashboard-redesign.plan.md`), PRD (`dashboard-redesign.prd.md`)

---

## Context Anchor

| Dimension | Content |
|-----------|---------|
| **WHY** | 대시보드가 CLI 보조 도구에 머물러 독자적 가치 부재. Pro 전환 동력 없음 |
| **WHO** | Pro 결제 사용자: 솔로 개발자, 팀 리드, DevOps/SRE |
| **RISK** | ZK 원칙 위반 없이 키 메타데이터 표시, IA 전면 재설계 70% 재작성 |
| **SUCCESS** | 온보딩 80%+, Free→Pro 15%+, 주간 재방문 40%+ |
| **SCOPE** | Dashboard 전면 재설계 + API 확장 + CLI 유도 + 랜딩 개선 |

---

## 1. 아키텍처 개요

### 1.1 전체 시스템 플로우

```
┌─────────────┐     ┌──────────────┐     ┌──────────────┐
│  tene CLI   │────▶│  Go API      │────▶│ PostgreSQL   │
│  (로컬)     │     │  (Echo v4)   │     │ + S3 (MinIO) │
└──────┬──────┘     └──────┬───────┘     └──────────────┘
       │                   │
       │ push/pull         │ REST API
       │ (binary blob)     │
       │            ┌──────┴───────┐
       │            │  Dashboard   │
       └───────────▶│  (Next.js)   │
        OAuth       │  app.tene.sh │
        redirect    └──────────────┘
```

### 1.2 핵심 설계 결정

#### D-01: 시크릿 키 이름 메타데이터 전송

**결정**: Push 시 **키 이름 + 환경 목록**을 별도 메타데이터로 전송한다.

**근거**:
- 현재 Push는 암호화된 blob만 전송 → 서버는 vault 안의 키 이름을 모름
- Dashboard에서 "환경별 시크릿 키 목록"을 표시하려면 키 이름이 필요
- **값(value)은 절대 전송하지 않음** — 키 이름은 메타데이터이며 ZK 원칙 위반 아님
- Doppler/Infisical 모두 키 이름을 서버에 저장 (값만 암호화)

**구현**:
```go
// Push request에 metadata 필드 추가
POST /api/v1/vaults/:id/push
Content-Type: multipart/form-data

Part 1: "blob" — 암호화된 vault envelope (binary)
Part 2: "metadata" — JSON
{
  "environments": ["default", "staging", "prod"],
  "keys": {
    "default": [
      {"name": "DB_HOST", "version": 1, "updated_at": "2026-04-08T..."},
      {"name": "JWT_SECRET", "version": 2, "updated_at": "2026-04-07T..."}
    ],
    "staging": [
      {"name": "STAGING_VAR", "version": 1, "updated_at": "2026-04-08T..."}
    ]
  },
  "secret_count": 3
}
```

#### D-02: Dashboard = Pro 전용 (Paywall)

**결정**: Free 사용자는 `/upgrade` 페이지만 접근 가능. Dashboard 전체는 Pro 전용.

**구현**: Next.js middleware에서 JWT `plan` 클레임 체크.

#### D-03: 메뉴 IA 재편

**결정**: 4개 메뉴 + 사용자 프로필 드롭다운
```
Projects | Team | Activity | Settings
[avatar ▾] → Profile / Billing / Sign out
```

---

## 2. 데이터 모델 확장

### 2.1 새 테이블: vault_key_metadata

Push 시 전송되는 키 메타데이터를 저장:

```sql
-- Migration: 000011_create_vault_key_metadata.up.sql
CREATE TABLE vault_key_metadata (
    vault_id    UUID NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,
    environment TEXT NOT NULL,
    key_name    TEXT NOT NULL,
    version     INTEGER NOT NULL DEFAULT 1,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (vault_id, environment, key_name)
);

CREATE INDEX idx_vault_key_metadata_vault ON vault_key_metadata(vault_id);
```

### 2.2 새 테이블: onboarding_progress

```sql
-- Migration: 000012_create_onboarding_progress.up.sql
CREATE TABLE onboarding_progress (
    user_id       UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    cli_installed BOOLEAN NOT NULL DEFAULT false,
    first_push    BOOLEAN NOT NULL DEFAULT false,
    second_device BOOLEAN NOT NULL DEFAULT false,
    completed     BOOLEAN NOT NULL DEFAULT false,
    dismissed     BOOLEAN NOT NULL DEFAULT false,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 2.3 기존 테이블 변경 없음

users, vaults, teams, team_members, devices, audit_logs — 모두 현재 스키마 유지.

---

## 3. API 설계

### 3.1 신규 API

#### GET /api/v1/vaults/:id/keys

Vault의 시크릿 키 메타데이터 조회 (값 제외).

```
GET /api/v1/vaults/:id/keys?env=default
Authorization: Bearer {token}

Response 200:
{
  "ok": true,
  "data": {
    "vault_id": "783b8b77-...",
    "environment": "default",
    "keys": [
      {"name": "DB_HOST", "version": 1, "updated_at": "2026-04-08T08:29:04Z"},
      {"name": "JWT_SECRET", "version": 2, "updated_at": "2026-04-07T10:00:00Z"}
    ],
    "environments": ["default", "staging", "prod"]
  }
}
```

- `env` 쿼리 파라미터 없으면 모든 환경의 키 반환
- Pro plan 필수 (middleware)
- vault 소유자 또는 팀 멤버만 접근 가능

#### GET /api/v1/onboarding/status

```
GET /api/v1/onboarding/status
Authorization: Bearer {token}

Response 200:
{
  "ok": true,
  "data": {
    "cli_installed": true,
    "first_push": true,
    "second_device": false,
    "completed": false,
    "dismissed": false
  }
}
```

#### POST /api/v1/onboarding/dismiss

```
POST /api/v1/onboarding/dismiss
Authorization: Bearer {token}

Response 200:
{
  "ok": true,
  "data": {"message": "onboarding dismissed"}
}
```

### 3.2 수정 API

#### POST /api/v1/vaults/:id/push (수정)

기존 binary blob → **multipart/form-data**로 변경:

```
POST /api/v1/vaults/:id/push
Content-Type: multipart/form-data
If-Match: {version}

Part "blob": binary envelope
Part "metadata": JSON {environments, keys, secret_count}
```

**호환성**: metadata 파트가 없으면 기존 동작 유지 (blob만 처리).

#### POST /api/v1/auth/signout (수정)

현재 refresh_token을 body로 받지만, Dashboard에서는 Zustand store에서 가져옴.
변경 없이 기존 API 사용 가능.

### 3.3 기존 API 변경 없음

모든 기존 API(vaults CRUD, teams CRUD, devices CRUD, audit, billing)는 그대로 유지.
Dashboard는 이 API들을 그대로 호출.

---

## 4. 프론트엔드 아키텍처

### 4.1 라우트 구조

```
apps/dashboard/src/app/
├── (auth)/
│   ├── login/page.tsx              # GitHub OAuth 로그인
│   └── upgrade/page.tsx            # Free → Pro 결제 유도 (NEW)
├── auth/
│   └── callback/page.tsx           # OAuth 콜백 처리
├── (dashboard)/
│   ├── layout.tsx                  # Sidebar + UserMenu + CommandPalette
│   ├── page.tsx                    # Projects 목록 (메인, 기존 Overview 대체)
│   ├── projects/
│   │   └── [id]/page.tsx           # Project 상세 (환경 탭 + 키 테이블)
│   ├── team/page.tsx               # 팀 관리 (생성 + 초대 + 역할)
│   ├── activity/page.tsx           # 감사 로그 (필터 + Pagination)
│   └── settings/page.tsx           # Devices + (향후 API Keys)
├── globals.css
└── layout.tsx                      # Root (Geist 폰트, Providers)
```

### 4.2 컴포넌트 구조

```
src/components/
├── layout/
│   ├── sidebar.tsx                 # 4개 메뉴 + 로고
│   ├── mobile-nav.tsx              # 하단 바 (4개 메뉴)
│   ├── user-menu.tsx               # 아바타 드롭다운 (Profile/Billing/Sign out)
│   └── command-palette.tsx         # Cmd+K 커맨드 팔레트
├── projects/
│   ├── project-card.tsx            # 프로젝트 카드 (목록용)
│   ├── env-tabs.tsx                # 환경 탭 (default/staging/prod)
│   ├── secret-key-table.tsx        # 시크릿 키 목록 테이블 (값 비공개)
│   └── project-empty.tsx           # 빈 상태 (첫 push 가이드)
├── team/
│   ├── team-create-form.tsx        # 팀 생성 폼
│   ├── member-table.tsx            # 멤버 테이블
│   └── invite-modal.tsx            # 초대 모달 (email + role + envs)
├── activity/
│   ├── activity-table.tsx          # 감사 로그 테이블
│   └── activity-filters.tsx        # 필터 (action/project/user)
├── settings/
│   └── device-card.tsx             # 디바이스 카드 (온라인/Revoke)
├── onboarding/
│   └── onboarding-checklist.tsx    # 단계별 체크리스트
└── ui/
    ├── badge.tsx                   # 재사용 배지 컴포넌트
    └── empty-state.tsx             # 재사용 빈 상태
```

### 4.3 상태 관리

#### Zustand Store (auth-store.ts) — 수정

```typescript
interface AuthState {
  accessToken: string | null;
  refreshToken: string | null;
  user: {
    id: string;
    plan: "free" | "pro";
    email: string;
    name: string;
    avatar_url: string;
  } | null;
  isAuthenticated: boolean;
  login(accessToken: string, refreshToken: string): void;
  logout(): void;
  setUser(user: AuthState["user"]): void;
}
```

변경점:
- `user`에 `name`, `avatar_url` 추가 (프로필 메뉴용)
- `hydrated` 제거 — `useAuthReady`에서 `!!accessToken`으로 판단

#### TanStack Query Keys

```typescript
const queryKeys = {
  vaults: ["vaults"],
  vault: (id: string) => ["vaults", id],
  vaultKeys: (id: string, env?: string) => ["vault-keys", id, env],
  teams: ["teams"],
  teamMembers: (teamId: string) => ["team-members", teamId],
  devices: ["devices"],
  auditLogs: (filters: AuditFilter) => ["audit-logs", filters],
  onboarding: ["onboarding"],
};
```

### 4.4 API 클라이언트 (api.ts) — 확장

기존 메서드 모두 유지 + 신규 추가:

```typescript
class ApiClient {
  // 기존 유지
  listVaults(): Promise<Vault[]>
  getVault(id: string): Promise<Vault>
  listTeams(): Promise<Team[]>
  createTeam(name: string, slug: string): Promise<Team>
  listTeamMembers(teamId: string): Promise<TeamMember[]>
  inviteTeamMember(teamId: string, email: string, role: string): Promise<TeamMember>
  removeTeamMember(teamId: string, userId: string): Promise<void>
  updateMemberRole(teamId: string, userId: string, role: string): Promise<void>
  listDevices(): Promise<Device[]>
  deleteDevice(id: string): Promise<void>
  listAuditLogs(params?: AuditFilter): Promise<AuditLog[]>
  getSubscription(): Promise<Subscription>
  createCheckout(email: string): Promise<{ checkout_url: string }>
  getPortal(): Promise<{ portal_url: string }>
  getMe(): Promise<UserProfile>
  exchangeAuthCode(code: string): Promise<TokenPair>
  signout(refreshToken?: string): Promise<void>

  // 신규 추가
  getVaultKeys(vaultId: string, env?: string): Promise<VaultKeysResponse>
  getOnboardingStatus(): Promise<OnboardingStatus>
  dismissOnboarding(): Promise<void>
}
```

### 4.5 타입 정의 (types.ts) — 신규 파일

```typescript
// 기존 타입 (api.ts에서 분리)
export interface Vault {
  id: string;
  user_id: string;
  team_id?: string;
  project_name: string;
  s3_key: string;
  vault_version: number;
  vault_hash: string;
  secret_count: number;
  size: number;
  created_at: string;
  updated_at: string;
  last_pushed_at?: string;
}

export interface Team { id: string; name: string; slug: string; owner_id: string; created_at: string; updated_at: string; }
export interface TeamMember { team_id: string; user_id: string; role: "admin" | "member"; env_permissions: string[]; joined_at: string; }
export interface Device { id: string; device_name: string; last_seen_at: string; created_at: string; }
export interface AuditLog { id: string; user_id: string; vault_id?: string; action: string; detail?: string; ip_address?: string; created_at: string; }

// 신규 타입
export interface VaultKeyEntry {
  name: string;
  version: number;
  updated_at: string;
}

export interface VaultKeysResponse {
  vault_id: string;
  environment: string;
  keys: VaultKeyEntry[];
  environments: string[];
}

export interface OnboardingStatus {
  cli_installed: boolean;
  first_push: boolean;
  second_device: boolean;
  completed: boolean;
  dismissed: boolean;
}

export interface UserProfile {
  user_id: string;
  email: string;
  name: string;
  avatar_url: string;
  plan: "free" | "pro";
}

export interface AuditFilter {
  action?: string;
  limit?: number;
  offset?: number;
}
```

---

## 5. 페이지별 상세 설계

### 5.1 Login 페이지 (/login)

**변경 최소**: 현재 로직 유지. 로고 SVG + "Continue with GitHub" + ZK 안내문.

변경점:
- `intent=upgrade` 파라미터 처리 유지
- 로그인 성공 후: `/auth/callback` → plan 체크 → pro면 `/`, free면 `/upgrade`

### 5.2 Upgrade 페이지 (/upgrade) — 신규

**접근**: Free 사용자가 OAuth 로그인 후 자동 리다이렉트.

```
┌─────────────────────────────────────────────────┐
│  [logo] tene cloud                              │
│                                                 │
│  Unlock your Dashboard                          │
│  Manage vaults, team, and audit logs            │
│  with zero-knowledge encryption.                │
│                                                 │
│  ┌─────────────────────────────────────────┐    │
│  │  Pro — $5/month                         │    │
│  │                                         │    │
│  │  ✓ Vault cloud sync (push/pull)         │    │
│  │  ✓ Team secret sharing + RBAC           │    │
│  │  ✓ Full audit log                       │    │
│  │  ✓ Device management                    │    │
│  │  ✓ Dashboard access                     │    │
│  │                                         │    │
│  │  [Upgrade to Pro — $5/month]            │    │
│  └─────────────────────────────────────────┘    │
│                                                 │
│  Already using CLI? That's free forever.        │
│  [← Back to tene.sh]                            │
└─────────────────────────────────────────────────┘
```

**구현 로직**:
```typescript
// upgrade/page.tsx
const user = useAuthStore((s) => s.user);
const handleUpgrade = async () => {
  const { checkout_url } = await api.createCheckout(user.email);
  window.location.href = checkout_url;
};
// 결제 완료 → LemonSqueezy redirect → /billing?success=true
// → webhook → DB plan 업데이트 → 재로그인 시 JWT plan="pro"
```

### 5.3 Auth Callback (/auth/callback) — 수정

현재 로직 + **plan 체크 추가**:

```typescript
// callback/page.tsx
useEffect(() => {
  const code = searchParams.get("code");
  if (code) {
    api.exchangeAuthCode(code).then(({ access_token, refresh_token }) => {
      authStore.login(access_token, refresh_token);
      // /auth/me 호출하여 plan 확인
      api.getMe().then((me) => {
        authStore.setUser(me);
        if (me.plan === "pro") {
          router.replace("/");         // Dashboard 진입
        } else {
          router.replace("/upgrade");  // 결제 유도
        }
      });
    });
  }
}, []);
```

### 5.4 Middleware — 수정

```typescript
// middleware.ts
const publicPaths = ["/login", "/auth/callback", "/upgrade"];

export function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl;

  // Public paths + static assets
  if (publicPaths.some((p) => pathname.startsWith(p))) {
    return NextResponse.next();
  }

  // Auth check
  const token = request.cookies.get("tene_access_token");
  if (!token) {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  // Plan check: decode JWT payload (no verification needed, middleware is for routing only)
  try {
    const payload = JSON.parse(atob(token.value.split(".")[1]));
    if (payload.plan !== "pro") {
      return NextResponse.redirect(new URL("/upgrade", request.url));
    }
  } catch {
    return NextResponse.redirect(new URL("/login", request.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ["/((?!_next/static|_next/image|favicon\\.svg|icon\\.svg|logo\\.svg|api).*)"],
};
```

### 5.5 Dashboard Layout — 재작성

```typescript
// (dashboard)/layout.tsx
export default function DashboardLayout({ children }) {
  return (
    <div className="flex h-screen">
      <Sidebar />
      <main className="flex-1 overflow-y-auto p-6 lg:p-8">
        {children}
      </main>
      <MobileNav />
      <CommandPalette />
    </div>
  );
}
```

### 5.6 Projects 페이지 (/) — 메인

```typescript
// (dashboard)/page.tsx — Projects 목록
export default function ProjectsPage() {
  const authReady = useAuthReady();
  const { data: vaults, isLoading } = useQuery({
    queryKey: ["vaults"],
    queryFn: () => api.listVaults(),
    enabled: authReady,
  });
  const { data: onboarding } = useQuery({
    queryKey: ["onboarding"],
    queryFn: () => api.getOnboardingStatus(),
    enabled: authReady,
  });

  return (
    <div className="space-y-6">
      {/* 온보딩 체크리스트 (미완료 시) */}
      {onboarding && !onboarding.completed && !onboarding.dismissed && (
        <OnboardingChecklist status={onboarding} />
      )}

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Projects</h1>
          <p className="text-muted text-sm mt-1">
            {vaults?.length || 0} synced projects
          </p>
        </div>
      </div>

      {isLoading ? <ProjectsSkeleton /> :
       !vaults?.length ? <ProjectEmpty /> :
       <div className="grid gap-4">
         {vaults.map((v) => <ProjectCard key={v.id} vault={v} />)}
       </div>
      }
    </div>
  );
}
```

### 5.7 Project Detail (/projects/[id]) — 신규

```typescript
// (dashboard)/projects/[id]/page.tsx
export default function ProjectDetailPage({ params }) {
  const { id } = params;
  const authReady = useAuthReady();
  const [activeEnv, setActiveEnv] = useState<string>();

  const { data: vault } = useQuery({
    queryKey: ["vaults", id],
    queryFn: () => api.getVault(id),
    enabled: authReady,
  });

  const { data: keysData } = useQuery({
    queryKey: ["vault-keys", id, activeEnv],
    queryFn: () => api.getVaultKeys(id, activeEnv),
    enabled: authReady && !!vault,
  });

  // 첫 로드 시 environments에서 첫 번째를 activeEnv로 설정
  useEffect(() => {
    if (keysData?.environments?.length && !activeEnv) {
      setActiveEnv(keysData.environments[0]);
    }
  }, [keysData]);

  return (
    <div className="space-y-6">
      {/* Breadcrumb */}
      <nav className="flex items-center gap-2 text-sm text-muted">
        <Link href="/" className="hover:text-foreground">Projects</Link>
        <span>/</span>
        <span className="text-foreground">{vault?.project_name}</span>
      </nav>

      {/* Header: 프로젝트 이름 + 메타데이터 */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold font-mono">{vault?.project_name}</h1>
          <p className="text-muted text-sm mt-1">
            v{vault?.vault_version} · {vault?.secret_count} secrets ·
            {vault?.size ? formatBytes(vault.size) : "0 B"} ·
            synced {vault?.last_pushed_at ? timeAgo(vault.last_pushed_at) : "never"}
          </p>
        </div>
      </div>

      {/* 환경 탭 */}
      {keysData?.environments && (
        <EnvTabs
          environments={keysData.environments}
          active={activeEnv || ""}
          onChange={setActiveEnv}
        />
      )}

      {/* 시크릿 키 테이블 */}
      <SecretKeyTable keys={keysData?.keys || []} />

      {/* ZK 안내 */}
      <div className="rounded-xl border border-dashed border-border p-4 text-center text-xs text-muted">
        <p>Secret values are <strong>never</strong> visible here.</p>
        <p>Use <code className="font-mono text-accent">tene get KEY</code> in your CLI.</p>
        <p className="mt-1 text-muted/60">Zero-Knowledge: XChaCha20-Poly1305 encryption</p>
      </div>
    </div>
  );
}
```

### 5.8 Team 페이지 (/team) — 재작성

```typescript
// (dashboard)/team/page.tsx
export default function TeamPage() {
  const authReady = useAuthReady();
  const queryClient = useQueryClient();

  const { data: teams } = useQuery({
    queryKey: ["teams"],
    queryFn: () => api.listTeams(),
    enabled: authReady,
  });

  const team = teams?.[0]; // 현재 1팀만 지원

  const { data: members } = useQuery({
    queryKey: ["team-members", team?.id],
    queryFn: () => api.listTeamMembers(team!.id),
    enabled: authReady && !!team,
  });

  // 팀이 없으면 생성 폼 표시
  if (!team) return <TeamCreateForm />;

  // 팀이 있으면 멤버 목록 + 초대 버튼
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{team.name}</h1>
          <p className="text-muted text-sm mt-1">{members?.length || 0} members</p>
        </div>
        <InviteButton teamId={team.id} />
      </div>

      <MemberTable members={members || []} teamId={team.id} />

      <div className="rounded-xl border border-dashed border-border p-4 text-xs text-muted text-center">
        Project keys shared via X25519 ECDH — the server never sees plaintext keys.
        Removing a member triggers automatic key rotation.
      </div>
    </div>
  );
}
```

### 5.9 Activity 페이지 (/activity) — 재작성

```typescript
// (dashboard)/activity/page.tsx
export default function ActivityPage() {
  const authReady = useAuthReady();
  const [filters, setFilters] = useState<AuditFilter>({ limit: 50, offset: 0 });

  const { data: logs, isLoading } = useQuery({
    queryKey: ["audit-logs", filters],
    queryFn: () => api.listAuditLogs(filters),
    enabled: authReady,
  });

  const loadMore = () => {
    setFilters((f) => ({ ...f, offset: (f.offset || 0) + 50 }));
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Activity</h1>
        <p className="text-muted text-sm mt-1">Security activity across your account</p>
      </div>

      <ActivityFilters value={filters} onChange={setFilters} />
      <ActivityTable logs={logs || []} isLoading={isLoading} />

      {logs?.length === 50 && (
        <button onClick={loadMore} className="w-full py-2 text-sm text-muted hover:text-foreground">
          Load more...
        </button>
      )}
    </div>
  );
}
```

### 5.10 Settings 페이지 (/settings) — 신규

```typescript
// (dashboard)/settings/page.tsx
export default function SettingsPage() {
  const authReady = useAuthReady();
  const queryClient = useQueryClient();

  const { data: devices } = useQuery({
    queryKey: ["devices"],
    queryFn: () => api.listDevices(),
    enabled: authReady,
  });

  const revokeMutation = useMutation({
    mutationFn: (id: string) => api.deleteDevice(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["devices"] }),
  });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-muted text-sm mt-1">Devices and account settings</p>
      </div>

      <section>
        <h2 className="text-sm font-semibold mb-4">Registered Devices</h2>
        <div className="grid gap-3 sm:grid-cols-2">
          {devices?.map((d) => (
            <DeviceCard key={d.id} device={d} onRevoke={() => revokeMutation.mutate(d.id)} />
          ))}
          {!devices?.length && <EmptyState icon={Monitor} message="No devices registered" hint="Devices are registered when you run tene login" />}
        </div>
      </section>
    </div>
  );
}
```

---

## 6. 주요 컴포넌트 상세

### 6.1 Sidebar

```typescript
const nav = [
  { href: "/", label: "Projects", icon: FolderKey },
  { href: "/team", label: "Team", icon: Users },
  { href: "/activity", label: "Activity", icon: ScrollText },
  { href: "/settings", label: "Settings", icon: Settings },
];
```

하단에 사용자 정보 대신 `<UserMenu />` 배치.

### 6.2 UserMenu (우측 상단 또는 사이드바 하단)

```typescript
// components/layout/user-menu.tsx
export function UserMenu() {
  const user = useAuthStore((s) => s.user);
  const logout = useAuthStore((s) => s.logout);
  const [open, setOpen] = useState(false);

  return (
    <div className="relative">
      <button onClick={() => setOpen(!open)} className="flex items-center gap-2">
        <img src={user?.avatar_url} className="w-7 h-7 rounded-full" />
        <span className="text-sm truncate max-w-[120px]">{user?.name}</span>
        <ChevronDown size={14} />
      </button>

      {open && (
        <div className="absolute bottom-full left-0 mb-2 w-64 rounded-xl border border-border bg-surface p-2 shadow-lg">
          <div className="px-3 py-2 text-xs text-muted">{user?.email}</div>
          <div className="px-3 py-1 text-xs">
            Plan: <span className="text-accent font-medium">{user?.plan}</span>
          </div>
          <div className="border-t border-border my-1" />
          <MenuLink href="/settings" icon={Monitor} label="Devices" />
          <MenuButton icon={CreditCard} label="Billing" onClick={openBillingPortal} />
          <div className="border-t border-border my-1" />
          <MenuButton icon={LogOut} label="Sign out" onClick={handleSignout} />
        </div>
      )}
    </div>
  );
}
```

### 6.3 CommandPalette (Cmd+K)

```typescript
// components/layout/command-palette.tsx
import { Command } from "cmdk";

export function CommandPalette() {
  const [open, setOpen] = useState(false);
  const { data: vaults } = useQuery({ queryKey: ["vaults"], queryFn: () => api.listVaults() });
  const router = useRouter();

  useEffect(() => {
    const handler = (e: KeyboardEvent) => {
      if (e.key === "k" && (e.metaKey || e.ctrlKey)) {
        e.preventDefault();
        setOpen((o) => !o);
      }
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, []);

  return (
    <Command.Dialog open={open} onOpenChange={setOpen}>
      <Command.Input placeholder="Search projects, commands..." />
      <Command.List>
        <Command.Group heading="Projects">
          {vaults?.map((v) => (
            <Command.Item key={v.id} onSelect={() => { router.push(`/projects/${v.id}`); setOpen(false); }}>
              <FolderKey size={14} /> {v.project_name}
            </Command.Item>
          ))}
        </Command.Group>
        <Command.Group heading="Navigation">
          <Command.Item onSelect={() => router.push("/team")}>Go to Team</Command.Item>
          <Command.Item onSelect={() => router.push("/activity")}>Go to Activity</Command.Item>
          <Command.Item onSelect={() => router.push("/settings")}>Go to Settings</Command.Item>
        </Command.Group>
        <Command.Group heading="CLI Commands">
          <Command.Item onSelect={() => copyToClipboard("tene push")}>Copy 'tene push'</Command.Item>
          <Command.Item onSelect={() => copyToClipboard("tene pull")}>Copy 'tene pull'</Command.Item>
        </Command.Group>
      </Command.List>
    </Command.Dialog>
  );
}
```

### 6.4 OnboardingChecklist

```typescript
export function OnboardingChecklist({ status }: { status: OnboardingStatus }) {
  const steps = [
    { done: true, label: "GitHub account connected", hint: null },
    { done: status.cli_installed, label: "Install CLI", hint: "curl -fsSL tene.sh/install.sh | sh" },
    { done: status.first_push, label: "Push your first vault", hint: "cd your-project && tene init && tene push" },
    { done: status.second_device, label: "Connect a second device", hint: "tene login && tene pull", optional: true },
  ];

  return (
    <div className="rounded-xl border border-accent/20 bg-accent/5 p-5">
      <h2 className="font-semibold mb-3">Get started with Tene Cloud</h2>
      <div className="space-y-3">
        {steps.map((step, i) => (
          <div key={i} className="flex items-start gap-3">
            <span className={step.done ? "text-accent" : "text-muted"}>
              {step.done ? "✓" : `${i + 1}`}
            </span>
            <div>
              <span className={step.done ? "line-through text-muted" : ""}>{step.label}</span>
              {step.optional && <span className="text-muted text-xs ml-1">(optional)</span>}
              {!step.done && step.hint && (
                <code className="block mt-1 text-xs font-mono text-accent bg-surface px-2 py-1 rounded">
                  {step.hint}
                </code>
              )}
            </div>
          </div>
        ))}
      </div>
      <button onClick={() => api.dismissOnboarding()} className="mt-3 text-xs text-muted hover:text-foreground">
        Skip onboarding
      </button>
    </div>
  );
}
```

---

## 7. CLI 수정

### 7.1 Push — 키 메타데이터 전송 추가

`internal/cli/push.go` 수정:

```go
// Push 시 vault.db에서 키 이름 + 환경 목록 추출
func extractKeyMetadata(vault *vault.Store) (map[string][]KeyMeta, []string, error) {
    envs, _ := vault.ListEnvironments()
    keys := make(map[string][]KeyMeta)
    for _, env := range envs {
        secrets, _ := vault.ListSecrets(env)
        for _, s := range secrets {
            keys[env] = append(keys[env], KeyMeta{
                Name:      s.Name,
                Version:   s.Version,
                UpdatedAt: s.UpdatedAt,
            })
        }
    }
    return keys, envs, nil
}
```

Push API 호출을 multipart/form-data로 변경:
- Part 1: `blob` — 기존 암호화된 envelope
- Part 2: `metadata` — JSON (environments + keys + secret_count)

### 7.2 Init — Pro 유도 메시지

`internal/cli/init.go` 마지막에 추가:

```go
fmt.Fprintf(cmd.ErrOrStderr(), "\n  Tip: Sync across devices with Tene Cloud ($5/mo)\n")
fmt.Fprintf(cmd.ErrOrStderr(), "       Run: tene login\n")
```

### 7.3 Login — Plan별 메시지 분기

`internal/cli/login.go` 성공 응답에서:

```go
if plan == "pro" {
    fmt.Fprintf(cmd.ErrOrStderr(), "  Dashboard: https://app.tene.sh\n")
    fmt.Fprintf(cmd.ErrOrStderr(), "  Run 'tene push' to sync your vault.\n")
} else {
    fmt.Fprintf(cmd.ErrOrStderr(), "\n  Upgrade to Pro for cloud sync, team sharing, and dashboard.\n")
    fmt.Fprintf(cmd.ErrOrStderr(), "  → https://app.tene.sh/upgrade\n")
}
```

---

## 8. 랜딩 페이지 수정

### 8.1 Nav 추가

`apps/web/src/components/nav.tsx`:
```typescript
// Desktop links에 Dashboard 추가
<a href="https://app.tene.sh" className="...">Dashboard</a>
```

### 8.2 Hero CTA

현재 "Get Started" → 변경:
```html
<a href="https://tene.sh/install.sh">Get Started — Free CLI</a>
<a href="https://app.tene.sh">Dashboard →</a>
```

### 8.3 Pricing CTA

Free 카드: `[Install CLI]` → `tene.sh/install.sh`
Pro 카드: `[Open Dashboard →]` → `app.tene.sh/login?intent=upgrade`

---

## 9. 백엔드 API 구현

### 9.1 새 Handler: VaultKeysHandler

```go
// internal/api/handler/vault_keys.go

type VaultKeyMetadataStore interface {
    GetKeyMetadata(ctx context.Context, vaultID, env string) ([]domain.VaultKeyMeta, error)
    GetEnvironments(ctx context.Context, vaultID string) ([]string, error)
    UpsertKeyMetadata(ctx context.Context, vaultID string, metadata domain.VaultMetadataPayload) error
}

func (h *VaultHandler) ListKeys(c echo.Context) error {
    claims := middleware.GetClaims(c)
    vaultID := c.Param("id")
    env := c.QueryParam("env")

    // vault 소유권 확인
    vault, err := h.store.GetVault(vaultID, claims.UserID)
    if err != nil { return response.Err(c, domain.ErrNotFound) }

    envs, _ := h.keyStore.GetEnvironments(c.Request().Context(), vaultID)
    if env == "" && len(envs) > 0 { env = envs[0] }

    keys, _ := h.keyStore.GetKeyMetadata(c.Request().Context(), vaultID, env)

    return response.OK(c, http.StatusOK, map[string]any{
        "vault_id":     vaultID,
        "environment":  env,
        "keys":         keys,
        "environments": envs,
    })
}
```

### 9.2 Push Handler 수정 — multipart 지원

```go
func (h *VaultHandler) Push(c echo.Context) error {
    // ... 기존 인증/plan 체크 ...

    contentType := c.Request().Header.Get("Content-Type")

    if strings.HasPrefix(contentType, "multipart/form-data") {
        // 새 형식: blob + metadata
        blob, _ := c.FormFile("blob")
        metaField := c.FormValue("metadata")
        // blob 처리 (기존 로직)
        // metadata 파싱 후 vault_key_metadata 테이블 upsert
    } else {
        // 구 형식: raw binary (하위 호환)
        // 기존 로직 그대로
    }
}
```

### 9.3 Onboarding Handler

```go
// internal/api/handler/onboarding.go

func (h *OnboardingHandler) GetStatus(c echo.Context) error {
    claims := middleware.GetClaims(c)

    // DB에서 조회, 없으면 기본값 계산
    status, err := h.store.GetOnboardingProgress(ctx, claims.UserID)
    if err != nil {
        // 자동 계산
        vaultCount, _ := h.vaultStore.CountVaults(ctx, claims.UserID)
        deviceCount, _ := h.deviceStore.CountDevices(ctx, claims.UserID)
        status = &domain.OnboardingProgress{
            CLIInstalled:  deviceCount > 0 || vaultCount > 0,
            FirstPush:     vaultCount > 0,
            SecondDevice:  deviceCount >= 2,
        }
        status.Completed = status.CLIInstalled && status.FirstPush
    }
    return response.OK(c, http.StatusOK, status)
}
```

---

## 10. 디자인 시스템

### 10.1 유지 (변경 없음)

```css
--background: #0a0a0a;
--foreground: #ededed;
--accent: #00ff88;
--accent-dim: #00cc6a;
--surface: #141414;
--surface-2: #1e1e1e;
--border: #2a2a2a;
--muted: #888888;
--danger: #ff4444;
--warning: #ffaa00;
```

폰트: Geist Sans + Geist Mono. 아이콘: Lucide React.

### 10.2 추가 (신규)

```css
/* CommandPalette overlay */
.cmd-overlay { background: rgba(0, 0, 0, 0.6); backdrop-filter: blur(4px); }
.cmd-dialog { background: var(--surface); border: 1px solid var(--border); }

/* Env tabs */
.env-tab { border-bottom: 2px solid transparent; }
.env-tab-active { border-bottom-color: var(--accent); color: var(--accent); }
```

### 10.3 의존성 추가

```json
{
  "cmdk": "^1.0.0"
}
```

---

## 11. Implementation Guide

### 11.1 Phase 1: 기반 (1주)

| # | Task | Files | 의존성 |
|---|------|-------|--------|
| 1-1 | DB 마이그레이션 (vault_key_metadata, onboarding_progress) | `migrations/000011_*.sql`, `000012_*.sql` | 없음 |
| 1-2 | Middleware Plan 체크 + /upgrade 라우트 | `middleware.ts`, `upgrade/page.tsx` | 없음 |
| 1-3 | 사이드바 재작성 (4메뉴 + logo) | `sidebar.tsx`, `mobile-nav.tsx` | 없음 |
| 1-4 | UserMenu 컴포넌트 | `user-menu.tsx` | auth-store user 필드 확장 |
| 1-5 | auth-store 수정 (user 필드 확장 + callback plan 체크) | `auth-store.ts`, `callback/page.tsx` | 1-2 |
| 1-6 | CLI 유도 메시지 (init, login) | `init.go`, `login.go` | 없음 |

### 11.2 Phase 2: 핵심 페이지 (2주)

| # | Task | Files | 의존성 |
|---|------|-------|--------|
| 2-1 | API: GET /vaults/:id/keys | `vault_keys.go`, `vault_key_repo.go` | 1-1 |
| 2-2 | API: Push multipart 지원 | `vault.go` (push handler) | 1-1, 2-1 |
| 2-3 | CLI: Push 키 메타데이터 전송 | `push.go` | 2-2 |
| 2-4 | Projects 목록 페이지 | `page.tsx`, `project-card.tsx` | 없음 |
| 2-5 | Project Detail 페이지 (환경탭 + 키 테이블) | `projects/[id]/page.tsx`, `env-tabs.tsx`, `secret-key-table.tsx` | 2-1 |
| 2-6 | Activity 페이지 (필터 + Pagination) | `activity/page.tsx`, `activity-table.tsx`, `activity-filters.tsx` | 없음 |
| 2-7 | Team 페이지 (생성 + 멤버 관리) | `team/page.tsx`, `team-create-form.tsx`, `member-table.tsx`, `invite-modal.tsx` | 없음 |

### 11.3 Phase 3: 고급 기능 (1주)

| # | Task | Files | 의존성 |
|---|------|-------|--------|
| 3-1 | API: Onboarding status/dismiss | `onboarding.go` | 1-1 |
| 3-2 | 온보딩 체크리스트 UI | `onboarding-checklist.tsx` | 3-1 |
| 3-3 | Cmd+K 커맨드 팔레트 | `command-palette.tsx` | `cmdk` 설치 |
| 3-4 | Settings 페이지 (Devices) | `settings/page.tsx`, `device-card.tsx` | 없음 |
| 3-5 | 랜딩 페이지 개선 (Dashboard 링크 + CTA) | `nav.tsx`, `hero.tsx`, `pricing.tsx` | 없음 |

### 11.4 Phase 4: 검증 (3일)

| # | Task |
|---|------|
| 4-1 | E2E: OAuth → Free → /upgrade → 결제 → Dashboard |
| 4-2 | E2E: CLI push → Dashboard Projects 반영 |
| 4-3 | E2E: Team 생성 → 초대 → 멤버 관리 |
| 4-4 | 모바일 반응형 테스트 |
| 4-5 | Cmd+K 키보드 탐색 테스트 |

### 11.3 Session Guide

| Module | Scope | Tasks |
|--------|-------|:-----:|
| module-1 | 기반 (미들웨어, 사이드바, 인증) | 6 |
| module-2 | API + 핵심 페이지 | 7 |
| module-3 | 고급 기능 | 5 |
| module-4 | 검증 | 5 |

---

## 12. PRD-Plan-Design 교차 검증

### 12.1 Plan 요구사항 충족 확인

| Plan 요구사항 | Design 섹션 | 구현 가능? |
|--------------|------------|:---:|
| Pro Paywall (Free → /upgrade) | §5.2, §5.4 | ✅ middleware + upgrade 페이지 |
| 사용자 프로필 (아바타 + 로그아웃 + Plan) | §6.2 UserMenu | ✅ /auth/me 데이터 활용 |
| Projects 페이지 (환경별 키 목록) | §5.6, §5.7, §9.1 | ✅ 신규 API + vault_key_metadata |
| Activity 필터 + Pagination | §5.9 | ✅ 기존 audit API (offset/limit 지원) |
| Team 대시보드 관리 | §5.8 | ✅ 기존 team API 활용 |
| 온보딩 체크리스트 | §6.4, §9.3 | ✅ 신규 API + DB |
| Cmd+K 커맨드 팔레트 | §6.3 | ✅ cmdk 라이브러리 |
| Settings (Devices) | §5.10 | ✅ 기존 devices API |
| 랜딩 개선 | §8 | ✅ nav/hero/pricing 수정 |
| CLI Pro 유도 | §7 | ✅ init.go/login.go 수정 |

### 12.2 ZK 원칙 위반 여부 확인

| 데이터 | 서버 저장? | ZK 위반? |
|--------|:---:|:---:|
| 시크릿 **값** (plaintext) | ❌ 절대 없음 | ✅ 안전 |
| 시크릿 **키 이름** | ✅ vault_key_metadata | ❌ — 키 이름은 메타데이터, 경쟁사도 동일 |
| 환경 목록 (dev/staging/prod) | ✅ vault_key_metadata | ❌ — 환경 이름은 공개 정보 |
| 암호화된 vault blob | ✅ S3 (SSE-S3) | ✅ — L2 Sync Envelope 유지 |
| 팀 멤버 wrapped key | ✅ DB | ✅ — X25519 ECDH 유지 |

### 12.3 기존 API 호환성 확인

| API | 변경 내용 | 하위 호환? |
|-----|----------|:---:|
| POST /vaults/:id/push | multipart 지원 추가 | ✅ — raw binary도 유지 |
| 기타 모든 API | 변경 없음 | ✅ |

### 12.4 기술 제약 확인

| 제약 | 상태 | 해결 |
|------|:---:|------|
| Push multipart → 50MB limit 유지? | ✅ | Echo bodyLimit은 라우트별 설정됨 |
| JWT plan 클레임 → middleware 디코딩 | ✅ | Base64 디코딩만 (서명 검증 불필요, 라우팅용) |
| vault_key_metadata UPSERT 성능 | ✅ | PRIMARY KEY ON CONFLICT 사용 |
| cmdk 라이브러리 Next.js 15 호환 | ✅ | cmdk v1.0 — React 19 지원 확인 필요 |
