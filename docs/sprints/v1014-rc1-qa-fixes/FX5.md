# FX5 — `tene run --help` short-circuit

> **Bug**: B6 (HIGH — UX regression, most-used verb)
> **Files touched**: `internal/cli/run.go`, `internal/cli/run_help_test.go` (new), `CHANGELOG.md`
> **PR target**: `staging`
> **Commit prefix**: `fix(cli):`

## 1. Problem

```
$ tene run --help
Error: No command specified. Usage: tene run -- <command>
```

Every other tene verb returns help text for `--help`. Only `run` errors
out because the command sets `DisableFlagParsing: true` (to forward
unknown flags to the child process after `--`), which suppresses
cobra's built-in `--help` handling. The custom `parseFlagsBeforeDash`
in run.go handles `--env`, `--json`, `--quiet` but not `--help`, so
`--help` falls through to the "no command specified" guard.

## 2. Design

Before the `extractArgsAfterDash` call in `runRun`, scan the args for
`--help` / `-h` and short-circuit to `cmd.Help()`:

```go
func runRun(cmd *cobra.Command, args []string) error {
    parseFlagsBeforeDash(args)

    // FX5 (B6): with DisableFlagParsing=true we must handle --help
    // / -h ourselves before the "no command specified" guard, which
    // would otherwise report `tene run --help` as a user error
    // instead of showing the help text.
    if hasHelpFlag(args) {
        return cmd.Help()
    }

    cmdArgs := extractArgsAfterDash(args)
    if len(cmdArgs) == 0 {
        return teneerr.New("COMMAND_NOT_FOUND", "No command specified. Usage: tene run -- <command>", 1)
    }
    ...
}

func hasHelpFlag(args []string) bool {
    for _, a := range args {
        if a == "--" {
            return false // -- means "rest is child command"; child's --help wins
        }
        if a == "--help" || a == "-h" {
            return true
        }
    }
    return false
}
```

The `--` early termination matters: `tene run -- python -h` invokes the
child's `-h`, NOT cobra's help. By scanning only the args before `--`
we preserve that contract.

## 3. Test plan

`internal/cli/run_help_test.go` (new):

| Case | args | Expected |
|---|---|---|
| `tene run --help` | `["run", "--help"]` | exit 0, help text in stdout, no error |
| `tene run -h` | `["run", "-h"]` | same |
| `tene run --env dev --help` | flag-then-help | help text |
| `tene run --help --env dev` | help-first | help text |
| `tene run -- python -h` | child's -h, not run's | "no command specified" → no actually `python` IS specified. Returns COMMAND_NOT_FOUND only if cmdArgs is empty. Here cmdArgs = ["python", "-h"]. So it would TRY to run python (which may not exist on CI). For the unit test we use a guaranteed cmd or verify the flag was not consumed by the run command (i.e. NOT returning help text). |
| `tene run` (bare) | `["run"]` | "No command specified" error (existing behaviour preserved) |

For the child-help isolation test, the cleanest pattern is to check that
`hasHelpFlag(args)` returns false when `--` precedes the flag. That is a
pure unit test on the helper.

## 4. Acceptance criteria

- [ ] Unit tests pass
- [ ] `tene run --help` shows help text and exits 0
- [ ] `tene run -- python -h` does NOT show tene's help (regression check)
- [ ] golangci-lint clean
- [ ] CHANGELOG entry
