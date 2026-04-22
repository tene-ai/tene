# PDCA Design — tene-cli ClawHub Skill

- **Feature**: `clawhub-skill-tene-cli`
- **References**: [Plan](../01-plan/clawhub-skill-tene-cli.md), [Research](../research/clawhub-skill-publishing.md), [skill-format.md](https://github.com/openclaw/clawhub/blob/main/docs/skill-format.md)

## 1. Directory Layout

```
skills/tene-cli/
├── SKILL.md              # Agent-facing instructions (trigger + safety + workflows)
├── README.md             # Marketplace listing (human-facing)
├── .clawhubignore        # Exclude sensitive patterns from publish
├── examples/
│   ├── 01-init-and-first-secret.md
│   ├── 02-inject-into-dev-server.md
│   └── 03-multi-env-ci-cd.md
└── tests/
    └── test.md           # Expected behavior test cases (documentation only)
```

## 2. SKILL.md — Frontmatter Design

### 2.1 Required fields
- `name: tene-cli` — kebab-case, matches slug
- `description` — single sentence, **packed with trigger keywords**: secrets, API keys, credentials, env vars, .env, injection, encrypted, local-first
- `version: 1.0.0` — first stable release of the skill

### 2.2 metadata.openclaw
- `emoji: "🔐"`
- `homepage: https://tene.sh`
- `os: [macos, linux]` — install.sh unsupported on Windows per install.sh:1
- `requires.bins: [tene]` — hard dependency; no fallback
- **NO `requires.env`** — tene does not require any env var at runtime. `TENE_MASTER_PASSWORD` is optional (CI mode). Declaring it would trigger "metadata mismatch" security flag.
- `install:` — single `kind: download` entry pointing to the curl installer

### 2.3 Rationale for omissions
- No `primaryEnv` — no single credential env dominates (master key comes from keychain or prompt)
- No `command-dispatch: tool` — skill is prompt-only, not a direct slash command wrapper
- No `user-invocable: false` — keep auto-discoverable by model
- No `anyBins` — tene has no alternative implementations
- No `brew` install spec — no tap exists yet; adding would mislead users
- No `nix` block — no flake yet

## 3. SKILL.md — Body Structure

Sections in order:

1. **When to use this skill** — trigger conditions (user mentions secrets/keys/.env/credentials, or repo contains `.tene/` or `CLAUDE.md` referencing tene)
2. **Critical safety rules** — 5 non-negotiable AI guardrails (derived from `internal/claudemd/template.go` + project `CLAUDE.md`)
3. **Install** — one-liner + verify
4. **Core workflows** — 6 subsections: init, list/check, set, run, export/import, passwd/recover
5. **Commands reference** — compact table of active commands with top flags
6. **Environment variable interactions** — `TENE_MASTER_PASSWORD`, `TENE_KEYCHAIN_FALLBACK`, `NO_COLOR`, `--no-keychain`
7. **Troubleshooting** — top 8 error codes from `pkg/errors/codes.go` with fixes
8. **Architecture note** — 3-line summary of crypto (Argon2id + XChaCha20-Poly1305 + OS keychain) so agents can answer "is this safe?"
9. **Further reading** — links

Length target: ~350–450 lines of markdown. Dense, scannable, token-efficient.

## 4. README.md — Design

Outward-facing marketplace copy:
1. Logo/title line + badges row (install, license, GitHub stars)
2. Tagline: "Local-first encrypted secrets for AI-native projects"
3. "Why" — 3 bullets (zero-knowledge, no cloud, baked-in AI safety)
4. Quick start — 4-line shell block
5. What this skill does — how it integrates with Claude Code / OpenClaw
6. Commands — same compact table as SKILL.md §5
7. License note — Skill is MIT-0; tene CLI is MIT
8. Links

## 5. `.clawhubignore` Design

Patterns to exclude from publish (belt-and-suspenders; these files shouldn't exist in `skills/tene-cli/` anyway):
```
.tene/
*.tene.enc
.env
.env.*
*.key
*.pem
*.p12
node_modules/
dist/
tmp/
.DS_Store
```

## 6. Examples — Design

### 6.1 `01-init-and-first-secret.md`
Scenario: fresh macOS machine, starting a new Node project.
Narrative: install → `tene init` → save Stripe test key from user prompt → run `npm test` with injection.

### 6.2 `02-inject-into-dev-server.md`
Scenario: user has a `.env` file they want to eliminate.
Narrative: `tene import .env --overwrite` → `rm .env && echo '.env' >> .gitignore` → replace all `npm start`/`next dev` invocations with `tene run -- <cmd>`.

### 6.3 `03-multi-env-ci-cd.md`
Scenario: local dev vs staging vs prod; GitHub Actions pipeline.
Narrative: create `local`, `staging`, `prod` envs → `tene set X --env prod` → GitHub Actions uses `TENE_MASTER_PASSWORD` secret + `--no-keychain` flag → `tene run --env prod -- ./deploy.sh`.

Each example file uses the format:
```markdown
## <Example Title>

**Scenario**: <1-2 lines>

**Input signal**: <what the user says>

**Expected AI behavior**:
- Step 1 ...
- Step 2 ...

**Commands executed**:
\`\`\`bash
<commands>
\`\`\`

**Unsafe patterns to avoid**: <specific to this example>
```

## 7. tests/test.md — Design

3 test scenarios covering the core safety boundaries:
1. Listing secrets safely (agent uses `tene list`, not `tene get`/`export`)
2. Running dev server (agent uses `tene run -- npm start`, never creates `.env`)
3. Refusing to print values (agent declines to run `tene get X`; instructs user)

Format follows the ClawHub blog "test.md" convention: `## Test Name` + `Input:` + `Expected behavior:`.

## 8. Publishing Workflow Design

### 8.1 Pre-flight (local, safe)
```bash
# Structural validation
test -f skills/tene-cli/SKILL.md
head -20 skills/tene-cli/SKILL.md | grep -E '^(name|description|version):'

# Size check
du -sh skills/tene-cli

# Slug regex
echo "tene-cli" | grep -E '^[a-z0-9][a-z0-9-]*$'

# Dry-run (if clawhub CLI installed)
clawhub skill publish ./skills/tene-cli --dry-run --no-input || true
```

### 8.2 Publish (user-gated, shared state)
```bash
clawhub login                       # interactive GitHub OAuth
clawhub whoami                      # verify

clawhub skill publish ./skills/tene-cli \
  --slug tene-cli \
  --name "Tene CLI — Local-First Secrets" \
  --version 1.0.0 \
  --tags latest \
  --changelog "Initial release: 100% command coverage, AI safety rules, curl installer support."
```

### 8.3 Post-publish verification
```bash
clawhub inspect tene-cli
# Expect: version 1.0.0, files list includes SKILL.md + README.md + examples/* + tests/*

# Fresh install test
mkdir -p /tmp/tene-skill-qa && cd /tmp/tene-skill-qa
clawhub install tene-cli
clawhub list
test -f skills/tene-cli/SKILL.md
```

## 9. Quality Gates (Gap Detection Target)

| Gate | Threshold | How measured |
|---|---|---|
| Coverage of active tene commands | 100% | checklist present |
| Safety rules from source CLAUDE.md preserved | 100% verbatim intent | diff |
| Frontmatter YAML validity | must parse | yq / python yaml |
| No plaintext secrets or `.tene/` in bundle | zero | grep / find |
| README + SKILL.md length ratio | README < SKILL.md | wc -l |
| Match rate (design ↔ implementation) | ≥ 90% | gap-detector equivalent |

If match rate < 90%, iterate up to 5 times.

## 10. Traceability

Every claim in SKILL.md must trace to:
- A tene source file (command matrix section 1-12), OR
- An openclaw/clawhub doc (skill-format.md)

No invented behavior, no speculative features.
