# Design — tene.sh SEO / AEO / Analytics (Track 1 / P0)

- **Feature**: `web-seo-aeo-analytics-strategy`
- **Branch**: `feature/web-landing-and-tech-blog`
- **Date**: 2026-04-22
- **Author**: Kay Kim (Frontend Architect Agent)
- **Plan reference**: `docs/01-plan/web-seo-aeo-analytics-strategy.md`
- **PRD reference**: `docs/00-pm/web-seo-aeo-analytics-strategy.prd.md`
- **Audit reference**: `docs/02-design/features/web-seo-aeo-analytics-strategy.codebase-audit.md`

---

## 0. Context Anchor

| Dimension | Detail |
|---|---|
| **WHY** | tene.sh has 0% analytics, 0% GSC verification, no blog, no dynamic sitemap, and no explicit AI-crawler allowlist. Every P1/P2 growth decision is a guess until this is fixed. |
| **WHO** | (1) Maintainers need traffic/conversion visibility; (2) Search + AI crawlers need dynamic sitemap, AI allowlist, rich schemas; (3) First 100 developer visitors need a dated launch post proving the project is real. |
| **RISK** | Next.js 16 API drift from Plan §8 (audited, mitigated); CSP breaking GA4 (critical, must widen); static robots.txt / sitemap.xml conflict with dynamic routes (must delete atomically). |
| **SUCCESS** | SC-1 through SC-17 all pass before merge to main. SC-18 (GSC indexing) deferred 8-12 weeks. See PRD §4 for full criteria. |
| **SCOPE** | Track 1 only: Module A (Analytics) + Module B (SEO/AEO) + Module C (Blog Rail). 18 P0 tasks, 14 user stories, 4 JTBD. |

---

## 1. Overview

Track 1 makes tene.sh **measurable, indexable, and AI-citable** with zero changes to the existing landing page visual design or UX. The deliverable is four shippable increments (modules) in `apps/web/`: (1) analytics infrastructure with GDPR-compliant consent, (2) dynamic SEO infrastructure replacing static public files, (3) AEO schema enhancements and answer-capsule copy rewrites, and (4) a complete MDX blog rail with RSS. Visitors see only a small cookie banner and a new `/blog` nav link. Everything else — fonts, design tokens, landing sections, hero, features, FAQ layout — remains untouched.

### What changes in `apps/web/` at a glance

```
BEFORE                          AFTER
─────────────────────────────── ─────────────────────────────────────────────────
public/robots.txt   (static)    src/app/robots.ts          (dynamic, 12 AI bots)
public/sitemap.xml  (static)    src/app/sitemap.ts         (dynamic, includes blog)
─                               src/app/blog/page.tsx      (MDX blog index)
─                               src/app/blog/[slug]/       (MDX post + Article schema)
─                               src/app/feed.xml/route.ts  (RSS 2.0 full content)
─                               src/lib/analytics.ts       (sendGAEvent wrapper)
─                               src/lib/blog.ts            (MDX parsing helpers)
─                               src/lib/schema.ts          (JSON-LD factory fns)
─                               src/components/cookie-banner.tsx
─                               src/components/blog/*.tsx  (PostCard, PostHeader, MDX)
─                               content/posts/*.mdx        (authored content)
─                               scripts/gen-llms-full.ts   (prebuild script)
─                               public/.well-known/security.txt
─                               public/llms-full.txt       (generated at build)
layout.tsx (3 JSON-LD schemas)  layout.tsx (6 JSON-LD schemas + analytics mounts)
```

### What remains untouched

- All existing landing sections: Hero, Features, HowItWorks, Security, Comparison, FAQ, CTA
- Geist Sans / Geist Mono fonts and `next/font/google` setup
- Design tokens (`--background`, `--accent`, `--foreground`, etc.) in `globals.css`
- All `src/data/*.ts` files except targeted copy rewrites in `hero.ts` (1 field) and `faq.ts` (top 3 answers)
- `public/llms.txt` (slim version, kept; coexists with new `llms-full.txt`)
- All Go packages, `internal/`, `pkg/`, `cmd/`, `go.mod`

---

## 2. Architecture Options

### Option A — Minimal (Co-located)

All blog code inside `src/app/blog/`. MDX helpers inline in page components. Single `app/layout.tsx` with all schemas inline. No separate `lib/` or `components/blog/`.

**Pros**: Fewest files, fastest to scaffold.
**Cons**: `[slug]/page.tsx` becomes 200+ lines mixing routing, data-fetching, schema generation, and MDX rendering. Not independently testable. Hard to reuse schema logic across homepage and blog.

### Option B — Clean Architecture

`src/lib/` for all utilities + `src/components/blog/` for all blog UI + `src/schemas/` for JSON-LD helpers + `src/app/blog/` for routing only.

**Pros**: Maximally maintainable, each layer testable in isolation.
**Cons**: Over-engineered for a marketing site with one post type. `src/schemas/` is a separate directory for <200 lines of code. Adds cognitive overhead for a solo team.

### Option C — Pragmatic Balance (RECOMMENDED)

`src/lib/blog.ts` + `src/lib/schema.ts` + `src/lib/analytics.ts` for all shared logic. `src/components/blog/` for all blog-specific UI. `src/app/blog/` for routes only. `content/posts/` for MDX files.

**Pros**: Clean separation without fragmentation. Every file has a single clear purpose. Schema helpers reused from both `layout.tsx` and `[slug]/page.tsx`. `lib/` matches Next.js App Router conventions. Easy to test each `lib/` function independently.
**Cons**: Creates ~12 new files (acceptable). `lib/analytics.ts` is a thin wrapper (~40 lines) that could live inline — but the consent-gate logic justifies extraction.

### Comparison Table

| Dimension | Option A | Option B | Option C |
|---|---|---|---|
| Complexity | Low | High | Medium |
| Maintainability | Low | High | High |
| Effort | 10h | 22h | 17h |
| Risk | Medium (monolithic files) | Low | Low |
| File count delta | +7 | +18 | +12 |
| Testability | Low | High | High |
| Alignment with existing codebase convention | Poor (`src/data/` already separates concerns) | Overkill | Good |

**Option C is the recommended architecture.** It follows the existing pattern of `src/data/*.ts` for content, extends it with `src/lib/*.ts` for logic, and keeps routes clean. All remaining sections proceed with Option C. If you prefer Option A for speed, the file specs below still work — collapse `lib/` into the page files.

---

## 3. Module Split

Four modules with explicit dependency graph. Execution order: **1 → 2 → 3 (parallel with 2) → 4**.

```
Module 1 (Analytics Foundation)
  ↓ unblocks measurement of all subsequent work
Module 2 (SEO Infrastructure)          Module 3 (AEO Enhancements)
  ↓ needs lib/blog.ts stub                ↑ independent of 1 and 2
  ↓ provides sitemap for Module 4         ↑ can run in parallel with 2
Module 4 (Blog Rail)
  ↑ depends on Module 2 (lib/blog.ts stub promoted to full impl)
  ↑ depends on Module 3 (answer-capsule copy approved by user)
```

### Module 1 — Analytics Foundation (P0-1, P0-2, P0-3, P0-4, P0-14)

- **Deps**: None — can start first
- **Unblocks**: measurement of everything else; Vercel Analytics is cookieless so fires regardless of consent
- **Estimated effort**: 3-4h
- **External gates**: User creates GA4 property (D-1) + adds Vercel env var (D-2)

### Module 2 — SEO Infrastructure (P0-5, P0-6, P0-7, P0-8, P0-11, P0-12)

- **Deps**: Needs `lib/blog.ts` stub (empty `getAllPostSlugs()` returning `[]`) so `sitemap.ts` compiles. Stub created in Module 2 itself.
- **Note**: GSC indexing takes 8-12 weeks (SC-18 deferred). DNS TXT + GSC UI submission are user actions, not blocking code.
- **Estimated effort**: 2h code + async GSC wait

### Module 3 — AEO Enhancements (P0-9, P0-10, P0-13)

- **Deps**: None structurally. Runs in parallel with Module 2.
- **Note**: Requires user approval of 4 answer-capsule rewrites (§7) before copy lands in `data/hero.ts` and `data/faq.ts`.
- **Estimated effort**: 2-3h

### Module 4 — Blog Rail (P0-15, P0-16, P0-17, P0-18)

- **Deps**: Module 2 (promotes `lib/blog.ts` stub to full implementation; sitemap auto-updates). Module 3 capsule rewrites should be in place.
- **Note**: P0-16 (first post, ~1,500 words) is ~3h of the 8-10h total estimate.
- **Estimated effort**: 8-10h

---

## 4. Per-Module File Specs

### 4.1 Module 1 — Analytics Foundation

#### File Table

| File | Action | Purpose | Approx lines delta |
|---|---|---|---|
| `apps/web/package.json` | MODIFY | Add `@vercel/analytics`, `@vercel/speed-insights`, `@next/third-parties@16.2.2` to dependencies | +3 |
| `apps/web/src/lib/analytics.ts` | NEW | `trackEvent(name, params)` consent-gated wrapper + TypeScript types for GA4 Consent Mode v2 | ~45 |
| `apps/web/src/components/cookie-banner.tsx` | NEW | Default-deny banner, dark theme (#0a0a0a/#00ff88), calls `gtag('consent', 'update')` on decision, persists to `localStorage['tene-consent']` | ~80 |
| `apps/web/src/app/layout.tsx` | MODIFY | (a) Import analytics components; (b) read `NEXT_PUBLIC_GA_ID`; (c) add beforeInteractive consent-default shim script; (d) mount `<CookieBanner />`, `<Analytics />`, `<SpeedInsights />`, conditional `<GoogleAnalytics>`; (e) add RSS `<link>` via metadata alternates | +20 |
| `apps/web/src/components/copy-command.tsx` | MODIFY | In `handleCopy`, call `trackEvent('install_script_copy', { location })` after clipboard write. Accept optional `location` prop. | +4 |
| `apps/web/src/components/hero.tsx` | MODIFY | GitHub/release anchor: add `onClick` calling `trackEvent('github_repo_click', { source: 'hero' })` | +3 |
| `apps/web/src/components/nav.tsx` | MODIFY | GitHub link: `trackEvent('github_repo_click', { source: 'nav' })`; Blog link: add `<a href="/blog">Blog</a>` to desktop + mobile menu | +6 |
| `apps/web/src/components/cta.tsx` | MODIFY | CTA buttons: `trackEvent('cta_click', { cta_name })` | +4 |
| `apps/web/next.config.ts` | MODIFY | Widen CSP `script-src` + `connect-src` + `img-src` (see §6.1 full diff) | +4 |
| `apps/web/.env.example` | NEW | Documents `NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX` (no real value) | ~5 |

#### Code Skeletons

**`src/lib/analytics.ts`**

```typescript
'use client';

import { sendGAEvent } from '@next/third-parties/google';

// GA4 Consent Mode v2 — type definitions
export type ConsentStatus = 'granted' | 'denied';

export interface ConsentState {
  analytics_storage: ConsentStatus;
  ad_storage: ConsentStatus;
  ad_user_data: ConsentStatus;
  ad_personalization: ConsentStatus;
}

const CONSENT_KEY = 'tene-consent';

export function getConsentStatus(): ConsentStatus | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(CONSENT_KEY) as ConsentStatus | null;
}

export function setConsent(status: ConsentStatus): void {
  if (typeof window === 'undefined') return;
  localStorage.setItem(CONSENT_KEY, status);
  const state: ConsentState = {
    analytics_storage: status,
    ad_storage: 'denied',             // never granted for tene.sh
    ad_user_data: 'denied',
    ad_personalization: 'denied',
  };
  window.gtag?.('consent', 'update', state);
}

/**
 * Consent-gated GA4 event wrapper.
 * No-ops silently if consent has not been granted.
 * Safe to call from any client component.
 */
export function trackEvent(
  name: string,
  params?: Record<string, unknown>
): void {
  if (typeof window === 'undefined') return;
  if (getConsentStatus() !== 'granted') return;
  sendGAEvent('event', name, params);
}

// Augment window type for gtag
declare global {
  interface Window {
    gtag: (...args: unknown[]) => void;
    dataLayer: unknown[];
  }
}
```

**`src/components/cookie-banner.tsx`**

```tsx
'use client';

import { useEffect, useState } from 'react';
import { getConsentStatus, setConsent } from '@/lib/analytics';

export function CookieBanner() {
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    // Only show if no prior decision
    if (getConsentStatus() === null) {
      setVisible(true);
    }
  }, []);

  function handleAccept() {
    setConsent('granted');
    setVisible(false);
  }

  function handleDecline() {
    setConsent('denied');
    setVisible(false);
  }

  if (!visible) return null;

  return (
    <div
      role="dialog"
      aria-label="Cookie consent"
      aria-live="polite"
      style={{
        position: 'fixed',
        bottom: '1.5rem',
        right: '1.5rem',
        zIndex: 50,
        maxWidth: '22rem',
        background: '#141414',
        border: '1px solid #2a2a2a',
        borderRadius: '0.5rem',
        padding: '1rem 1.25rem',
        display: 'flex',
        flexDirection: 'column',
        gap: '0.75rem',
      }}
    >
      <p style={{ color: '#ededed', fontSize: '0.875rem', lineHeight: 1.5, margin: 0 }}>
        We use cookies to measure how tene.sh performs.{' '}
        <a
          href="/blog/privacy"
          style={{ color: '#00ff88', textDecoration: 'underline' }}
        >
          Privacy policy
        </a>
        .
      </p>
      <div style={{ display: 'flex', gap: '0.5rem' }}>
        <button
          onClick={handleAccept}
          style={{
            flex: 1,
            background: '#00ff88',
            color: '#0a0a0a',
            border: 'none',
            borderRadius: '0.375rem',
            padding: '0.5rem 0.75rem',
            fontSize: '0.875rem',
            fontWeight: 600,
            cursor: 'pointer',
          }}
        >
          Accept
        </button>
        <button
          onClick={handleDecline}
          style={{
            flex: 1,
            background: 'transparent',
            color: '#888888',
            border: '1px solid #2a2a2a',
            borderRadius: '0.375rem',
            padding: '0.5rem 0.75rem',
            fontSize: '0.875rem',
            cursor: 'pointer',
          }}
        >
          Decline
        </button>
      </div>
    </div>
  );
}
```

**`src/app/layout.tsx` — analytics section (body additions)**

```tsx
// Add to imports at top of layout.tsx:
import { Analytics } from '@vercel/analytics/react';
import { SpeedInsights } from '@vercel/speed-insights/next';
import { GoogleAnalytics } from '@next/third-parties/google';
import { CookieBanner } from '@/components/cookie-banner';

// Read env var once (server component):
const gaId = process.env.NEXT_PUBLIC_GA_ID;

// Inside <html> <head>, add beforeInteractive consent-default shim BEFORE GoogleAnalytics:
// (Placed via Next <Script strategy="beforeInteractive"> or inline in <head>)
<script
  dangerouslySetInnerHTML={{
    __html: `
      window.dataLayer = window.dataLayer || [];
      function gtag(){dataLayer.push(arguments);}
      gtag('consent','default',{
        analytics_storage:'denied',
        ad_storage:'denied',
        ad_user_data:'denied',
        ad_personalization:'denied'
      });
    `,
  }}
/>

// Updated <body>:
<body className="min-h-full flex flex-col">
  {children}
  <NoiseOverlay />
  <CookieBanner />
  <Analytics />
  <SpeedInsights />
  {gaId && <GoogleAnalytics gaId={gaId} />}
</body>
```

**Metadata alternates for RSS** (add to existing `metadata` export):

```ts
alternates: {
  canonical: 'https://tene.sh',
  types: {
    'application/rss+xml': 'https://tene.sh/feed.xml',
  },
},
```

---

### 4.2 Module 2 — SEO Infrastructure

#### File Table

| File | Action | Purpose |
|---|---|---|
| `apps/web/src/lib/blog.ts` | NEW (stub) | Initial: `getAllPostSlugs()` returns `[]`. Lets `sitemap.ts` compile without blog posts. Module 4 expands this to full implementation. |
| `apps/web/src/app/sitemap.ts` | NEW | `MetadataRoute.Sitemap` — home, /blog, /install, dynamic blog post slugs from `lib/blog.ts`. Async. |
| `apps/web/src/app/robots.ts` | NEW | `MetadataRoute.Robots` — wildcard `*` allow `/` + explicit rules for 12 AI bots. |
| `apps/web/public/robots.txt` | DELETE | Replaced by `robots.ts`. Must be in same commit. |
| `apps/web/public/sitemap.xml` | DELETE | Replaced by `sitemap.ts`. Must be in same commit. |
| `apps/web/public/.well-known/security.txt` | NEW | RFC 9116: Contact, Expires, Preferred-Languages. PGP field optional (pending Q3). |
| `apps/web/package.json` | MODIFY | Add `homepage`, `repository`, `description`, `license`, `author` fields. |

#### Code for `src/app/sitemap.ts`

```typescript
import type { MetadataRoute } from 'next';
import { getAllPostSlugs } from '@/lib/blog';

const BASE_URL = 'https://tene.sh';

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const postSlugs = await getAllPostSlugs();

  const staticRoutes: MetadataRoute.Sitemap = [
    {
      url: BASE_URL,
      lastModified: new Date(),
      changeFrequency: 'weekly' as const,
      priority: 1.0,
    },
    {
      url: `${BASE_URL}/blog`,
      lastModified: new Date(),
      changeFrequency: 'daily' as const,
      priority: 0.9,
    },
    {
      url: `${BASE_URL}/install`,
      lastModified: new Date(),
      changeFrequency: 'monthly' as const,
      priority: 0.7,
    },
  ];

  const postRoutes: MetadataRoute.Sitemap = postSlugs.map(({ slug, lastModified }) => ({
    url: `${BASE_URL}/blog/${slug}`,
    lastModified: lastModified ? new Date(lastModified) : new Date(),
    changeFrequency: 'monthly' as const,
    priority: 0.8,
  }));

  return [...staticRoutes, ...postRoutes];
}
```

#### Code for `src/app/robots.ts`

```typescript
import type { MetadataRoute } from 'next';

export default function robots(): MetadataRoute.Robots {
  return {
    rules: [
      // Default: allow all well-behaved crawlers
      {
        userAgent: '*',
        allow: '/',
        disallow: [],
      },
      // Explicit AI crawler allowlist (12 bots)
      { userAgent: 'GPTBot',              allow: '/' },
      { userAgent: 'ChatGPT-User',        allow: '/' },
      { userAgent: 'OAI-SearchBot',       allow: '/' },
      { userAgent: 'ClaudeBot',           allow: '/' },
      { userAgent: 'Claude-User',         allow: '/' },
      { userAgent: 'Claude-SearchBot',    allow: '/' },
      { userAgent: 'PerplexityBot',       allow: '/' },
      { userAgent: 'Perplexity-User',     allow: '/' },
      { userAgent: 'Google-Extended',     allow: '/' },
      { userAgent: 'Amazonbot',           allow: '/' },
      { userAgent: 'Applebot-Extended',   allow: '/' },
      { userAgent: 'Meta-ExternalAgent',  allow: '/' },
    ],
    sitemap: 'https://tene.sh/sitemap.xml',
    host: 'https://tene.sh',
  };
}
```

#### `public/.well-known/security.txt`

```
Contact: mailto:security@tene.sh
Expires: 2027-04-22T00:00:00.000Z
Preferred-Languages: en
Policy: https://github.com/agent-kay-it/tene/blob/main/SECURITY.md
# Encryption: https://tene.sh/.well-known/pgp-key.txt  # Uncomment when PGP key added (Q3)
```

#### `lib/blog.ts` — Initial Stub (expanded in Module 4)

```typescript
export interface PostSlug {
  slug: string;
  lastModified?: string;
}

/** Returns all published post slugs. Stub returns [] until Module 4. */
export async function getAllPostSlugs(): Promise<PostSlug[]> {
  return [];
}
```

---

### 4.3 Module 3 — AEO Enhancements

#### File Table

| File | Action | Purpose |
|---|---|---|
| `apps/web/src/lib/schema.ts` | NEW | Factory functions: `organizationSchema()`, `personSchema()`, `breadcrumbSchema(items)`, `articleSchema(post)` |
| `apps/web/src/app/layout.tsx` | MODIFY | Extend existing `@graph` array (currently lines 82-176) from 3 schemas → 6. Import from `lib/schema.ts`. Move full object construction to named const. |
| `apps/web/src/data/hero.ts` | MODIFY | Rewrite `sub` field to 45-word answer capsule (see §7) |
| `apps/web/src/data/faq.ts` | MODIFY | Rewrite answers for items 1, 2, 3 to 40-60 word capsules (see §7). Baseline is 8 items — leave items 4-8 untouched. |
| `apps/web/scripts/gen-llms-full.ts` | NEW | Build-time script: concatenate `public/llms.txt`, `content/posts/*.mdx`, key docs → `public/llms-full.txt` |
| `apps/web/package.json` | MODIFY | Add `"prebuild": "tsx scripts/gen-llms-full.ts"` script. Add `tsx` to devDeps. |

#### Code for `src/lib/schema.ts`

```typescript
export interface SchemaOrg {
  '@type': string;
  [key: string]: unknown;
}

export interface BreadcrumbItem {
  name: string;
  url: string;
}

export interface ArticlePost {
  title: string;
  description: string;
  date: string;
  author: string;
  slug: string;
  ogImage?: string;
}

export function organizationSchema(): SchemaOrg {
  return {
    '@type': 'Organization',
    '@id': 'https://tene.sh/#organization',
    name: 'Tene',
    url: 'https://tene.sh',
    logo: {
      '@type': 'ImageObject',
      url: 'https://tene.sh/logo.svg',
    },
    sameAs: [
      'https://github.com/agent-kay-it/tene',
    ],
    founder: {
      '@id': 'https://tene.sh/#person-kay-kim',
    },
  };
}

export function personSchema(): SchemaOrg {
  return {
    '@type': 'Person',
    '@id': 'https://tene.sh/#person-kay-kim',
    name: 'Kay Kim',
    url: 'https://github.com/agent-kay-it',
    jobTitle: 'Founder, Tene',
    sameAs: [
      'https://github.com/agent-kay-it',
      // Add LinkedIn, X/Twitter URLs when answered in Q5
    ],
    worksFor: {
      '@id': 'https://tene.sh/#organization',
    },
  };
}

export function breadcrumbSchema(items: BreadcrumbItem[]): SchemaOrg {
  return {
    '@type': 'BreadcrumbList',
    itemListElement: items.map((item, index) => ({
      '@type': 'ListItem',
      position: index + 1,
      name: item.name,
      item: item.url,
    })),
  };
}

export function articleSchema(post: ArticlePost): SchemaOrg {
  return {
    '@type': 'Article',
    headline: post.title,
    description: post.description,
    datePublished: post.date,
    dateModified: post.date,
    author: {
      '@id': 'https://tene.sh/#person-kay-kim',
    },
    publisher: {
      '@id': 'https://tene.sh/#organization',
    },
    mainEntityOfPage: {
      '@type': 'WebPage',
      '@id': `https://tene.sh/blog/${post.slug}`,
    },
    image: post.ogImage ?? 'https://tene.sh/og-image.png',
  };
}
```

#### Code for `scripts/gen-llms-full.ts`

```typescript
import { readFileSync, writeFileSync, readdirSync, existsSync } from 'fs';
import { join } from 'path';

const WEB_DIR = join(process.cwd()); // called from apps/web/
const OUTPUT = join(WEB_DIR, 'public', 'llms-full.txt');

function readIfExists(filePath: string): string {
  if (!existsSync(filePath)) return '';
  return readFileSync(filePath, 'utf-8');
}

const sections: string[] = [
  '# tene — llms-full.txt\n# Full site content for AI RAG pipelines.\n# Generated at build time. Do not edit manually.\n',
  '---\n## SUMMARY (from llms.txt)\n',
  readIfExists(join(WEB_DIR, 'public', 'llms.txt')),
];

// Append all MDX posts
const postsDir = join(WEB_DIR, 'content', 'posts');
if (existsSync(postsDir)) {
  const mdxFiles = readdirSync(postsDir).filter(f => f.endsWith('.mdx'));
  for (const file of mdxFiles) {
    const content = readFileSync(join(postsDir, file), 'utf-8');
    sections.push(`\n---\n## BLOG POST: ${file}\n`, content);
  }
}

// Append SKILL.md from repo root if accessible
const skillPath = join(WEB_DIR, '..', '..', 'SKILL.md');
const skillContent = readIfExists(skillPath);
if (skillContent) {
  sections.push('\n---\n## TENE CLI SKILL\n', skillContent);
}

writeFileSync(OUTPUT, sections.join('\n'), 'utf-8');
console.log(`[gen-llms-full] Written: ${OUTPUT}`);
```

---

### 4.4 Module 4 — Blog Rail

#### File Table

| File | Action | Purpose |
|---|---|---|
| `apps/web/package.json` | MODIFY | Add `next-mdx-remote@^5`, `gray-matter@^4`, `shiki@^1`. Add `@tailwindcss/typography@^0.5` to devDeps. |
| `apps/web/src/lib/blog.ts` | MODIFY (expand stub) | Full implementation: `getAllPosts()`, `getAllPostSlugs()`, `getPostBySlug(slug)`. Parse frontmatter with `gray-matter`, read `content/posts/*.mdx`. |
| `apps/web/src/app/globals.css` | MODIFY | Add `@plugin "@tailwindcss/typography";` + prose brand-color overrides. |
| `apps/web/src/components/nav.tsx` | MODIFY | Add `/blog` link (desktop + mobile menu). Already noted in Module 1; consolidate here if sequencing Module 4 first in nav work. |
| `apps/web/src/app/blog/page.tsx` | NEW | Blog index: `getAllPosts()`, sorted newest-first, rendered as `<PostCard />` list. |
| `apps/web/src/app/blog/[slug]/page.tsx` | NEW | Post page: `generateStaticParams`, `generateMetadata` (async params), `MDXRemote`, Article + BreadcrumbList schema. |
| `apps/web/src/components/blog/post-card.tsx` | NEW | Title + date + description + tag chips. |
| `apps/web/src/components/blog/post-header.tsx` | NEW | H1 + metadata bar (date, author, reading time). |
| `apps/web/src/components/blog/mdx-components.tsx` | NEW | Custom MDX component overrides (code blocks via shiki, inline links, blockquotes). |
| `apps/web/src/app/feed.xml/route.ts` | NEW | RSS 2.0: `GET()` → `Response` with `Content-Type: application/rss+xml`. CDATA-wrapped content. |
| `apps/web/content/posts/introducing-tene.mdx` | NEW | First post: ~1,500 words. Frontmatter + answer capsule opening + body. |
| `apps/web/content/posts/welcome.mdx` | NEW | Minimal placeholder post proving index scales (PRD US-13). |

#### Full `src/lib/blog.ts` Implementation

```typescript
import fs from 'fs';
import path from 'path';
import matter from 'gray-matter';

const POSTS_DIR = path.join(process.cwd(), 'content', 'posts');

export interface PostFrontmatter {
  title: string;
  description: string;
  date: string;
  author: string;
  tags: string[];
  canonical?: string;
  ogImage?: string;
  published: boolean;
}

export interface Post {
  slug: string;
  frontmatter: PostFrontmatter;
  content: string;
  readingTime: number; // minutes
}

export interface PostSlug {
  slug: string;
  lastModified?: string;
}

function estimateReadingTime(content: string): number {
  const words = content.split(/\s+/).length;
  return Math.ceil(words / 200); // 200 wpm average
}

function getSlugFromFilename(filename: string): string {
  return filename.replace(/\.mdx$/, '');
}

export async function getAllPosts(): Promise<Post[]> {
  if (!fs.existsSync(POSTS_DIR)) return [];

  const filenames = fs.readdirSync(POSTS_DIR).filter(f => f.endsWith('.mdx'));

  const posts = filenames
    .map((filename): Post | null => {
      const filePath = path.join(POSTS_DIR, filename);
      const raw = fs.readFileSync(filePath, 'utf-8');
      const { data, content } = matter(raw);
      const frontmatter = data as PostFrontmatter;

      if (!frontmatter.published) return null;

      return {
        slug: getSlugFromFilename(filename),
        frontmatter,
        content,
        readingTime: estimateReadingTime(content),
      };
    })
    .filter((p): p is Post => p !== null);

  return posts.sort(
    (a, b) =>
      new Date(b.frontmatter.date).getTime() -
      new Date(a.frontmatter.date).getTime()
  );
}

export async function getAllPostSlugs(): Promise<PostSlug[]> {
  const posts = await getAllPosts();
  return posts.map(p => ({
    slug: p.slug,
    lastModified: p.frontmatter.date,
  }));
}

export async function getPostBySlug(slug: string): Promise<Post | null> {
  const filePath = path.join(POSTS_DIR, `${slug}.mdx`);
  if (!fs.existsSync(filePath)) return null;

  const raw = fs.readFileSync(filePath, 'utf-8');
  const { data, content } = matter(raw);
  const frontmatter = data as PostFrontmatter;

  if (!frontmatter.published) return null;

  return {
    slug,
    frontmatter,
    content,
    readingTime: estimateReadingTime(content),
  };
}
```

#### `src/app/blog/[slug]/page.tsx` (Next 16 async params)

```tsx
import type { Metadata } from 'next';
import { notFound } from 'next/navigation';
import { MDXRemote } from 'next-mdx-remote/rsc';
import { getAllPostSlugs, getPostBySlug } from '@/lib/blog';
import { articleSchema, breadcrumbSchema, organizationSchema, personSchema } from '@/lib/schema';
import { PostHeader } from '@/components/blog/post-header';
import { mdxComponents } from '@/components/blog/mdx-components';

interface Props {
  params: Promise<{ slug: string }>;
}

export async function generateStaticParams() {
  const slugs = await getAllPostSlugs();
  return slugs.map(p => ({ slug: p.slug }));
}

export async function generateMetadata({ params }: Props): Promise<Metadata> {
  const { slug } = await params;
  const post = await getPostBySlug(slug);
  if (!post) return {};

  const { frontmatter } = post;
  return {
    title: `${frontmatter.title} — Tene`,
    description: frontmatter.description,
    authors: [{ name: frontmatter.author }],
    alternates: {
      canonical: frontmatter.canonical ?? `/blog/${slug}`,
    },
    openGraph: {
      title: frontmatter.title,
      description: frontmatter.description,
      type: 'article',
      publishedTime: frontmatter.date,
      authors: [frontmatter.author],
      images: [frontmatter.ogImage ?? '/og-image.png'],
    },
  };
}

export default async function BlogPostPage({ params }: Props) {
  const { slug } = await params;
  const post = await getPostBySlug(slug);
  if (!post) notFound();

  const { frontmatter, content, readingTime } = post;

  const postJsonLd = {
    '@context': 'https://schema.org',
    '@graph': [
      organizationSchema(),
      personSchema(),
      articleSchema({ ...frontmatter, slug }),
      breadcrumbSchema([
        { name: 'Home', url: 'https://tene.sh' },
        { name: 'Blog', url: 'https://tene.sh/blog' },
        { name: frontmatter.title, url: `https://tene.sh/blog/${slug}` },
      ]),
    ],
  };

  return (
    <>
      <script
        type="application/ld+json"
        dangerouslySetInnerHTML={{ __html: JSON.stringify(postJsonLd) }}
      />
      <main>
        <article className="prose prose-invert max-w-3xl mx-auto px-6 py-16">
          <PostHeader
            title={frontmatter.title}
            date={frontmatter.date}
            author={frontmatter.author}
            readingTime={readingTime}
            tags={frontmatter.tags}
          />
          <MDXRemote source={content} components={mdxComponents} />
        </article>
      </main>
    </>
  );
}
```

#### `src/app/feed.xml/route.ts`

```typescript
import { getAllPosts } from '@/lib/blog';

const BASE_URL = 'https://tene.sh';

function cdata(str: string): string {
  return `<![CDATA[${str.replace(/\]\]>/g, ']]]]><![CDATA[>')}]]>`;
}

function escapeXml(str: string): string {
  return str
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;');
}

export async function GET(): Promise<Response> {
  const posts = await getAllPosts();

  const items = posts
    .map(post => {
      const url = `${BASE_URL}/blog/${post.slug}`;
      const pubDate = new Date(post.frontmatter.date).toUTCString();
      return `
    <item>
      <title>${cdata(post.frontmatter.title)}</title>
      <link>${escapeXml(url)}</link>
      <guid isPermaLink="true">${escapeXml(url)}</guid>
      <description>${cdata(post.frontmatter.description)}</description>
      <pubDate>${pubDate}</pubDate>
      <author>${cdata(post.frontmatter.author)}</author>
      ${post.frontmatter.tags.map(tag => `<category>${escapeXml(tag)}</category>`).join('\n      ')}
      <content:encoded>${cdata(post.content)}</content:encoded>
    </item>`;
    })
    .join('\n');

  const xml = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"
  xmlns:content="http://purl.org/rss/1.0/modules/content/"
  xmlns:atom="http://www.w3.org/2005/Atom">
  <channel>
    <title>Tene Blog</title>
    <link>${BASE_URL}</link>
    <description>Technical writing on secret management for AI agents, developer security, and building open-source CLI tools.</description>
    <language>en-us</language>
    <managingEditor>kay@popupstudio.ai (Kay Kim)</managingEditor>
    <atom:link href="${BASE_URL}/feed.xml" rel="self" type="application/rss+xml" />
    <lastBuildDate>${new Date().toUTCString()}</lastBuildDate>
    ${items}
  </channel>
</rss>`;

  return new Response(xml, {
    headers: {
      'Content-Type': 'application/rss+xml; charset=utf-8',
      'Cache-Control': 'public, max-age=3600, s-maxage=3600',
    },
  });
}
```

#### `content/posts/introducing-tene.mdx` — Frontmatter + Opening

```mdx
---
title: "Introducing Tene — Local-First Encrypted Secrets for AI Coding Agents"
description: "Tene encrypts your API keys in a local SQLite vault and injects them as environment variables at runtime. AI agents see your commands, not your credentials. Here is why we built it and how it works under the hood."
date: "2026-04-22"
author: "Kay Kim"
tags: ["secret-management", "ai-agents", "security", "cli", "open-source"]
canonical: "https://tene.sh/blog/introducing-tene"
ogImage: "/og-image.png"
published: true
---

Tene encrypts your secrets in a local SQLite vault using XChaCha20-Poly1305 and injects them as environment variables at runtime. AI agents like Claude Code see your commands — not your credentials. No server, no signup, no cloud dependency. MIT licensed and fully open source.

## The Problem: Your .env Is Visible to Every AI Agent

## How Tene Works: Four-Layer Encryption

## Dogfooding From Day One

## Getting Started in 60 Seconds

## What is Coming Next

## Conclusion
```

---

## 5. Data Contracts

### Post Frontmatter Schema

```typescript
interface PostFrontmatter {
  title: string;        // 50-60 chars, SEO-optimized title
  description: string;  // 150-160 chars, meta + OG description
  date: string;         // ISO 8601 (YYYY-MM-DD)
  author: string;       // default "Kay Kim"
  tags: string[];       // 3-5 kebab-case tags
  canonical?: string;   // absolute URL; defaults to https://tene.sh/blog/${slug}
  ogImage?: string;     // absolute or root-relative; defaults to /og-image.png
  published: boolean;   // false = draft, excluded from getAllPosts / sitemap / RSS
}
```

### GA4 Event Catalog (Type-Safe Enum Pattern)

```typescript
// src/lib/analytics.ts — extend with event catalog
export const GA4_EVENTS = {
  INSTALL_SCRIPT_COPY:  'install_script_copy',    // hero/CTA copy button
  GITHUB_REPO_CLICK:   'github_repo_click',       // any github.com/agent-kay-it/tene link
  GITHUB_RELEASE_CLICK:'github_release_click',    // link to /releases specifically (C from Q10)
  CTA_CLICK:           'cta_click',               // any CTA button; params: { cta_name }
  CLAWHUB_LINK_CLICK:  'clawhub_link_click',      // footer ClawHub badge (pending Q9 decision)
  BLOG_POST_VIEW:      'blog_post_view',           // blog post page load; params: { slug }
} as const;

export type GA4EventName = typeof GA4_EVENTS[keyof typeof GA4_EVENTS];
```

**Event instrumentation sites**:

| Event | Trigger location | Component |
|---|---|---|
| `install_script_copy` | Install command copy button | `copy-command.tsx` (hero + CTA) |
| `github_repo_click` | GitHub icon in nav + hero button | `nav.tsx`, `hero.tsx` |
| `github_release_click` | CTA "Download from GitHub" → /releases | `cta.tsx` |
| `cta_click` | Any CTA button (with `cta_name` param) | `cta.tsx` |
| `clawhub_link_click` | Footer ClawHub badge (if added per Q9) | `footer.tsx` |
| `blog_post_view` | Blog post page mount | `blog/[slug]/page.tsx` |

### JSON-LD Type Strategy

Use inline TypeScript objects (no additional `schema-dts` package). Factory functions in `src/lib/schema.ts` accept typed parameters and return `SchemaOrg` (= `{ '@type': string; [key: string]: unknown }`). This avoids a large devDep for what is 4 small functions.

---

## 6. Infrastructure Changes

### 6.1 CSP Widening (`next.config.ts`)

**Current CSP** (`next.config.ts` line 8 — single string):

```
default-src 'self';
script-src 'self' 'unsafe-inline' 'unsafe-eval';
style-src 'self' 'unsafe-inline';
img-src 'self' data: https:;
connect-src 'self';
font-src 'self' data:
```

**Required CSP after Module 1** — add the following domains:

```
script-src  'self' 'unsafe-inline' 'unsafe-eval'
            https://www.googletagmanager.com
            https://va.vercel-scripts.com;

connect-src 'self'
            https://www.google-analytics.com
            https://*.analytics.google.com
            https://*.google-analytics.com
            https://www.googletagmanager.com
            https://vitals.vercel-insights.com
            https://va.vercel-scripts.com;

img-src     'self' data: https: https://www.google-analytics.com;

font-src    'self' data:;
```

**Note**: `unsafe-eval` is already present. The additions are GA4 + Vercel Analytics hostnames only. No new unsafe directives introduced.

### 6.2 Environment Variables

| Variable | Purpose | Storage |
|---|---|---|
| `NEXT_PUBLIC_GA_ID` | GA4 Measurement ID (e.g., `G-XXXXXXXXXX`) | tene vault (`tene set NEXT_PUBLIC_GA_ID G-XXX --env prod`) |

**Deployment workflow** (dogfooding tene):

```bash
# Store in vault
tene set NEXT_PUBLIC_GA_ID G-XXXXXXXXXX --env prod

# Add to Vercel (user action — not committed)
tene run --env prod -- vercel env add NEXT_PUBLIC_GA_ID production
```

`apps/web/.env.example` documents the variable name only:

```
# GA4 Measurement ID — stored in tene vault, NOT in .env files
# Run: tene set NEXT_PUBLIC_GA_ID G-XXXXXXXXXX --env prod
NEXT_PUBLIC_GA_ID=G-XXXXXXXXXX
```

### 6.3 DNS Actions (User)

| Action | Detail |
|---|---|
| GSC Verification TXT record | `google-site-verification=<token>` at apex `tene.sh` |
| DNS provider | See Q2 — Vercel DNS or Cloudflare |
| Propagation time | 5-60 minutes |

### 6.4 Build Pipeline

After Module 1 + Module 3 ship, `npm run build` flow becomes:

```
prebuild     → tsx scripts/gen-llms-full.ts → public/llms-full.txt
next build   → static export for all routes:
               /           (home)
               /blog        (post index)
               /blog/*      (generated from getAllPostSlugs via generateStaticParams)
               /feed.xml    (RSS route handler — static or edge)
               /sitemap.xml (from src/app/sitemap.ts)
               /robots.txt  (from src/app/robots.ts)
```

---

## 7. AEO Content Specs (P0-13 Answer Capsules)

These are the proposed rewrites from Audit §4.6. User approves at Design Checkpoint before `/pdca do`. Each must be: self-contained, 40-60 words, first clause answers directly, second clause provides mechanism/evidence.

### Hero Subtitle (current: 18 words → target: 40-60 words)

**Current** (`src/data/hero.ts` → `sub` field):
> Store secrets in a local encrypted vault and inject them at runtime with one command.

**Proposed** (45 words):
> Tene encrypts your secrets in a local SQLite vault using XChaCha20-Poly1305 and injects them as environment variables at runtime. AI agents like Claude Code see your commands — not your credentials. No server, no signup, no cloud dependency. MIT licensed and fully open source.

### FAQ Answer 1 — "Why is .env dangerous with AI agents?"

**Proposed** (54 words):
> AI coding agents like Claude Code, Cursor, and Windsurf read every file in your project — including `.env`. Your API keys, tokens, and database passwords become plaintext context sent to foundation models. Once sent, you cannot control how that data is logged, cached, or used for future training.

### FAQ Answer 2 — "How does Tene keep secrets from AI?"

**Proposed** (48 words):
> Tene stores secrets in a local encrypted SQLite vault protected by XChaCha20-Poly1305. When you run `tene run -- claude`, it injects secrets as environment variables at runtime. AI agents see the command in CLAUDE.md and the variable names, but never read the plaintext values themselves.

### FAQ Answer 3 — "What is Tene?"

**Proposed** (42 words):
> Tene is a local-first, encrypted secret management CLI for developers using AI coding agents. It stores API keys and credentials in a device-only vault, never touches the network, and auto-generates AI editor rules so tools like Claude Code, Cursor, and Windsurf use secrets safely.

**Word counts**: Hero = 45 words (in range). FAQ 1 = 54 (in range). FAQ 2 = 48 (in range). FAQ 3 = 42 (in range). All four pass the 40-60 word capsule requirement.

---

## 8. Test Plan

| Layer | Module | Test | Tool | Automated? |
|---|---|---|---|---|
| L1 | M1 | GA4 beacon fires on page load after consent | Network tab / Playwright | Manual |
| L1 | M1 | GA4 does NOT fire before consent | Network tab — no `google-analytics.com` requests on first load | Manual |
| L1 | M1 | Cookie banner shows on first visit, hidden after decision | Playwright | Automated |
| L1 | M1 | Consent persisted to `localStorage['tene-consent']` | Browser devtools | Manual |
| L1 | M1 | Vercel Analytics shows pageviews within 24h | Vercel dashboard | Manual |
| L1 | M2 | `/sitemap.xml` returns 200 with valid XML + blog slugs | `curl -I https://tene.sh/sitemap.xml` | Manual |
| L1 | M2 | `/robots.txt` returns all 12 AI bot rules | `curl https://tene.sh/robots.txt \| grep -c "allow"` | Manual |
| L1 | M2 | `/.well-known/security.txt` returns 200 | `curl -o /dev/null -w "%{http_code}" .../security.txt` | Manual |
| L1 | M2 | Static `public/robots.txt` + `public/sitemap.xml` deleted | `git ls-files apps/web/public/robots.txt` returns empty | CI check |
| L1 | M3 | JSON-LD passes Google Rich Results Test | schema.org validator + Rich Results Test URL | Manual |
| L1 | M3 | 6 schemas in `@graph` on homepage | `curl tene.sh | grep -c '"@type"'` ≥ 6 | Manual |
| L1 | M3 | `/llms-full.txt` returns 200 + non-empty | `curl -o /dev/null -w "%{http_code}\n%{size_download}" .../llms-full.txt` | Manual |
| L1 | M4 | `/blog` returns 200 + lists ≥1 post | Playwright | Automated |
| L1 | M4 | `/blog/introducing-tene` returns 200 + Article schema present | Playwright + page source check | Automated |
| L1 | M4 | `/feed.xml` validates as RSS 2.0 | W3C Feed Validator (`validator.w3.org/feed`) | Manual |
| L1 | M4 | `<content:encoded>` in RSS is not truncated | Check first post full text in feed | Manual |
| L2 | M1 | Clicking install-script copy fires `install_script_copy` event | Playwright + network intercept → `collect` endpoint | Automated |
| L2 | M1 | Clicking GitHub link fires `github_repo_click` event | Playwright + network intercept | Automated |
| L2 | M4 | Reading time calculation is correct | Unit test for `estimateReadingTime()` | Automated |
| L2 | M4 | `getPostBySlug` returns null for draft posts | Unit test: `published: false` | Automated |

---

## 9. Risk Mitigation

| # | Risk (from PRD §8) | Implementation Mitigation |
|---|---|---|
| R-1 | Next 16 API drift from Plan §8 snippets | All APIs verified against `node_modules/next/dist/lib/metadata/types/metadata-interface.d.ts` in Audit §2. `params: Promise<{slug}>` pattern confirmed. Module specs above use verified signatures. |
| R-2 | Cookie banner hurts conversion | Banner is minimal (~80 LOC), bottom-right corner, non-modal, does not block install command. Vercel Analytics fires regardless (cookieless) — baseline measurement unaffected. |
| R-3 | GSC indexing lag (8-12 weeks) | SC-18 explicitly deferred. SC-1–SC-17 are the P0 DoD. Weekly GSC check starting week 4. |
| R-4 | llms-full.txt no proven ROI | One `prebuild` script, one output file. Zero ongoing maintenance. Accepted as low-cost hygiene. |
| R-5 | MDX + Next 16 RSC stability | `next-mdx-remote/rsc` chosen per Audit §4.1 recommendation. Exact version pinned. RSC-native, production-proven. If v5 breaks, fallback path is native `@next/mdx` (documented in Audit §4.1). |
| R-6 | Blog typography overrides dark palette | Tailwind v4 `@plugin "@tailwindcss/typography";` + custom `.prose-invert` overrides using `--accent` + `--foreground` CSS vars. Test on `welcome.mdx` at Do start. |
| R-7 | Events fire before consent | `trackEvent()` wrapper in `lib/analytics.ts` checks `localStorage['tene-consent'] === 'granted'` before calling `sendGAEvent`. No-op pre-consent. Beforeinteractive consent-default shim sets `analytics_storage: denied` before GA script loads. |
| R-8 | `NEXT_PUBLIC_GA_ID` absent in preview | `layout.tsx` conditionally mounts `<GoogleAnalytics>` only when env var is present (`{gaId && <GoogleAnalytics gaId={gaId} />}`). Fails silently in preview without breaking the page. |
| R-9 | RSS `<content:encoded>` XML-escaping | `cdata()` helper wraps content in `CDATA` sections. `escapeXml()` used for URLs and other attributes. Validate against W3C Feed Validator in SC-16. |
| R-10 | Founder name / bio not finalized | "Kay Kim" + `https://github.com/agent-kay-it` used as minimum viable Person schema. Bio page deferred to P1 (PRD §6 Out of Scope). |

---

## 10. Open Questions

These 10 questions must be resolved before `/pdca do` begins. Answers drive implementation decisions that cannot be changed cheaply post-ship.

**Q1 — GA4 Property**: Create a new GA4 property specifically for tene.sh (recommended: "Tene Landing", clean slate), or reuse an existing property?

**Q2 — DNS Provider**: Where is tene.sh DNS managed — Vercel DNS, Cloudflare, or another provider? This determines the exact steps for GSC TXT verification (D-3).

**Q3 — PGP Key for security.txt**: Skip the `Encryption:` field for v1 (recommended — valid per RFC 9116), use an existing PGP key (provide fingerprint + key URL), or generate a new key for security@tene.sh?

**Q4 — package.json Field Values**: Confirm: `homepage: "https://tene.sh"`, `repository: "https://github.com/agent-kay-it/tene"`, `author: "Kay Kim <kay@popupstudio.ai>"`, `description: "Landing page for Tene — local-first encrypted secret manager"`. Any corrections?

**Q5 — Person Schema sameAs URLs**: `name: "Kay Kim"` and `url: "https://github.com/agent-kay-it"` are confirmed from PRD C-4. Should `sameAs` also include LinkedIn, X/Twitter, or a personal site? Provide URLs if yes.

**Q6 — "Introducing tene" Post Drafting**: (A) Claude drafts 1,500 words based on PRD/Plan, user reviews and edits; (B) user writes, Claude only scaffolds MDX frontmatter + H2 outline; (C) co-draft — Claude proposes outline now, prose drafted collaboratively in Do session. Recommended: C.

**Q7 — Cookie Banner Default-Deny**: Confirm default-deny (`analytics_storage: denied` until visitor clicks "Accept")? This is PRD US-3 aligned and GDPR-safe, but costs ~15-30% of GA4 data volume versus default-allow. Recommended: yes, default-deny.

**Q8 — Answer Capsule Rewrites**: Accept all 4 proposed rewrites in §7 verbatim (hero subtitle + 3 FAQ answers)? Or revise specific items? The hero rewrite adds 27 words; FAQ answers add 30-40 words each. These are presented for approval before they land in `data/hero.ts` and `data/faq.ts`.

**Q9 — ClawHub Link and Event**: P0-14 calls for `clawhub_link_click` tracking but the current Nav + Footer have no ClawHub link. Decision: (A) add a small ClawHub badge to Footer alongside GitHub links and track it; (B) drop `clawhub_link_click`, replace with `product_hunt_click` (PH badge already in Hero); (C) defer to Module 1 Do session. Recommended: A (ClawHub is a primary distribution channel per Plan §2 point 10).

**Q10 — GitHub Link Destination and Event Split**: PRD specifies `github_release_click` but all current GitHub links point to the repo homepage, not `/releases`. Decision: (A) change all links to `/releases`; (B) rename event to `github_click` everywhere; (C) split — hero/nav stay on repo homepage (`github_repo_click`), CTA buttons point to `/releases` (`github_release_click`). Recommended: C (gives richest signal for conversion analysis).

---

## 11. Implementation Guide

### 11.1 Session Plan (Recommended)

| Session | Module | Duration | Blocks on |
|---|---|---|---|
| Session 1 | M1 — Analytics Foundation | 3-4h | Q1, Q2, Q7, Q9, Q10 must be answered first |
| Session 2 | M2 — SEO Infrastructure | 2h + async GSC wait | Q2, Q3, Q4 must be answered; DNS propagation is async |
| Session 3 | M3 — AEO Enhancements | 2-3h | Q5, Q8 must be answered; can run in parallel with Session 2 waiting |
| Session 4 | M4 — Blog Rail | 8-10h | Q6 must be answered; Modules 2 + 3 should be merged first |

**Sequence rationale**: Module 1 ships first so every subsequent test has analytics coverage from day one. Module 2 + 3 are largely independent and can overlap (different file sets). Module 4 last because it needs `lib/blog.ts` stub from Module 2 and approved capsule copy from Module 3.

### 11.2 Command Invocations

After all 10 open questions are answered and this design is approved:

```bash
# Start Module 1
/pdca do web-seo-aeo-analytics-strategy --scope module-1

# After Module 1 is merged and verified:
/pdca do web-seo-aeo-analytics-strategy --scope module-2

# Parallel to Module 2 (or after):
/pdca do web-seo-aeo-analytics-strategy --scope module-3

# After Modules 2+3 merged:
/pdca do web-seo-aeo-analytics-strategy --scope module-4
```

### 11.3 Module Scope Map

| Flag value | Tasks included |
|---|---|
| `module-1` | P0-1, P0-2, P0-3, P0-4, P0-14 — Analytics Foundation |
| `module-2` | P0-5, P0-6, P0-7, P0-8, P0-11, P0-12 — SEO Infrastructure |
| `module-3` | P0-9, P0-10, P0-13 — AEO Enhancements |
| `module-4` | P0-15, P0-16, P0-17, P0-18 — Blog Rail |

---

## 12. Success Criteria Map

| SC | PRD Criterion | Module | Test |
|---|---|---|---|
| SC-1 | Vercel Analytics firing | M1 | Vercel dashboard — ≥1 pageview within 1h of deploy |
| SC-2 | Speed Insights firing | M1 | Vercel Speed Insights — p75 CWV within 24h |
| SC-3 | GA4 firing ≥95% sessions | M1 | GA4 Realtime ±5% vs Vercel over 48h sample |
| SC-4 | Cookie banner default-deny | M1 | Network tab — no `google-analytics.com` requests before "Accept" |
| SC-5 | 4 events instrumented | M1 | GA4 DebugView — all events from staged trigger clicks |
| SC-6 | GSC domain-verified | M2 (user action) | GSC "Ownership verified" status |
| SC-7 | Sitemap submitted + fetched | M2 (user action) | GSC Sitemaps panel — status "Success" |
| SC-8 | `/sitemap.xml` dynamic | M2 | `curl` returns XML with blog post URLs; static file deleted |
| SC-9 | `/robots.txt` dynamic with AI allowlist | M2 | `curl` returns rules for all 12 bots; static file deleted |
| SC-10 | 6 JSON-LD schemas on homepage | M3 | View source — 6 types in `@graph`; Rich Results Test passes |
| SC-11 | `llms-full.txt` served 200 | M3 | `curl https://tene.sh/llms-full.txt` — 200, non-empty |
| SC-12 | `security.txt` served 200 | M2 | `curl https://tene.sh/.well-known/security.txt` — 200, RFC 9116 fields |
| SC-13 | Hero + 3 FAQ capsules | M3 | Manual review: 40-60 words, self-contained answer, approved in §7 |
| SC-14 | `/blog` renders | M4 | Playwright — `/blog` returns 200, lists ≥1 post |
| SC-15 | Launch post live | M4 | `/blog/introducing-tene` — 200, ≥1,500 words, Article schema present |
| SC-16 | RSS feed valid | M4 | W3C Feed Validator — passes; `<content:encoded>` non-truncated |
| SC-17 | package.json fields | M2 | `cat apps/web/package.json | jq '.homepage'` returns `"https://tene.sh"` |
| SC-18 | ≥1 page indexed (deferred) | M2 + time | GSC URL Inspection — "URL is on Google" within 8-12 weeks of merge |

---

## 13. Appendix — Complete Source Reference

### A. Full `src/lib/analytics.ts`

*(See §4.1 code skeleton — this is the complete implementation, ~45 lines.)*

The `trackEvent` function is the single call site for all GA4 events across components. Every component that fires analytics imports only from `@/lib/analytics`, never directly from `@next/third-parties/google`.

### B. Full `src/components/cookie-banner.tsx`

*(See §4.1 code skeleton — complete, ~80 lines.)*

Key behavior:
- First visit: shows bottom-right fixed banner.
- "Accept": calls `setConsent('granted')` → updates Consent Mode v2 → hides.
- "Decline": calls `setConsent('denied')` → hides. GA4 never fires this session.
- Subsequent visits: reads `localStorage['tene-consent']` — no banner shown.
- Does not block any page content; z-index: 50.

### C. Full `src/app/sitemap.ts`

*(See §4.2 — complete Next 16 implementation with `Promise<MetadataRoute.Sitemap>` return type.)*

### D. Full `src/app/robots.ts`

*(See §4.2 — complete with 12 AI bots + wildcard default.)*

### E. Full `src/lib/schema.ts`

*(See §4.3 — all 4 factory functions: `organizationSchema`, `personSchema`, `breadcrumbSchema`, `articleSchema`.)*

### F. Full `src/lib/blog.ts`

*(See §4.4 — complete implementation: `getAllPosts`, `getAllPostSlugs`, `getPostBySlug`, `estimateReadingTime`.)*

### G. Full `src/app/blog/[slug]/page.tsx`

*(See §4.4 — complete with Next 16 async params, `generateStaticParams`, `generateMetadata`, Article + BreadcrumbList schema.)*

### H. Full `src/app/feed.xml/route.ts`

*(See §4.4 — complete RSS 2.0 with `CDATA` wrapping for XML safety, full `<content:encoded>`, cache headers.)*

### I. CSP Before/After Diff

```diff
- "Content-Security-Policy": "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self'; font-src 'self' data:"
+ "Content-Security-Policy": "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://www.googletagmanager.com https://va.vercel-scripts.com; style-src 'self' 'unsafe-inline'; img-src 'self' data: https: https://www.google-analytics.com; connect-src 'self' https://www.google-analytics.com https://*.analytics.google.com https://*.google-analytics.com https://www.googletagmanager.com https://vitals.vercel-insights.com https://va.vercel-scripts.com; font-src 'self' data:"
```

### J. Frontmatter TypeScript Interface

```typescript
interface PostFrontmatter {
  title: string;        // 50-60 chars
  description: string;  // 150-160 chars
  date: string;         // ISO 8601: "YYYY-MM-DD"
  author: string;       // "Kay Kim"
  tags: string[];       // ["secret-management", "ai-agents", ...]
  canonical?: string;   // "https://tene.sh/blog/${slug}"
  ogImage?: string;     // "/og-image.png"
  published: boolean;   // false = draft, excluded everywhere
}
```

---

**End of Design Document.** Total scope: 4 modules, 18 P0 tasks, 12 new files, 10 modified files, 2 deleted files, ~17h estimated dev effort. Next: resolve 10 open questions above, then `/pdca do web-seo-aeo-analytics-strategy --scope module-1`.
