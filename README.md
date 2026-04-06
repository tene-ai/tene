# Tene

**Secret management that AI agents understand.**

Tene is a local-first, encrypted secret management CLI. It stores your API keys, tokens, and credentials in an encrypted SQLite vault on your device — no server, no signup, no cloud dependency.

When you run `tene init`, it generates a `CLAUDE.md` file so Claude Code automatically knows how to use your secrets.

## Why Tene?

- **No server** — secrets never leave your device. Nothing to hack.
- **No signup** — set a master password and start. No account, no email.
- **AI-native** — Claude Code auto-detects tene via generated `CLAUDE.md`
- **Encrypted** — XChaCha20-Poly1305 + Argon2id. OS keychain for master key.
- **Single binary** — built in Go, ~10MB, no runtime needed.

## Install

```bash
# Go users (recommended)
go install github.com/tomo-kay/tene/cmd/tene@latest

# Or download binary from GitHub Releases
# https://github.com/tomo-kay/tene/releases

# Or build from source
git clone https://github.com/tomo-kay/tene.git
cd tene && go build -o tene ./cmd/tene
```

## Quick Start

```bash
# 1. Initialize — creates encrypted vault + CLAUDE.md
tene init

# 2. Store secrets
tene set STRIPE_KEY sk_test_51Hxxxxx
tene set OPENAI_API_KEY sk-proj-xxxxx

# 3. Run with secrets injected as environment variables
tene run -- claude

# That's it. Claude Code reads CLAUDE.md and knows how to use tene.
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
| `tene init` | Create vault, set master password, generate CLAUDE.md |
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

### AI Agent Usage

Claude Code can call tene directly from bash:

```bash
# Get a single secret
STRIPE_KEY=$(tene get STRIPE_KEY)

# JSON output for parsing
tene get STRIPE_KEY --json

# List all available secrets
tene list --json
```

### Migrate from .env

```bash
tene import .env
# ✓ 5 secrets imported (encrypted)
# Tip: You can now delete .env and use tene run instead.
```

## What Tene Does / Doesn't Do

### Does

- Store secrets locally with XChaCha20-Poly1305 encryption
- Inject secrets as environment variables via `tene run`
- Generate CLAUDE.md so Claude Code auto-detects your secrets
- Support multiple environments (dev, staging, prod)
- Provide encrypted backup via `tene export --encrypted`

### Doesn't (yet)

- Check API key expiration dates
- Auto-rotate secrets
- Sync across devices (cloud sync is being validated)
- Share secrets with team members

## Comparison

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
