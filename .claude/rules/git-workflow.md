# Git Workflow Rules

## Branch Strategy

```
feature/* or fix/* → PR → main (release)
```

- `main`: production-ready. Direct commits forbidden.
- `feature/*`, `fix/*`: branch from main.

## Rules (MUST follow)

1. **Never commit directly to main.**
2. **Always branch from main:**
   ```bash
   git checkout main && git pull origin main
   git checkout -b fix/issue-name
   ```
3. **PR workflow:** fix branch → main PR → CI passes → merge
4. **No cherry-pick.** Squash merge changes commit hashes, causing divergence.
5. **No force-push to main.**
6. **Delete fix/feature branches after merge** (local + remote).

## Anti-patterns (never do)

- `git commit` on main directly
- `git cherry-pick` across branches
- `git push --force` on main
- `git tag -f` (GoReleaser asset collision)
- Guessing at problems — check logs first
