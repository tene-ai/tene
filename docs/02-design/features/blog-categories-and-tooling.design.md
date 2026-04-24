# Blog Categories + Tooling — Design

| 항목 | 값 |
|------|-----|
| Feature | `blog-categories-and-tooling` |
| 기반 Plan | `docs/01-plan/blog-categories-and-tooling.plan.md` |
| 브랜치 | `feature/blog-categories` |

---

## 1. 아키텍처 결정

### 1.1 Data layer — 2개 export 병행

`src/lib/tags.ts`에 `CATEGORY_VOCABULARY`와 `TAG_VOCABULARY`를 동시 export. Category는 1개 필수 + Tag는 2-4개 복수. 서로 독립 평면.

### 1.2 URL 구조 — Option A (기존 URL 전부 보존)

- 글별 canonical URL 전부 그대로 (`https://tene.sh/blog/<slug>`)
- `/blog/tag/<tag>` 페이지 유지 (15개 vocabulary + dropped `go`/`ai` 301)
- `/blog/category/<category>` 신규 (4개)
- `/blog?tags=a,b` query-based 클라이언트 필터 (sitemap 비대상)

### 1.3 Frontmatter — `category` 필수화 전략

- 단일 PR에서 vocabulary 확장 + 9편 마이그레이션 + parser validation 동시 진행
- blog.ts parser는 category 누락 시 build-time 에러 throw → 역호환성 보호 불필요 (전편 마이그 완료 가정)

### 1.4 기존 tag 중 `go`·`ai`·`vibe-coding` 처리

- `go`: xchacha20 글 1편만 보유. tag 제거 후 `architecture`로 치환
- `ai`: 5편에 부착. 전부 vibe-coding category로 이동 예정이므로 tag 제거 (중복 방지)
- `vibe-coding`: tag vocabulary에서 **제거**, category로 승격. 5편의 글에서 tag 제거

---

## 2. 코드 변경 상세 (파일별)

### 2.1 `src/lib/tags.ts`

```ts
// NEW — 최상단
export const CATEGORY_VOCABULARY = {
  tools:        "Tools",
  engineering:  "Engineering",
  "vibe-coding": "Vibe Coding",
  philosophy:   "Philosophy",
} as const;

export type CategoryKey = keyof typeof CATEGORY_VOCABULARY;

export function isValidCategory(cat: string): cat is CategoryKey {
  return cat in CATEGORY_VOCABULARY;
}

export function getCategoryLabel(cat: string): string {
  return (CATEGORY_VOCABULARY as Record<string, string>)[cat] ?? cat;
}

// REVISED
export const TAG_VOCABULARY = {
  tene:                 "tene",
  bkit:                 "bkit",
  security:             "Security",
  cryptography:         "Cryptography",
  architecture:         "Architecture",
  devsecops:            "DevSecOps",
  cli:                  "CLI",
  "harness-engineering":"Harness Engineering",
  workflow:             "Workflow",
  "claude-code":        "Claude Code",
  cursor:               "Cursor",
  tutorial:             "Tutorial",
  comparison:           "Comparison",
  "deep-dive":          "Deep Dive",
  "founder-story":      "Founder Story",
} as const;
```

### 2.2 `src/lib/blog.ts`

```ts
// BlogPostFrontmatter type에 category 추가 (필수)
export type BlogPostFrontmatter = {
  // ... 기존
  category: CategoryKey;  // NEW — 필수
  tags: TagKey[];         // 타입 좁힘
  // ...
};

// loadPost() 내 validation
const category = data.category as string;
if (!category) {
  throw new Error(`[blog] ${slug}: 'category' frontmatter field is required`);
}
if (!isValidCategory(category)) {
  throw new Error(
    `[blog] ${slug}: invalid category '${category}'. ` +
    `Allowed: ${Object.keys(CATEGORY_VOCABULARY).join(", ")}`,
  );
}

// 신규 helper
export function getAllCategories(): Array<{ category: string; count: number; label: string }> {
  const counts = new Map<string, number>();
  for (const post of getAllPosts()) {
    counts.set(post.category, (counts.get(post.category) ?? 0) + 1);
  }
  // Return all 4 even if count=0 (empty philosophy still visible)
  return Object.keys(CATEGORY_VOCABULARY).map((cat) => ({
    category: cat,
    count: counts.get(cat) ?? 0,
    label: getCategoryLabel(cat),
  }));
}

export function getPostsByCategory(cat: string): BlogPostMeta[] {
  return getAllPosts().filter((p) => p.category === cat);
}
```

### 2.3 Migration — 9편 MDX frontmatter

매핑 테이블 (Plan §2.3에서 확정):

| Slug | category | tags |
|---|---|---|
| `ai-reads-env` | `vibe-coding` | `security, devsecops, tene` |
| `bkit-harness-engineering-vibe-coding` | `vibe-coding` | `bkit, harness-engineering, workflow` |
| `bkit-pdca-for-claude-code` | `vibe-coding` | `bkit, workflow, tutorial, claude-code` |
| `claude-code-safe-api-keys` | `vibe-coding` | `tene, claude-code, security, tutorial` |
| `cursor-secret-management-2026` | `vibe-coding` | `tene, cursor, security, tutorial` |
| `doppler-alternative-journey` | `tools` | `tene, comparison, founder-story` |
| `dotenv-vault-alternatives` | `tools` | `tene, comparison, tutorial, devsecops` |
| `migrate-env-60s` | `tools` | `tene, tutorial, cli` |
| `xchacha20-for-devs` | `engineering` | `cryptography, architecture, security, tene` |

### 2.4 Routes

#### 2.4.1 `/blog/category/[category]/page.tsx` (신규)

`/blog/tag/[tag]/page.tsx` 패턴 그대로 복제:
- `generateStaticParams()` → 4개 카테고리
- `generateMetadata()` → canonical, OG, title
- 본문: Hero + 해당 카테고리 글 그리드 (`PostCard` 재사용)
- JSON-LD: `CollectionPage` + `BreadcrumbList`

#### 2.4.2 `/blog/page.tsx` (수정)

- Hero 아래에 `<CategoryPills>` 추가 (4 pill)
- 기존 `<TagCloud>` 제거 → `<TagFilter>` 로 교체
- 메인 글 목록은 URL `?tags=` 파라미터 기반 클라이언트 필터 적용

#### 2.4.3 `/blog/tag/[tag]/page.tsx` (수정 — 최소)

- `generateStaticParams()`가 vocabulary 기반 15개 + 기존 dropping 2개 제외
- 본문 변경 없음

### 2.5 Components

#### 2.5.1 `src/components/blog/category-pills.tsx` (신규)

```tsx
type Props = { activeCategory?: CategoryKey };

export function CategoryPills({ activeCategory }: Props) {
  const categories = getAllCategories();  // SSR, build-time
  return (
    <div className="flex flex-wrap gap-2 justify-center">
      {categories.map(({ category, count, label }) => (
        <Link
          key={category}
          href={`/blog/category/${category}`}
          aria-current={activeCategory === category ? "page" : undefined}
          className={cn(
            "rounded-full border px-4 py-2 text-sm transition-colors",
            activeCategory === category
              ? "border-accent bg-accent/10 text-foreground"
              : "border-border bg-surface/60 text-muted hover:border-accent/40 hover:text-foreground",
          )}
        >
          {label} <span className="text-xs opacity-60">({count})</span>
        </Link>
      ))}
    </div>
  );
}
```

#### 2.5.2 `src/components/blog/category-badge.tsx` (신규)

```tsx
type Props = { category: CategoryKey; size?: "sm" | "md" };

export function CategoryBadge({ category, size = "sm" }: Props) {
  const label = getCategoryLabel(category);
  const colors: Record<CategoryKey, string> = {
    tools:          "bg-blue-500/10 text-blue-400 border-blue-500/20",
    engineering:    "bg-purple-500/10 text-purple-400 border-purple-500/20",
    "vibe-coding":  "bg-green-500/10 text-green-400 border-green-500/20",
    philosophy:     "bg-amber-500/10 text-amber-400 border-amber-500/20",
  };
  return (
    <span className={cn(
      "inline-flex items-center rounded border",
      size === "sm" ? "px-2 py-0.5 text-xs" : "px-3 py-1 text-sm",
      colors[category],
    )}>
      {label}
    </span>
  );
}
```

#### 2.5.3 `src/components/blog/tag-filter.tsx` (신규)

클라이언트 컴포넌트. URL 쿼리 `?tags=x,y` 동기화.

- 기본: `<details>` 태그로 접힘
- 펼치면: 검색 input + top 5 tag pills + "Show all 15" 토글
- 선택/해제 시 URL 업데이트 (`useSearchParams` + `useRouter.replace`)
- AND 모드 (선택된 모든 tag 포함하는 글만 표시)
- Props: `allTags: {tag, count}[]`, `onChange(selected: string[])`

Parent (`/blog/page.tsx`)에서 필터된 글 목록 렌더.

#### 2.5.4 `src/components/blog/post-card.tsx` (수정)

- 카드 상단에 `<CategoryBadge category={post.category} />` 추가
- tag 표시 3개 → 2개로 줄임 (+ `+N more` 생략 기호)

### 2.6 `src/app/sitemap.ts` (수정)

```ts
// 추가
const blogCategoryUrls: MetadataRoute.Sitemap =
  getAllCategories().map(({ category }) => ({
    url: `${base}/blog/category/${category}`,
    lastModified,
    changeFrequency: "monthly",
    priority: 0.6,
  }));

// return에 ...blogCategoryUrls 포함
// blogTagUrls는 기존 유지 (vocabulary 기반 15개만)
```

### 2.7 Landing (`src/app/page.tsx`) — 블로그 섹션 개선

기존: 최신 블로그 3편 (또는 N편) 그리드
변경: 카테고리별 최신 1편씩 4개 그리드 (없는 카테고리는 "Coming soon" 플레이스홀더)

실제 블로그 섹션이 landing에 있는지 확인 후 적용. 없으면 생략 가능.

### 2.8 /blog-new 스킬 업데이트 (`tene-biz/.claude/`)

#### 2.8.1 `.claude/rules/blog-content.md`

- §1 Frontmatter schema에 `category: {slug}` 필수 필드 추가
- §2 tag vocabulary 섹션 전면 교체 (10 → 15)
- §2에 Category 섹션 신설 (4개 설명)

#### 2.8.2 `.claude/rules/blog-pdca-workflow.md`

- Phase 1 체크리스트에 "Category 결정" 1줄 추가
- AskUserQuestion 템플릿에 Q0 추가 (Category 선택)

#### 2.8.3 `.claude/templates/blog/article.template.mdx`

- frontmatter 예시에 `category:` 1줄 추가

#### 2.8.4 `.claude/templates/blog/pre-publish-checklist.md` (존재 시)

- category 유효성 확인 항목 추가

---

## 3. 비기능 요구 (NFR)

- **SEO**: 9편 canonical URL 불변. sitemap 추가 4개. Google Search Console 재제출.
- **성능**: 모든 페이지 정적 (SSG). 필터는 클라이언트 전용 JS — 번들 < 5 KB gzipped.
- **접근성**: CategoryPills에 `aria-current`, TagFilter에 `aria-expanded`.
- **Layout contract**: `.claude/rules/blog-content.md §7.1` 준수. 카테고리 Pill이 모바일 viewport(375/390px)에서 overflow 없이 wrap.

## 4. 검증

### 4.1 빌드 검증
- `npx next build` 에러 0
- `/blog/category/<cat>` 4개 정적 경로 생성
- 기존 9편 canonical URL 불변

### 4.2 Chrome MCP QA (Phase 6)
- 9편 상세 레이아웃 (hero/article/toc/related 폭)
- `/blog` 인덱스 카테고리 pill + tag filter 인터랙션
- `/blog/category/<cat>` 4개 페이지 메타/레이아웃
- Console errors = 0

### 4.3 SEO 회귀
- `grep -c "@type\":\"BlogPosting\"" .next/server/app/blog/<slug>.html` → 1 (9편)
- `curl /sitemap.xml | grep -c "/blog/category/"` → 4
- `curl /blog/tag/ai` → 301 or 404 (no longer vocabulary)

---

## 5. 구현 순서 (tasks)

1. `src/lib/tags.ts` — vocabulary 2개 export
2. `src/lib/blog.ts` — category 타입 + helpers + validation
3. 9편 MDX frontmatter 마이그레이션
4. Build check (Day 1 끝)
5. `/blog/category/[category]/page.tsx` 신규 route
6. `CategoryPills` · `CategoryBadge` · `TagFilter` 3개 컴포넌트
7. `PostCard` 에 CategoryBadge 통합
8. `/blog/page.tsx` TagCloud → TagFilter 교체 + CategoryPills 추가
9. `sitemap.ts` category URL 추가
10. `/blog-new` 스킬 4개 파일 업데이트
11. 전체 빌드 + Chrome MCP QA
12. Gap 분석 + 반복 개선

---

## 6. 참조

- Plan: `docs/01-plan/blog-categories-and-tooling.plan.md`
- 기존 tag 파일: `apps/web/src/lib/tags.ts`
- 기존 blog parser: `apps/web/src/lib/blog.ts`
- 기존 tag route: `apps/web/src/app/blog/tag/[tag]/page.tsx` (참고 패턴)
- Layout contract: `tene-biz/.claude/rules/blog-content.md §7.1`
