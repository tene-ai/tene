# FX3 — `tene update` SemVer-aware comparison + RC channel handling

> **Bug**: B3 (CRITICAL — release infra)
> **Invariant introduced**: I-13
> **Files touched**: `internal/cli/update.go`, `internal/cli/update_semver_test.go` (new), `go.mod` (promote `golang.org/x/mod` from indirect to direct), `CHANGELOG.md`
> **PR target**: `staging`
> **Commit prefix**: `fix(update):`

## 1. Problem

`update.go:85` and `:100` rely on a single-character `!=` comparison:

```go
"updateAvailable": currentDisplay != latestTag && currentDisplay != "vdev",
```

The QA report (RECHECK-3) reproduces the consequence:

```json
$ tene update --check --json
{
  "currentVersion":  "v1.0.14-rc1",
  "latestVersion":   "v1.0.13",
  "targetVersion":   "v1.0.13",
  "updateAvailable": true
}
```

A user on v1.0.14-rc1 typing `tene update` (no args) is auto-downgraded to
v1.0.13 because `targetVersion := latestTag` and the proceed-or-not branch
only checks string inequality, not version ordering. RC1 users are one
keystroke from bricking their own upgrade.

## 2. Design

### 2.1 Helper `shouldOfferUpdate`

Use the standard `golang.org/x/mod/semver` package (already an indirect
dep at v0.30.0 — promote it to direct). The helper centralises the
"is `latest` strictly newer AND on a channel we accept" decision:

```go
// shouldOfferUpdate reports whether tene should offer to upgrade from
// `current` to `latest`. Sprint v1014-rc1-qa-fixes / FX3 (invariant I-13).
//
// Rules (all evaluated against semver.Compare):
//   - latest <= current  → never offer (covers B3 RC-to-stable downgrade).
//   - current is "dev"   → never offer (no comparable baseline).
//   - either tag is not a valid semver → never offer (fail-closed; we
//     do not invent a comparison for malformed input).
//   - latest is a pre-release (rc/beta/alpha) AND includePrerelease is
//     false → never offer (a stable user should never be nudged to RC
//     automatically; they must pass --include-prerelease).
//   - otherwise → offer.
//
// A few representative scenarios:
//   ("v1.0.13",     "v1.0.14",     false) → true   (normal upgrade)
//   ("v1.0.14-rc1", "v1.0.13",     false) → false  (B3 fix: RC > stable)
//   ("v1.0.14-rc1", "v1.0.14",     false) → true   (RC → stable upgrade)
//   ("v1.0.13",     "v1.0.14-rc1", false) → false  (no auto RC for stable)
//   ("v1.0.13",     "v1.0.14-rc1", true)  → true   (explicit opt-in)
//   ("v1.0.13",     "v1.0.13",     false) → false  (already up to date)
//   ("vdev",        anything,      _)     → false  (no baseline)
func shouldOfferUpdate(current, latest string, includePrerelease bool) bool {
    if current == "vdev" || current == "dev" || current == "" {
        return false
    }
    if !semver.IsValid(current) || !semver.IsValid(latest) {
        return false
    }
    if semver.Compare(latest, current) <= 0 {
        return false
    }
    if !includePrerelease && semver.Prerelease(latest) != "" {
        return false
    }
    return true
}
```

### 2.2 New flag `--include-prerelease`

CLI surface:

```go
var updateFlagIncludePrerelease bool

func init() {
    updateCmd.Flags().BoolVar(&updateFlagCheck, "check", false, "Check for updates without installing")
    updateCmd.Flags().BoolVar(&updateFlagIncludePrerelease, "include-prerelease", false,
        "Allow upgrading to RC/beta releases (opt-in)")
}
```

### 2.3 `runUpdate` integration

`updateAvailable` JSON field and the text-mode "Update available!" line
both call `shouldOfferUpdate`. The proceed-with-install gate gains an
extra check: when `targetVersion == latestTag` (user typed `tene update`
without an explicit version) AND `shouldOfferUpdate` returns false, the
install branch is skipped and the command exits 0 with an
"Already up to date." line.

An explicit `tene update v1.0.13` (user supplies a target) is left
untouched — that is a manual choice, not an automatic recommendation,
and it is sometimes legitimate (rolling back a broken release).

### 2.4 Invariant I-13

> `tene update` and `tene update --check` never recommend a target whose
> semver precedence is lower than the current version. Pre-release tags
> (rc/beta/alpha) are excluded from the automatic recommendation unless
> the user passes `--include-prerelease`.

## 3. Test plan

`internal/cli/update_semver_test.go` (new) — pure unit test on
`shouldOfferUpdate`, no network:

| Case | current | latest | includePrerelease | Expected |
|---|---|---|---|---|
| stable upgrade | v1.0.13 | v1.0.14 | false | true |
| B3 fix — RC > stable | v1.0.14-rc1 | v1.0.13 | false | false |
| RC → stable upgrade | v1.0.14-rc1 | v1.0.14 | false | true |
| stable to RC blocked | v1.0.13 | v1.0.14-rc1 | false | false |
| opt-in to RC | v1.0.13 | v1.0.14-rc1 | true | true |
| already up to date | v1.0.13 | v1.0.13 | false | false |
| RC to newer RC | v1.0.14-rc1 | v1.0.14-rc2 | true | true |
| RC to newer RC w/o opt-in | v1.0.14-rc1 | v1.0.14-rc2 | false | false |
| dev baseline | vdev | v1.0.14 | false | false |
| malformed input | v1.0.13 | not-semver | false | false |

## 4. Acceptance criteria

- [ ] 10 unit test cases pass
- [ ] `tene update --check --json` on v1.0.14-rc1 with latest=v1.0.13 → `updateAvailable: false`
- [ ] `golangci-lint run` 0 issues
- [ ] go.mod's `golang.org/x/mod` promoted from indirect to direct
- [ ] CHANGELOG entry added
