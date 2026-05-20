# FX2 — Destructive Ops Fail-Closed (env delete + delete)

> **Bugs**: B2 (CRITICAL — data loss), B9 (MEDIUM — `--force` flag missing on `env delete`)
> **Invariant introduced**: I-12
> **Files touched**: `internal/cli/helpers.go`, `internal/cli/env.go`, `internal/cli/delete.go` (none — share existing var), `internal/cli/testhelper_test.go` (flag reset), `internal/cli/env_delete_safety_test.go` (new), `CHANGELOG.md`
> **PR target**: `staging`
> **Commit prefix**: `fix(cli):`

## 1. Problem

QA evidence (`docs/05-qa/tene-cli-v1.0.14-rc1.qa-report.md` RECHECK-1):

```bash
$ tene env create testdel && tene set CANARY_KEY canary_value --env testdel
$ tene env delete testdel    # no prompt, no --force, no TTY
Environment "testdel" deleted (1 secrets removed).   # ← silent data loss
```

Two cooperating defects:

- **B2** `internal/cli/helpers.go:97-100` — `promptConfirm()` defaults to
  *yes* on non-TTY:
  ```go
  func promptConfirm(msg string) bool {
      if !isTerminal() {
          return true // non-interactive defaults to yes
      }
  ```
  Every destructive op (`tene delete KEY`, `tene env delete`,
  `tene audit prune` — wait, audit prune has its own `--force` flag and
  separate prompt path; check helpers usage) uses this helper, so the
  fail-open default leaks into all of them.
- **B9** `internal/cli/env.go:25-30` — `envDeleteCmd` has no `--force`
  flag wired, so even if a user wants to suppress a (future) confirm
  prompt they cannot.

## 2. Design

### 2.1 `promptConfirm` becomes fail-closed (I-12)

Replace the `return true` non-TTY branch with `return false` plus a
one-line stderr explanation so the operator understands the refusal:

```go
func promptConfirm(msg string) bool {
    if !isTerminal() {
        // Sprint v1014-rc1-qa-fixes / FX2 (invariant I-12).
        // Destructive operations default to "no" on non-TTY input so
        // CI/CD pipelines, log-redirected scripts, and AI agent contexts
        // cannot silently consent to data loss. Callers that legitimately
        // want unattended execution must opt in with --force.
        fmt.Fprintln(os.Stderr, "Refusing to confirm a destructive operation on a non-interactive shell.")
        fmt.Fprintln(os.Stderr, "Pass --force to skip the prompt, or run in a terminal.")
        return false
    }
    fmt.Fprintf(os.Stderr, "%s (y/N) ", msg)
    reader := bufio.NewReader(os.Stdin)
    answer, _ := reader.ReadString('\n')
    answer = strings.TrimSpace(strings.ToLower(answer))
    return answer == "y" || answer == "yes"
}
```

### 2.2 `envDeleteCmd` gets its own `--force` flag

Introduce a separate package-level boolean so `tene env delete` does not
inherit the `tene delete KEY` flag state (the two verbs already live in
different cobra subtrees; sharing was an accident of laziness).

```go
var envDeleteFlagForce bool

func init() {
    envCmd.AddCommand(envCreateCmd)
    envCmd.AddCommand(envDeleteCmd)
    envCmd.AddCommand(envListCmd)
    envDeleteCmd.Flags().BoolVar(&envDeleteFlagForce, "force", false, "Skip confirmation prompt")
}
```

Then in `runEnvDelete`, replace the `if !deleteFlagForce` reference with
`if !envDeleteFlagForce`. `testhelper_test.go::resetFlags()` adds
`envDeleteFlagForce = false` to its reset list so per-test isolation
holds.

### 2.3 Concomitant behaviour for `tene delete KEY`

`tene delete KEY` also uses `promptConfirm`. Once `promptConfirm` fails
closed, `tene delete KEY` on a non-TTY without `--force` will also refuse.
This is intentional and matches the same I-12 logic — silent secret
deletion in CI is just as bad as silent env deletion. The CHANGELOG
BREAKING entry covers both verbs.

### 2.4 Invariant I-12

> Destructive operations (`tene delete`, `tene env delete`) require either
> an interactive yes/yes-answer OR an explicit `--force`. Non-TTY without
> `--force` refuses and exits with non-zero, surfacing the refusal to
> stderr.

## 3. Test plan

### 3.1 New file `internal/cli/env_delete_safety_test.go`

| Case | Setup | Expected |
|---|---|---|
| `tene env delete X` (no --force, non-TTY) | env X exists with secrets | exit non-zero; env still present; secret count unchanged |
| `tene env delete X --force` (non-TTY) | env X exists with secrets | exit 0; env gone; secrets gone |
| `tene env delete default` | active env | exit non-zero ("It is the default environment"), --force does NOT bypass |
| `tene env delete <active>` | env switched to it | exit non-zero ("Switch to another first"), --force does NOT bypass |
| `tene delete KEY` (no --force, non-TTY) | KEY exists | exit non-zero; secret still present |
| `tene delete KEY --force` (non-TTY) | KEY exists | exit 0; secret gone |

### 3.2 Manual sandbox replay

```bash
mkdir /tmp/fx2-sb && cd /tmp/fx2-sb
TENE_MASTER_PASSWORD=p tene init proj --no-keychain --claude --quiet
TENE_MASTER_PASSWORD=p tene env create staging
TENE_MASTER_PASSWORD=p tene set CANARY v --env staging
TENE_MASTER_PASSWORD=p tene env delete staging      # expect refusal
TENE_MASTER_PASSWORD=p tene env list                # expect staging still there
TENE_MASTER_PASSWORD=p tene env delete staging --force  # expect success
```

## 4. Acceptance criteria

- [ ] 3.1's 6 test cases pass
- [ ] `go test -race ./...` green
- [ ] `golangci-lint run` 0 issues
- [ ] Cross-repo (tene-cloud) build still green
- [ ] CHANGELOG.md ⚠️ Breaking entry added
