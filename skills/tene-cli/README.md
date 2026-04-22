# tene-cli — Local-First Encrypted Secrets for AI-Native Projects

> ClawHub skill that teaches Claude Code / OpenClaw to safely use the
> [tene](https://tene.sh) CLI for secret management — with baked-in AI
> safety guardrails so agents never leak plaintext into chat.

## Why this skill?

**Tene** is a local-first, zero-knowledge secret manager. Secrets live in an
encrypted SQLite vault in your project's `.tene/` directory. The master key is
derived from your password with Argon2id and cached in the OS keychain. The
`tene run -- <cmd>` pattern injects secrets into child processes as env vars —
they never hit disk, stdout, shell history, or `ps` output.

This skill wraps that philosophy into a drop-in package for AI coding agents:

- **Zero cloud dependency** — tene works offline; the skill reflects that.
- **AI-safe by default** — the skill explicitly forbids `tene get`, plain
  `tene export`, and reading anything in `.tene/`.
- **100% command coverage** — every active command is documented, with flag
  placement rules, key-name regex, env name regex, and exit code map.
- **Real workflows** — three end-to-end examples: project init, migration
  from `.env`, multi-environment CI/CD.

## Quick start

```bash
# 1. Install tene
curl -sSfL https://tene.sh/install.sh | sh

# 2. Initialize a vault in your project
cd my-project
tene init

# 3. Add a secret (value entered interactively — never as an argument)
tene set STRIPE_KEY

# 4. Run any command with secrets injected
tene run -- npm start
```

## How this skill helps your AI agent

After installing the skill, any Claude Code / OpenClaw session in a
tene-enabled project will automatically know to:

- Use `tene list` (not `tene get`) when asked "what secrets are set?"
- Wrap dev/test commands in `tene run --` instead of creating `.env` files
- Refuse to print decrypted values into chat
- Explain key-name rules (`^[A-Z][A-Z0-9_]*$`) and env-name rules
  (`^[a-z][a-z0-9-]*$`) when the user hits a validation error
- Suggest `tene recover` (not a fresh `tene init`) when the user forgets their
  master password

## Commands documented

| Category | Commands |
|---|---|
| Setup | `init`, `passwd`, `recover` |
| Secrets | `set`, `list`, `delete`, `get` (flagged unsafe) |
| Execution | `run` |
| Backup | `import`, `export` (flagged unsafe without `--encrypted`) |
| Environments | `env [list/create/delete]`, `env <name>` |
| Diagnostics | `whoami`, `version`, `update` |

Cloud-side commands (`login`, `push`, `pull`, `sync`, `billing`, `team`) are
currently disabled in the tene CLI and are intentionally not documented here.
They'll be added in a future version of this skill when the CLI re-enables
them.

## Requirements

- macOS or Linux (Windows users can use tene but the curl installer is
  POSIX-only)
- tene v1.0.0 or later (verify with `tene version`)
- ClawHub-compatible agent runtime: Claude Code, OpenClaw, or any tool that
  loads `SKILL.md` files

## License

This skill (documentation only) is published on ClawHub under **MIT-0**
(no attribution required, commercial use allowed). The underlying
[tene CLI](https://github.com/tomo-kay/tene) is **MIT**.

## Links

- Tene homepage: https://tene.sh
- Source: https://github.com/tomo-kay/tene
- ClawHub skill page: https://clawhub.ai/tene-cli _(after publish)_
- Issue tracker: https://github.com/tomo-kay/tene/issues
