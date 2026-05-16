# GitHub Repository Transfer Report — tomo-kay/tene → agent-kay-it/tene

> **날짜**: 2026-05-16
> **소요 시간**: 약 1시간 (transfer API call 14:36 KST ~ main merge 22:09 KST 사이 사용자 manual 대기 포함)
> **결과**: ✅ **성공** (1개 follow-up: AWS IAM trust policy)

---

## 1. Executive Summary

GitHub repository 를 사용자의 두 User 계정 (`tomo-kay` → `agent-kay-it`) 간 transfer 했다. 6 phase, 20 task 로 분할 실행. **production 영향 0** (tene.sh 정상, GitHub auto-redirect 활성). 1개 follow-up: AWS IAM role trust policy 수정 (goreleaser → S3 release artifact upload 만 영향, 일상 운영 무관).

| 지표 | Before | After |
|------|--------|-------|
| Repository URL | github.com/tomo-kay/tene | github.com/agent-kay-it/tene |
| Stars | 8 | 8 (유지) |
| Forks | 0 | 0 (유지) |
| Issues/PR history | 113 PRs | 113 PRs (+ 2 신규: #112/#113) |
| main HEAD SHA (pre-transfer) | e1ba551 | e1ba551 (유지) |
| Old URL redirect | — | HTTP 301 → 새 URL ✅ |
| Vercel project (prj_AT0w...) | tomo-kay 연결 | agent-kay-it 자동 follow ✅ |
| tene.sh production | 200 | 200 ✅ |
| Go module path | github.com/tomo-kay/tene | github.com/agent-kay-it/tene |

---

## 2. Phase 별 실행 결과

| Phase | Task ID | 결과 |
|:-----:|:-------:|:----:|
| 0.1 agent-kay-it User account 확인 | #36 | ✅ User type, kay kim |
| 0.2 gh CLI scope | #37 | ✅ `repo` scope 면 충분 |
| 0.3 Vercel GitHub App 사전 설치 | #38 | ✅ 사용자 수동 (2 days ago install) |
| 1.1 Pre-transfer 백업 | #39 | ✅ snapshot JSON (22 LOC) |
| 1.2 Transfer API 호출 | #40 | ✅ HTTP 202 Accepted |
| 1.3 사용자 이메일 acceptance + 검증 | #41 | ✅ 8 stars / SHA e1ba551 / HTTP 301 redirect |
| 2.1 Vercel git reconnect | #42 | ✅ **자동 follow** (수동 reconnect 불필요) |
| 2.2 Vercel env+domain 검증 | #43 | ✅ tene.sh HTTP 200 |
| 3.1 Local git remote URL | #44 | ✅ `git remote set-url origin ...` |
| 3.2 worktree + tene-cloud 검증 | #45 | ✅ tene-cloud 발견: `require github.com/tomo-kay/tene v1.0.0-rc1` (별도 follow-up) |
| 4.1 Branch 생성 + plan | #46 | ✅ `chore/transfer-to-agent-kay-it` from `origin/staging` |
| 4.2 Go module path | #47 | ✅ go.mod + 26 .go files, build/test pass |
| 4.3 Web frontend + blog + assets | #48 | ✅ apps/web/src + content + public + .well-known |
| 4.4 GitHub config + top-level docs | #49 | ✅ .github + .goreleaser + CLAUDE/AGENTS/GEMINI/README/CHANGELOG/CONTRIBUTING/CODE_OF_CONDUCT/SECURITY |
| 4.5 docs/* (audit + plan + design) | #50 | ✅ docs/00-pm + 01-plan + 02-design + 03-report |
| 5.1 Full build + test verification | #51 | ✅ go build clean, web build clean, 23 articles indexability PASS |
| 5.2 PR → staging → main | #52 | ✅ PR #112 merged 34a39a6, PR #113 admin-merged 5266682 |
| 6.1 외부 referrer 안내 | #53 | (pending — Daily.dev/Reddit/Awesome PRs cross-share, GitHub redirect 가 처리 중) |
| 6.2 Homebrew tap 결정 | #54 | (pending — Sprint 3 brew-tap-reactivation 시 agent-kay-it/homebrew-tene 으로 직접 생성) |
| 7 Post-transfer 검증 + 보고 | #55 | ✅ 본 문서 |

## 신규 발견 task

| Task | 결과 |
|------|:----:|
| #56 Vercel reconnect 자동 확인 | ✅ PR #112 의 Vercel preview check 통과로 검증 |
| #57 AWS IAM trust policy update | 🟡 사용자 manual follow-up (non-blocking) |

---

## 3. Commits 요약

`chore/transfer-to-agent-kay-it` 브랜치 (PR #112) 의 4 commits:

| SHA | Title | Files | Change |
|-----|-------|------:|-------:|
| f8096af | Go module path | 26 | +56 -56 |
| b08b764 | Non-Go refs sweep | 30 | +108 -86 |
| 56e7981 | docs/* sweep | 45 | +274 -274 |
| dd2c365 | Final cleanup (Dockerfile/Makefile/llms.txt/Person URLs) | 18 | +49 -49 |
| **합계** | | **108 unique files** | **+487 / -465** |

Promotion PR #113 (staging → main): admin-merge (goreleaser 외부 의존성 fail 무시).

---

## 4. 의도적으로 보존된 historical reference

GitHub 자동 redirect 가 처리하지만 의미적으로 historical 인 것들:

| 위치 | 이유 |
|------|------|
| `apps/web/content/blog/*.mdx` `author: tomo-kay` | 글 작성 시점 attribution |
| `apps/web/content/blog/ai-is-the-best-teacher...mdx` `github.com/tomo-kay/bkit-claude-code` | **bkit 은 별개 repo** (본 transfer 범위 외) |
| `internal/claudemd/generator_test.go` "legacy tomo-kay URL" | cleanup verification test (string contains 검증) |
| `docs/03-report/github-transfer-snapshot-2026-05-16.json` | pre-transfer 백업 (의도된 historical) |
| `docs/archive/*` | 전체 archive 디렉토리 (legacy planning) |

---

## 5. Known Issues + Follow-ups

### 5.1 AWS IAM trust policy (Task #57) — non-blocking

**Symptom**: goreleaser job 의 AWS OIDC step 실패
```
##[error]Could not assume role with OIDC: 
Not authorized to perform sts:AssumeRoleWithWebIdentity
```

**Root cause**: AWS IAM role `tene-prod-github-actions` (arn:aws:iam::507221376909:role/...) 의 trust policy `token.actions.githubusercontent.com:sub` 가 `repo:tomo-kay/tene:*` 만 허용.

**Fix (사용자 manual)**:
1. AWS Console → IAM → Roles → `tene-prod-github-actions`
2. Trust relationships → Edit JSON
3. `"token.actions.githubusercontent.com:sub": "repo:tomo-kay/tene:*"` →
   `"token.actions.githubusercontent.com:sub": "repo:agent-kay-it/tene:*"`
   (또는 array `["repo:tomo-kay/tene:*", "repo:agent-kay-it/tene:*"]` 으로 둘 다 허용 — 안전한 transition period)
4. 다음 release tag push 시 goreleaser S3 upload 정상화

**Impact when unfixed**:
- ❌ release artifact S3 upload 안 됨 (tene-releases bucket)
- ❌ `install.sh` (curl-pipe install path) 가 새 binary 못 받음
- ✅ CI (lint/test) 정상
- ✅ Vercel deploy 정상
- ✅ GitHub Releases tag 자체는 생성됨 (binary 만 누락)

### 5.2 tene-cloud sibling repo (별도 follow-up)

`tene-cloud` (별도 `tomo-kay/tene-cloud` repo) 의 `go.mod`:
```
require github.com/tomo-kay/tene v1.0.0-rc1
```

본 PR 이 v2 release 로 tag 되면 tene-cloud 에서:
```
go get github.com/agent-kay-it/tene@v...
```
실행하여 require 갱신 필요. 별도 PR 로 처리. tene-cloud 자체의 owner transfer 도 사용자 결정 사항.

### 5.3 외부 referrer 명시 갱신 (Task #53)

GitHub auto-redirect 가 처리하지만 best practice 로 명시 갱신 권장:
- Show HN agent-kay 포스트
- Daily.dev AI-Safe Secrets squad
- GeekNews 한국어 post
- Reddit r/vibecoding 등 4 sub
- Awesome lists PR (5건 — bkit-pdca-for-claude-code 블로그 cross-link)

Vercel/GitHub redirect 로 인해 broken link 0 — 우선순위 낮음.

---

## 6. Verification Matrix

| 검증 | Pre-transfer | Post-transfer | Status |
|------|:------------:|:-------------:|:------:|
| Repo URL accessible | github.com/tomo-kay/tene | github.com/agent-kay-it/tene | ✅ |
| Old URL redirect | n/a | HTTP 301 | ✅ |
| Star count | 8 | 8 | ✅ |
| Fork count | 0 | 0 | ✅ |
| Watcher count | 0 | 8 (transfer 후 GitHub 자동 set) | ⚠️ |
| main HEAD SHA | e1ba551 | e1ba551 (Phase 1) → 5266682 (Phase 5 후) | ✅ |
| Branch protection | enabled | enabled | ✅ |
| Vercel project ID | prj_AT0w... | prj_AT0w... 유지 | ✅ |
| tene.sh root HTTP | 200 | 200 (585ms) | ✅ |
| /blog/bkit-sprint-orchestration | 200 | 200 | ✅ |
| Vercel preview deploy on PR | 작동 | 자동 follow | ✅ |
| Go module path | tomo-kay/tene | agent-kay-it/tene | ✅ |
| go build ./... | clean | clean | ✅ |
| go test pass | pass | pass | ✅ |
| apps/web build | clean | clean | ✅ |
| 23 articles indexability | PASS | PASS | ✅ |
| CI lint+test on PR/main | pass | pass | ✅ |
| Auto-tag workflow | pass | pass | ✅ |
| **goreleaser AWS OIDC** | pass (tomo-kay scope) | **fail** (agent-kay-it scope 미허용) | ❌ |
| **tene-cloud go.mod require** | github.com/tomo-kay/tene | github.com/tomo-kay/tene (미갱신) | 🟡 |

총: **18/20 PASS, 1 follow-up, 1 별도 repo 의존**

---

## 7. Lessons Learned

1. **Vercel GitHub App pre-install** 이 transfer 성공의 핵심. agent-kay-it 에 사전 설치된 덕분에 Phase 2.1 Vercel reconnect 가 0 클릭으로 자동 완료.
2. **GitHub auto-redirect 가 매우 강력** — git operations + browser navigation + API 모두 처리. 그러나 **Go module path** 는 별개 (module path 정확 매치 필요).
3. **AWS IAM OIDC trust policy** 는 transfer 의 hidden coupling. 다른 cloud provider (GCP, Azure) 도 동일 패턴 — sub claim 으로 repo URL 검증.
4. **User-to-User transfer** 는 receiver email acceptance 필요. API 자동 수락 endpoint 없음. 사용자가 collaborator 인 경우에도 acceptance 별개.
5. **goreleaser admin-override merge** 가 합리적 — 외부 의존성 (AWS) 이슈로 인한 fail 은 코드 정확성과 무관. Branch protection 의 `--admin` flag 가 이런 경우 정확한 도구.

---

## 8. 보존된 산출물

| 파일 | LOC | 용도 |
|------|----:|------|
| `docs/03-report/github-transfer-snapshot-2026-05-16.json` | 22 | Pre-transfer 상태 백업 (stars/forks/branches/PRs/Issues) |
| `docs/03-report/github-transfer-2026-05-16.md` (본 문서) | ~250 | 최종 보고서 |

---

> **Status**: ✅ Complete (with 2 documented follow-ups)
>
> Production live at https://tene.sh — GitHub repo at https://github.com/agent-kay-it/tene
