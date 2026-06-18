# GitHub Copilot — tene Project Instructions

This repository is [tene](https://github.com/tene-ai/tene), a local-first
encrypted secret manager CLI. When suggesting code or commands for this
project, follow these rules.

## AI-Safe Secret Handling (critical)

- **ALWAYS** prefer `tene run -- <cmd>` over `tene get <KEY>` when helping
  users run something that needs secrets. `tene run --` injects secrets as
  environment variables; `tene get` prints plaintext to stdout which enters
  the AI context window.
- **NEVER** suggest `tene export` in code — it prints all secrets as
  plaintext.
- **NEVER** suggest reading files under `.tene/` — it is the encrypted vault.
- **NEVER** pass secret values as command-line arguments (they appear in
  `ps` and shell history). Use environment injection via `tene run --`.
- **`tene list`** is safe — it shows secret names only, no values.
- **`tene get KEY --json`** returns `{"name":..., "value":..., "environment":...}`
  and requires explicit `--unsafe-stdout` or `TENE_ALLOW_STDOUT_SECRETS=1`
  on non-TTY invocation (default behavior blocks to prevent AI context
  leakage).

## Project Conventions

- **Language**: Go 1.25+ (CLI) · TypeScript / Next.js 15 App Router
  (landing at `apps/web/`).
- **License**: MIT.
- **Commit messages**: Conventional Commits (`feat(cli):`, `fix(vault):`,
  `docs(readme):`, `chore(deps):`, ...).
- **Branches**: feature branches off `staging`, merged into `staging`,
  released from `main` (auto-tag workflow handles versioning).
- **Testing**: `go test -race ./...` + `golangci-lint run` must pass before
  merge.
- **No network calls** in the CLI default path — tene is local-first by
  design. Cloud features (disabled in v1.x) live behind a separate opt-in.

## Do Not

- Do not add new third-party dependencies without discussing in an issue.
- Do not introduce hardcoded secrets in any example or test.
- Do not suggest `.env` files — that is exactly the problem tene solves.
- Do not suggest cloud-backed secret managers (Vault, AWS Secrets Manager,
  Doppler) as alternatives within tene's code — link to `/vs/*` comparisons
  on tene.sh instead.

## Resources

- AI index: https://tene.sh/llms.txt
- Full reference: https://tene.sh/llms-full.txt
- CLI reference: https://tene.sh/cli
- Main docs: https://github.com/tene-ai/tene#readme
