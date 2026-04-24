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

- `tene completion [bash|zsh|fish|powershell]` — generate shell completion scripts.
- `.github/copilot-instructions.md`, `.github/FUNDING.yml`.
- `SECURITY.md`, `CODE_OF_CONDUCT.md`, `CONTRIBUTING.md` (GitHub Community Health 42% → 90%+).
- GitHub Issue templates (bug + feature) and PR template.
- Homebrew tap (`brew install tomo-kay/tene/tene` · pending first release cut).
- Docker image on GHCR (`docker run ghcr.io/tomo-kay/tene` · pending first release cut).
- Man pages + shell completions bundled in release archives (pending first release cut).
- Landing `/cli` public route with the full CLI reference.
- Visual breadcrumb on `/vs/*` and `/blog/*` pages.
- Trust section on homepage with live GitHub badges + maintainer bio.
- `<link rel="ai-index">` + `/.well-known/ai.json` on the landing site.

### Changed

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
  `tomo-kay/homebrew-tene` tap repository and
  `HOMEBREW_TAP_GITHUB_TOKEN` secret are set up. Re-enable
  instructions are preserved inline as a comment block.
