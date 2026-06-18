# tene CLI Reference

> Canonical reference for every tene command, flag, exit code, and JSON
> schema. Published at https://tene.sh/cli.

## Global Flags

All commands inherit the following persistent flags from the root:

| Flag | Short | Type | Default | Description |
|------|:-----:|------|---------|-------------|
| `--json` | | bool | `false` | Emit JSON to stdout |
| `--quiet` | `-q` | bool | `false` | Minimal output (errors only) |
| `--env` | `-e` | string | active env | Target environment |
| `--dir` | | string | cwd | Project directory |
| `--no-color` | | bool | `false` | Disable colorized output |
| `--no-keychain` | | bool | `false` | Skip OS keychain (CI/CD) |

## Exit Codes

| Code | Name | Meaning |
|:----:|------|---------|
| `0` | Success | Command completed |
| `1` | GenericError | Uncategorized error |
| `2` | UsageError | Wrong flag / arg shape (Cobra default) |
| `2` | STDOUT_SECRET_BLOCKED | `tene get` on non-TTY without `--unsafe-stdout` / `TENE_ALLOW_STDOUT_SECRETS=1` |
| `3` | VAULT_NOT_FOUND | `.tene/vault.db` missing in the project directory |
| `4` | AUTH_REQUIRED | Master password / keychain failed |
| `5` | SECRET_NOT_FOUND | KEY does not exist in the target environment |
| `6` | DECRYPT_FAILED | Ciphertext + key mismatch |
| `7` | INTERACTIVE_REQUIRED | No password source on non-TTY |
| `127` | COMMAND_NOT_FOUND | `tene run -- <cmd>` target binary not in PATH |

## Commands

### `tene init [project-name]`

Create an encrypted vault and generate AI editor context files.

**Default** (no flags): writes `CLAUDE.md`, `.cursor/rules/tene.mdc`,
`.windsurfrules`, `GEMINI.md`, `AGENTS.md` to the project root.

**Flags**: `--claude`, `--cursor`, `--windsurf`, `--gemini`, `--codex` to
generate a single editor file.

**Side effects**: creates `.tene/vault.db` (encrypted), adds `.tene/` to
`.gitignore`, stores master key in OS keychain (unless `--no-keychain`),
issues a 12-word BIP-39 recovery key.

### `tene set KEY VALUE` · `tene set KEY --stdin`

Encrypt and store a secret.

**Flags**:

- `--stdin` — read value from stdin (avoids shell history).
- `--overwrite` — replace an existing secret.
- `--env <name>` — target a specific environment.

### `tene get KEY`

Decrypt and print a secret.

⚠️ **Security**: refuses non-TTY stdout by default. This prevents accidental
leakage into AI agent context windows, log aggregators, and shell history.

**Flags**:

- `--json` — emit `{ok, name, value, environment}`. Works on non-TTY, but
  prints a one-line warning to stderr when not explicitly opted in.
- `--unsafe-stdout` — explicit per-invocation override for non-TTY stdout.

**Environment override**: `TENE_ALLOW_STDOUT_SECRETS=1` also permits non-TTY
stdout.

**Exit codes**: `0` (printed), `2` (STDOUT_SECRET_BLOCKED), `5` (SECRET_NOT_FOUND).

**Preferred alternative**: `tene run -- <cmd>` — never prints plaintext.

### `tene run -- COMMAND [ARGS...]`

Inject all secrets from the active (or `--env`) environment as environment
variables, then execute the command. Child stdout / stderr / stdin are
inherited; tene's own diagnostic messages go to stderr so the child's
stdout remains clean.

**Exit code**: the child process's exit code is propagated unchanged.

**JSON output** (`--json`): emits a single-line JSON object to **stderr**
(not stdout, because stdout is the child process):
`{"injectedCount":N,"environment":"env","command":"..."}`.

### `tene list`

List secret names in the active (or `--env`) environment. Values are masked
in human output; the `--json` form also returns `preview` fields masked.

**JSON schema** (`tene list --json`):

```json
{
  "ok": true,
  "project": "my-app",
  "environment": "default",
  "count": 3,
  "secrets": [
    {
      "name": "STRIPE_KEY",
      "preview": "*****",
      "version": 1,
      "updatedAt": "2026-04-22T12:00:00Z"
    }
  ]
}
```

**AI-safe**: never returns plaintext values.

### `tene delete KEY`

Delete a secret from the active (or `--env`) environment. Prompts for
confirmation unless `--quiet` is set.

### `tene import .env | backup.tene.enc`

One-shot migration:

- `.env` file → encrypt each `KEY=value` line into the vault.
- `.tene.enc` (encrypted backup from `tene export --encrypted`) → restore.

**Flags**: `--encrypted` (import mode toggles based on file extension but
`--encrypted` forces the encrypted-backup path).

### `tene export [--encrypted] [--file path]`

Export vault contents.

⚠️ **Without `--encrypted`**, this prints every secret as plaintext. Avoid
using in AI context. Use `--encrypted --file backup.tene.enc` to produce a
portable encrypted backup.

### `tene env [name] | env list | env create NAME`

Multi-environment workflow:

- `tene env` — show active environment.
- `tene env list` — list all environments.
- `tene env create staging` — create a new environment.
- `tene env staging` — switch active environment.

### `tene passwd`

Change the master password. Vault is re-encrypted with the new key.

### `tene recover`

Restore the master key from a 12-word BIP-39 mnemonic. Useful if the OS
keychain is wiped or the password is forgotten.

### `tene whoami`

Print the active vault path, environment, and project name.

### `tene version`

Print version, commit, and build date.

### `tene update`

Check for and install the latest release (re-runs `install.sh`).

### `tene completion [bash|zsh|fish|powershell]`

Emit a shell completion script to stdout.

```bash
source <(tene completion bash)                                 # Bash (one-off)
tene completion bash > $(brew --prefix)/etc/bash_completion.d/tene  # Bash (Homebrew)
tene completion zsh > "${fpath[1]}/_tene"                      # Zsh
tene completion fish > ~/.config/fish/completions/tene.fish    # Fish
tene completion powershell | Out-String | Invoke-Expression    # PowerShell
```

## Environment Variables (Input)

| Variable | Purpose |
|----------|---------|
| `TENE_MASTER_PASSWORD` | Non-interactive master password (CI/CD) |
| `TENE_ALLOW_STDOUT_SECRETS` | `=1` or `=true` allows `tene get` on non-TTY |
| `NO_COLOR` | Any non-empty value disables color (also via `--no-color`) |

## JSON Schema Conventions

- **Success envelope**: `{ "ok": true, ...fields }`
- **Error envelope**: `{ "ok": false, "error": "CODE", "message": "..." }`
- **All timestamps**: RFC 3339 (`2026-04-23T12:00:00Z`)
- **Secret values**: only present in `tene get --json`. Every other command
  returns names only.

## Man Page

After Homebrew install (`brew install tene-ai/tene/tene`), the man page is
available:

```bash
man tene
```

## FAQ

**What is tene?**

tene is a local-first encrypted secret manager CLI. It encrypts API keys with XChaCha20-Poly1305 and injects them at runtime via tene run, so AI coding agents never see plaintext values.

**How do I install tene?**

Run curl -sSfL https://tene.sh/install.sh | sh on macOS or Linux. The installer auto-detects your platform, downloads the latest release binary, and places it on your PATH. No Go toolchain or account is required.

**How do I run a command with secrets injected?**

tene run -- `<command>` launches your command with every secret in the active environment exposed as a process-scoped environment variable. The values never appear on stdout or in shell history. Example: tene run -- npm start.

**Can my AI assistant read tene secrets?**

No. tene get refuses to print secrets to a non-TTY stdout (exit code 2, STDOUT_SECRET_BLOCKED) so an AI agent or log pipeline cannot pipe the value out. Use tene run -- `<command>` instead, which keeps secrets in the child process environment only.

**Where are secrets stored?**

Encrypted in a local SQLite vault at .tene/vault.db, scoped per project directory. The master key is held in your OS keychain (Keychain on macOS, libsecret on Linux, Credential Manager on Windows). A 12-word BIP-39 recovery key is issued on tene init for offline backup.

## Resources

- Repository: https://github.com/tene-ai/tene
- LLM index: https://tene.sh/llms.txt
- Full reference: https://tene.sh/llms-full.txt
- Issues: https://github.com/tene-ai/tene/issues
- Discussions: https://github.com/tene-ai/tene/discussions
