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

- **Destructive verbs refuse to proceed on a non-interactive shell
  without `--force`** (sprint v1014-rc1-qa-fixes, FX2, invariant I-12).

  Previously `promptConfirm()` returned `true` whenever stdin was not a
  TTY. That made `tene env delete prod` (and `tene delete KEY`) silently
  succeed in CI/CD pipelines, log-redirected scripts, and AI agent
  contexts — the very places destructive intent should be most carefully
  guarded. QA filed this as B2 (CRITICAL data loss).

  v1.0.14 inverts the default: a non-TTY invocation without `--force` is
  refused with a stderr message and a non-zero exit code. The
  destructive op does not run.

  **Migration**

  - In a terminal, behaviour is unchanged: tene asks `(y/N)` and acts on
    your answer.
  - In automation, add `--force` to any `tene delete KEY` or
    `tene env delete <name>` invocation that *intends* to proceed
    unattended. `tene env delete prod --force` is now a working flag
    (it raised "unknown flag" in rc1, the B9 piece of this fix).

- **`--no-keychain` no longer writes the master key to a shared
  `~/.tene/keyfile`** (sprint v1014-rc1-qa-fixes, FX1, invariant I-11).

  Through v1.0.13 and v1.0.14-rc1, `--no-keychain` quietly fell back to a
  single shared file at `~/.tene/keyfile`. Two projects on the same machine
  using `--no-keychain` overwrote each other's master keys, and the *most
  recently written* key was returned for every subsequent decrypt — which
  meant a wrong `TENE_MASTER_PASSWORD` (or no password at all) still
  decrypted a sibling project's vault. QA report B1 in tene-biz for the
  full reproduction.

  `--no-keychain` now means what its name says: **no persistence anywhere**.
  Every invocation must resolve the master password from
  `TENE_MASTER_PASSWORD` env var or the interactive prompt.

  **Migration**

  - Most users: nothing to do. Drop `--no-keychain` and use the OS keychain
    (the default), or keep `--no-keychain` and ensure `TENE_MASTER_PASSWORD`
    is set in your CI/CD environment.
  - To preserve the v1.0.13 file-backed behaviour, set a path explicitly:
    `export TENE_KEYFILE=$HOME/.tene/keyfile-myproject` (or any path you
    control). The chosen file is created with mode `0600`; you are
    responsible for its location and isolation between projects.
  - If you scripted around the old shared `~/.tene/keyfile` path, switch
    to per-project `TENE_KEYFILE` files. The legacy path itself is left on
    disk so other processes that still rely on it can keep working until
    you remove it.

### Added

- **`auth.IsCobraSynthetic(*cobra.Command) bool`** — exported helper
  that returns true for cobra's auto-generated `help`, `__complete`,
  `__completeNoDesc` verbs. Sprint v1014-rc1-qa-fixes/FX4 promotes
  the previously-private `isCobraInternal` to public so the runtime
  dispatcher in `cli/root.go` can share the predicate with the
  startup-time walker, ensuring static + runtime stay in lockstep.
- **`auth.ValidateStrict(*cobra.Command) error`** — bidirectional
  startup validator. Reports both "missing tier declaration" (forward)
  and "stale entry" (reverse) drifts with distinct prefixes so the
  fix direction is obvious. `root.go init()` now calls this instead
  of `Validate`; the looser `Validate` remains for unit tests with
  synthetic trees that do not need to populate every tier entry.
- **`tene update --include-prerelease`** — opt-in flag for pulling RC/beta
  releases. Without it the update-check path treats the stable channel
  as the only auto-recommendation source, which is what closes the B3
  downgrade vector (sprint v1014-rc1-qa-fixes, FX3, invariant I-13).
- **`shouldOfferUpdate` SemVer-aware helper** — replaces the
  single-character `!=` comparison that drove the B3 RC-to-stable
  downgrade. Uses `golang.org/x/mod/semver.Compare` so the decision
  table is the standard Go community implementation.
- **`TENE_KEYFILE` env var** — explicit opt-in to a file-backed master-key
  store when running with `--no-keychain`. The path is the user's choice;
  the file is created with mode `0600` on first `tene init`. This is the
  documented migration path for the v1.0.13 `--no-keychain` behaviour
  (sprint v1014-rc1-qa-fixes, FX1).
- **`keychain.NullStore`** — internal type representing "no persistent
  storage". Selected automatically when `--no-keychain` is passed without
  a `TENE_KEYFILE` override. Backs invariant I-11.
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

- **B4 (HIGH): `tene help` and `tene help <verb>` returned a dispatch
  error**. The PR #116 permission-tier dispatcher refused to dispatch
  any command without a CommandTier entry, including cobra's synthetic
  `help` verb. The auth-package walker already knew to skip `help`;
  the runtime dispatcher in `cli/root.go` now mirrors that skip via
  the exported `auth.IsCobraSynthetic` predicate. The same fix covers
  `__complete` shell-completion helpers.
- **B5 (HIGH): `tene permissions` listed `logout` as a valid verb but
  `tene logout` returned "unknown command"**. The cloud feature whose
  scope included `logout` was unregistered from `root.go init()` while
  the entry remained in `auth.CommandTier`. The entry has been removed
  and `auth.Validate` was extended to detect reverse drift (a
  CommandTier path with no registered verb) via the new
  `auth.ValidateStrict`. The bidirectional check now panics at binary
  startup if either direction drifts.
- **B3 (CRITICAL): `tene update` recommended a downgrade from RC to older
  stable**. A user on `v1.0.14-rc1` typing `tene update` was
  auto-downgraded to `v1.0.13` because `updateAvailable` was a single-
  character `!=` check that ignored SemVer ordering. The new
  `shouldOfferUpdate(current, latest, includePrerelease)` helper
  centralises the decision and never returns `true` when `latest`
  precedes `current`. The text-mode flow now prints
  "You are on vX, which is newer than the latest stable vY" with
  guidance to use `--include-prerelease` or an explicit version.
- **B2 (CRITICAL): silent destructive ops on non-TTY** — see the ⚠️
  Breaking entry above. The shared `promptConfirm()` helper at
  `internal/cli/helpers.go` now refuses non-TTY invocations without
  `--force`. Both `tene delete KEY` and `tene env delete` surface the
  refusal as a non-zero exit (previously they returned exit 0 + "no
  change"); CI pipelines that ignored the exit code now learn about
  cancelled runs.
- **B9: `tene env delete --force` was an unknown flag** — `envDeleteCmd`
  now declares its own `--force` BoolVar (`envDeleteFlagForce`),
  separate from the single-secret `deleteFlagForce`. The two destructive
  verbs no longer share global state, which the test harness will use
  to assert per-test isolation.
- **B1 (CRITICAL): cross-project key bleed under `--no-keychain`** — see the
  ⚠️ Breaking entry above for the full root cause and migration path. The
  fix lives in `internal/keychain/null_store.go` and `internal/cli/root.go`
  (`selectKeyStore` helper). Regression-pinned by
  `TestNoKeychain_CrossProjectIsolation` in
  `internal/cli/no_keychain_integration_test.go`.
- **`tene init` master-key status message** — was always "Master Key saved
  to OS Keychain" regardless of where the key actually landed. Now reflects
  the real destination: OS Keychain, file path with `TENE_KEYFILE`,
  auto-fallback file (with reason), or "NOT persisted (--no-keychain)" with
  guidance to set `TENE_MASTER_PASSWORD`.
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
