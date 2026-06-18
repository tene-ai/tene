# Contributing to Tene

Thanks for considering a contribution! Tene is a local-first encrypted secret
manager CLI. This document covers what we accept, the dev workflow, and the
style we care about.

## Code of Conduct

All participation is governed by [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md).

## Development Setup

### Prerequisites

- Go 1.25+
- `golangci-lint` 2.1+
- `make`
- (Optional) GoReleaser 2.x for release dry-runs

### Clone and build

```bash
git clone https://github.com/tene-ai/tene.git
cd tene
go build -o tene ./cmd/tene
go test -race ./...
golangci-lint run
```

### Run locally in a throwaway directory

```bash
mkdir -p /tmp/tene-dev && cd /tmp/tene-dev
TENE_MASTER_PASSWORD=devpass /path/to/tene init --no-keychain --claude
TENE_MASTER_PASSWORD=devpass /path/to/tene set DEMO_KEY demo_value --no-keychain
TENE_MASTER_PASSWORD=devpass /path/to/tene run --no-keychain -- env | grep DEMO_KEY
```

## Contribution Workflow

1. **Open an issue first** for non-trivial changes — this saves rework.
2. **Fork → branch**: create `fix/short-desc` or `feat/short-desc` off
   `staging`.
3. **Write tests** alongside the code (`*_test.go` in the same package).
4. **Verify locally**:

   ```bash
   go vet ./...
   go test -race ./...
   golangci-lint run
   ```

5. **Commit message**: Conventional Commits preferred
   (`feat(cli):`, `fix(vault):`, `docs(readme):`, `chore(deps):`).
6. **Open a PR to `staging`**. CI must pass.
7. **Security-sensitive** changes: mention `@agent-kay-it` in the PR body for
   dedicated review.

## Code Style

- Run `gofmt -s` + `golangci-lint run` before commit.
- No breaking CLI changes without a major version bump.
- **No new network calls** in the default CLI path — tene is local-first by
  design. Cloud features (disabled in v1.x) live behind an explicit opt-in.
- **Encrypt-at-rest invariants**: any new write path to the vault must go
  through `pkg/crypto/` primitives. Never place plaintext secrets on disk.
- **stdout/stderr discipline**: stdout is for data that callers parse
  (respect `--json`); stderr is for diagnostic and progress text. Secrets
  must never land in stdout unless explicitly opted in.

## Documentation

User-facing changes must update:

- `README.md`
- `apps/web/public/llms.txt` and `apps/web/public/llms-full.txt`
- `docs/cli-reference.md` (command semantics, flags, exit codes, JSON schema)
- Blog or `/vs/*` data if applicable (`apps/web/content/blog/*.mdx`,
  `apps/web/src/data/comparisons/*`)

## Release Process (maintainers only)

Pushing to `staging` auto-creates a `v*-rcN` tag and triggers GoReleaser for
RC builds. Pushing to `main` promotes the latest RC to a stable tag and runs
the full release pipeline (GitHub Release, Homebrew tap bump, Docker GHCR
push, S3 binaries, `LATEST_VERSION` pointer update).

Never manually create tags. The `auto-tag` workflow owns versioning.

## First-PR Ideas

Looking for something small to get started?

- Improve an error message in `pkg/errors/`
- Add a shell-completion smoke test in `internal/cli/completion_test.go`
- Add a new comparison page in `apps/web/src/data/comparisons/`
- Fix a typo in `llms.txt` / `llms-full.txt`
- Translate the `llms.txt` summary (experimental)

## License

By contributing, you agree that your contributions will be licensed under the
MIT License (see [LICENSE](./LICENSE)).
