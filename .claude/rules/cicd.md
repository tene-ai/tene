# Tene CLI — CI/CD Pipeline Rules

## Workflows

| Workflow | Trigger | Jobs |
|----------|---------|------|
| `ci.yml` | push main, PRs to main | test, lint |
| `release.yml` | push v* tag | GoReleaser → S3 + GitHub Releases |

## CI Rules
- **test**: `go test -race -coverprofile=coverage.out ./...`
- **lint**: `golangci-lint run` (v2.1.6, source-built with Go 1.25)
- NO deploy jobs (API deployment is in tene-cloud)
- NO Docker build (no Dockerfile in this repo)
- NO AWS credentials needed for CI (only for release via OIDC)

## Release Rules
- Tag `v*` triggers GoReleaser
- Builds: darwin/linux x amd64/arm64 + windows/amd64
- S3 upload: `tene-releases` bucket, OIDC auth (no stored keys)
- `LATEST_VERSION` file updated on S3 (cache-control: max-age=60)
- GitHub Releases: dual publish (S3 primary + GitHub secondary)
- Checksums: SHA-256 (`checksums.txt` in each version directory)

## Anti-patterns (NEVER do)

1. NEVER add `deploy-prod` or `deploy-staging` jobs to ci.yml
2. NEVER add ECS/ECR/Docker deployment steps
3. NEVER add `staging` branch trigger to ci.yml
4. NEVER add database migration steps
5. NEVER add Docker build/push steps
6. NEVER store AWS credentials as secrets (use OIDC only)
