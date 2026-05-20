# Changelog

All notable changes to tene will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and tene adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### ⚠️ Breaking

- `tene get <KEY>` now refuses to print plaintext to non-TTY stdout by default.
  This prevents accidental secret leakage into AI agent context windows, log
  aggregators, and shell history files.

  **Migration**

  - If you use `tene get` interactively in a terminal → no change, works as before.
  - If you use `tene get` in a script or pipeline, explicitly opt in:
    - Pass `--unsafe-stdout` on the command line, **or**
    - Set `TENE_ALLOW_STDOUT_SECRETS=1` in the environment, **or**
    - Switch to `tene run -- <command>` (recommended — secrets never touch stdout).
  - `tene get --json` still works on non-TTY but emits a one-line warning on stderr.

### Added

- **`tene permissions`** — print the 3-tier permission table (text + `--json` modes). Shows which commands need a password and which run silently. 26 commands classified as `metaread` / `secretwrite` / `secretread`. Source: sprint cli-ux-permission-model F5.
- **`tene audit tail|show|prune`** — manage the local audit log (sprint F8). `tail [-n N]` shows recent rows; `show --since X --filter Y` queries by time / action prefix; `prune --older-than X` deletes old rows (requires interactive confirm or `--force`). All three accept `--json` for NDJSON output. Auto-deletion never happens — manual prune only (G10 invariant, single DELETE chokepoint enforced by static test).
- **`tene config`** — read / write vault-scoped settings stored in the `vault_meta` table (sprint F1). Keys: `preview.enabled` (bool), `preview.front` (int 0-8), `preview.back` (int 0-8), `audit.warnAtMB` (int 1-1000). `tene config preview.front=N` for N>0 prompts an explicit confirmation because it exposes API key prefixes (sk-, ghp_, AKIA…); `--force` skips the confirm for scripts.
- **`tene migrate fill-previews`** — populate the new preview column for existing v1 vaults (sprint F1). Idempotent — re-running is a no-op after the first pass.
- **Vault DB schema v2** — adds a `secrets.preview` plaintext column storing a substring of each value (default `0 chars from the start + 4 chars from the end`, so a leaked vault.db reveals at most `…aBc1` per row, without the API key prefixes that would identify the issuing service). Hard cap `front + back ≤ 8`. Auto-migration on first open uses `BEGIN IMMEDIATE` + `PRAGMA table_info` to be idempotent and crash-safe.
- **`Vault.ListSecretMetadata(env)`** API returning `[]domain.VaultKeyMeta` with the preview field — `tene list` (sprint F3) now uses this and never touches `encrypted_value`. Bench: `BenchmarkListWithPreview` ≈ 0.09 ms / 100 secrets on Apple M4 Pro (≈165× faster than the previous decrypt-loop path).
- **Threshold notice** — when the audit log size crosses `audit.warnAtMB` (default 50 MB), tene prints a one-line stderr notice and writes a per-project sentinel at `~/.tene/.audit-warned-<dir-hash>` so the same notice does not fire again for 24 h. `--quiet` suppresses entirely. Sentinel write failure does not block the command.
- **Keychain-fallback notice** (sprint F6) — first run on a host without an OS keychain (for example, a headless Linux box without libsecret) prints a one-time stderr notice naming the file-keystore path. Per-project sentinel at `~/.tene/.fallback-warned-<dir-hash>`; `--quiet` suppresses.
- **Declarative permission tier model** — `internal/auth.PermLevel` enum + `auth.CommandTier` map (26 entries) drive a `PersistentPreRunE` dispatcher. Any new cobra command without an entry panics at binary startup (G4), making accidental permission regressions impossible.
- **Audit row per CLI invocation** (sprint F4) — every command writes one `cli.<tier>.<verb>` row alongside the legacy action rows (for example `vault.init`, `secret.write`). The action column carries only the verb path; arguments and secret values are never recorded.
- `tene completion [bash|zsh|fish|powershell]` — generate shell completion scripts.
- `.github/copilot-instructions.md`, `.github/FUNDING.yml`.
- `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md` (GitHub Community Health 42% → 90%+).
- GitHub Issue templates (bug + feature) and PR template.
- Homebrew tap (`brew install agent-kay-it/tene/tene` · pending first release cut).
- Docker image on GHCR (`docker run ghcr.io/agent-kay-it/tene` · pending first release cut).
- Man pages + shell completions bundled in release archives (pending first release cut).
- Landing `/cli` public route with the full CLI reference.
- Visual breadcrumb on `/vs/*` and `/blog/*` pages.
- Trust section on homepage with live GitHub badges + maintainer bio.
- `<link rel="ai-index">` + `/.well-known/ai.json` on the landing site.

### Changed

- **`tene list`** no longer prompts for a master password and no longer decrypts values (sprint F3). It reads the new `secrets.preview` column directly. JSON output: the `preview` key is always a string (never `null`, never absent); empty values render as `""` in JSON and as `-` in the text 3-column `NAME | PREVIEW | UPDATED` layout.
- **`tene set` / `tene import`** derive and store the preview alongside the ciphertext in the same SQL transaction (sprint F1), so the metadata column never drifts out of sync with the encrypted value.
- **`tene init`** output now includes a 3-line hint pointing at `tene permissions` and `tene config preview.enabled=false`, so a first-time user understands which commands need a password and how to disable previews if they prefer.
- `tene --help` now includes a Resources footer with AI index, CLI reference, and docs links.
- Landing `<meta name="description">` shortened to 154 characters (was 211 → exceeded Google SERP cutoff).
- Landing robots.txt explicitly allow-lists 14 LLM crawlers (GPTBot, ClaudeBot, PerplexityBot, ...) and disallows `/.tene/` and `/api/`.
- Homepage JSON-LD now includes Organization + WebSite (SearchAction) alongside existing SoftwareApplication + FAQPage + HowTo.
- Hero sub-copy rewritten to open with "Tene is a local-first encrypted secret manager CLI."
- `install.sh` success message now prints next-step URL (README, CLI ref, llms.txt, Issues).
- README "Cloud Commands" section marked as "Coming soon — currently disabled in the v1.x CLI" to eliminate the documentation-vs-implementation mismatch.

### Removed

- `aggregateRating` field from `/vs/*` structured data. Using GitHub stars as
  `reviewCount` violates Schema.org semantics (reviewCount = reviews, not
  stars) and risks Google rich-result penalties. Will return with genuine
  user reviews in a future release.

### Fixed

- `/vs/*` Schema.org `SoftwareApplication` node now conforms to spec.
- `auto-tag.yml` workflow's `Update LATEST_VERSION` step now runs with
  `if: always()` and a S3 tarball existence check. Previously a
  non-critical GoReleaser failure (e.g. Homebrew formula publish)
  would skip this step and leave `install.sh` users on a stale
  version — v1.0.5, v1.0.6, and v1.0.7 each required a manual S3
  hotfix before the next release.
- Homebrew publishing disabled in `.goreleaser.yml` until the
  `agent-kay-it/homebrew-tene` tap repository and
  `HOMEBREW_TAP_GITHUB_TOKEN` secret are set up. Re-enable
  instructions are preserved inline as a comment block.

### Security

- **Vault DB schema v2 introduces a plaintext `preview` column** (sprint cli-ux-permission-model, Q2 user decision). Default settings (`preview.front=0, preview.back=4`) expose at most the last 4 characters of each secret in `vault.db`, deliberately suppressing API key prefixes (sk-, ghp_, AKIA…) so a leaked vault file cannot be trivially mapped to the issuing service. See `SECURITY.md` § "Vault DB Preview Column (Schema v2)" for the full threat model and opt-out instructions (`tene config preview.enabled=false`).
- **`tene config preview.front=N` for N>0 requires explicit confirmation** (or `--force` for scripts) because it lets API key prefixes appear in `vault.db`. The exact warning text is documented in `SECURITY.md`.
- **Audit log retention: auto-deletion never happens.** Only `tene audit prune` (PermSecretWrite tier, interactive confirm or `--force`) removes rows. The `DELETE FROM audit_log` SQL lives at a single chokepoint (`internal/vault/vault.go::PruneAuditLog`); the invariant is enforced by `TestG10_AuditAutoDeleteProhibition`, which fails the build if any other code path issues that DELETE.
- **`audit_log.action` column never carries arguments or secret values** (sprint F4). It carries only `cli.<tier>.<verb>` (e.g., `cli.secretwrite.set`). `TestAudit_NoArgsLeakedIntoAction` fails the build if this regresses.
