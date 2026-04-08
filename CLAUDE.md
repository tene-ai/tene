# Tene -- Agentic Secret Runtime

## Overview

Local-first encrypted secret management platform with Zero-Knowledge cloud sync.
CLI (Go) + Cloud API (Go/Echo) + Dashboard (Next.js) + AWS infrastructure.
AI agents auto-detect secrets via CLAUDE.md generation.

## Rules (detailed guides)

- [Architecture & Clean Code](.claude/rules/architecture.md) — 클린 아키텍처, 패키지 의존성, API 라우트, 암호화 체계, DB 스키마
- [Coding Conventions](.claude/rules/conventions.md) — Go/Frontend 코딩 규칙, 린팅, 디자인 시스템, Git 전략
- [Secret Management](.claude/rules/secrets.md) — tene 시크릿 주입, 환경별 설정, OAuth, CORS, CI/CD
- [Environments & Infrastructure](.claude/rules/environments.md) — local/staging/prod 환경, 포트, 서비스, 네트워크
- [Deployment Guide](.claude/rules/deployment.md) — 배포 절차, 롤백, 모니터링
- [Git Workflow](.claude/rules/git-workflow.md) — 브랜치 전략, PR 규칙, 배포 전 체크리스트, Anti-patterns

## Architecture

- Language: Go 1.25+
- CLI framework: cobra (spf13/cobra)
- API framework: labstack/echo v4
- Local DB: modernc.org/sqlite (pure Go, no CGo)
- Cloud DB: PostgreSQL 16 (RDS)
- Crypto: golang.org/x/crypto (XChaCha20-Poly1305, Argon2id, HKDF, X25519)
- Keychain: zalando/go-keyring (macOS Keychain, Linux libsecret, Windows Credential Vault)
- Recovery: tyler-smith/go-bip39 (12-word BIP-39 mnemonic)
- Storage: AWS S3 (SSE-S3) / MinIO (local)
- Auth: HS256 JWT + GitHub/Google OAuth (PKCE)
- Billing: LemonSqueezy (MoR)
- Dashboard: Next.js 15 + Tailwind CSS v4 + TanStack Query v5
- Infra: Terraform (AWS ECS Fargate, ALB, RDS, S3)
- CI/CD: GitHub Actions (OIDC) → ECR → ECS, Vercel (frontend)
- Build: goreleaser + Homebrew tap
- Test: Go testing + stretchr/testify
- Lint: golangci-lint

## Directory Structure

```
cmd/tene/              CLI entrypoint (main.go)
cmd/server/            Cloud API server entrypoint
internal/crypto/       Argon2id, XChaCha20-Poly1305, HKDF, X25519, team key wrapping
internal/vault/        SQLite vault CRUD, schema, migrations
internal/keychain/     OS keychain integration + file fallback
internal/claudemd/     CLAUDE.md generation and merge
internal/recovery/     BIP-39 mnemonic generation and master key recovery
internal/cli/          Cobra commands (login, push, pull, team, billing, etc.)
internal/api/          Echo server, handlers (auth, vault, team, billing, device, audit)
internal/api/middleware/ JWT auth, rate limit, CORS, security headers, RBAC
internal/api/storage/  S3 client for vault blob storage
internal/auth/         JWT + OAuth services
internal/sync/         Sync Envelope, push/pull engine, conflict resolution, 3-way merge
internal/billing/      LemonSqueezy integration
internal/domain/       Domain models + sentinel errors
internal/config/       CLI + Cloud config management
apps/web/              Next.js landing page (tene.sh, Vercel)
apps/dashboard/        Next.js dashboard (app.tene.sh, Vercel)
infra/terraform/       AWS infrastructure (12 modules)
migrations/            PostgreSQL migrations (7 tables)
scripts/               dev.sh, sync-secrets.sh
docs/                  PDCA documents (PM, Plan, Design, Analysis, Report)
```

## Development

```bash
# Local dev (all services: DB + S3 + API + Dashboard + Landing)
./scripts/dev.sh              # Start everything
./scripts/dev.sh status       # Check status
./scripts/dev.sh stop         # Stop everything

# Build
go build ./cmd/tene           # CLI
go build ./cmd/server          # API server

# Run tests
go test ./...

# Lint
golangci-lint run

# Secret management (dogfooding — tene manages its own secrets)
tene list --env local          # Local dev secrets
tene list --env prod           # Production secrets
tene run --env local -- go run ./cmd/server  # Run with injected secrets

# Sync secrets to AWS (for ECS deployment)
tene run --env prod -- ./scripts/sync-secrets.sh
```

## Key Commands

| Command | Purpose |
|---------|---------|
| `tene init` | Create vault + master password + recovery key + CLAUDE.md |
| `tene set KEY VALUE` | Encrypt and store a secret |
| `tene get KEY` | Decrypt and output a secret |
| `tene run -- CMD` | Inject all secrets as env vars and run command |
| `tene list` | List secret keys (values masked) |
| `tene delete KEY` | Delete a secret |
| `tene import .env` | Bulk import from .env file |
| `tene export` | Export as .env format |
| `tene export --encrypted` | Export encrypted vault backup |
| `tene env [name]` | Switch environment (dev/staging/prod) |
| `tene passwd` | Change master password + re-encrypt vault |
| `tene recover` | Restore master password via 12-word recovery key |
| `tene push` | Encrypt vault with Sync Envelope and upload to cloud |
| `tene pull` | Download and decrypt remote vault |
| `tene sync` | Push + Pull combined (requires Pro plan) |
| `tene login` | OAuth login to Tene Cloud (GitHub) |
| `tene logout` | Sign out and revoke tokens |
| `tene team create` | Create team + generate project key |
| `tene team invite` | Invite member with X25519 key wrapping |
| `tene team remove` | Remove member + trigger key rotation |
| `tene team list` | List teams |
| `tene billing` | View subscription status |
| `tene billing upgrade` | Open LemonSqueezy checkout |

## Key Decisions

- Zero-Knowledge: server never sees plaintext secrets (4-layer encryption)
- XChaCha20-Poly1305 + Argon2id + HKDF + X25519 ECDH (golang.org/x/crypto)
- Sync Envelope: L1 (secret values) + L2 (metadata) + L3 (TLS) + L4 (S3 SSE)
- Team key sharing via X25519 ECDH (no RSA), key rotation on member removal
- LemonSqueezy for billing (MoR, Korean individual account support)
- modernc.org/sqlite for local vault, PostgreSQL for cloud metadata
- goreleaser for multi-platform binaries + Homebrew tap
- Tene dogfooding: Tene manages its own production secrets
- Git branch: main → prod auto-deploy, feature/* → PR → preview

## Coding Conventions

- Go standard naming: camelCase for unexported, PascalCase for exported
- Errors: wrap with `fmt.Errorf("context: %w", err)`, define sentinel errors per package
- All public functions must have godoc comments
- Table-driven tests preferred
- No global state -- pass dependencies via struct fields
- internal/ packages are not importable from outside the module

## Security Model

- Master Password -> Argon2id KDF (64MB, 3 iterations) -> Master Key (256-bit)
- Master Key -> HKDF -> Encryption Key, SyncKey, DeviceKey, AuthHash
- Master Key cached in OS Keychain (go-keyring)
- 192-bit random nonce per encryption, key name as AAD
- Recovery: BIP-39 mnemonic -> Argon2id -> Recovery Key -> decrypt stored Master Key
- Cloud: Zero-Knowledge Sync Envelope (L2 wraps entire vault.db before upload)
- Team: X25519 ECDH shared secret -> HKDF(projectID) -> wrap Project Key per member
- Key rotation: member removal -> new PK -> re-wrap for remaining members

## Project Data

```
Per-project:
  .tene/vault.db       Encrypted secrets + metadata + audit log
  .tene/vault.json     Project metadata
  .tene/.gitignore     Auto-exclude from git
  CLAUDE.md            Auto-generated Claude Code context

Global:
  ~/.tene/config.json  CLI settings, analytics
```

## Frontend Apps

| App | Directory | Domain | Deployment |
|-----|-----------|--------|------------|
| Landing | `apps/web/` | `tene.sh` | Vercel |
| Dashboard | `apps/dashboard/` | `app.tene.sh` | Vercel |

Design system: dark-only (#0a0a0a), accent #00ff88, Geist Sans/Mono, Tailwind CSS v4.
See `apps/web/CLAUDE.md` and `apps/dashboard/CLAUDE.md` for frontend conventions.

## Cloud Infrastructure

AWS account `507221376909` (monsa-sandbox), region `ap-northeast-2`.
Terraform: `infra/terraform/environments/prod/`.
See [.claude/rules/environments.md](.claude/rules/environments.md) for full details.
