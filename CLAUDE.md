# Tene -- Agentic Secret Runtime

## Overview

Local-first encrypted secret management CLI built in Go.
AI agents (Claude Code) auto-detect secrets via CLAUDE.md generation.
No server, no cloud dependency, no signup required for MVP.

## Architecture

- Language: Go 1.22+
- CLI framework: cobra (spf13/cobra)
- DB: modernc.org/sqlite (pure Go, no CGo)
- Crypto: golang.org/x/crypto (XChaCha20-Poly1305, Argon2id, HKDF)
- Keychain: zalando/go-keyring (macOS Keychain, Linux libsecret, Windows Credential Vault)
- Recovery: tyler-smith/go-bip39 (12-word BIP-39 mnemonic)
- Build: goreleaser + Homebrew tap
- Test: Go testing + stretchr/testify
- Lint: golangci-lint

## Directory Structure

```
cmd/tene/              CLI entrypoint (main.go)
internal/crypto/       Argon2id KDF, XChaCha20-Poly1305 encrypt/decrypt, key derivation
internal/vault/        SQLite vault CRUD, schema, migrations
internal/keychain/     OS keychain integration + file fallback
internal/claudemd/     CLAUDE.md generation and merge
internal/recovery/     BIP-39 mnemonic generation and master key recovery
internal/cli/          Cobra command definitions (one file per command)
apps/web/              Next.js landing page (independent npm project)
docs/                  PDCA documents (PM, Plan, Design, Analysis, Report)
```

## Development

```bash
# Build
go build -o tene ./cmd/tene

# Run tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Lint
golangci-lint run

# Run specific package tests
go test ./internal/crypto/...
go test ./internal/vault/...
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
| `tene sync` | Fake Door -- shows cloud waitlist (not implemented) |

## Key Decisions

- No server, no cloud dependency (MVP Phase 1)
- XChaCha20-Poly1305 + Argon2id encryption (golang.org/x/crypto)
- Claude Code only (no Cursor/Windsurf in MVP, Phase 2)
- 12-word BIP-39 recovery key
- modernc.org/sqlite (pure Go, no CGo) for cross-compilation
- goreleaser for multi-platform binaries + Homebrew tap
- Fake Door: `tene sync` shows waitlist to validate cloud demand
- `tene export --encrypted` for manual backup until cloud

## Coding Conventions

- Go standard naming: camelCase for unexported, PascalCase for exported
- Errors: wrap with `fmt.Errorf("context: %w", err)`, define sentinel errors per package
- All public functions must have godoc comments
- Table-driven tests preferred
- No global state -- pass dependencies via struct fields
- internal/ packages are not importable from outside the module

## Security Model

- Master Password -> Argon2id KDF (64MB, 3 iterations) -> Master Key (256-bit)
- Master Key -> HKDF -> Encryption Key (for XChaCha20-Poly1305)
- Master Key cached in OS Keychain (go-keyring)
- 192-bit random nonce per encryption, key name as AAD
- Recovery: BIP-39 mnemonic -> Argon2id -> Recovery Key -> decrypt stored Master Key
- Zero network communication in MVP

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

## Landing Page (apps/web/)

Separate Next.js project. Do not mix with Go CLI code.
See `apps/web/CLAUDE.md` for frontend conventions.
