# Tene Architecture & Clean Code Guide

## Open Core Structure

```
tene/ (Public, MIT)
├── cmd/tene/              CLI entrypoint (main.go) → goreleaser → S3 + GitHub Releases
├── pkg/                   Shared packages (importable by tene-cloud)
│   ├── domain/            Models: User, Vault, Team, Device, AuditLog
│   ├── crypto/            XChaCha20-Poly1305, Argon2id, HKDF, X25519
│   └── errors/            CLI error codes with recovery suggestions
├── internal/              CLI-only packages (not importable externally)
│   ├── cli/               Cobra commands (25+)
│   ├── vault/             SQLite CRUD, schema, migrations
│   ├── keychain/          OS keychain (macOS/Linux) + file fallback
│   ├── sync/              Push/pull engine, 3-way merge, Sync Envelope
│   ├── config/            CLI + global config management
│   ├── recovery/          BIP-39 mnemonic generation/recovery
│   ├── claudemd/          CLAUDE.md auto-generation (5 AI editors)
│   └── encfile/           Encrypted file format (header + ciphertext)
├── apps/web/              Landing page (Next.js 15) → Vercel (tene.sh)
├── docs/                  PDCA documents (PM, Plan, Design, Report)
└── .claude/rules/         AI agent context rules
```

Cloud code (API server, dashboard, infrastructure) lives in the private `tene-cloud` repo.

## Clean Architecture Layers

```
Domain Layer (no external deps)
  └── pkg/domain/           Models: User, Vault, Team, Device, AuditLog
  └── pkg/errors/           CLI error codes with recovery suggestions

Infrastructure Layer (external libs only)
  └── pkg/crypto/           XChaCha20-Poly1305, Argon2id, HKDF, X25519
  └── internal/encfile/     Encrypted file format (header + ciphertext)
  └── internal/recovery/    BIP-39 mnemonic generation/recovery
  └── internal/keychain/    OS keychain (macOS/Linux) + file fallback
  └── internal/vault/       SQLite CRUD, schema, migrations
  └── internal/config/      CLI + global config management
  └── internal/claudemd/    CLAUDE.md auto-generation (5 AI editors)

Service Layer (business logic)
  └── internal/sync/        Push/pull engine, 3-way merge, Sync Envelope

Interface Layer (user-facing)
  └── internal/cli/         Cobra commands (25+ commands)
```

## Dependency Rules

- `pkg/domain/` → imports ONLY stdlib (`errors`, `time`)
- `pkg/crypto/`, `internal/vault/`, `internal/keychain/` → imports domain + external libs
- `internal/sync/` → imports crypto + domain
- `internal/cli/` → imports vault, sync, crypto, domain, errors, keychain, recovery, config
- **No circular imports.**
- **pkg/ packages must not import internal/ packages.**

## Crypto Architecture

```
Master Password
  → Argon2id (64MB, 3 iter, 4 threads) → Master Key (256-bit)
    → HKDF-SHA256("encryption") → Encryption Key
    → HKDF-SHA256("sync")       → Sync Key
    → HKDF-SHA256("device")     → Device Key
    → HKDF-SHA256("auth")       → Auth Hash

Secret Encryption:
  XChaCha20-Poly1305 (192-bit nonce, key name as AAD)

Sync Envelope:
  Seal: SyncKey + AAD(projectID:env) → XChaCha20-Poly1305(vault.db)
  Magic: "TENV" + version + nonce + ciphertext

Team Key Wrapping:
  X25519 ECDH → shared secret → HKDF(projectID) → wrap Project Key
```

## Sync Flow (Push/Pull/Merge)

### Push
1. Read local `vault.db` → Serialize secrets to JSON
2. Seal with Sync Envelope (SyncKey + AAD `projectID:env`)
3. SHA-256 hash for conflict detection
4. Upload to `POST /vaults/:id/push` with `If-Match` (optimistic locking)

### Pull
1. `GET /vaults/:id/pull` → presigned S3 URL (5-min TTL)
2. Download + verify SHA-256 hash
3. Open Sync Envelope (SyncKey + AAD)
4. 3-way merge: base (last sync) vs local vs remote
5. Update local vault.db

### 3-Way Merge Rules
```
Base=A, Local=B, Remote=A → Local wins (only local changed)
Base=A, Local=A, Remote=B → Remote wins (only remote changed)
Base=A, Local=B, Remote=C → CONFLICT (both changed)
Base=nil, Local=A         → Local addition
Base=nil, Remote=A        → Remote addition
Base=A, Local=nil, Remote=B → CONFLICT (deleted vs modified)
```

## CLI Commands (25+)

| Command | Description |
|---------|-------------|
| `init [name]` | Create vault + master password + recovery key + CLAUDE.md |
| `set KEY [VALUE]` | Encrypt and store (supports --stdin, --overwrite) |
| `get KEY` | Decrypt and output |
| `list` | List keys (values masked) |
| `delete KEY` | Delete (supports --force) |
| `run -- CMD` | Inject secrets as env vars |
| `import FILE` | Bulk import (.env or encrypted) |
| `export` | Export (.env or --encrypted) |
| `env [name]` | Switch environment |
| `passwd` | Change password (atomic 2-phase re-encrypt) |
| `recover` | Restore via 12-word BIP-39 mnemonic |
| `push` | Upload encrypted vault to cloud |
| `pull` | Download and decrypt remote vault |
| `sync` | Push + Pull combined (Pro) |
| `login` | GitHub OAuth to Tene Cloud |
| `logout` | Sign out + revoke tokens |
| `team create/invite/remove/list/members` | Team management |
| `billing` / `billing upgrade` | Subscription management |
| `whoami` | Current user info |
| `version` | Version info |
| `update` | Check + install updates (S3 with GitHub fallback) |

Global flags: `--json`, `--quiet/-q`, `--env/-e`, `--dir`, `--no-color`, `--no-keychain`

## SQLite Schema (Local) — internal/vault/schema.go

| Table | Purpose |
|-------|---------|
| vault_meta | Key-value metadata (salt, recovery_blob) |
| secrets | Encrypted name/value per environment |
| environments | Environment management |
| audit_log | Local operation history |
