# Example 3 — Multi-environment secrets and CI/CD pipelines

**Scenario**: User maintains separate secrets for local dev, staging, and
production. They want a clean local workflow plus a GitHub Actions pipeline
that deploys with the right secrets per environment.

**Input signal**:

> "I need different database URLs and API keys for dev, staging, and prod.
>  How do I keep them separated in tene, and how do I use this in CI?"

## Expected AI behavior

1. Explain tene's **environment** concept (one vault, multiple namespaces).
2. Create the three environments with `tene env create`.
3. Populate each env via `tene set --env <name>` (prompted or stdin).
4. Demonstrate local execution with `--env` before `--`.
5. For CI: show the pattern of passing `TENE_MASTER_PASSWORD` + `--no-keychain`
   because GitHub runners have no OS keychain and no TTY.
6. Commit the encrypted vault `.tene/vault.db` ONLY if the team has agreed to
   (it's encrypted with a password the runner gets from GitHub Secrets). Many
   teams prefer pulling the vault from a build artifact or — in future — from
   tene-cloud sync instead.

## Local setup

```bash
# Create named environments
tene env create local
tene env create staging
tene env create prod

# Check they exist
tene env list
#   * default  (0 secrets)  [active]
#     local    (0 secrets)
#     staging  (0 secrets)
#     prod     (0 secrets)

# Populate each env
tene set DATABASE_URL --env local    # prompts
tene set DATABASE_URL --env staging
tene set DATABASE_URL --env prod

# Bulk import from per-env .env files (if they exist)
tene import .env.local --env local --overwrite
tene import .env.prod  --env prod  --overwrite
rm .env.local .env.prod

# Make 'local' the default for dev work
tene env local
```

## Local execution per environment

```bash
# Flag placement: --env BEFORE the `--` separator (tene flag, not child flag)
tene run --env local   -- npm run dev
tene run --env staging -- ./scripts/smoke-test.sh
tene run --env prod    -- ./scripts/deploy.sh
```

## GitHub Actions pattern

GitHub Actions runners have no OS keychain and no interactive TTY, so:

1. Store `TENE_MASTER_PASSWORD` as a repository secret.
2. Pass `--no-keychain` to force file-backed key storage (it'll read the env
   var automatically during the first decrypt).
3. Ensure `.tene/vault.db` is checked in (encrypted, safe to commit if your
   master password is strong). Alternatively, fetch it as a build artifact.

```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    env:
      TENE_MASTER_PASSWORD: ${{ secrets.TENE_MASTER_PASSWORD }}
    steps:
      - uses: actions/checkout@v4
      - name: Install tene
        run: curl -sSfL https://tene.sh/install.sh | sh
      - name: Deploy with prod secrets
        run: tene run --env prod --no-keychain -- ./scripts/deploy.sh
```

### Rotating the GitHub secret

1. User generates a new master password locally: `tene passwd`.
2. This re-encrypts the vault under the new password.
3. Commit the updated `.tene/vault.db`.
4. Update the `TENE_MASTER_PASSWORD` GitHub Secret to match.

## Hardening checklist

- Separate CI master password from the developer master password (use
  `tene passwd` on a clone and maintain two vaults if needed).
- Rotate the master password after removing a team member who had access.
- Use `tene env delete <name>` + recreate to wipe secrets for an environment
  without affecting others.
- Never log the injected environ in CI (e.g. `env | grep KEY`) — GitHub
  Actions masks known secrets but not values it hasn't seen yet.

## Unsafe patterns to avoid in this flow

- ❌ Storing `TENE_MASTER_PASSWORD` in a `.env` file on the runner
- ❌ Omitting `--no-keychain` on GitHub Actions (it'll try to open a keychain
  that doesn't exist and fail with a confusing error)
- ❌ Running `tene export --env prod` in the CI job to "see what secrets are
  available" (leaks plaintext into runner logs)
- ❌ Using the same master password for dev machines and CI runners — a
  compromised CI runner now holds the keys to dev machines too
