# Tene

[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.22+-00ADD8.svg)](https://go.dev)
[![Version](https://img.shields.io/github/v/release/tomo-kay/tene?color=green)](https://github.com/tomo-kay/tene/releases)
[![CI](https://github.com/tomo-kay/tene/actions/workflows/ci.yml/badge.svg)](https://github.com/tomo-kay/tene/actions/workflows/ci.yml)
[![Author](https://img.shields.io/badge/Author-monsa-purple)](https://github.com/tomo-kay/tene)

<p align="center">
  <img src="branding/OG_image.png" alt="Tene — Secret management that AI agents understand" width="800">
</p>

**Your .env is not a secret. AI can read it.** | [Website](https://tene.sh) | [Releases](https://github.com/tomo-kay/tene/releases)

Tene is a local-first, encrypted secret management CLI. It encrypts your secrets and injects them at runtime — so AI agents can use them without ever seeing the values.

### Supported Platforms

| Platform | Architecture | Status |
|----------|-------------|:------:|
| macOS | Apple Silicon (arm64) | ✓ |
| macOS | Intel (amd64) | ✓ |
| Linux | x86_64 (amd64) | ✓ |
| Linux | ARM (arm64) | ✓ |
| Windows | x86_64 (via WSL) | ✓ |

## Why Tene?

### .env files are not secrets

Every AI coding agent — Claude Code, Cursor, Windsurf — reads your project files. That includes `.env`. Your API keys, database passwords, and tokens are sent to AI models as plaintext context.

```
  .env (plaintext)              AI Agent
  ┌──────────────────┐         ┌──────────────────────┐
  │ STRIPE_KEY=sk_xx │────────>│ Reads all project     │
  │ DB_PASS=s3cur3   │────────>│ files including .env  │
  └──────────────────┘         └──────────────────────┘
```

### Tene keeps secrets from AI

Tene stores secrets in an encrypted SQLite vault. When you run `tene run -- claude`, secrets are injected as environment variables at runtime. The AI agent never sees the actual values.

```
  .tene/vault.db (encrypted)    tene run -- claude
  ┌──────────────────┐         ┌──────────────────────┐
  │ ████████████████ │───X───> │ Secrets injected as   │
  │ (XChaCha20-Poly) │         │ env vars at runtime   │
  └──────────────────┘         │ AI sees: tene run     │
                               │ AI knows: nothing     │
                               └──────────────────────┘
```

### Free locally. $1/mo for cloud.

Local CLI is free forever — unlimited secrets, XChaCha20-Poly1305 encryption, OS keychain integration. Cloud sync ($1/user/month) eliminates repeated `tene init` + `tene set` across projects and machines. (Coming soon)

## Install

```bash
curl -sSfL https://tene.sh/install.sh | sh
```

Auto-detects your OS and architecture, downloads the latest binary from GitHub Releases.

### Other methods

<details>
<summary>With Go</summary>

```bash
go install github.com/tomo-kay/tene/cmd/tene@latest
```

</details>

<details>
<summary>Download binary manually</summary>

Download from [GitHub Releases](https://github.com/tomo-kay/tene/releases), then:

```bash
tar xzf tene_*.tar.gz
sudo mv tene /usr/local/bin/
```

</details>

<details>
<summary>Build from source</summary>

```bash
git clone https://github.com/tomo-kay/tene.git
cd tene && go build -o tene ./cmd/tene
sudo mv tene /usr/local/bin/
```

</details>

## Quick Start

```bash
# 1. Initialize — creates encrypted vault + CLAUDE.md
$ tene init

  Welcome to Tene! Let's set up your local secret vault.
  Master Password: ********
  Confirm: ********

  ✓ .tene/vault.db created (local encrypted vault)
  ✓ Generated CLAUDE.md, .cursor/rules/tene.mdc, .windsurfrules, GEMINI.md, AGENTS.md
  ✓ .tene/ added to .gitignore

  Recovery Key (write this down and keep it safe!):
  +--------------------------------------------------+
  |   apple banana cherry dolphin eagle frost          |
  |   grape harbor island jungle kite lemon            |
  +--------------------------------------------------+

# 2. Store secrets
$ tene set STRIPE_KEY sk_test_51Hxxxxx
  STRIPE_KEY saved (encrypted, default)

$ tene set OPENAI_API_KEY sk-proj-xxxxx
  OPENAI_API_KEY saved (encrypted, default)

# 3. Run with secrets injected as environment variables
$ tene run -- claude
  ✓ 2 secrets injected as environment variables
  ✓ Starting: claude

# That's it. AI editors read the generated files and know how to use tene.
```

## How It Works

```
Master Password
  └─ Argon2id (64MB memory, 3 iterations)
     └─ Master Key (256-bit) → OS Keychain
        └─ XChaCha20-Poly1305 (192-bit nonce)
           └─ SQLite vault (.tene/vault.db)

Network calls: none
Server: none
Attack surface: none
```

Your secrets are encrypted locally with XChaCha20-Poly1305. The master key is derived from your password via Argon2id and cached in the OS keychain (macOS Keychain, Linux libsecret, Windows Credential Vault). A 12-word BIP-39 recovery key is issued during `tene init`.

## Commands

| Command | Description |
|---------|-------------|
| `tene init` | Create vault, set master password, generate AI agent context files |
| `tene set KEY VALUE` | Encrypt and store a secret |
| `tene get KEY` | Decrypt and print a secret to stdout |
| `tene run -- CMD` | Inject secrets as env vars, run command |
| `tene list` | List secret names (values masked) |
| `tene delete KEY` | Delete a secret |
| `tene import .env` | Import secrets from a .env file |
| `tene export` | Export secrets as .env format |
| `tene export --encrypted` | Export encrypted vault backup (.tene.enc) |
| `tene env [name]` | Switch environment (dev/staging/prod) |
| `tene passwd` | Change master password, re-encrypt vault |
| `tene recover` | Recover vault with 12-word recovery key |
| `tene whoami` | Show current vault status |
| `tene sync` | Cloud sync waitlist (coming soon) |
| `tene version` | Print version number |
| `tene update` | Update to latest version (or `tene update v0.2.0`) |

### Global Flags

| Flag | Description |
|------|-------------|
| `--json` | JSON output (for AI agents and scripting) |
| `--env <name>` | Target environment (default: active) |
| `--quiet` | Minimal output (errors only) |
| `--no-keychain` | Skip OS keychain (for CI/CD) |
| `--no-color` | Disable color output |

### Supported AI Editors

`tene init` auto-generates context files for all major AI editors:

| AI Editor | Generated File | Format |
|-----------|---------------|--------|
| Claude Code | `CLAUDE.md` | Markdown |
| Cursor | `.cursor/rules/tene.mdc` | MDC (frontmatter) |
| Windsurf | `.windsurfrules` | Markdown |
| Gemini / Jules | `GEMINI.md` | Markdown |
| Codex / OpenAI | `AGENTS.md` | Markdown |

Each file contains a complete guide: 10 commands, 7 rules, example workflows. The AI editor reads it and knows how to use tene automatically.

### AI Agent Usage

Any AI agent can call tene directly from bash:

```bash
# Get a single secret
STRIPE_KEY=$(tene get STRIPE_KEY)

# JSON output for programmatic parsing
tene get STRIPE_KEY --json
# → {"name":"STRIPE_KEY","value":"sk_test_xxx","environment":"default"}

# List all available secrets
tene list --json
# → {"ok":true,"count":3,"secrets":[...]}

# JSON error output (for error handling)
tene get NONEXISTENT --json
# → {"ok":false,"error":"SECRET_NOT_FOUND","message":"..."}
```

### Detailed Command Usage

#### Set secrets

```bash
# Basic set
tene set STRIPE_KEY sk_test_51Hxxxxx

# Update an existing secret (requires --overwrite)
tene set STRIPE_KEY sk_live_NEW999 --overwrite

# Read value from stdin (avoids shell history)
echo "sk_test_xxx" | tene set STRIPE_KEY --stdin

# Set in a specific environment
tene set DATABASE_URL postgres://prod-host/db --env prod
```

#### Environment management

```bash
# List environments
tene env list
#   * default (active, 3 secrets)
#     staging (1 secret)
#     prod (2 secrets)

# Create a new environment
tene env create staging

# Switch active environment
tene env staging

# Delete an environment
tene env delete staging --force

# Set/get in a specific environment without switching
tene set API_KEY xxx --env prod
tene get API_KEY --env prod
tene run --env prod -- node server.js
```

#### Backup and restore

```bash
# Export as encrypted backup
tene export --encrypted --file backup.tene.enc

# Restore from encrypted backup (on another machine)
tene init
tene import backup.tene.enc --encrypted
```

#### Change master password

```bash
tene passwd
# Enter current password, set new password
# All secrets are re-encrypted with new key
# New 12-word recovery key is issued
```

#### Recover vault (forgot password)

```bash
tene recover
# Enter 12-word recovery key
# Set new master password
# All secrets remain intact
```

### Migrate from .env

```bash
tene import .env
# ✓ 5 secrets imported (encrypted)
# Tip: You can now delete .env and use tene run instead.
```

<p align="center">
  <img src="branding/tene_core_point.png" alt="Tene Features" width="800">
</p>

## What Tene Does / Doesn't Do

### Does

- Store secrets locally with XChaCha20-Poly1305 encryption
- Inject secrets as environment variables via `tene run`
- Generate context files for 5 AI editors (Claude Code, Cursor, Windsurf, Gemini, Codex)
- Support multiple environments (dev, staging, prod)
- Provide encrypted backup via `tene export --encrypted`
- Memory zeroing — master keys cleared from memory after use
- Structured error codes (`--json` error responses for AI parsing)
- Self-update via `tene update`

### Doesn't (yet)

- Check API key expiration dates
- Auto-rotate secrets
- Sync across devices (cloud sync is being validated)
- Share secrets with team members

## Comparison

<p align="center">
  <img src="branding/tene_compares.png" alt="How Tene compares" width="800">
</p>

| | Tene | .env | Doppler | Vault | Infisical |
|---|:---:|:---:|:---:|:---:|:---:|
| Local-first | ✓ | ✓ | ✗ | ✗ | ✗ |
| No server | ✓ | ✓ | ✗ | ✗ | ✗ |
| Encrypted | ✓ | ✗ | ✓ | ✓ | ✓ |
| AI auto-detect | ✓ | ✗ | ✗ | ✗ | ✗ |
| No signup | ✓ | ✓ | ✗ | ✗ | ✗ |
| 100% offline | ✓ | ✓ | ✗ | ✗ | ✗ |
| Open source | ✓ | ✓ | ✗ | ✗ | ✓ |
| Price | Free | Free | $21/user/mo | $1,152+/mo | $6/user/mo |

## Security

- **Encryption**: XChaCha20-Poly1305 (256-bit keys, 192-bit nonces)
- **Key derivation**: Argon2id (64MB memory, 3 iterations)
- **Key storage**: OS native keychain
- **Recovery**: 12-word BIP-39 mnemonic
- **Zero network**: no calls, no telemetry, no phone home
- **Open source**: every line of crypto code is auditable

Tene has no server. There is no database to breach, no API to exploit, no cloud to compromise. Your secrets exist only on your device.

## CI/CD Usage

Use `TENE_MASTER_PASSWORD` environment variable and `--no-keychain` flag for non-interactive environments:

```bash
# GitHub Actions example
env:
  TENE_MASTER_PASSWORD: ${{ secrets.TENE_MASTER_PASSWORD }}

steps:
  - run: tene get DATABASE_URL --no-keychain
  - run: tene run --no-keychain -- npm test
```

```bash
# Docker / CI script
export TENE_MASTER_PASSWORD="your-password"
tene get API_KEY --no-keychain --json
```

## Built With

- [Go](https://go.dev) — single binary, cross-platform
- [cobra](https://github.com/spf13/cobra) — CLI framework
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — pure Go SQLite
- [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) — XChaCha20-Poly1305, Argon2id, HKDF
- [go-keyring](https://github.com/zalando/go-keyring) — OS keychain
- [go-bip39](https://github.com/tyler-smith/go-bip39) — recovery key mnemonic

## Contributing

Tene is open source under the [MIT License](LICENSE).

```bash
git clone https://github.com/tomo-kay/tene.git
cd tene
go build -o tene ./cmd/tene
go test ./...
golangci-lint run
```

## License

MIT
