# PDCA Plan — ClawHub Skill for tene CLI

- **Feature**: `clawhub-skill-tene-cli`
- **Branch**: `feature/clawhub-skill-tene-cli` (from `staging`)
- **Date**: 2026-04-22
- **Automation Level**: L4 (full-auto up to publish step)

## 1. Problem Statement

Tene is a local-first encrypted secret manager used heavily by AI coding agents. Today, agents only know how to use tene if the project has a hand-written `CLAUDE.md` or the user types rules manually. There is no portable, single-command way to teach any Claude Code / OpenClaw session the safe usage patterns.

ClawHub (https://clawhub.ai) is the public skill registry for OpenClaw and Claude Code — the "npm for agent skills". Publishing a `tene-cli` skill there gives every AI user free, one-command activation of tene knowledge plus baked-in safety guardrails.

## 2. Success Criteria

| # | Criterion | Verification |
|---|---|---|
| SC1 | `SKILL.md` parses as valid YAML frontmatter | `yq` / ClawHub validator |
| SC2 | Slug `tene-cli` matches `^[a-z0-9][a-z0-9-]*$` | regex check |
| SC3 | Bundle size < 50 MB | `du -sh skills/tene-cli` |
| SC4 | Every `requires.env` / `requires.bins` declared in frontmatter matches actual tene behavior | manual cross-check vs command matrix |
| SC5 | 100% coverage of registered tene commands (active surface only; cloud cmds excluded) | checklist against command matrix |
| SC6 | All safety rules from `CLAUDE.md` + `internal/claudemd/template.go` preserved verbatim | diff against source |
| SC7 | `clawhub skill publish --dry-run` equivalent succeeds locally (if CLI available) | CLI output |
| SC8 | No secrets, `.tene/` contents, or plaintext sample values in the bundle | grep check |
| SC9 | Committed on feature branch; staging/main untouched | `git log --graph` |
| SC10 | Publish command prepared but NOT executed (user confirms before push to registry) | user gate |

## 3. In / Out of Scope

### In Scope (v1.0.0 of skill)
- All **active** tene CLI commands: `init`, `set`, `get`, `list`, `delete`, `run`, `import`, `export`, `env` (+ subcommands), `passwd`, `recover`, `version`, `update`, `whoami`
- Global flags: `--json`, `--quiet/-q`, `--env/-e`, `--dir`, `--no-color`, `--no-keychain`
- Env vars: `TENE_MASTER_PASSWORD`, `TENE_KEYCHAIN_FALLBACK`, `NO_COLOR`, `API_URL` (informational — cloud commands disabled)
- Install via `install.sh` (+ source build mention)
- AI safety rules (`tene get`, `tene export`, `.tene/`, CLI args, master password)
- Error code cheatsheet for top 10 most common errors
- OS: macOS + Linux (Windows works for CLI but install.sh doesn't; note limitation)
- Three real examples: init→set flow, npm start injection, multi-env CI pattern

### Out of Scope (deferred to v1.1+)
- Cloud commands (`login`, `push`, `pull`, `sync`, `billing`, `team`) — **disabled in current tene build** (root.go:80-90)
- Homebrew install — tap not yet configured in `.goreleaser.yml`
- Nix plugin spec — no Nix flake exists yet
- Windows-specific install guide

## 4. ClawHub Requirements (Compliance Matrix)

Based on [openclaw/clawhub/docs/skill-format.md](https://github.com/openclaw/clawhub/blob/main/docs/skill-format.md):

| Requirement | Plan |
|---|---|
| `SKILL.md` present at folder root | ✅ `skills/tene-cli/SKILL.md` |
| YAML frontmatter with `name`, `description` | ✅ both populated; `version: 1.0.0` |
| Text-based files only | ✅ only `.md` |
| Bundle ≤ 50 MB | ✅ ~15 KB |
| Slug regex `^[a-z0-9][a-z0-9-]*$` | ✅ `tene-cli` |
| `metadata.openclaw.requires` matches behavior | ✅ `bins: [tene]`; no `env` (master password is prompted/keychain, not env-required) |
| `metadata.openclaw.install` declared | ✅ `kind: download` pointing to `https://tene.sh/install.sh` |
| MIT-0 license acceptance | ✅ skill itself is docs-only; underlying CLI remains MIT |
| `.clawhubignore` prevents accidents | ✅ blocks `.tene/`, `.env*`, `*.key`, etc. |

## 5. tene CLI Coverage Matrix

### Active commands (must document)
- `init [name]` — vault init + AI editor file generation
- `set KEY [VALUE]` — flags: `--stdin`, `--overwrite`
- `get KEY` — **AI-UNSAFE** (prints plaintext)
- `list` — masked names only, AI-safe
- `delete KEY` — flag: `--force`
- `run -- CMD` — flag `--env/-e` must come before `--`
- `import FILE` — flags: `--overwrite`, `--encrypted`
- `export` — flags: `--file`, `--encrypted`; **AI-UNSAFE** without `--encrypted`
- `env [name | create | delete | list]`
- `passwd` — re-encrypts entire vault
- `recover` — BIP-39 mnemonic restore
- `version` / `update`
- `whoami` — project + keychain status

### Disabled (documented as "cloud/beta, not yet available" only if relevant)
- `login`, `logout`, `push`, `pull`, `sync`, `billing`, `team*`

## 6. Risks

| Risk | Impact | Mitigation |
|---|---|---|
| Frontmatter declares env that isn't actually required → security-analysis mismatch | Skill flagged / featured-denied | Only declare `bins: [tene]`; no `requires.env` |
| Accidentally bundle `.tene/` (encrypted vault) | Serious — leak project binary metadata | `.clawhubignore` + `clawhub inspect --files` before publish |
| `install.sh` URL changes | Install step breaks | Reference stable `https://tene.sh/install.sh`; also document GitHub Releases fallback |
| User expects Homebrew | Install friction | Explicitly note "curl installer" is current; brew pending |
| ClawHub `publish` requires OAuth + shared state | Can't fully automate at L4 | Prepare exact command; require user confirmation |
| Version drift (tene CLI v1.0.3 today; future changes) | Skill guidance goes stale | Skill frontmatter `version: 1.0.0`; bump on tene major releases |

## 7. Rollback

Feature is entirely additive. If anything goes wrong:
```bash
git checkout staging
git branch -D feature/clawhub-skill-tene-cli
# Already published on ClawHub? → clawhub delete tene-cli
```

## 8. Timeline (single session target)

1. Design doc (this session)
2. Implement skill files (this session)
3. Gap check + iterate (this session)
4. Local QA (this session)
5. Commit + prepare publish command (this session)
6. **User-gated**: `clawhub login` + `clawhub skill publish`
