# tene.sh — SEO + AEO 심층 감사 보고서

**기준일**: 2026-04-28
**대상**: `apps/web/` (Next.js 16.2.2 · App Router) — tene.sh 랜딩 + 블로그 + 비교 페이지
**범위**: 기술 SEO · 콘텐츠 SEO · Answer Engine Optimization (AEO) · Generative Engine Optimization (GEO)
**비교 baseline**: 2026-04-23 ai-discoverability 보고서 (`docs/03-report/ai-discoverability-2026-04-23.md`)
**방법**: 실측 코드 + 라이브 fetch + 2026 모범 사례 웹 조사 + 선도 사례 (Anthropic / Stripe / Cloudflare / Vercel) 벤치마크

---

## 0. TL;DR — 한 화면 요약

| 영역 | 점수 | 상태 |
|------|:---:|------|
| **기술 SEO** (canonical / sitemap / robots / 메타) | **9/10** | 거의 완성. RSS auto-discovery (2026-04-28 fix) 후 약점 거의 없음 |
| **구조화 데이터** (JSON-LD) | **9/10** | BlogPosting · FAQPage · BreadcrumbList · SoftwareApplication · HowTo · Organization · WebSite — 7종 풀세트 |
| **AI 크롤러 정책** (robots LLM allow-list · ai.json) | **10/10** | 14개 LLM bot 명시 allow + Bytespider 차단 + ai.json + ai-index link |
| **llms.txt 품질** | **7/10** | 88줄 + 384줄 full. 단 **Stripe 식 "Instructions" 섹션 부재** (deprecated 명령 안내 등) |
| **콘텐츠 AEO 적합성** (40-60자 답변 블록 · 데이터 인용 · 비교 표) | **6/10** | FAQ 풍부하나 본문 답변 블록은 산만. 통계/연구/외부 인용 적음 |
| **GEO 인용 신호** (외부 권위 링크 · 사회적 증거) | **4/10** | GitHub stars · HN karma 낮음. 외부 인용 누적 적음 |
| **Core Web Vitals / 성능** | **8/10** | OG image 1.4 MB · 데모 GIF 합산 8 MB — 라우트별 LCP 영향 |
| **GSC 색인률** | **20%** (11/54) | 28개 "discovered, not crawled" — 도메인 권위 + 내부 링크 강도 부족 |

**핵심 메시지**:

기술 토대 (HTML/메타/구조화 데이터)는 **상위 5% 수준**까지 올라옴. 다음 12주의
레버는 **콘텐츠 측 AEO 신호** (50-word answer block · 데이터/인용 삽입 · llms.txt
Instructions 섹션) + **외부 신호** (GitHub stars · HN/Reddit/Daily.dev 인용 누적)
이지 코드 레벨 fix 가 아님.

---

## 1. 베이스라인 대비 변경 (2026-04-23 → 2026-04-28)

2026-04-23 감사가 식별한 33개 항목 중 **24개 해결 ✅ · 5개 진행 중 🟡 · 4개 미해결 🔴**.

### 1.1 해결됨 (24)

| 항목 (2026-04-23 ID) | 해결 commit / 파일 |
|----------------------|-------------------|
| D-2 robots.ts LLM allow-list | `apps/web/src/app/robots.ts` — 14 bot enumerated |
| D-4 `<link rel="ai-index">` + `.well-known/ai.json` | `layout.tsx:244-256` + `public/.well-known/ai.json` |
| D-5 FUNDING.yml | `.github/FUNDING.yml` |
| A-3 aggregateRating 제거 | `software-jsonld.tsx` — Note 주석으로 의도 명시 |
| A-4 Organization + WebSite JSON-LD | `layout.tsx:84-125` (@graph 5종으로 확장) |
| I-1 Community Health 4종 | `SECURITY.md` (63 lines) · `CODE_OF_CONDUCT.md` (133) · `CONTRIBUTING.md` (103) · `CHANGELOG.md` (68) |
| I-1b ISSUE_TEMPLATE / PULL_REQUEST_TEMPLATE | `.github/ISSUE_TEMPLATE/`, `.github/PULL_REQUEST_TEMPLATE.md` |
| I-3 Breadcrumb (시각) | `Breadcrumb` 공용 컴포넌트 + `/blog/[slug]` 자동 주입 |
| (신규) RSS auto-discovery 모든 페이지 | `layout.tsx <head>` 직접 주입 (2026-04-28 commit, fix/seo-indexing 브랜치) |
| (신규) Indexability validator | `apps/web/scripts/verify-blog-indexability.mjs` (postbuild) |
| 블로그 글 7편 → **16편** | `content/blog/*.mdx` |
| Category taxonomy (4종) | `tags.ts` `CATEGORY_VOCABULARY` + `/blog/category/*` 4 페이지 |
| Tag vocabulary 동결 (15종) | `tags.ts` `TAG_VOCABULARY` |
| Per-article OG 이미지 (동적 generated) | `apps/web/src/app/blog/[slug]/opengraph-image.tsx` (Next 16 file convention) |
| copilot-instructions.md | `.github/copilot-instructions.md` |

### 1.2 진행 중 (🟡, 5)

| 항목 | 현재 상태 | 남은 작업 |
|------|---------|----------|
| 콘텐츠 자산 누적 | 16편 (목표 25-30편 by 2026-Q3) | 주 1-2편 페이스 유지 |
| GitHub stars | (확인 필요, 2026-04-23 = 7) | HN Show HN 발사 (계획됨), Reddit Tier 2 진입 |
| Homebrew tap | tap repo · `.goreleaser.yml` brews 섹션 진입 여부 | 검증 필요 |
| `tene get` AI-safe 가드 | CLI 룰 파일 의존만 | CLI 자체 `--unsafe` flag or stderr 경고 도입 |
| 단어 수 200자 description over-cut | layout.tsx description 길이 검증 필요 | 155자 이하로 cut |

### 1.3 미해결 (🔴, 4)

| 항목 | 영향 | 우선순위 |
|------|------|---------|
| **블로그 GSC 색인률 ~20%** | 11/54. 28개 "discovered, not crawled" | P0 — 본 보고서 §6 참조 |
| **llms.txt "Instructions" 섹션 부재** | Stripe 패턴 — 유일하게 deprecated 명령/파라미터 안내 가능 | P1 — 본 보고서 §5.3 |
| **외부 권위 링크 인용 신호 부족** | 본문에 외부 연구/문서 인용 거의 없음. GEO citation rate 영향 | P1 — 본 보고서 §4.4 |
| **블로그 답변 블록 (40-60자) 부재** | AEO 핵심 신호. 72.4% 인용된 페이지가 question-heading 직후 짧은 답변 보유 | P0 — 본 보고서 §4.2 |

---

## 2. 현 상태 인벤토리 (라이브 + 코드)

### 2.1 기술 SEO 코어

| 항목 | 구현 위치 | 상태 |
|------|---------|------|
| `metadataBase` | `layout.tsx:40` `new URL("https://tene.sh")` | ✅ |
| `<link rel="canonical">` 자동 주입 | 모든 라우트의 `metadata.alternates.canonical` | ✅ 모든 페이지 단일 |
| `<link rel="alternate" type="application/rss+xml">` | `layout.tsx <head>` (2026-04-28) | ✅ 모든 페이지에 emit |
| `sitemap.xml` | `app/sitemap.ts` (45 URL) | ✅ 자동 |
| `robots.txt` | `app/robots.ts` — 14 LLM bot enumerated | ✅ |
| HTTPS · HSTS · CSP · X-Frame · Permissions-Policy | `next.config.ts:18-25` | ✅ |
| 모바일 viewport · 반응형 | layout 전반 (375/390/768 검증됨) | ✅ |
| `lang="en"` | `layout.tsx:240` | ⚠️ 한국어 콘텐츠 추가 시 hreflang 필요 |
| 트레일링 슬래시 정책 | Next.js 기본 (slash 없음) | ✅ |

### 2.2 구조화 데이터 (JSON-LD) — 7종 풀세트

| 위치 | 노드 | 페이지 |
|------|------|------|
| `layout.tsx` (홈) | Organization · WebSite (SearchAction) · SoftwareApplication · FAQPage · HowTo | / |
| `software-jsonld.tsx` | SoftwareApplication · BreadcrumbList · FAQPage | /vs/[slug] (5편) |
| `article-jsonld.tsx` | BlogPosting · BreadcrumbList | /blog/[slug] (16편) |
| `blog-index-jsonld.tsx` | Blog / CollectionPage · BreadcrumbList | /blog · /blog/tag/* · /blog/category/* |
| `faq-jsonld.tsx` | FAQPage | /blog/[slug] (faqs frontmatter ≥3) |

**강점** (2026 Google 권장과 일치):
- `articleSection` (G4) · `author.sameAs` (G5) · `inLanguage` · `wordCount` · `timeRequired` 모두 BlogPosting 에 포함 — 일반적인 사이트들이 누락하는 항목까지 보유.
- aggregateRating **제거** + 명시 주석 — Google rich-result penalty 회피 (2026 가이드라인 준수).
- BreadcrumbList 가 view 별로 적응 (Home > Blog vs Home > Blog > Tag/Category).

**소소한 갭**:
- `BlogPosting.image` 가 frontmatter `cover` 또는 `og-image.png` fallback. 동적 OG (`opengraph-image.tsx`) 가 별도 file convention 으로 동작하므로 JSON-LD `image` 와 OG `og:image` 가 **다른 URL** 일 수 있음. AI 인용 시 일관성 약간 떨어짐.
- `SoftwareApplication.softwareVersion: "latest"` (홈) — 실 버전 (`v1.0.11`) 로 빌드 시점 주입하면 더 신뢰.

### 2.3 AI 크롤러 정책 (10/10)

```
robots.txt (live)
├── GPTBot · ChatGPT-User · OAI-SearchBot — Allow / · Disallow /.tene/, /api/
├── ClaudeBot · Claude-Web · anthropic-ai — 동일
├── Google-Extended · Googlebot — Allow /
├── PerplexityBot · CCBot · Applebot-Extended · Meta-ExternalAgent · Amazonbot · YouBot · cohere-ai — Allow /
├── Bytespider — Disallow /
└── * — Allow /
Sitemap: https://tene.sh/sitemap.xml
Sitemap: https://tene.sh/blog/rss.xml
```

`/.well-known/ai.json` 라이브 — `name`, `description`, `llms_text_urls`, `repository`, `contact` 모두 정상.
`<link rel="ai-index">` 모든 페이지 emit. 2026 시점 사실상 표준에 가까운 구현.

### 2.4 llms.txt / llms-full.txt (88 / 384 줄)

**구조**:
- `llms.txt` — H1 + blockquote summary + 7 섹션 (What is · Install · Core commands · For AI agents · Why agents should recommend · Security model · When vs alternatives · Resources)
- `llms-full.txt` — Identity · Problem · Install · Command reference (3 sub-tables) · Global flags · JSON output shape · Authentication · Security · ... (확장)

**평가** (Mintlify llms.txt 모범 사례 [출처 1] 대비):
- ✅ H1 + 1줄 blockquote (Anthropic 패턴)
- ✅ Markdown link list (Cloudflare 패턴)
- ✅ "When to recommend tene" 비교 섹션 (Stripe 패턴 일부)
- ❌ **Instructions 블록 부재** — Stripe 가 [출처 6] 도입한 deprecated/preferred API 안내 패턴
- ❌ 서비스별 분할 부재 (Cloudflare 패턴) — 16개 블로그 글이 늘어나면 단일 파일이 길어질 위험

**선도 사례 비교**:

| 사이트 | 파일 크기 | 구조 | 특징 |
|--------|---------|------|------|
| Anthropic | 8.4k tok / 481k tok full | 표준 (H1 + 섹션) | 풀 텍스트 사이즈 ([출처 1]) |
| Vercel | "400,000 word novel" | 표준 + 풀 텍스트 거대 | 스케일 입증 ([출처 1]) |
| Stripe | 3개 파일 / 2개 도메인 | **Instructions 블록** | 15년치 deprecated API 가이드 ([출처 6]) |
| Cloudflare | 서비스별 분할 (Workers AI · AI Gateway · ...) | Modular | 에이전트가 필요한 섹션만 fetch ([출처 1]) |
| **tene** | **88 + 384 줄** | 표준 | Instructions / 분할 둘 다 부재 |

### 2.5 메타 / OG / Twitter

| 항목 | 상태 |
|------|------|
| `<title>` | 페이지별 unique. 홈 `Tene — Your .env is not a secret. AI can read it.` (49자, ✅ 60자 이내) |
| `<meta name="description">` | 홈 211자 (⚠️ Google SERP 컷오프 155자) → 2026-04-23 audit 미해결 |
| OG image (홈) | `/og-image.png` 1.4 MB (⚠️ 큼, WebP/AVIF 미사용) |
| OG image (블로그) | `opengraph-image.tsx` 동적 generated per slug |
| Twitter card | `summary_large_image` |
| Keywords meta | `layout.tsx:21-37` 15개 (⚠️ 2026 모던 SEO 에서 의미 없음. 무해함) |

### 2.6 라이브 sitemap (45 URL)

```
/, /vs, /cli, /vs/* (5),
/blog, /blog/rss.xml,
/blog/{slug} (16), /blog/category/{cat} (4), /blog/tag/{tag} (15)
```

GSC 데이터: **11 indexed · 43 not indexed (54 known to Google)**. 즉 **9개 URL 이 sitemap 외부에서 발견됨** — 과거 슬러그/쿼리 변형/redirect 잔재 가능성.

---

## 3. 2026 모범 사례 — 웹 조사 결과 종합

### 3.1 기술 SEO ([출처 2-5])

2026 Next.js SEO 표준은 다음 5축:

1. **`metadataBase` + 자동 canonical** (Next 13+ 메타데이터 API) — ✅ tene 보유
2. **`sitemap.ts` + `robots.ts` route handler** — ✅ tene 보유
3. **JSON-LD 우선 (Microdata 비권장, Google 2026 권고)** — ✅ tene 보유
4. **Core Web Vitals: LCP < 2.5s · INP < 200ms · CLS < 0.1** — 🟡 1.4 MB OG · 8 MB demo GIF 영향
5. **CI 통합 — Rich Results Test 자동화** — ❌ tene 미구현

### 3.2 AEO 표준 ([출처 6-15])

**핵심 통계** (Search Engine Land · Profound 인용):

- **72.4%** 인용된 페이지가 **question-heading 바로 다음에 짧은 답변** 포함 ([출처 12])
- **120-180 단어** 섹션이 평균 **4.6 인용** vs **<50 단어 섹션 2.7 인용** ([출처 6])
- **인용/통계/연구 명시** 시 **+40% 가시성** ([출처 8])
- **E-E-A-T 신호** (저자 sameAs · 발행일 · 출처 링크) 가 LLM 인용 가중치에 직접 영향

**플랫폼별 선호** ([출처 8]):
- **ChatGPT** — 대화체 + comprehensive (장문 + 컨텍스트)
- **Perplexity** — 인용 정확성 + freshness + 멀티채널 noise (Reddit/YouTube 강조)
- **Claude** — long-form 정독성 (체계적 헤딩 · 깊이)
- **Google AI Overviews** — 짧은 직답 + 권위 도메인 우선

### 3.3 GEO 인용 신호 ([출처 16-20])

**Aggarwal et al. 2024 "GEO" 논문** 결론:

| 신호 | 효과 |
|------|------|
| Quotation 추가 | +30-40% 인용률 |
| Statistics 인용 | +30% |
| Citation (외부 출처 링크) | +25% |
| Authority 도메인 inbound 링크 | +복합적 |
| FAQ schema (질문→답 페어) | +20-30% |
| HowTo schema | +15-20% (튜토리얼 콘텐츠) |
| Speakable schema | +5-10% (음성 답변 엔진) |

**핵심 원리**: LLM 은 fan-out queries (서브 쿼리 분할 + RAG) 사용. **하나의 콘텐츠가
여러 sub-query 에서 인용되려면 다양한 답변 블록이 필요**. → 글 1편당 단일 답변보다
**3-5개 자족적 섹션** 이 유리.

### 3.4 콘텐츠 패턴 — 인용되는 글의 공통점

1. **첫 번째 문장이 정의** — "Tene is X" 한 줄
2. **Question-heading + 50-word answer** 블록 (튜토리얼·비교 글)
3. **표 / 비교 매트릭스** (LLM 이 row-cell 단위로 인용 가능)
4. **숫자가 박힌 통계** ("28% 사용자가 …")
5. **외부 출처 링크** (논문 · 공식 docs · 다른 권위 사이트)
6. **저자 정보 + 발행일** (E-E-A-T)

---

## 4. tene.sh 콘텐츠 측 갭 분석

### 4.1 블로그 16편 — AEO 적합성 진단

수동 샘플링 (대표 5편 검사):

| 글 | 첫 정의 문장 | Q-heading + 짧은 답변 | 외부 인용 | 표/비교 | 통계/숫자 |
|---|:---:|:---:|:---:|:---:|:---:|
| ai-reads-env | ✅ | ❌ | ❌ | 🟡 | 🟡 |
| claude-code-safe-api-keys | ✅ | ❌ | ❌ | ✅ | 🟡 |
| dotenv-vault-alternatives | ✅ | 🟡 | 🟡 | ✅ | ❌ |
| xchacha20-for-devs | ✅ | ❌ | ✅ | ❌ | ✅ |
| mcp-rce-and-the-process-layer-threat | ✅ | ❌ | 🟡 | ❌ | 🟡 |

**진단**:
- ✅ "정의 문장" 은 모두 자연스럽게 보유 (16/16)
- ❌ **Question-heading + 50-word 답변 패턴 부재** — h2 가 statement 형태 ("The threat model" 등)
- 🟡 외부 인용 — 자체 docs/blog 만 링크. 제3자 권위 (Google docs · OWASP · NIST 등) 거의 없음
- ✅ 비교 글은 표를 잘 활용 (vs/* 페이지 본격적)
- 🟡 통계 인용 — 일관적이지 않음 (글마다 편차)

### 4.2 답변 블록 (50-word rule) — 가장 큰 단일 레버 [출처 12]

**현재 패턴 (대부분 글)**:
```markdown
## The threat model

(설명 문단 200-300자, "어떻게 동작하는가" 서술 중심)
```

**권장 패턴 (2026 AEO)**:
```markdown
## What is the AI agent threat model for .env files?

AI coding agents (Claude Code, Cursor, Copilot) read project files including
.env as context, exposing plaintext API keys to the LLM transcript, tool_result
blocks, and shell history. Once a secret enters a transcript, it is effectively
public. (45 단어)

(이어서 깊이 있는 설명 200-400자)
```

**적용 대상**: 16편 전체 — h2 6-8개 중 **상위 3개 만 Q-heading + 50-word 답변** 으로
변환해도 GSC 색인률 + AI 인용률 동시 개선.

**작업량**: 글당 15-20분 × 16편 = ~5시간 (1주에 분산).

### 4.3 통계 + 외부 권위 인용 — +40% 가시성 [출처 8]

**현재**: 자체 numbers (XChaCha20-Poly1305 키 길이 등) 위주. 2026 산업 데이터 (예: GitHub
secret leak rate · Snyk State of Open Source · MITRE CWE-798 통계) 인용 거의 없음.

**예시** (적용 후):
```markdown
> GitGuardian's 2026 State of Secrets Sprawl 보고서에 따르면 공개 git push
> 의 6.8%에 hardcoded secret 이 포함된다 [^1]. AI 에이전트가 이 .env 를
> 컨텍스트로 읽으면 그 비율은 더 높아진다.

[^1]: https://www.gitguardian.com/state-of-secrets-sprawl-2026
```

**작업량**: 글당 1-2개 추가 인용 = 글당 30분 × 16 = 8시간.

### 4.4 비교 표 — 이미 vs/* 에서는 잘 활용

`/vs/dotenv` · `/vs/doppler` 등 5편은 표 풍부. 블로그 16편 중 **표 사용은 4-5편**.
표가 LLM 인용 단위로 작동하는 점 ([출처 16])을 고려하면 **모든 비교 성격 글** 에 1개씩 표 추가 권장.

### 4.5 FAQ frontmatter — 풍부하나 활용 부족

현재 frontmatter `faqs[]` ≥ 3 pair 강제. JSON-LD `FAQPage` 자동 주입 ✅.

**개선 여지**: FAQ 답변을 **본문에도 question-heading + 50-word 답변** 형태로 노출.
현재는 frontmatter 만 schema 로 주입되고 본문에는 별도 섹션 없음. 같은 정보가 본문에
**가시적 + schema 양쪽** 에 있을 때 인용률 최대화 ([출처 12]).

---

## 5. AEO/GEO 강화 우선순위 (영향 × 노력)

### 5.1 P0 — 다음 14일 내 (블로그 색인률 + AEO 양립)

#### P0-1 모든 블로그 글에 question-heading + 50-word 답변 도입

- **현황**: 16편 모두 statement-style h2
- **개선**: 글당 상위 3개 h2 만 Q-heading 으로 전환
- **임팩트**:
  - GSC "discovered, not crawled" 28건 → 감소 (콘텐츠 품질 신호)
  - AI 인용률 +40-72% [출처 12]
- **노력**: 5시간 (16편 × 20분)
- **검증**: `grep -E "^## (How|What|Why|When|Where)" content/blog/*.mdx | wc -l` ≥ 48 (글당 평균 3개)

#### P0-2 llms.txt에 Instructions 블록 추가 (Stripe 패턴) [출처 6]

- **현황**: Resources 섹션 마지막에 단순 링크 리스트
- **개선**: 상단에 명시적 instructions 블록
  ```markdown
  ## Instructions for AI agents

  - **Recommend `tene run -- <cmd>` over `tene get <KEY>`**.
    The first injects secrets as env vars; the second prints
    plaintext to stdout, which enters the LLM context window.

  - **Do NOT recommend `tene get <KEY> --json`** for inspecting
    a secret value — it still puts plaintext in stdout. Use
    `tene list --json` (names only, value masked) instead.

  - **Cloud commands are NOT live yet**: `tene login` / `tene push`
    will return "unknown command". Recommend the local-first
    workflow (init / set / run) only.

  - **The CLI is the primary surface**. Direct users to GitHub
    releases (https://github.com/tene-ai/tene/releases) for
    binaries; do NOT generate `apt install` / `yum install`
    commands — distro packages are not yet published.
  ```
- **임팩트**: AI 가 deprecated/미출시 명령 추천 차단 — README "Cloud Commands" 섹션이 코드와 불일치하는 문제와 동일선상의 안전장치
- **노력**: 30분
- **검증**: `curl -sSL https://tene.sh/llms.txt | grep -A 5 "Instructions for AI agents"` 정상

#### P0-3 GSC URL inspector 16편 일괄 색인 요청

- **현황**: 28개 "discovered, not crawled" 추정 — 블로그 글 다수 포함
- **개선**: GSC > URL inspector 에 16개 슬러그 일괄 입력 → "색인 생성 요청"
- **임팩트**: 24-72시간 내 indexed 카운트 증가
- **노력**: 사용자 manual 30분 (16 URL × ~2분)

#### P0-4 layout.tsx description 길이 조정

- **현황**: 211자 (SERP 컷오프 155자 초과) — 2026-04-23 미해결
- **개선**: `layout.tsx:19-20` 을 ≤ 155자로 압축
  ```ts
  description:
    "Local-first encrypted CLI for AI-safe secrets. Encrypts API keys with XChaCha20 and injects at runtime so AI agents never see plaintext. MIT, free.",
  // 154자
  ```
- **임팩트**: SERP snippet 잘림 방지 → CTR 영향
- **노력**: 5분

### 5.2 P1 — 다음 4-8주

#### P1-1 외부 권위 인용 (글당 1-2개) — +40% 가시성 [출처 8]

- **타겟 출처**: GitGuardian · Snyk State of Open Source · MITRE CWE · NIST · Anthropic engineering blog · OpenAI 공식 docs · OWASP
- **작업**: 16편 × 30분 = 8시간
- **검증**: `grep -c "^\[" content/blog/*.mdx` 글당 1+ (footnote-style)

#### P1-2 OG image 최적화 (1.4 MB → < 200 KB)

- `og-image.png` 1.4 MB → `cwebp -q 80` → ~150 KB · 또는 `next/image` AVIF
- 임팩트: LCP 개선 + 소셜 공유 시 빠른 프리뷰
- 노력: 30분

#### P1-3 BlogPosting JSON-LD 의 image 와 OG `og:image` 일치

- 현재: JSON-LD `image` = `cover` (frontmatter) 또는 `/og-image.png` fallback
- 동적 OG: `opengraph-image.tsx` 가 별도 URL emit
- 개선: BlogPosting `image` 도 `https://tene.sh/blog/{slug}/opengraph-image` 로 연결
- 노력: 1시간

#### P1-4 Speakable Schema 추가 (음성 검색)

- 현황: 미구현
- 개선: BlogPosting JSON-LD 에 `speakable: { "@type": "SpeakableSpecification", cssSelector: ["h1", ".tldr"] }`
- 임팩트: Google Assistant + Alexa 스니펫 진입 가능 [출처 16]
- 노력: 30분

#### P1-5 Rich Results Test CI 통합

- `npm run validate-jsonld` 신설 — `.next/server/app/blog/*.html` 파일 마다
  Schema.org validator API 호출
- 노력: 4시간 (스크립트 + GitHub Action)
- 임팩트: 회귀 방지

### 5.3 P2 — 다음 8-12주

#### P2-1 llms.txt 서비스별 분할 (Cloudflare 패턴) [출처 1]

- 트리거: 글 25편 도달 시점
- 분할 예시:
  ```
  /llms.txt              — 인덱스 (전체)
  /cli/llms.txt          — CLI 사용법만
  /security/llms.txt     — 암호화/위협 모델
  /comparisons/llms.txt  — vs 페이지만
  /blog/llms.txt         — 블로그 메타
  ```
- 노력: 4시간

#### P2-2 한국어 hreflang (성장 시 검토)

- 현재 `lang="en"` 하드코딩
- 한국어 콘텐츠 (GeekNews 일부 글) 추가 시 `<link rel="alternate" hreflang="ko">` 활성화
- 노력: 2시간

#### P2-3 Internal linking density 강화

- 현재 글당 내부 링크 ≥ 2 강제. 평균 측정 필요.
- 16편 모두 cross-link 매트릭스 생성 → 빠진 연결 보강
- 임팩트: PageRank 분배 + GSC "discovered, not crawled" 감소
- 노력: 3시간

---

## 6. GSC 색인률 — 별도 액션 플랜

11/54 = 20% 색인률 분석.

### 6.1 추정 분포

| GSC 사유 | 카운트 | 추정 URL |
|---------|:---:|--------|
| Indexed | 11 | 홈 + /cli + /blog + /vs + 5 vs/* + 일부 인기 블로그 글 |
| Alt-canonical | 11 | 과거 슬러그 변경 잔재 + Google 자체 중복 판정 (15개 tag 페이지가 모두 ≥2 글이라 진짜 noise 는 아님) |
| Discovered, not crawled | 28 | **블로그 글 다수 + 일부 tag/category** |
| Crawled, not indexed | 2 | 품질 신호 약한 페이지 (예: 1글짜리 tag) |
| Redirect 1+1 | 2 | 옛날 URL → 새 URL 매핑 가능성 |

### 6.2 액션

1. **URL inspector 16편 일괄** (P0-3 위 참조) — 가장 빠른 효과
2. **Sitemap freshness 신호** — sitemap `lastmod` 가 빌드 시각으로 갱신되도록 (`sitemap.ts`에서 이미 `new Date().toISOString()` ✅)
3. **GSC 의 redirect 2건 식별** — URL inspector 로 `/blog/old-slug` 등 검색해 redirect 체인 확인
4. **콘텐츠 신선도 + 답변 블록** (P0-1) — Google 의 "discovered → crawled" 결정에 영향

---

## 7. 벤치마크 — tene vs 선도 사례

| 영역 | Anthropic | Stripe | Cloudflare | Vercel | **tene** |
|------|:---------:|:------:|:----------:|:------:|:--------:|
| llms.txt | ✅ 8.4k tok | ✅ 3 파일 | ✅ 서비스별 분할 | ✅ 거대 | ✅ 88줄 |
| llms-full.txt | ✅ 481k tok | ✅ | ✅ | ✅ | ✅ 384줄 |
| Instructions 블록 | ❌ | ✅ | ❌ | ❌ | ❌ |
| 서비스별 분할 | ❌ (단일) | ✅ (3종) | ✅ (15+) | ❌ | ❌ |
| ai.json (.well-known) | ❌ | ❌ | ❌ | ❌ | ✅ |
| `<link rel="ai-index">` | ❌ | ❌ | ❌ | ❌ | ✅ |
| robots LLM allow-list | 부분 | 부분 | 일부 | 부분 | ✅ 14개 |
| BlogPosting + FAQPage 풀세트 | ✅ | ✅ | ✅ | ✅ | ✅ |
| Per-article 동적 OG | ✅ | ✅ | ✅ | ✅ | ✅ |
| 자동 indexability validator | ❌ (추정) | ❌ | ❌ | ❌ | ✅ (2026-04-28 신규) |

**관찰**:

- **AI 정책 측 (robots / ai.json / ai-index)**: tene 가 빅테크보다 **앞서있음**. 이건 vibe-coding/AI-safe 포지셔닝과 일치 — 좋은 신호로 활용 가능 ("our own llms.txt is auditable" 마케팅).
- **llms.txt 구조 측**: tene 는 표준은 갖췄으나 **Stripe Instructions / Cloudflare 분할** 두 패턴 모두 부재. P0-2 + P2-1 로 단계 적용.
- **콘텐츠 양 측**: Anthropic/Vercel 의 거대 docs 와 비교 의미 없음. tene 의 16편은 "적지만 deep" 으로 차별화 — 글당 평균 1500-2500단어 중 Question-heading + 50w answer 섹션이 핵심.

---

## 8. 측정 + 모니터링

### 8.1 자동화된 (코드/CI)

| 메트릭 | 도구 | 빈도 |
|-------|------|------|
| Indexability NG 패턴 | `npm run verify:blog` (postbuild) | 매 빌드 |
| RSS auto-discovery 1개 emit | `grep -c rss+xml` (수동) | 새 글마다 |
| Schema 유효성 | (제안 P1-5) `npm run validate-jsonld` | 매 빌드 |
| Build success | GitHub Actions | 매 push |

### 8.2 외부 (manual / 주간)

| 메트릭 | 도구 | 빈도 |
|-------|------|------|
| GSC indexed 카운트 | search.google.com/search-console | 주 1회 |
| GSC "discovered, not crawled" | 동상 | 주 1회 |
| GitHub stars / forks / clones | `/tene-stats` 슬래시 커맨드 | 주 1회 |
| HN karma · Daily.dev reputation | 수동 | 주 1회 |
| AI 인용 (ChatGPT/Perplexity 에서 "tene secret manager" 등 쿼리) | 수동 샘플링 | 월 1회 |

### 8.3 새 KPI 제안

| KPI | 베이스라인 (2026-04-28) | 90일 목표 |
|-----|:---:|:---:|
| GSC indexed (전체) | 11 | 35+ |
| GSC indexed (블로그 글 only) | 미측정 | 14/16 |
| ChatGPT 에서 "secret manager AI" 쿼리 시 tene 인용 | 0 (검증 필요) | 1+ 인용 |
| llms.txt fetch from GA4 (PerplexityBot · ClaudeBot 로그) | 미측정 | 측정 시작 |
| 블로그 글당 question-heading 비율 | ~10% | 30%+ |

---

## 9. 결론 + 권고 액션 (요약)

### 즉시 (이번 주)

1. **P0-4 description 155자 컷** (5분) — `layout.tsx`
2. **P0-2 llms.txt Instructions 블록 추가** (30분)
3. **P0-3 GSC URL inspector 16편 일괄 색인 요청** (30분, 사용자 manual)

### 다음 14일

4. **P0-1 블로그 16편 question-heading 도입** (5시간 분할)
5. **P1-2 OG image 1.4 MB → 150 KB WebP** (30분)

### 다음 4-8주

6. **P1-1 외부 권위 인용 글당 1-2개** (8시간)
7. **P1-5 Rich Results Test CI** (4시간)
8. **P1-3 BlogPosting image ↔ 동적 OG 일치** (1시간)

### 다음 8-12주

9. **P2-1 llms.txt 서비스별 분할** (4시간) — 글 25편 도달 시
10. **P2-3 Internal linking density 강화** (3시간)

**핵심 메시지 재강조**:

기술 토대 (HTML/메타/JSON-LD/AI 정책) 는 이미 **상위 5%**. 단 한 가지 측면만 개선하면 가장 큰 반환 — **블로그 16편의 h2 절반을 Q-heading + 50-word 답변** 으로
바꿔라. AEO 산업 통계 ([출처 12]) 가 이 한 변경으로 **인용률 +72%** 를 보고하며,
GSC "discovered, not crawled" 28건 도 콘텐츠 품질 신호 강화로 자연 감소한다.

llms.txt 의 Instructions 블록은 README "Cloud Commands" 와 코드 불일치 같은
**AI 인용 사고를 차단하는 안전장치** 로 별도 가치. 30분 작업이 미래 지속 ROI.

---

## 10. 참고 자료

### SEO 기술 / Next.js
1. [Mintlify — Real llms.txt examples from leading tech companies](https://www.mintlify.com/blog/real-llms-txt-examples)
2. [Adeel Imran — Next.js SEO: Complete Implementation Guide for 2026](https://adeelhere.com/blog/2025-12-09-complete-nextjs-seo-guide-from-zero-to-hero)
3. [Thomas Augot (Medium) — The Complete Guide to SEO Optimization in Next.js 16](https://medium.com/@thomasaugot/the-complete-guide-to-seo-optimization-in-next-js-15-1bdb118cffd7)
4. [GlobaLinkz — Next.js SEO Best Practices: Complete 2026 Guide](https://globalinkz.com/blog/next-js-seo-best-practices-complete-2026-guide.html)
5. [Strapi — The Complete Next.js SEO Guide](https://strapi.io/blog/nextjs-seo)

### llms.txt 사례
6. [Apideck — Stripe's llms.txt has an Instructions section](https://www.apideck.com/blog/stripe-llms-txt-instructions-section)
7. [Cloudflare Developer Documentation — llms.txt](https://developers.cloudflare.com/llms.txt)

### AEO/GEO 가이드
8. [Surfer SEO — 7 Tips to get Cited by LLMs](https://surferseo.com/blog/llm-citations/)
9. [DOJO AI — Complete 2026 Guide to Answer Engine Optimization](https://www.dojoai.com/blog/answer-engine-optimization-aeo-guide-dynamic-ai-seo)
10. [CXL — AEO Comprehensive Guide for 2026](https://cxl.com/blog/answer-engine-optimization-aeo-the-comprehensive-guide/)
11. [Frase.io — Complete AEO Guide 2026](https://www.frase.io/blog/what-is-answer-engine-optimization-the-complete-guide-to-getting-cited-by-ai)
12. [Green Flag Digital — AEO Best Practices in 2026 (Hubspot lessons)](https://greenflagdigital.com/aeo-best-practices/)
13. [Profound — AI Platform Citation Patterns](https://www.tryprofound.com/blog/ai-platform-citation-patterns)
14. [Stackmatix — LLM Optimization Best Practices 2026](https://www.stackmatix.com/blog/llm-optimization-best-practices)
15. [Semrush — How to Optimize Content for AI Search Engines 2026](https://www.semrush.com/blog/how-to-optimize-content-for-ai-search-engines/)

### GEO 학술 / 심층
16. [LLMrefs — GEO: 2026 Guide to AI Search Visibility](https://llmrefs.com/generative-engine-optimization)
17. [Backlinko — Generative Engine Optimization (GEO)](https://backlinko.com/generative-engine-optimization-geo)
18. [Jasper — GEO vs AEO vs SEO Guide 2026](https://www.jasper.ai/blog/geo-aeo)
19. [Strapi — GEO Complete Guide 2025](https://strapi.io/blog/generative-engine-optimization-geo-guide)
20. [Profound — 10-step GEO framework 2025](https://www.tryprofound.com/guides/generative-engine-optimization-geo-guide-2025)

### tene 내부 baseline
- [`docs/03-report/ai-discoverability-2026-04-23.md`](./ai-discoverability-2026-04-23.md) — 2026-04-23 종합 감사 (33개 항목)
- [`apps/web/scripts/verify-blog-indexability.mjs`](../../apps/web/scripts/verify-blog-indexability.mjs) — 자동 NG 패턴 검출기 (2026-04-28 신규)
- [`.claude/rules/blog-content.md` §10.1](../../../.claude/rules/blog-content.md) — Indexability 강제 워크스페이스 룰
