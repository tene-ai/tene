# Blog Categorization + /blog-new Skill Enhancement — Plan

| 항목 | 값 |
|------|-----|
| Feature | `blog-categories-and-tooling` |
| 기반 논의 | 2026-04-23 대화 세션 (PRD 없음 — 점진 개선) |
| 브랜치 | `feature/blog-categories` (origin/staging 베이스) |
| 후속 feature | `blog-content-batch-2026-04-23` (별도 plan 문서) |
| 총 공수 | ~10-12 인시간 |
| 기간 | 1-2일 집중 |

---

## 1. 컨텍스트

### 1.1 문제 (As-is)

현재 블로그(`apps/web/content/blog/*.mdx`)는 **단일 평면 tag vocabulary 10개**만 운영:

```
security · ai · cli · go · devsecops · cryptography ·
tutorial · comparison · architecture · vibe-coding
```

- 제품/기술 중심이라 **개념·철학 글을 담을 카테고리 부재**
- `ai`가 너무 포괄적 — "AI 보조 개발"과 "AI 노동 철학"이 같은 tag로 묶임
- `go`는 1편만 사용 (과잉 세분화)
- 탐색 단위가 tag 하나뿐 — 독자가 "이 블로그는 무엇에 관한 블로그인가" 파악 어려움

### 1.2 목표 (To-be)

**2-layer taxonomy** 도입:

- **Category (필수 1개)** — 4개 버킷. 블로그 탐색의 primary 축.
  - `tools`, `engineering`, `vibe-coding`, `philosophy`
- **Tag (2-4개 복수 선택)** — 15개 curated vocabulary. 세부 주제 분류 + 교차 탐색.

독자 경험:
- 카테고리 4개 pill로 즉시 콘텐츠 정체성 이해
- 인스타그램식 tag 필터로 교차 탐색 (`#security` + `#claude-code`)
- 기존 URL 전부 보존 (SEO 자산 보호)

### 1.3 확정 결정 (대화 기반)

사용자와 확정된 사항:

| 질문 | 결정 |
|------|------|
| Category 4개 구성 | ✅ `tools` / `engineering` / `vibe-coding` / `philosophy` |
| Tag vocabulary 유지 여부 | ✅ 유지 (SEO 자산 보존 + 다축 표현력) |
| URL 구조 | ✅ Option A — `/blog/<slug>` 유지 |
| 언어 | ✅ 영문 전용 (category/tag 화면 표시 + slug) |
| Tag 필터 UI 범위 | ✅ 표준 (접힘 + 검색창 + URL 쿼리 동기화) |

---

## 2. 스코프 (구현 항목 22개)

### 2.1 Data model — TAG_VOCABULARY 재편 + CATEGORY_VOCABULARY 신설

**파일**: `apps/web/src/lib/tags.ts`

```ts
// NEW
export const CATEGORY_VOCABULARY = {
  tools:       "Tools",        // Products (tene, bkit, ADK family)
  engineering: "Engineering",  // Concepts, architecture, crypto
  "vibe-coding": "Vibe Coding", // AI-assisted dev practice
  philosophy:  "Philosophy",   // AI era essays, opinion, founder story
} as const;

export type CategoryKey = keyof typeof CATEGORY_VOCABULARY;

// REVISED
export const TAG_VOCABULARY = {
  // Products
  tene:                 "tene",
  bkit:                 "bkit",
  // Technical domain
  security:             "Security",
  cryptography:         "Cryptography",
  architecture:         "Architecture",
  devsecops:            "DevSecOps",
  cli:                  "CLI",
  // AI era concepts
  "harness-engineering":"Harness Engineering",
  workflow:             "Workflow",
  "claude-code":        "Claude Code",
  cursor:               "Cursor",
  // Format / character
  tutorial:             "Tutorial",
  comparison:           "Comparison",
  "deep-dive":          "Deep Dive",
  "founder-story":      "Founder Story",
} as const;
```

**제외**: `ai` (과포괄, category로 대체), `go` (과세분화), `vibe-coding` (category로 승격)

### 2.2 Frontmatter schema 확장

**파일**: `apps/web/src/lib/blog.ts` (BlogPostMeta 타입 + 파싱)

```ts
export interface BlogPostMeta {
  slug: string;
  title: string;
  description: string;
  publishedAt: string;
  updatedAt?: string;
  category: CategoryKey;          // NEW - 필수
  tags: TagKey[];                  // 기존, 2-4개
  author: string;
  cover?: string;
  canonicalUrl?: string;
  faqs: { question: string; answer: string }[];
}
```

`gray-matter` 파싱 후 validation에서 `category` 누락/잘못된 값 빌드 실패 처리.

### 2.3 기존 9편 MDX frontmatter 마이그레이션

각 파일 frontmatter에 `category:` 1줄 추가 + tags 재편.

| 글 | 새 Category | 새 Tags |
|---|---|---|
| `ai-reads-env.mdx` | `vibe-coding` | security, devsecops, tene |
| `bkit-harness-engineering-vibe-coding.mdx` | `vibe-coding` | bkit, harness-engineering, workflow |
| `bkit-pdca-for-claude-code.mdx` | `vibe-coding` | bkit, workflow, tutorial, claude-code |
| `claude-code-safe-api-keys.mdx` | `vibe-coding` | tene, claude-code, security, tutorial |
| `cursor-secret-management-2026.mdx` | `vibe-coding` | tene, cursor, security, tutorial |
| `doppler-alternative-journey.mdx` | `tools` | tene, comparison, founder-story |
| `dotenv-vault-alternatives.mdx` | `tools` | tene, comparison, tutorial, devsecops |
| `migrate-env-60s.mdx` | `tools` | tene, tutorial, cli |
| `xchacha20-for-devs.mdx` | `engineering` | cryptography, architecture, security, tene |

분포: `vibe-coding` 5 · `tools` 3 · `engineering` 1 · `philosophy` 0.
→ philosophy 0편 상태. 후속 plan(`blog-content-batch-*`)에서 채움.

### 2.4 Routes

**유지 (변경 없음)**:
- `/blog` — 인덱스
- `/blog/<slug>` — 개별 글 (canonical 유지 → **SEO 보호**)
- `/blog/tag/<tag>` — tag 필터 페이지 (15개)

**신규**:
- `/blog/category/<category>` — category 필터 페이지 (4개)
- `/blog?tags=a,b` — 인스타식 멀티 tag 필터 (query param 기반 클라이언트 필터)

**파일**:
- `apps/web/src/app/blog/category/[category]/page.tsx` (신규, generateStaticParams)
- `apps/web/src/app/blog/page.tsx` (수정 — 필터 UI 추가)

### 2.5 UI 컴포넌트

| 위치 | 컴포넌트 | 동작 |
|---|---|---|
| `/blog` 상단 | 카테고리 pill 4개 | 클릭 시 `/blog/category/<slug>` 이동 |
| `/blog` 중단 | `<TagFilter>` | 기본 접힘. "Filter by topic ▾" 클릭 시 확장: 검색창 + top 5 tag pills + "Show all" → 전체 15개 |
| `/blog/category/<x>` 상단 | category 제목 + 설명 + 해당 글 목록 | sitemap + SEO |
| `/blog/tag/<x>` 상단 | tag 제목 + count + 해당 글 목록 | 기존 동작 유지 |
| 글 카드 | category 배지 (색상 구분) + tag 2개만 표시 + `+N more` | 모바일에서도 깔끔 |
| 글 상세 | Hero 아래에 category + 전체 tag 표시 | 기존 유사 |

**신규 파일**:
- `apps/web/src/components/blog/category-pills.tsx`
- `apps/web/src/components/blog/tag-filter.tsx` (검색 + 선택 + URL 동기화)
- `apps/web/src/components/blog/category-badge.tsx`

### 2.6 Sitemap / RSS / robots

**파일**: `apps/web/src/app/sitemap.ts`

추가 URL:
- `/blog/category/tools`
- `/blog/category/engineering`
- `/blog/category/vibe-coding`
- `/blog/category/philosophy`

`/blog?tags=...`는 URL query라 sitemap 비대상.

**RSS**: 변경 없음 (글별 canonical 그대로).

### 2.7 Landing page blog 섹션 개선

**파일**: `apps/web/src/app/page.tsx` (또는 Hero 주변 섹션)

- Blog preview 섹션에 **카테고리별 최신 1편씩 표시** (4×1 grid)
- 현재는 최신 3편만 표시 — category별 균등 노출로 변경

### 2.8 /blog-new 스킬 및 문서 업데이트

**스킬 관련 파일** (tene-biz 쪽):
- `.claude/commands/blog-new.md`:
  - Phase 1 AskUserQuestion에 Category 질문 추가 (3개 → 4개 질문)
  - Plan 출력 포맷에 Category 필드 추가
- `.claude/rules/blog-content.md`:
  - §1 Frontmatter 스키마 `category` 필수 필드 추가
  - §2 Tag vocabulary 업데이트 (10개 → 15개)
  - §2에 Category 정의 섹션 추가
- `.claude/rules/blog-pdca-workflow.md`:
  - Phase 1 체크리스트에 "category 결정" 1줄 추가
- `.claude/templates/blog/article.template.mdx`:
  - frontmatter 예시에 `category:` 추가
- `.claude/templates/blog/pre-publish-checklist.md`:
  - category 유효성 확인 항목 추가

### 2.9 Layout contract 검증 재확인

**파일**: `.claude/rules/blog-content.md §7.1`

변경 없음 예상. 카테고리 배지 + tag filter 추가되지만:
- hero / breadcrumb / related 폭 = `max-w-3xl` 유지
- article grid = `720px 220px` 유지
- shiki code block 스타일 유지

Chrome MCP QA 시 추가 체크: 카테고리 pill이 모바일(375/390px)에서 overflow 없이 wrap되는지.

---

## 3. Role 분배 (Agent Teams 기준)

| Role | 담당 | 공수 |
|------|------|------|
| **Frontend Architect** | 2.4 Routes · 2.5 UI 5종 · 2.6 Sitemap · 2.7 Landing | 6h |
| **Product Manager** | 2.3 9편 frontmatter 마이그 + 분류 리뷰 | 1h |
| **Skill Author** (사용자 + Claude main) | 2.8 `.claude/` 4종 업데이트 | 2h |
| **QA** (Chrome MCP subagent) | 2.9 Layout contract + 9편 회귀 검증 | 1h |
| **Backend** | 2.1-2.2 tags.ts + blog.ts schema | 1h |

---

## 4. 타임라인

### Day 1 — Schema + Migration (4h)

- [ ] 2.1 `tags.ts` CATEGORY_VOCABULARY 추가 + TAG_VOCABULARY 재편 (30m)
- [ ] 2.2 `blog.ts` parser + 타입 확장 (30m)
- [ ] 2.3 기존 9편 frontmatter 마이그레이션 (1h, 각 ~7분)
- [ ] 2.8 `.claude/rules/blog-content.md` vocabulary 섹션 업데이트 (1h)
- [ ] Build + 기존 9편 검증 (빌드 통과, 카테고리 페이지 생성 확인) (1h)

### Day 2 — UI + Routes + Skill (5-6h)

- [ ] 2.4 `/blog/category/[category]/page.tsx` 신규 (1h)
- [ ] 2.5 `<CategoryPills>` · `<TagFilter>` · `<CategoryBadge>` 3개 (2.5h)
- [ ] 2.6 sitemap.ts 업데이트 (15m)
- [ ] 2.7 landing blog 섹션 — 카테고리별 1편씩 (45m)
- [ ] 2.8 blog-new 스킬 Phase 1 수정 (30m)

### Day 2 끝 — QA + Commit (1h)

- [ ] 2.9 Chrome MCP 회귀 테스트 (9편 + 4 카테고리 페이지 + 인덱스 필터) (30m)
- [ ] README / CHANGELOG 업데이트 (15m)
- [ ] Commit 분할 (2-3 커밋) + PR to staging (15m)

---

## 5. 검증 (Success Criteria)

### 5.1 자동 빌드 검증 (CI 통과 + HTML grep)

- `npx next build` → 에러 0 (9편 + 새 라우트 4개 SSG 포함)
- `grep -c "category" .next/server/app/blog/<slug>.html` → 각 글 ≥ 1
- `curl /sitemap.xml | grep -c "/blog/category/"` → 4
- `curl /blog/category/tools` → 200
- `curl /blog/category/philosophy` → 200 (글 0편이어도 "no posts yet" 플레이스홀더)

### 5.2 Chrome MCP layout 회귀 (blog-content.md §7.1)

기존 9편 전부 재검증:
- `articleWidth` = 720 (lg+)
- `gridCols` = `"720px 220px"`
- `horizontalScroll` = false (375/390/768px viewport)
- 신규: `<CategoryPills>` overflow 없음
- 신규: `<TagFilter>` 클릭 후 URL 쿼리 업데이트, 뒤로가기 시 상태 복원

### 5.3 SEO 영향 확인

- 기존 9편 canonical URL **변경 없음** (frontmatter의 `canonicalUrl`은 현행 유지)
- 기존 `/blog/tag/<tag>` 10개 URL 유지 (`go`, `ai` 2개는 제거 — 리디렉션 또는 404)
- 신규 `/blog/category/*` 4개 URL 생성 + sitemap 등록
- `article:section` meta는 category로 업데이트 (Schema.org `BlogPosting.articleSection`)

### 5.4 /blog-new 스킬 회귀

다음 이터레이션 때 `/blog-new` 실행 시:
- AskUserQuestion에 Category 선택지 4개 노출
- Plan 출력에 Category 필드 포함
- 생성된 MDX에 `category:` 포함
- Phase 4 Check에서 category 값이 vocabulary에 속하는지 검증

---

## 6. 리스크 & 완화

| 리스크 | 확률 | 영향 | 완화 |
|------|:---:|:---:|------|
| 기존 `go`·`ai` tag 페이지 유입 트래픽 유실 | 중 | 소 | 301 redirect: `/blog/tag/ai` → `/blog/category/vibe-coding`, `/blog/tag/go` → `/blog/tag/cli` |
| 카테고리 배지가 카드 레이아웃 깨뜨림 | 낮 | 중 | Chrome MCP QA 모바일 375px 필수 |
| `<TagFilter>` URL 쿼리 동기화 bug (뒤로가기 깨짐) | 중 | 중 | `useSearchParams` + `router.replace` 패턴. 기본 테스트 케이스 3개 |
| /blog-new 템플릿 불일치로 새 글 빌드 실패 | 낮 | 중 | Phase 4 build 체크 템플릿 기반 |
| `category` 필드 누락된 기존 글 빌드 실패 | 중 | 고 | 마이그레이션 누락 없이 9편 전부 수정, CI grep 검증 |

---

## 7. Rollback

문제 발생 시:

1. **빌드 실패**: feature 브랜치 버리고 staging으로 복귀. 기존 9편 frontmatter 수정 없음 → 영향 0.
2. **배포 후 SEO 하락 감지 (7일 모니터)**: category 페이지만 noindex 처리 후 원인 조사. frontmatter 롤백 가능.
3. **/blog-new 스킬 회귀**: `.claude/` 디렉터리만 이전 버전으로 복구. 본 repo 코드는 영향 없음 (symlink 구조).

---

## 8. 의존성 및 후속 작업

- **선행**: 없음
- **병행**: homebrew 결정 (`feature/homebrew-tap` 브랜치 paused) — 별개
- **후속**:
  - `blog-content-batch-2026-04-23.plan.md` — 오늘 작성할 글 batch (별도 plan 문서에서 관리)
  - 추후: `/blog/category/philosophy`가 비어 있으므로 우선순위 높음
  - 추후: `ai-discoverability` plan에서 언급된 Blog 관련 항목 (article:section meta 등) 통합

---

## 9. 참조

- 대화 세션: 2026-04-23 (blog categorization discussion)
- 기존 tag 정의: `apps/web/src/lib/tags.ts` (10개)
- 기존 blog parser: `apps/web/src/lib/blog.ts`
- Layout contract: `tene-biz/.claude/rules/blog-content.md §7.1`
- Blog PDCA workflow: `tene-biz/.claude/rules/blog-pdca-workflow.md`
- Article 템플릿: `tene-biz/.claude/templates/blog/article.template.mdx`
- Existing blog posts: `apps/web/content/blog/*.mdx` (9편)

---

**작성**: 2026-04-23 (Claude main, discussion-based, no separate PRD)
**검토 대기**: 사용자 승인 후 `/pdca design` 단계 진입
