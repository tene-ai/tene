# tene AI 발견성 개선 — Plan (팀 구성 + 실행 계획)

| 항목 | 값 |
|------|-----|
| Feature | `ai-discoverability` |
| 기반 PRD | `docs/00-pm/ai-discoverability.prd.md` |
| 기반 감사 | `docs/03-report/ai-discoverability-2026-04-23.md` |
| 설계 문서 | `docs/02-design/features/ai-discoverability.design.md` |
| 코드베이스 감사 | `docs/02-design/features/ai-discoverability.codebase-audit.md` |
| 브랜치 | `feature/ai-discoverability-audit` (origin/staging 베이스) |
| 총 공수 | 24-32 인시간 |
| 기간 | W1 (6h) → W2 (8h) → M1 (10h) → 감사 (8h) |

---

## 1. 팀 구성 (Agent Teams)

### 1.1 역할별 책임 및 담당 항목 매핑

22개 개선 항목을 7개 role 에 분배. 각 role 은 담당 항목의 설계·구현·테스트 주체.

| Role (Agent) | 담당 항목 | 파일 스코프 | 공수 |
|-------------|----------|------------|-----|
| **CTO Lead** (`cto-lead`) | 조율 · W1 승인 · 품질 gate · 충돌 조정 | 전체 | 2h |
| **Product Manager** (`product-manager`) | PRD 유지 · 변경 관리 · Success Criteria 추적 | `docs/00-pm/`, `docs/stats/ai-discoverability.md` | 2h |
| **Frontend Architect** (`frontend-architect`) | **D-2 · D-4 · A-2 · A-3 · A-4 · A-5 · I-2 · I-3** (랜딩 8건) | `apps/web/src/app/`, `apps/web/src/components/` | 6h |
| **Infra Architect** (`infra-architect`) | **S-1 · S-2 · S-4 · S-5** (배포 4건) + **D-1 · D-3 · D-5 · I-1** (GitHub repo 4건) | `.goreleaser.yml`, `Dockerfile`, `.github/`, root trust 파일 | 6h |
| **Backend (Go) 개발자** | **U-1 · U-4 · S-3** + shell completion 등록 (S-4 의 CLI 측) | `internal/cli/*.go`, `cmd/tene/main.go`, `apps/web/public/install.sh` | 4h |
| **Security Architect** (`security-architect`) | U-1 설계 리뷰 · SECURITY.md 감수 · GPG 서명 계획 (S-5) · 3rd-party audit 발주 검토 | `SECURITY.md`, `get.go` diff review, `.goreleaser.yml signs:` | 3h |
| **QA Strategist** (`qa-strategist`) | **U-5 docs/cli-reference.md** + §6 자동/수동 검증 체크리스트 실행 | `docs/cli-reference.md`, `tests/` | 5h |
| **Growth Agent** (수동) | 블랙박스 검증 · Dev.to 크로스포스트 · Awesome-list PR · HN/Reddit 시딩 | `docs/stats/`, 외부 채널 | 4h/주 (ongoing) |

**총 24h (Growth Agent 제외)**. Growth 는 분기 ongoing.

### 1.2 병렬화 가능 그룹 (Gantt 화)

- **Group A (독립 · 병렬 가능)**: Frontend Architect (랜딩 8건) ⊥ Infra Architect (배포 4건 + GitHub 4건)
- **Group B (의존)**: Backend Go (U-1 안전 가드) → Security Architect 리뷰 → 배포
- **Group C (의존)**: Infra Architect (S-1 Homebrew tap 리포 생성) → Backend Go (S-4 shell completion 등록) → Infra Architect (S-4 goreleaser 패키징)

```
W1 ┬── Frontend [D-2,D-4,A-2,A-3,A-4,A-5] ─────────────┐
   ├── Infra    [D-1,D-3,D-5,A-1(README),S-3(install.sh)] ─┤
   └── Backend  [U-4,A-1(root.go)] ─────────────────────┤
                                                          ├── W1 합류
W2 ┬── Infra    [I-1 4종 파일, S-1 Homebrew, S-2 Docker, S-4 man/completion] ┐
   ├── Backend  [U-1 get.go 가드, S-4 completion cmd] ──────────┤
   └── Security [U-1 리뷰, SECURITY.md 감수] ───────────────────┤
                                                                  ├── W2 합류
M1 ┬── Frontend [I-2 Trust, I-3 Breadcrumb] ──┐
   ├── QA       [U-5 CLI 레퍼런스] ────────────┤
   └── Security [S-5 GPG 서명 (선택)] ─────────┤
                                                ├── M1 합류
검증 ── QA + Growth [§6 자동/수동 체크리스트] ─┘
```

---

## 2. 마일스톤 & 타임라인

### 2.1 Milestone 1 — Week 1 Quick Wins (6h, 2026-04-24 ~ 04-30)

**목표**: 퍼널 차단 🔴 중 7건 즉시 해소. 외부에서 즉시 체감되는 변화.

| 항목 | 담당 | 소요 | 의존성 | DoD |
|:---:|------|:---:|-------|-----|
| D-1 Custom OG 이미지 | Infra | 1분 | — | `gh repo view --json usesCustomOpenGraphImage` = true |
| D-2 robots.ts LLM 명시 | Frontend | 5분 | — | `curl tene.sh/robots.txt | grep ClaudeBot` 매치 |
| D-3 Discussions 활성화 + Q 3개 시딩 | Infra | 15분 | — | `gh api repos/... --jq .has_discussions` = true, 3개 Q 게시 |
| D-4 `<link rel=ai-index>` + `.well-known/ai.json` | Frontend | 10분 | — | `curl tene.sh/.well-known/ai.json` 200 |
| D-5 `.github/FUNDING.yml` | Infra | 5분 | — | 파일 존재 + GitHub Sponsor 버튼 표시 |
| **A-1** README Cloud Commands 불일치 해소 | Infra | 2분 | — | `grep -c "tene login" README.md` = 0 or Coming soon |
| A-2 description 154자 | Frontend | 1분 | — | `layout.tsx:19-20` 교체 + 글자 수 ≤ 155 |
| A-3 aggregateRating 제거 | Frontend | 5분 | — | `grep -r aggregateRating apps/web/src` = 0 |
| A-4 Organization + WebSite JSON-LD | Frontend | 10분 | — | `curl tene.sh | grep '@type":"Organization"' ` 매치 |
| A-5 "Tene is X" 1문장 | Frontend | 5분 | `src/data/hero.ts` 읽기 | 히어로 `sub` 첫 단어가 "Tene is..." |
| S-3 install.sh 힌트 확장 | Backend | 3분 | — | `install.sh` 출력에 `llms.txt` URL 포함 |
| U-2 `.github/copilot-instructions.md` | Infra | 5분 | — | 파일 존재 + AGENTS.md 와 동일 룰 |
| U-4 `tene --help` 푸터 리소스 | Backend | 15분 | — | `tene --help | grep llms.txt` 매치 |

**W1 종료 gate**: 13개 항목 completed, 자동 검증 §6.1 중 8개 통과.

### 2.2 Milestone 2 — Week 2 구조 개선 (8h, 2026-05-01 ~ 05-07)

**목표**: 설치 채널 해금 + 신뢰 신호 4종 + CLI-level 안전 가드.

| 항목 | 담당 | 소요 | 의존성 | DoD |
|:---:|------|:---:|-------|-----|
| I-1a SECURITY.md | Security | 30분 | README `## Security` 이관 | GitHub "Security Policy" 활성화 |
| I-1b CODE_OF_CONDUCT.md | Infra | 15분 | Covenant v2.1 기본 | Community Health score 반영 |
| I-1c CONTRIBUTING.md | Infra | 45분 | README `## Contributing` 확장 | 신규 기여자 onboarding 가능 |
| I-1d `.github/ISSUE_TEMPLATE/bug_report.yml` + `feature_request.yml` | Infra | 30분 | — | New Issue 탭에 2 템플릿 표시 |
| I-1e `.github/PULL_REQUEST_TEMPLATE.md` | Infra | 10분 | — | 새 PR 본문 자동 채움 |
| **S-1** Homebrew tap 활성화 | Infra | 30분 | 신규 리포 `agent-kay-it/homebrew-tap`, `HOMEBREW_TAP_GITHUB_TOKEN` 시크릿 | `brew install agent-kay-it/tap/tene && tene version` |
| **S-2** Docker GHCR 배포 | Infra | 45분 | `Dockerfile` 생성, `packages: write` 권한 | `docker run ghcr.io/agent-kay-it/tene version` |
| S-4 man_pages + shell completion | Infra + Backend | 1h | Backend: `completion` 명령 등록 · mangen 구현 | archive 에 `completions/*`, `manpages/*` 포함 |
| **U-1** `tene get` 비대화형 차단 | Backend + Security | 2h | CHANGELOG + migration 노트 | `tene get X | cat` → exit 2, `STDOUT_SECRET_BLOCKED` 에러, `--unsafe-stdout` 또는 env 로 escape |
| W2 자동 검증 | QA | 30분 | 상기 11개 완료 | §6.1 체크리스트 80%+ pass |

**W2 종료 gate**: `brew install` · `docker run` 양쪽 작동, Community Health ≥ 85%, 자동 검증 §6.1 의 P0-P1 전수 통과.

### 2.3 Milestone 3 — Month 1 흥미·컨버전 개선 (10h, 2026-05-08 ~ 05-31)

**목표**: 랜딩 UX + 컨텐츠 자산 + 실측 KPI 셋업.

| 항목 | 담당 | 소요 | 의존성 | DoD |
|:---:|------|:---:|-------|-----|
| **I-2** Trust Section | Frontend | 3h | Shields.io 배지 URL, early user quote 수집 | 홈에 stars 라이브 + quote 1+ 렌더 |
| I-3 Breadcrumb 컴포넌트 | Frontend | 1h | — | `/vs/*`, `/blog/*` 상단에 trail 표시 |
| **U-5** `docs/cli-reference.md` 캐노니컬 페이지 + `/cli` 라우트 | QA + Frontend | 3-4h | 모든 명령 `--json` 스키마 수집 | `tene.sh/cli` 200, 모든 명령/flag/exit code 문서화 |
| S-5 GPG 서명 체크섬 (선택) | Security | 1h | GPG 키 생성 + public 배포 | `install.sh` 에 `gpg --verify` 경로 추가 |
| GSC 등록 + Sitemap 제출 | Growth | 30분 | D-2 완료 | "Impressions" 데이터 수집 시작 |
| Homebrew Core PR 검토 준비 | Infra | 1h | tap 30일 안정 · stars ≥ 75 확보 시점 | PR 초안 작성 |
| Dev.to 크로스포스트 4편 (canonical=tene.sh) | Growth | 4h (주 1편) | 블로그 MDX 4편 기존 활용 | 4편 발행, 각 canonical URL = tene.sh/blog/{slug} |

**M1 종료 gate**: §6.1 자동 검증 14개 중 13개 통과. Lighthouse Performance ≥ 90. Dev.to 4편 발행.

### 2.4 Milestone 4 — Quarter (8h + ongoing, 2026-06-01 ~ 07-23)

**목표**: AI 블랙박스 검증 1차 + 권위 백링크.

- XChaCha20 3rd-party audit 발주 ($2-5K) + 보고서 블로그
- Awesome-list PR 4건 병렬 제출 (mahseema/awesome-ai-tools, agarrharr/awesome-cli-apps, devsecops/awesome-devsecops, sbilly/awesome-security)
- HN Show HN 재런칭 (feature 완료 후 major milestone 으로)
- 블랙박스 테스트 8종 수행 (§6.2)
- KPI 재측정 → `docs/stats/ai-discoverability.md` 월별 누적

---

## 3. 의존성 그래프 (전체)

```
D-1 OG 이미지 ─────────┐ (독립)
D-2 robots.ts ─────────┤
D-3 Discussions ───────┤
D-4 ai-index + ai.json ┤
D-5 FUNDING.yml ───────┤
A-1 README 수정 ───────┤ → root.go 주석 해제 검토 (별도)
A-2 description ───────┤
A-3 aggregateRating ───┤
A-4 Organization LD ───┤
A-5 Tene is X ─────────┤ ← src/data/hero.ts 읽기 필요
S-3 install.sh ────────┤
U-2 copilot-instr ─────┤
U-4 tene --help ───────┘ → root.go 수정 (독립)

I-1 4파일 ─────┐ (SECURITY.md ← README §Security 참조)
               │
S-1 Homebrew ──┴── 신규 리포 생성 + PAT 시크릿
S-2 Docker ────┬── Dockerfile 선행 작성
               │
S-4 man/compl ─┴── Backend completion 명령 등록 → goreleaser 패키징
U-1 get.go ────── Security 설계 리뷰 → CHANGELOG → 배포

I-2 Trust ─────┬── (선행 없음)
I-3 Breadcrumb ┘
U-5 CLI ref ────── 모든 명령 --json 스키마 수집 (Backend 협조)
S-5 GPG ────────── GPG 키 생성 + public 배포
```

**Critical path (최단 배포 경로)**:
```
W1 Quick Wins (6h, 병렬) → W2 S-1 Homebrew + S-2 Docker + U-1 get.go (5h) → 배포 검증 → M1 Trust + CLI ref (7h) → 블랙박스 검증
```

---

## 4. 리소스 할당

### 4.1 인원

1인 solo maintainer 기준. Agent Teams 로 역할 분담 시뮬레이션 — 각 역할을 순차/병렬로 메인 에이전트가 에뮬레이션.

### 4.2 비용

- **필수**: $0 (OSS · 본인 작업)
- **선택 (M1-M2)**: $2,000-5,000 (XChaCha20 3rd-party audit, Zellic/Trail of Bits 마이크로 감사)
- **선택 (M3+)**: Homebrew Core PR 대기 중 광고 $0 유지

### 4.3 인프라

- GitHub Actions 분량 증가 (Docker build 추가로 ~2배) — free tier 내 처리 가능
- S3 `tene-releases` 버킷 기존 사용 지속
- 신규 public 리포 `agent-kay-it/homebrew-tap` (비용 0)
- GHCR (GitHub Container Registry) — public 무료

---

## 5. 승인 체크포인트 (Gate)

### 5.1 Pre-implementation (Design Gate)

- [ ] PRD (`docs/00-pm/ai-discoverability.prd.md`) 승인 → **tomo-kay**
- [ ] Plan (본 문서) 승인
- [ ] Design (`docs/02-design/features/ai-discoverability.design.md`) 승인
- [ ] Codebase Audit 리뷰 (`docs/02-design/features/ai-discoverability.codebase-audit.md`) 완료
- [ ] U-1 breaking change 위험 평가 + CHANGELOG 초안 리뷰

### 5.2 W1 종료 Gate

- [ ] 13개 P0 Quick Win 항목 completed
- [ ] `feature/ai-discoverability-audit` 브랜치에서 13개 커밋 또는 squash
- [ ] `git push origin feature/ai-discoverability-audit` + PR open (target: staging)
- [ ] `gh pr checks` CI 통과
- [ ] 자동 검증 §6.1 체크리스트 8/14 통과

### 5.3 W2 종료 Gate

- [ ] `brew install agent-kay-it/tap/tene` 실측 성공
- [ ] `docker run ghcr.io/agent-kay-it/tene version` 실측 성공
- [ ] Community Health ≥ 85% (`gh api .../community/profile`)
- [ ] U-1 breaking change 배포 전 CHANGELOG + release note 1차 초안 완료
- [ ] Security Architect 의 U-1 get.go diff 리뷰 pass

### 5.4 M1 종료 Gate

- [ ] Lighthouse Home Performance ≥ 90, SEO ≥ 95
- [ ] `tene.sh/cli` 200 OK, 모든 명령 문서화
- [ ] GSC 에 Sitemap 등록 + "Impressions" 데이터 수집 시작
- [ ] 블로그 Dev.to 크로스포스트 4편 모두 발행
- [ ] 자동 검증 §6.1 의 13/14 통과

### 5.5 M3 Blackbox Gate (2026-07-23)

- [ ] 블랙박스 테스트 8종 중 6종 이상 목표 달성
- [ ] GitHub stars ≥ 200 (Homebrew Core 요건 접근)
- [ ] 월 Homebrew install ≥ 500 (brew analytics)

---

## 6. 커뮤니케이션 & 리포팅

### 6.1 내부

- **매주 금요일 17:00 KST**: W+1 진행 상황 Slack 요약 (수동)
- **매일**: `docs/stats/github-stats.md` 자동 업데이트 (`/tene-stats`)
- **브랜치 push 시**: PR 자동 생성 + `gh pr checks` 모니터링

### 6.2 외부

- W2 종료 시: HN/Reddit 에 "shipped: brew install tene" 1포스트
- M1 종료 시: Product Hunt 업데이트 (major improvement 카테고리)
- M3 종료 시: 블로그 "tene 90일 성장 리포트" 발행

---

## 7. Contingency (리스크 발현 시 대응)

| 리스크 시나리오 | 대응 |
|---------------|-----|
| S-1 Homebrew tap PR 자동 생성 실패 (GoReleaser 토큰 이슈) | `HOMEBREW_TAP_GITHUB_TOKEN` 재생성 후 재릴리스 |
| U-1 배포 후 사용자 이슈 폭주 | 24시간 내 hotfix — `TENE_ALLOW_STDOUT_SECRETS=1` 명시화 + 에러 메시지 확장 |
| Docker GHCR 이미지 크기 > 50MB | multi-stage build 로 재빌드 (FROM scratch) |
| aggregateRating 제거 후 Google SERP 트래픽 감소 | 2주 관찰, 유의미하면 honest `Review` 배열로 복귀 |
| Community Health 파일이 Plugin/스팸 PR 유발 | `.github/CODEOWNERS` 로 기본 리뷰어 설정 |

---

## 8. 다음 Phase

이 Plan 승인 후:

1. **Codebase Audit** (`docs/02-design/features/ai-discoverability.codebase-audit.md`) — 수정 대상 파일 전수 맵핑, 현재 상태 증거
2. **Design** (`docs/02-design/features/ai-discoverability.design.md`) — 22개 항목 각각의 diff 수준 상세 설계

> Note: 기존에 생성된 `docs/01-plan/web-seo-aeo-analytics-strategy.md`, `docs/00-pm/web-seo-aeo-analytics-strategy.prd.md`, `docs/02-design/features/web-seo-aeo-analytics-strategy.*` 는 **이전 feature 의 유사 작업**. 본 feature (`ai-discoverability`) 는 감사 보고서 기반 재설계로, 기존 파일과 **별도 유지**. W2 완료 후 merge 여부 재평가.

---

**승인란**: tomo-kay __________________ (서명)  2026-__-__
