# FX4 — Dispatch Hook Robustness + Reverse-Drift Validate

> **Bugs**: B4 (HIGH — `tene help` returns dispatch error), B5 (HIGH — `logout` phantom in permissions table)
> **Invariant**: I-10 strengthened (now enforces *both* directions)
> **Files touched**: `internal/auth/permissions.go`, `internal/auth/permissions_test.go`, `internal/cli/root.go`, `internal/cli/permissions_dispatch_test.go`, `CHANGELOG.md`
> **PR target**: `staging`
> **Commit prefix**: `fix(auth):` + `fix(cli):`

## 1. Problem

### B4 — `tene help` is broken

```
$ tene help
Error: internal: command "help" has no PermLevel entry in
internal/auth.CommandTier — refusing to dispatch
$ tene help set
Error: internal: command "help" has no PermLevel entry in ...
```

Root cause: `internal/cli/root.go:120-141` (`rootPersistentPreRunE`)
looks up every command path in `auth.TierFor` and refuses dispatch on
miss. `auth.Validate()` already knows to skip cobra's synthetic `help`
command in its tree walk (via `isCobraInternal`), but that
skip-logic was never plumbed into the runtime dispatcher. The fix
mirrors the same predicate at the dispatch site.

### B5 — `logout` phantom

`internal/auth/permissions.go:119` carries `"logout": PermMetaRead` but
`internal/cli/root.go:243-254` intentionally does NOT register the
logout cobra command (cloud feature disabled). The QA reproduction:

```
$ tene permissions | grep logout
logout       metaread            no       ← listed
$ tene logout
Error: unknown command "logout" for "tene"  ← but doesn't exist
```

Root cause: `auth.Validate` only checks one direction (every registered
verb has a tier). The reverse direction (every tier has a registered
verb) was unenforced, so when the cloud verbs were unregistered the
stale `logout` entry survived in the table.

## 2. Design

### 2.1 Export the synthetic-command predicate

`auth/permissions.go` currently has `isCobraInternal(c *cobra.Command)`
as a private helper for the `walk()` function. Promote it to public
`IsCobraSynthetic(c *cobra.Command) bool` so the cli package's
`rootPersistentPreRunE` can reuse the same matcher. Rename in
permissions.go and update its single call site.

### 2.2 Dispatch hook skips synthetic commands

In `cli/root.go:rootPersistentPreRunE`, before the `auth.TierFor` call:

```go
if auth.IsCobraSynthetic(cmd) {
    // cobra's auto-generated help / __complete commands never reach
    // user-facing RunE in the normal flow; their job is to format
    // text for the parent command. Treating them like real verbs
    // and demanding a CommandTier entry was the root cause of B4
    // (`tene help` returned "no PermLevel entry").
    return nil
}
```

This sits before the audit-log emit too — synthetic commands do not
generate `cli.<tier>.<verb>` audit rows (they have no tier, and the
operator already sees the help text in stdout).

### 2.3 Remove the stale `logout` entry

`auth.CommandTier` drops `"logout": PermMetaRead`. The cloud feature
is currently disabled (`root.go:243-254`); when it is re-enabled the
re-registering PR will reintroduce the entry as part of its scope.
The permissions table now lists only the 19 verbs the binary actually
dispatches (15 PermMetaRead + 5 PermSecretWrite + 5 PermSecretRead =
25 total, down from rc1's 26).

### 2.4 Reverse-drift check in `auth.Validate`

Extend `auth.Validate(rootCmd)` to also assert every `CommandTier` key
is reachable in the rootCmd subtree. The reverse walk uses
`rootCmd.Find()` for each path — if `Find` returns `nil` the entry is
stale and the error names it:

```go
// (existing forward direction check first, unchanged)
// New reverse direction:
for path := range CommandTier {
    parts := strings.Fields(path)
    cmd, _, err := rootCmd.Find(parts)
    if err != nil || cmd == rootCmd || cmd.Name() != parts[len(parts)-1] {
        orphans = append(orphans, path)
    }
}
```

Orphans concatenate into the same error message format the forward
check uses, with a "stale entry" prefix so the operator knows whether
to add a tier or remove one.

### 2.5 Audit emit also skips synthetic commands

`emitCliAuditRow(dir, auditActionFor(tier, path))` is called *after*
the tier check passes. With the early return above for synthetic
commands, we naturally skip the audit row too — which is correct (a
`tene help` invocation is not a security-relevant audit event).

## 3. Test plan

### 3.1 `internal/auth/permissions_test.go` updates

- Drop `logout` from the expected map (line 104) → 15 entries in
  PermMetaRead, 25 total.
- Update `TestCommandTier_Counts` expected values: `PermMetaRead = 15`.
- Add `TestValidate_FailReverseDrift` (synthetic tree where a
  CommandTier-listed verb is not registered): error must mention the
  orphan path and the phrase "stale entry" so the operator knows the
  fix direction.
- Add `TestValidate_PassesProductionTreeWithoutLogout` (sanity: after
  the FX4 changes, `Validate(rootCmd)` returns nil on the real tree).

### 3.2 `internal/cli/permissions_dispatch_test.go` new cases

- `TestPersistentPreRunE_HelpSubcommandSkipped`: synthesize a "help"
  child via `rootCmd.InitDefaultHelpCmd()` (mirrors the existing
  `TestValidate_IgnoresHelpCommand` pattern), call
  `rootPersistentPreRunE(helpCmd, nil)`, expect nil error.
- `TestPersistentPreRunE_CompleteSubcommandSkipped`: same for the
  `__complete` cobra command.

### 3.3 In-vivo verification post-build

```bash
go build -o /tmp/tene-fx4-bin ./cmd/tene
/tmp/tene-fx4-bin help            # expect: help text, exit 0
/tmp/tene-fx4-bin help set        # expect: set's help, exit 0
/tmp/tene-fx4-bin permissions | grep -c logout   # expect: 0
```

## 4. Acceptance criteria

- [ ] auth.permissions_test.go: 25 entries, 15/5/5 split
- [ ] cli.permissions_dispatch_test.go: 2 new tests pass
- [ ] `tene help` works (exit 0)
- [ ] `tene permissions` no longer lists `logout`
- [ ] Reverse-drift Validate panics at startup if a stale entry is
  reintroduced — verified by unit test
- [ ] go test -race / golangci-lint clean
- [ ] CHANGELOG entries for B4 + B5
