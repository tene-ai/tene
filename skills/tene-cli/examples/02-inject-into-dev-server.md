# Example 2 — Migrate from `.env` to `tene run --`

**Scenario**: User has an existing project with a `.env` file. They want to
move all secrets into tene and eliminate the plaintext file.

**Input signal**:

> "We've got a `.env` file committed... well, in `.gitignore`, but still on
>  disk on every dev machine. I want to replace it with tene so secrets are
>  encrypted at rest."

## Expected AI behavior

1. Check for `.tene/`. If missing, run `tene init` first.
2. Run `tene import .env --overwrite` to bulk-import (`--overwrite` because
   the user wants tene to be the source of truth after migration).
3. Verify with `tene list` — names should match the `.env` file keys.
4. Delete the plaintext `.env` file.
5. Ensure `.env` is in `.gitignore` and `.env*` is in any local
   `.dockerignore`.
6. Search the codebase for any script that invokes `npm start`, `next dev`,
   `python ...`, `go run ...`, etc. and propose wrapping them with `tene run --`.
7. Update the README and any CI config accordingly.

## Commands executed

```bash
# Initialize if needed
[ -d .tene ] || tene init

# Bulk import from the existing .env
tene import .env --overwrite

# Verify
tene list
# → DATABASE_URL, JWT_SECRET, STRIPE_KEY, ...

# Remove the plaintext file
rm .env

# Ensure .env stays out of source control
grep -q '^\.env$' .gitignore || echo '.env' >> .gitignore
grep -q '^\.env\.\*$' .gitignore || echo '.env.*' >> .gitignore

# Update npm scripts in package.json:
#   "dev":  "tene run -- next dev"
#   "start":"tene run -- next start"
#   "test": "tene run -- jest"

# Confirm the new flow works
tene run -- npm run dev
```

## Suggested `package.json` diff

```diff
 "scripts": {
-  "dev":   "next dev",
-  "build": "next build",
-  "start": "next start",
-  "test":  "jest"
+  "dev":   "tene run -- next dev",
+  "build": "tene run -- next build",
+  "start": "tene run -- next start",
+  "test":  "tene run -- jest"
 }
```

(Note: `tene run --` is only needed for commands that read secrets at
runtime. `next build` typically does not need secrets unless the build uses
`NEXT_PUBLIC_*` values that pull from the environment — leave it wrapped to
be safe.)

## Docker compose pattern

```yaml
services:
  app:
    build: .
    # Inject tene secrets via env_file or through the entrypoint
    entrypoint: ["tene", "run", "--"]
    command: ["node", "server.js"]
```

For Docker, the tene binary must be available inside the container. A
multi-stage Dockerfile can `COPY --from=tene-install /usr/local/bin/tene
/usr/local/bin/tene`. Alternatively, use `tene export --encrypted` to ship an
encrypted backup and decrypt inside the container.

## Unsafe patterns to avoid in this flow

- ❌ Running `tene export` to "verify" the migration preserved values (this
  writes plaintext secrets back to disk — defeats the whole purpose)
- ❌ Leaving the old `.env` file as "backup" — it's still the same plaintext
- ❌ Committing `.env.example` with real values (only placeholder values)
- ❌ Setting secrets in `docker-compose.yml` via `environment:` blocks
  (Docker Compose stores plaintext)
