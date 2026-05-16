---
template: design
version: 1.3
description: tene CLI v2.0 Sprint 1 — audit-reader (P0-A1 + P0-E1 + P0-G1 + P0-K1) Design Document
variables:
  - feature: audit-reader
  - sprint: tene-cli-v2-s1
  - date: 2026-05-13
  - author: cto-lead (frontend-architect + security-architect perspective)
  - project: tene CLI
  - version: v2.0.0 (target)
  - planPath: docs/01-plan/features/audit-reader.plan.md
  - trustLevel: L2
---

# audit-reader Design Document

> **Summary**: `tene audit --since/--actor/--limit/--json` Cobra subcommand 신설 + `audit_log.actor` 컬럼 + 8-entry exit code 표 정렬 + `get.go` U-1 guard JSON warning fix + `keychain` Set+Delete 프로빙 제거 + `init` next-step 3-line.
>
> **Project**: tene CLI
> **Version**: target v2.0.0
> **Sprint**: tene-cli-v2-s1 (W1-W2)
> **Author**: cto-lead (frontend-architect for CLI UX + security-architect for SQL injection defense)
> **Date**: 2026-05-13
> **Status**: Draft (L2)
> **Planning Doc**: [audit-reader.plan.md](../../01-plan/features/audit-reader.plan.md)

---

## Context Anchor

| Key | Value |
|-----|-------|
| **WHY** | audit_log 기록만 / read CLI 없음 ("theater"); exit code drift (광고 vs 실제); JSON guard 미검증; keychain 시동 20-30ms overhead. v2.0 Show HN 메시지 #3 "AI session forensics" prerequisite. |
| **WHO** | **AI-vibe coder (50%)** + **Sec-conscious team (15%)** + **Indie OSS dev (25%)** — exit code 정확성, forensics 가시화. |
| **RISK** | (a) SQL injection (mitigation: `?` placeholder). (b) exit code 2→8 BREAKING (mitigation: migration guide). (c) v1 vault `actor` 컬럼 부재 (mitigation: COALESCE). (d) keychain probe 제거 후 fallback path. |
| **SUCCESS** | `tene audit --since 24h --json` 동작; exit code 1:1 mapping; JSON+nonTTY stderr emit; `time tene version` ≤ 35ms; `init` 3-line footer. |
| **SCOPE** | Sprint 1 Trk-C (3 PR: #8 + #11 + #12). 5 dev-day. +460 LOC. internal/cli (audit.go new) + internal/vault + pkg/errors + internal/keychain + docs. |

---

## 1. Overview

### 1.1 Design Goals

1. **AI forensics gateway**: `tene audit` CLI 가 SQLite audit_log 의 자연스러운 read surface
2. **SQL injection-safe**: `?` placeholder 강제 (모든 user-controlled flag 값)
3. **Exit code accuracy**: docs ↔ code 1:1 mapping; BREAKING 변경은 migration guide
4. **JSON guard symmetry**: TTY 차단 시 JSON 모드도 stderr emit (regression detection 가능)
5. **Performance**: keychain probe 제거로 매 명령 8-30ms 절감
6. **Backward-compatible audit_log**: 신규 vault 만 `actor` 컬럼; 기존 v1 vault `SELECT COALESCE(actor, 'human')`
7. **Versioned info domain**: HKDF info `tene/audit/v1` (sibling `crypto-v2-keys` 공유)
8. **Meta-audit**: audit reader 자체도 audit_log 에 `audit.read` 기록

### 1.2 Design Principles

- **CLI-first UX**: human-readable default + `--json` flag (consistent with other tene commands)
- **Defense in depth**: prepared placeholders + `--limit` cap + input sanitization
- **YAGNI on output formats**: human + JSON (no csv/ndjson; v2.1)
- **Explicit cap**: `--limit` enforced max 1000 (server-side query cap)
- **Versioned table change**: `actor` column with DEFAULT 'human' (no migration needed for new vaults)

---

## 2. Architecture Options

### 2.0 Architecture Comparison

| Criteria | Option A: Single audit.go | Option B: audit/ subpackage | Option C: tene log alias + structured logger |
|----------|:-:|:-:|:-:|
| **Approach** | `internal/cli/audit.go` 단일 file + RunE | `internal/cli/audit/` 패키지 + sub-subcommands | `tene log` rename + slog integration |
| **New Files** | 2 (audit.go + audit_test.go) | 5+ (root_cmd, list, since, format) | 다수 |
| **Modified Files** | ~6 (vault.go + schema + exit codes + get.go + keychain + init) | ~6 (same) | many |
| **Complexity** | Low | Medium | High |
| **Maintainability** | High | Medium (over-engineered for v2.0) | High but rewrites audit semantics |
| **Effort** | Medium | Medium | High |
| **Extensibility** | Medium (single file grows) | High (subcommands easy to add) | Highest |
| **Risk** | Low | Low | High (BREAKING semantic) |
| **Recommendation** | **Default choice** | Future (v2.1 — `tene audit list/export/verify`) | Out of scope |

**Selected**: **Option A — Single audit.go** — **Rationale**: v2.0 scope 는 read only. YAGNI — sub-subcommand 분할은 `tene audit export/verify` 추가 시점에. Single file 은 navigation/test 가장 쉽다. 추가 명령은 같은 file 안에 helper 함수로 자라거나 v2.1 에서 패키지로 promote.

### 2.1 Component Diagram

```
┌────────────────────────────────────────────────────────────────────┐
│                       internal/cli/                                 │
├────────────────────────────────────────────────────────────────────┤
│  audit.go [NEW]                                                     │
│    var auditCmd = &cobra.Command{Use: "audit", RunE: runAudit}     │
│    func runAudit(cmd, args) error {                                 │
│      parseSinceFlag → time.Time                                     │
│      parseActorFlag → string                                        │
│      parseLimitFlag → int (capped 1-1000)                          │
│      entries, _ := app.Vault.GetAuditLog(since, actor, limit)       │
│      if flagJSON: printJSON(entries)                                │
│      else: printHumanReadable(entries)                              │
│      _ = app.Vault.AddAuditLog("audit.read", details, "", actor)    │
│    }                                                                │
│                                                                    │
│  init.go [MODIFIED line ~221]                                       │
│    next-step footer 3 lines (tene set, list, run)                  │
│                                                                    │
│  get.go [MODIFIED lines 93-99]                                      │
│    Guard: TTY stdout → stderr "Refusing..." (now also JSON mode)    │
│                                                                    │
│  root.go [MODIFIED]                                                 │
│    rootCmd.AddCommand(auditCmd) added                              │
└─────────────────────────────────────────────────────────────────────┘
            ▼
┌────────────────────────────────────────────────────────────────────┐
│                       internal/vault/                               │
├────────────────────────────────────────────────────────────────────┤
│  schema.go [MODIFIED]                                               │
│    CREATE TABLE audit_log (                                         │
│      ...                                                            │
│      actor TEXT DEFAULT 'human'   ← NEW column                     │
│    )                                                                │
│                                                                    │
│  vault.go [MODIFIED]                                                │
│    AddAuditLog(action, resource, details, actor) — new param        │
│    GetAuditLog(since, actor, limit) — NEW function                 │
└─────────────────────────────────────────────────────────────────────┘
            ▼
┌────────────────────────────────────────────────────────────────────┐
│                       pkg/errors/                                   │
├────────────────────────────────────────────────────────────────────┤
│  codes.go [MODIFIED]                                                │
│    ErrGeneral (1), ErrInvalidInput (2),                            │
│    ErrVaultNotFound (3), ErrInvalidPassword/AuthFailed (4),         │
│    ErrSecretNotFound (5), ErrDecryptFailed (6),                    │
│    ErrInteractiveRequired (7), ErrStdoutSecretBlocked (8)          │
└─────────────────────────────────────────────────────────────────────┘
            ▼
┌────────────────────────────────────────────────────────────────────┐
│                       internal/keychain/                            │
├────────────────────────────────────────────────────────────────────┤
│  keychain.go [MODIFIED lines 91-97]                                 │
│    NewStore() — no probing; just return &Store{}                    │
│    Load() — first call lazy-init + fallback decision                │
└─────────────────────────────────────────────────────────────────────┘
            ▼
┌────────────────────────────────────────────────────────────────────┐
│                       docs/                                         │
├────────────────────────────────────────────────────────────────────┤
│  cli-reference.md [MODIFIED] — sync exit code table                 │
│  migration/exit-codes.md [NEW] — v1 → v2 mapping                    │
└─────────────────────────────────────────────────────────────────────┘
```

### 2.2 Data Flow — `tene audit --since 24h --actor ai --json`

```
User: tene audit --since 24h --actor ai --limit 50 --json
  │
  ▼
runAudit(cmd, args)
  │
  ├─ since ← parseSinceFlag("24h")  // returns time.Time = (now - 24h)
  │     │
  │     ├─ Try time.ParseDuration("24h") → 24*time.Hour
  │     ├─ Fall back to regex 1d/7d/1w/1month
  │     └─ since = time.Now().Add(-d)
  │
  ├─ actor ← parseActorFlag("ai")  // "human" | "ai" | "any" (default "any")
  │
  ├─ limit ← parseLimitFlag(50)  // 1 ≤ limit ≤ 1000
  │
  ├─ app, _ ← loadApp()
  │
  ├─ entries ← app.Vault.GetAuditLog(since, actor, limit)
  │     │
  │     ▼
  │  internal/vault/vault.go:GetAuditLog
  │     │
  │     ├─ sql ← "SELECT timestamp, action, resource_name, details, COALESCE(actor, 'human') AS actor
  │     │        FROM audit_log
  │     │        WHERE timestamp >= ?
  │     │          AND (? = 'any' OR actor = ?)
  │     │        ORDER BY timestamp DESC
  │     │        LIMIT ?"
  │     ├─ rows ← db.Query(sql, since.Unix(), actor, actor, limit)
  │     └─ for rows.Next() { scan into AuditEntry; append }
  │
  ├─ if flagJSON:
  │     printJSON(entries)  // [] if empty (never null)
  │  else:
  │     for entry in entries:
  │       fmt.Printf("[%s] [%s] %s %s (%s)\n",
  │           entry.Timestamp.Format(RFC3339),
  │           entry.Actor,
  │           entry.Action,
  │           entry.ResourceName,
  │           entry.Details)
  │
  ├─ // Meta-audit: log this read (doesn't recursive into next audit query)
  │  _ = app.Vault.AddAuditLog(
  │       "audit.read",
  │       "",
  │       fmt.Sprintf("since=%s actor=%s limit=%d", since, actor, limit),
  │       os.Getenv("TENE_ACTOR_ID"))
  │
  └─ return nil  // exit 0
```

### 2.3 Dependencies

| Component | Depends On | Purpose |
|-----------|-----------|---------|
| `internal/cli/audit.go` | `internal/vault.Vault.GetAuditLog`, `time`, `regexp`, `cobra` | Command + parsing |
| `internal/vault/vault.go:GetAuditLog` | `modernc.org/sqlite`, `internal/vault/types.go:AuditEntry` | SQL read |
| `internal/vault/schema.go` | (existing) | Schema change (actor column) |
| `pkg/errors/codes.go` | n/a | 8 exit code constants |
| `internal/keychain/keychain.go` | `github.com/zalando/go-keyring`, `internal/keychain/fallback.go` | Lazy init |

---

## 3. Data Model

### 3.1 Entity Definition

```go
// internal/vault/types.go (existing struct, ADD `Actor` field)
type AuditEntry struct {
    ID           int64  `json:"-"`                        // internal; not in JSON output
    Timestamp    int64  `json:"timestamp"`                // Unix seconds
    Action       string `json:"action"`                   // e.g., "vault.passwd_changed"
    ResourceName string `json:"resource_name,omitempty"`  // e.g., "AWS_KEY"
    Details      string `json:"details,omitempty"`        // free-form
    Actor        string `json:"actor"`                    // e.g., "human", "ai", "cursor-session-..."
}
```

### 3.2 Entity Relationships

```
[Vault] 1 ──── N [AuditEntry]
```

### 3.3 Database Schema

```sql
-- internal/vault/schema.go (MODIFIED — only takes effect on new vaults)
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp INTEGER NOT NULL,
    action TEXT NOT NULL,
    resource_name TEXT,
    details TEXT,
    actor TEXT DEFAULT 'human'   -- NEW (Sprint 2 backfills for old vaults via migration 002_audit_log_v2)
);

CREATE INDEX IF NOT EXISTS idx_audit_log_timestamp ON audit_log(timestamp);
-- NEW index — speeds up "WHERE timestamp >= ?" queries
```

### 3.4 Action Vocabulary (existing 9 + 3 new)

| Action | Triggered By | Resource | Details |
|--------|-------------|----------|---------|
| `vault.init` | `tene init` | "" | "" |
| `vault.passwd_changed` | `tene passwd` (success) | "" | "" |
| `vault.passwd_failed` (NEW; sibling passwd-verify) | `tene passwd` (wrong) | "" | "" |
| `vault.recovered` | `tene recover` | "" | "" |
| `secret.created` | `tene set` | secret name | "" |
| `secret.updated` | `tene set` (existing) | secret name | "" |
| `secret.deleted` | `tene delete` | secret name | "" |
| `secret.read` | `tene get`, `tene run`, `tene list` | secret name(s) | "" |
| `vault.exported` | `tene export` | "" | file path |
| `vault.imported` | `tene import` | "" | file path / source |
| `audit.read` (NEW; this design) | `tene audit` | "" | "since=... actor=... limit=..." |

---

## 4. API Specification

### 4.1 CLI Command — `tene audit`

```
$ tene audit --help
Read the vault's audit log. Shows recent actions on secrets with timestamps,
actors (human/AI/automation), and event metadata.

Usage:
  tene audit [flags]

Flags:
  --since DURATION  Show entries newer than DURATION (e.g., 24h, 7d, 30d, 1month). Default: 24h.
  --actor STRING    Filter by actor: human, ai, or any. Default: any.
  --limit N         Max number of entries (1-1000). Default: 50.
  --json            Output as JSON array.
  -h, --help        Show this help.

Exit codes:
  0   OK
  1   General error
  2   Invalid input (e.g., bad --since format)
  6   Decrypt failed (corrupted vault.db)

Examples:
  # Show last 24 hours of activity (default)
  tene audit

  # Show last week of AI agent activity as JSON
  tene audit --since 7d --actor ai --json

  # Audit forensics: find when AWS_KEY was last read
  tene audit --since 30d --json | jq '.[] | select(.resource_name == "AWS_KEY")'
```

### 4.2 New Vault API

#### `GetAuditLog(since time.Time, actor string, limit int) ([]AuditEntry, error)`

```go
// internal/vault/vault.go (NEW)
//
// GetAuditLog returns recent audit log entries matching the filters.
//
// Parameters:
//   since: only entries with timestamp >= since are returned
//   actor: "human", "ai", or "any" — filter on actor column
//   limit: max number of entries (1 ≤ limit ≤ 1000)
//
// SQL uses ? placeholders; user input from CLI flags is sanitized by the
// driver, eliminating injection risk. Limit is enforced at the SQL level
// (LIMIT clause), so callers cannot bypass it.
//
// Returns entries ordered by timestamp DESC. Empty result returns []AuditEntry{}, nil.
func (v *Vault) GetAuditLog(since time.Time, actor string, limit int) ([]AuditEntry, error) {
    // Defensive clamp (UI parses; this is the boundary)
    if limit < 1 {
        limit = 50
    } else if limit > 1000 {
        limit = 1000
    }
    if actor == "" {
        actor = "any"
    }

    query := `
        SELECT id, timestamp, action,
               COALESCE(resource_name, '') AS resource_name,
               COALESCE(details, '') AS details,
               COALESCE(actor, 'human') AS actor
        FROM audit_log
        WHERE timestamp >= ?
          AND (? = 'any' OR actor = ?)
        ORDER BY timestamp DESC
        LIMIT ?
    `
    rows, err := v.db.Query(query, since.Unix(), actor, actor, limit)
    if err != nil {
        return nil, fmt.Errorf("query audit log: %w", err)
    }
    defer rows.Close()

    out := make([]AuditEntry, 0, limit)
    for rows.Next() {
        var e AuditEntry
        if err := rows.Scan(&e.ID, &e.Timestamp, &e.Action, &e.ResourceName, &e.Details, &e.Actor); err != nil {
            return nil, fmt.Errorf("scan audit row: %w", err)
        }
        out = append(out, e)
    }
    if err := rows.Err(); err != nil {
        return nil, fmt.Errorf("iterate audit rows: %w", err)
    }
    return out, nil
}
```

#### `AddAuditLog(action, resourceName, details, actor string) error` (signature change)

```go
// internal/vault/vault.go (MODIFIED — added `actor` parameter)
//
// Before:
//   func (v *Vault) AddAuditLog(action, resourceName, details string) error
//
// After:
//   func (v *Vault) AddAuditLog(action, resourceName, details, actor string) error
//
// `actor` should be derived from os.Getenv("TENE_ACTOR_ID") or default to "human".
// Helper `resolveActor()` in internal/cli/root.go.
func (v *Vault) AddAuditLog(action, resourceName, details, actor string) error {
    if actor == "" {
        actor = "human"
    }
    _, err := v.db.Exec(
        `INSERT INTO audit_log (timestamp, action, resource_name, details, actor)
         VALUES (?, ?, ?, ?, ?)`,
        time.Now().Unix(), action, resourceName, details, actor,
    )
    if err != nil {
        return fmt.Errorf("insert audit log: %w", err)
    }
    return nil
}
```

#### `resolveActor() string` (helper in cli/root.go)

```go
// internal/cli/root.go (NEW helper)
func resolveActor() string {
    if v := os.Getenv("TENE_ACTOR_ID"); v != "" {
        return v
    }
    return "human"
}
```

### 4.3 Modified `runGet` (P0-G1)

```go
// internal/cli/get.go (MODIFIED lines 93-99)
//
// BEFORE: TTY guard only triggers for non-JSON mode
// AFTER: TTY guard fires for any TTY stdout (JSON or human)

func runGet(cmd *cobra.Command, args []string) error {
    // ... (existing flow)

    // U-1 guard: prevent printing secret to a TTY stdout regardless of format
    if isTerminal(os.Stdout) {
        fmt.Fprintln(os.Stderr,
            "Refusing to print secret to a TTY (use `--json | cat` to pipe, or redirect to a file). Exit 8.")
        return teneerr.ErrStdoutSecretBlocked  // NEW exit code: 8
    }

    // ... (existing output)
}
```

### 4.4 Exit Code Table (P0-E1)

```go
// pkg/errors/codes.go (MODIFIED — full 8-entry table)

var (
    // Exit code 0: success (no error type; cmd returns nil)

    // Exit code 1: General/unclassified error
    ErrGeneral = TeneError{Code: 1, Message: "general error", Type: "GENERAL_ERROR"}

    // Exit code 2: Invalid user input (CLI flag, file format, etc.)
    ErrInvalidInput = TeneError{Code: 2, Message: "invalid input", Type: "INVALID_INPUT"}

    // Exit code 3: Vault not found (~/.tene/vault.db missing)
    ErrVaultNotFound = TeneError{Code: 3, Message: "vault not found", Type: "VAULT_NOT_FOUND"}

    // Exit code 4: Authentication failed (master password mismatch)
    ErrInvalidPassword = TeneError{Code: 4, Message: "invalid password", Type: "AUTH_FAILED"}
    ErrAuthFailed      = ErrInvalidPassword  // alias

    // Exit code 5: Secret not found
    ErrSecretNotFound = TeneError{Code: 5, Message: "secret not found", Type: "SECRET_NOT_FOUND"}

    // Exit code 6: Decryption failed (AAD mismatch, corrupted ciphertext, unsupported KDF)
    ErrDecryptFailed   = TeneError{Code: 6, Message: "decryption failed", Type: "DECRYPT_FAILED"}
    ErrUnsupportedKDF  = TeneError{Code: 6, Message: "unsupported KDF algorithm", Type: "DECRYPT_FAILED"}

    // Exit code 7: Interactive prompt required but stdin is non-TTY
    ErrInteractiveRequired = TeneError{Code: 7, Message: "interactive prompt required", Type: "INTERACTIVE_REQUIRED"}

    // Exit code 8: TTY stdout guard refused secret output (BREAKING change from v1.0.8 exit 2)
    ErrStdoutSecretBlocked = TeneError{Code: 8, Message: "refusing to print secret to TTY", Type: "STDOUT_SECRET_BLOCKED"}
)
```

### 4.5 Modified `keychain.NewStore` (P0-K1)

```go
// internal/keychain/keychain.go (MODIFIED lines 91-97)
//
// BEFORE:
//   func NewStore() (*Store, error) {
//       err := keyring.Set(serviceName, "_probe", "_dummy")
//       if err != nil { return nil, fmt.Errorf("keychain unavailable: %w", err) }
//       _ = keyring.Delete(serviceName, "_probe")
//       return &Store{...}, nil
//   }
// AFTER (lazy init): no probe; first Load() determines fallback.

func NewStore() *Store {
    return &Store{
        primary:  keyringStore{},  // delegates to zalando/go-keyring
        fallback: nil,              // resolved lazily on first Load() failure
    }
}

func (s *Store) Load() ([]byte, error) {
    key, err := s.primary.Load()
    if err == nil {
        return key, nil
    }
    // Lazy fallback decision
    if s.fallback == nil {
        s.fallback = fallback.New()
        fmt.Fprintln(os.Stderr, "warning: OS keychain unavailable; using file-based fallback (`.tene/keystore.enc`)")
    }
    return s.fallback.Load()
}
```

> Note: this introduces a write-on-read mutation on first failure. Concurrency: tene CLI is single-process; no race. If we later parallelize, mutex needed.

### 4.6 Modified `tene init` (P1, T1-C7)

```go
// internal/cli/init.go (MODIFIED line ~221)
fmt.Println()
fmt.Println("Next steps:")
fmt.Println("  tene set FOO bar           # store a secret")
fmt.Println("  tene list                  # list all secrets")
fmt.Println("  tene run -- npm start      # run command with secrets injected")
fmt.Println()
fmt.Println("Run `tene --help` for all commands.")
```

---

## 5. UI/UX Design

### 5.1 `tene audit` (human output)

```
$ tene audit --since 24h
[2026-05-13T10:30:42Z] [human]              vault.init
[2026-05-13T10:30:55Z] [human]              secret.created  AWS_KEY
[2026-05-13T10:31:02Z] [cursor-session-...] secret.read     AWS_KEY
[2026-05-13T10:35:14Z] [human]              secret.created  STRIPE_KEY
[2026-05-13T10:40:00Z] [cursor-session-...] secret.read     STRIPE_KEY
[2026-05-13T11:00:23Z] [human]              vault.exported  /tmp/backup.tene.enc
```

### 5.2 `tene audit --json`

```json
[
  {
    "timestamp": 1715592623,
    "action": "vault.exported",
    "resource_name": "",
    "details": "/tmp/backup.tene.enc",
    "actor": "human"
  },
  {
    "timestamp": 1715587200,
    "action": "secret.read",
    "resource_name": "STRIPE_KEY",
    "details": "",
    "actor": "cursor-session-abc123"
  }
]
```

### 5.3 `tene audit --since 7d --actor ai --json | jq ...`

```
$ tene audit --since 7d --actor ai --json | jq '.[].resource_name' | sort -u
"AWS_KEY"
"OPENAI_API_KEY"
"STRIPE_KEY"
```

### 5.4 `tene init` next-step footer (NEW)

```
$ tene init
Master Password: ********
Confirm Master Password: ********

  Initializing vault at ~/.tene/vault.db ...
  Generated 32-byte salt (Argon2id).
  Storing master key in OS Keychain.

  Recovery Key (write this down and keep it safe!):
  +--------------------------------------------------+
  |   word1 word2 word3 ...                          |
  +--------------------------------------------------+

  Vault initialized successfully.

Next steps:
  tene set FOO bar           # store a secret
  tene list                  # list all secrets
  tene run -- npm start      # run command with secrets injected

Run `tene --help` for all commands.
$
```

### 5.5 `tene get FOO --json` on TTY (NEW; P0-G1 fix)

```
$ tene get FOO --json
Refusing to print secret to a TTY (use `--json | cat` to pipe, or redirect to a file). Exit 8.
$ echo $?
8

$ tene get FOO --json | cat
{"name":"FOO","value":"my-secret","environment":"default"}
$ echo $?
0
```

### 5.6 Page UI Checklist

#### `tene audit` command

- [ ] Help text shows usage, flags, exit codes, examples
- [ ] `--since 24h` (default) shows last 24 hours
- [ ] `--since 7d` shorthand recognized
- [ ] `--since 30d` shorthand recognized
- [ ] `--since 1month` shorthand recognized (= 720h)
- [ ] `--since invalid` → exit 2 + "invalid --since format"
- [ ] `--actor human|ai|any` filter
- [ ] `--limit 50` (default), 1 ≤ N ≤ 1000
- [ ] `--limit 5000` → clamped to 1000 (silent or warn?) — design choice: warn on stderr
- [ ] `--json` outputs `[]` for empty (not `null`)
- [ ] human output: `[timestamp] [actor] action resource_name (details)`
- [ ] meta-audit entry on success: `audit.read` row with details "since=X actor=Y limit=Z"
- [ ] SQL injection input: `--actor "h'; DROP TABLE audit_log; --"` → audit_log intact
- [ ] v1 vault (no actor column): `SELECT ... COALESCE(actor, 'human')` → actor="human"

#### `tene get --json` TTY guard

- [ ] TTY stdout (terminal) → exit 8 + stderr "Refusing..."
- [ ] Pipe (`| cat`) → exit 0 + JSON output
- [ ] Redirect (`> /tmp/foo`) → exit 0 + JSON output
- [ ] `--json` flag → same TTY guard applies

#### `tene init` next-step footer

- [ ] 3 lines (set, list, run)
- [ ] Trailing `tene --help` line
- [ ] Empty line between commands and help line

---

## 6. Error Handling

### 6.1 Error Code Table (full 0-8)

See §4.4 for full Go code. UX table:

| Code | Name | Cause | Example |
|------|------|-------|---------|
| 0 | OK | Success | `tene get FOO` (piped) |
| 1 | GENERAL_ERROR | Internal failure | corrupted SQLite, panic recovery |
| 2 | INVALID_INPUT | Bad CLI flag / argument | `tene audit --since "bad"` |
| 3 | VAULT_NOT_FOUND | `~/.tene/vault.db` missing | `tene get FOO` (vault not initialized) |
| 4 | AUTH_FAILED | Master password mismatch (sibling `passwd-verify`) | `tene passwd` (wrong old) |
| 5 | SECRET_NOT_FOUND | `tene get NAME` (NAME not exists) | `tene get NOPE` |
| 6 | DECRYPT_FAILED | AAD mismatch, corruption, unsupported KDF | corrupt `.tene.enc`, KDFAlg=0xFF |
| 7 | INTERACTIVE_REQUIRED | non-TTY without `TENE_MASTER_PASSWORD` | `echo "" \| tene passwd` |
| 8 | STDOUT_SECRET_BLOCKED | TTY guard | `tene get FOO` (terminal, no pipe) — BREAKING from v1.0.8 (was 2) |

### 6.2 Migration Guide — exit code change

See `docs/migration/exit-codes.md` (NEW; structure):

```markdown
# Exit code migration — tene v1.0.x → v2.0

## What changed

| v1.0.x | v2.0 | Description |
|--------|------|-------------|
| 2 (STDOUT_BLOCKED) | 8 | Refusing to print secret to TTY — moved out of "auth" group |
| 2 (everything else) | (varies) | See full table below |

## Why

In v1.0.x, exit code 2 was overloaded for both authentication failures and the
TTY guard. v2.0 separates these: auth uses 4 (AUTH_FAILED), TTY guard uses 8
(STDOUT_SECRET_BLOCKED).

## Automation script updates

### Before (v1.0.x)
```bash
tene get FOO
if [ $? -eq 2 ]; then
  echo "auth or stdout block — please retry"
fi
```

### After (v2.0)
```bash
tene get FOO
case $? in
  0) echo "ok" ;;
  4) echo "auth failed; check password" ;;
  5) echo "secret not found" ;;
  8) echo "refusing to print to TTY; use a pipe" ;;
  *) echo "other error: $?" ;;
esac
```
```

---

## 7. Security Considerations

### 7.1 SQL Injection Defense

- [x] **All user input via `?` placeholder**: `db.Query(sql, since.Unix(), actor, actor, limit)`
- [x] **`actor` flag validated**: only `human` / `ai` / `any` accepted (whitelist in parseActorFlag); other values → exit 2
- [x] **`limit` clamped**: 1 ≤ N ≤ 1000 at boundary
- [x] **`since` parsed by Go**: `time.ParseDuration` / regex — no raw SQL inserted
- [x] **Integration test**: `tene audit --actor "h'; DROP TABLE audit_log; --"` → audit_log row count unchanged

### 7.2 Audit log integrity

- [x] **`audit.read` meta-audit**: every read logged (forensics on forensics)
- [x] **No row deletion via CLI**: `tene audit` is read-only; no `--delete` or `--clear` flag
- [x] **`actor` column not user-controlled** for reads — only filter; insert uses `TENE_ACTOR_ID` env (caller can set, but cannot rewrite history)
- [x] **Out of scope for v2.0**: chain hash (tamper-evident), signed audit log — v2.1+

### 7.3 keychain probe removal

- [x] **No security regression**: probe just verifies keychain availability; removing it just defers the check
- [x] **Fallback path warning**: first failed Load() prints stderr warning before falling back to file-based keystore
- [x] **File keystore is encrypted**: `.tene/keystore.enc` uses XChaCha20-Poly1305 with master key (unchanged)

---

## 8. Test Plan

### 8.1 Test Scope

| Type | Target | Tool | Phase |
|------|--------|------|-------|
| L1: Unit | `parseSinceFlag` (24h / 7d / 30d / 1month / invalid) | `go test` | Do |
| L1: Unit | `parseActorFlag` (human / ai / any / invalid) | `go test` | Do |
| L1: Unit | `parseLimitFlag` (5 / 0 / 5000 / -1) | `go test` | Do |
| L1: Unit | `GetAuditLog` SQL — empty, single, multi, filter, limit | `go test` | Do |
| L1: Unit | `resolveActor()` (env set / unset) | `go test` | Do |
| L2: CLI | `tene audit --since 24h --json` end-to-end | testhelper_test.go | Do |
| L2: Security | SQL injection: `--actor "h'; DROP TABLE audit_log; --"` | integration_test.go | Do |
| L2: Backcompat | v1 vault fixture (no actor column) + `tene audit` | integration_test.go | Do |
| L2: Guard | `tene get FOO --json` TTY → exit 8 + stderr | testhelper_test.go | Do |
| L2: Guard | `tene get FOO --json \| cat` (non-TTY) → exit 0 + JSON | testhelper_test.go | Do |
| L4: Performance | `time tene version` (43ms → ≤ 35ms after keychain probe removal) | `time` + bash | Do |
| L3: UX | `tene init` produces 3-line next-step footer | testhelper_test.go | Do |

### 8.2 L1: Unit Test Scenarios

| # | Target | Test | Expected |
|---|--------|------|---------|
| 1 | `parseSinceFlag` | "24h" | 24*time.Hour |
| 2 | `parseSinceFlag` | "7d" | 7*24*time.Hour |
| 3 | `parseSinceFlag` | "30d" | 30*24*time.Hour |
| 4 | `parseSinceFlag` | "1month" | 720*time.Hour |
| 5 | `parseSinceFlag` | "garbage" | error → `ErrInvalidInput` |
| 6 | `parseActorFlag` | "" | "any" (default) |
| 7 | `parseActorFlag` | "ai" | "ai" |
| 8 | `parseActorFlag` | "AI" | "ai" (lowercase) |
| 9 | `parseActorFlag` | "robot" | error → `ErrInvalidInput` |
| 10 | `parseLimitFlag` | 50 | 50 |
| 11 | `parseLimitFlag` | 0 | 50 (default) |
| 12 | `parseLimitFlag` | 5000 | 1000 + stderr warn |
| 13 | `parseLimitFlag` | -1 | 50 (default) |
| 14 | `GetAuditLog` | empty table | `[]AuditEntry{}, nil` |
| 15 | `GetAuditLog` | 100 rows, since 1h, limit 50 | last 50 rows (DESC) |
| 16 | `GetAuditLog` | actor="ai" filter | only actor='ai' rows |
| 17 | `GetAuditLog` | actor="any" filter | all rows |
| 18 | `resolveActor` | env `TENE_ACTOR_ID=cursor-1` | "cursor-1" |
| 19 | `resolveActor` | env unset | "human" |

### 8.3 L2: CLI Test Scenarios

| # | Command | Setup | Expected |
|---|---------|-------|----------|
| 1 | `tene audit --since 24h --json` | empty audit_log | `[]` (not null) |
| 2 | `tene audit --since 24h --json` | 3 rows | JSON array of 3 |
| 3 | `tene audit --since 24h` | 3 rows | human output (no JSON) |
| 4 | `tene audit --since 24h --actor ai` | 5 rows (2 ai) | 2 rows |
| 5 | `tene audit --since invalid` | n/a | exit 2 + "invalid --since" |
| 6 | `tene audit --limit 0` | 100 rows | 50 rows (default) |
| 7 | `tene audit --limit 5000` | 100 rows | 100 rows + stderr warn |
| 8 | `tene audit --actor "h'; DROP TABLE audit_log;--"` | filled audit_log | exit 2 (actor whitelist rejects) + table intact |
| 9 | `tene get FOO --json` (TTY stdout) | FOO exists | exit 8 + stderr "Refusing..." |
| 10 | `tene get FOO --json \| cat` | FOO exists | exit 0 + JSON output |
| 11 | `tene init` (fresh dir) | n/a | 3-line next-step footer |
| 12 | `time tene version` | warm cache | wall < 35ms |

### 8.4 L3: E2E Forensics Scenario

| # | Scenario | Steps | Expected |
|---|----------|-------|----------|
| 1 | AI session reads secret; forensics | (a) `TENE_ACTOR_ID=cursor-1 tene get AWS_KEY` (b) `tene audit --json \| jq '.[] \| select(.resource_name == "AWS_KEY")'` | row with actor="cursor-1" |

### 8.5 L4: Performance Benchmark

| # | Command | Before | After |
|---|---------|-------:|------:|
| 1 | `time tene version` (1st call) | 43ms | ≤ 35ms |
| 2 | `time tene version` (warm) | 43ms | ≤ 35ms |
| 3 | `time tene list` | (n/a) | ≤ 50ms |

---

## 9. Clean Architecture

### 9.1 Layer Structure

| Layer | Component | Location |
|-------|-----------|----------|
| **Presentation** | `audit.go` Cobra command, output formatting | `internal/cli/audit.go` |
| **Application** | `runAudit` flag parsing, filter dispatch | `internal/cli/audit.go` |
| **Domain** | (none — read-only data; no business rules) | n/a |
| **Infrastructure** | `Vault.GetAuditLog`, `AddAuditLog` SQL | `internal/vault/vault.go` |
| **Cross-cutting** | 8 exit code constants | `pkg/errors/codes.go` |

### 9.2 Dependency Rules

```
internal/cli/audit.go ─→ internal/vault (Vault.GetAuditLog)
                       ─→ pkg/errors (ErrInvalidInput)
                       ─→ stdlib (time, regexp, encoding/json)
```

### 9.3 File Import Rules

| From | Imports | Forbidden |
|------|---------|-----------|
| `internal/cli/audit.go` | `time`, `regexp`, `encoding/json`, `github.com/spf13/cobra`, `internal/vault`, `pkg/errors` | direct SQL/sqlite |
| `internal/vault/vault.go` | `modernc.org/sqlite`, `fmt`, `time` | `internal/cli/*` |

### 9.4 This Feature's Layer Assignment

| Component | Layer | Location |
|-----------|-------|----------|
| `auditCmd`, `runAudit` | Presentation | `internal/cli/audit.go` |
| `parseSinceFlag`, `parseActorFlag`, `parseLimitFlag` | Presentation | `internal/cli/audit.go` |
| `Vault.GetAuditLog`, `Vault.AddAuditLog` | Infrastructure | `internal/vault/vault.go` |
| Exit code constants | Cross-cutting | `pkg/errors/codes.go` |

---

## 10. Coding Convention Reference

### 10.1 Naming Conventions

| Target | Rule | Example |
|--------|------|---------|
| Cobra commands | `{name}Cmd` | `auditCmd`, `getCmd` |
| Cobra RunE | `run{Name}` | `runAudit`, `runGet` |
| Flag parsers | `parse{Name}Flag` | `parseSinceFlag` |
| SQL placeholder | `?` (positional) | `WHERE timestamp >= ?` |
| Audit action strings | `{resource}.{verb}` snake | `"audit.read"`, `"vault.passwd_failed"` |
| Env vars | `TENE_{NAME}` UPPER_SNAKE | `TENE_ACTOR_ID` |
| Exit code constants | `Err{Cause}` | `ErrInvalidInput`, `ErrStdoutSecretBlocked` |

### 10.2 Import Order

```go
package cli

import (
    // 1. stdlib
    "encoding/json"
    "fmt"
    "os"
    "regexp"
    "strings"
    "time"

    // 2. external
    "github.com/spf13/cobra"

    // 3. internal
    "github.com/agent-kay-it/tene/internal/vault"
    teneerr "github.com/agent-kay-it/tene/pkg/errors"
)
```

### 10.3 Environment Variables

| Variable | Purpose | Scope |
|----------|---------|-------|
| `TENE_ACTOR_ID` | audit_log actor column (e.g., "cursor-session-abc", "claude-code") | Optional |

---

## 11. Implementation Guide

### 11.1 File Structure

```
tene/
├── internal/
│   ├── cli/
│   │   ├── audit.go          # NEW (~150 LOC)
│   │   ├── audit_test.go     # NEW (~120 LOC)
│   │   ├── root.go           # MODIFIED (rootCmd.AddCommand(auditCmd); resolveActor())
│   │   ├── get.go:93-99      # MODIFIED (JSON+TTY guard)
│   │   ├── get_guard_test.go # MODIFIED (confess removed; real stderr assert)
│   │   ├── init.go:221       # MODIFIED (3-line footer)
│   │   └── init_test.go      # MODIFIED (footer assert)
│   ├── vault/
│   │   ├── schema.go         # MODIFIED (actor column)
│   │   ├── vault.go          # MODIFIED (AddAuditLog signature, new GetAuditLog)
│   │   ├── types.go          # MODIFIED (AuditEntry.Actor field)
│   │   └── vault_test.go     # MODIFIED (test new APIs)
│   └── keychain/
│       └── keychain.go:91-97 # MODIFIED (probe removed, lazy fallback)
├── pkg/
│   └── errors/
│       └── codes.go          # MODIFIED (8 entries)
└── docs/
    ├── cli-reference.md      # MODIFIED (exit code table sync)
    └── migration/
        └── exit-codes.md     # NEW (v1 → v2 mapping)
```

### 11.2 Implementation Order

> **PR #8** (`feat(cli): tene audit reader command (--since/--actor/--json)`):

1. [ ] `internal/vault/types.go` — `AuditEntry.Actor string`
2. [ ] `internal/vault/schema.go` — `audit_log` 에 `actor TEXT DEFAULT 'human'` 컬럼 (CREATE TABLE only — new vaults; existing v1 vaults backfilled in s2)
3. [ ] `internal/vault/schema.go` — `CREATE INDEX idx_audit_log_timestamp`
4. [ ] `internal/vault/vault.go` — `AddAuditLog(action, resource, details, actor)` add param
5. [ ] `internal/vault/vault.go` — `GetAuditLog(since, actor, limit)` NEW
6. [ ] `internal/cli/root.go` — `resolveActor()` helper
7. [ ] `internal/cli/audit.go` (NEW) — auditCmd + runAudit + 3 parsers
8. [ ] `internal/cli/audit_test.go` — Unit tests for parsers + GetAuditLog mock
9. [ ] `internal/cli/root.go` — `rootCmd.AddCommand(auditCmd)`
10. [ ] (sweep) Update all 9 existing `AddAuditLog` callers to pass `resolveActor()` as 4th arg
11. [ ] Integration test: SQL injection attempt → audit_log intact
12. [ ] CHANGELOG.md — "feat(cli): tene audit reader for forensics (--since/--actor/--limit/--json)"

> **PR #11** (`fix(errors+docs): exit code drift fix + STDOUT_SECRET_BLOCKED → 8`):

13. [ ] `pkg/errors/codes.go` — full 8-entry table (see §4.4)
14. [ ] `pkg/errors/codes_test.go` — verify each code emits expected Code
15. [ ] `internal/cli/get.go:93-99` — JSON+TTY guard (exit 8)
16. [ ] `internal/cli/get_guard_test.go` — actual stderr capture + assert
17. [ ] `docs/cli-reference.md` — exit code table sync
18. [ ] `docs/migration/exit-codes.md` (NEW) — v1 → v2 mapping + bash example
19. [ ] CHANGELOG.md — "BREAKING: exit code 2 (STDOUT_BLOCKED) → 8; new auth exit code 4 (AUTH_FAILED) — see docs/migration/exit-codes.md"

> **PR #12** (`perf(keychain): remove startup Set+Delete probe (save 8ms)`):

20. [ ] `internal/keychain/keychain.go:91-97` — `NewStore()` no probe; just `return &Store{...}`
21. [ ] `internal/keychain/keychain.go` — `Load()` lazy fallback (warn on stderr first failure)
22. [ ] `internal/keychain/keychain_test.go` — verify no Set+Delete during NewStore
23. [ ] `internal/keychain/keychain_bench_test.go` — `BenchmarkNewStore` (target ≤ 1ms)
24. [ ] Manual: `time tene version` before/after — confirm ≥ 8ms savings
25. [ ] CHANGELOG.md — "perf(keychain): remove startup probe — saves 8-30ms per invocation on macOS"

> **PR (folded into #8 or #11)** (`feat(cli): tene init next-step 3-line`):

26. [ ] `internal/cli/init.go:221` — 3-line footer (set, list, run) + help line
27. [ ] `internal/cli/init_test.go` — assert footer text
28. [ ] CHANGELOG.md — "feat(cli): tene init shows 3-line next-step guide"

### 11.3 Session Guide

#### Module Map

| Module | Scope Key | Description | Estimated Turns |
|--------|-----------|-------------|:---------------:|
| audit command + vault API | `module-1` | PR #8 (audit.go + GetAuditLog + AddAuditLog signature + schema actor) | 30-40 |
| exit code drift + guard | `module-2` | PR #11 (8-code table + get.go guard + docs migration) | 20-25 |
| keychain perf + init UX | `module-3` | PR #12 (keychain lazy) + init footer | 15-20 |

#### Recommended Session Plan

| Session | Phase | Scope | Turns |
|---------|-------|-------|:-----:|
| Session 1 | Plan + Design | (this document) | 30-35 (current) |
| Session 2 | Do | `--scope module-1` (PR #8) | 35-45 |
| Session 3 | Do | `--scope module-2` (PR #11) | 20-25 |
| Session 4 | Do | `--scope module-3` (PR #12 + init) | 20-25 |
| Session 5 | Check + Report | All 3 modules | 25-30 |

---

## 12. Edge Cases & Failure Modes

| # | Scenario | Behavior |
|---|----------|----------|
| 1 | `tene audit` on a v1.0.x vault (no actor column) | SQL `COALESCE(actor, 'human')` returns "human" — works |
| 2 | `tene audit --since 100y` (very large window) | SQL fine; just slow; `--limit 1000` caps results |
| 3 | `tene audit --limit -1` | clamped to 50 (default) |
| 4 | `TENE_ACTOR_ID=""` empty env | treated as unset → "human" |
| 5 | `TENE_ACTOR_ID="cursor session $(rm -rf)"` shell injection | tene treats as opaque string; SQL placeholder safe |
| 6 | `--actor "Robot"` (case sensitivity) | parseActorFlag normalizes to lowercase; "robot" → exit 2 |
| 7 | `tene audit --json` on empty result | output: `[]\n` (not `null`) |
| 8 | Concurrent `tene audit` invocations | SQLite WAL serializes; each read is consistent snapshot |
| 9 | `tene audit` while another process is writing audit_log | reader sees committed state; potential 0 ms staleness |
| 10 | `tene get FOO --json > /tmp/foo` (redirect) | non-TTY stdout → exit 0 + JSON written |
| 11 | `tene get FOO --json 2>&1 > /tmp/foo` (stderr to stdout, stdout redirect) | guard sees non-TTY stdout → exit 0 + JSON written (correct behavior) |
| 12 | Keychain returns permanent error on Load() | fallback (file keystore) used; stderr warn shown once |
| 13 | Both keychain and file keystore fail | exit 1 + "no keystore available" |

---

## 13. Sequence Diagram — `tene audit --since 24h --json`

```
User           audit.go            vault.go               SQLite
 │                │                    │                     │
 │ tene audit...  │                    │                     │
 ├───────────────▶│                    │                     │
 │                │ parseSinceFlag     │                     │
 │                │ parseActorFlag     │                     │
 │                │ parseLimitFlag     │                     │
 │                │ loadApp() (existing)                     │
 │                │ GetAuditLog(since, actor, limit)         │
 │                ├───────────────────▶│                     │
 │                │                    │ db.Query(...)        │
 │                │                    ├────────────────────▶│
 │                │                    │◀─── rows ───────────│
 │                │                    │ scan loop           │
 │                │                    │ []AuditEntry        │
 │                │◀───────────────────┤                     │
 │                │ flagJSON ? printJSON : printHuman        │
 │                │ AddAuditLog("audit.read", ...)           │
 │                ├───────────────────▶│                     │
 │                │                    │ db.Exec(...)        │
 │                │                    ├────────────────────▶│
 │                │                    │◀─── OK ─────────────│
 │                │◀───────────────────┤                     │
 │ JSON output    │                    │                     │
 │◀───────────────┤                    │                     │
 │ exit 0         │                    │                     │
```

---

## 14. Related & Cross-Sprint References

- **Sprint 2 `vault-v2-migration`** — `002_audit_log_v2` migration adds `actor` column to existing v1 vaults (backfill); also introduces `resource_name_hmac` + `encrypted_details` (P0-V1)
- **Sprint 6 `documentation-migration`** — `docs/migration/exit-codes.md` (NEW in PR #11) becomes part of the v1→v2 migration guide
- **Sibling `passwd-verify.design.md`** — uses new audit action `vault.passwd_failed` + new exit code 4 (AUTH_FAILED) — both defined here
- **Sibling `crypto-v2-keys.design.md`** — uses HKDF info `tene/audit/v1` (PurposeAudit constant) — but only s2's `002_audit_log_v2` enables HMAC; s1 uses unencrypted columns
- **External dependency** — `modernc.org/sqlite` v1.39.0 — `?` placeholder support (already in use)

---

## Version History

| Version | Date | Changes | Author |
|---------|------|---------|--------|
| 0.1 | 2026-05-13 | Initial draft (Sprint 1 design phase, L2 boundary) | cto-lead (frontend-architect + security-architect perspectives) |
