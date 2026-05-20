# FX6 — Polish (B7, B8, B10–B13)

> **Bugs**: B7 (MEDIUM — passwd docs), B8 (MEDIUM — list --quiet), B10/B11 (LOW — env list pluralisation + spacing), B12 (LOW — config KEY-only normalisation), B13 (LOW — --dir error specificity)
> **Files touched**: `internal/cli/helpers.go`, `internal/cli/list.go`, `internal/cli/env.go`, `internal/cli/config.go`, `internal/cli/root.go`, `internal/cli/passwd.go`, `internal/cli/fx6_polish_test.go` (new), `pkg/errors/errors.go` (new error code), `CHANGELOG.md`
> **PR target**: `staging`
> **Commit prefix**: `fix(cli):` + `docs(cli):` (one commit covering both)

## 1. Per-bug scope

### B7 — `tene passwd` requires interactive terminal

Already enforced at `passwd.go:19-21` with `teneerr.ErrInteractiveRequired`.
The QA finding was that the wording "This command requires an interactive
terminal." sounds like a missing feature, not a deliberate security
stance. Add a `Long:` description to `passwdCmd` that explains why
(forces witnessed password rotation; closes a CI-driven brute-force
vector). No behavioural change.

### B8 — `tene list --quiet` ignored

`list.go:renderListText` prints the project banner, table header, table
rows, and footer hint unconditionally. With `--quiet` the user expects
"minimal output (errors only)" — matching the existing `tene set --quiet`
contract. Implement by short-circuiting `renderListText` to return early
when `flagQuiet` is set on a non-empty result. For empty results the
text path already returns its single line; with --quiet it becomes
silent too.

For JSON output (`--json`), `--quiet` does not apply (JSON consumers
parse the structured payload — there is no decorative output to
suppress).

### B10 — `1 secrets` plural form

`env.go:140` (`fmt.Printf("...%d secrets)\n", count)`) and `list.go:123`
(`fmt.Printf("\n  %d secrets in ...", len(metas))`) both render incorrect
plurals when count == 1. Add a tiny `pluralize(count int, singular
string) string` helper in `helpers.go` returning `"secret"` /
`"secrets"`; use it at both call sites and at the `env.delete` output
("1 secret removed" not "1 secrets removed").

### B11 — `(  0 secrets)` double-space

`env.go:130-140` builds the line:

```go
} else {
    active = " ("
}
fmt.Printf("  %s %s%s %d secrets)\n", marker, e.Name, active, count)
```

For non-active envs `active = " ("` and the format inserts a `space`
between `%s%s` and `%d`, producing `"  dev (  0 secrets)"`. Restructure
to two distinct format strings (active vs not) so the spacing is
explicit per branch.

### B12 — `tene config KEY` returns "unknown" for valid keys

`config.go:172-174`'s `printSingleConfig` calls `vaultcfg.IsKnown(key)`.
Config keys are stored with a `config.` prefix (the table prints
`config.preview.enabled = true`), but users naturally type the
trailing form (`tene config preview.enabled`). Normalise the key by
trying both forms: if `IsKnown(key)` is false, try `IsKnown("config."+key)`.

### B13 — `--dir` error specificity

`root.go:loadApp()`:

```go
if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
    return nil, teneerr.ErrVaultNotFound
}
```

Returns the same error whether the directory itself is missing or the
directory exists without a vault. Distinguish by checking the parent
dir first:

```go
if _, err := os.Stat(dir); os.IsNotExist(err) {
    return nil, teneerr.New("DIR_NOT_FOUND",
        fmt.Sprintf("Directory %q does not exist.", dir), 1)
}
if _, err := os.Stat(vaultPath); os.IsNotExist(err) {
    return nil, teneerr.ErrVaultNotFound
}
```

## 2. Test plan

`internal/cli/fx6_polish_test.go` (new) — focused tests per bug:

- `TestPluralize` — table-driven on the helper.
- `TestListQuietSuppressesOutput` — `tene list --quiet` on a non-empty
  vault produces empty stdout.
- `TestEnvListSingularPluralFormatting` — exactly one env with one
  secret renders `"1 secret"` not `"1 secrets"`.
- `TestEnvListNoDoubleSpaceForNonActive` — regex confirms no double
  space between paren and count.
- `TestConfigKeyOnlyAcceptsBareKey` — `tene config preview.enabled`
  returns the value without an error.
- `TestLoadAppDirNotFoundIsDistinctError` — pure helper test on the
  new error code path.

## 3. Acceptance criteria

- [ ] all 6 sub-tests pass
- [ ] full cli + auth + keychain regression remains green
- [ ] golangci-lint clean
- [ ] CHANGELOG entries for B7–B13 in the Fixed section
