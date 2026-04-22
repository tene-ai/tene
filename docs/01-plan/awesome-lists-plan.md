# Awesome-list PR Playbook — tene

> Phase 3 of the `ai-discoverability` feature. Six target lists, six PRs, one
> at a time with ≥ 3 days spacing so maintainers don't see them as a spam
> burst.
>
> Plan ref:    `docs/01-plan/features/ai-discoverability.plan.md` §4 Phase 3
> Design ref:  `docs/02-design/features/ai-discoverability.design.md` §2.4
> PR body:     `.claude/templates/growth/awesome-pr-template.md`

---

## Submission order (≥ 3 day spacing)

| # | Target | Priority | Fit | Expected result |
|--:|--------|:--------:|----:|-----------------|
| 1 | `mahseema/awesome-ai-tools` | ⭐⭐⭐ | AI-tools fit best; weakest star gate | Most likely merge |
| 2 | `awesome-claude-code` (verify existence) | ⭐⭐⭐ | Perfect topical fit | Pending existence |
| 3 | `sbilly/awesome-security` | ⭐⭐ | Secret Management section | Strict; writeup required |
| 4 | `devsecops/awesome-devsecops` | ⭐⭐ | Secrets Management section | Medium strictness |
| 5 | `agarrharr/awesome-cli-apps` | ⭐⭐ | Development / Security | Medium strictness |
| 6 | `avelino/awesome-go` | ⭐ | Strict star gate (~100+ typical) | **Defer** until stars ≥ 100 |

---

## 1. `mahseema/awesome-ai-tools`

- URL: https://github.com/mahseema/awesome-ai-tools
- README path: `README.md`
- Target section: `AI Development Tools` (or `Security`; confirm by reading
  existing entries first).
- Alphabetical position: after entries starting with `t` — place before
  "Textgen" / similar.

**Line to add** (one line, keep leading `* `):

```
* [tene](https://github.com/tomo-kay/tene) — Local-first AI-safe secret manager CLI. XChaCha20-Poly1305 encryption, runtime env-var injection, auto-generates rules for Claude Code, Cursor, Windsurf.
```

**Commands**:

```bash
gh repo fork mahseema/awesome-ai-tools --clone --remote
cd awesome-ai-tools
git checkout -b add-tene
$EDITOR README.md   # place the line alphabetically in the right section
git add README.md
git commit -m "Add tene — local-first AI-safe secret manager CLI"
git push -u origin add-tene
gh pr create --title "Add tene — local-first AI-safe secret manager CLI" \
  --body-file ../tene-biz/.claude/templates/growth/awesome-pr-template.md
```

---

## 2. `awesome-claude-code` (existence check first)

```bash
gh search repos "awesome-claude-code" --limit 10 --json fullName,stargazerCount,description,updatedAt
```

If a strong candidate exists (> 50 stars, updated in the last 90 days):
apply the same pattern. If none exists, skip — **do not** open a stub
awesome list; that is a separate project.

---

## 3. `sbilly/awesome-security`

- URL: https://github.com/sbilly/awesome-security
- Target section: `Secret Management` (search the README; if the section
  does not exist, do **not** create one in the same PR — propose it in a
  separate issue first).
- Strictness: high. Expect a maintainer to ask for a short technical
  writeup on the crypto model. Have `docs/02-design/features/ai-discoverability.design.md`
  §2.2 content ready as backup.

---

## 4. `devsecops/awesome-devsecops`

- URL: https://github.com/devsecops/awesome-devsecops
- Target section: `Secrets Management`
- Medium strictness. Good fit for the AI-agent angle.

---

## 5. `agarrharr/awesome-cli-apps`

- URL: https://github.com/agarrharr/awesome-cli-apps
- Target section: `Development/Passwords` or `Development/Security`. Confirm
  by reading the README and the existing alphabetical placement.
- Medium strictness.

---

## 6. `avelino/awesome-go` (defer)

- URL: https://github.com/avelino/awesome-go
- Target section: `Security > Secrets Manager`
- **Star gate**: awesome-go is unusually strict; the maintainer has
  historically required ≥ ~100 stars + active maintenance + tests + CI. At
  5 stars, an immediate PR will be closed with a "not yet" comment.
- **Defer** this PR until the repo has ≥ 100 stars. Track in Run History.

---

## Tracking

After each PR is opened, append a row to
`tene/docs/stats/ai-discoverability.md` → `## Run History`:

| Date | Target list | PR URL | State | Note |

States: `open` · `merged` · `rejected` · `stale` · `deferred`.

---

## Common rejection reasons + recovery

- **"Too new / not enough stars"**: close politely; revisit in ~6 months.
- **"One-line description is too marketing-y"**: rewrite to pure function
  ("Encrypted secret manager CLI with runtime env-var injection"), resubmit.
- **"Similar project already listed"**: check if the listed one is
  maintained. If not, propose a replacement in a separate issue before
  re-opening the PR.
- **"Wrong category"**: move the line, force-push to your PR branch.

---

## Do NOT

- Open multiple awesome-list PRs the same day.
- Edit the list's `CONTRIBUTING.md` in your PR.
- Bundle tene with another project.
- Argue publicly with a maintainer. If they reject, accept and move on.
- Open a PR the day after a major repo event (mass-add PRs, spam waves) —
  your PR will get swept up in the moderation backlash.
