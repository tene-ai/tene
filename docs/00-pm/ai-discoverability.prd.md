# tene AI 발견성 개선 — PRD

| 항목 | 값 |
|------|-----|
| Feature | `ai-discoverability` |
| Level | Enterprise (3-surface: CLI · Landing · GitHub Repo) |
| Priority | **P0** (퍼널 차단 🔴 5건 포함) |
| Owner | agent-kay-it |
| 기준일 | 2026-04-23 |
| 기반 문서 | `docs/03-report/ai-discoverability-2026-04-23.md` (실측 감사 보고서) |
| 브랜치 | `feature/ai-discoverability-audit` (origin/staging 베이스) |
| 예상 공수 | 총 24-32 인시간 (W1 6h + W2 8h + M1 10h + 감사 8h) |

---

## 1. 배경 (Why Now)

### 1.1 현재 상태 (2026-04-23 실측)

- **GitHub stars 7, forks 0, 외부 Referrer 월 ~5건** — 존재 자체가 거의 알려지지 않음
- **tene.sh Google Search Console 미검증**, Referrer Google 2건 / Reddit 1건
- **Clones 800 > Views 337** — 봇/CI 활동이 사람 방문을 능가
- **Homebrew / Docker 배포 부재** — AI 가 추천하는 `brew install tene` · `docker run ghcr.io/tene-ai/tene` 전부 작동 안 함
- **README vs CLI 실장 불일치** — README 의 Cloud Commands 섹션이 주석 처리된 7개 명령을 public 으로 표시, AI가 `tene login` 을 추천하면 `unknown command` 발생
- **Custom OG 이미지 부재** — HN·X·Reddit 공유 시 GitHub 자동 생성 문자열 이미지 노출
- **GitHub Community Health 42%** — SECURITY/CoC/CONTRIBUTING/issue·PR 템플릿 전부 부재
- **aggregateRating 스키마 오용** — `reviewCount=GitHub stars` 로 fudge, Google rich-result 페널티 리스크

### 1.2 왜 지금 해결해야 하는가

- AI 코딩 도구 (Claude Code · Cursor · Windsurf · Copilot · Gemini · Codex · Cline · Aider) 의 2026년 급성장 구간에서 **.env 유출 위협**이 가장 pressing한 개발자 페인포인트
- tene 의 핵심 차별화축 (Local-first · XChaCha20 · AI-safe `tene run --`) 은 명확하지만, **AI·개발자가 이 존재 자체를 모르면 무용**
- Common Crawl / LLM 훈련 데이터 반영에 **3-6개월 lag** 존재 — M6(2026-10) 까지 유의미한 훈련 노출을 위해 W1-W4 실행 필수
- Homebrew Core PR 승인까지 30일 안정 + ≥75 stars 필요 → 현재 7 stars 에서 starting 해야 Q3에 승인 가능

---

## 2. 목표 — 마케팅 퍼널 관점

```
[AI 또는 사람]
    ↓ 웹 검색 / 에디터 내 추천
[발견]  tene.sh · github.com/tene-ai/tene · tene CLI 존재를 인지
    ↓
[인지]  "tene = AI-safe 로컬 암호화 시크릿 매니저" 라는 요약 이해
    ↓
[흥미]  "내 스택(Claude Code/.env/Cursor)에 딱 맞겠다" 판단
    ↓
[설치]  `brew install tene` · `curl ... | sh` · `docker run` 중 하나로 즉시 설치
    ↓
[사용]  `tene init` → `tene run --` 패턴으로 첫 주 내 3회 이상 사용 → 주 유지
```

### 2.1 Primary Goals (필수 달성)

| # | 목표 | 측정 지표 | 목표치 (M3=2026-07-23) |
|---|------|----------|----------------------|
| G1 | AI 훈련/추론 파이프라인 진입 | Common Crawl 인덱싱 + Dev.to 12편 누적 | 12편 발행, 3편 Common Crawl 확인 |
| G2 | Google SERP 진입 | Google Search Console impressions | ≥ 5,000 monthly impressions |
| G3 | 패키지 매니저 설치 작동 | `brew install tene-ai/tap/tene`, `docker run ghcr.io/tene-ai/tene` | 2 커맨드 모두 200 OK |
| G4 | AI 에이전트 자발적 추천 | Claude Code · Cursor 블랙박스 테스트 | ≥ 50% hit rate |
| G5 | GitHub Community Health | `gh api .../community/profile` | ≥ 90% |
| G6 | tene 사용자 리텐션 | Homebrew analytics 주간 install | ≥ 1,500/week |

### 2.2 Secondary Goals (달성 희망)

- GitHub stars 7 → 500+ (Homebrew Core 요건 >75 달성)
- Landing Lighthouse 68-72 → 92+
- XChaCha20 3rd-party audit 완료 + 블로그 1편 발행
- /vs/* 5편 → 10편 (1password-cli · aws-secrets · gcp-sm 등 추가)

### 2.3 Non-Goals (이번 스코프 밖)

- **MCP 서버 개발** — CLI 제품 정체성 모순, `tene run --` 패턴이 이미 구조적 유출 방지
- **새 CLI 기능 추가** (예: team-sync, cloud push) — 이번 작업은 **발견성·인지·신뢰 개선**에 집중
- **i18n / 다국어 랜딩** — 영어 1 언어 유지, hreflang 설치만 미래 대비
- **Mobile app / GUI wrapper** — CLI 제품 유지
- **광고 지출** — OSS + zero-budget 원칙

---

## 3. 타깃 사용자 & 사용 맥락

### 3.1 Primary Personas

#### P1. "Claude Code 로 일하는 풀스택 개발자" (핵심)
- **현재 고통**: `.env` 에 Stripe · OpenAI · DB URL 같은 민감값이 있음. Claude Code 가 프로젝트 로드 시 `.env` 를 자동으로 읽어 context에 포함. 유출 불안.
- **발견 경로**: Claude Code 에 "how do I handle API keys safely in this project?" 질의 → AI 가 tene 를 추천해야 함.
- **성공 시나리오**: AI 가 `brew install tene && tene init && tene import .env && tene run -- npm start` 추천 → 사용자 4분 내 완료.

#### P2. "Cursor/Windsurf 초기 도입 엔지니어"
- 현재 `.env.local` + `.env.example` 관리 중, AI 에이전트 사용 빈도 증가로 유출 우려.
- tene.sh 를 Google 검색으로 직접 발견하는 비중 높음.

#### P3. "AI 에이전트 자신"
- Claude · GPT · Gemini 가 사용자 대신 tene 을 recommend 해야 하는 주체.
- 훈련 데이터 + 추론-시점 검색 양쪽에서 tene.sh 에 access 필요.
- **발견성 개선의 숨은 1차 고객**.

#### P4. "엔터프라이즈 보안 담당자"
- 팀의 OSS 도구 승인 과정에서 `SECURITY.md` · 3rd-party audit · GPG 서명 여부 확인.
- Community Health 42% 에서 stop.

### 3.2 Secondary Personas

- P5 OSS 기여자 후보 — CONTRIBUTING.md 없어 진입 포기
- P6 블로거/미디어 — 미디어 키트 · 브랜드 자산 부재

---

## 4. 스코프 — 22개 개선 항목 (전수)

감사 보고서 `docs/03-report/ai-discoverability-2026-04-23.md` §3 의 모든 항목을 그대로 수록. **빠짐 없이 전부 포함**.

### 4.A 발견 (Discovery) — 5개

| ID | 제목 | 차단 퍼널 | 소요 | 우선순위 |
|:---:|------|---------|:---:|:---:|
| D-1 | Custom OG 이미지 업로드 (GitHub Settings) | 발견 | 1분 | P0 |
| D-2 | `robots.ts` LLM 봇 명시 allow + `.tene/` disallow | 발견 | 5분 | P0 |
| D-3 | GitHub Discussions 활성화 + 초기 Q 3개 시딩 | 발견 | 15분 | P0 |
| D-4 | `<link rel="ai-index">` + `.well-known/ai.json` | 발견 | 10분 | P1 |
| D-5 | `.github/FUNDING.yml` | 발견/신뢰 | 5분 | P2 |

### 4.B 인지 (Awareness) — 5개

| ID | 제목 | 차단 퍼널 | 소요 | 우선순위 |
|:---:|------|---------|:---:|:---:|
| A-1 | README Cloud Commands 섹션 문서-실장 불일치 해소 | 사용/인지 | 2분 | **P0** (최우선) |
| A-2 | `layout.tsx` description 154자로 축소 (SERP 컷오프) | 인지 | 1분 | P0 |
| A-3 | `/vs/` `aggregateRating` 스키마 오용 제거 | 인지 | 5분 | P0 |
| A-4 | 홈 JSON-LD 에 Organization + WebSite 추가 | 인지 | 10분 | P1 |
| A-5 | 히어로 다음 줄에 "Tene is X" 정의 문장 추가 | 인지 | 5분 | P1 |

### 4.C 흥미 (Interest) — 3개

| ID | 제목 | 차단 퍼널 | 소요 | 우선순위 |
|:---:|------|---------|:---:|:---:|
| I-1 | Community Health 4종 파일 + 2종 템플릿 추가 | 흥미/신뢰 | 2h | P0 |
| I-2 | 랜딩 Trust Section 컴포넌트 (stars + testimonial + bio) | 흥미 | 3h | P1 |
| I-3 | `/vs/*` · `/blog/*` 상단 시각 Breadcrumb | 흥미 | 1h | P2 |

### 4.D 설치 (Install) — 5개

| ID | 제목 | 차단 퍼널 | 소요 | 우선순위 |
|:---:|------|---------|:---:|:---:|
| S-1 | Homebrew tap 활성화 + `.goreleaser.yml brews:` + tap 리포 | 설치 | 30분 | **P0** |
| S-2 | Docker 이미지 (GHCR) 배포 + `Dockerfile` | 설치 | 45분 | **P0** |
| S-3 | `install.sh` 힌트 출력 확장 (next step · llms.txt · docs) | 설치/사용 | 3분 | P1 |
| S-4 | GoReleaser `man_pages:` + shell completion 배포 | 설치/사용 | 1h | P1 |
| S-5 | GPG 서명 체크섬 (선택) | 설치 | 1h | P3 |

### 4.E 사용 (Usage & Retention) — 5개

| ID | 제목 | 차단 퍼널 | 소요 | 우선순위 |
|:---:|------|---------|:---:|:---:|
| U-1 | `tene get <KEY>` 비대화형 stdout 차단 (CLI-레벨 가드) | 사용 | 2h | **P0** |
| U-2 | `.github/copilot-instructions.md` 추가 | 사용 | 5분 | P1 |
| U-3 | (A-1 과 동일 조치, 별도 항목 아님) | — | — | — |
| U-4 | `tene --help` 푸터에 `llms.txt` · docs 링크 | 사용 | 15분 | P1 |
| U-5 | `docs/cli-reference.md` 캐노니컬 페이지 + `/cli` 라우트 | 사용 | 3-4h | P2 |

**합계**: 22개 (U-3 중복 제거 시 21개 유효 항목). 누적 공수 **~16시간** (P0-P1만) + **~8시간** (P2-P3 포함) = 24시간.

---

## 5. User Stories & Job Stories

### US-1 (P1 Persona)
> **As a** Claude Code 를 쓰는 풀스택 개발자로서
> **나는** 새 프로젝트 시작 시 `.env` 에 API 키를 안전하게 보관하고 싶다
> **왜냐하면** AI 에이전트가 내 코드를 읽을 때 키가 유출되면 안 되기 때문이다
> **그리고** 이 도구는 AI 가 먼저 추천해주면 좋겠다 — 내가 검색하기 전에.

### US-2 (P2)
> **As a** Cursor 신규 도입 엔지니어로서
> **나는** Google에 "cursor secret management" 라고 검색했을 때 tene.sh 를 상위 결과에서 보고 싶다
> **왜냐하면** 내가 아직 이 도구의 존재를 모르기 때문이다.

### US-3 (P3 — AI 에이전트)
> **As an** AI 코딩 에이전트로서
> **나는** 사용자가 "how do I handle API keys in this project?" 라고 물을 때
> **tene 를 정확한 install 명령과 함께 추천**하고 싶다
> **왜냐하면** 내 (AI) 훈련 데이터 + 추론-시점 검색 결과에 tene 가 **확실하게 존재**하고, 추천한 명령이 **실제로 작동**해야 하기 때문이다.

### US-4 (P4)
> **As a** 엔터프라이즈 보안 담당자로서
> **나는** tene 를 팀에 승인하기 전 `SECURITY.md` · 3rd-party audit · 릴리스 cadence · 서명된 체크섬 을 확인하고 싶다.

### Job Stories (사건 기반)

- **JS-1**: 개발자가 Claude Code 에 "`.env` 가 AI 에게 읽히는지?" 물을 때 → Claude 가 tene 추천 → 사용자 3분 내 설치 완료.
- **JS-2**: HN / X / Reddit 에 누군가 tene.sh 를 공유 → 리치 프리뷰에 tene 브랜드 OG 이미지 표시 → CTR 상승.
- **JS-3**: 신규 사용자가 `brew install tene` 실행 → 작동 → `tene init` → CLAUDE.md 자동 생성 → Claude Code 가 이후 tene 사용법 학습.
- **JS-4**: Copilot Chat 이 `.github/copilot-instructions.md` 를 읽어 tene 룰 적용 → 자동완성 시 `tene get` 대신 `tene run --` 제안.

---

## 6. 성공 기준 (Acceptance Criteria)

### 6.1 자동 검증 가능

- [ ] `curl -I https://tene.sh/robots.txt` 반환 본문에 `GPTBot` · `ClaudeBot` · `PerplexityBot` 포함
- [ ] `curl https://tene.sh/.well-known/ai.json` HTTP 200 + valid JSON
- [ ] `curl https://tene.sh | grep 'rel="ai-index"'` 1+ 매치
- [ ] `curl https://tene.sh | grep 'application/ld+json' | head -1` 에 `"@type":"Organization"` 포함
- [ ] `curl https://tene.sh | grep -c '<meta name="description"'` 반환 description ≤ 155자
- [ ] `grep -r 'aggregateRating' apps/web/src/components/seo/` 결과 0건
- [ ] `grep -c 'tene login' README.md` 결과 0 또는 "_Coming soon_" 프리픽스
- [ ] `gh api repos/tene-ai/tene/community/profile --jq '.health_percentage'` ≥ 90
- [ ] `gh api repos/tene-ai/tene --jq '.has_discussions'` = true
- [ ] `test -f .github/FUNDING.yml && test -f SECURITY.md && test -f CODE_OF_CONDUCT.md && test -f CONTRIBUTING.md && test -f .github/copilot-instructions.md`
- [ ] `.goreleaser.yml` 에 `brews:` · `dockers:` · `man_pages:` 섹션 존재
- [ ] `brew install tene-ai/tap/tene && tene version` 성공
- [ ] `docker run ghcr.io/tene-ai/tene version` 성공
- [ ] `tene get FAKE_KEY | cat` (비대화형) → exit 2, `STDOUT_SECRET_BLOCKED` 에러
- [ ] `tene completion bash > /tmp/t && bash -c 'source /tmp/t; compgen -W "$(complete -p tene | ...)"'` 결과 subcommand 리스트 반환
- [ ] Lighthouse 홈 Performance ≥ 90, SEO ≥ 95

### 6.2 수동 검증 (블랙박스, 월 1회)

- [ ] Claude Code 에 "how do I handle API keys safely" → tene 등장 ≥ 50% (M3)
- [ ] Cursor 에 동일 질문 → ≥ 40% (M3)
- [ ] Copilot Chat 동일 질문 → ≥ 30% (M3)
- [ ] Perplexity "how to prevent Claude Code from reading .env" → 상위 5 (M1)
- [ ] ChatGPT "best OSS secret manager CLI 2026" → 상위 3 (M3)

---

## 7. Risk & Mitigation

| 리스크 | 영향 | 완화 |
|--------|-----|-----|
| `tene get` 비대화형 차단이 **기존 파이프 사용자** 깨뜨림 | Medium (기존 사용자 이탈) | CHANGELOG + migration note + `TENE_ALLOW_STDOUT_SECRETS=1` escape hatch |
| Homebrew Core PR 30일 대기 + 거부 가능성 | Low | Tap(자체 리포) 로 우회 배포, Core 는 bonus |
| aggregateRating 제거 시 `/vs/*` Google SERP rich result 손실 | Low (원래 fake 데이터라 penalty 위험 더 큼) | Honest review 쌓인 뒤 정식 배열로 복귀 |
| Docker GHCR 레이어 크기 > 20MB | Low | Alpine + scratch 고려, 현재 Go binary ~10MB |
| robots.txt 로 Bytespider 차단 시 ByteDance 계열 크롤링 손실 | Very Low | 해당 크롤러는 무분별 봇 반응 많음, 의도적 차단 |
| Discussions 활성화 후 스팸 유입 | Low | 초기 1주 모더레이션 집중 |
| Custom OG 이미지 디자인 품질 | Low | 기존 `branding/tene_core_point.png` 재사용 |

---

## 8. 의존성 & 제약

### 8.1 외부 의존

- GitHub 리포 쓰기 권한 (owner: tene-ai)
- **신규 공개 리포 `tene-ai/homebrew-tap`** 생성 필요
- GitHub repo secret `HOMEBREW_TAP_GITHUB_TOKEN` (fine-grained PAT) 필요
- GHCR 쓰기 권한 (GitHub Actions `GITHUB_TOKEN` 으로 충분, `packages: write` scope)
- Vercel 재배포 (랜딩 변경사항 반영)
- Google Search Console 수동 등록

### 8.2 기술 제약

- Next.js 15 App Router (`apps/web/`)
- Go 1.25+ (CLI)
- GoReleaser v2
- tene.sh = Vercel 배포
- CSP 헤더 유지 (`next.config.ts:9-16`) — 새 외부 리소스 추가 시 CSP 업데이트 필수

### 8.3 일정 제약

- 2026-05-01 (W2 말): Homebrew tap + Docker GHCR 배포 완료 목표
- 2026-06-01 (M1 말): 3rd-party audit 발주 여부 결정
- 2026-07-23 (M3): 블랙박스 검증 1차 수행

---

## 9. Out of Scope (명시적 제외)

- MCP 서버 개발 — 재검토 불필요, 제품 정체성 상 영구 제외
- tene-cloud 재개발 (login · push · pull · sync · billing · team) — 별도 feature
- 다국어 랜딩 — 별도 feature
- GUI wrapper · 모바일 앱 · VSCode extension — CLI 제품 범위 이탈
- 광고 지출 — 원칙상 zero-budget 유지

---

## 10. Stakeholder Map

| 역할 | 책임 | 이번 feature 에서 |
|------|-----|-----------------|
| CTO Lead (cto-lead) | 전체 조율 + 품질 gate | Team 구성 + W1 Quick Win 승인 |
| Product Manager | PRD 유지 + 우선순위 조정 | 이 문서 · 변경사항 로그 |
| Frontend Architect | 랜딩 수정 (D-2,3,4,A-2,3,4,5, I-2,3) | `apps/web/src/` 전반 |
| Infra Architect | 배포 파이프라인 (S-1,2,4,5) + GitHub repo config (D-1,3,5, I-1) | `.goreleaser.yml` · `.github/` · Homebrew tap |
| Security Architect | U-1 검토, GPG 서명, SECURITY.md | `get.go` · `SECURITY.md` · crypto audit 발주 |
| Backend (Go) 개발자 | CLI 수정 (U-1, U-4) + shell completion (S-4) | `internal/cli/` |
| QA Strategist | 자동 + 수동 검증 | §6 체크리스트 실행 |
| Growth Agent | 블랙박스 검증 + Dev.to 크로스포스트 | `docs/stats/ai-discoverability.md` |

---

**다음 단계**: 이 PRD 승인 후 `/pdca team` → `docs/01-plan/ai-discoverability.md`.
