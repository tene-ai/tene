# Open Core — Public Repo Rules (tene)

## This Repo Contains
- `cmd/tene/` — CLI entrypoint
- `pkg/domain/`, `pkg/crypto/`, `pkg/errors/` — Shared packages (importable by tene-cloud)
- `internal/` — CLI-only packages (cli, vault, keychain, sync, config, recovery, claudemd, encfile)
- `apps/web/` — Landing page (tene.sh)

## This Repo Does NOT Contain
- API server code (`cmd/server/`) — lives in tene-cloud
- Dashboard code (`apps/dashboard/`) — lives in tene-cloud
- Terraform/infrastructure (`infra/`) — lives in tene-cloud
- Database migrations (`migrations/`) — lives in tene-cloud
- Docker files — lives in tene-cloud
- OAuth/JWT auth logic — lives in tene-cloud
- Billing/LemonSqueezy — lives in tene-cloud

## Rules

1. **NEVER** create `cmd/server/`, `internal/api/`, `internal/auth/`, `internal/billing/`, `internal/repository/`
2. **NEVER** add echo, pgx, aws-sdk, jwt, oauth2 to go.mod
3. **NEVER** add `deploy-prod` or `deploy-staging` to CI workflows
4. **NEVER** create Dockerfile or docker-compose.yml
5. `pkg/` packages are shared — changes affect tene-cloud. Verify: `go build ./...` in both repos after pkg/ changes.
6. `internal/` packages are CLI-only — safe to modify freely.

## Dependency Direction
```
tene-cloud imports tene/pkg/* (one-way)
tene NEVER imports tene-cloud
```
