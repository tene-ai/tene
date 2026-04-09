# Tene CLI Environments

## Overview

Tene supports multiple environments per vault for isolating secrets.

| Environment | Purpose |
|-------------|---------|
| **default** | Default environment when none specified |
| **local** | Local development secrets |
| **staging** | Staging/QA secrets |
| **prod** | Production secrets |

## CLI Usage

```bash
# Set secret in specific environment
tene set API_KEY "value" --env prod

# List secrets in an environment
tene list --env local

# Run command with secrets injected
tene run --env local -- go run ./cmd/tene

# Switch default environment
tene env staging
```

## Local Development

### Landing Page (apps/web)

| Port | Service |
|:----:|---------|
| 3000 | Landing Page (Next.js 15) |

```bash
cd apps/web && npx next dev
```

## Secret Injection

```bash
# tene injects vault secrets as environment variables
tene run --env local -- <command>
```

### IMPORTANT: `tene run --env` Flag

The `--env` flag in `tene run` is parsed manually before the `--` separator
(due to `DisableFlagParsing: true` for child command passthrough).

```bash
# Correct — uses local environment
tene run --env local -- my-command

# Without --env — uses "default" environment
tene run -- my-command
```

## CI/CD

### GitHub Actions
- **CI** (`ci.yml`): `go test -race` + `golangci-lint` on push to main / PRs
- **Release** (`release.yml`): GoReleaser → S3 + GitHub Releases on `v*` tags
- AWS auth: OIDC (no stored secrets)
