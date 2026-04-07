# Tene Secret Management (Dogfooding)

Tene manages its own secrets using itself. Zero hardcoded credentials in the codebase.

## Secret Flow by Environment

```
tene vault (source of truth)
    │
    ├── local:   tene run --env local -- go run ./cmd/server
    │            → env vars injected directly to process
    │
    ├── staging: tene run --env staging -- ./scripts/sync-secrets.sh
    │            → Tene vault → AWS Secrets Manager → ECS
    │
    └── prod:    tene run --env prod -- ./scripts/sync-secrets.sh
                 → Tene vault → AWS Secrets Manager → ECS
```

## Required Secrets per Environment

### Go API Server (`cmd/server/main.go`)

| Key | Required | Source | Description |
|-----|:--------:|--------|-------------|
| JWT_SECRET | Yes | `envRequired()` | JWT signing (min 32 bytes) |
| GITHUB_CLIENT_ID | No | `envOr("", "")` | GitHub OAuth App |
| GITHUB_CLIENT_SECRET | No | `envOr("", "")` | GitHub OAuth App |
| CALLBACK_BASE | No | default `http://127.0.0.1:8080` | OAuth callback base URL |
| DASHBOARD_URL | No | default `https://app.tene.sh` | Post-checkout redirect |
| S3_BUCKET | No | from Config | Vault blob storage |
| S3_ENDPOINT | No | local MinIO only | S3-compatible endpoint |
| AWS_REGION | No | from Config | AWS region |
| DB_HOST | No | from ECS env | PostgreSQL host |
| DB_PORT | No | default `5432` | PostgreSQL port |
| DB_NAME | No | from ECS env | Database name |
| DB_PASSWORD | No | from Secrets Manager | Database password |
| LEMON_API_KEY | No | from Secrets Manager | LemonSqueezy API |
| LEMON_WEBHOOK_SECRET | No | from Secrets Manager | Webhook HMAC |
| LEMON_STORE_ID | No | from ECS env | Store identifier |
| LEMON_VARIANT_PRO | No | from ECS env | Pro plan variant |

### Frontend (`apps/dashboard/`, `apps/web/`)

| Key | Where Set | Description |
|-----|-----------|-------------|
| NEXT_PUBLIC_API_URL | Vercel env / dev.sh | API endpoint |
| NEXT_PUBLIC_DASHBOARD_URL | dev.sh only | Local dashboard URL |

### Local Development

```bash
# All secrets are in tene vault (local environment)
tene list --env local

# API server gets secrets via tene run
tene run --env local -- go run ./cmd/server

# Frontend gets NEXT_PUBLIC_* via scripts/dev.sh inline
NEXT_PUBLIC_API_URL=http://localhost:8080 npx next dev
```

### IMPORTANT: `tene run --env` Flag

The `--env` flag in `tene run` is parsed manually before the `--` separator
(due to `DisableFlagParsing: true` for child command passthrough).

```bash
# Correct — uses local environment
tene run --env local -- go run ./cmd/server

# Without --env — uses "default" environment (may have wrong values!)
tene run -- go run ./cmd/server
```

## GitHub OAuth App Configuration

| App | Client ID | Callback URL | Usage |
|-----|-----------|--------------|-------|
| Tene (Prod) | `Ov23li3Dnr6...` | `https://api.tene.sh/api/v1/auth/github/callback` | Production |
| Tene Local | `Ov23litON0g...` | `http://localhost:8080/api/v1/auth/github/callback` | Local dev |

**Critical**: Local vault must use the **Local** app's Client ID, not Production.
Store in vault: `tene set GITHUB_CLIENT_ID "local-id" --env local`

## CI/CD Secret Sources

### GitHub Actions → ECS Deploy
- AWS credentials: **OIDC** (no stored secrets) via `tene-prod-github-actions` role
- Docker: builds from source, no secrets baked in

### ECS Task Definition
- **Environment vars**: non-sensitive (PORT, DB_HOST, S3_BUCKET, etc.)
- **Secrets**: sensitive values from AWS Secrets Manager (`DB_PASSWORD`, `JWT_SECRET`, etc.)

### Vercel (Frontend)
- `NEXT_PUBLIC_API_URL` set in Vercel project settings
- No other secrets needed (dashboard is client-side only)

### Terraform
```bash
# Secrets passed via tene + TF_VAR_*
TF_VAR_db_password="$(tene get DB_PASSWORD --env prod)" \
TF_VAR_jwt_secret="$(tene get JWT_SECRET --env prod)" \
terraform plan
```

## CORS Configuration

CORS origins are environment-configurable (no localhost in production):

```go
// Default: production origins only
corsOrigins := []string{"https://tene.sh", "https://app.tene.sh"}

// Local dev: set CORS_EXTRA_ORIGINS env var
// e.g., tene set CORS_EXTRA_ORIGINS "http://localhost:3000,http://localhost:3001" --env local
```

## Security Rules for AI Agents

From `internal/claudemd/template.go`:
1. Never hardcode secrets in source code
2. Never create .env files
3. Access secrets via `tene run -- <command>`
4. Never run `tene get <KEY>` in AI conversations (values may leak)
5. Never run `tene export` (outputs all plaintext to stdout)

## .gitignore Rules

```gitignore
# Build binaries only (not source dirs!)
/tene           # not cmd/tene/
/server         # not cmd/server/
/tene-dev

# Secret-related
.tene/          # local vault directory
.env
.env.*
*.tene.enc
```
