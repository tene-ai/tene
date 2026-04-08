# Git Workflow Rules

## Branch Strategy

```
feature/* or fix/* → PR → staging (QA) → PR → main (prod auto-deploy)
```

- `main`: 프로덕션. 직접 커밋 절대 금지.
- `staging`: QA/검증. 직접 커밋 절대 금지.
- `feature/*`, `fix/*`: 항상 staging에서 분기.

## Rules (MUST follow)

1. **절대 main이나 staging에 직접 커밋하지 않는다.**
2. **항상 staging에서 feature/fix 브랜치를 딴다:**
   ```bash
   git checkout staging && git pull origin staging
   git checkout -b fix/issue-name
   ```
3. **PR 순서:** fix 브랜치 → staging PR → CI 통과 → merge → staging QA → main PR
4. **cherry-pick 금지.** squash merge로 커밋 해시가 달라져서 브랜치 diverge됨.
5. **force-push 금지.** (태그 제외)
6. **로컬 동기화:** `git reset --hard origin/xxx`로 리모트와 맞춘다. pull로 merge commit 만들지 않는다.
7. **작업 완료 후 fix/feature 브랜치 삭제** (로컬 + 리모트 모두).

## Sync (main ↔ staging)

staging → main 머지 후 staging이 뒤처지면:
```bash
gh pr create --base staging --head main --title "sync: main → staging"
gh pr merge --merge
```

## 배포 전 체크리스트

staging 배포 전:
- [ ] Docker 빌드 테스트: `docker build -f Dockerfile.server .`
- [ ] DB SSL 연결 확인 (sslmode=require)
- [ ] CORS origins에 staging 도메인 포함 확인
- [ ] 환경변수 비교: local vs staging vs prod
- [ ] ECS 다중 태스크 환경에서 in-memory state 사용 여부 확인

## Anti-patterns (절대 하지 말 것)

- `git commit` on main/staging directly
- `git cherry-pick` across main/staging
- `git push --force` on main/staging
- `git tag -f` (GoReleaser asset 충돌)
- 추측으로 문제 해결 시도 — 로그 먼저 확인
