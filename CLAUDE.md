# Tene -- CLI Secret Manager (Open Core)

## Overview

Local-first encrypted secret management CLI with optional cloud sync.
Open Core model: this public repo contains the CLI + shared packages (`pkg/`).
The private `tene-cloud` repo imports `pkg/` for the API server and dashboard.

## Rules (detailed guides)

- [Coding Conventions](.claude/rules/conventions.md) -- Go coding rules, linting, Git strategy
- [Secret Management](.claude/rules/secrets.md) -- tene secret injection, environment config
- [Git Workflow](.claude/rules/git-workflow.md) -- branch strategy, PR rules, anti-patterns

## Architecture

- Language: Go 1.25+
- CLI framework: cobra (spf13/cobra)
- Local DB: modernc.org/sqlite (pure Go, no CGo)
- Crypto: golang.org/x/crypto (XChaCha20-Poly1305, Argon2id, HKDF, X25519)
- Keychain: zalando/go-keyring (macOS Keychain, Linux libsecret, Windows Credential Vault)
- Recovery: tyler-smith/go-bip39 (12-word BIP-39 mnemonic)
- Build: goreleaser + Homebrew tap
- Test: Go testing + stretchr/testify
- Lint: golangci-lint v2

## Directory Structure

```
cmd/tene/              CLI entrypoint (main.go)
pkg/domain/            Domain models (User, Vault, Team, Device) -- shared with tene-cloud
pkg/crypto/            XChaCha20-Poly1305, Argon2id, HKDF, X25519 -- shared with tene-cloud
pkg/errors/            Sentinel errors and error codes -- shared with tene-cloud
internal/cli/          Cobra commands (25+ commands)
internal/vault/        SQLite vault CRUD, schema, migrations
internal/keychain/     OS keychain integration + file fallback
internal/sync/         Sync Envelope, push/pull engine, conflict resolution, 3-way merge
internal/config/       CLI + global config management
internal/recovery/     BIP-39 mnemonic generation and master key recovery
internal/claudemd/     CLAUDE.md auto-generation (5 AI editors)
internal/encfile/      Encrypted file format (header + ciphertext)
apps/web/              Next.js landing page (tene.sh, Vercel)
```

## Shared Packages (pkg/)

Packages under `pkg/` are the public API of this module, imported by `github.com/tene-ai/tene-cloud`:

- `pkg/domain` -- Domain models and types
- `pkg/crypto` -- Encryption, key derivation, key wrapping
- `pkg/errors` -- Structured error codes with recovery hints

**Do not break the public API of pkg/ without coordinating with tene-cloud.**

## Development

```bash
# Build
go build ./cmd/tene

# Run tests
go test ./...

# Lint
golangci-lint run
```

## Key Commands

| Command | Purpose |
|---------|---------|
| `tene init` | Create vault + master password + recovery key + CLAUDE.md |
| `tene set KEY VALUE` | Encrypt and store a secret |
| `tene get KEY` | Decrypt and output a secret |
| `tene run -- CMD` | Inject all secrets as env vars and run command |
| `tene list` | List secret keys (values masked) |
| `tene delete KEY` | Delete a secret |
| `tene import .env` | Bulk import from .env file |
| `tene export` | Export as .env format |
| `tene export --encrypted` | Export encrypted vault backup |
| `tene env [name]` | Switch environment (dev/staging/prod) |
| `tene passwd` | Change master password + re-encrypt vault |
| `tene recover` | Restore master password via 12-word recovery key |
| `tene push` | Encrypt vault with Sync Envelope and upload to cloud |
| `tene pull` | Download and decrypt remote vault |
| `tene sync` | Push + Pull combined (requires Pro plan) |
| `tene login` | OAuth login to Tene Cloud |
| `tene logout` | Sign out and revoke tokens |
| `tene team create` | Create team + generate project key |
| `tene team invite` | Invite member with X25519 key wrapping |
| `tene team remove` | Remove member + trigger key rotation |
| `tene billing` | View subscription status |
| `tene billing upgrade` | Open LemonSqueezy checkout |

## Key Decisions

- Zero-Knowledge: server never sees plaintext secrets (4-layer encryption)
- XChaCha20-Poly1305 + Argon2id + HKDF + X25519 ECDH (golang.org/x/crypto)
- Sync Envelope: L1 (secret values) + L2 (metadata) + L3 (TLS) + L4 (S3 SSE)
- Team key sharing via X25519 ECDH (no RSA), key rotation on member removal
- modernc.org/sqlite for local vault (pure Go, no CGo)
- goreleaser for multi-platform binaries + Homebrew tap
- Git branch: main -> prod auto-deploy, feature/* -> PR -> preview

## Coding Conventions

- Go standard naming: camelCase for unexported, PascalCase for exported
- Errors: wrap with `fmt.Errorf("context: %w", err)`, define sentinel errors per package
- All public functions must have godoc comments
- Table-driven tests preferred
- No global state -- pass dependencies via struct fields
- internal/ packages are not importable from outside the module
- pkg/ packages are the shared public API -- maintain backward compatibility

## Security Model

- Master Password -> Argon2id KDF (64MB, 3 iterations) -> Master Key (256-bit)
- Master Key -> HKDF -> Encryption Key, SyncKey, DeviceKey, AuthHash
- Master Key cached in OS Keychain (go-keyring)
- 192-bit random nonce per encryption, key name as AAD
- Recovery: BIP-39 mnemonic -> Argon2id -> Recovery Key -> decrypt stored Master Key
- Cloud: Zero-Knowledge Sync Envelope (L2 wraps entire vault.db before upload)
- Team: X25519 ECDH shared secret -> HKDF(projectID) -> wrap Project Key per member

## Project Data

```
Per-project:
  .tene/vault.db       Encrypted secrets + metadata + audit log
  .tene/vault.json     Project metadata
  .tene/.gitignore     Auto-exclude from git
  CLAUDE.md            Auto-generated Claude Code context

Global:
  ~/.tene/config.json  CLI settings, analytics
```
