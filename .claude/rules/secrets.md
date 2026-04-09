# Tene Secret Management (Dogfooding)

Tene manages its own secrets using itself. Zero hardcoded credentials in the codebase.

## Secret Injection

```bash
# Run any command with secrets injected as env vars
tene run --env local -- <command>

# Check what secrets are available
tene list --env local
```

## Security Rules for AI Agents

From `internal/claudemd/template.go`:
1. Never hardcode secrets in source code
2. Never create .env files
3. Access secrets via `tene run -- <command>`
4. Never run `tene get <KEY>` in AI conversations (values may leak)
5. Never run `tene export` (outputs all plaintext to stdout)

## .gitignore Rules

```gitignore
# Build binaries only (not source dirs!)
/tene           # not cmd/tene/
/tene-dev

# Secret-related
.tene/          # local vault directory
.env
.env.*
*.tene.enc
```
