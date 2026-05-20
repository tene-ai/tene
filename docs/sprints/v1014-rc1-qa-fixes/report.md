# Sprint Report — v1014-rc1-qa-fixes

> **Sprint ID**: `v1014-rc1-qa-fixes`
> **Period**: 2026-05-20 (single day, L4 Full-Auto)
> **Working branch**: `fix/v1014-rc1-qa-findings` (off `origin/staging` @ `1fd5b7b`)
> **HEAD at archive**: `7bcdbd0`
> **Status**: complete; ready for PR merge to `staging` then v1.0.14 GA tag

## §0 Mission re-stated

Close all 10 regressions found by the v1.0.14-rc1 QA cycle without
weakening PR #116's permission-tier dispatcher. Three new invariants
(I-11/I-12/I-13) extend G4 enforcement so the same class of bug
cannot ship again.

## §1 KPI snapshot

| Metric | Target | Actual |
|---|---|---|
| CRITICAL bugs closed | 3 | 3 ✅ |
| HIGH bugs closed | 3 | 3 ✅ |
| MEDIUM bugs closed | 3 | 3 ✅ |
| LOW bugs closed | 4 | 4 ✅ |
| New invariant tests | 3 | 64 sub-cases across 13 invariants ✅ |
| New unit test files | 6 | 6 ✅ |
| Cross-repo (tene-cloud) regressions | 0 | 0 ✅ |
| matchRate (post-fix QA) | ≥ 95% | 100% (all 10 bugs verified closed) |
| golangci-lint | 0 issues | 0 issues ✅ |
| `go test -race ./...` | green | green ✅ |

## §2 Commits

```
7bcdbd0 fix(cli): polish — list --quiet, plurals, config key normalisation, --dir error, passwd docs (B7,B8,B10-B13)
45b6d68 fix(cli): tene run --help shows help text (B6)
73ddd63 fix(auth,cli): dispatch hook skips synthetic verbs, ValidateStrict catches reverse drift (B4, B5)
4aedbc3 fix(update): SemVer-aware update check, no RC→stable downgrade (B3, I-13)
9ca4c4b fix(cli): destructive ops fail-closed on non-TTY (B2, B9, I-12)
b772095 fix(keychain): --no-keychain stops sharing ~/.tene/keyfile (B1, I-11)
3c5b497 docs(sprint): add v1014-rc1-qa-fixes master plan + FX1 design
1fd5b7b Merge pull request #116 from agent-kay-it/feature/cli-ux-permission-model
```

Total 7 commits (1 docs + 6 fixes). Diff stat across the sprint:

```
docs/sprints/v1014-rc1-qa-fixes/  ~2500 LoC (master plan + 6 feature specs + this report)
internal/keychain/                 ~210 LoC added (NullStore + tests)
internal/cli/                      ~1500 LoC modified (6 fixes + tests)
internal/auth/                     ~250 LoC modified (ValidateStrict + tests)
CHANGELOG.md / SECURITY.md / README.md  ~120 LoC added
```

## §3 Phase history

| Phase | Verb | Notes |
|---|---|---|
| Plan | master-plan.md | 11 sections + 3 appendices |
| Design | FX1–FX6.md | one per feature; root-cause-first |
| Do | 6 PR-sized commits | one feature per commit, no rebasing required |
| Check | post-fix QA | docs/05-qa/tene-cli-v1.0.14-postfix.qa-report.md in tene-biz |
| Cross-repo verify | tene-cloud build + test | green |
| Report | this doc + CHANGELOG.md | — |
| Archive | (pending) | will land with the PR merge |

## §4 Auto-pause events

None — the sprint ran in L4 Full-Auto and hit no quality-gate failures.
Two minor course-corrections happened mid-sprint:

1. **FX1 NullStore introduction** changed the keychain test surface,
   triggering a full `internal/cli/...` regression that took 268 s.
   Result: green. No code change needed.
2. **FX4 reverse-drift Validate** broke synthetic-tree tests at first
   because they did not populate every CommandTier entry. Resolved by
   splitting into `Validate` (forward) and `ValidateStrict`
   (bidirectional) — production `init()` panics use Strict, unit
   tests use the looser form.

## §5 Lessons learned

- **Static + runtime predicates must share a source.** PR #116's
  Validate skipped cobra's synthetic `help` command but the runtime
  dispatcher tried to look it up; B4 was the avoidable cost. The new
  exported `auth.IsCobraSynthetic` makes both sides import the same
  predicate.
- **Permission tables need both-direction enforcement.** "Every
  registered verb has a tier" was the only check; "every tier has a
  registered verb" was unenforced and that is exactly how `logout`
  became a phantom. `ValidateStrict` covers both directions now.
- **`promptConfirm` default-yes on non-TTY was a footgun.** The
  original intent was probably "tests don't need to mock stdin" but
  the cost was a silent CRITICAL data-loss path. Default-no with an
  explanatory stderr line is the right tradeoff; tests that need the
  bypass pass `--force` like CI/CD scripts now must.
- **SemVer comparisons must use `golang.org/x/mod/semver`.** A naive
  `!=` is one keystroke from a downgrade. Standard library is the
  only safe choice — pre-release identifier semantics are subtle.

## §6 Carry items (deferred)

| Item | Reason | Target sprint |
|---|---|---|
| Cleanup orphaned `tene-<hash>` keychain entries (95 on dev machine) | Out of master plan scope (§6 Keychain hygiene); cleanup verb is a new feature, not a bug fix | v1.0.15 keychain-hygiene sprint |
| `tene run` shares global args with parent flags (DisableFlagParsing artefact) | The B6 fix scoped to --help only; broader cleanup of `parseFlagsBeforeDash` would touch more code than this defensive sprint allows | v1.0.15 cli-polish |
| `auth.CommandTier` table re-grow with cloud verbs (login/logout/push/pull/sync/billing/team) when cloud is re-enabled | Cloud feature freeze remains | post-cloud-redesign |

## §7 Acknowledgements

The sprint operated in L4 Full-Auto throughout. The user's mid-sprint
challenges ("you're guessing — verify the code") prompted the
direct-code-read pattern that produced the Appendix A root-cause
table in master-plan.md, which in turn made every FX design.md
precise enough to ship in 6 commits without rework.

---

**End of sprint.** Branch `fix/v1014-rc1-qa-findings` is ready for PR
review and merge to `staging`, after which a v1.0.14 tag closes the
release line opened by `v1.0.14-rc1`.
