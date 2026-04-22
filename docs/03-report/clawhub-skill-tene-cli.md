# PDCA Completion Report — ClawHub Skill for tene CLI

- **Feature**: `clawhub-skill-tene-cli`
- **Branch**: `feature/clawhub-skill-tene-cli` (from `staging`)
- **Date completed**: 2026-04-22
- **Final match rate**: **97%** (PASS — threshold 90%)
- **Iterations**: 1 (one auto-fix loop after initial gap scan)

## 1. Summary

Designed, implemented, gap-tested, and QA-validated a publishable ClawHub
skill for the tene CLI. All local work done on a feature branch; publish
step is prepared but user-gated (requires interactive GitHub OAuth and
produces shared registry state).

## 2. Deliverables

| Path | Purpose |
|---|---|
| `docs/01-plan/clawhub-skill-tene-cli.md` | Plan: scope, success criteria, risks |
| `docs/02-design/clawhub-skill-tene-cli.md` | Design: folder layout, SKILL.md structure |
| `docs/03-report/clawhub-skill-tene-cli.md` | This report |
| `docs/research/clawhub-skill-publishing.md` | ClawHub deep-dive (from prior turn) |
| `skills/tene-cli/SKILL.md` | Agent-facing instructions + safety rules |
| `skills/tene-cli/README.md` | Marketplace-facing description |
| `skills/tene-cli/.clawhubignore` | Publish exclusion patterns |
| `skills/tene-cli/examples/01-init-and-first-secret.md` | Fresh project init example |
| `skills/tene-cli/examples/02-inject-into-dev-server.md` | `.env` migration example |
| `skills/tene-cli/examples/03-multi-env-ci-cd.md` | Multi-env + CI/CD example |
| `skills/tene-cli/tests/test.md` | 6 expected-behavior test scenarios |

Total bundle size: **48 KB** (limit: 50 MB).

## 3. Success Criteria Verification

| # | Criterion | Status |
|---|---|:---:|
| SC1 | SKILL.md parses as valid YAML | ✅ verified via Ruby YAML.safe_load |
| SC2 | Slug `tene-cli` matches `^[a-z0-9][a-z0-9-]*$` | ✅ |
| SC3 | Bundle size < 50 MB | ✅ 48 KB |
| SC4 | Frontmatter declarations match actual tene behavior | ✅ `requires.bins: [tene]` only; no fictional env vars |
| SC5 | 100% coverage of active tene commands | ✅ 14/14 (`init, set, get, list, delete, run, import, export, env, passwd, recover, version, update, whoami`) |
| SC6 | Safety rules preserved from source | ✅ 5 rules enumerated (4 NEVER + 1 use-tene-list) |
| SC7 | `clawhub skill publish --dry-run` local check | ⏸️ deferred (clawhub CLI not installed locally; command prepared) |
| SC8 | No secrets or `.tene/` contents in bundle | ✅ grep confirmed; only educational references |
| SC9 | Feature branch only; staging/main untouched | ✅ `feature/clawhub-skill-tene-cli` |
| SC10 | Publish command prepared; not executed | ✅ see §5 |

## 4. Iterate Loop Outcome

### Pass 1 gap scan → 94% match rate
Issues detected by gap-detector:
1. **CRITICAL** (later verified false-positive): claim about `TENE_KEYCHAIN_FALLBACK` env var.
   - Gap-detector said it didn't exist; source verification showed it **does** exist at `internal/keychain/keychain.go:82`. Kept as-is.
2. **Real fix** — `--codex` flag output path was wrong: documented as `.codex/rules.md`; actual is `AGENTS.md` (`internal/cli/init.go:47`, `internal/claudemd/template.go:87`). Fixed.
3. **Real fix** — claimed `.tene/recovery.json` exists as a separate file; actually the recovery blob lives inside `vault.db` (`vault_meta` table, `recovery_blob` key). Fixed.

### Pass 2 gap scan → 97% match rate → PASS

## 5. Publish Command (user-gated, NOT executed)

### 5.1 Prerequisites the user must satisfy
1. Install the clawhub CLI (one of):
   ```bash
   # Preferred — Bun-based
   curl -fsSL https://bun.sh/install | bash
   bun install -g clawhub

   # Or — if the npm/standalone path lands
   # npm install -g clawhub          # check availability
   # brew install clawhub            # if tap is published
   ```

2. Authenticate with GitHub OAuth:
   ```bash
   clawhub login            # opens browser → GitHub → loopback callback
   clawhub whoami           # verify
   ```

### 5.2 Dry run (safe, recommended)
```bash
cd /Users/popup-kay/Documents/GitHub/agentkay/tene

# If clawhub supports --dry-run:
clawhub skill publish ./skills/tene-cli --dry-run --no-input

# Or inspect the eventual payload manually:
find skills/tene-cli -type f | sort
du -sh skills/tene-cli
```

### 5.3 First publish (writes to the ClawHub registry)
```bash
cd /Users/popup-kay/Documents/GitHub/agentkay/tene

clawhub skill publish ./skills/tene-cli \
  --slug tene-cli \
  --name "Tene CLI — Local-First Secrets" \
  --version 1.0.0 \
  --tags latest \
  --changelog "Initial release: 100% active command coverage, AI safety rules baked in, curl installer support (Homebrew tap pending)."
```

Expected response: a URL of the form `https://clawhub.ai/<your-github-handle>/tene-cli`.

### 5.4 Post-publish verification
```bash
# From any machine:
clawhub inspect tene-cli
clawhub inspect tene-cli --files

# Install test in a scratch directory:
mkdir -p /tmp/tene-skill-qa && cd /tmp/tene-skill-qa
clawhub install tene-cli
clawhub list
test -f skills/tene-cli/SKILL.md && echo "OK: install produced SKILL.md"
```

### 5.5 Future updates
```bash
# Bump version per semver (patch/minor/major), then:
clawhub skill publish ./skills/tene-cli \
  --slug tene-cli \
  --version 1.1.0 \
  --tags latest \
  --changelog "<what changed>"
```

## 6. Git Status (as of this report)

```
On branch feature/clawhub-skill-tene-cli
Untracked (to be committed in same PR):
  docs/01-plan/clawhub-skill-tene-cli.md
  docs/02-design/clawhub-skill-tene-cli.md
  docs/03-report/clawhub-skill-tene-cli.md
  docs/research/clawhub-skill-publishing.md   (from earlier turn)
  skills/tene-cli/
```

## 7. Next Actions for User

1. **Review**: read `skills/tene-cli/SKILL.md` and confirm the agent instructions match your expectations.
2. **Commit**: run the commit step (prepared below in §8).
3. **Open PR** to `staging` (per `.claude/rules/git-workflow.md`):
   ```bash
   gh pr create --base staging --title "feat: add ClawHub skill for tene CLI" --body "..."
   ```
4. **Merge to staging**, then **PR staging → main** (normal flow).
5. **Publish to ClawHub** using §5.3 (requires GitHub OAuth).
6. **Announce** in Discord / README once live at `https://clawhub.ai/<handle>/tene-cli`.

## 8. Commit Preparation

Suggested commit message:

```
feat(skill): add ClawHub tene-cli skill (v1.0.0)

- SKILL.md with 100% coverage of active tene commands
- 5 AI safety rules (no tene get, no plain export, no .tene/ reads,
  no CLI-arg secrets, use tene list for introspection)
- 3 end-to-end examples + 6 expected-behavior test cases
- Frontmatter: requires.bins=[tene], install=download(install.sh),
  os=[macos, linux]
- PDCA docs under docs/01-plan/, docs/02-design/, docs/03-report/
- ClawHub research under docs/research/

Final design-impl match rate: 97% (PASS).
Bundle size: 48 KB. Slug: tene-cli.
```

## 9. Risks Not Yet Addressed

- **ClawHub CLI install path still ambiguous**: `bun install -g clawhub` is the documented path in `openclaw/clawhub/docs/quickstart.md`, but end-user availability via npm/Homebrew should be confirmed before publishing public docs pointing at the skill.
- **Version drift**: if tene releases v2 with breaking changes, this skill must be updated. Suggest a CI job in `tene-cloud` or a follow-up PR here that diffs `internal/cli/` against `skills/tene-cli/SKILL.md` on each release tag.
- **Homebrew install spec**: deliberately omitted because no tap exists yet. When the tap lands, bump to v1.1.0 with an additional `install: kind: brew` entry.
