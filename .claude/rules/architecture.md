# Tene Architecture & Clean Code Guide

## Monorepo Structure

```
tene/
├── cmd/tene/              CLI entrypoint (main.go) → goreleaser → GitHub Releases
├── cmd/server/            API server entrypoint (main.go) → Docker → ECR → ECS
├── apps/web/              Landing page (Next.js 15) → Vercel (tene.sh)
├── apps/dashboard/        Dashboard (Next.js 15) → Vercel (app.tene.sh)
├── internal/              Go shared packages (not importable externally)
├── infra/terraform/       AWS infrastructure (12+ modules)
├── migrations/            PostgreSQL schema migrations (000001-000008)
├── scripts/               dev.sh, sync-secrets.sh
├── docs/                  PDCA documents (PM, Plan, Design, Report)
└── .claude/rules/         AI agent context rules
```

## Clean Architecture Layers

```
Domain Layer (no external deps)
  └── internal/domain/        Models: User, Vault, Team, Device, AuditLog
  └── internal/errors/        CLI error codes with recovery suggestions

Infrastructure Layer (external libs only)
  └── internal/crypto/        XChaCha20-Poly1305, Argon2id, HKDF, X25519
  └── internal/encfile/       Encrypted file format (header + ciphertext)
  └── internal/recovery/      BIP-39 mnemonic generation/recovery
  └── internal/keychain/      OS keychain (macOS/Linux) + file fallback
  └── internal/vault/         SQLite CRUD, schema, migrations
  └── internal/config/        CLI + global config management
  └── internal/claudemd/      CLAUDE.md auto-generation (5 AI editors)

Service Layer (business logic)
  └── internal/auth/          OAuth (GitHub PKCE) + JWT (HS256, 15min/30day)
  └── internal/billing/       LemonSqueezy integration + HMAC webhooks
  └── internal/sync/          Push/pull engine, 3-way merge, Sync Envelope

Interface Layer (user-facing)
  └── internal/cli/           Cobra commands (25+ commands)
  └── internal/api/           Echo server, routes, middleware
  └── internal/api/handler/   HTTP handlers (auth, vault, team, billing, etc.)
  └── internal/api/middleware/ JWT auth, rate limit, RBAC, security headers
  └── internal/api/response/  Structured JSON responses
  └── internal/api/storage/   S3 client for vault blobs
```

## Dependency Rules

- `domain/` → imports ONLY stdlib (`errors`, `time`)
- `crypto/`, `vault/`, `keychain/` → imports domain + external libs
- `auth/`, `billing/`, `sync/` → imports crypto + domain
- `api/handler/` → imports domain + auth + billing (via interfaces)
- `cli/` → imports vault, sync, crypto, domain, errors, keychain, recovery, config
- **No circular imports. No `api/` ↔ `cli/` cross-dependency.**

## Interface Segregation

Handlers depend on interfaces, not concrete types:

```go
// handler/vault.go
type VaultStore interface {
    CreateVault(v *domain.Vault) error
    GetVault(id, userID string) (*domain.Vault, error)
    ListVaults(userID string) ([]domain.Vault, error)
    // ...
}

// handler/team.go
type TeamStore interface {
    CreateTeam(t *domain.Team) error
    IsMember(teamID, userID string) bool
    IsAdmin(teamID, userID string) bool
    GetEnvPermissions(teamID, userID string) ([]string, error)
    // ...
}
```

In-memory stores (`MemVaultStore`, `MemTeamStore`) implement these for development.
PostgreSQL stores will implement the same interfaces for production.

## Dependency Injection

All dependencies are injected via constructors in `api/server.go`:

```go
func NewServer(cfg Config) *echo.Echo {
    jwtSvc := auth.NewJWTService(cfg.JWTSecret)
    oauthSvc := auth.NewOAuthService(cfg.GitHubClientID, ...)
    vaultStore := handler.NewMemVaultStore()  // swap for PG in prod
    authH := handler.NewAuthHandler(oauthSvc, jwtSvc)
    vaultH := handler.NewVaultHandler(vaultStore, s3Client)
    // ...
}
```

No global state, no service locators, no `init()` side effects.

## API Routes

```
Public:
  GET  /health                              Liveness check
  GET  /health/ready                        Readiness check
  GET  /api/v1/auth/github/authorize        OAuth initiation (PKCE)
  GET  /api/v1/auth/github/callback         OAuth callback
  POST /api/v1/auth/refresh                 Token rotation
  POST /api/v1/billing/webhook              LemonSqueezy webhook (HMAC)
  POST /api/v1/waitlist                     Waitlist registration

Authenticated (JWT Bearer):
  GET  /api/v1/auth/me                      Current user
  POST /api/v1/auth/signout                 Revoke tokens
  GET  /api/v1/vaults                       List vaults
  POST /api/v1/vaults                       Create vault
  POST /api/v1/vaults/:id/push             Push (50MB limit)
  GET  /api/v1/vaults/:id/pull             Pull (presigned URL)
  DELETE /api/v1/vaults/:id                Delete vault
  GET  /api/v1/billing/subscription         Subscription status
  POST /api/v1/billing/checkout             LemonSqueezy checkout
  POST /api/v1/billing/portal              Customer portal
  POST /api/v1/teams                       Create team
  GET  /api/v1/teams                       List teams
  POST /api/v1/teams/:id/invite            Invite member
  DELETE /api/v1/teams/:id/members/:uid    Remove member
  PATCH /api/v1/teams/:id/members/:uid/role Update role
  POST /api/v1/devices                     Register device
  GET  /api/v1/devices                     List devices
  DELETE /api/v1/devices/:id               Delete device
  GET  /api/v1/audit                       Audit logs
```

## Crypto Architecture

```
Master Password
  → Argon2id (64MB, 3 iter, 4 threads) → Master Key (256-bit)
    → HKDF-SHA256("encryption") → Encryption Key
    → HKDF-SHA256("sync")       → Sync Key
    → HKDF-SHA256("device")     → Device Key
    → HKDF-SHA256("auth")       → Auth Hash

Secret Encryption:
  XChaCha20-Poly1305 (192-bit nonce, key name as AAD)

Sync Envelope (L2):
  Seal: SyncKey + AAD(projectID:env) → XChaCha20-Poly1305(vault.db)
  Magic: "TENV" + version + nonce + ciphertext

Team Key Wrapping:
  X25519 ECDH → shared secret → HKDF(projectID) → wrap Project Key
  Key rotation on member removal → re-wrap for remaining members

Cloud 4-Layer Encryption:
  L1: Secret values (XChaCha20-Poly1305)
  L2: Sync Envelope (entire vault.db)
  L3: TLS in transit
  L4: S3 SSE-S3 (AES-256) at rest
```

## Sync Flow (Push/Pull/Merge)

### Push
1. Read local `vault.db` → Serialize secrets to JSON
2. Seal with Sync Envelope (SyncKey + AAD `projectID:env`)
3. SHA-256 hash for conflict detection
4. Upload to `POST /vaults/:id/push` with `If-Match` (optimistic locking)
5. S3 stores with SSE-S3 (L4)

### Pull
1. `GET /vaults/:id/pull` → presigned S3 URL (5-min TTL)
2. Download + verify SHA-256 hash
3. Open Sync Envelope (SyncKey + AAD)
4. 3-way merge: base (last sync) vs local vs remote
5. Update local vault.db

### 3-Way Merge Rules
```
Base=A, Local=B, Remote=A → Local wins (only local changed)
Base=A, Local=A, Remote=B → Remote wins (only remote changed)
Base=A, Local=B, Remote=C → CONFLICT (both changed)
Base=nil, Local=A         → Local addition
Base=nil, Remote=A        → Remote addition
Base=A, Local=nil, Remote=B → CONFLICT (deleted vs modified)
```

## Auth Flow

### GitHub OAuth (PKCE)
1. Generate PKCE verifier (32 bytes) + S256 challenge
2. Generate random state (16 bytes), store with 5-min TTL
3. Redirect to GitHub `/login/oauth/authorize`
4. Callback: validate state → exchange code+verifier → fetch user
5. Issue access token (15min) + refresh token (30day, SHA-256 hashed)

### Token Refresh (H-04: Family Tracking)
- Each refresh token belongs to a "family"
- On use: old token deleted, new token issued with same family
- Reuse detection: if token not found, revoke entire family

### JWT Claims
```go
Subject: userID, Plan: "free"|"pro", DeviceID, Scope: "user"|"team"
Issuer: "tene", Audience: "tene-api", JTI: unique ID
```

## CLI Commands (25+)

| Command | Description |
|---------|-------------|
| `init [name]` | Create vault + master password + recovery key + CLAUDE.md |
| `set KEY [VALUE]` | Encrypt and store (supports --stdin, --overwrite) |
| `get KEY` | Decrypt and output |
| `list` | List keys (values masked) |
| `delete KEY` | Delete (supports --force) |
| `run -- CMD` | Inject secrets as env vars |
| `import FILE` | Bulk import (.env or encrypted) |
| `export` | Export (.env or --encrypted) |
| `env [name]` | Switch environment |
| `passwd` | Change password (atomic 2-phase re-encrypt) |
| `recover` | Restore via 12-word BIP-39 mnemonic |
| `push` | Upload encrypted vault to cloud |
| `pull` | Download and decrypt remote vault |
| `sync` | Push + Pull combined (Pro) |
| `login` | GitHub OAuth to Tene Cloud |
| `logout` | Sign out + revoke tokens |
| `team create` | Create team + generate project key |
| `team invite` | Invite with X25519 wrapped key |
| `team remove` | Remove + trigger key rotation |
| `team list` | List teams |
| `team members` | List team members |
| `billing` | Subscription status |
| `billing upgrade` | LemonSqueezy checkout |
| `whoami` | Current user info |
| `version` | Version info |
| `update` | Check + install updates |

Global flags: `--json`, `--quiet/-q`, `--env/-e`, `--dir`, `--no-color`, `--no-keychain`

## Database Schema

### PostgreSQL (Cloud) — 8 migrations

| Table | Purpose |
|-------|---------|
| users | OAuth users, plan, public keys |
| refresh_tokens | SHA-256 hashed, family tracking |
| vaults | Cloud vault metadata, version, hash |
| devices | X25519 public keys per device |
| audit_logs | Partitioned by created_at |
| waitlist | Email registration |
| teams | Team entities with slugs |
| team_members | Membership, roles, wrapped keys |

### SQLite (Local) — internal/vault/schema.go

| Table | Purpose |
|-------|---------|
| vault_meta | Key-value metadata (salt, recovery_blob) |
| secrets | Encrypted name/value per environment |
| environments | Environment management |
| audit_log | Local operation history |
