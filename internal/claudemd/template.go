package claudemd

const SectionHeader = "# Secrets Management"

const SecretsMdTemplate = `# Secrets Management — tene

This project uses [tene](https://github.com/tene-ai/tene) for secret management.
Secrets are encrypted locally with XChaCha20-Poly1305. No server, no cloud.

## Quick Reference

| Task | Command |
|------|---------|
| List available secrets | ` + "`tene list`" + ` |
| Get a secret value | ` + "`tene get <KEY>`" + ` |
| Get secret as JSON | ` + "`tene get <KEY> --json`" + ` |
| Set a secret | ` + "`tene set <KEY> <VALUE>`" + ` |
| Set from stdin | ` + "`echo \"value\" | tene set <KEY> --stdin`" + ` |
| Run with secrets injected | ` + "`tene run -- <command>`" + ` |
| Import from .env | ` + "`tene import .env`" + ` |
| Export to .env | ` + "`tene export`" + ` |
| Switch environment | ` + "`tene env <name>`" + ` |
| List environments | ` + "`tene env list`" + ` |

## Setup

Run ` + "`tene init`" + ` to create the vault, set a master password, generate a recovery key, and produce this context file.

## Rules

1. **Never hardcode secrets** in source code, config files, or comments
2. **Never create .env files** — use ` + "`tene run -- <command>`" + ` to inject secrets as environment variables
3. **Access secrets via environment variables** — ` + "`process.env.KEY_NAME`" + ` (Node.js), ` + "`os.Getenv(\"KEY_NAME\")`" + ` (Go), ` + "`os.environ[\"KEY_NAME\"]`" + ` (Python)
4. **Use ` + "`tene list`" + ` first** to check what secrets are available before asking the user
5. **If a secret is missing**, tell the user to run: ` + "`tene set <KEY> <VALUE>`" + `
6. **Use ` + "`--json`" + ` flag** for programmatic parsing: ` + "`tene get KEY --json`" + ` returns ` + "`{\"name\":\"KEY\",\"value\":\"...\",\"environment\":\"...\"}`" + `
7. **Environment isolation**: secrets are scoped per environment (default, dev, staging, prod). Use ` + "`--env <name>`" + ` to target a specific environment
8. **Never run ` + "`tene get <KEY>`" + `** to inspect secret values — the plaintext value appears in stdout and enters the AI context window. Use ` + "`tene run -- <command>`" + ` instead.
9. **Never run ` + "`tene export`" + `** — it outputs all plaintext secrets to stdout. If you need to check which secrets exist, use ` + "`tene list`" + ` (shows names only).
10. **ALWAYS use ` + "`tene run -- <command>`" + `** to inject secrets into development commands. This is the primary workflow — secrets are injected as environment variables and never appear in stdout.
11. **NEVER read or cat files in ` + "`.tene/`" + ` directory** — it contains the encrypted vault database. Even encrypted data should not enter AI context.
12. **If you need a secret value for debugging**, tell the user: "Please check the value with ` + "`tene get KEY`" + ` in a separate terminal." Do not run it yourself.
13. **NEVER pass secret values as command-line arguments** — use environment variables via ` + "`tene run --`" + ` instead. Command arguments appear in process listings and shell history.
14. **Use ` + "`tene list`" + ` to check available secret names** — it only shows names, not values. This is safe to run in AI context.

## Available Environments

Run ` + "`tene env list`" + ` to see all environments and which is active.

## Example Workflows

### Starting a new feature that needs an API key
1. Check if key exists: ` + "`tene list`" + `
2. If not: ask user to run ` + "`tene set STRIPE_KEY sk_test_xxx`" + `
3. Use in code via ` + "`process.env.STRIPE_KEY`" + `
4. Run/test with: ` + "`tene run -- npm test`" + `

### Running the project
` + "```bash" + `
tene run -- npm start          # Node.js
tene run -- go run .           # Go
tene run -- python main.py     # Python
` + "```" + `

## Resources

- Concise agent-readable summary: https://tene.sh/llms.txt
- Extended reference (all commands, security model, FAQ): https://tene.sh/llms-full.txt
- Repository: https://github.com/tene-ai/tene
- Website: https://tene.sh
`

// CursorMdcFrontmatter is the frontmatter for .mdc files used by Cursor.
const CursorMdcFrontmatter = `---
description: Secret management with tene
globs:
alwaysApply: true
---
`

// AgentFile describes an AI editor context file.
type AgentFile struct {
	Name   string // e.g. "claude", "cursor"
	Path   string // relative path from project root
	Format string // "markdown" or "mdc"
}

// AgentFiles lists all supported AI editor context files.
var AgentFiles = []AgentFile{
	{Name: "claude", Path: "CLAUDE.md", Format: "markdown"},
	{Name: "cursor", Path: ".cursor/rules/tene.mdc", Format: "mdc"},
	{Name: "windsurf", Path: ".windsurfrules", Format: "markdown"},
	{Name: "gemini", Path: "GEMINI.md", Format: "markdown"},
	{Name: "codex", Path: "AGENTS.md", Format: "markdown"},
}
