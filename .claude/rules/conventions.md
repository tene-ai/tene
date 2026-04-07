# Tene Coding Conventions

## Go Conventions

### Naming
- **Exported**: PascalCase (`DeriveSubKey`, `NewVault`)
- **Unexported**: camelCase (`loadApp`, `resolveEnv`)
- **Packages**: lowercase, single word (`crypto`, `vault`, `sync`)
- **Test files**: `*_test.go` in same package
- **Error constructors**: `New*` pattern (`NewSecretNotFound(key, env)`)

### Error Handling
```go
// Always wrap with context
return fmt.Errorf("vault: set secret: %w", err)

// Sentinel errors per package
var ErrNotFound = errors.New("not found")

// CLI errors with codes and recovery hints
var ErrVaultNotFound = &TeneError{
    Code:    "VAULT_NOT_FOUND",
    Message: "Not in a Tene project. Run \"tene init\" first.",
    Exit:    1,
}
```

### Unchecked Returns (errcheck)
```go
// defer Close — use closure pattern
defer func() { _ = app.Vault.Close() }()

// fmt.Fprint* in CLI output — assign to blank
_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "  Done\n")

// Non-critical operations
_ = os.Remove(tempFile)
```

### Error String Style (staticcheck ST1005)
```go
// Correct: lowercase, no trailing punctuation
return fmt.Errorf("cannot create directory: %w", err)

// Wrong:
return fmt.Errorf("Cannot create directory: %w", err)  // capitalized
return fmt.Errorf("no secrets found.")                   // trailing period
```

### Dependencies
- No global mutable state — pass via struct fields
- CLI flags are the exception (Cobra convention)
- HTTP clients: package-level singletons (`var cliHTTPClient`)
- All handler deps injected via constructors

### Testing
- Table-driven tests preferred
- Test helpers in `testhelper_test.go`
- Use `t.TempDir()` for isolation
- Error returns in tests: `require.NoError(t, err)` or `_ = x()`

## Linting — golangci-lint v2

Config: `.golangci.yml` (version: "2")

**Enabled linters:**
- `errcheck` — unchecked error returns (check-type-assertions: true)
- `govet` — suspicious constructs
- `ineffassign` — useless assignments
- `staticcheck` — advanced static analysis
- `unused` — unused code
- `misspell` — common misspellings
- `exhaustive` — exhaustive switch (default-signifies-exhaustive: true)

**CI**: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.1.6`
(Source-built with Go 1.25 — pre-built binaries use Go 1.24)

## Frontend Conventions (Next.js 15)

### Design System (Dark-only)
```css
--background: #0a0a0a;
--foreground: #ededed;
--accent: #00ff88;         /* neon green */
--accent-dim: #00cc6a;
--surface: #141414;
--surface-2: #1e1e1e;
--border: #2a2a2a;
--muted: #888888;
--danger: #ff4444;         /* dashboard only */
--warning: #ffaa00;        /* dashboard only */
```

### Fonts
- Primary: Geist Sans (via next/font/google)
- Mono: Geist Mono
- Fallback: Arial, Helvetica, sans-serif

### Security Headers (both apps)
```typescript
// next.config.ts
const apiUrl = process.env.NEXT_PUBLIC_API_URL || "https://api.tene.sh";
// HSTS, X-Frame-Options: DENY, CSP (dynamic connect-src), Permissions-Policy
```

### Environment Variables
```
NEXT_PUBLIC_API_URL       API endpoint (default: https://api.tene.sh)
NEXT_PUBLIC_DASHBOARD_URL Dashboard URL (local dev only)
```

### Dashboard Tech Stack
- TanStack Query v5 — server state management
- Zustand — auth state (access token)
- Lucide React — icons
- clsx — conditional classes

### Component Patterns
- Server Components by default (App Router)
- Client Components only when needed (`"use client"`)
- Semantic HTML with ARIA attributes
- `<Link>` for internal navigation (not `<a>`)

## Git Conventions

### Branch Strategy
```
main → prod auto-deploy (protected, PRs only)
feature/* → PR → Vercel Preview + CI
hotfix/* → PR → fast merge → main → deploy
```

### Commit Messages
```
feat: new feature
fix: bug fix
fix(ci): CI/CD fix
fix(lint): linter fixes
docs: documentation
refactor: code restructure
```

### CI Pipeline (GitHub Actions)
```
Push to main:
  test (go test -race) + lint (golangci-lint v2) → deploy (Docker → ECR → ECS)

Tag v*:
  GoReleaser → GitHub Releases (multi-platform binaries)

Dashboard changes:
  TypeScript check + build verification → Vercel auto-deploy
```
